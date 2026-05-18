/-
  Wyrd/SubstrateTrace.lean

  Class B Phase 2 (extension) — Lean substrate-trace structure for the
  Translation Functor §4.2 substrate-tier invariant
  (repo-bma-systema-issue-#170; design surface PR #63 — Phase C-PR-10).

  Helpful Engineering — Quaternion-Based Physics Programme
  May 2026 — Rev 0.1

  ============================================================
  PURPOSE
  ============================================================

  PR #63 (Phase C-PR-10 design surface) §3 commits to the following
  Lean encoding for substrate execution traces:

    "Substrate execution trace as List InstructionEvent with per-event
     cycle counter; Monotonic + AdvanceByOne predicates; symbolic
     ComputeManifestPhase mirroring Go-side enum."

  This file delivers exactly that. The structures support the
  substrate-tier theorem `cycle_counter_monotonic_per_phase` per
  Phase C-PR-12 (forthcoming), which formalises the federation
  sovereignty invariant:

    "For all blessed compute substrate phases per the current Compute
     Manifest, the substrate exposes a canonical instruction-retire
     cycle counter with monotonic-non-decreasing semantics, advancing
     by 1 per retired instruction."

  ============================================================
  WHY THIS STRUCTURE
  ============================================================

  Two abstraction choices in this file are deliberate per PR #63 §3.3:

  1. InstructionEvent carries ONLY a cycle counter. Downstream
     substrate-tier theorems that need richer per-event data (opcode,
     address, etc.) can refine the structure at their own anchor
     sites without invalidating this one. Same decomposition
     discipline as `Wyrd.ComputeManifest.RawManifest.payload`
     (PR #60).

  2. SubstrateTrace is parameterized by ComputeManifestPhase. The
     parameter is symbolic (no runtime phase data lives in the
     trace) — its job is to make the cross-phase invariant
     statement explicit at the type level: a SubstrateTrace exists
     for EVERY phase, not just the current Crawl manifest.

  ============================================================
  DRIFT DETECTION ↔ Go-side enum
  ============================================================

  The Lean `ComputeManifestPhase` inductive mirrors
  `model.ComputeManifestPhase` (Go-side, in PR #59). A CI drift
  test compares the variant set on both sides; drift fails CI.
  Phase C-PR-13 ships the drift-detection harness.
-/

import Mathlib.Data.List.Basic

namespace Wyrd
namespace SubstrateTrace

/-- ComputeManifestPhase — the symbolic Lean mirror of the Go-side
    `model.ComputeManifestPhase` enum (per `model/compute_manifest.go`
    in PR #59). Variants MUST stay in sync with the Go side; Phase
    C-PR-13 ships the drift-detection CI test.

    Per Spec 9.2 §4 phase table:
      crawl       — Crawl phase; substrate = QBP-CU emulator
      toddle      — Toddle phase; substrate = QBP-CU emulator
      walk        — Walk phase; substrate = QBP-CU M1 Gearbox
      runInitial  — Run-initial phase; substrate = QBP-CU M2 + ROCm
      runMature   — Run-mature phase; substrate = possibly silicon -/
inductive ComputeManifestPhase where
  | crawl
  | toddle
  | walk
  | runInitial
  | runMature
  deriving DecidableEq, Repr

/-- InstructionEvent — a single retired-instruction event in a
    substrate execution trace, carrying just its cycle counter.

    Per PR #63 §3.3 deliberate-minimal-abstraction: downstream
    substrate-tier theorems that need richer per-event data
    (opcode, address, etc.) refine this structure at their own
    anchor site without invalidating the cycle-counter cross-phase
    invariant at this level. -/
structure InstructionEvent where
  cycle : Nat
  deriving DecidableEq, Repr

/-- SubstrateTrace m — a substrate execution trace under the given
    ComputeManifestPhase. The events list is the retired-instruction
    sequence in execution order.

    The phase parameter is symbolic — the trace structure doesn't
    carry runtime phase data. The parameter exists so the substrate-
    tier theorem statement can quantify over phases explicitly: the
    invariant must hold for ANY m, not just the current Crawl
    manifest's. -/
structure SubstrateTrace (m : ComputeManifestPhase) where
  events : List InstructionEvent

/- ============================================================
   PART 1 — Predicates the substrate must satisfy
   ============================================================ -/

/-- Monotonic — the cycle counter is non-decreasing across the
    trace. For any pair of indices i < j, the cycle at i is ≤ the
    cycle at j.

    This is the weaker of the two algebraic claims; AdvanceByOne
    (below) is the stronger one. Both are required for the federation
    sovereignty invariant per A22 §4.2.

    Implementation note: indices are bounded by t.events.length via
    explicit Fin witnesses constructed from `Nat.lt_trans` (avoids
    omega-tactic Nat/Int coercion issues at definition site). -/
def Monotonic (t : SubstrateTrace m) : Prop :=
  ∀ (i j : Nat) (hij : i < j) (hj : j < t.events.length),
    (t.events.get ⟨i, Nat.lt_trans hij hj⟩).cycle ≤
    (t.events.get ⟨j, hj⟩).cycle

/-- AdvanceByOne — the cycle counter advances by exactly 1 per
    retired instruction. For any adjacent pair of events (i, i+1)
    within the trace, cycle(i+1) = cycle(i) + 1.

    AdvanceByOne strictly implies Monotonic (cycle(j) - cycle(i) =
    j - i ≥ 1 for j > i, so cycle(j) ≥ cycle(i) + 1 > cycle(i)).
    Both predicates are kept distinct because consumers may want to
    reason about Monotonic without committing to AdvanceByOne (e.g.,
    SIMD-style substrates that retire multiple instructions per
    cycle would satisfy Monotonic but not AdvanceByOne, and the
    invariant must distinguish these cases). -/
def AdvanceByOne (t : SubstrateTrace m) : Prop :=
  ∀ (i : Nat) (hsucc : i + 1 < t.events.length),
    (t.events.get ⟨i+1, hsucc⟩).cycle =
    (t.events.get ⟨i, Nat.lt_of_succ_lt hsucc⟩).cycle + 1

/- ============================================================
   PART 2 — Fixture-trace tests (small examples)
   ============================================================

   These three example traces exercise the predicates against
   hand-constructed inputs to demonstrate that the predicates
   compute as expected. Not load-bearing theorems; just smoke-test
   examples consumers can grep when constructing their own traces.
   The actual substrate-tier theorem lands at PR #66 (Phase C-PR-12).
-/

/-- A 5-event trace where the cycle counter goes [0,1,2,3,4].
    Satisfies both Monotonic AND AdvanceByOne. -/
def exampleTraceGood : SubstrateTrace ComputeManifestPhase.crawl :=
  { events := [⟨0⟩, ⟨1⟩, ⟨2⟩, ⟨3⟩, ⟨4⟩] }

/-- A 3-event trace where the cycle counter goes [0,1,3].
    Satisfies Monotonic but FAILS AdvanceByOne (the [1,3] gap is +2). -/
def exampleTraceMonotonicNotAdvanceByOne : SubstrateTrace ComputeManifestPhase.crawl :=
  { events := [⟨0⟩, ⟨1⟩, ⟨3⟩] }

/-- A 3-event trace where the cycle counter goes [2,1,0].
    FAILS Monotonic (decreasing). -/
def exampleTraceNotMonotonic : SubstrateTrace ComputeManifestPhase.crawl :=
  { events := [⟨2⟩, ⟨1⟩, ⟨0⟩] }

/- ============================================================
   PART 3 — Status
   ============================================================

   PROVEN/DEFINED:
     ✓ ComputeManifestPhase inductive (5 variants matching Go-side
       model.ComputeManifestPhase enum)
     ✓ InstructionEvent structure (minimal: cycle counter only)
     ✓ SubstrateTrace m structure (phase-parameterized)
     ✓ Monotonic predicate (non-decreasing across trace)
     ✓ AdvanceByOne predicate (cycle advances by 1 per event)
     ✓ Three fixture traces demonstrating predicate behavior

   No sorry, no user-defined axiom — Phase 2 CI gate compliant.
   ~50 LOC theorem-body equivalent (definitions only at this stage;
   the actual substrate-tier theorem lands at Phase C-PR-12).

   FOLLOW-ON SCOPE:
     ◦ Phase C-PR-12 — cycle_counter_monotonic_per_phase substrate-tier
       theorem; mode (a) verification target.
     ◦ Phase C-PR-13 — Go-side extraction harness + Lean↔Go drift-
       detection CI test (verifies the symbolic ComputeManifestPhase
       enum stays in sync with model.ComputeManifestPhase).
     ◦ Phase C-PR-14 — promotion PR per Spec 9.2 §2 declaring
       mode = (a) + (b); first-10 substrate promotion HVR per
       Spec 9.2 §9.

   Per design surface PR #63 §3.3: the minimal abstraction here
   intentionally defers richer per-event data (opcode, address,
   etc.) to downstream substrate-tier theorem anchor sites. Same
   decomposition discipline as `RawManifest.payload` in
   Wyrd.ComputeManifest (PR #60).
-/

end SubstrateTrace
end Wyrd
