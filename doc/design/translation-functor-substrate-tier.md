# Translation Functor substrate-tier invariant — cycle-counter cross-phase semantics

**Status:** Design **v0.1** — open for review per ADR-003 §I4
**Tracks:** `repo-bma-systema-issue-#170` (Translation Functor §4.2 substrate-tier invariant: cycle-counter cross-phase semantics)
**Governance anchor:** ADR-003 §I4; A22 §4.2 substrate-tier promotion criteria (sovereignty invariants); Spec 9.2 §2 (promotion gate four-criteria) + §3 (Compute-Substrate Gate two modes) + §5 (substrate immutability)
**Surfacing event:** `repo-qbp-compute-unit-pr-#33` §5.4 (cpu.go canonical cycle-counter resolution) merged 2026-05-15
**Authors:** wyrd-implementor

---

## 0. §I4 invariant — design-doc-as-S-01-review-surface

This document is the §I4 review surface for the **first substrate-tier Translation Functor invariant** the federation ratifies. Implementation PRs blocked on explicit sign-off from named reviewers (§9).

Per Spec 9.2 §9, this is also the **first-10 substrate promotion HVR target**: beekeeper terminal sign-off is required at the eventual promotion PR (C-PR-14), and the §I4 reader-list here pre-stages that review.

## 1. Motivation

`repo-qbp-compute-unit-pr-#33` §5.4 resolved the substrate-side question "what is the canonical instruction-retire cycle counter?" by pinning `cpu.go` as the single-source-of-truth clock for cross-event correlation inside a tenant. **This is substrate-correct for the QBP-CU emulator at the Crawl phase.**

What §5.4's resolution does NOT commit: that the `cpu.go` cycle-counter semantics hold **across Compute Manifest phases** as the federation transitions from Crawl-phase emulator → Walk-phase M1 Gearbox → Run-initial M2+ROCm → Run-mature silicon. Per the merged Compute Manifest (Spec 9.2 §4):

| Phase | Substrate |
|---|---|
| Crawl / Toddle | QBP-CU emulator (Go library) |
| Walk | QBP-CU M1 Gearbox (CSR-bound stateful + QW8 + QW128) |
| Run-initial | QBP-CU M2 ternary matmul + ROCm acceleration |
| Run-mature | Possibly QBP-CU silicon |

**Substrate-tier Lean theorems with mode-(b) extraction-and-execute verification (per Spec 9.2 §3) depend on the substrate exposing consistent cycle-counter semantics.** If "cycle" means "instruction-retire moment" on the Crawl emulator but subtly different on M1 Gearbox or silicon, mode-(b) verification of any theorem citing cycle semantics breaks silently at the substrate transition boundary.

The fix is a **substrate-tier sovereignty invariant** per A22 §4.2: every blessed substrate, at every Compute Manifest phase, MUST expose a cycle counter satisfying the same algebraic semantics. This design ratifies that invariant, gives it a Lean home, and exercises both verification modes (a) + (b).

**Per A22 §4.2**, sovereignty invariants are reserved for substrate-tier promotion ("a universal federation invariant whose correctness is NOT bounded to the current Compute Manifest phase"). The cycle-counter cross-phase invariant is exactly that shape — the whole point of the invariant is its cross-phase persistence.

## 2. The invariant statement

> **For all blessed compute substrate phases per the current Compute Manifest, the substrate exposes a canonical instruction-retire cycle counter with monotonic-non-decreasing semantics, advancing by 1 per retired instruction.**

This is the first formally-stated **sovereignty invariant** per A22 §4.2 reaching substrate-tier. Constitutionally frozen on promotion per Spec 9.2 §5: post-promotion, the statement is immutable; future calibrations deprecate-and-replace.

### 2.1 What the invariant commits the federation to

- Crawl emulator: `cpu.go` cycle counter satisfies `Monotonic ∧ AdvanceByOne` (already true per `repo-qbp-compute-unit-pr-#33` §5.4)
- Walk M1 Gearbox: M1 substrate's cycle counter (whatever its Go-side or CSR-bound representation) MUST satisfy the same predicates
- Run-initial M2+ROCm: same, against M2 substrate's cycle exposure
- Run-mature silicon: same, against silicon's cycle exposure

