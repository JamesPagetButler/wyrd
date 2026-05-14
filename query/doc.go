// Package query is the read-only API surface over a model.Graph.
//
// Constructed with [New]; never mutates the underlying graph. Safe for
// concurrent use because all underlying Graph reads go through
// model.Graph's RWMutex (PR #14 + ADR-003 §I3).
//
// v0.1 ships four methods covering the load-bearing read patterns:
// GetNode, GetHyperedge, IncidentEdges, NeighborNodes. The DSL is
// deferred per James's Q1 decision (live-test seq=65 option A).
//
// Soundness: per Wyrd.Hypergraph.hyperedge_preserves_incident_edges
// (Phase 2 C-20a), IncidentEdges respects the structural invariant
// that adding a non-incident hyperedge leaves a node's incident set
// unchanged. Per Wyrd.Projection.kernel_supervisor_safe (Phase 1
// T2.2), reads at any tier are safe regardless of caller tier — no
// ReadCapability required for query operations (consistent with the
// capability v0.2 design § 4 Option A).
//
// Directionality (PR #26 §2.1, §I4-approved): IncidentEdges returns
// flattened membership. model.Hyperedge.Nodes is []NodeID with no
// orientation metadata, so any "oriented" answer at v0.1 would be a
// fiction inferred from slice index. Oriented traversal pairs with a
// future oriented-hyperedge model extension (Wyrd issue #30) as a
// separate primitive.
package query
