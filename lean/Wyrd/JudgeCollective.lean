/-
  Wyrd/JudgeCollective.lean

  Class C Phase 3 — Judge collective determinism + vote aggregation.

  Helpful Engineering — Quaternion-Based Physics Programme
  April 2026 — Rev 0.1

  ============================================================
  PURPOSE
  ============================================================

  BMA's governance uses a judge collective: when a proposal needs
  evaluation (especially constitutional changes — see Constitutional.lean
  for C-21d), each judge in the collective votes APPROVE / MINOR_CONCERN /
  MAJOR_CONCERN / VETO, and the votes aggregate into a collective result.

  This file proves:

    (C-21c) judge_collective_deterministic:
      Same proposal + same context + same judges yields the same
      collective vote (judges are pure deterministic functions, vote
      aggregation is order-independent).

  Plus the supporting structural results:
    - Vote aggregation is commutative (order-of-evaluation invariant)
    - Vote aggregation is associative (parenthesization invariant)
    - Vote aggregation respects vote priority (VETO absorbs, then
      MAJOR > MINOR > APPROVE in escalation)

  ============================================================
  MODEL
  ============================================================

  - Proposal: a candidate decision being evaluated (abstract; could be
    code update, architectural change, capability grant, etc.)
  - Context: world state at evaluation time (abstract; encoded as a
    hash for our determinism proofs)
  - Vote: the four-valued judgment (APPROVE / MINOR / MAJOR / VETO)
  - Judge: a pure function (Proposal → Context → Vote)
  - Aggregate: a binary operation combining two votes

  The aggregate forms a commutative semigroup with VETO as absorbing,
  APPROVE as identity. Determinism follows from purity of judges plus
  algebraic properties of the aggregate.
-/

import Mathlib.Data.List.Basic

namespace Wyrd
namespace JudgeCollective

/- ============================================================
   PART 1 — Proposal, Context, Vote
   ============================================================ -/

/-- A proposal under evaluation by the judge collective. The token
    is opaque; what matters for our determinism proof is identity. -/
structure Proposal where
  proposalId : Nat
  deriving DecidableEq, Repr

/-- The context at evaluation time. Abstracted as a hash that
    captures all relevant world state. Two evaluations in the same
    context are by definition evaluating against the same state. -/
structure Context where
  contextHash : Nat
  deriving DecidableEq, Repr

/-- The four-valued vote outcome.

    Priority order (most-restrictive to least-restrictive):
      VETO > MAJOR_CONCERN > MINOR_CONCERN > APPROVE

    The aggregation rule (see `aggregate` below) takes the more
    restrictive of two votes. -/
inductive Vote : Type where
  | approve
  | minorConcern
  | majorConcern
  | veto
  deriving DecidableEq, Repr

/- ============================================================
   PART 2 — Judges as pure functions
   ============================================================ -/

/-- A judge is a pure function from (proposal, context) to vote.
    Purity is the key property: same input → same output, no
    hidden state, no I/O effects. -/
def Judge := Proposal → Context → Vote

/- ============================================================
   PART 3 — Vote aggregation as a commutative monoid
   ============================================================ -/

/-- Aggregate two votes — take the more restrictive.
    VETO absorbs (any vote with VETO becomes VETO).
    APPROVE is identity (APPROVE with anything else is the other vote). -/
def aggregate : Vote → Vote → Vote
  | Vote.veto,         _                  => Vote.veto
  | _,                 Vote.veto          => Vote.veto
  | Vote.majorConcern, _                  => Vote.majorConcern
  | _,                 Vote.majorConcern  => Vote.majorConcern
  | Vote.minorConcern, _                  => Vote.minorConcern
  | _,                 Vote.minorConcern  => Vote.minorConcern
  | Vote.approve,      Vote.approve       => Vote.approve

/-- Aggregate is commutative — order of two votes doesn't matter. -/
theorem aggregate_comm (v1 v2 : Vote) : aggregate v1 v2 = aggregate v2 v1 := by
  cases v1 <;> cases v2 <;> rfl

/-- Aggregate is associative — parenthesization doesn't matter. -/
theorem aggregate_assoc (v1 v2 v3 : Vote) :
    aggregate (aggregate v1 v2) v3 = aggregate v1 (aggregate v2 v3) := by
  cases v1 <;> cases v2 <;> cases v3 <;> rfl

/-- APPROVE is the identity for aggregation (left). -/
theorem aggregate_approve_left (v : Vote) : aggregate Vote.approve v = v := by
  cases v <;> rfl

/-- APPROVE is the identity for aggregation (right). -/
theorem aggregate_approve_right (v : Vote) : aggregate v Vote.approve = v := by
  cases v <;> rfl

/-- VETO is absorbing for aggregation (left). -/
theorem aggregate_veto_left (v : Vote) : aggregate Vote.veto v = Vote.veto := by
  cases v <;> rfl

/-- VETO is absorbing for aggregation (right). -/
theorem aggregate_veto_right (v : Vote) : aggregate v Vote.veto = Vote.veto := by
  cases v <;> rfl

/-- Left-commutativity: a · (b · c) = b · (a · c). Combines comm + assoc. -/
theorem aggregate_left_comm (a b c : Vote) :
    aggregate a (aggregate b c) = aggregate b (aggregate a c) := by
  rw [← aggregate_assoc, aggregate_comm a b, aggregate_assoc]

/- ============================================================
   PART 4 — Running the judge collective
   ============================================================ -/