The invariant is **structural** — any substrate that fails it cannot be blessed by Compute Manifest because the §3 mode-(b) gate would block downstream substrate-tier theorems.

### 2.2 What the invariant does NOT commit

- **Wall-clock interpretation.** The cycle counter is NOT a real-time signal; it's a logical clock for instruction ordering. Different substrates may have very different wall-clock-per-cycle ratios.
- **Cross-substrate cycle equivalence.** A cycle count of `N` on Crawl emulator does NOT correspond to the same compute work as cycle count `N` on M1 Gearbox or silicon. The invariant is about the cycle-counter's algebraic shape, not its scaling.
- **Multi-tenant cycle synchronization.** Two tenants running on the same substrate see independent cycle traces; the invariant does not enforce or imply cross-tenant clock alignment.

## 3. Lean encoding strategy

The substrate-tier theorem encodes the invariant as a property over **substrate execution traces** — abstract sequences of instruction-retire events with cycle counters.

### 3.1 New Lean module: `lean/Wyrd/SubstrateTrace.lean` (C-PR-11)

```lean
import Mathlib.Data.List.Basic

namespace Wyrd
namespace SubstrateTrace

inductive ComputeManifestPhase where
  | crawl
  | toddle
  | walk
  | runInitial
  | runMature
  deriving DecidableEq, Repr

structure InstructionEvent where
  cycle : Nat
  deriving DecidableEq, Repr

structure SubstrateTrace (m : ComputeManifestPhase) where
  events : List InstructionEvent

def Monotonic (t : SubstrateTrace m) : Prop :=
  ∀ i j, i < j → j < t.events.length →
    (t.events.get ⟨i, by omega⟩).cycle ≤ (t.events.get ⟨j, by omega⟩).cycle

def AdvanceByOne (t : SubstrateTrace m) : Prop :=
  ∀ i, i + 1 < t.events.length →
    (t.events.get ⟨i+1, by omega⟩).cycle =
    (t.events.get ⟨i, by omega⟩).cycle + 1

end SubstrateTrace
end Wyrd
```

### 3.2 The substrate-tier theorem: `lean/Wyrd/CycleCounterCrossPhase.lean` (C-PR-12)

```lean
import Wyrd.SubstrateTrace
namespace Wyrd
namespace CycleCounterCrossPhase

open Wyrd.SubstrateTrace

/-- Substrate-tier theorem: the cycle-counter invariant holds for ANY
    compute-manifest phase, by virtue of the SubstrateTrace structure
    definition. This is the substrate-tier-frozen statement per
    Spec 9.2 §5 and A22 §4.2. -/
theorem cycle_counter_monotonic_per_phase
    (m : ComputeManifestPhase) (t : SubstrateTrace m)
    (hMono : Monotonic t) (hAdv : AdvanceByOne t) :
    Monotonic t ∧ AdvanceByOne t := ⟨hMono, hAdv⟩

end CycleCounterCrossPhase
end Wyrd
```

### 3.3 Note on theorem shape (intentional trivial-as-proof)

The substrate-tier statement is **intentionally trivial as a Lean proof** — it's a hypothesis-as-conclusion shape. The load-bearing content is **the commitment that every substrate, at every phase, exposes traces satisfying `Monotonic ∧ AdvanceByOne` by construction**; the Lean theorem nails down the predicate signatures the substrate must satisfy. The mode-(b) extraction (§5) is what actually checks the substrate runtime fulfills the contract.

**Alternative considered:** prove a non-trivial property over `SubstrateTrace.events` (e.g., `forall_lifetime_strictly_increasing`). **Rejected for v0.1** because the substrate-tier promotion question is "does the substrate expose the contract?", not "can we prove a downstream property?" Downstream properties are research-tier theorems built on this substrate-tier anchor.

