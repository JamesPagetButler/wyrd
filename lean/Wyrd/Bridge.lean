/-
  Wyrd/Bridge.lean

  Class B Phase 2 — Bridge promotion atomicity (count-preservation form).

  Helpful Engineering — Quaternion-Based Physics Programme
  April 2026 — Rev 0.1

  ============================================================
  PURPOSE
  ============================================================

  The Bridge layer (per Contextus spec §8) promotes Insight Signals
  from Contextus producers to CTH evaluators. Atomicity of this
  promotion is load-bearing for Class B correctness: a signal that
  is "halfway through" the bridge must not be observable by either
  side, lest it be evaluated twice or lost entirely.

  This file proves:

    (C-20c) bridge_promote_preserves_count:
      A bridge promotion that moves signal s from the Contextus
      queue to the CTH queue preserves the total signal count.
      Equivalently: signals are neither created nor destroyed by
      promotion; they only move between queues.

  ============================================================
  SCOPE NOTE
  ============================================================

  This is the "conservation" form of atomicity, not the full
  state-machine "no partial observable state" form. The conservation
  property is sufficient for Class B correctness claims (no signal
  loss, no signal duplication) and is provable directly from the
  promotion operation's definition.

  The full state-machine atomicity property (which requires modeling
  observers, transitions, and visibility) is deferred. The conservation
  property is the substantive content most consumers need; a future
  Phase 3 extension can add the observer model if production
  experience reveals gaps.
-/

import Mathlib.Data.Finset.Basic
import Mathlib.Data.Finset.Card
import Mathlib.Data.Finset.Insert
import Mathlib.Tactic.Linarith

namespace Wyrd
namespace Bridge

variable {Signal : Type*} [DecidableEq Signal]

/- ============================================================
   PART 1 — Bridge state model
   ============================================================ -/

/-- The Bridge state: signals partitioned across the Contextus queue
    (producer side) and the CTH queue (evaluator side). -/
structure State (Signal : Type*) [DecidableEq Signal] where
  contextusQueue : Finset Signal
  cthQueue : Finset Signal

/-- Total signal count visible to external observers. -/
def State.signalCount (b : State Signal) : ℕ :=
  b.contextusQueue.card + b.cthQueue.card

/- ============================================================
   PART 2 — Atomic promotion operation
   ============================================================ -/

/-- The atomic promotion operation: move signal s from contextusQueue
    to cthQueue. Defined only when s is in contextusQueue and not
    already in cthQueue (no duplicate-promotion). -/
def State.promote (b : State Signal) (s : Signal)
    (_h_in : s ∈ b.contextusQueue) (_h_out : s ∉ b.cthQueue) : State Signal :=
  { contextusQueue := b.contextusQueue.erase s
    cthQueue := insert s b.cthQueue }

@[simp] theorem State.promote_contextusQueue (b : State Signal) (s : Signal)
    (h_in : s ∈ b.contextusQueue) (h_out : s ∉ b.cthQueue) :
    (b.promote s h_in h_out).contextusQueue = b.contextusQueue.erase s := rfl

@[simp] theorem State.promote_cthQueue (b : State Signal) (s : Signal)
    (h_in : s ∈ b.contextusQueue) (h_out : s ∉ b.cthQueue) :
    (b.promote s h_in h_out).cthQueue = insert s b.cthQueue := rfl

/- ============================================================
   PART 3 — Theorem C-20c: promotion preserves signal count
   ============================================================ -/

/-- THEOREM C-20c: Bridge promotion preserves the total signal count.

    READING: signals are neither lost nor duplicated when the Bridge
    moves them from Contextus to CTH. The count visible to any external
    observer (sum of both queue cardinalities) is invariant under
    promotion.

    SECURITY / CORRECTNESS INTERPRETATION: this is the conservation
    form of atomicity. An attacker (or bug) cannot use promotion to
    inflate the signal count (creating phantom signals on the CTH side
    that never came from Contextus) or deflate it (dropping signals
    silently). Combined with the precondition (s ∈ contextusQueue,
    s ∉ cthQueue), this rules out the two failure modes that matter
    most for Class B integrity:

      1. Lost signal: post-promote, signal would be in neither queue
      2. Duplicated signal: post-promote, signal would be in both queues

    Both are excluded by the type signature plus this theorem. -/
