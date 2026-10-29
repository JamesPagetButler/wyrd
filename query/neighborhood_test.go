package query

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JamesPagetButler/wyrd/model"
)

// Contract tests for API.Neighborhood (bma-systema #229 Wyrd-side;
// scoping ratified live-test seq=539/542/544). The fixture mirrors
// the BMA boot-time mirror shape: an NT_SEED-anchored / life-
// certificate-anchored neighborhood with provenance edges.

// Shared fixture literals (hoisted to satisfy goconst; mirror the BMA
// boot-time-mirror node types).
const (
	certID   model.NodeID   = "cert"
	seed1ID  model.NodeID   = "seed1"
	seed2ID  model.NodeID   = "seed2"
	obs1ID   model.NodeID   = "obs1"
	obs2ID   model.NodeID   = "obs2"
	obs3ID   model.NodeID   = "obs3"
	orphanID model.NodeID   = "orphan"
	certType model.NodeType = "bma.lineage.life-certificate"
	seedType model.NodeType = "bma.seed"
	obsType  model.NodeType = "bma.observation"
)

func nbNode(id model.NodeID, typ model.NodeType, immune bool, sal float64) model.Node {
	return model.Node{
		ID:         id,
		Type:       typ,
		Tier:       model.TierComplex,
		Created:    time.Unix(0, 0),
		TierImmune: immune,
		Salience:   sal,
	}
}

func nbEdge(id model.HyperedgeID, nodes ...model.NodeID) model.Hyperedge {
	return model.Hyperedge{
		ID:          id,
		Nodes:       nodes,
		Weight:      model.NewComplexWeight(1, 0),
		IsSymmetric: true,
		Created:     time.Unix(0, 0),
	}
}

// fixtureGraph builds:
//
//	cert (NT_LIFE_CERTIFICATE, immune) ── e-cs1 ── seed1 (immune)
//	cert ── e-cs2 ── seed2 (immune)
//	seed1 ── e-s1o1 ── obs1 ─ e-o1o3 ─ obs3   (depth 2, 3 from cert)
//	seed2 ── e-s2o2 ── obs2
//	orphan (no edges)
//	ext: edge e-out with one endpoint outside any depth-2 set
func fixtureGraph(t *testing.T) *model.Graph {
	t.Helper()
	g := model.NewGraph()
	nodes := []model.Node{
		nbNode(certID, certType, true, 1.0),
		nbNode(seed1ID, seedType, true, 1.0),
		nbNode(seed2ID, seedType, true, 1.0),
		nbNode(obs1ID, obsType, false, 0.3),
		nbNode(obs2ID, obsType, false, 0.2),
		nbNode(obs3ID, obsType, false, 0.1),
		nbNode(orphanID, obsType, false, 0.0),
	}
	for _, n := range nodes {
		if err := g.AddNode(n); err != nil {
			t.Fatalf("AddNode(%s): %v", n.ID, err)
		}
	}
	edges := []model.Hyperedge{
		nbEdge("e-cs1", certID, seed1ID),
		nbEdge("e-cs2", certID, seed2ID),
		nbEdge("e-s1o1", seed1ID, obs1ID),
		nbEdge("e-s2o2", seed2ID, obs2ID),
		nbEdge("e-o1o3", obs1ID, obs3ID),
	}
	for _, e := range edges {
		if err := g.AddHyperedge(e); err != nil {
			t.Fatalf("AddHyperedge(%s): %v", e.ID, err)
		}
	}
	return g
}

