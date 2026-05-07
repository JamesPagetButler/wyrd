# Wyrd `query/` Subpackage — Read-API for the Hypergraph Substrate

**Status:** Design **v0.1** — open for review per ADR-003 §I4
**Tracks:** Q1 from the 2026-05-06 closeout (`live-test` seq=65); paired with Q8 at M1 (BMA reins-side query primitive)
**Governance anchor:** [ADR-003](https://github.com/JamesPagetButler/qbp-compute-unit/blob/feat/issue-7-lean2rom/architecture/adr-003-m1-wdevent-observer-invariants.md) §I1, §I3, §I4
**Companion theorems:** `Wyrd.Hypergraph.hyperedge_preserves_incident_edges` (Phase 2 C-20a); `Wyrd.Projection.kernel_supervisor_safe` (Phase 1 T2.2)
**Authors:** wyrd-implementor

---

## 0. §I4 invariant — design-doc-as-S-01-review-surface

> "Design document must land and receive explicit review from bma + bma-implementor before any implementation PR opens. This is an S-01 requirement."
> — `@bma`, `live-test` seq=33; landed in ADR-003 §I4.

This document is the §I4 review surface for the `query/` subpackage implementation PR. It must receive explicit review from `@bma` and `@bma-implementor` before any code lands. **§I4 explicitly applies to cross-project interface contracts** (Q2 of the 2026-05-06 closeout, confirmed by `@bma` `live-test` seq=64); `query/` is a cross-project interface (BMA + Contextus consume it), so the gate fully applies.

## 1. Motivation

`@bma` `live-test` seq=55 flagged a material gap: **BMA can reason *about* the hypergraph but cannot read nodes or traverse edges directly.** This blocks three downstream items:

- Gemini's Meta-Watchdog framework needs to walk the graph to compute Hypergraph Laplacian smoothness (Tang 2023; Hypergraph-Inference-BMA paper §3.1).
- The Walk-α Laplacian primitives (`compute/laplacian.go`, post-Q4) need a node iterator and a neighbour-traversal primitive.
- Contextus Spec v1.3 §4.6.6's "fresh BMA in 2027 subscribes to a scope and queries without re-ingesting" relies on a query interface that doesn't exist yet.

Today `model.Graph` exposes:

- `NodeCount`, `EdgeCount`
- `Node(id) (Node, bool)`, `Hyperedge(id) (Hyperedge, bool)`
- `IncidentEdges(v NodeID) []HyperedgeID`
- `Nodes() []Node`, `Hyperedges() []Hyperedge` — full snapshot

That covers single-element lookup and full-snapshot iteration but not **traversal-from-a-starting-node**. A consumer wanting "all nodes neighbouring N within distance 1" has to call `IncidentEdges(N)`, then `Hyperedge(eid)` for each, then collect `e.Nodes`, then dedupe. That's idiomatic for a one-off but bad for hot paths and surfaces enough complexity that consumers will write incompatible-but-equivalent helpers.

`query/` is the package where these idioms become typed, testable, soundness-cited primitives.

## 2. Decision: lean read API at v0.1

Per James's Q1 decision (`live-test` seq=65 option A): **new `query/` subpackage with a minimum viable read API. DSL deferred.** Both Contextus and BMA can use it. The four named methods:

```go
package query

import "github.com/JamesPagetButler/wyrd/model"

// API is the read-only query surface over a model.Graph. Constructed
// with [New]; never mutates the underlying graph. Safe for concurrent
// use because all underlying Graph reads go through model.Graph's
// RWMutex (PR #14 / ADR-003 §I3).
type API struct {
    g *model.Graph
}

// New returns a query API over the given graph.
func New(g *model.Graph) *API

// GetNode returns the node with the given ID and reports whether it
// exists. Equivalent to model.Graph.Node; surfaced here so consumers
// can hold a query.API handle without also threading the *model.Graph.
func (q *API) GetNode(id model.NodeID) (model.Node, bool)

// GetHyperedge returns the hyperedge with the given ID and reports
// whether it exists. Equivalent to model.Graph.Hyperedge; surfaced
// here for the same reason as GetNode.
func (q *API) GetHyperedge(id model.HyperedgeID) (model.Hyperedge, bool)

// IncidentEdges returns the IDs of hyperedges incident on v, in
// unspecified order. The returned slice is freshly allocated.
//
// Soundness: per Wyrd.Hypergraph.hyperedge_preserves_incident_edges
// (Phase 2 C-20a), adding a hyperedge that does not touch v leaves
// IncidentEdges(v) unchanged. Concurrent writers do not invalidate
// the snapshot of v's incident set captured here.
func (q *API) IncidentEdges(v model.NodeID) []model.HyperedgeID

// NeighborNodes returns the IDs of all nodes connected to v by some
// hyperedge incident on v, excluding v itself, in unspecified order
// without duplicates. The returned slice is freshly allocated.
//
// "Connected by some hyperedge" means: there exists an e with
// v ∈ e.Nodes and target ∈ e.Nodes. The result is the *combinatorial*
// neighbour set — directionality (if any, encoded in Hyperedge.Nodes
// ordering) is ignored at v0.1; consumers needing directed traversal
// must filter the result themselves. v0.2 may add a
// DirectedNeighbors variant if a use case justifies it.
func (q *API) NeighborNodes(v model.NodeID) []model.NodeID
```

Four methods. No DSL, no fluent-builder, no graph algorithms. Just the
minimum that lets BMA + Contextus + the Laplacian primitive walk one
step out from a starting node.

## 3. What's deliberately NOT in v0.1

Per Q1's framing ("DSL deferred"):

- **No multi-hop traversal.** `NeighborNodes(v)` returns 1-hop only. A
  k-hop traversal walks N times in caller code or waits for v0.2.
- **No filtering by `Type` or `Tier` at the query level.** Consumers
  filter the returned slice. v0.2 may add `NodesByType(prefix)` or
  similar if BMA's classification work needs it.
- **No subgraph extraction.** The "give me the subgraph induced by
  this node set" operation is a v0.2+ candidate (and arguably belongs
  in `compute/` rather than `query/`).
