/-
  Wyrd/HolographicHypergraphQuaternion.lean

  PROT-HH-001 Theorem 2ℍ — Formal proof of multi-beam irreducibility for
  quaternion-valued (polarisation-state) recordings.

  Helpful Engineering — Quaternion-Based Physics Programme
  May 2026 — Rev 0.1 (Wyrd corpus v1.5 candidate)

  ============================================================
  PURPOSE
  ============================================================

  Lifts the ℝ-valued Theorem 2 (HolographicHypergraph.lean) to the
  quaternion case. The composition law for polarisation states is
  multiplicative (Jones calculus): when beam A's polarisation is rotated
  to beam B and then to beam C, the cumulative rotation is q_AB · q_BC,
  not q_AB + q_BC. The triangle constraint becomes

      q_AC = q_AB · q_BC

  using the (non-commutative) Hamilton product. The embedding

      tripleToPairsH : (q_AB, q_BC) ↦ (q_AB, q_AB · q_BC, q_BC)

  enforces the multiplicative constraint on its image. Theorem 2ℍ
  shows the embedding is not surjective: some IndepPairsH configuration
  violates the constraint and cannot arise from any triple.

  ============================================================
  WITNESS
  ============================================================

  Counterexample: q_AB = i, q_AC = i, q_BC = j. Any preimage triple
  would force i · j = i, but i · j = k (Hamilton), and k ≠ i because
  k.imK = 1 while i.imK = 0. The numerical-witness companion at
  BMA/projects/holographic-hypergraph/quaternion-witness/ confirms

      ‖i − (i·j)‖² = 2 exactly across QBP-CU emulator widths W64..W512.

  This exactness is structural (i·j = k for any commutative-ring
  base), so the witness lifts cleanly to Lean over Quaternion ℝ.

  ============================================================
  CONNECTION TO THE WYRD CORPUS
  ============================================================

  Companion to:
    - Wyrd.HolographicHypergraph (the ℝ-valued Phase 4 v1.4 file).
    - PROT-HH-001 PRED-HH-09 (polarisation-dependent write speed
      from quaternion-norm scaling of VSG).
    - PROT-HH-002 §3 consequence 2 ("algebraic privilege is physical")
      Level 1 substrate.

  Witness verified numerically on QBP-CU emulator: see
    BMA/projects/holographic-hypergraph/quaternion-witness/main.go.

  See `Wyrd-Proofs-Reference-v1.5.md` (forthcoming) §31 for the corpus map.
-/

import Mathlib.Algebra.Quaternion
import Mathlib.Data.Complex.Basic
import Mathlib.Tactic.NormNum

namespace Wyrd
namespace HolographicHypergraphQuaternion

/- ============================================================
   PART 1 — Quaternion-valued recording types
   ============================================================ -/

/-- A triple-coherent recording with quaternion polarisation states.
    Two states determine the third via the multiplicative triangle
    constraint q_AC = q_AB · q_BC; only two degrees of freedom. -/
structure TripleCoherentH where
  /-- Polarisation rotation from beam A to beam B. -/
  qAB : Quaternion ℝ
  /-- Polarisation rotation from beam B to beam C. -/
  qBC : Quaternion ℝ

/-- Three independent pair recordings with quaternion polarisation
    rotations. Three independent reference frames; the recovered
    rotations need not satisfy q_AC = q_AB · q_BC. Three DOF. -/
structure IndepPairsH where
  /-- Polarisation rotation recovered from the (A, B) pair recording. -/
  qAB : Quaternion ℝ
  /-- Polarisation rotation recovered from the (A, C) pair recording. -/
  qAC : Quaternion ℝ
  /-- Polarisation rotation recovered from the (B, C) pair recording. -/
  qBC : Quaternion ℝ

/- ============================================================
   PART 2 — The embedding tripleToPairsH
   ============================================================ -/

/-- The natural embedding: a triple-coherent recording induces a
    consistent pairwise representation, with q_AC = q_AB · q_BC. -/
def tripleToPairsH (tc : TripleCoherentH) : IndepPairsH :=
  { qAB := tc.qAB
    qAC := tc.qAB * tc.qBC
    qBC := tc.qBC }

@[simp] theorem tripleToPairsH_qAB (tc : TripleCoherentH) :
    (tripleToPairsH tc).qAB = tc.qAB := rfl

@[simp] theorem tripleToPairsH_qAC (tc : TripleCoherentH) :
    (tripleToPairsH tc).qAC = tc.qAB * tc.qBC := rfl

@[simp] theorem tripleToPairsH_qBC (tc : TripleCoherentH) :
    (tripleToPairsH tc).qBC = tc.qBC := rfl

/-- Multiplicative triangle predicate: an IndepPairsH satisfies the
    consistency constraint when q_AC = q_AB · q_BC. -/
def IndepPairsH.IsConsistent (ip : IndepPairsH) : Prop :=
  ip.qAC = ip.qAB * ip.qBC

/-- Every triple-coherent recording produces a consistent pairwise
    representation by construction. -/
theorem tripleToPairsH_consistent (tc : TripleCoherentH) :
    (tripleToPairsH tc).IsConsistent := by
  unfold IndepPairsH.IsConsistent tripleToPairsH
  rfl

/- ============================================================
   PART 3 — Theorem 2ℍ: tripleToPairsH is NOT surjective
   ============================================================ -/

