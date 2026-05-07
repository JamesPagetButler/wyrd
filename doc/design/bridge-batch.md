# Bridge Batch Atomicity at the Wyrd Mutation Boundary

**Status:** Design **v0.1** — open for review per ADR-003 §I4
**Tracks:** `qbp-cu-walk` seq=5 item (3); `live-test` seq=22 (S-01 requirement); seq=36 (symmetric remove-batch broadening); seq=11 (`@bma-implementor` ack — all-or-nothing with rollback)
**Governance anchor:** [`qbp-compute-unit/architecture/adr-003-m1-wdevent-observer-invariants.md`](https://github.com/JamesPagetButler/qbp-compute-unit/blob/feat/issue-7-lean2rom/architecture/adr-003-m1-wdevent-observer-invariants.md) §I3, §I4
**Companion theorems:** `Wyrd.Bridge.bridge_promote_preserves_count` (Phase 2 C-20c); `Wyrd.Bridge.bridge_promote_exactly_one_side`
**Lean target (v1.6):** `Wyrd.Bridge.bridge_promote_batch_preserves_count` (induction over batch list)
**Authors:** wyrd-implementor, with input from `@bma` (S-01 framing, seq=22, seq=36), `@bma-implementor` (atomicity rationale, seq=11), `@qbp-architecture` (decision deferral to BMA, seq=7)

---

## 0. §I4 invariant — design-doc-as-S-01-review-surface

> "Design document must land and receive explicit review from bma + bma-implementor before any implementation PR opens. This is an S-01 requirement — the design doc is the review surface through which beekeeper-gated oversight operates before decisions are woven into running code."
> — `@bma`, `live-test` seq=33; landed in ADR-003 §I4.

This document is the §I4 review surface for the eventual `Bridge.PromoteBatch` / `Bridge.RemoveBatch` implementation PR. It must receive explicit review from `@bma` and `@bma-implementor` before any code lands.

## 1. Motivation

BMA's sleep cycle compresses episodic memory into semantic memory in bursts of thousands of hyperedge moves per cycle. The current `Bridge.Promote` is single-edge atomic with destination-staging rollback (`Wyrd.Bridge.bridge_promote_preserves_count`, Phase 2 C-20c). Calling `Promote` in a loop over a batch would:

- Make the cycle's coherence dependent on the loop completing, with no rollback if the loop crashes partway.
- Expose intermediate state to the WDEvent observer between iterations — the I3 "observer-out-for-the-full-duration" invariant requires the full structural action to be atomic from the observer's perspective.
- Force every caller to write the rollback machinery themselves.

A typed batch primitive solves all three.

### 1.1 Atomicity is an S-01 requirement, not a performance choice

> "PromoteBatch atomicity → I3. If PromoteBatch runs non-atomically, a watchdog event landing mid-batch creates an apparent-completion-without-completion window — the failure mode I flagged for I3. Atomicity here is an S-01 requirement, not just a performance concern."
> — `@bma`, `live-test` seq=22.

The `Bridge.Promote{Batch,RemoveBatch}` operations MUST be all-or-nothing. The partial-failure-with-manifest alternative was on the table at v0.1 of this discussion; **it is closed** per `@bma`'s S-01 framing and `@bma-implementor`'s seq=11 reasoning:

- Sleep cycle's whole point is producing coherent semantic state.
- Bursty-but-rare workload (~thousands per 300s cycle) tolerates restart cost.
- Pre-validation in BMA's compress functor means failures at `PromoteBatch` are exceptional (probably tier-mismatch caught at the capability boundary), not routine.
- Rollback machinery centralised in Wyrd; not duplicated across every caller.

A future `PromoteBatchPartial` is opt-in, not the default — and only if a use case appears that genuinely needs partial semantics.

## 2. Proposed surface

Two new methods on `compute.Bridge`, both atomic with rollback:

```go
// PromoteBatch atomically moves the named hyperedges from Source to
// Destination. Either all listed edges land in Destination and are
// removed from Source, or none do — on any failure the Source and
// Destination graphs are restored to their pre-call state.
//
// Soundness:
//   - per-edge: same as Bridge.Promote, citing
//     Wyrd.Bridge.bridge_promote_preserves_count (Phase 2 C-20c) and
//     Wyrd.Bridge.bridge_promote_exactly_one_side.
//   - batch-level: a forthcoming Lean theorem
//     Wyrd.Bridge.bridge_promote_batch_preserves_count (Phase 4 v1.6,
//     induction over batch list) lifts C-20c to the batch case. The
//     induction is small — the inductive step is one application of
//     C-20c plus the trivial "Source ∪ Destination edge count is
//     preserved by stage + commit + remove" identity.
//   - I3 atomicity: the write Lock on both graphs is held continuously
//     across the whole batch, satisfying the "observer out for the
//     full duration of any structural action" requirement
//     (ADR-003 §I3) for the observer's perspective. (See §3 lock-
//     ordering note for the two-graph case.)
//
// Errors:
//   - ErrBridgeUnknownEdge — any id is missing in Source.
//   - ErrBridgeAlreadyPromoted — any id already exists in Destination.
//   - returns the FIRST failure encountered along with full rollback;
//     no partial commit visible to subsequent callers.
func (b *Bridge) PromoteBatch(ids []HyperedgeID) error

// RemoveBatch atomically removes the named hyperedges from a single
// graph. All-or-nothing semantics matching PromoteBatch. The
// symmetric-remove case lands here per @bma seq=36; this lets
// Contextus eviction reuse the same atomicity primitive.
//
// Soundness anchor: Wyrd.Hypergraph.hyperedge_preserves_incident_edges
// (Phase 2 C-20a, generalised to a batch lemma with Phase 4 v1.6
// induction). For any node v not incident on any e in ids,
// IncidentEdges(v) is unchanged by RemoveBatch.
func (g *Graph) RemoveBatch(ids []HyperedgeID) error
```

## 3. Implementation sketch

### 3.1 PromoteBatch

```go
func (b *Bridge) PromoteBatch(ids []HyperedgeID) error {
    // Take both graph locks in canonical order (Source first if pointer
    // address < Destination; else Destination first) to avoid deadlock
    // with concurrent reverse-direction batches.
    first, second := orderLocks(b.Source, b.Destination)
    first.mu.Lock()
    second.mu.Lock()
    defer first.mu.Unlock()
    defer second.mu.Unlock()

    // Phase 1 (no mutation): preflight every id. Catch missing-in-Source
    // and already-in-Destination errors here so we never start mutating
    // and then have to roll back.
    edges := make([]Hyperedge, 0, len(ids))
    for _, id := range ids {
        e, ok := b.Source.edges[id]
        if !ok {
            return fmt.Errorf("%w: %s", ErrBridgeUnknownEdge, id)
        }
        if _, exists := b.Destination.edges[id]; exists {
            return fmt.Errorf("%w: %s", ErrBridgeAlreadyPromoted, id)
        }
        edges = append(edges, e)
    }

    // Phase 2 (mutate): all preflight checks passed; do the moves.
    // No further error checking is needed for routine cases — the
    // graphs were validated under the same lock. In adversarial cases
    // (e.g., disk-backed Walk-phase store returns I/O error mid-write),
    // we'd need to track committed indices and roll back; v0.1 ships
    // in-memory only, so this is a TODO for the MuninnDB-backed
    // implementation tracked in Wyrd issue #1.
    for _, e := range edges {
        b.Destination.edges[e.ID] = e
        for _, v := range e.Nodes {
            b.Destination.incidence[v][e.ID] = struct{}{}
        }
        delete(b.Source.edges, e.ID)
        for _, v := range e.Nodes {
            delete(b.Source.incidence[v], e.ID)
        }
    }
    return nil
}
```

The two-phase pattern (preflight-validate, then commit) keeps v0.1 simple. v0.2 may need a journal-style rollback when storage backends can fail mid-commit; for the in-memory Crawl backend, the preflight is sufficient.

### 3.2 RemoveBatch

Same shape, single graph; preflight checks every id is present, then deletes in one pass under the write lock.

### 3.3 Lock-ordering note

`Bridge.Source` and `Bridge.Destination` are both `*model.Graph`. Taking two locks risks deadlock if a peer batch goes the other direction. Solutions:

- **(a)** Order locks by pointer address (chosen in §3.1). Cheap, reliable.
- **(b)** Single global "bridge mutex" on the Bridge type. Simpler, serialises all bridge ops globally. Probably fine at Crawl/Walk; revisit at Run.
- **(c)** Wait-and-retry with a stochastic backoff. Doesn't make sense for an atomic primitive — partial waits aren't all-or-nothing.

**Lean: (a).** Order-by-address is well-understood and adds zero global contention. (b) becomes attractive only if profiling shows two-graph contention is real.

## 4. What this design does NOT include (deliberate v0.1 scope)

Per `@wyrd-implementor` `interface-prep` seq=5 + `@bma-implementor` seq=11, the following are explicitly **out of scope for v0.1**:

- **`PromoteBatchPartial`** — partial-failure-manifest variant. Future opt-in only; no use case yet justifies it.
- **`BatchValidator` hook** — Gemini's seq=49 Hypergraph-Laplacian-smoothness primitive belongs as a v0.2 additive: a typed `BatchValidator interface` and `Bridge.PromoteBatchValidated(ids, validators)` method that runs validators in the preflight phase. Keeping the Crawl atomicity primitive separable from Walk inference primitives.
- **Cross-graph references** — if any edge in the batch references a node that exists only in Source, that's a missing-node error in Destination after promotion. v0.1 leaves this as the caller's preflight responsibility (BMA's compress functor already does this); v0.2 may add a `PromoteBatchWithNodes(ids, []NodeID)` for self-contained batches.
- **Storage-backend-aware journaling** — the rollback story is in-memory-perfect at Crawl. MuninnDB-backed PromoteBatch (Wyrd issue #1) needs a journal; design lands when MuninnDB integration starts.
- **Permission re-check during commit** — capability-enforced batches will land via `PromoteBatchWithCapability(ids, cap)` after PR #15 (capability enforcement) merges. The cap check happens in preflight; commit is unconditional.

## 5. Interaction with PR #15 (capability enforcement)

When PR #15's `WriteCapability` lands, the capability-gated form is:

```go
func (b *Bridge) PromoteBatchWithCapability(ids []HyperedgeID, cap WriteCapability) error
```

Preflight adds one check per edge: `cap.AllowsWrite(e.Tier())`. A single tier mismatch fails the whole batch in preflight (no mutation occurred). This ordering is intentional — the capability check is an I1+I3 gate per `@bma` seq=22's mapping; failing it is a governance event, not a "partial success" event.

The bare `PromoteBatch` keeps working at default-`TierComplex` for backward compatibility (§3 of PR #15's design mirrored here).

## 6. Soundness anchors

- `Wyrd.Bridge.bridge_promote_preserves_count` (Phase 2 C-20c) — single-edge count conservation.
- `Wyrd.Bridge.bridge_promote_exactly_one_side` (Phase 2) — post-promotion the edge is in exactly one of Source/Destination.
- `Wyrd.Bridge.bridge_promote_batch_preserves_count` (**Phase 4 v1.6, forthcoming**) — batch-level count conservation, by induction over the batch list. Inductive step is one application of C-20c. Worth ~50 lines of Lean; intentionally a small proof to keep the soundness story tight.
- ADR-003 §I3 — observer is gated OUT for the full duration of any structural action; the two-graph write Lock acquisition in §3.1 is the runtime enforcement.

## 7. Migration path

1. Land `compute.Bridge.PromoteBatch` (this design, no validator hook).
2. Land `model.Graph.RemoveBatch` (same atomicity, single graph).
3. Land `compute.Bridge.PromoteBatchWithCapability` (after PR #15 merges).
4. Land `Wyrd.Bridge.bridge_promote_batch_preserves_count` (Phase 4 v1.6 Lean lift; small induction).
5. (v0.2 follow-up) Add `BatchValidator` interface + `PromoteBatchValidated` for Gemini's Laplacian-smoothness primitive and friends.

This implementation PR ships steps 1–2 only. Steps 3–5 are tracked separately:

- Step 3 blocks on PR #15 v0.2 merge.
- Step 4 is a new Lean file (`lean/Wyrd/BridgeBatch.lean`); independent of the Go landing.
- Step 5 is a separate v0.2 design that opens after the Hypergraph Laplacian primitive (Gemini's seq=49 item 1) has its own substrate-readiness review.

Per §I4: the Go implementation PR opens after this v0.1 doc receives explicit review from `@bma` and `@bma-implementor`. v0.1 is the review surface.

## 8. Open questions for review

1. **Lock-ordering choice (§3.3)** — order-by-pointer-address (proposed) vs. global Bridge mutex. My lean: order-by-address.
2. **Deadlock test coverage** — do we want a `TestPromoteBatch_ConcurrentReverseDirection` that pairs Source↔Destination batches in opposite directions, asserting no deadlock? My lean: yes, ship with v0.1.
3. **Error reporting** — first-failure-with-rollback (proposed) vs. all-failures-collected. My lean: first-failure; the all-or-nothing semantics mean only one error matters.
4. **Symmetric `RemoveBatch` on `model.Graph` vs `compute`** — `Graph.RemoveBatch` (proposed) keeps the API near the data; alternative is `compute.RemoveBatch(g, ids)` for symmetry with `compute.Bridge.PromoteBatch`. My lean: on `Graph` because `RemoveBatch` is single-graph and doesn't depend on bridge semantics.

## 9. What this PR ships, concretely

```
compute/bridge.go         — Bridge.PromoteBatch + lock-ordering helper
model/graph.go            — Graph.RemoveBatch
compute/bridge_test.go    — preflight tests, atomicity tests, deadlock test
model/graph_test.go       — RemoveBatch atomicity tests
```

No public API breakage. `Promote` keeps working unchanged.

---

## Cross-references

- `qbp-cu-walk` seq=5 item (3) — original raise; seq=7 (`@qbp-architecture` defer); seq=11 (`@bma-implementor` all-or-nothing decision); seq=14 (consolidated ack)
- `live-test` seq=22 (`@bma` S-01-required framing); seq=36 (broaden to symmetric remove-batch); seq=37 (consensus framing)
- `interface-prep-2026-05-06` seq=5 (Crawl-atomicity ⊥ Walk-inference separation; BatchValidator deferred to v0.2)
- ADR-003 §I3, §I4
- `Wyrd.Bridge.bridge_promote_preserves_count` (Phase 2 C-20c)

---

*Status: DRAFT v0.1 — open for §I4 review. Implementation PR blocked on explicit @bma + @bma-implementor approval of this form.*
