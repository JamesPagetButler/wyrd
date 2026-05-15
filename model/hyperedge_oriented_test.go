package model

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func mkOrientedNode(id NodeID) Node {
	return Node{ID: id, Type: testIssuer, Tier: TierComplex, Created: time.Unix(0, 0)}
}

// TestHyperedge_Validate_SymmetricWithHeadsRejected confirms that
// an IsSymmetric=true edge with non-empty Heads/Tails fails validation.
func TestHyperedge_Validate_SymmetricWithHeadsRejected(t *testing.T) {
	e := Hyperedge{
		ID:          "e",
		Nodes:       []NodeID{"a", "b"},
		Weight:      Weight{Tier: TierComplex},
		IsSymmetric: true,
		Heads:       []int{0},
		Created:     time.Unix(0, 0),
	}
	err := e.Validate()
	if err == nil || !strings.Contains(err.Error(), "symmetric edge must have empty Heads/Tails") {
		t.Errorf("expected symmetric+Heads rejection; got %v", err)
	}
}

func TestHyperedge_Validate_HeadIndexOutOfRange(t *testing.T) {
	e := Hyperedge{
		ID:      "e",
		Nodes:   []NodeID{"a", "b"},
		Weight:  Weight{Tier: TierComplex},
		Heads:   []int{2},
		Created: time.Unix(0, 0),
	}
	if err := e.Validate(); err == nil {
		t.Error("expected out-of-range head index rejection")
	}
}

func TestHyperedge_Validate_NegativeHeadIndex(t *testing.T) {
	e := Hyperedge{
		ID:      "e",
		Nodes:   []NodeID{"a", "b"},
		Weight:  Weight{Tier: TierComplex},
		Heads:   []int{-1},
		Created: time.Unix(0, 0),
	}
	if err := e.Validate(); err == nil {
		t.Error("expected negative head index rejection")
	}
}

// TestHyperedge_Validate_HeadAndTailOverlap confirms a node can't be
// both head and tail.
func TestHyperedge_Validate_HeadAndTailOverlap(t *testing.T) {
	e := Hyperedge{
		ID:      "e",
		Nodes:   []NodeID{"a", "b"},
		Weight:  Weight{Tier: TierComplex},
		Heads:   []int{0},
		Tails:   []int{0},
		Created: time.Unix(0, 0),
	}
	err := e.Validate()
	if err == nil || !strings.Contains(err.Error(), "in both") {
		t.Errorf("expected head/tail overlap rejection; got %v", err)
	}
}

func TestHyperedge_Validate_DuplicateInHeads(t *testing.T) {
	e := Hyperedge{
		ID:      "e",
		Nodes:   []NodeID{"a", "b"},
		Weight:  Weight{Tier: TierComplex},
		Heads:   []int{0, 0},
		Created: time.Unix(0, 0),
	}
	if err := e.Validate(); err == nil {
		t.Error("expected duplicate-in-Heads rejection")
	}
}

// TestHyperedge_Validate_OrientedHappyPath — a clean bipartite edge.
func TestHyperedge_Validate_OrientedBipartite(t *testing.T) {
	e := Hyperedge{
		ID:      "e",
		Nodes:   []NodeID{"src", "sink"},
		Weight:  Weight{Tier: TierComplex},
		Heads:   []int{0},
		Tails:   []int{1},
		Created: time.Unix(0, 0),
	}
	if err := e.Validate(); err != nil {
		t.Errorf("clean bipartite should validate; got %v", err)
	}
}

// TestHyperedge_Validate_TransitNodes — PR #31 §3 N-to-M-with-transit.
func TestHyperedge_Validate_TransitNodes(t *testing.T) {
	e := Hyperedge{
		ID:      "e",
		Nodes:   []NodeID{"src", "ctx", "sink"},
		Weight:  Weight{Tier: TierComplex},
		Heads:   []int{0},
		Tails:   []int{2},
		Created: time.Unix(0, 0),
	}
	if err := e.Validate(); err != nil {
		t.Errorf("transit-node pattern should validate; got %v", err)
	}
}

