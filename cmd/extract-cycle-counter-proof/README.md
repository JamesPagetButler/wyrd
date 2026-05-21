# `cmd/extract-cycle-counter-proof` — pragmatic mode-(b) extraction harness

Phase C-PR-13 deliverable for [`repo-bma-systema-issue-#170`](https://github.com/JamesPagetButler/bma-systema/issues/170). This binary exercises the substrate-tier theorem `Wyrd.CycleCounterCrossPhase.cycle_counter_monotonic_per_phase` (PR #66, merged `c81b7a7`) against the actual QBP-CU emulator runtime — the **mode-(b) extraction-and-execute** verification path per Spec 9.2 §3 + §3.1 amendment.

## Why this exists

Per Spec 9.2 §3.1 amendment ([`repo-inter` PR #6](https://github.com/JamesPagetButler/inter/pull/6), merged `e774069`):

> Lean 4 lacks a stable extract-to-executable pipeline analogous to Lean 3's. The federation has explicitly accepted a pragmatic-extraction discipline: hand-write the substrate-runtime harness in Go, with paired doc-comments referencing the Lean theorem, plus CI drift-detection.

This binary implements that pragmatic-extraction discipline for the first federation substrate-tier sovereignty invariant per A22 §4.2.

## How it works

1. **Loads the Compute Manifest** (`manifest/CURRENT` → `manifest/compute-manifest-v0_2.yaml`) to confirm the substrate identity matches what the harness was written against (`github.com/JamesPagetButler/qbp-compute-unit`).
2. **Builds a QADD program** of ≥1000 single-cycle instructions encoded for QBP-CU's RISC-V-derived ISA.
3. **Steps the emulator** instruction-by-instruction, capturing `CPU.Cycles` after each Step into an `InstructionEvent` slice.
4. **Validates the captured trace** against `Monotonic` + `AdvanceByOne` predicates — the Go-side paired definitions of `lean/Wyrd/SubstrateTrace.lean`'s Lean predicates.
5. **Emits a structured run log** to stdout (and optionally to `testdata/crawl-emulator-run.log`).
6. **Exit codes:**
   - `0` — both predicates hold; mode-(b) verification PASSES
   - `1` — predicate violation (Monotonic or AdvanceByOne); mode-(b) verification FAILS
   - `2` — manifest load / runtime error (paperwork bug; investigate manifest health)

## Run it

```bash
# Default: 1024 instructions, output to stdout
go run ./cmd/extract-cycle-counter-proof --root .

# Custom step count + persist log
go run ./cmd/extract-cycle-counter-proof --root . --steps 4096 --log /tmp/run.log

# As the federation CI workflow consumes it:
go run ./cmd/extract-cycle-counter-proof --root . --log cmd/extract-cycle-counter-proof/testdata/crawl-emulator-run.log
```

## Committed run log

`testdata/crawl-emulator-run.log` is a sample run committed at PR-open time. It documents:

- The Compute Manifest substrate identity at the time of capture
- The 1024-instruction trace's first/last cycle counter values
- Predicate verdicts (Monotonic + AdvanceByOne; both currently `true`)
- Verdict: `mode_b_eligible`

## SCOPE — single-cycle opcode subset (load-bearing limitation)

The QBP-CU emulator's `QROT` opcode increments `CPU.Cycles` by 2 per retired instruction ("Two QMULs" — composite operation per [`repo-qbp-compute-unit-pr-#33`](https://github.com/JamesPagetButler/qbp-compute-unit/pull/33) §5.4). This **violates the strict-equality `AdvanceByOne` predicate** (`cycle[i+1] = cycle[i] + 1`).

This harness uses **only the single-cycle opcode subset** (QADD; QMUL/QCONJ/QNORM/FANO are also single-cycle):

| Opcode | Funct7 | Cycles per retired instruction |
|---|---|---|
| QMUL | 0 | 1 |
| **QADD** | **1** | **1 — used by this harness** |
| QROT | 2 | 2 (excluded — composite "Two QMULs") |
| FANO | 3 | 1 |
| QCONJ | 4 | 1 |
| QNORM | 5 | 1 |

The exclusion is acknowledged in PR #63 §10 NOT-DECIDED:

> Cycle-counter monotonicity vs concurrent dispatch ... out of scope for v0.1; the `cpu.go canonical` resolution in repo-qbp-compute-unit-pr-#33 §5.4 is the Crawl-phase answer; Walk-α may require an A22 §4.2 amendment.

v0.2 theorem refinement has two candidate paths:

- **(option 1)** Relax `AdvanceByOne` to admit composite-op cycle accounting (`cycle[i+1] ≥ cycle[i] + 1`). Admits more substrates but tells consumers less about timing structure.
- **(option 2)** Add a per-instruction `cycle_cost` field on `InstructionEvent` so the predicate becomes `cycle[i+1] = cycle[i] + cycle_cost(opcode_i)`. Preserves the strict-equality contract per cost-class; consumers can reason about cycle-budget allocation per opcode.

Per @qbp-cu-implementor PR #67 §I4 read, **option 2 is the substrate-publisher-preferred path** (preserves strict-equality contract; admits composite ops without information loss). `cycle_cost` would be parameterized as `ComputeManifestPhase → Opcode → Nat` and documented per-phase per-opcode in the Compute Manifest `verified_invariants` schema slot (already reserved per PR #58 design §2.2). The current harness validates the substrate against the v0.1 strict-equality predicate using the documented single-cycle subset; full-coverage verification of every opcode is a v0.2 follow-up.

## Lean ↔ Go drift detection

The Lean predicate definitions in `lean/Wyrd/SubstrateTrace.lean` and the Go validator functions in `main.go` carry **paired doc-comments referencing each other**. The `drift_test.go` unit test computes SHA-256 of both files and compares against the committed snapshot at `testdata/lean-go-parity.snap`:

```
lean.SubstrateTrace          <sha256-hex>
go.extractCycleCounterProof  <sha256-hex>
```

CI runs the drift test on every PR touching either file. **Any modification to either side trips the test**, including comment edits. The forcing function is the load-bearing discipline — every change to the predicate-pairing region requires an explicit "I checked the other side" gesture from the developer.

### Regenerating the snapshot

When you intentionally update both paired sides in lockstep:

```bash
go test -run TestLeanGoParityDrift ./cmd/extract-cycle-counter-proof -args -update
```

The `-update` flag regenerates `testdata/lean-go-parity.snap` from current source. Commit the regenerated snapshot alongside the paired-source change.

## Federation-canonical use (C-PR-14 + beyond)

This harness produces mode-(b) verification **evidence** — the committed run log at `testdata/crawl-emulator-run.log` — that the promotion PR per Spec 9.2 §2 (C-PR-14; first-10 substrate-tier promotion per Spec 9.2 §9) cites as the mode-(b) verification artifact.

**The federation CI workflow does NOT invoke this harness directly.** Per @bma-implementor PR #67 §I4 read, the CI workflow (`.github/workflows/ci-compute-manifest.yml`, Phase B-PR-8 merged `e02e67a`) runs `cmd/mode-b-eligibility-check` only — which validates the **Compute Manifest credibility window** (Spec 9.2 §3.1 amendment), not the substrate-trace predicates this harness validates.

The actual mode-(b) verification chain is:

1. Developer runs this harness locally against the current substrate (`go run ./cmd/extract-cycle-counter-proof --root . --log testdata/crawl-emulator-run.log`)
2. Predicates validated; verdict logged; log file committed
3. CI runs `TestLeanGoParityDrift` → paired Lean predicate source ↔ Go validator source can't drift without an explicit `-update` ack gesture
4. CI runs `cmd/mode-b-eligibility-check` → Compute Manifest credibility window verified for the `mode-b-promotion`-labeled PR
5. C-PR-14 promotion PR cites the committed run log + the CI drift-test green + the credibility-window verdict as **combined mode-(b) evidence**
6. Beekeeper HVR (Spec 9.2 §9 first-10) weighs all three artifacts together

A future workflow refinement may add this harness to the CI gate directly (so the verification chain is fully automated rather than depending on developer-committed evidence). That's a Phase D+ refinement; the current evidence-based path is correct for v0.1 + the drift-test forcing function preserves correctness across the merge boundary.

## Related artifacts

- `lean/Wyrd/SubstrateTrace.lean` (PR #64) — Lean predicate source
- `lean/Wyrd/CycleCounterCrossPhase.lean` (PR #66) — substrate-tier theorem
- `doc/design/translation-functor-substrate-tier.md` (PR #63) — design surface
- `inter/spec/BMA-Spec-Addendum-9_2-Federation-Lean-Promotion-Protocol.md` (PR #6 amendment) — Spec 9.2 §3.1 substrate-credibility-window
- `model/compute_manifest.go` (PR #62) — `IsModeBEligible` predicate
- `.github/workflows/ci-compute-manifest.yml` (PR #65) — federation CI mode-(b) gate
