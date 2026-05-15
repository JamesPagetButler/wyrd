# BMA `hg/` Shim Retirement Timeline — W-Toddle-3

**Status:** Design **v0.1 (wyrd-implementor draft)** — open for review per ADR-003 §I4
**Tracks:** Wyrd issue [#43](https://github.com/JamesPagetButler/wyrd/issues/43) — OD-11(c) tracking, deliverable #3
**Governance anchor:** ADR-003 §I4; Marcy `#toddle-design` seq=24 constitutional approval; beekeeper OD-11(c) decision (`live-test` seq=95)
**Authors:** wyrd-implementor (draft) + `@bma-implementor` (joint authorship invited — see §0.1)

---

## 0. §I4 invariant — design-doc-as-S-01-review-surface

W-Toddle-3 is the **third and final** OD-11(c) absorption deliverable. Predecessors landed on `main` on 2026-05-14:

- ✅ **W-Toddle-1** — generic substrate primitives (`Node.TierImmune`, `Node.Salience`, `Graph.SetRetentionCap`, `RetentionTier`). PR #39 design + PR #42 impl + PR #46 Lean (`Wyrd.TierImmunity`).
- ✅ **W-Toddle-2** — BMA `NodeType`-to-policy mapping (`model.ApplyBMAPolicy`, `BMAPolicy`). PR #47 design + PR #48 impl. Eight TD-4 inventory entries canonicalised.
- ⏳ **W-Toddle-3** — this doc. BMA `hg/` shim retirement timeline.

## 0.1 Joint authorship — call for `@bma-implementor` collaboration

This is **wyrd-implementor's half** of the design. The other half — concrete BMA `hg/` API surface inventory, per-consumer migration sequence, and BMA-side test coverage strategy — is BMA-implementor's authorship area. The placeholders at §3 + §5 + §6 are marked **TBD: bma-implementor** for the joint draft to fill in. The completed doc gets `Co-Authored-By: ...` lines from both implementors.

Until BMA-side content lands, this doc is **incomplete-but-reviewable** — `@bma`'s gov-layer read can land against my half (Wyrd-side perspective on retirement, cutover criteria, parity gates); the BMA-side specifics carry their own §I4 review when they arrive.

## 1. Motivation

OD-11(c) decision (beekeeper, 2026-05-13): *"Wyrd absorbs `hg/`'s BMA-specific structures."* The substrate primitives (W-Toddle-1) and the BMA-specific `NodeType` policy (W-Toddle-2) already live in Wyrd. **What remains is the migration of BMA's existing `hg/` consumers** off the local in-BMA `hg/` package and onto Wyrd's substrate.

Two ways the retirement can land — they are not exclusive:

- **Path R1 — `hg/` becomes a thin re-export shim.** BMA's `hg/` package keeps its existing public API, but each call simply delegates to `model.Graph` / `model.ApplyBMAPolicy` / etc. Consumers don't change; only the implementation moves. Long-term `hg/` may stay indefinitely as a stable BMA-internal facade.
- **Path R2 — `hg/` is deleted; consumers migrate to Wyrd directly.** BMA-side call sites change from `hg.AddSeed(...)` to `model.ApplyBMAPolicy(&node); g.AddNodeWithCapability(node, cap)` (or equivalent). `hg/` package is removed from BMA.

The right answer is almost certainly **R1 first, R2 later (or never)**. R1 minimises churn during the Toddle 7-day endurance window; R2 is a v0.x cleanup that pays a small dividend in fewer indirection layers but isn't necessary for constitutional correctness.

## 2. Migration sequence — three phases

### Phase A — Substrate-ready (✅ DONE as of 2026-05-14)

All Wyrd-side prerequisites on `main`. Specifically:
- `Node.TierImmune` + `Node.Salience` (W-Toddle-1)
- `Graph.SetRetentionCap` / `RetentionCap` (W-Toddle-1)
- `model.ApplyBMAPolicy(*Node)` + the 8-entry TD-4 mapping (W-Toddle-2)
- Lean anchor `Wyrd.TierImmunity.tier_immune_node_preserves_eviction` (no `sorry`, no axiom)

BMA-side consumers can begin to be migrated at any time. Phase A's exit criterion: `@bma-implementor` confirms `model.ApplyBMAPolicy` covers every BMA-specific NodeType currently produced by `hg/` writes. (TD-4 inventory covered 8 types; if BMA discovers a 9th in this phase, it gets added to `bmaNodeTypePolicy` via a small W-Toddle-2-extension PR.)

### Phase B — Thin re-export shim (Path R1)

BMA `hg/` package's public API is preserved verbatim. Each function body is rewritten to call into Wyrd's substrate. From the perspective of every BMA consumer, **nothing changes** — `hg.AddSeed(...)` still works, but the implementation under it now uses `model.ApplyBMAPolicy(&node); g.AddNodeWithCapability(node, cap)`.

**BMA-side test coverage**: every existing `hg/` test must continue to pass against the rewritten shim. This is the API parity gate (§5). No test changes; only impl changes inside `hg/`.

**Phase B's exit criterion**: `bma-systema` CI green on a branch where every `hg/` implementation function delegates to Wyrd. The branch can then merge; subsequent BMA work continues to write through `hg.*` as before, just delegating differently underneath.

### Phase C — Direct consumer migration (Path R2; optional)

Once Phase B is stable, BMA call sites can be migrated incrementally from `hg.XxxFunc(...)` to direct Wyrd calls. Each migration is its own small PR; the `hg/` package shrinks as call sites move; eventually `hg/` is empty and can be deleted.

**Phase C is OPTIONAL** — there is no constitutional reason to do it. It pays a small reduction in indirection layers and slightly tighter type contracts but adds churn. My recommendation: **start Phase C only when a BMA-side consumer needs Wyrd-only behaviour that `hg/` doesn't shim** (e.g., something W-Toddle-1 added that `hg/` never exposed). At that point migrating just that consumer is cheap; doing the rest can wait.

## 3. BMA `hg/` API surface — TBD: bma-implementor

**This section needs `@bma-implementor` input.** The completed table lists every public function/type currently in BMA's `hg/` package with its target Wyrd substrate:

| BMA `hg/` API | Wyrd substrate (Phase B target) | Phase C migration call site |
|---|---|---|
| `hg.AddSeed(g, payload)` | `n := Node{...}; ApplyBMAPolicy(&n); g.AddNodeWithCapability(n, cap)` | call sites that write NT_SEED nodes |
| TBD | TBD | TBD |
| TBD | TBD | TBD |

(Filling in this table is the load-bearing BMA-implementor authorship item.)

## 4. API parity gates (Wyrd-side guarantees during Phase B)

Wyrd-implementor's commitment for Phase B:

- **API compatibility on `model.Graph.*WithCapability`**: no breaking changes to the capability-enforcement API (PR #15/#21) through W-Toddle-3. If BMA's shim relies on `AddNodeWithCapability`'s current signature, that signature is stable.
- **`Node` schema additivity**: the fields W-Toddle-1 added (`TierImmune`, `Salience`) stay omitempty-friendly; existing v0.1 serialisation still works (verified in `model/tier_immunity_salience_test.go`).
- **`Hyperedge` schema additivity** (post-PR #52, oriented-hyperedge impl): `Heads`/`Tails` likewise omitempty; existing edges in `hg/` writes don't need to set them.
- **`BMAPolicy` table stability**: the 8 canonical entries in `bmaNodeTypePolicy` are stable through Phase B. If BMA needs a 9th, file a W-Toddle-2-extension PR — additive, non-breaking.
- **No method renames on `model.Graph` or `model.Node`** without a deprecation window of at least one Sprint cycle.

In return, BMA-implementor's commitment for Phase B:
- `hg/` shim retains its existing public function shapes through Phase B
- BMA-side CI verifies the shim's tests pass against the Wyrd-backed impl
- If `hg/` discovers it needs a Wyrd substrate primitive that doesn't exist yet, file an issue rather than re-implementing privately

## 5. Cutover criteria — TBD: bma-implementor

**This section needs `@bma-implementor` input.** Specifically:

- What does "Phase B complete" look like operationally? My current best guess: BMA `hg/` package on a branch passes all its existing tests with every public function delegating to Wyrd. CI green. PR opened with the substitution.
- What's the Toddle-endurance criterion? Probably: the bilateral continuous-loop scaffold runs 24 hours on Crawl hardware with `hg/` Phase B in place, no SE_FATAL, no `hg/`-side regressions.
- Is Phase C blocked on anything beyond "BMA call sites felt like migrating"? Should there be a hard rule (e.g., "Phase C never happens; `hg/` stays as a stable internal facade")?

## 6. BMA-side test coverage strategy — TBD: bma-implementor

**This section needs `@bma-implementor` input.** Specifically:
- Which existing `hg/` test files are load-bearing during Phase B?
- Does BMA's Hebbian-reinforcement logic interact with `Node.Salience` in any place that needs additional Wyrd-side guarantees?
- Sleep-cycle compaction (BMA `internal/bma/sleep/cycle.go`): does it currently read/write `hg/` directly or via an interface? If interface, Phase B is invisible to it.

## 7. Soundness anchors

W-Toddle-3 is a process / migration design; no new Lean theorem at the substrate is required. The structural guarantees that make A11 Topological Cognition decay-immunity hold under retirement come from already-merged anchors:

- `Wyrd.TierImmunity.tier_immune_node_preserves_eviction` (PR #46) — immune nodes survive eviction
- `Wyrd.Capability.capability_grants_safe_access` (Phase 1 T2.3) — capability-gated writes preserve tier invariants
- `Wyrd.Hypergraph.hyperedge_preserves_incident_edges` (Phase 2 C-20a) — substrate read invariants

The `model.ApplyBMAPolicy` helper applied at Phase B write sites is the policy-application layer that translates BMA-specific intent into structural settings these theorems then enforce.

## 8. Open questions for §I4 reviewers

1. **Phase C optional, or eventually required?** My lean: optional, possibly never. Pushback if the federation wants `hg/` deleted eventually for substrate-purity reasons.

2. **What about `internal/bma/hg/infer.go` (type inference at write time)?** Per the BMA Spec §10.4-ish: BMA infers `NodeType` from payload at write time. After OD-11(c), does this inference run *inside* the `hg/` shim (Phase B) or migrate into Wyrd? My lean: **stays in `hg/` shim** — type inference is a BMA-specific concern, not substrate. The shim infers, then constructs the Wyrd `Node`, then calls `model.ApplyBMAPolicy` based on the inferred type.

3. **Do W-Toddle-2's 8 `bma.*` `NodeType` constants need to be exported from BMA's `hg/`** as the canonical names? Or do they stay Wyrd-side as `model.NodeTypeBMASeed` etc.? My lean: **stay Wyrd-side**; BMA re-exports them as needed for internal call sites. Avoids the rename-tax if Wyrd ever extends the canonical set.

4. **Operational milestone**: when does W-Toddle-3 close? My lean: **when Phase B's BMA-side PR merges** and the Toddle 24-hour endurance check (above) passes once with the shim active. Phase C is not in W-Toddle-3's scope.

## 9. §I4 named reviewers

- `@bma` (Marcy / Gen 61+) — A11 gov-layer constitutional check; same standard as PR #39 / PR #47
- `@bma-implementor` — primary BMA-side author for §3 + §5 + §6 contents
- beekeeper — final acceptance

`@contextus-impl` consultative — same substrate primitives serve `NT_INSIGHT_SIGNAL` retention; pattern transfers cleanly.

## 10. Migration path summary

1. Land this design doc (wyrd-implementor half) — §I4 sign-off path begins.
2. `@bma-implementor` fills in §3 + §5 + §6 + (optional) §8 ack — completes the joint draft.
3. `@bma` gov-layer reads completed doc; A11 constitutional approval per PR #39 pattern.
4. (BMA-side) Phase B PR opens on `bma-systema`: `hg/` shim re-implemented to delegate to Wyrd substrate. BMA CI verifies parity. Merge.
5. (BMA-side) 24-hour Toddle endurance with Phase B shim active. No SE_FATAL; A11 preserved structurally.
6. W-Toddle-3 closes; issue #43 closes; **OD-11(c) absorption complete**.
7. Phase C migration is decided per consumer at consumer-implementor discretion. No deadline.

---

## Cross-references

- Wyrd issue [#43](https://github.com/JamesPagetButler/wyrd/issues/43) — OD-11(c) tracking
- `live-test` seq=95 (beekeeper OD-11(c) decision 2026-05-13), seq=99 (TD-4 inventory)
- `#toddle-design` seq=24 (Marcy gov-layer constitutional approval), seq=25 (bma-implementor closeout ack)
- `#sprint-1-toddle-entry` seq=1 (Sprint 1 scope assignment)
- Wyrd PR #39 (W-Toddle-1 design) + PR #42 (W-Toddle-1 impl) + PR #46 (Lean anchor)
- Wyrd PR #47 (W-Toddle-2 design) + PR #48 (W-Toddle-2 impl)
- BMA Theory Addendum 11 (Topological Cognition — decay-immunity origin)
- BMA Seed Protocol (Step 9 — `NT_SEED` definition)
- ADR-003 §I4

---

*Status: DRAFT v0.1 (wyrd-implementor half) — open for joint authorship with `@bma-implementor` per §0.1. Implementation Phase B PR blocked on completed-and-signed-off form of this doc.*