- **No Laplacian, no eigenvectors, no spectral analysis.** Those land
  in `compute/laplacian.go` (post-Q4 `gonum/mat` import) which
  *consumes* `query.API`.
- **No write methods.** `query.API` is read-only. Mutation goes
  through `model.Graph.{AddNode, AddHyperedge, RemoveHyperedge,
  *WithCapability}`.
- **No graph algorithms** (BFS, DFS, shortest-path, connected
  components). Those are M1+ work and should each have their own §I4
  review surface if they go into `query/`.
- **No streaming / iterator interface.** v0.1 returns slices; v0.2
  may add channel-based iterators if profiling shows allocation
  pressure.

## 4. I-8 reins-side primitive (paired with Q8)

Per James's Q8 decision (`live-test` seq=65 option A): **M1 — paired
with Q1**. The reins-side beekeeper-driven query primitive in BMA is
the `bma query ...` reins command (or equivalent) that takes a
`query.API` over the running BMA's `model.Graph` and exposes the
four methods to the beekeeper as a CLI. That's BMA-side work; this
PR ships only the Go API that BMA's reins layer will wrap.

`@bma-implementor` owns the reins-side wrapper. The shape of the
wrapper is BMA's call; my responsibility is to ensure the Go API
under it is cleanly callable.

## 5. Concurrency contract

`query.API` holds a `*model.Graph` and dispatches every method
through the underlying `Graph.{Node,Hyperedge,IncidentEdges}` calls.
Those methods take `Graph.mu.RLock()` per PR #14 + the godoc
clarification in PR #18 ("Lock acquisition as I3 enforcement point").

This means:

- All `query.API` reads are I1-compatible (observer-only path).
- All `query.API` reads do NOT block structural mutations during the
  Lock window — they wait via RWMutex semantics.
- A single `query.API` instance is safe for concurrent use by
  multiple goroutines.
- The `NeighborNodes` method takes the RLock multiple times in a
  single call (once per `IncidentEdges`, once per `Hyperedge` lookup
  per incident edge). This is intentionally not held-as-one-lock: the
  visible result reflects the graph state across however many lock
  windows the iteration takes. A concurrent writer that adds an edge
  mid-`NeighborNodes` may or may not have its addition reflected in
  the result — this is consistent with `Nodes()` / `Hyperedges()`'s
  existing snapshot semantics. The result is always *valid* (no
  partial nodes, no dangling refs); it just isn't a single
  point-in-time view across iteration steps.

If a consumer needs strict single-snapshot semantics, the v0.2
addition of a `Snapshot()` method that takes one RLock and freezes
the result is a clean follow-up; not required at v0.1.

## 6. Error handling

`GetNode` and `GetHyperedge` return `(value, ok)` pairs matching the
underlying `model.Graph` methods. `IncidentEdges` and `NeighborNodes`
return the empty slice (not nil) when the queried node doesn't exist
— this matches a "no neighbours" reading and keeps the iteration
target uniform. No error type at v0.1.

A consumer wanting to distinguish "v doesn't exist" from "v has no
neighbours" calls `GetNode(v)` first.

## 7. Soundness anchors

- `Wyrd.Hypergraph.hyperedge_preserves_incident_edges` (Phase 2 C-20a)
  — `IncidentEdges` semantics: adding a non-incident edge leaves a
  node's incident set unchanged.
- `Wyrd.Projection.kernel_supervisor_safe` (Phase 1 T2.2) — reads at
  any tier are safe regardless of caller tier; no `ReadCapability`
  required for query operations (consistent with the capability
  v0.2 design § 4 Option A — unrestricted reads).
- ADR-003 §I1 — observer-read path at the query layer; no mutation
  side-effects.