/-- Run the judge collective on a proposal in a given context.
    Each judge produces a vote; aggregate folds them into a
    collective vote. The fold starts from APPROVE (identity). -/
def run (judges : List Judge) (p : Proposal) (ctx : Context) : Vote :=
  judges.foldr (fun j acc => aggregate (j p ctx) acc) Vote.approve

/- ============================================================
   PART 5 — Theorem C-21c: collective determinism
   ============================================================ -/

/-- THEOREM C-21c (basic form): the collective vote is a deterministic
    function of (judges, proposal, context).

    READING: same judges + same proposal + same context = same vote.
    No hidden state, no time-of-day variance, no inter-evaluation
    dependence. The collective is a pure function.

    This is structurally trivial in Lean (functions ARE deterministic);
    the substantive content is proving the *more interesting* corollary
    that the order of judges doesn't matter (commutative under
    permutation), which we prove next. -/
theorem judge_collective_deterministic
    (judges : List Judge) (p : Proposal) (ctx : Context) :
    run judges p ctx = run judges p ctx := rfl

/-- THEOREM C-21c (substantive form): the collective vote is invariant
    under permutation of the judge list.

    READING: order doesn't matter. Whether the judges are queried
    sequentially, in parallel, or in some hand-crafted order, the
    final aggregate vote is the same. This certifies the optimization
    C-OPT-5 (parallel judge evaluation): we can fan out judge queries
    in any order without affecting the outcome.

    PROOF STRATEGY: aggregate is a commutative monoid (proven above),
    so foldr over a permuted list yields the same result. -/
theorem judge_collective_perm_invariant
    (j1 j2 : List Judge) (p : Proposal) (ctx : Context)
    (h_perm : j1.Perm j2) :
    run j1 p ctx = run j2 p ctx := by
  unfold run
  induction h_perm with
  | nil => rfl
  | cons x _ ih => simp [List.foldr_cons]; rw [ih]
  | swap x y l =>
    -- swap two adjacent judges in the fold
    -- LHS after simp: aggregate (y p ctx) (aggregate (x p ctx) (foldr l))
    -- RHS:            aggregate (x p ctx) (aggregate (y p ctx) (foldr l))
    -- Connected by left-commutativity of aggregate.
    simp [List.foldr_cons]
    exact aggregate_left_comm (y p ctx) (x p ctx) _
  | trans _ _ ih1 ih2 => exact ih1.trans ih2

/-- COROLLARY: the empty judge list trivially returns APPROVE. -/
theorem judge_collective_empty (p : Proposal) (ctx : Context) :
    run [] p ctx = Vote.approve := rfl

/-- COROLLARY: a single judge's vote is the collective vote. -/
theorem judge_collective_singleton (j : Judge) (p : Proposal) (ctx : Context) :
    run [j] p ctx = j p ctx := by
  unfold run
  simp [List.foldr_cons, List.foldr_nil]
  exact aggregate_approve_right (j p ctx)

/-- COROLLARY: if any judge votes VETO, the collective votes VETO.

    SECURITY INTERPRETATION: a single VETO is sufficient to block.
    This is the "any judge can stop" property — required for the
    constitutional protection where any judge can prevent a
    self-modification (per Theory v2.0 §judge collective). -/
theorem judge_collective_veto_propagates
    (judges : List Judge) (p : Proposal) (ctx : Context)
    (j : Judge) (h_in : j ∈ judges) (h_veto : j p ctx = Vote.veto) :
    run judges p ctx = Vote.veto := by
  induction judges with
  | nil => exact absurd h_in List.not_mem_nil
  | cons head rest ih =>
    simp [run, List.foldr_cons]
    rcases List.mem_cons.mp h_in with rfl | h_in_rest
    · -- j is the head; head p ctx = veto
      rw [h_veto]
      exact aggregate_veto_left _
    · -- j is in the rest; recursive case
      have h_rest : run rest p ctx = Vote.veto := ih h_in_rest
      simp [run] at h_rest
      rw [h_rest]
      exact aggregate_veto_right _

/- ============================================================
   PART 6 — Status
   ============================================================

   PROVEN:
     ✓ Proposal, Context, Vote types
     ✓ Judge as a pure function type
     ✓ aggregate operation (the four-valued lattice)
     ✓ aggregate_comm: commutativity
     ✓ aggregate_assoc: associativity
     ✓ aggregate_approve_left/right: APPROVE is identity
     ✓ aggregate_veto_left/right: VETO is absorbing
     ✓ run: judge-collective folding
     ✓ C-21c (basic): judge_collective_deterministic
     ✓ C-21c (substantive): judge_collective_perm_invariant — order doesn't matter
     ✓ judge_collective_empty: empty list trivially APPROVE
     ✓ judge_collective_singleton: single judge = collective
     ✓ judge_collective_veto_propagates: any VETO blocks the collective

   DEFERRED (not blocking):
     ◦ Domain-weighted voting (per CLAUDE.md, judges have domain weights;
       our model treats all judges equal-weight; weights are a refinement)
     ◦ Time-out semantics (a judge that doesn't respond — currently we
       require all judges to vote; production may need timeout handling)
     ◦ Vote caching (per C-OPT-7; orthogonal to determinism, can layer)

   This file is COMPLETE for the C-21c deliverable, plus exports the
   judge / vote / run / aggregate primitives that Constitutional.lean
   uses for C-21d.
-/

end JudgeCollective
end Wyrd
