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
	return nil
}