- ADR-003 §I3 — RWMutex `RLock` window per query method; observer is
  *not* gated out during reads (it's gated out during writes only).

## 8. Migration path

1. Land `query/` package skeleton (this design's API surface) with
   tests.
2. Update `doc/integration/{cth,bma,contextus}.md` with usage
   examples — three sketches; one per consumer.
3. (Walk-α) `compute/laplacian.go` consumes `query.API`. The Laplacian
   primitive's eigenvector iteration drives the `gonum/mat` import
   landing per Q4.
4. (Walk-α, BMA-side) `bma query ...` reins command wraps
   `query.API`. Out of scope for this PR.

This PR ships steps 1–2.

## 9. What this PR ships, concretely

```
query/
  doc.go             — package doc citing ADR-003 §I1/§I3
  query.go           — API type, New, GetNode, GetHyperedge, IncidentEdges, NeighborNodes
  query_test.go      — happy-path tests for each method, edge cases (empty graph, missing node, isolated node, self-loops)
doc/integration/
  bma.md             — usage sketch added to "Hardware backend" section
  cth.md             — usage sketch added (CTH walks signal subgraphs via query)
  contextus.md       — usage sketch added (scope-membership traversal)
```

No public API breakage. All existing `model.Graph` methods unchanged.

## 10. Open questions for review

1. **Method naming — `GetNode` vs `Node`.** `query.API.Node(id)` would be a method-name collision with `model.Graph.Node` from a Go-stdlib-style perspective; `GetNode` is unambiguous. My lean: **GetNode**. Pushback OK if the team prefers the shorter form.

2. **`NeighborNodes` directionality at v0.1.** Combinatorial-only as proposed. Consumers needing directed traversal call `IncidentEdges(v)` and walk `Hyperedge.Nodes` themselves. My lean: **stay combinatorial at v0.1**, add `DirectedNeighbors` only when a use case justifies it. If BMA's M1 observer hot-path needs directionality at the query layer, flag now and I'll add a `Direction enum` parameter.

3. **Result-allocation policy on empty result.** `NeighborNodes(missing-v)` returns empty slice, not nil. Mirrors `IncidentEdges`. My lean: **empty slice everywhere** for predictable iteration. Pushback if anyone wants nil for "doesn't exist."

4. **Should `query/` import `compute/` for typed errors, or stay independent?** v0.1 has no error types so the question is moot at this stage. v0.2 may surface validation errors (e.g. "node ID format invalid"); when that lands, the import question reopens. My lean: **stay independent** so `compute/` can import `query/` cleanly without a cycle.

## 11. Why a new package vs extending `model.Graph`

This is the substantive call against option B from Q1's table ("extend
`model.Graph` with traversal helpers in `model/`"). James chose option
A — separate package. The reasoning that justifies the choice
operationally:

- **Separation of concerns.** `model/` defines the data; `query/`
  defines read access patterns; `compute/` defines algebra +
  traversal-using algorithms. Keeping the three layers separable
  makes future moves (a `query/` re-implementation against MuninnDB
  at Walk-phase, for example) cleanly substitutable.
- **`*model.Graph` is an in-memory implementation.** The `query.API`
  abstraction lets MuninnDB-backed graphs (Wyrd issue #1) ship a
  conforming implementation by giving them a `query.New(muninn.G)`-
  shaped constructor. Whether to `interface`-ify is a v0.2 decision;
  the package boundary makes that decision possible without breaking
  consumers.
- **Discoverability.** `query.API` advertises itself as the read
  surface; a new contributor reading `model/graph.go` shouldn't
  encounter 30 traversal helpers buried alongside `AddHyperedge`.

## 12. Items NOT decided here (defer to a future doc)

- Multi-graph queries (e.g., "find me all nodes in *either* the
  episodic or semantic graph that have edges to node X"). v0.x
  candidate; out of scope.
- Performance benchmarks for `NeighborNodes` at high arity. The
  unknown is "what's the realistic graph size at Walk-phase?"; we
  measure once BMA's M1 sleep cycle is wired up.
- Subscription / change-notification (`OnNodeAdded`, `OnEdgeRemoved`).
  M2+ work; sits adjacent to the WDEvent observer rather than in
  `query/`.

---

## Cross-references

- 2026-05-06 closeout: `live-test` seq=61 (questions); `live-test`
  seq=65 (James's decisions; Q1 = A, Q8 = A paired)
- ADR-003 §I1 / §I3 / §I4
- PR #14 (RWMutex) — concurrency substrate this query API rides on
- PR #21 (capability impl) — read-tier capability noted for context
- Hypergraph-Inference-BMA paper (Gemini, 2026-05-05) — Laplacian
  primitive consumer at Walk-α
- Contextus Spec v1.3 §4.6.6 — scope-bounded query consumer

---

*Status: DRAFT v0.1 — open for §I4 review. Implementation PR blocked on explicit @bma + @bma-implementor approval of this form.*
