package query

import (
	"slices"
	"testing"
	"time"

	"github.com/JamesPagetButler/wyrd/model"
)

// makeOrientedTestGraph constructs a graph with one outgoing, one
// incoming, one transit, and one symmetric edge incident on node "v".
func makeOrientedTestGraph(t *testing.T) *model.Graph {
	t.Helper()
	g := model.NewGraph()
	for _, id := range []model.NodeID{"v", "a", "b", "c", "d", "e", "f"} {
		if err := g.AddNode(model.Node{
			ID:      id,
			Type:    testIssuer,
			Tier:    model.TierComplex,
			Created: time.Unix(0, 0),
		}); err != nil {
			t.Fatalf("add node %s: %v", id, err)
		}
	}
	// e_out: v is head (outgoing); a is tail.
	_ = g.AddHyperedge(model.Hyperedge{
		ID:      "e_out",
		Nodes:   []model.NodeID{"v", "a"},
		Weight:  model.NewQuaternionWeight(1, 0, 0, 0),
		Heads:   []int{0},
		Tails:   []int{1},
		Created: time.Unix(0, 0),
	})
	// e_in: b is head; v is tail (incoming).
	_ = g.AddHyperedge(model.Hyperedge{
		ID:      "e_in",
		Nodes:   []model.NodeID{"b", "v"},
		Weight:  model.NewQuaternionWeight(1, 0, 0, 0),
		Heads:   []int{0},
		Tails:   []int{1},
		Created: time.Unix(0, 0),
	})
	// e_trans: c is head; d is tail; v is transit.
	_ = g.AddHyperedge(model.Hyperedge{
		ID:      "e_trans",
		Nodes:   []model.NodeID{"c", "v", "d"},
		Weight:  model.NewQuaternionWeight(1, 0, 0, 0),
		Heads:   []int{0},
		Tails:   []int{2},
		Created: time.Unix(0, 0),
	})
	// e_sym: not oriented; v + e + f symmetric.
	_ = g.AddHyperedge(model.Hyperedge{
		ID:          "e_sym",
		Nodes:       []model.NodeID{"v", "e", "f"},
		Weight:      model.NewQuaternionWeight(1, 0, 0, 0),
		IsSymmetric: true,
		Created:     time.Unix(0, 0),
	})
	return g
}

func TestIncidentOrientedEdges_AllFourBuckets(t *testing.T) {
	g := makeOrientedTestGraph(t)
	q := New(g)
	got := q.IncidentOrientedEdges("v")

	if len(got.Outgoing) != 1 || got.Outgoing[0] != "e_out" {
		t.Errorf("Outgoing = %v, want [e_out]", got.Outgoing)
	}
	if len(got.Incoming) != 1 || got.Incoming[0] != "e_in" {
		t.Errorf("Incoming = %v, want [e_in]", got.Incoming)
	}
	if len(got.Transit) != 1 || got.Transit[0] != "e_trans" {
		t.Errorf("Transit = %v, want [e_trans]", got.Transit)
	}
	if len(got.Symmetric) != 1 || got.Symmetric[0] != "e_sym" {
		t.Errorf("Symmetric = %v, want [e_sym]", got.Symmetric)
	}
}

// TestIncidentOrientedEdges_PartitionEqualsIncidentEdges verifies the
// PR #31 §2.1 invariant: union of all four buckets = IncidentEdges(v).
func TestIncidentOrientedEdges_PartitionEqualsIncidentEdges(t *testing.T) {
	g := makeOrientedTestGraph(t)
	q := New(g)
	parts := q.IncidentOrientedEdges("v")
	flat := q.IncidentEdges("v")

	all := append([]model.HyperedgeID{}, parts.Outgoing...)
	all = append(all, parts.Incoming...)
	all = append(all, parts.Transit...)
	all = append(all, parts.Symmetric...)

	if len(all) != len(flat) {
		t.Fatalf("partition len = %d, IncidentEdges len = %d", len(all), len(flat))
	}
	slices.Sort(all)
	slices.Sort(flat)
	for i := range all {
		if all[i] != flat[i] {
			t.Errorf("partition[%d] = %s, IncidentEdges[%d] = %s", i, all[i], i, flat[i])
		}
	}
}

func TestIncidentOrientedEdges_EmptyForUnknownNode(t *testing.T) {
	g := model.NewGraph()
	q := New(g)
	got := q.IncidentOrientedEdges("ghost")
	if got.Outgoing == nil || got.Incoming == nil || got.Transit == nil || got.Symmetric == nil {
		t.Error("buckets should be empty slices (not nil) for unknown node")
	}
	if len(got.Outgoing)+len(got.Incoming)+len(got.Transit)+len(got.Symmetric) != 0 {
		t.Errorf("all buckets should be empty for unknown node; got %+v", got)
	}
}

// TestIncidentOrientedEdges_BackwardCompatible — an edge with no
// orientation metadata at all goes into Symmetric regardless of
// IsSymmetric flag value. (v0.1 backward-compat: edges that pre-date
// the orientation extension have IsOriented()==false.)
func TestIncidentOrientedEdges_BackwardCompatible(t *testing.T) {
	g := model.NewGraph()
	_ = g.AddNode(model.Node{ID: "v", Type: testIssuer, Tier: model.TierComplex, Created: time.Unix(0, 0)})
	_ = g.AddNode(model.Node{ID: "a", Type: testIssuer, Tier: model.TierComplex, Created: time.Unix(0, 0)})
	// IsSymmetric=false but no Heads/Tails — the v0.1-style edge.
	_ = g.AddHyperedge(model.Hyperedge{
		ID:      "e",
		Nodes:   []model.NodeID{"v", "a"},
		Weight:  model.NewQuaternionWeight(1, 0, 0, 0),
		Created: time.Unix(0, 0),
	})
	q := New(g)
	got := q.IncidentOrientedEdges("v")
	if len(got.Symmetric) != 1 || got.Symmetric[0] != "e" {
		t.Errorf("v0.1-style edge should land in Symmetric; got %+v", got)
	}
}
