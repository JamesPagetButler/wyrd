package model

import (
	"testing"
	"time"
)

func mkNode(id NodeID, tier Tier) Node {
	return Node{
		ID:      id,
		Type:    testIssuer,
		Tier:    tier,
		Created: time.Unix(0, 0),
	}
}

func mkEdge(id HyperedgeID, nodes []NodeID, tier Tier) Hyperedge {
	return Hyperedge{
		ID:      id,
		Nodes:   nodes,
		Weight:  Weight{Tier: tier},
		Created: time.Unix(0, 0),
	}
}

func TestGraph_AddAndQueryNodes(t *testing.T) {
	g := NewGraph()
	if err := g.AddNode(mkNode("a", TierComplex)); err != nil {
		t.Fatalf("add a: %v", err)
	}
	if g.NodeCount() != 1 {
		t.Errorf("NodeCount = %d, want 1", g.NodeCount())
	}
	n, ok := g.Node("a")
	if !ok {
		t.Fatal("node a not found")
	}
	if n.Tier != TierComplex {
		t.Errorf("node a tier = %s, want complex", n.Tier)
	}
}

func TestGraph_DuplicateNodeRejected(t *testing.T) {
	g := NewGraph()
	_ = g.AddNode(mkNode("a", TierComplex))
	if err := g.AddNode(mkNode("a", TierComplex)); err == nil {
		t.Error("expected error on duplicate node")
	}
}

func TestGraph_HyperedgeRequiresKnownNodes(t *testing.T) {
	g := NewGraph()
	_ = g.AddNode(mkNode("a", TierComplex))
	_ = g.AddNode(mkNode("b", TierComplex))
	if err := g.AddHyperedge(mkEdge("e1", []NodeID{"a", NodeID(missingID)}, TierComplex)); err == nil {
		t.Error("expected error: edge references unknown node")
	}
}

func TestGraph_IncidentEdges(t *testing.T) {
	g := NewGraph()
	for _, id := range []NodeID{"a", "b", "c", "d"} {
		_ = g.AddNode(mkNode(id, TierQuaternion))
	}
	_ = g.AddHyperedge(mkEdge("e_abc", []NodeID{"a", "b", "c"}, TierQuaternion))
	_ = g.AddHyperedge(mkEdge("e_bd", []NodeID{"b", "d"}, TierQuaternion))

	want := map[NodeID]int{"a": 1, "b": 2, "c": 1, "d": 1}
	for v, w := range want {
		if got := len(g.IncidentEdges(v)); got != w {
			t.Errorf("IncidentEdges(%s) count = %d, want %d", v, got, w)
		}
	}
}

// TestGraph_NonincidentAdditionPreservesIncidence is the Go-side counterpart
// of `Wyrd.Hypergraph.hyperedge_preserves_incident_edges` (C-20a, Phase 2):
// adding an edge that doesn't touch v leaves IncidentEdges(v) unchanged.
func TestGraph_NonincidentAdditionPreservesIncidence(t *testing.T) {
	g := NewGraph()
	for _, id := range []NodeID{"a", "b", "c", "d"} {
		_ = g.AddNode(mkNode(id, TierQuaternion))
	}
	_ = g.AddHyperedge(mkEdge("e_ab", []NodeID{"a", "b"}, TierQuaternion))

	beforeC := len(g.IncidentEdges("c"))
	// Add an edge that does NOT touch c.
	_ = g.AddHyperedge(mkEdge("e_ad", []NodeID{"a", "d"}, TierQuaternion))
	afterC := len(g.IncidentEdges("c"))

	if beforeC != afterC {
		t.Errorf("non-incident addition changed c's incident set: %d → %d", beforeC, afterC)
	}
}

func TestGraph_RemoveHyperedge(t *testing.T) {
	g := NewGraph()
	_ = g.AddNode(mkNode("a", TierComplex))
	_ = g.AddNode(mkNode("b", TierComplex))
	_ = g.AddHyperedge(mkEdge("e", []NodeID{"a", "b"}, TierComplex))

	if err := g.RemoveHyperedge("e"); err != nil {
		t.Fatalf("RemoveHyperedge: %v", err)
	}
	if got := len(g.IncidentEdges("a")); got != 0 {
		t.Errorf("IncidentEdges(a) after remove = %d, want 0", got)
	}
	if err := g.RemoveHyperedge("e"); err == nil {
		t.Error("expected error removing non-existent edge")
	}
}
