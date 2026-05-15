package query

import (
	"slices"

	"github.com/JamesPagetButler/wyrd/model"
)

// OrientedIncidence partitions a node's incident edges by the node's
// role in each edge's orientation (per PR #31 §5).
//
//   - Incoming: v ∈ e.TailNodes (v is a sink — the edge points into v
//     from one of e.HeadNodes).
//   - Outgoing: v ∈ e.HeadNodes (v is a source — the edge points out
//     from v toward one of e.TailNodes).
//   - Transit:  v ∈ e.Nodes but v is neither head nor tail (the edge
//     touches v as context, not as a directional endpoint).
//   - Symmetric: e.IsOriented() == false (v's role is unoriented for
//     this edge; the edge carries no orientation metadata).
//
// Soundness: per Wyrd.HypergraphOriented.oriented_edge_preserves_
// incident_edges (forthcoming; PR #31 §4.3 reduction pattern), the
// union of all four buckets equals query.API.IncidentEdges(v). No
// edge incident on v appears in zero or more than one bucket.
type OrientedIncidence struct {
	Incoming  []model.HyperedgeID
	Outgoing  []model.HyperedgeID
	Transit   []model.HyperedgeID
	Symmetric []model.HyperedgeID
}

// IncidentOrientedEdges returns the incident edges of v partitioned by
// v's role in each edge's orientation. The returned slices are freshly
// allocated; empty slices (not nil) for buckets with no edges.
//
// Companion to [API.IncidentEdges] (which returns the flattened total).
// Per PR #31 §2.1: `IncidentEdges(v) == union of all four buckets`.
// Consumers needing oriented traversal call this; consumers wanting
// the combinatorial total call IncidentEdges.
func (q *API) IncidentOrientedEdges(v model.NodeID) OrientedIncidence {
	out := OrientedIncidence{
		Incoming:  []model.HyperedgeID{},
		Outgoing:  []model.HyperedgeID{},
		Transit:   []model.HyperedgeID{},
		Symmetric: []model.HyperedgeID{},
	}
	for _, eid := range q.g.IncidentEdges(v) {
		e, ok := q.g.Hyperedge(eid)
		if !ok {
			continue
		}
		if !e.IsOriented() {
			out.Symmetric = append(out.Symmetric, eid)
			continue
		}
		// Find v's index in e.Nodes.
		idx := -1
		for i, n := range e.Nodes {
			if n == v {
				idx = i
				break
			}
		}
		if idx < 0 {
			// Shouldn't happen — IncidentEdges only returns edges
			// containing v. Defensive skip.
			continue
		}
		role := classifyIndex(idx, e.Heads, e.Tails)
		switch role {
		case roleHead:
			out.Outgoing = append(out.Outgoing, eid)
		case roleTail:
			out.Incoming = append(out.Incoming, eid)
		case roleTransit:
			out.Transit = append(out.Transit, eid)
		}
	}
	return out
}

type roleClass int

const (
	roleTransit roleClass = iota
	roleHead
	roleTail
)

func classifyIndex(idx int, heads, tails []int) roleClass {
	if slices.Contains(heads, idx) {
		return roleHead
	}
	if slices.Contains(tails, idx) {
		return roleTail
	}
	return roleTransit
}
