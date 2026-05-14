package model

import (
	"fmt"
	"time"
)

// NodeID uniquely identifies a hypergraph node within a Wyrd Graph.
// IDs are opaque strings; consumers (CTH, BMA, Contextus) choose their
// own naming conventions. We recommend a `<scope>:<local-id>` form
// (e.g., `cth:anchor:T2-mass-prediction`, `bma:engram:abc123`).
type NodeID string

// NodeType is a free-form classification chosen by the consumer.
// Reserved Wyrd-internal types are prefixed with `wyrd.` and MUST NOT
// be used by external consumers.
type NodeType string

// Reserved Wyrd-internal node types.
const (
	NodeTypeWyrdInternal NodeType = "wyrd.internal"
)

// Node is a vertex in the Wyrd hypergraph.
//
// Soundness: nodes carry a Tier indicating the algebraic tier at which
// operations involving them must execute. A process operating below the
// node's tier may read it (downward projection is safe per
// `Wyrd.Projection.kernel_supervisor_safe`) but cannot author edges
// referencing it without a capability of at least that tier (per
// `Wyrd.Capability.capability_grants_safe_access`).
type Node struct {
	ID      NodeID    `json:"id"`
	Type    NodeType  `json:"type"`
	Tier    Tier      `json:"tier"`
	Created time.Time `json:"created"`
	// Payload is opaque to Wyrd; consumers serialise their own data.
	Payload []byte `json:"payload,omitempty"`

	// TierImmune marks the node as exempt from all automatic eviction
	// paths (cap-per-tier saturation, sleep-cycle compaction, etc.).
	// Used for NT_SEED (BMA seed protocol Step 9), foundation theorems,
	// and any node whose deletion would invalidate downstream invariants.
	// Default false preserves v0.1 wire-format compatibility.
	//
	// TierImmune blocks EVICTION (automatic, policy-driven) but NOT
	// explicit deletion via Graph.RemoveNodeWithCapability — that's a
	// user mutation through the capability layer.
	//
	// Soundness: per (forthcoming)
	// Wyrd.Hypergraph.tier_immune_node_preserves_eviction (W-Toddle-1),
	// adding or evicting other nodes does not change the membership of
	// {v : v.TierImmune}.
	TierImmune bool `json:"tier_immune,omitempty"`

	// Salience modulates eviction priority. Range 0.0..1.0.
	//   0.0 (default): no priority modulation
	//   higher values: stronger retention under pressure
	// Hebbian reinforcement increments Salience (capped at 1.0);
	// Ebbinghaus decay decrements it over time. Default 0.0 preserves
	// v0.1 wire-format compatibility.
	//
	// Eviction priority order under saturation: TierImmune nodes
	// excluded; among the remainder, ascending Salience evicted first.
	Salience float64 `json:"salience,omitempty"`
}

// Validate returns an error if any required field is missing or invalid.
func (n Node) Validate() error {
	if n.ID == "" {
		return fmt.Errorf("model: node: empty id")
	}
	if n.Type == "" {
		return fmt.Errorf("model: node %s: empty type", n.ID)
	}
	if !n.Tier.IsValid() {
		return fmt.Errorf("model: node %s: invalid tier %v", n.ID, n.Tier)
	}
	if n.Salience < 0.0 || n.Salience > 1.0 {
		return fmt.Errorf("model: node %s: salience %g out of [0,1]", n.ID, n.Salience)
	}
	return nil
}
