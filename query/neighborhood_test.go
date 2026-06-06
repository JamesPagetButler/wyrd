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
		nbNode("cert", "bma.lineage.life-certificate", true, 1.0),
		nbNode("seed1", "bma.seed", true, 1.0),
		nbNode("seed2", "bma.seed", true, 1.0),
		nbNode("obs1", "bma.observation", false, 0.3),
		nbNode("obs2", "bma.observation", false, 0.2),
		nbNode("obs3", "bma.observation", false, 0.1),
		nbNode("orphan", "bma.observation", false, 0.0),
	}
	for _, n := range nodes {
		if err := g.AddNode(n); err != nil {
			t.Fatalf("AddNode(%s): %v", n.ID, err)
		}
	}
	edges := []model.Hyperedge{
		nbEdge("e-cs1", "cert", "seed1"),
		nbEdge("e-cs2", "cert", "seed2"),
		nbEdge("e-s1o1", "seed1", "obs1"),
		nbEdge("e-s2o2", "seed2", "obs2"),
		nbEdge("e-o1o3", "obs1", "obs3"),
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
	sg, err := q.Neighborhood("cert", 2, 100)
	if err != nil {
		t.Fatalf("Neighborhood: %v", err)
	}
	// Depth 2 from cert: cert(0), seed1+seed2(1), obs1+obs2(2).
	// obs3 is 3 hops; orphan unreachable.
	wantIDs := []model.NodeID{"cert", "seed1", "seed2", "obs1", "obs2"}
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
	wantHops := map[model.NodeID]int{"cert": 0, "seed1": 1, "seed2": 1, "obs1": 2, "obs2": 2}
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
	a, err := q.Neighborhood("cert", 2, 100)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	b, err := q.Neighborhood("cert", 2, 100)
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

func TestNeighborhood_ExactFitBeyondRingSetsTruncated(t *testing.T) {
	q := New(fixtureGraph(t))
	// Budget 3 = cert + both seeds (ring 1 fits exactly); ring 2
	// (obs1, obs2) exists beyond → Truncated.
	sg, err := q.Neighborhood("cert", 2, 3)
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
	if _, err := q.Neighborhood("cert", 0, 100); err == nil {
		t.Error("depth 0 must be rejected")
	}
	if _, err := q.Neighborhood("cert", 2, 0); err == nil {
		t.Error("maxNodes 0 must be rejected")
	}
}

func TestNeighborhood_OrphanAnchor(t *testing.T) {
	q := New(fixtureGraph(t))
	sg, err := q.Neighborhood("orphan", 2, 100)
	if err != nil {
		t.Fatalf("Neighborhood: %v", err)
	}
	if len(sg.Nodes) != 1 || sg.Nodes[0].ID != "orphan" {
		t.Errorf("orphan shard = %v, want [orphan] only", sg.Nodes)
	}
	if len(sg.Edges) != 0 {
		t.Errorf("orphan shard has %d edges, want 0", len(sg.Edges))
	}
	if sg.Truncated {
		t.Error("orphan shard is complete — Truncated must be false")
	}
}

func TestNeighborhood_OrientationCarriedInWire(t *testing.T) {
	// Seam check vs bma-systema PR #244: the boot-time mirror's
	// provenance edges are ORIENTED (cert→seed "read-at-founding":
	// IsSymmetric=false, Heads=[0], Tails=[1]). The shard must carry
	// the arrows — a navigational map without direction loses the
	// provenance semantics the mirror exists to record.
	g := model.NewGraph()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(g.AddNode(nbNode("cert", "bma.lineage.life-certificate", true, 1.0)))
	must(g.AddNode(nbNode("seed1", "bma.seed", true, 1.0)))
	prov := model.Hyperedge{
		ID:          "prov-read-at-founding-seed1",
		Nodes:       []model.NodeID{"cert", "seed1"},
		Weight:      model.NewComplexWeight(1, 0),
		IsSymmetric: false,
		Created:     time.Unix(0, 0),
		Heads:       []int{0}, // source: cert
		Tails:       []int{1}, // sink: seed
	}
	must(g.AddHyperedge(prov))

	sg, err := New(g).Neighborhood("cert", 1, 10)
	if err != nil {
		t.Fatalf("Neighborhood: %v", err)
	}
	if len(sg.Edges) != 1 {
		t.Fatalf("len(edges) = %d, want 1", len(sg.Edges))
	}
	e := sg.Edges[0]
	if e.IsSymmetric {
		t.Error("oriented edge marked symmetric in shard")
	}
	if !reflect.DeepEqual(e.Heads, []int{0}) || !reflect.DeepEqual(e.Tails, []int{1}) {
		t.Errorf("orientation dropped: Heads=%v Tails=%v, want [0]/[1]", e.Heads, e.Tails)
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