func TestNeighborhood_Depth2CertAnchored(t *testing.T) {
	q := New(fixtureGraph(t))
	sg, err := q.Neighborhood(certID, 2, 100)
	if err != nil {
		t.Fatalf("Neighborhood: %v", err)
	}
	// Depth 2 from cert: cert(0), seed1+seed2(1), obs1+obs2(2).
	// obs3 is 3 hops; orphan unreachable.
	wantIDs := []model.NodeID{certID, seed1ID, seed2ID, obs1ID, obs2ID}
	gotIDs := make([]model.NodeID, len(sg.Nodes))
	for i, n := range sg.Nodes {
		gotIDs[i] = n.ID
	}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Errorf("nodes = %v, want %v (BFS ring order, lexicographic within ring)", gotIDs, wantIDs)
	}
	if sg.Truncated {
		t.Error("Truncated should be false — budget not hit")
	}
	// Hops annotation.
	wantHops := map[model.NodeID]int{certID: 0, seed1ID: 1, seed2ID: 1, obs1ID: 2, obs2ID: 2}
	for _, n := range sg.Nodes {
		if n.Hops != wantHops[n.ID] {
			t.Errorf("node %s hops = %d, want %d", n.ID, n.Hops, wantHops[n.ID])
		}
	}
	// Induced edges: e-cs1, e-cs2, e-s1o1, e-s2o2 are inside;
	// e-o1o3 has obs3 outside → excluded.
	wantEdges := []model.HyperedgeID{"e-cs1", "e-cs2", "e-s1o1", "e-s2o2"}
	gotEdges := make([]model.HyperedgeID, len(sg.Edges))
	for i, e := range sg.Edges {
		gotEdges[i] = e.ID
	}
	if !reflect.DeepEqual(gotEdges, wantEdges) {
		t.Errorf("edges = %v, want %v (induced-only, lexicographic)", gotEdges, wantEdges)
	}
}

func TestNeighborhood_Deterministic(t *testing.T) {
	q := New(fixtureGraph(t))
	a, err := q.Neighborhood(certID, 2, 100)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	b, err := q.Neighborhood(certID, 2, 100)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if string(ja) != string(jb) {
		t.Error("two calls over the same graph state are not byte-identical")
	}
}

func TestNeighborhood_BudgetEvictionPriority(t *testing.T) {
	// Anchor with a depth-1 ring of 5: one TierImmune seed, one
	// high-salience obs, three low-salience obs. Budget of 3 total
	// (anchor + 2): the immune seed and the high-salience obs must
	// survive the cut.
	g := model.NewGraph()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(g.AddNode(nbNode("a", "bma.seed", true, 1.0)))
	must(g.AddNode(nbNode("imm", "bma.seed", true, 1.0)))
	must(g.AddNode(nbNode("hot", "bma.observation", false, 0.9)))
	must(g.AddNode(nbNode("c1", "bma.observation", false, 0.1)))
	must(g.AddNode(nbNode("c2", "bma.observation", false, 0.1)))
	must(g.AddNode(nbNode("c3", "bma.observation", false, 0.1)))
	for i, id := range []model.NodeID{"imm", "hot", "c1", "c2", "c3"} {
		must(g.AddHyperedge(nbEdge(model.HyperedgeID(fmt.Sprintf("e%d", i)), "a", id)))
	}

	sg, err := New(g).Neighborhood("a", 1, 3)
	if err != nil {
		t.Fatalf("Neighborhood: %v", err)
	}
	if !sg.Truncated {
		t.Error("Truncated must be true — ring was cut")
	}
	got := map[model.NodeID]bool{}
	for _, n := range sg.Nodes {
		got[n.ID] = true
	}
	if !got["imm"] {
		t.Error("TierImmune node evicted from its own shard — priority broken")
	}
	if !got["hot"] {
		t.Error("high-salience node evicted before low-salience peers")
	}
	if len(sg.Nodes) != 3 {
		t.Errorf("len(nodes) = %d, want 3 (budget)", len(sg.Nodes))
	}
}

// wyrd#104 TierImmune-guarantee test IDs (hoisted to satisfy goconst;
// distinct from the other fixtures' IDs so no literal repeats ≥3×).
const (
	tiAnchor model.NodeID = "ti-anchor"
	tiImm1   model.NodeID = "ti-imm1"
	tiImm2   model.NodeID = "ti-imm2"
	tiImm3   model.NodeID = "ti-imm3"
	tiHot    model.NodeID = "ti-hot"
	tiC1     model.NodeID = "ti-c1"
	tiC2     model.NodeID = "ti-c2"
)