theorem bridge_promote_preserves_count (b : State Signal) (s : Signal)
    (h_in : s ∈ b.contextusQueue) (h_out : s ∉ b.cthQueue) :
    (b.promote s h_in h_out).signalCount = b.signalCount := by
  unfold State.signalCount
  rw [State.promote_contextusQueue, State.promote_cthQueue]
  rw [Finset.card_erase_of_mem h_in]
  rw [Finset.card_insert_of_notMem h_out]
  -- Goal is now: (b.contextusQueue.card - 1) + (b.cthQueue.card + 1)
  --            = b.contextusQueue.card + b.cthQueue.card
  have h_card_pos : b.contextusQueue.card ≥ 1 := Finset.card_pos.mpr ⟨s, h_in⟩
  omega

/- ============================================================
   PART 4 — Corollaries
   ============================================================ -/

/-- COROLLARY: post-promote, the signal is in cthQueue. -/
theorem bridge_promote_signal_in_cth (b : State Signal) (s : Signal)
    (h_in : s ∈ b.contextusQueue) (h_out : s ∉ b.cthQueue) :
    s ∈ (b.promote s h_in h_out).cthQueue := by
  rw [State.promote_cthQueue]
  exact Finset.mem_insert_self s b.cthQueue

/-- COROLLARY: post-promote, the signal is NOT in contextusQueue. -/
theorem bridge_promote_signal_not_in_contextus (b : State Signal) (s : Signal)
    (h_in : s ∈ b.contextusQueue) (h_out : s ∉ b.cthQueue) :
    s ∉ (b.promote s h_in h_out).contextusQueue := by
  rw [State.promote_contextusQueue]
  intro h
  exact (Finset.mem_erase.mp h).1 rfl

/-- COMBINED COROLLARY: the signal is in EXACTLY ONE queue post-promotion.
    This is the "no partial state" claim in conservation form: the signal
    is in cth (yes) and not in contextus (yes) — never both, never neither. -/
theorem bridge_promote_exactly_one_side (b : State Signal) (s : Signal)
    (h_in : s ∈ b.contextusQueue) (h_out : s ∉ b.cthQueue) :
    s ∈ (b.promote s h_in h_out).cthQueue ∧
    s ∉ (b.promote s h_in h_out).contextusQueue :=
  ⟨bridge_promote_signal_in_cth b s h_in h_out,
   bridge_promote_signal_not_in_contextus b s h_in h_out⟩

/- ============================================================
   PART 5 — Status
   ============================================================

   PROVEN:
     ✓ State (Bridge state model) with two queues
     ✓ State.signalCount (observable total)
     ✓ State.promote (the atomic promotion operation)
     ✓ Promotion-side simp lemmas (contextusQueue, cthQueue)
     ✓ C-20c: bridge_promote_preserves_count
     ✓ bridge_promote_signal_in_cth (post-state in CTH)
     ✓ bridge_promote_signal_not_in_contextus (post-state not in Contextus)
     ✓ bridge_promote_exactly_one_side (combined: in one, not both)

   DEFERRED (Phase 3, not blocking):
     ◦ Full state-machine atomicity with explicit observers and
       step semantics. The conservation form proven here is the
       substantive content for Class B integrity; the state-machine
       form would add formal modeling of partial-state non-observability.

     ◦ Multi-signal batched promotion. Currently single-signal; batches
       are a downstream optimization (per B-OPT-6 in Workload-ISA v0.2).

     ◦ Failure / abort semantics. Currently we model successful atomic
       promotion only. Aborts that revert state would extend this with
       a complementary `abort` operation and conservation under abort.

   This file is COMPLETE for the C-20c conservation deliverable.
-/

end Bridge
end Wyrd
