# Oriented-Hyperedge Schema Extension — `model.Hyperedge` v0.2

**Status:** Design **v0.1** — open for review per ADR-003 §I4
**Tracks:** Wyrd issue [#30](https://github.com/JamesPagetButler/wyrd/issues/30); paired with the BMA M1 Oriented Laplacian primitive
**Governance anchor:** [ADR-003](https://github.com/JamesPagetButler/qbp-compute-unit/blob/feat/issue-7-lean2rom/architecture/adr-003-m1-wdevent-observer-invariants.md) §I4
**Companion theorems:** `Wyrd.Hypergraph.hyperedge_preserves_incident_edges` (Phase 2 C-20a); forthcoming `Wyrd.Hypergraph.oriented_edge_preserves_incident_edges` (this design's anchor)
**Authors:** wyrd-implementor

---

## 0. §I4 invariant — design-doc-as-S-01-review-surface

> "Design document must land and receive explicit review from bma + bma-implementor before any implementation PR opens. This is an S-01 requirement."
> — `@bma`, `live-test` seq=33; landed in ADR-003 §I4.

This is a **substrate schema change**: it touches `model.Hyperedge`, the JSON wire format, the soundness theorems that depend on Hyperedge structure, and the query API contract for directed traversal. §I4 squarely applies. Implementation PR opens only after explicit `@bma` + `@bma-implementor` sign-off on this doc.

## 1. Motivation

The query subpackage v0.1 settled on **flattened** incidence semantics (`doc/design/query-subpackage.md` §2.1) for a structural reason: `model.Hyperedge.Nodes` is `[]NodeID` with no orientation metadata, and `IsSymmetric` is a bare `bool` with no slot for **which** nodes are heads vs tails when `IsSymmetric == false`. There is nothing in the data layer to distinguish.

That's a clean v0.1 stance for the query API, but it **structurally blocks** the BMA M1 Oriented Laplacian primitive (Hypergraph-Inference-BMA paper, Gemini, 2026-05-05). The primitive needs head/tail distinction to compute oriented Laplacian eigenvectors over CTH opcode flow.

`@bma` `live-test` seq=74 flagged this as **M1-prerequisite**: the substrate must land before M1 opens or the primitive is blocked. seq=77 explicitly asked this issue + design doc be filed now so it appears in M1 scope before planning closes.

The design problem: extend `Hyperedge` with head/tail metadata in a way that

- preserves backward compatibility (existing symmetric edges remain valid without modification),
- keeps `Nodes` as the authoritative all-nodes set (no duplicate-tracking),
- supports both bipartite (single-head, single-tail) and N-to-M flow patterns (multi-source observations into a multi-target write),
- emits a clean Lean soundness anchor that proves the fundamental invariant: **adding an oriented edge that doesn't touch v leaves IncidentEdges(v) unchanged**, regardless of orientation,
- allows the paired query primitive (`IncidentOrientedEdges` / `IncomingEdges` / `OutgoingEdges`) to ship in the same PR with predictable semantics.

## 2. Decision: head/tail as `[]int` indices into `Nodes`

```go
// model/hyperedge.go (proposed v0.2 schema)

type Hyperedge struct {
    ID          HyperedgeID `json:"id"`
    Nodes       []NodeID    `json:"nodes"`
    Weight      Weight      `json:"weight"`
    IsSymmetric bool        `json:"is_symmetric"`
    Created     time.Time   `json:"created"`

    // NEW v0.2: when IsSymmetric == false, Heads and Tails are
    // disjoint subsets of the index set {0..len(Nodes)-1} encoding
    // orientation. Either or both may be empty (see §3 for the
    // semantics of partial orientation). When IsSymmetric == true,
    // Heads and Tails MUST be empty (validation enforces this — they
    // are meaningless for symmetric edges).
    Heads []int `json:"heads,omitempty"`
    Tails []int `json:"tails,omitempty"`
}
```

**Why indices into `Nodes` rather than separate `[]NodeID` slices:**

- **`Nodes` stays authoritative.** The all-nodes set is one canonical place; `IncidentEdges(v)` semantics never change because v's incidence to the edge is "v ∈ Nodes," full stop, regardless of orientation.
- **No duplicate-tracking surface.** A `Heads []NodeID` field would force consumers to keep `Heads ⊆ Nodes` themselves; `[]int` indices into `Nodes` make that invariant a pointer-arithmetic question, not a set-equality question.
- **Cheaper validation.** The validator just checks `0 ≤ idx < len(Nodes)`, `Heads ∩ Tails == ∅`, and (per §3) `len(Heads) + len(Tails) ≤ len(Nodes)`. No string comparisons.
- **`omitempty` does the right thing for symmetric edges.** A symmetric edge serialises to the v0.1 form; a symmetric Hyperedge from v0.1 deserialises into a v0.2 Hyperedge with `Heads == nil`, `Tails == nil`. Wire format compatible in both directions.

## 3. Orientation patterns supported

Three patterns the schema admits, each with a use case from the BMA / Contextus / CTH triangle:

| Pattern | Heads | Tails | Use case |
|---|---|---|---|
| **Symmetric** | empty (or `nil`) | empty (or `nil`) | `IsSymmetric == true` enforces this. Contextus scope-membership; co-activation hyperedges between BMA Tier-1 nodes. |
| **Strict bipartite** | one or more | one or more, disjoint | `Heads ∪ Tails == {0..len(Nodes)-1}` (no transit nodes). CTH opcode dependency edges with explicit upstream / downstream. |
| **N-to-M with transit** | one or more | one or more | `Heads ∩ Tails == ∅` but `Heads ∪ Tails ⊊ {0..len(Nodes)-1}`; the leftover indices are "transit / context" nodes participating in the edge but not directional. Multi-source observation into single-target write where some nodes are referenced for context rather than as flow endpoints. |

**Validation invariant (per the v0.2 `Hyperedge.Validate()`):**

```
IsSymmetric == true ⇒ len(Heads) == 0 AND len(Tails) == 0
IsSymmetric == false ⇒
  ∀ idx ∈ Heads ∪ Tails: 0 ≤ idx < len(Nodes)
  Heads ∩ Tails == ∅      (a node cannot be both source and sink in one edge)
  no requirement that Heads ∪ Tails == {0..len(Nodes)-1}  (transit nodes allowed)
```

Self-loops are unchanged: a node ID may appear in `Nodes` more than once; `Heads` / `Tails` index by position, so `Heads = [0]`, `Tails = [1]`, `Nodes = [v, v, w]` is a valid self-arrow with a transit node.

## 4. Soundness anchor — Lean theorem to add

`lean/Wyrd/HypergraphOriented.lean` (new file):

```lean
/-! Phase 2 (extension) — Oriented hyperedge incidence preservation.

Companion to `Hypergraph.lean` C-20a. Proves the fundamental invariant
that the v0.2 oriented schema preserves: adding an oriented edge that
does not touch v leaves IncidentEdges(v) unchanged, regardless of
which side of the orientation v would have been on if it were touched.
-/

theorem oriented_edge_preserves_incident_edges
  (g : Hypergraph) (e : OrientedHyperedge) (v : NodeID) :
    v ∉ e.nodes →
    IncidentEdges (insertOriented g e) v = IncidentEdges g v := by
  -- proof reuses the proof of hyperedge_preserves_incident_edges
  -- with the orientation metadata observed-but-not-touched.
  sorry  -- placeholder for the actual proof in the schema PR
```

The Go-side annotation in `model/hyperedge.go`:

```go
// Soundness: per Wyrd.Hypergraph.oriented_edge_preserves_incident_edges
// (Phase 2 v0.2 extension), adding an oriented edge whose Nodes set
// excludes v leaves IncidentEdges(v) unchanged regardless of orientation.
// The flattened-incidence contract from query.API §2.1 is preserved
// in the v0.2 schema: orientation is observed by oriented-traversal
// primitives only, never by IncidentEdges itself.
```

## 5. Paired query primitive

The schema extension lands with a paired query primitive in `query/`:

```go
// query/oriented.go (new file, lands in the v0.2 schema PR)

// IncidentOrientedEdges returns the incident edges of v partitioned by
// v's role in each edge's orientation:
//
//   - Incoming: v ∈ e.Tails (v is a sink of edge e — the edge points
//     into v from one of e.Heads).
//   - Outgoing: v ∈ e.Heads (v is a source of edge e — the edge points
//     out from v toward one of e.Tails).
//   - Transit:  v ∈ e.Nodes but v ∉ e.Heads ∪ e.Tails (the edge
//     touches v as a context / transit node, not as a directional
//     endpoint).
//   - Symmetric: e.IsSymmetric == true (v's role is unoriented for
//     this edge).
//
// Soundness: per Wyrd.Hypergraph.oriented_edge_preserves_incident_edges,
// the union of all four buckets equals the result of IncidentEdges(v).
// No edge incident on v appears in zero or more than one bucket.
type OrientedIncidence struct {
    Incoming  []model.HyperedgeID
    Outgoing  []model.HyperedgeID
    Transit   []model.HyperedgeID
    Symmetric []model.HyperedgeID
}

func (q *API) IncidentOrientedEdges(v model.NodeID) OrientedIncidence
```

This is **additive** — no breaking change to `IncidentEdges`, which keeps its v0.1 flattened semantics. Consumers needing oriented traversal call the new primitive; consumers wanting the combinatorial total call `IncidentEdges` and get the union, exactly as today.

## 6. Backward compatibility

- **Existing edges unchanged.** All v0.1 edges are deserialised with `Heads == nil`, `Tails == nil`. The validator's `IsSymmetric == true ⇒ Heads + Tails empty` rule is satisfied; the validator's `IsSymmetric == false` rule is satisfied because `len(empty) == 0` makes the disjointness and bounds invariants trivially true.
- **Wire format.** v0.1 JSON deserialises into v0.2 Hyperedge cleanly (omitempty on the new fields). v0.2 Hyperedges with empty `Heads` and `Tails` serialise to a wire form indistinguishable from v0.1.
- **Soundness theorems.** `hyperedge_preserves_incident_edges` (Phase 2 C-20a) is unchanged. The new `oriented_edge_preserves_incident_edges` is a strictly additive theorem proving that the *new* schema preserves the *old* invariant.
- **Storage backends.** The forthcoming MuninnDB (Wyrd issue #1) marshals the v0.2 schema natively; v0.1 storage backends would need a migration path, but no v0.1 storage backend exists yet — the in-memory `model.Graph` is the only consumer.

## 7. What lands in the schema PR

```
model/hyperedge.go        — schema extension (Heads, Tails fields)
model/hyperedge_test.go   — validator tests for the new invariants
model/graph_test.go       — round-trip JSON tests confirming v0.1 ↔ v0.2 wire compatibility
lean/Wyrd/HypergraphOriented.lean — new Lean file with the soundness theorem (no sorry)
lean/Wyrd.lean            — import the new file
query/oriented.go         — IncidentOrientedEdges API
query/oriented_test.go    — happy-path + transit-only edge + self-loop tests
doc/integration/bma.md    — usage sketch for the Oriented Laplacian primitive
```

## 8. What lands AFTER the schema PR (separate issues)

- **The actual Oriented Laplacian body** in `compute/laplacian.go` — replaces the v0.1 stub from PR #29 with the symmetric / oriented dispatch on `IsSymmetric`. Tracked by the follow-on of issue #24.
- **MuninnDB backend awareness** — when Wyrd issue #1 lands, the storage layer marshals/unmarshals the new fields; tracked in issue #1's scope.

## 9. Open questions for review

1. **Validation strictness on `Heads ∪ Tails ≠ Nodes`.** Should the validator *reject* edges where `Heads ∪ Tails ⊊ Nodes` (strict bipartite only), or accept transit-node patterns (the proposal in §3)? My lean: **accept transit nodes**. The N-to-M flow with context is a real pattern in the Hypergraph-Inference-BMA paper (Tang 2023 §3.2's "incidence-with-context"). Pushback if BMA's M1 primitive only ever sees strict bipartite.

2. **`Heads` / `Tails` as `[]int` vs `[]uint16`.** `[]int` is idiomatic Go and matches existing slice conventions; `[]uint16` would shave memory at very high arity. My lean: **`[]int`** for v0.2; revisit if profiling shows allocation pressure at Walk-α.

3. **Self-loop semantics.** A self-loop edge with `Nodes = [v, v]`, `Heads = [0]`, `Tails = [1]` represents v→v as a directional self-arrow. Is this a real use case, or should the validator reject self-loop orientation as nonsensical? My lean: **accept**. CTH opcode flow can have self-edges (a node depending on its own prior version); rejecting them now closes a door we can't reopen without a v0.3.

4. **Multi-edge JSON repeatability.** If the same edge ID is serialised twice, does v0.2 require both encodings to have identical `Heads` / `Tails` (canonical ordering inside the slices)? My lean: **canonical-sort `Heads` and `Tails` before serialisation** (ascending integers); this makes two encodings of the same edge byte-equal. Validator can either enforce or normalise; I'd normalise (silently sort on Validate).

## 10. Migration path

1. Land this design doc — `@bma` + `@bma-implementor` §I4 sign-off.
2. Open the schema PR per §7. CI must stay green; existing tests must pass without modification (backward compatibility is structurally enforced).
3. The Lean theorem in `HypergraphOriented.lean` ships proven (no `sorry`) in the same PR; CI's "no sorry, no axiom" gate covers it.
4. Once the schema PR merges, the Oriented Laplacian body PR opens and replaces the v0.1 `compute/laplacian.go` stub. That PR is **not** §I4-gated independently because the substrate (this design's surface) covers it; the body is a numerical implementation of an already-approved primitive shape.
5. BMA M1 work consumes both: the schema (for graph reads) and the Laplacian primitive (for the Meta-Watchdog signal).

## 11. What this design does NOT decide

- **Eviction / aging semantics for oriented edges.** Same as for symmetric edges; no special treatment. If oriented edges should age differently (e.g., tail-only nodes evict before head-only nodes), that's a separate retention-tier discussion that belongs in BMA's sleep-cycle design.
- **Higher-rank tensor structure.** `Heads` and `Tails` are flat sets, not ordered tuples. If a future use case needs "the i-th head connects to the j-th tail" matrix structure, that's a v0.3 conversation; v0.2 is the partition-into-roles step.
- **Storage-layer index optimisation.** MuninnDB issue #1 will likely want a directional-incidence index for oriented-traversal hot paths. That's an issue #1 implementation detail, not a v0.2 schema concern.

---

## Cross-references

- Wyrd issue [#30](https://github.com/JamesPagetButler/wyrd/issues/30) — issue this design closes
- `live-test` seq=72 / 74 / 77 (bma raises, signs off on flattened v0.1, asks issue + design filed)
- `doc/design/query-subpackage.md` §2.1 (directionality contract — the path forward this doc executes)
- ADR-003 §I4
- Hypergraph-Inference-BMA paper (Gemini, 2026-05-05) — Oriented Laplacian primitive consumer at M1

---

*Status: DRAFT v0.1 — open for §I4 review. Implementation PR blocked on explicit @bma + @bma-implementor approval of this form.*