// wyrd#104: when the mandatory TierImmune set fits the budget, ALL of
// them are retained and the remaining budget is filled by optional
// (non-immune) nodes by Salience — TierImmune is never dropped.
func TestNeighborhood_TierImmuneAllRetainedOptionalFills(t *testing.T) {
	g := model.NewGraph()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(g.AddNode(nbNode(tiAnchor, seedType, true, 1.0)))
	// Depth-1 ring: 2 mandatory (immune) + 3 optional; budget after
	// anchor = 3, so 1 optional slot remains after the 2 mandatory.
	must(g.AddNode(nbNode(tiImm1, seedType, true, 1.0)))
	must(g.AddNode(nbNode(tiImm2, seedType, true, 1.0)))
	must(g.AddNode(nbNode(tiHot, obsType, false, 0.9)))
	must(g.AddNode(nbNode(tiC1, obsType, false, 0.1)))
	must(g.AddNode(nbNode(tiC2, obsType, false, 0.1)))
	for _, id := range []model.NodeID{tiImm1, tiImm2, tiHot, tiC1, tiC2} {
		must(g.AddHyperedge(nbEdge(model.HyperedgeID("tie-"+string(id)), tiAnchor, id)))
	}

	sg, err := New(g).Neighborhood(tiAnchor, 1, 4) // anchor + 3
	if err != nil {
		t.Fatalf("Neighborhood: %v", err)
	}
	got := map[model.NodeID]bool{}
	for _, n := range sg.Nodes {
		got[n.ID] = true
	}
	if !got[tiImm1] || !got[tiImm2] {
		t.Fatalf("a TierImmune node was dropped: got=%v — wyrd#104 invariant broken", got)
	}
	if !got[tiHot] {
		t.Error("highest-salience optional not retained in the remaining budget")
	}
	if len(sg.Nodes) != 4 {
		t.Errorf("len(nodes) = %d, want 4 (anchor + 2 immune + 1 optional)", len(sg.Nodes))
	}
	if !sg.Truncated {
		t.Error("Truncated must be true — optional nodes were cut")
	}
}

// wyrd#104 (the KILL test): when the mandatory TierImmune set at a ring
// alone exceeds the budget, Neighborhood fails LOUD rather than silently
// dropping a defeater. The old "sort-then-ring[:budget]" returned a
// Subgraph with an immune node silently cut — this test KILLS that.
func TestNeighborhood_TierImmuneMandatoryOverflowErrorsLoud(t *testing.T) {
	g := model.NewGraph()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(g.AddNode(nbNode(tiAnchor, seedType, true, 1.0)))
	// Depth-1 ring of 3 mandatory immune nodes; budget after anchor = 2 < 3.
	for _, id := range []model.NodeID{tiImm1, tiImm2, tiImm3} {
		must(g.AddNode(nbNode(id, seedType, true, 1.0)))
		must(g.AddHyperedge(nbEdge(model.HyperedgeID("tie-"+string(id)), tiAnchor, id)))
	}

	sg, err := New(g).Neighborhood(tiAnchor, 1, 3) // budget after anchor = 2, but 3 mandatory
	if err == nil {
		t.Fatalf("expected a loud error (3 mandatory TierImmune > budget 2), got none; sg=%+v — old ring[:budget] would silently drop a defeater", sg)
	}
	if len(sg.Nodes) != 0 {
		t.Errorf("Subgraph must be zero on error, got %d nodes (no partial shard presented as complete)", len(sg.Nodes))
	}
	if !errors.Is(err, ErrTierImmuneExcluded) || !strings.Contains(err.Error(), string(tiImm3)) {
		t.Errorf("error should name the excluded TierImmune node (tiImm3, dropped by the budget), got: %v", err)
	}
}

