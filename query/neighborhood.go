package query

import (
	"fmt"
	"sort"

	"github.com/JamesPagetButler/wyrd/model"
)

// Subgraph is the wire-format result of [API.Neighborhood] — the
// "holographic shard" of bma-systema issue #229. It is a MAP of the
// anchor's neighborhood, not the territory: node payloads are
// deliberately excluded so a ~100-node shard fits the BMA bridge's
// 50KB injection budget; content fetch is a follow-up [API.GetNode]
// by ID.
//
// Determinism: for the same underlying graph state, Neighborhood
// returns a byte-identical Subgraph (BFS ring order; NodeID-
// lexicographic within a ring; HyperedgeID-lexicographic edges).
// Injection diffing and contract tests rely on this.
type Subgraph struct {
	Anchor model.NodeID `json:"anchor"`
	Depth  int          `json:"depth"`
	// Truncated reports that the node budget cut the traversal —
	// the shard is a prefix of the full neighborhood, not all of it.
	Truncated bool           `json:"truncated"`
	Nodes     []SubgraphNode `json:"nodes"`
	// Edges is induced-only: an edge appears iff EVERY endpoint is
	// in Nodes. No dangling references in the wire format.
	Edges []SubgraphEdge `json:"edges"`
}

// SubgraphNode is the shard projection of a [model.Node]: identity,
// type, tier, distance-from-anchor, and retention metadata — no
// payload (map, not territory).
type SubgraphNode struct {
	ID   model.NodeID   `json:"id"`
	Type model.NodeType `json:"type"`
	Tier model.Tier     `json:"tier"`
	// Hops is the BFS distance from the anchor (0 = the anchor).
	Hops       int     `json:"hops"`
	Salience   float64 `json:"salience,omitempty"`
	TierImmune bool    `json:"tier_immune,omitempty"`
}

// SubgraphEdge is the shard projection of a [model.Hyperedge]:
// identity, membership, symmetry, and orientation. Heads/Tails are
// carried verbatim (indices into Nodes, per PR #31 §3) so oriented
// provenance edges — cert→seed "read-at-founding", observation→cert —
// keep their direction in the shard: a navigational map needs its
// arrows. Empty for symmetric edges, mirroring the model invariant.
type SubgraphEdge struct {
	ID          model.HyperedgeID `json:"id"`
	Nodes       []model.NodeID    `json:"nodes"`
	IsSymmetric bool              `json:"is_symmetric,omitempty"`
	Heads       []int             `json:"heads,omitempty"`
	Tails       []int             `json:"tails,omitempty"`
}

