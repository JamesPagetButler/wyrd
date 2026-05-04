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
	return nil
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