// wyrd#104 (second silent-drop path, found by qbp-architecture): when an
// earlier hop fills the budget exactly, the budget==0 branch skipped the
// next ring with Truncated=true — silently dropping a TierImmune defeater
// at a deeper hop. Now it fails LOUD. Repro: depth 2, maxNodes 2 — anchor
// + one optional at hop 1 fill the budget; the hop-2 TierImmune node must
// not vanish with err=nil. KILLS the old budget==0 break.
func TestNeighborhood_TierImmuneDeeperHopBudgetExhaustedErrorsLoud(t *testing.T) {
	g := model.NewGraph()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(g.AddNode(nbNode(tiAnchor, seedType, true, 1.0)))
	must(g.AddNode(nbNode(tiHot, obsType, false, 0.9)))  // hop-1 optional
	must(g.AddNode(nbNode(tiImm1, seedType, true, 1.0))) // hop-2 mandatory defeater
	must(g.AddHyperedge(nbEdge("tie-a-hot", tiAnchor, tiHot)))
	must(g.AddHyperedge(nbEdge("tie-hot-imm", tiHot, tiImm1)))

	// anchor + the hop-1 optional fill maxNodes=2 exactly; the hop-2
	// TierImmune node is beyond the budget (the budget==0 branch).
	sg, err := New(g).Neighborhood(tiAnchor, 2, 2)
	if err == nil {
		t.Fatalf("expected a loud error (hop-2 TierImmune unreachable within budget), got none; sg=%+v — old budget==0 break silently dropped the defeater", sg)
	}
	if len(sg.Nodes) != 0 {
		t.Errorf("Subgraph must be zero on error, got %d nodes", len(sg.Nodes))
	}
	if !errors.Is(err, ErrTierImmuneExcluded) || !strings.Contains(err.Error(), string(tiImm1)) {
		t.Errorf("error should name the excluded TierImmune node, got: %v", err)
	}
}

// wyrd#104 §I4 KILL tests — qbp-architecture's probes P1/P2 (issuecomment-
// 6487654183): a TierImmune node can be lost without any cut site touching
// it, by sitting beyond an exhausted budget (P1) or behind a cut optional
// (P2). The whole-traversal check must fail loud on both.

// P1: a TierImmune node sits one hop beyond an exactly-exhausted budget.
// a→o1→o2→imm, Neighborhood(a,3,2): anchor + o1 fill the budget, {o2} hits
// budget==0 and breaks (benign — no immune in {o2}), so the hop-3 immune
// node is never reached.
func TestNeighborhood_P1_ImmuneBeyondExhaustedBudget(t *testing.T) {
	g := model.NewGraph()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(g.AddNode(nbNode(tiAnchor, seedType, true, 1.0)))
	must(g.AddNode(nbNode(tiC1, obsType, false, 0.5)))   // o1
	must(g.AddNode(nbNode(tiC2, obsType, false, 0.5)))   // o2
	must(g.AddNode(nbNode(tiImm1, seedType, true, 1.0))) // imm (hop 3)
	must(g.AddHyperedge(nbEdge("p1-a-o1", tiAnchor, tiC1)))
	must(g.AddHyperedge(nbEdge("p1-o1-o2", tiC1, tiC2)))
	must(g.AddHyperedge(nbEdge("p1-o2-imm", tiC2, tiImm1)))

	sg, err := New(g).Neighborhood(tiAnchor, 3, 2)
	if err == nil {
		t.Fatalf("P1: hop-3 TierImmune silently dropped (err=nil); sg=%+v", sg)
	}
	if len(sg.Nodes) != 0 {
		t.Errorf("Subgraph must be zero on error, got %d nodes", len(sg.Nodes))
	}
	if !errors.Is(err, ErrTierImmuneExcluded) || !strings.Contains(err.Error(), string(tiImm1)) {
		t.Errorf("error should name the excluded TierImmune node, got: %v", err)
	}
}