func TestHyperedge_IsOriented(t *testing.T) {
	cases := []struct {
		name string
		e    Hyperedge
		want bool
	}{
		{"empty heads + tails", Hyperedge{Nodes: []NodeID{"a", "b"}}, false},
		{"heads only", Hyperedge{Nodes: []NodeID{"a", "b"}, Heads: []int{0}}, true},
		{"tails only", Hyperedge{Nodes: []NodeID{"a", "b"}, Tails: []int{1}}, true},
		{"both", Hyperedge{Nodes: []NodeID{"a", "b"}, Heads: []int{0}, Tails: []int{1}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.e.IsOriented(); got != tc.want {
				t.Errorf("IsOriented = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHyperedge_HeadNodes_TailNodes_TransitNodes(t *testing.T) {
	e := Hyperedge{
		ID:      "e",
		Nodes:   []NodeID{"a", "b", "c", "d"},
		Heads:   []int{0, 2},
		Tails:   []int{1},
		Created: time.Unix(0, 0),
	}
	if got := e.HeadNodes(); len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Errorf("HeadNodes = %v, want [a c]", got)
	}
	if got := e.TailNodes(); len(got) != 1 || got[0] != "b" {
		t.Errorf("TailNodes = %v, want [b]", got)
	}
	if got := e.TransitNodes(); len(got) != 1 || got[0] != "d" {
		t.Errorf("TransitNodes = %v, want [d]", got)
	}
}

// TestHyperedge_JSON_v01CompatibleBothDirections confirms backward
// compatibility per PR #31 §6:
//   - v0.1 JSON (no heads/tails fields) deserialises into v0.2
//   - v0.2 with empty heads/tails serialises to v0.1-indistinguishable form
func TestHyperedge_JSON_v01CompatibleBothDirections(t *testing.T) {
	v01JSON := `{"id":"e","nodes":["a","b"],"weight":{"components":[1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"tier":"complex"},"is_symmetric":false,"created":"1970-01-01T00:00:00Z"}`
	var e Hyperedge
	if err := json.Unmarshal([]byte(v01JSON), &e); err != nil {
		t.Fatalf("unmarshal v0.1: %v", err)
	}
	if e.IsOriented() {
		t.Error("v0.1 edge should not be oriented")
	}
	if len(e.Heads) != 0 || len(e.Tails) != 0 {
		t.Errorf("v0.1 edge should have empty Heads+Tails; got %v / %v", e.Heads, e.Tails)
	}

	// Now serialise + check the omitempty.
	out, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(out)
	if strings.Contains(s, "heads") || strings.Contains(s, "tails") {
		t.Errorf("default heads/tails should be omitted via omitempty; got %s", s)
	}
}

func TestHyperedge_JSON_OrientedRoundTrip(t *testing.T) {
	in := Hyperedge{
		ID:      "e",
		Nodes:   []NodeID{"a", "b", "c"},
		Weight:  NewQuaternionWeight(1, 0, 0, 0),
		Heads:   []int{0},
		Tails:   []int{2},
		Created: time.Unix(0, 0).UTC(),
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"heads":[0]`) {
		t.Errorf("Heads not in JSON: %s", b)
	}
	if !strings.Contains(string(b), `"tails":[2]`) {
		t.Errorf("Tails not in JSON: %s", b)
	}
	var out Hyperedge
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !out.IsOriented() {
		t.Error("round-tripped edge should be oriented")
	}
}

// TestGraph_AddHyperedge_AcceptsOrientedEdge — confirms the model.Graph
// path accepts oriented edges end-to-end.
func TestGraph_AddHyperedge_AcceptsOrientedEdge(t *testing.T) {
	g := NewGraph()
	for _, id := range []NodeID{"src", "sink"} {
		_ = g.AddNode(mkOrientedNode(id))
	}
	e := Hyperedge{
		ID:      "e",
		Nodes:   []NodeID{"src", "sink"},
		Weight:  NewQuaternionWeight(1, 0, 0, 0),
		Heads:   []int{0},
		Tails:   []int{1},
		Created: time.Unix(0, 0),
	}
	if err := g.AddHyperedge(e); err != nil {
		t.Fatalf("AddHyperedge: %v", err)
	}
	got, ok := g.Hyperedge("e")
	if !ok {
		t.Fatal("edge not found")
	}
	if got.HeadNodes()[0] != "src" || got.TailNodes()[0] != "sink" {
		t.Errorf("orientation lost in graph round-trip")
	}
}
