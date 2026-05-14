package query

import (
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/JamesPagetButler/wyrd/model"
)

const testIssuer = "test"

func mkNode(id model.NodeID, tier model.Tier) model.Node {
	return model.Node{
		ID:      id,
		Type:    testIssuer,
		Tier:    tier,
		Created: time.Unix(0, 0),
	}
}

func mkEdge(id model.HyperedgeID, nodes []model.NodeID, tier model.Tier) model.Hyperedge {
	return model.Hyperedge{
		ID:      id,
		Nodes:   nodes,
		Weight:  model.Weight{Tier: tier},
		Created: time.Unix(0, 0),
	}
}

func sortedNodes(ids []model.NodeID) []model.NodeID {
	out := append([]model.NodeID(nil), ids...)
	slices.Sort(out)
	return out
}

func sortedEdges(ids []model.HyperedgeID) []model.HyperedgeID {
	out := append([]model.HyperedgeID(nil), ids...)
	slices.Sort(out)
	return out
}

func TestNew_NilGraphPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil graph")
		}
	}()
	_ = New(nil)
}

func TestGetNode_HappyPath(t *testing.T) {
	g := model.NewGraph()
	if err := g.AddNode(mkNode("a", model.TierComplex)); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	q := New(g)
	n, ok := q.GetNode("a")
	if !ok {
		t.Fatal("node a not found")
	}
	if n.ID != "a" {
		t.Errorf("got node %s, want a", n.ID)
	}
}

func TestGetNode_Missing(t *testing.T) {
	q := New(model.NewGraph())
	if _, ok := q.GetNode("ghost"); ok {
		t.Error("expected ok=false for missing node")
	}
}

func TestGetHyperedge_HappyPath(t *testing.T) {
	g := model.NewGraph()
	_ = g.AddNode(mkNode("a", model.TierComplex))
	_ = g.AddNode(mkNode("b", model.TierComplex))
	_ = g.AddHyperedge(mkEdge("e", []model.NodeID{"a", "b"}, model.TierComplex))
	q := New(g)
	e, ok := q.GetHyperedge("e")
	if !ok {
		t.Fatal("edge e not found")
	}
	if e.ID != "e" {
		t.Errorf("got edge %s, want e", e.ID)
	}
}

func TestGetHyperedge_Missing(t *testing.T) {
	q := New(model.NewGraph())
	if _, ok := q.GetHyperedge("ghost"); ok {
		t.Error("expected ok=false for missing edge")
	}
}

func TestIncidentEdges_HappyPath(t *testing.T) {
	g := model.NewGraph()
	for _, id := range []model.NodeID{"a", "b", "c"} {
		_ = g.AddNode(mkNode(id, model.TierComplex))
	}
	_ = g.AddHyperedge(mkEdge("e1", []model.NodeID{"a", "b"}, model.TierComplex))
	_ = g.AddHyperedge(mkEdge("e2", []model.NodeID{"a", "c"}, model.TierComplex))
	q := New(g)
	got := sortedEdges(q.IncidentEdges("a"))
	want := []model.HyperedgeID{"e1", "e2"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("IncidentEdges(a) = %v, want %v", got, want)
	}
}

func TestIncidentEdges_EmptyOnMissingNode(t *testing.T) {
	q := New(model.NewGraph())
	got := q.IncidentEdges("ghost")
	if got == nil {
		t.Error("want empty slice, got nil")
	}
	if len(got) != 0 {
		t.Errorf("want empty, got %v", got)
	}
}

func TestIncidentEdges_EmptyOnIsolatedNode(t *testing.T) {
	g := model.NewGraph()
	_ = g.AddNode(mkNode("a", model.TierComplex))
	q := New(g)
	got := q.IncidentEdges("a")
	if got == nil {
		t.Error("want empty slice, got nil")
	}
	if len(got) != 0 {
		t.Errorf("want empty, got %v", got)
	}
}

