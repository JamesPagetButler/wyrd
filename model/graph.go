package model

import (
	"fmt"
	"sync"
)

// Graph is an in-memory typed hypergraph. It maintains node and edge
// dictionaries plus an incidence index (node → edges containing it)
// to support fast incident-edge queries.
//
// Soundness: per `Wyrd.Hypergraph.hyperedge_preserves_incident_edges`
// (Phase 2 C-20a), adding a hyperedge that does not touch a node v
// leaves v's incident set unchanged. The incidence index is rebuilt on
// every AddHyperedge to make this property structural rather than
// implementation-fragile.
//
// # Concurrency
//
// Graph is safe for concurrent use by multiple goroutines. An internal
// [sync.RWMutex] guards all reads and writes; readers hold an RLock,
// writers hold a Lock. The mutex is read-heavy biased — typical
// consumers (BMA observer, CTH ρ_net snapshotter, scout-agent reads)
// run alongside a single sleep-cycle writer goroutine, and Go's
// RWMutex prefers readers when no writer is contending.
//
// Methods that return slices ([Graph.Nodes], [Graph.Hyperedges],
// [Graph.IncidentEdges]) return freshly-allocated copies so callers
// may iterate without holding the lock; this trades allocation for
// caller simplicity. Snapshot calls in hot paths should be batched.
//
// # Lock acquisition as I3 enforcement point
//
// Acquiring the write lock on Graph is not just a concurrency
// primitive; it is the implementation of ADR-003 §I3, the
// algebraic-isolation-aware lock boundary that gates the WDEvent
// observer OUT for the full duration of any structural action.
// In governance terms (ADR-003 §I3 / §I4):
//
//   - Read methods (RLock) are the I1 observer-read path. The
//     observer is read-only on BMA state within a tick — reads do
//     not block structural mutations and they never escalate into
//     writes within the same call.
//   - Write methods (Lock) are the I3 fire point. While the write
//     lock is held, no observer goroutine can be reading state in
//     parallel (RWMutex semantics); during this window the
//     beekeeper-gated interrupt path through the mutation boundary
//     is the only authorised sequence to firmware-tier state.
//
// Per @bma on #live-test 2026-05-06 seq=22, "the RWMutex becomes
// the implementation of the lock boundary, not just a concurrency
// detail." Future edits to Graph that touch the lock acquisition
// pattern are governance-relevant per ADR-003 §I4 and require
// explicit review before they can land.
type Graph struct {
	mu        sync.RWMutex
	nodes     map[NodeID]Node
	edges     map[HyperedgeID]Hyperedge
	incidence map[NodeID]map[HyperedgeID]struct{}
	// retentionCaps holds per-RetentionTier eviction caps (0 = disabled
	// / infinite). Per Contextus Spec v1.3 §5.4 + §9.1 (cap-per-tier
	// retention). Eviction execution is deferred to a separate primitive
	// (W-Toddle-2); this field is the policy contract only.
	//
	// Per @contextus-impl PR #39 review: the retention axis (Skeleton /
	// Distant / Peripheral / Near / Core) is intentionally separate from
	// model.Tier (algebraic Cayley-Dickson tower); see retention.go.
	retentionCaps map[RetentionTier]int
}

// NewGraph returns an empty Graph.
func NewGraph() *Graph {
	return &Graph{
		nodes:         make(map[NodeID]Node),
		edges:         make(map[HyperedgeID]Hyperedge),
		incidence:     make(map[NodeID]map[HyperedgeID]struct{}),
		retentionCaps: make(map[RetentionTier]int),
	}
}

// SetRetentionCap sets the maximum number of nodes that may be held at
// the given retention tier before automatic eviction triggers.
// cap == 0 disables eviction at that retention tier (effectively
// infinite).
//
// Per Contextus Spec v1.3 §5.4 + §9.1 (cap-per-retention-tier).
// Eviction execution (walking the saturated tier and dropping nodes)
// is deferred to a separate primitive (W-Toddle-2); this method is
// the policy contract only.
//
// Note the type: [RetentionTier], NOT [Tier]. The two axes are
// orthogonal — [Tier] is algebraic privilege (ℂ ⊂ ℍ ⊂ 𝕆 ⊂ 𝕊);
// [RetentionTier] is Spec v1.3 §9.1 retention (Skeleton / Distant /
// Peripheral / Near / Core). Per @contextus-impl PR #39 review:
// keeping these typed separately prevents axis-confusion at call sites.
//
// Eviction order under saturation: nodes with TierImmune=true are
// excluded; among the remainder, ascending Salience is evicted first.
func (g *Graph) SetRetentionCap(rt RetentionTier, cap int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if cap < 0 {
		cap = 0
	}
	g.retentionCaps[rt] = cap
}

// RetentionCap returns the eviction cap currently set for the given
// retention tier. Returns 0 if no cap is set (eviction disabled at
// that tier).
func (g *Graph) RetentionCap(rt RetentionTier) int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.retentionCaps[rt]
}

// NodeCount returns the number of nodes in the graph.
func (g *Graph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}

// EdgeCount returns the number of hyperedges in the graph.
func (g *Graph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.edges)
}

