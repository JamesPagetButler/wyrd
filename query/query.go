package query

import (
	"github.com/JamesPagetButler/wyrd/model"
)

// API is the read-only query surface over a model.Graph. Constructed
// with [New]; never mutates the underlying graph. Safe for concurrent
// use because all underlying Graph reads go through model.Graph's
// RWMutex (PR #14 + ADR-003 §I3).
type API struct {
	g *model.Graph
}

// New returns a query API over the given graph. Panics if g is nil
// (which is always a programmer error — there is no defensible
// recovery from a nil graph at the query layer).
func New(g *model.Graph) *API {
	if g == nil {
		panic("query: graph is nil")
	}
	return &API{g: g}
}

// GetNode returns the node with the given ID and reports whether it
// exists. Equivalent to model.Graph.Node; surfaced here so consumers
// can hold a query.API handle without also threading the *model.Graph.
func (q *API) GetNode(id model.NodeID) (model.Node, bool) {
	return q.g.Node(id)
}

// GetHyperedge returns the hyperedge with the given ID and reports
// whether it exists. Equivalent to model.Graph.Hyperedge; surfaced
// here for the same reason as GetNode.
func (q *API) GetHyperedge(id model.HyperedgeID) (model.Hyperedge, bool) {
	return q.g.Hyperedge(id)
}

// IncidentEdges returns the IDs of hyperedges incident on v, in
// unspecified order. The returned slice is freshly allocated; the
// empty slice (not nil) is returned when v has no incident edges or
// does not exist in the graph.
//
// Membership is FLATTENED: an edge whose Nodes slice contains v is
// included regardless of position. The current model.Hyperedge has
// no head/tail distinction, so flattened is the only possible answer
// from the data. Oriented traversal goes through a future separate
// primitive (Wyrd issue #30).
//
// Soundness: per Wyrd.Hypergraph.hyperedge_preserves_incident_edges
// (Phase 2 C-20a), adding a hyperedge that does not touch v leaves
// IncidentEdges(v) unchanged.
func (q *API) IncidentEdges(v model.NodeID) []model.HyperedgeID {
	got := q.g.IncidentEdges(v)
	if got == nil {
		return []model.HyperedgeID{}
	}
	return got
}

// NeighborNodes returns the IDs of all nodes connected to v by some
// hyperedge incident on v, excluding v itself, in unspecified order
// without duplicates. The returned slice is freshly allocated; the
// empty slice (not nil) is returned when v has no neighbors or does
// not exist in the graph.
//
// "Connected by some hyperedge" means: there exists an e with
// v ∈ e.Nodes and target ∈ e.Nodes. The result is the *combinatorial*
// neighbour set — directionality (if any, encoded in Hyperedge.Nodes
// ordering) is ignored at v0.1.
//
// Self-incident edges and NeighborNodes (per PR #26 §6, per
// @qbp-implementor review NEIGHBOR-EXCLUDES-SELF): an arity-1
// hyperedge with e.Nodes = [v] makes IncidentEdges(v) return [e.ID]
// but NeighborNodes(v) return []. The two answers are intentional
// and consistent with each method's contract: IncidentEdges reports
// edge membership; NeighborNodes reports *other* nodes reachable in
// one hop.
//
// Concurrency: takes RLock multiple times across the call (once per
// IncidentEdges, once per Hyperedge lookup). The result is always
// *valid* (no partial nodes, no dangling refs); it just isn't a
// single point-in-time view across iteration steps. See PR #26 §5.
func (q *API) NeighborNodes(v model.NodeID) []model.NodeID {
	edges := q.g.IncidentEdges(v)
	if len(edges) == 0 {
		return []model.NodeID{}
	}
	seen := make(map[model.NodeID]struct{}, len(edges))
	out := make([]model.NodeID, 0, len(edges))
	for _, eid := range edges {
		e, ok := q.g.Hyperedge(eid)
		if !ok {
			continue
		}
		for _, n := range e.Nodes {
			if n == v {
				continue
			}
			if _, dup := seen[n]; dup {
				continue
			}
			seen[n] = struct{}{}
			out = append(out, n)
		}
	}
	return out
}