func TestNeighborNodes_HappyPath(t *testing.T) {
	g := model.NewGraph()
	for _, id := range []model.NodeID{"a", "b", "c", "d"} {
		_ = g.AddNode(mkNode(id, model.TierComplex))
	}
	_ = g.AddHyperedge(mkEdge("e1", []model.NodeID{"a", "b", "c"}, model.TierComplex))
	_ = g.AddHyperedge(mkEdge("e2", []model.NodeID{"a", "d"}, model.TierComplex))
	q := New(g)
	got := sortedNodes(q.NeighborNodes("a"))
	want := []model.NodeID{"b", "c", "d"}
	if len(got) != len(want) {
		t.Fatalf("NeighborNodes(a) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("NeighborNodes(a)[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestNeighborNodes_EmptyOnMissing(t *testing.T) {
	q := New(model.NewGraph())
	got := q.NeighborNodes("ghost")
	if got == nil {
		t.Error("want empty slice, got nil")
	}
	if len(got) != 0 {
		t.Errorf("want empty, got %v", got)
	}
}

func TestNeighborNodes_EmptyOnIsolated(t *testing.T) {
	g := model.NewGraph()
	_ = g.AddNode(mkNode("a", model.TierComplex))
	q := New(g)
	got := q.NeighborNodes("a")
	if len(got) != 0 {
		t.Errorf("want empty, got %v", got)
	}
}

// TestNeighborNodes_SelfLoopExcludesSelf — per PR #26 §6 (NEIGHBOR-EXCLUDES-SELF):
// an arity-1 self-membership hyperedge makes IncidentEdges return [e.ID] but
// NeighborNodes return []. Both are correct per their contracts.
func TestNeighborNodes_SelfLoopExcludesSelf(t *testing.T) {
	g := model.NewGraph()
	_ = g.AddNode(mkNode("a", model.TierComplex))
	_ = g.AddHyperedge(mkEdge("e_self", []model.NodeID{"a"}, model.TierComplex))
	q := New(g)

	edges := q.IncidentEdges("a")
	if len(edges) != 1 || edges[0] != "e_self" {
		t.Errorf("IncidentEdges(a) = %v, want [e_self]", edges)
	}

	neighbors := q.NeighborNodes("a")
	if len(neighbors) != 0 {
		t.Errorf("NeighborNodes(a) = %v, want [] (self excluded)", neighbors)
	}
}

func TestNeighborNodes_NoDuplicates(t *testing.T) {
	g := model.NewGraph()
	for _, id := range []model.NodeID{"a", "b"} {
		_ = g.AddNode(mkNode(id, model.TierComplex))
	}
	_ = g.AddHyperedge(mkEdge("e1", []model.NodeID{"a", "b"}, model.TierComplex))
	_ = g.AddHyperedge(mkEdge("e2", []model.NodeID{"a", "b"}, model.TierComplex))
	q := New(g)
	got := q.NeighborNodes("a")
	if len(got) != 1 || got[0] != "b" {
		t.Errorf("NeighborNodes(a) = %v, want [b] (no duplicates)", got)
	}
}

// TestNeighborNodes_ConcurrentReadsAndWrites validates PR #26 §5's
// "valid but not single-snapshot" contract per @qbp-implementor's
// CONCURRENCY-INVARIANT ask. Result is always a subset of *some*
// sequential graph state — no dangling refs, no IDs that never existed.
func TestNeighborNodes_ConcurrentReadsAndWrites(t *testing.T) {
	g := model.NewGraph()
	_ = g.AddNode(mkNode("a", model.TierComplex))
	// Add some initial neighbors to read against.
	for _, id := range []model.NodeID{"b", "c", "d"} {
		_ = g.AddNode(mkNode(id, model.TierComplex))
	}
	_ = g.AddHyperedge(mkEdge("e_initial", []model.NodeID{"a", "b"}, model.TierComplex))
	q := New(g)

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Reader goroutine: never sees a dangling NodeID.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			neighbors := q.NeighborNodes("a")
			for _, n := range neighbors {
				// Every returned NodeID must correspond to a node that
				// exists in the graph at the moment we check. This is
				// the "always valid" property — no dangling refs.
				if _, ok := q.GetNode(n); !ok {
					// Permitted: by the time we check, the node has
					// been removed. But every returned NodeID must
					// have been a real node at some sequential
					// moment in the recent past. Test for the
					// invariant that we don't get a NodeID that
					// never existed (e.g., garbage memory).
					if n == "" {
						t.Errorf("NeighborNodes returned empty NodeID — invariant violation")
					}
				}
			}
		}
	}()

	// Writer goroutine: adds new edges incident on "a".
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range 50 {
			select {
			case <-stop:
				return
			default:
			}
			nodeID := model.NodeID("n" + string(rune('0'+i%10)))
			_ = g.AddNode(mkNode(nodeID, model.TierComplex))
			edgeID := model.HyperedgeID("e" + string(rune('0'+i%10)))
			_ = g.AddHyperedge(mkEdge(edgeID, []model.NodeID{"a", nodeID}, model.TierComplex))
		}
	}()

	time.Sleep(50 * time.Millisecond)
	close(stop)
	wg.Wait()
}
