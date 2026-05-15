package model

import (
	"fmt"
	"time"
)

// HyperedgeID uniquely identifies a hyperedge within a Wyrd Graph.
type HyperedgeID string

// Hyperedge connects k ≥ 1 nodes with a single tier-tagged weight.
//
// Soundness — irreducibility: per `Wyrd.HolographicHypergraph.theorem2_irreducibility`
// (Phase 4 v1.4) and the higher-arity generalisation
// `Wyrd.HolographicHypergraphHigherArity.theorem2_irreducibility_n_arity`
// (Phase 4 v1.5), a hyperedge of arity k ≥ 3 carries information that
// CANNOT be recovered from any decomposition into smaller-arity edges.
// Consumers MUST NOT silently split a k-edge into k(k−1)/2 pair edges
// at storage or transit time — the joint constraint is information that
// pair edges cannot encode. Use [Graph.AddHyperedge], not pair shims.
//
// The Nodes slice is ordered. Edges may be directional (the consumer
// may interpret position 0 as source, position N-1 as sink, etc.) or
// symmetric; Wyrd does not impose a convention. The IsSymmetric flag
// records the convention so that incidence queries can be specialised.
type Hyperedge struct {
	ID          HyperedgeID `json:"id"`
	Nodes       []NodeID    `json:"nodes"`
	Weight      Weight      `json:"weight"`
	IsSymmetric bool        `json:"is_symmetric"`
	Created     time.Time   `json:"created"`

	// Heads and Tails are indices into Nodes encoding orientation. When
	// IsSymmetric == false, Heads identifies the source-side nodes
	// and Tails the sink-side nodes; their semantics are tenant-
	// defined (CTH opcode flow: Heads = upstream, Tails = downstream;
	// other consumers may interpret differently per their own design).
	// Indices in Heads ∪ Tails must lie within [0, len(Nodes)); their
	// intersection must be empty. Nodes not in either set are
	// "transit" — present in the edge but not directional endpoints
	// (per PR #31 §3 "N-to-M-with-transit" pattern).
	//
	// When IsSymmetric == true, both slices must be empty (validated).
	// Both fields omitempty for v0.1 wire-format compatibility.
	//
	// Soundness: per Wyrd.HypergraphOriented.oriented_edge_preserves
	// _incident_edges (Phase 2 extension, forthcoming Lean theorem
	// following PR #31 §4.3 reduction pattern), adding an oriented
	// edge whose Nodes set excludes v leaves IncidentEdges(v)
	// unchanged regardless of orientation metadata.
	Heads []int `json:"heads,omitempty"`
	Tails []int `json:"tails,omitempty"`
}

// Arity returns the number of distinct nodes the hyperedge connects.
// (The Nodes slice may contain a node multiple times for self-loops;
// arity counts the multiset length.)
func (e Hyperedge) Arity() int {
	return len(e.Nodes)
}

// Tier returns the tier of the edge's weight; this is the operational
// tier of any computation involving the edge.
func (e Hyperedge) Tier() Tier {
	return e.Weight.Tier
}

// Validate returns an error if the edge is malformed.
func (e Hyperedge) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("model: hyperedge: empty id")
	}
	if len(e.Nodes) == 0 {
		return fmt.Errorf("model: hyperedge %s: empty node list", e.ID)
	}
	for i, n := range e.Nodes {
		if n == "" {
			return fmt.Errorf("model: hyperedge %s: node[%d] empty", e.ID, i)
		}
	}
	if err := e.Weight.Validate(); err != nil {
		return fmt.Errorf("model: hyperedge %s: %w", e.ID, err)
	}
	// Orientation invariants per PR #31 §3.
	if e.IsSymmetric && (len(e.Heads) > 0 || len(e.Tails) > 0) {
		return fmt.Errorf("model: hyperedge %s: symmetric edge must have empty Heads/Tails", e.ID)
	}
	n := len(e.Nodes)
	seen := make(map[int]string, len(e.Heads)+len(e.Tails))
	for _, idx := range e.Heads {
		if idx < 0 || idx >= n {
			return fmt.Errorf("model: hyperedge %s: Heads index %d out of range [0,%d)", e.ID, idx, n)
		}
		if prev, ok := seen[idx]; ok {
			return fmt.Errorf("model: hyperedge %s: index %d in both %s and Heads", e.ID, idx, prev)
		}
		seen[idx] = "Heads"
	}
	for _, idx := range e.Tails {
		if idx < 0 || idx >= n {
			return fmt.Errorf("model: hyperedge %s: Tails index %d out of range [0,%d)", e.ID, idx, n)
		}
		if prev, ok := seen[idx]; ok {
			return fmt.Errorf("model: hyperedge %s: index %d in both %s and Tails", e.ID, idx, prev)
		}
		seen[idx] = "Tails"
	}
	return nil
}

// IsOriented reports whether the edge carries non-trivial orientation
// metadata (at least one Head or Tail index set). An edge can be
// IsSymmetric=false without being oriented if both Heads and Tails
// are empty — that's the "v0.1 backward-compatible" case where
// orientation isn't expressed even though it isn't structurally
// forbidden either.
func (e Hyperedge) IsOriented() bool {
	return len(e.Heads) > 0 || len(e.Tails) > 0
}

// HeadNodes returns the NodeIDs of the head-side endpoints. Convenience
// wrapper over the Heads index slice.
func (e Hyperedge) HeadNodes() []NodeID {
	if len(e.Heads) == 0 {
		return nil
	}
	out := make([]NodeID, 0, len(e.Heads))
	for _, idx := range e.Heads {
		if idx >= 0 && idx < len(e.Nodes) {
			out = append(out, e.Nodes[idx])
		}
	}
	return out
}

// TailNodes returns the NodeIDs of the tail-side endpoints.
func (e Hyperedge) TailNodes() []NodeID {
	if len(e.Tails) == 0 {
		return nil
	}
	out := make([]NodeID, 0, len(e.Tails))
	for _, idx := range e.Tails {
		if idx >= 0 && idx < len(e.Nodes) {
			out = append(out, e.Nodes[idx])
		}
	}
	return out
}

// TransitNodes returns the NodeIDs of nodes that are part of the edge
// but neither head nor tail endpoints. Per PR #31 §3, transit nodes
// participate as context.
func (e Hyperedge) TransitNodes() []NodeID {
	if !e.IsOriented() {
		return nil
	}
	roles := make(map[int]bool, len(e.Heads)+len(e.Tails))
	for _, i := range e.Heads {
		roles[i] = true
	}
	for _, i := range e.Tails {
		roles[i] = true
	}
	var out []NodeID
	for i, n := range e.Nodes {
		if !roles[i] {
			out = append(out, n)
		}
	}
	return out
}

// Incident reports whether v is among the edge's nodes.
func (e Hyperedge) Incident(v NodeID) bool {
	for _, n := range e.Nodes {
		if n == v {
			return true
		}
	}
	return false
}