// Neighborhood returns the depth-bounded BFS neighborhood of anchor
// as a deterministic, budget-capped [Subgraph] — the Crawl-phase
// graph-traversal primitive (bma-systema #229; the Crawl seed of the
// Walk-α typed-query surface).
//
// Traversal: breadth-first, ring by ring, up to depth hops. Within a
// ring, nodes are ordered NodeID-lexicographically. When admitting a
// full ring would exceed maxNodes, the ring is cut by retention
// priority — TierImmune nodes first, then Salience descending, then
// NodeID ascending as the deterministic tie-break — so seeds and
// lineage records never fall out of their own shard. A cut (or an
// unvisited nonempty ring beyond an exactly-full budget) sets
// Truncated.
//
// Neighbor semantics match [API.NeighborNodes]: combinatorial
// membership, orientation ignored at v0.1, arity-1 self-edges do not
// produce neighbors.
//
// Errors: wraps [model.ErrNodeNotFound] if anchor does not exist;
// depth < 1 or maxNodes < 1 is rejected.
//
// Concurrency: like the other API methods, takes RLocks across
// multiple Graph reads — the result is always well-formed but is not
// a single point-in-time snapshot under concurrent mutation (PR #26
// §5 semantics).
func (q *API) Neighborhood(anchor model.NodeID, depth, maxNodes int) (Subgraph, error) {
	if depth < 1 {
		return Subgraph{}, fmt.Errorf("query: neighborhood: depth must be >= 1, got %d", depth)
	}
	if maxNodes < 1 {
		return Subgraph{}, fmt.Errorf("query: neighborhood: maxNodes must be >= 1, got %d", maxNodes)
	}
	anchorNode, ok := q.g.Node(anchor)
	if !ok {
		return Subgraph{}, fmt.Errorf("%w: %s", model.ErrNodeNotFound, anchor)
	}

	out := Subgraph{Anchor: anchor, Depth: depth}
	hops := map[model.NodeID]int{anchor: 0}
	out.Nodes = append(out.Nodes, projectNode(anchorNode, 0))
	frontier := []model.NodeID{anchor}

	for h := 1; h <= depth; h++ {
		ring := q.nextRing(frontier, hops)
		if len(ring) == 0 {
			break
		}
		budget := maxNodes - len(out.Nodes)
		if budget == 0 {
			// Exactly-full shard with a nonempty unvisited ring
			// beyond it: the shard is a prefix.
			out.Truncated = true
			break
		}
		if len(ring) > budget {
			out.Truncated = true
			// Priority eviction: TierImmune first, Salience
			// descending, NodeID ascending.
			sort.Slice(ring, func(i, j int) bool {
				a, b := ring[i], ring[j]
				if a.TierImmune != b.TierImmune {
					return a.TierImmune
				}
				if a.Salience != b.Salience {
					return a.Salience > b.Salience
				}
				return a.ID < b.ID
			})
			ring = ring[:budget]
			// Restore the within-ring output ordering.
			sort.Slice(ring, func(i, j int) bool { return ring[i].ID < ring[j].ID })
		}
		frontier = frontier[:0]
		for _, n := range ring {
			hops[n.ID] = h
			out.Nodes = append(out.Nodes, projectNode(n, h))
			frontier = append(frontier, n.ID)
		}
	}

	out.Edges = q.inducedEdges(hops)
	return out, nil
}

// nextRing collects the not-yet-visited neighbor nodes of the
// frontier, NodeID-lexicographically ordered.
func (q *API) nextRing(frontier []model.NodeID, visited map[model.NodeID]int) []model.Node {
	seen := make(map[model.NodeID]struct{})
	for _, v := range frontier {
		for _, nb := range q.NeighborNodes(v) {
			if _, vis := visited[nb]; vis {
				continue
			}
			seen[nb] = struct{}{}
		}
	}
	ring := make([]model.Node, 0, len(seen))
	for id := range seen {
		if n, ok := q.g.Node(id); ok {
			ring = append(ring, n)
		}
	}
	sort.Slice(ring, func(i, j int) bool { return ring[i].ID < ring[j].ID })
	return ring
}

// inducedEdges returns every hyperedge whose endpoints ALL lie in the
// visited set, HyperedgeID-lexicographically ordered.
func (q *API) inducedEdges(visited map[model.NodeID]int) []SubgraphEdge {
	edgeSeen := make(map[model.HyperedgeID]struct{})
	out := []SubgraphEdge{}
	for id := range visited {
		for _, eid := range q.IncidentEdges(id) {
			if _, dup := edgeSeen[eid]; dup {
				continue
			}
			edgeSeen[eid] = struct{}{}
			e, ok := q.g.Hyperedge(eid)
			if !ok {
				continue
			}
			inside := true
			for _, v := range e.Nodes {
				if _, vis := visited[v]; !vis {
					inside = false
					break
				}
			}
			if inside {
				out = append(out, SubgraphEdge{
					ID:          e.ID,
					Nodes:       append([]model.NodeID(nil), e.Nodes...),
					IsSymmetric: e.IsSymmetric,
					Heads:       append([]int(nil), e.Heads...),
					Tails:       append([]int(nil), e.Tails...),
				})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func projectNode(n model.Node, hops int) SubgraphNode {
	return SubgraphNode{
		ID:         n.ID,
		Type:       n.Type,
		Tier:       n.Tier,
		Hops:       hops,
		Salience:   n.Salience,
		TierImmune: n.TierImmune,
	}
}