// P2: optional cut at hop 1 drops the only path to a hop-2 immune node.
// a→{hi1,hi2,lo}, lo→imm, Neighborhood(a,2,3): the hop-1 cut evicts lo (the
// lowest-salience optional, no immune in that ring), so imm — reachable only
// through lo — never enters a ring.
func TestNeighborhood_P2_CutOptionalHidesDeeperImmune(t *testing.T) {
	g := model.NewGraph()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(g.AddNode(nbNode(tiAnchor, seedType, true, 1.0)))
	must(g.AddNode(nbNode(tiHot, obsType, false, 0.9)))  // hi1
	must(g.AddNode(nbNode(tiC2, obsType, false, 0.8)))   // hi2
	must(g.AddNode(nbNode(tiC1, obsType, false, 0.1)))   // lo
	must(g.AddNode(nbNode(tiImm1, seedType, true, 1.0))) // imm (behind lo)
	must(g.AddHyperedge(nbEdge("p2-a-hi1", tiAnchor, tiHot)))
	must(g.AddHyperedge(nbEdge("p2-a-hi2", tiAnchor, tiC2)))
	must(g.AddHyperedge(nbEdge("p2-a-lo", tiAnchor, tiC1)))
	must(g.AddHyperedge(nbEdge("p2-lo-imm", tiC1, tiImm1)))

	sg, err := New(g).Neighborhood(tiAnchor, 2, 3)
	if err == nil {
		t.Fatalf("P2: hop-2 TierImmune silently dropped (path cut, err=nil); sg=%+v", sg)
	}
	if len(sg.Nodes) != 0 {
		t.Errorf("Subgraph must be zero on error, got %d nodes", len(sg.Nodes))
	}
	if !errors.Is(err, ErrTierImmuneExcluded) || !strings.Contains(err.Error(), string(tiImm1)) {
		t.Errorf("error should name the excluded TierImmune node, got: %v", err)
	}
}

// wyrd#104 follow-up (arch): the budget-exclusion error wraps the exported
// sentinel ErrTierImmuneExcluded, so callers (bma#358 --inject / boot
// auto-inject) match with errors.Is, never the message text.
func TestNeighborhood_ErrTierImmuneExcluded_IsMatchable(t *testing.T) {
	g := model.NewGraph()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(g.AddNode(nbNode(tiAnchor, seedType, true, 1.0)))
	for _, id := range []model.NodeID{tiImm1, tiImm2, tiImm3} {
		must(g.AddNode(nbNode(id, seedType, true, 1.0)))
		must(g.AddHyperedge(nbEdge(model.HyperedgeID("se-"+string(id)), tiAnchor, id)))
	}
	_, err := New(g).Neighborhood(tiAnchor, 1, 3) // 3 mandatory immune > budget 2
	if !errors.Is(err, ErrTierImmuneExcluded) {
		t.Fatalf("budget-exclusion error must wrap ErrTierImmuneExcluded (errors.Is), got: %v", err)
	}
	// The sentinel is specific: a params error must NOT match it.
	if _, perr := New(g).Neighborhood(tiAnchor, 0, 3); errors.Is(perr, ErrTierImmuneExcluded) {
		t.Errorf("a params error must not match ErrTierImmuneExcluded, got: %v", perr)
	}
}

func TestNeighborhood_ExactFitBeyondRingSetsTruncated(t *testing.T) {
	q := New(fixtureGraph(t))
	// Budget 3 = cert + both seeds (ring 1 fits exactly); ring 2
	// (obs1, obs2) exists beyond → Truncated.
	sg, err := q.Neighborhood(certID, 2, 3)
	if err != nil {
		t.Fatalf("Neighborhood: %v", err)
	}
	if len(sg.Nodes) != 3 {
		t.Fatalf("len(nodes) = %d, want 3", len(sg.Nodes))
	}
	if !sg.Truncated {
		t.Error("Truncated must be true — depth-2 ring exists beyond the exactly-full budget")
	}
}

func TestNeighborhood_AnchorNotFound(t *testing.T) {
	q := New(fixtureGraph(t))
	_, err := q.Neighborhood("ghost", 2, 100)
	if !errors.Is(err, model.ErrNodeNotFound) {
		t.Errorf("error %v does not unwrap to model.ErrNodeNotFound", err)
	}
}

func TestNeighborhood_RejectsBadParams(t *testing.T) {
	q := New(fixtureGraph(t))
	if _, err := q.Neighborhood(certID, 0, 100); err == nil {
		t.Error("depth 0 must be rejected")
	}
	if _, err := q.Neighborhood(certID, 2, 0); err == nil {
		t.Error("maxNodes 0 must be rejected")
	}
}