// AddNode inserts a node. Returns an error if the node is malformed
// or its ID collides with an existing node.
func (g *Graph) AddNode(n Node) error {
	if err := n.Validate(); err != nil {
		return err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, exists := g.nodes[n.ID]; exists {
		return fmt.Errorf("model: graph: node %s already exists", n.ID)
	}
	g.nodes[n.ID] = n
	g.incidence[n.ID] = make(map[HyperedgeID]struct{})
	return nil
}

// Node returns the node with the given ID and reports whether it exists.
func (g *Graph) Node(id NodeID) (Node, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	n, ok := g.nodes[id]
	return n, ok
}

// AddHyperedge inserts a hyperedge. Returns an error if:
//   - the edge is malformed
//   - any of its nodes are not present in the graph
//   - the edge ID collides with an existing edge
func (g *Graph) AddHyperedge(e Hyperedge) error {
	if err := e.Validate(); err != nil {
		return err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, exists := g.edges[e.ID]; exists {
		return fmt.Errorf("model: graph: hyperedge %s already exists", e.ID)
	}
	for _, v := range e.Nodes {
		if _, exists := g.nodes[v]; !exists {
			return fmt.Errorf("model: graph: hyperedge %s references unknown node %s", e.ID, v)
		}
	}
	g.edges[e.ID] = e
	for _, v := range e.Nodes {
		g.incidence[v][e.ID] = struct{}{}
	}
	return nil
}

// Hyperedge returns the hyperedge with the given ID and reports whether
// it exists.
func (g *Graph) Hyperedge(id HyperedgeID) (Hyperedge, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	e, ok := g.edges[id]
	return e, ok
}

// IncidentEdges returns the IDs of hyperedges incident on v, in
// unspecified order. The returned slice is freshly allocated; callers
// may modify it.
//
// Soundness: this is the Go counterpart of
// `Wyrd.Hypergraph.incidentEdges` (Phase 2 v1.1); after AddHyperedge
// of a non-incident edge, IncidentEdges(v) returns the same set as
// before by `hyperedge_preserves_incident_edges` (C-20a).
func (g *Graph) IncidentEdges(v NodeID) []HyperedgeID {
	g.mu.RLock()
	defer g.mu.RUnlock()
	set, ok := g.incidence[v]
	if !ok {
		return nil
	}
	out := make([]HyperedgeID, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	return out
}

// RemoveHyperedge removes a hyperedge by ID. Returns an error if the
// edge does not exist.
func (g *Graph) RemoveHyperedge(id HyperedgeID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	e, ok := g.edges[id]
	if !ok {
		return fmt.Errorf("model: graph: hyperedge %s does not exist", id)
	}
	delete(g.edges, id)
	for _, v := range e.Nodes {
		delete(g.incidence[v], id)
	}
	return nil
}

// Nodes returns a snapshot slice of every node in the graph in
// unspecified order.
func (g *Graph) Nodes() []Node {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		out = append(out, n)
	}
	return out
}

// Hyperedges returns a snapshot slice of every hyperedge in the graph
// in unspecified order.
func (g *Graph) Hyperedges() []Hyperedge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]Hyperedge, 0, len(g.edges))
	for _, e := range g.edges {
		out = append(out, e)
	}
	return out
}

// AddNodeWithCapability is the capability-gated form of [Graph.AddNode].
// Returns [CapabilityError] (Unwraps to [ErrCapabilityViolation]) if
// cap does not authorise writes at n.Tier; otherwise behaves exactly
// like [Graph.AddNode].
//
// Soundness: per `Wyrd.Capability.capability_grants_safe_access`
// (Phase 1 T2.3), a holder at tier T performing a tier-T' write
// (T' ≤ T) is safe; the runtime check rejects exactly the
// `Wyrd.Foundations.no_surjection_*` (T2.1.a/b/c) cases.
//
// I1+I3 framing per `@bma` `live-test` seq=22: the capability check
// at the mutation boundary is the implementation of both the I1
// observer-vs-mutator separation and the I3 beekeeper-gated
// interrupt fire point. See `doc/design/capability-enforcement.md`
// v0.2 §1.
func (g *Graph) AddNodeWithCapability(n Node, cap WriteCapability) error {
	if err := cap.AllowsWrite(n.Tier); err != nil {
		return err
	}
	return g.AddNode(n)
}

// AddHyperedgeWithCapability is the capability-gated form of
// [Graph.AddHyperedge]. Returns [CapabilityError] if cap does not
// authorise writes at e.Tier(); otherwise behaves exactly like
// [Graph.AddHyperedge].
func (g *Graph) AddHyperedgeWithCapability(e Hyperedge, cap WriteCapability) error {
	if err := cap.AllowsWrite(e.Tier()); err != nil {
		return err
	}
	return g.AddHyperedge(e)
}

// RemoveHyperedgeWithCapability is the capability-gated form of
// [Graph.RemoveHyperedge]. The tier check uses the *current* tier
// of the edge being removed; if the edge does not exist, the
// underlying RemoveHyperedge returns its own error and no
// capability check is performed (no I1/I3 boundary is being
// crossed by a no-op).
func (g *Graph) RemoveHyperedgeWithCapability(id HyperedgeID, cap WriteCapability) error {
	g.mu.RLock()
	e, ok := g.edges[id]
	g.mu.RUnlock()
	if !ok {
		// Defer to the underlying method's "not found" semantics.
		return g.RemoveHyperedge(id)
	}
	if err := cap.AllowsWrite(e.Tier()); err != nil {
		return err
	}
	return g.RemoveHyperedge(id)
}
