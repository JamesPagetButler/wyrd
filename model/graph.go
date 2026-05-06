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
type Graph struct {
	mu        sync.RWMutex
	nodes     map[NodeID]Node
	edges     map[HyperedgeID]Hyperedge
	incidence map[NodeID]map[HyperedgeID]struct{}
}

// NewGraph returns an empty Graph.
func NewGraph() *Graph {
	return &Graph{
		nodes:     make(map[NodeID]Node),
		edges:     make(map[HyperedgeID]Hyperedge),
		incidence: make(map[NodeID]map[HyperedgeID]struct{}),
	}
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
