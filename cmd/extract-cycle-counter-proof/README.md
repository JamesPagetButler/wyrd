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

v0.2 theorem refinement may relax `AdvanceByOne` to admit composite-op cycle accounting (or admit a per-instruction `cycle_cost` field on `InstructionEvent`). The current harness validates the substrate against the v0.1 strict-equality predicate using the documented subset; full-coverage verification of every opcode is a v0.2 follow-up.

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

The promotion PR per Spec 9.2 §2 (C-PR-14; first-10 substrate-tier promotion per Spec 9.2 §9) cites this harness's run log as the mode-(b) verification evidence. The federation CI workflow (`.github/workflows/ci-compute-manifest.yml`, Phase B-PR-8 merged `e02e67a`) exercises this binary on any PR labeled `mode-b-promotion`; passing CI = mode-(b) verification clean against the current Compute Manifest substrate.

## Related artifacts

- `lean/Wyrd/SubstrateTrace.lean` (PR #64) — Lean predicate source
- `lean/Wyrd/CycleCounterCrossPhase.lean` (PR #66) — substrate-tier theorem
- `doc/design/translation-functor-substrate-tier.md` (PR #63) — design surface
- `inter/spec/BMA-Spec-Addendum-9_2-Federation-Lean-Promotion-Protocol.md` (PR #6 amendment) — Spec 9.2 §3.1 substrate-credibility-window
- `model/compute_manifest.go` (PR #62) — `IsModeBEligible` predicate
- `.github/workflows/ci-compute-manifest.yml` (PR #65) — federation CI mode-(b) gate