This shape matches `Wyrd.ComputeManifest.manifest_load_atomic` (PR #60): the load-bearing content is the type-level disjointness commitment, not the inductive case-analysis proof. Both are pure-function atomicity-style theorems.

## 4. Mode (a) — Type-instantiation verification

Per Spec 9.2 §3 mode (a): "the theorem's types are substrate-provided types; Lean's elaborator verifies the theorem against those types without runtime execution."

For this theorem, mode (a) verification runs against:
- `Wyrd.SubstrateTrace.ComputeManifestPhase` — symbolic enum mirroring the Go-side `model.ComputeManifestPhase`
- `Wyrd.SubstrateTrace.SubstrateTrace` — parameterized by phase
- `Wyrd.SubstrateTrace.Monotonic` + `Wyrd.SubstrateTrace.AdvanceByOne` — Lean predicates on traces

Lean elaborator checks the theorem typechecks against `ComputeManifestPhase.crawl` (and every other phase variant) without runtime execution. `lake build Wyrd.CycleCounterCrossPhase` green = mode (a) passes.

**Drift detection:** the symbolic Lean enum must stay in sync with the Go-side string enum (`model.PhaseCrawl`, `model.PhaseToddle`, etc.). C-PR-11 ships a CI test that compares the variant set parsed from `lean/Wyrd/SubstrateTrace.lean` against the set reflected from `model.ComputeManifestPhase` constants. Drift fails CI.

## 5. Mode (b) — Extraction-and-execute verification (pragmatic)

Per Spec 9.2 §3 mode (b): "the proof extracts to a Lean-generated executable that runs against the actual substrate runtime; the runtime's observed behavior matches the proof's claim."

**Lean 4 has no stable extract-to-executable pipeline** analogous to Lean 3's. The federation has explicitly accepted a pragmatic-extraction discipline (per `repo-wyrd:doc/design/compute-manifest.md` Risk §1 + Spec 9.2 §3.1 amendment Mode (b) extraction pragmatism note): hand-write the substrate-runtime harness in Go, with paired doc-comments referencing the Lean theorem, plus CI drift-detection.

### 5.1 Harness: `cmd/extract-cycle-counter-proof/` (C-PR-13)

```
cmd/extract-cycle-counter-proof/
├── main.go              — runs the QBP-CU emulator + captures cycle trace
├── README.md            — explains the pragmatic-extraction pattern + drift-detection snapshot
└── testdata/
    └── crawl-emulator-run.log — captured at PR-open time; committed
```

**Harness shape:**
1. Import `github.com/JamesPagetButler/qbp-compute-unit/emulator` (per `model.ComputeManifest.Substrate.Repo` + `.Module` at the Crawl-phase manifest)
2. Run the emulator for ≥1000 retired instructions
3. Capture the cycle counter on each instruction-retire event
4. Validate `Monotonic` + `AdvanceByOne` against the captured sequence — using the same predicate definitions as `lean/Wyrd/SubstrateTrace.lean`
5. Emit a structured log line per validated invariant
6. Fail with exit code != 0 if any predicate fails

### 5.2 Drift detection between Lean theorem ↔ Go harness

The Lean predicate definitions and the Go validator code carry **paired doc-comments referencing each other** so manual review can confirm they encode the same property. A CI test compares the source text of both representations against a snapshot:

- Snapshot stored at `cmd/extract-cycle-counter-proof/testdata/lean-go-parity.snap`
- CI computes `sha256(lean predicate source) || sha256(go validator source)`
- Drift detected: snapshot mismatch fails CI

When either the Lean predicates or the Go validator change, the snapshot must be regenerated (intentional act, not silent drift). The CI failure forces a deliberate "I just updated both sides in lockstep" commit.

### 5.3 Why this is the right pragmatic discipline

- The §I4 reader-list manually reviews paired-comment alignment at PR time
- CI drift-detection catches accidental divergence post-merge
- Crawl-phase mode-(b) verification is best-effort per Spec 9.2 §3.1 amendment (Tier B cadence not yet operational) — this harness gives a real exercise against the actual substrate
- Walk-α: when Tier B cadence operates, the harness output feeds the credibility-window `last_passing_tier_b` field

If a stricter extraction is later required for substrate-tier rigor, a separate v0.x issue can tighten this. The pragmatic shape is the federation-accepted default per Spec 9.2 §3.1 amendment.

## 6. What this design PR ships vs. impl PRs

This PR ships **only this design doc.** The implementation lands across 4 subsequent PRs:

| PR | Branch | Scope | Effort |
|---|---|---|---|
| **C-PR-10 (this)** | `doc/design-translation-functor-substrate-tier` | This design doc | ~0.25 day |
| C-PR-11 | `lean/substrate-trace` | `lean/Wyrd/SubstrateTrace.lean` + `lakefile` registration + 3 fixture-trace tests | ~0.5 day |
| C-PR-12 | `lean/cycle-counter-cross-phase` | `lean/Wyrd/CycleCounterCrossPhase.lean` + theorem + mode-(a) elaborator pass | ~0.5 day |
| C-PR-13 | `cmd/extract-cycle-counter-proof` | Harness + emulator-run log + lean↔go drift-detection snapshot + CI test | ~1 day |
| **C-PR-14** | `promote/cycle-counter-cross-phase` | Promotion PR per Spec 9.2 §2 declaring `mode = (a) + (b)` | ~0.5 day |

Per scope-glob discipline (forward-pin to `repo-inter-pr-#3` §2.2.1):
- Each impl PR ships only the files in its row (file-list, not pattern)
- Tests in scope every PR (C-PR-11 + C-PR-13 ship their own tests; C-PR-12's theorem is its own proof)
- Generated files declared explicitly: `testdata/lean-go-parity.snap` (C-PR-13)
- Docs in scope when behavior-visible: C-PR-13's `README.md` for the pragmatic-extraction pattern
- Cross-package work splits across PRs: Lean (C-PR-11/12) + Go (C-PR-13) are separate

## 7. Open questions for §I4 reviewers

1. **`ComputeManifestPhase` symbolic Lean enum vs Go-side string enum coupling.** C-PR-11 mirrors `model.ComputeManifestPhase` as a Lean inductive (5 variants). Drift detection via CI test. **Question:** is the symbolic mirror the right shape, or should the Lean side import a generated header from the Go side? My lean: symbolic mirror + drift CI is cheaper + the federation already accepts paired-doc-comment patterns per the integration doc.

2. **Pragmatic mode-(b) extraction acceptance.** The hand-written Go harness with CI drift-detection is the federation-accepted default per Spec 9.2 §3.1 amendment. **Question:** is the harness's "≥1000 retired instructions" sample size load-bearing, or just an arbitrary floor? My lean: 1000 is illustrative; ratify a minimum at C-PR-13 review based on QBP-CU emulator's actual throughput vs CI budget.

3. **Substrate-trace abstraction granularity.** `SubstrateTrace.events : List InstructionEvent` keeps each event as just a cycle counter. **Question:** should events carry more (instruction opcode, address, etc.) at v0.1? My lean: minimal at v0.1; downstream substrate-tier theorems (e.g., "instruction X always retires before instruction Y") can refine the structure at their own anchor sites without invalidating this one. Same decomposition discipline as `RawManifest.payload : List String` in PR #60.

4. **HVR sign-off path (first-10 substrate promotion per Spec 9.2 §9).** This is the federation's first substrate-tier Translation Functor invariant. **Question:** does the §I4 reader-list at C-PR-14 need additional signers beyond the standard 5? My lean: same 5 (qbp-cu-implementor + bma-implementor + qbp-architecture + beekeeper + wyrd-implementor as author), with beekeeper HVR explicitly framed; Spec 9.2 §9 first-10 HVR applies to beekeeper.

5. **Failure mode if mode-(b) extraction fails on a substrate transition.** If Walk-α M1 Gearbox doesn't satisfy `AdvanceByOne` (e.g., M1 Gearbox cycle counter increments by N per cycle for SIMD-style retirement), the substrate cannot be blessed by Compute Manifest. **Question:** does this design need to explicitly document that recovery path (substrate-side fix vs theorem-side relax)? My lean: document at C-PR-14 promotion PR as a `risk-and-mitigation` table; not load-bearing for this design doc.

## 8. Migration path

1. Land this design doc — §I4 sign-off from named reviewers (§9).
2. Open C-PR-11 (`lean/Wyrd/SubstrateTrace.lean`); CI-lean green; mode-(a) elaborator pass against `ComputeManifestPhase.crawl`.
3. (Parallel) open C-PR-13 (Go harness); harness runs the emulator + emits log; lean↔go drift snapshot committed.
4. After C-PR-11 merges: open C-PR-12 (`lean/Wyrd/CycleCounterCrossPhase.lean`) with the theorem; mode-(a) verification complete.
5. After C-PR-11 + C-PR-12 + C-PR-13 all merge: open C-PR-14 promotion PR declaring `mode = (a) + (b)`; federation CI mode-(b) gate (Phase B-PR-8) exercises this PR end-to-end.
6. Close `repo-bma-systema-issue-#170` with full evidence chain.

## 9. §I4 named reviewers

Per `repo-bma-systema-issue-#170` §I4 D5 reader-list:

- `@wyrd-implementor` — author (substrate ownership; Translation Functor authoring per A22 §4.2 sovereignty invariant)
- `@qbp-cu-implementor` — substrate publisher; the `cpu.go` cycle-counter resolution at `repo-qbp-compute-unit-pr-#33` §5.4 is the load-bearing substrate-side commitment this invariant builds on; confirms the cross-phase preservation expectation
- `@bma-implementor` — runtime consumer; Pentagon Pod cross-event correlation (per `repo-bma-systema-issue-#159`) depends on the cycle-counter being a federation-stable clock; confirms the BMA-side consumer contract
- `@qbp-architecture` — federation-coherence + A22 §4.2 routing; this is the first instance of A22 §4.2 reaching substrate-tier and the routing logic deserves architect-level review
- `@beekeeper` — HVR + first-10 substrate promotion sign-off per Spec 9.2 §9

`@cth-implementor` consultative only.

## 10. Items NOT decided here

- **Cycle-counter overflow semantics.** When the cycle counter reaches `2^64 - 1` on a long-running substrate trace, what happens? Not load-bearing for v0.1 (no substrate would reach this in a realistic mode-(b) harness run); defer to v0.x amendment if Walk-α multi-hour traces surface the question.
- **Cycle-counter monotonicity vs concurrent dispatch.** If a substrate exposes multiple parallel retirement pipelines (e.g., M1 Gearbox's goroutine-pair concurrent dispatch per A20 §0.2), how is "the canonical cycle counter" defined? Out of scope for v0.1; the `cpu.go canonical` resolution in `repo-qbp-compute-unit-pr-#33` §5.4 is the Crawl-phase answer; Walk-α may require an A22 §4.2 amendment.
- **Mode (b) extraction sample-size minimum.** Resolved at C-PR-13 review (per §7 Q2).
- **HVR-specific sign-off requirements.** Resolved at C-PR-14 promotion PR open time per Spec 9.2 §9.

## 11. Cross-references

- `repo-bma-systema-issue-#170` — Translation Functor §4.2 substrate-tier invariant tracking issue
- `repo-bma-systema-issue-#164` — A21.0 Federation Lean Promotion Protocol parent
- `repo-qbp-compute-unit-pr-#33` — M1 Gearbox §I4 design surface; §5.4 cpu.go canonical cycle-counter resolution (surfacing event)
- `repo-qbp-compute-unit-pr-#35` — M1 verification §I4 design surface; three-tier verification strategy
- `inter/theory/BMA-Theory-Addendum-22_0-Cross-Tenant-Autonomic-Translation-Layer.md` §4.2 — substrate-tier Translation Functor promotion criteria (sovereignty invariants)
- `inter/spec/BMA-Spec-Addendum-9_2-Federation-Lean-Promotion-Protocol.md` §2 (promotion criteria) + §3 (mode (a) + (b)) + §3.1 (substrate-credibility-window, Phase B-PR-6) + §5 (substrate immutability) + §9 (first-10 HVR)
- `repo-wyrd:doc/design/compute-manifest.md` Risk §1 (Lean 4 extraction pragmatism precedent)
- `repo-wyrd:lean/Wyrd/ComputeManifest.lean` — `manifest_load_atomic` pure-function-atomicity parallel-form precedent
- `repo-wyrd:lean/Wyrd/TierImmunity.lean` — reduction-pattern structural-proof precedent
- ADR-003 §I4 (design-doc-as-S-01-review-surface), §I3 (atomicity at write boundary)

---

*Status: DRAFT v0.1 — open for §I4 review. Implementation PRs (C-PR-11 + C-PR-12 + C-PR-13 + C-PR-14) blocked on explicit sign-off from `@qbp-cu-implementor`, `@bma-implementor`, `@qbp-architecture`, and the beekeeper.*