func TestNeighborhood_OrphanAnchor(t *testing.T) {
	q := New(fixtureGraph(t))
	sg, err := q.Neighborhood(orphanID, 2, 100)
	if err != nil {
		t.Fatalf("Neighborhood: %v", err)
	}
	if len(sg.Nodes) != 1 || sg.Nodes[0].ID != orphanID {
		t.Errorf("orphan shard = %v, want [orphan] only", sg.Nodes)
	}
	if len(sg.Edges) != 0 {
		t.Errorf("orphan shard has %d edges, want 0", len(sg.Edges))
	}
	if sg.Truncated {
		t.Error("orphan shard is complete — Truncated must be false")
	}
}

func TestNeighborhood_WriteShapeEqualsReadShape(t *testing.T) {
	// Condition-1 round-trip (qbp-architecture §I4, PR #85): the
	// architectural requirement is read-shape ≡ write-shape — what
	// the BMA boot-time mirror WRITES (a model.Hyperedge provenance
	// edge) is exactly what Neighborhood PROJECTS back (a
	// SubgraphEdge). This asserts the round-trip mechanically rather
	// than by assumption, using the mirror's literal oriented
	// provenance shape (cert→seed "read-at-founding":
	// IsSymmetric=false, Heads=[0], Tails=[1]) from bma-systema #244.
	g := model.NewGraph()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(g.AddNode(nbNode(certID, certType, true, 1.0)))
	must(g.AddNode(nbNode(seed1ID, seedType, true, 1.0)))

	// WRITE side: the exact hyperedge the boot-time mirror emits.
	written := model.Hyperedge{
		ID:          "prov-read-at-founding-seed1",
		Nodes:       []model.NodeID{certID, seed1ID},
		Weight:      model.NewComplexWeight(1, 0),
		IsSymmetric: false,
		Created:     time.Unix(0, 0),
		Heads:       []int{0}, // source: cert
		Tails:       []int{1}, // sink: seed
	}
	must(g.AddHyperedge(written))

	// READ side: what Neighborhood projects back.
	sg, err := New(g).Neighborhood(certID, 1, 10)
	if err != nil {
		t.Fatalf("Neighborhood: %v", err)
	}
	if len(sg.Edges) != 1 {
		t.Fatalf("len(edges) = %d, want 1", len(sg.Edges))
	}
	got := sg.Edges[0]

	// Round-trip equality: every field the SubgraphEdge carries must
	// equal the written hyperedge's value. Drop a field from the
	// projection (as the pre-c5b2e8e shape dropped orientation) and
	// this fails.
	want := SubgraphEdge{
		ID:          written.ID,
		Nodes:       written.Nodes,
		IsSymmetric: written.IsSymmetric,
		Heads:       written.Heads,
		Tails:       written.Tails,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("read-shape != write-shape:\n got  %+v\n want %+v", got, want)
	}
}

func TestNeighborhood_PayloadExcludedFromWire(t *testing.T) {
	// Condition 1 (live-test seq=544): the Subgraph JSON is the
	// Wyrd-owned canonical wire shape. Assert the shard is a MAP —
	// no payload field in the marshaled output even when the source
	// node carries one — and the field names are the contract's.
	g := model.NewGraph()
	n := nbNode("a", "bma.seed", true, 1.0)
	n.Payload = []byte(`{"secret":"territory, not map"}`)
	if err := g.AddNode(n); err != nil {
		t.Fatal(err)
	}
	sg, err := New(g).Neighborhood("a", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(sg)
	if err != nil {
		t.Fatal(err)
	}
	var asMap map[string]any
	if err := json.Unmarshal(raw, &asMap); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"anchor", "depth", "truncated", "nodes", "edges"} {
		if _, ok := asMap[want]; !ok {
			t.Errorf("wire format missing contract field %q", want)
		}
	}
	if strings.Contains(string(raw), "payload") || strings.Contains(string(raw), "territory") {
		t.Error("payload leaked into the wire format — the shard must be a map, not the territory")
	}
}
