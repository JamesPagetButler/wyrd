package model

import (
	"errors"
	"fmt"
	"unsafe"
)

// ErrBatchEdgeNotFound is returned when a PromoteBatch / RemoveBatch
// preflight check finds an id that does not exist in the source graph.
var ErrBatchEdgeNotFound = errors.New("model: batch: hyperedge not found in source")

// ErrBatchEdgeAlreadyExists is returned when a PromoteBatch preflight
// finds an id that already exists in the destination graph.
var ErrBatchEdgeAlreadyExists = errors.New("model: batch: hyperedge already exists in destination")

// ErrBatchMissingNode is returned when a PromoteBatch preflight finds
// an edge whose nodes are not all present in the destination graph.
// Consumers (e.g., BMA's sleep-cycle compactor) are responsible for
// mirroring nodes into the destination before calling PromoteBatch.
var ErrBatchMissingNode = errors.New("model: batch: edge references node not in destination")

// PromoteBatch atomically moves the named hyperedges from g (source)
// to dst (destination). Either all listed edges land in dst and are
// removed from g, or none do — on any failure both graphs are restored
// to their pre-call state.
//
// All-or-nothing per ADR-003 §I3 (atomicity is an S-01 requirement
// per @bma `live-test` seq=22; partial-failure-with-manifest was
// considered and explicitly closed). Lock-ordered by pointer address
// to avoid two-graph deadlock with reverse-direction concurrent batches.
//
// Soundness:
//   - per-edge: same as [Bridge.Promote] (compute pkg), citing
//     `Wyrd.Bridge.bridge_promote_preserves_count` (Phase 2 C-20c).
//   - batch-level: a forthcoming Lean theorem
//     `Wyrd.Bridge.bridge_promote_batch_preserves_count` (Phase 4
//     v1.6, induction over the batch list) lifts C-20c to the batch
//     case. The induction is small — the inductive step is one
//     application of C-20c plus the trivial sum-conservation identity.
//   - I3 atomicity: the write Lock on both graphs is held continuously
//     across the whole batch, satisfying the "observer out for the
//     full duration of any structural action" requirement (ADR-003
//     §I3).
//
// Caller responsibility: every node referenced by an edge in ids
// MUST already exist in dst before this call. The preflight returns
// [ErrBatchMissingNode] if this is violated; no mutation occurs.
//
// Errors: [ErrBatchEdgeNotFound], [ErrBatchEdgeAlreadyExists],
// [ErrBatchMissingNode]; first failure encountered fails the whole
// batch with rollback (which in v0.1 is a no-op since preflight
// completes before any mutation).
func (g *Graph) PromoteBatch(dst *Graph, ids []HyperedgeID) error {
	if g == nil {
		return fmt.Errorf("model: graph: nil source")
	}
	if dst == nil {
		return fmt.Errorf("model: graph: nil destination")
	}
	if g == dst {
		return fmt.Errorf("model: graph: source and destination are the same graph")
	}

	first, second := orderLocks(g, dst)
	first.mu.Lock()
	defer first.mu.Unlock()
	second.mu.Lock()
	defer second.mu.Unlock()

	// Phase 1 (preflight, no mutation): validate every id can be
	// promoted. Catch missing-in-source, already-in-destination, and
	// missing-node-in-destination errors here; fail the batch before
	// any state changes.
	edges := make([]Hyperedge, 0, len(ids))
	for _, id := range ids {
		e, ok := g.edges[id]
		if !ok {
			return fmt.Errorf("%w: %s", ErrBatchEdgeNotFound, id)
		}
		if _, exists := dst.edges[id]; exists {
			return fmt.Errorf("%w: %s", ErrBatchEdgeAlreadyExists, id)
		}
		for _, v := range e.Nodes {
			if _, exists := dst.nodes[v]; !exists {
				return fmt.Errorf("%w: edge %s references node %s",
					ErrBatchMissingNode, e.ID, v)
			}
		}
		edges = append(edges, e)
	}

	// Phase 2 (commit): preflight passed; do the moves under the same
	// lock window. No further validation needed — the graphs were
	// inspected under this exact lock, no concurrent mutation possible.
	for _, e := range edges {
		dst.edges[e.ID] = e
		for _, v := range e.Nodes {
			dst.incidence[v][e.ID] = struct{}{}
		}
		delete(g.edges, e.ID)
		for _, v := range e.Nodes {
			delete(g.incidence[v], e.ID)
		}
	}
	return nil
}

// RemoveBatch atomically removes the named hyperedges from g. All-or-
// nothing semantics matching [Graph.PromoteBatch]: every id must exist;
// if any is missing the whole batch fails before any mutation.
//
// Soundness: per `Wyrd.Hypergraph.hyperedge_preserves_incident_edges`
// (Phase 2 C-20a, generalised to the batch case in Phase 4 v1.6 by
// induction), for any node v not incident on any e in ids,
// IncidentEdges(v) is unchanged by RemoveBatch.
//
// Errors: [ErrBatchEdgeNotFound] for the first missing id encountered
// during preflight; no mutation occurs on error.
func (g *Graph) RemoveBatch(ids []HyperedgeID) error {
	if g == nil {
		return fmt.Errorf("model: graph: nil graph")
	}
	g.mu.Lock()
	defer g.mu.Unlock()

	// Phase 1 (preflight): validate every id exists.
	edges := make([]Hyperedge, 0, len(ids))
	for _, id := range ids {
		e, ok := g.edges[id]
		if !ok {
			return fmt.Errorf("%w: %s", ErrBatchEdgeNotFound, id)
		}
		edges = append(edges, e)
	}

	// Phase 2 (commit): delete under the same lock window.
	for _, e := range edges {
		delete(g.edges, e.ID)
		for _, v := range e.Nodes {
			delete(g.incidence[v], e.ID)
		}
	}
	return nil
}

// (PromoteBatchWithCapability and RemoveBatchWithCapability — the
// capability-gated forms — land in a follow-up PR after the
// capability-enforcement implementation merges per `doc/design/
// bridge-batch.md` §5. They're a small wrapper: preflight adds one
// `cap.AllowsWrite(e.Tier())` check per edge before the existing
// preflight checks; commit phase is identical.)

// orderLocks returns the two graphs in a deterministic order based on
// pointer address. Used by two-graph operations (PromoteBatch) to
// acquire locks in a globally-consistent order, preventing deadlock
// with concurrent reverse-direction batches.
func orderLocks(a, b *Graph) (first, second *Graph) {
	// #nosec G103 -- pointer-address comparison is the standard Go idiom
	// for deterministic two-mutex lock ordering (Go memory model permits
	// this; no aliasing or arithmetic is performed). Without it, two
	// concurrent reverse-direction PromoteBatch calls deadlock.
	if uintptr(unsafe.Pointer(a)) < uintptr(unsafe.Pointer(b)) {
		return a, b
	}
	return b, a
}
