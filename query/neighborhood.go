package query

import (
	"errors"
	"fmt"
	"sort"

	"github.com/JamesPagetButler/wyrd/model"
)

// ErrTierImmuneExcluded is wrapped by [API.Neighborhood] when the node
// budget would exclude a TierImmune node reachable within depth — the
// Nidavellir#38 "baby" invariant (wyrd#104). Callers detect it with
// errors.Is rather than matching the message text.
var ErrTierImmuneExcluded = errors.New("TierImmune node excluded by node budget")

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
// full ring would exceed maxNodes it is cut by retention priority —
// TierImmune first (best-effort), then Salience descending, then NodeID
// ascending — and a cut (or an unvisited nonempty ring beyond an
// exactly-full budget) sets Truncated.
//
// TierImmune guarantee (the Nidavellir#38 "baby" invariant, wyrd#104):
// no TierImmune node reachable within depth — a seed, lineage record,
// defeasibility override, or sole-alternative-path node — is ever
// silently excluded from the shard. This is a WHOLE-TRAVERSAL property,
// not a per-ring one: a locally-benign truncation or exact-full break
// can foreclose the only path to a deeper mandatory node, so it is
// enforced once, after the budgeted pass, by a budget-free reachability
// check that walks the full depth-bounded neighbourhood regardless of
// maxNodes. If any reachable TierImmune node is missing from the shard the
// call fails loud (see Errors) rather than return a silently-incomplete
// shard. (Detect-and-fail-loud is the invariant floor; preserving a path
// to an includable mandatory node by evicting optionals is a future
// enhancement.)
//
// Neighbor semantics match [API.NeighborNodes]: combinatorial
// membership, orientation ignored at v0.1, arity-1 self-edges do not
// produce neighbors.
//
// Errors: wraps [model.ErrNodeNotFound] if anchor does not exist;
// depth < 1 or maxNodes < 1 is rejected; and a budget that excludes any
// TierImmune node reachable within depth is a loud error naming the
// excluded set (callers MUST surface it, not swallow it). The returned
// Subgraph is zero on any error — never a partial shard presented as
// complete.
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
			// Exactly-full shard with a nonempty unvisited ring beyond it:
			// the shard is a prefix. (Whether that prefix silently excludes
			// a reachable TierImmune node is settled once, after the loop,
			// by the whole-traversal check below — not here.)
			out.Truncated = true
			break
		}
		if len(ring) > budget {
			out.Truncated = true
			// Per-ring retention is BEST-EFFORT: TierImmune first (so a
			// mandatory node is kept whenever it fits), then Salience
			// descending, NodeID ascending as the deterministic tie-break.
			// This is not the TierImmune guarantee — that is the whole-
			// traversal check after the loop (wyrd#104; a cut here can still
			// foreclose the only path to a deeper mandatory node, P1/P2).
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

	// Whole-traversal TierImmune guarantee (wyrd#104 §I4 — arch + cth + edda).
	// Per-ring retention above is best-effort; the actual invariant is that no
	// TierImmune node reachable within depth is SILENTLY excluded. A locally-
	// benign truncation or exact-full break can foreclose the only path to a
	// deeper mandatory node (probes P1/P2), so the guarantee is argued over the
	// whole depth bound, not per step: a budget-free, ID-only BFS finds every
	// reachable TierImmune node; any absent from the materialised shard is a
	// loud error (zero Subgraph), never a silently-incomplete one. (Detect-and-
	// fail-loud is the required floor; preserving a path to an includable
	// mandatory node by evicting optionals along it is a future enhancement.)
	if missing := q.unreachedTierImmune(anchor, depth, hops); len(missing) > 0 {
		return Subgraph{}, fmt.Errorf(
			"query: neighborhood: %w: %d node(s) reachable within depth %d (maxNodes %d) %v — raise maxNodes or narrow depth",
			ErrTierImmuneExcluded, len(missing), depth, maxNodes, missing)
	}

	out.Edges = q.inducedEdges(hops)
	return out, nil
}

// unreachedTierImmune does a budget-free, ID-only BFS from anchor to depth
// and returns the sorted NodeIDs of the TierImmune nodes reachable within
// that bound that are ABSENT from present (the materialised shard's hops
// map). A TierImmune node — a defeasibility override, sole-alternative-path
// node, seed, or lineage record — must never be silently excluded from a
// shard (the Nidavellir#38 "baby" invariant; wyrd#104). This single whole-
// traversal check is the sole enforcement point: it catches a direct cut, an
// exact-full break, and a path foreclosed by a pruned optional parent alike,
// which no per-ring / per-cut-site check can establish.
func (q *API) unreachedTierImmune(anchor model.NodeID, depth int, present map[model.NodeID]int) []model.NodeID {
	visited := map[model.NodeID]int{anchor: 0}
	frontier := []model.NodeID{anchor}
	var missing []model.NodeID
	for h := 1; h <= depth; h++ {
		ring := q.nextRing(frontier, visited)
		if len(ring) == 0 {
			break
		}
		frontier = frontier[:0]
		for _, n := range ring {
			visited[n.ID] = h
			frontier = append(frontier, n.ID)
			if n.TierImmune {
				if _, ok := present[n.ID]; !ok {
					missing = append(missing, n.ID)
				}
			}
		}
	}
	sort.Slice(missing, func(i, j int) bool { return missing[i] < missing[j] })
	return missing
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