/-- THEOREM 2ℍ (PROT-HH-001 §3.1, quaternion extension): a quaternion-
    valued 3-beam coherent recording is NOT equivalent to three
    independent quaternion pair recordings.

    Formally: the embedding `tripleToPairsH` is not surjective. The
    independent-pair configuration ⟨i, i, j⟩ (q_AB = i, q_AC = i,
    q_BC = j) cannot arise from any triple, since `tripleToPairsH`
    would force i · j = i, but in Quaternion ℝ we have i · j = k ≠ i
    (the imK component disagrees: k.imK = 1, i.imK = 0).

    The witness counterexample is verified numerically on the QBP-CU
    emulator at widths W64..W512 — see the project README at
    BMA/projects/holographic-hypergraph/quaternion-witness/. -/
theorem theorem2_irreducibility_quaternion :
    ¬ Function.Surjective tripleToPairsH := by
  intro h_surj
  -- Witness: ⟨i, i, j⟩ in IndepPairsH (q_AB = i, q_AC = i, q_BC = j).
  obtain ⟨tc, h_tc⟩ := h_surj
    { qAB := ⟨0, 1, 0, 0⟩, qAC := ⟨0, 1, 0, 0⟩, qBC := ⟨0, 0, 1, 0⟩ }
  -- h_tc : tripleToPairsH tc = ⟨i, i, j⟩.
  simp only [tripleToPairsH, IndepPairsH.mk.injEq] at h_tc
  obtain ⟨h_AB, h_AC, h_BC⟩ := h_tc
  -- h_AB : tc.qAB = i,  h_AC : tc.qAB * tc.qBC = i,  h_BC : tc.qBC = j.
  rw [h_AB, h_BC] at h_AC
  -- h_AC : (⟨0,1,0,0⟩ : Quaternion ℝ) * ⟨0,0,1,0⟩ = ⟨0,1,0,0⟩
  -- Apply imK; LHS reduces to 1 (i·j = k), RHS to 0.
  have h_imK := congrArg (·.imK) h_AC
  simp only [Quaternion.imK_mul] at h_imK
  norm_num at h_imK

/- ============================================================
   PART 4 — Image characterisation and injectivity
   ============================================================ -/

/-- The image of `tripleToPairsH` is exactly the multiplicative-
    consistency subspace. -/
theorem tripleToPairsH_image (ip : IndepPairsH) :
    (∃ tc : TripleCoherentH, tripleToPairsH tc = ip) ↔ ip.IsConsistent := by
  constructor
  · rintro ⟨tc, rfl⟩
    exact tripleToPairsH_consistent tc
  · intro h_consistent
    refine ⟨{ qAB := ip.qAB, qBC := ip.qBC }, ?_⟩
    unfold tripleToPairsH IndepPairsH.IsConsistent at *
    cases ip
    simp_all

/-- `tripleToPairsH` is INJECTIVE: different triples produce different
    pairwise representations. The qAB and qBC components are preserved
    coordinate-wise; equality on the image forces equality of the
    preimage's two free coordinates. -/
theorem tripleToPairsH_injective : Function.Injective tripleToPairsH := by
  intro tc1 tc2 h_eq
  simp only [tripleToPairsH, IndepPairsH.mk.injEq] at h_eq
  obtain ⟨h_AB, _, h_BC⟩ := h_eq
  cases tc1; cases tc2
  congr

/-- COMBINED RESULT: `tripleToPairsH` is injective but NOT surjective.
    Quaternion-valued triple recordings and independent-pair recordings
    are formally information-distinct, just as the ℝ-valued case
    (companion theorem `Wyrd.HolographicHypergraph.tripleToPairs_inj_not_surj`). -/
theorem tripleToPairsH_inj_not_surj :
    Function.Injective tripleToPairsH ∧ ¬ Function.Surjective tripleToPairsH :=
  ⟨tripleToPairsH_injective, theorem2_irreducibility_quaternion⟩

/- ============================================================
   PART 5 — Status and integration
   ============================================================

   PROVEN:
     ✓ TripleCoherentH and IndepPairsH structures (quaternion-valued)
     ✓ tripleToPairsH embedding with simp lemmas for projections
     ✓ tripleToPairsH_consistent — every triple's image satisfies
       the multiplicative triangle constraint
     ✓ THEOREM 2ℍ: theorem2_irreducibility_quaternion — embedding
       is NOT surjective; ⟨i, i, j⟩ is a structural witness
     ✓ tripleToPairsH_image — image is exactly the consistency subspace
     ✓ tripleToPairsH_injective — no information lost in the embedding
     ✓ tripleToPairsH_inj_not_surj — combined inj-but-not-surj

   DEFERRED (Phase 4 follow-ups beyond v1.5):
     ◦ Higher-arity n-beam ↛ (n-1)-beam (separate ticket).
     ◦ Information-theoretic codimension form (Sprint phase).
     ◦ Polarisation-norm-dependent VSG scaling (PRED-HH-09 Level 2+).

   The quaternion-extended core load-bearing claim of PROT-HH-001 §3.1
   is now formally established in Lean. PROT-HH-002 v0.2 should cite
   this theorem as Level 1 substrate for consequence 2 of the
   algebra-as-schema thesis.

   COMPANION DEMONSTRATION:
     `BMA/projects/holographic-hypergraph/quaternion-witness/main.go`
     verifies the witness numerically on QBP-CU at widths W64..W512.
     The exact match across all precisions (residual norm² = 2.0 to
     60+ digits, identical at every tier) is the Level 2 signal that
     the witness is structural, not floating-point.
-/

end HolographicHypergraphQuaternion
end Wyrd
