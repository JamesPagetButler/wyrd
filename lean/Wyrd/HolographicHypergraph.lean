/-
  Wyrd/HolographicHypergraph.lean

  PROT-HH-001 Theorem 2 — Formal proof of multi-beam irreducibility.

  Helpful Engineering — Quaternion-Based Physics Programme
  April 2026 — Rev 0.1

  ============================================================
  PURPOSE
  ============================================================

  Hardens the load-bearing claim of PROT-HH-001 (Holographic Hypergraph
  Storage): a 3-beam coherent recording is NOT equivalent to three
  independent pairwise recordings, even in a fully linear-response medium.

  The irreducibility resides in PHASE COHERENCE: a 3-beam recording
  binds three relative phases (φ₁₂, φ₂₃, φ₁₃) under the constraint
  φ₁₃ = φ₁₂ + φ₂₃ (triangle consistency). Three independent pair
  recordings have three independent absolute phase references; the three
  relative phases recovered from them need not satisfy this triangle
  consistency.

  Companion files:
    - sim_theorem2.py at BMA/projects/holographic-hypergraph/
        Numerical demonstration over a sweep of phase-drift magnitudes.
    - PROT-HH-002-qbp-algebra-as-schema.md (forthcoming)
        Stratified theory protocol citing this file as Level 1 evidence.

  ============================================================
  FORMAL CLAIM
  ============================================================

  Define the natural embedding
      tripleToPairs : TripleCoherent → IndepPairs
  that maps a 3-beam recording (parametrised by 2 relative phases) to its
  pairwise representation (3 relative phases, with the constraint
  phase13 = phase12 + phase23).

  Theorem 2 (irreducibility): tripleToPairs is NOT surjective. There
  exist pairwise recording configurations that are NOT the image of any
  triple-coherent recording — namely, those that violate the triangle
  consistency.

  Corollaries proven in this file:
    * The image of tripleToPairs is exactly the consistency subspace.
    * tripleToPairs is injective (different triples → different pairs).
    * The two encodings are *information-distinct* — neither is a
      faithful surrogate for the other.

  ============================================================
  CONNECTION TO THE WYRD CORPUS
  ============================================================

  This file extends the Wyrd Lean corpus from Phase 1-3 (algebraic
  privilege, Class B hypergraph, Class C operational) to Phase 4
  (physical instantiation theorems). Theorem 2 is the formal core of
  PROT-HH-002 Level 1: "QBP algebra-as-schema is *expressible* and
  enforces non-equivalent encodings."

  See `Wyrd-Proofs-Reference-v1.3.md` for the corpus map.
-/

import Mathlib.Analysis.SpecialFunctions.Trigonometric.Basic
import Mathlib.Tactic.Linarith
import Mathlib.Tactic.NormNum

namespace Wyrd
namespace HolographicHypergraph

/- ============================================================
   PART 1 — Triple-coherent and independent-pair recording types
   ============================================================ -/

/-- A triple-coherent recording captures three beams in a single
    phase-referenced event. The third relative phase (φ₁₃) is
    determined by the first two: φ₁₃ = φ₁₂ + φ₂₃ (triangle
    consistency). Hence only 2 degrees of freedom. -/
structure TripleCoherent where
  /-- Relative phase between beams 1 and 2 (φ₁ - φ₂). -/
  phase12 : ℝ
  /-- Relative phase between beams 2 and 3 (φ₂ - φ₃). -/
  phase23 : ℝ

/-- An independent-pair recording is three separate pair recordings,
    each with its own phase reference. No constraint among the three
    relative phases. Hence 3 degrees of freedom. -/
structure IndepPairs where
  /-- Relative phase recovered from the (1,2) pair recording. -/
  phase12 : ℝ
  /-- Relative phase recovered from the (1,3) pair recording. -/
  phase13 : ℝ
  /-- Relative phase recovered from the (2,3) pair recording. -/
  phase23 : ℝ

/- ============================================================
   PART 2 — The embedding tripleToPairs
   ============================================================ -/

/-- The natural embedding: a triple-coherent recording induces a
    consistent pairwise representation, with φ₁₃ = φ₁₂ + φ₂₃. -/
def tripleToPairs (tc : TripleCoherent) : IndepPairs :=
  { phase12 := tc.phase12
    phase13 := tc.phase12 + tc.phase23
    phase23 := tc.phase23 }

@[simp] theorem tripleToPairs_phase12 (tc : TripleCoherent) :
    (tripleToPairs tc).phase12 = tc.phase12 := rfl

@[simp] theorem tripleToPairs_phase13 (tc : TripleCoherent) :
    (tripleToPairs tc).phase13 = tc.phase12 + tc.phase23 := rfl

@[simp] theorem tripleToPairs_phase23 (tc : TripleCoherent) :
    (tripleToPairs tc).phase23 = tc.phase23 := rfl

/-- The image of `tripleToPairs` is exactly the triangle-consistency
    subspace: pairs satisfying phase13 = phase12 + phase23. -/
def IndepPairs.IsConsistent (ip : IndepPairs) : Prop :=
  ip.phase13 = ip.phase12 + ip.phase23

/-- Every triple-coherent recording produces a consistent pairwise
    representation by construction. -/
theorem tripleToPairs_consistent (tc : TripleCoherent) :
    (tripleToPairs tc).IsConsistent := by
  unfold IndepPairs.IsConsistent tripleToPairs
  rfl

/- ============================================================
   PART 3 — Theorem 2: tripleToPairs is NOT surjective
   ============================================================ -/

/-- THEOREM 2 (PROT-HH-001 §3.1): A 3-beam coherent recording is NOT
    equivalent to three independent pair recordings.

    Formally: the embedding `tripleToPairs` is not surjective. There
    exist independent-pair configurations that are not the image of
    any triple-coherent recording — namely, those that violate the
    triangle consistency φ₁₃ = φ₁₂ + φ₂₃.

    PROOF: counterexample. The pair configuration
        ⟨phase12 := 0, phase13 := 0, phase23 := π⟩
    cannot arise from any triple, since `tripleToPairs` would force
    phase13 = phase12 + phase23 = 0 + π = π, but the configuration
    has phase13 = 0. The contradiction reduces to π = 0, which fails
    by `Real.pi_ne_zero`.

    PHYSICAL INTERPRETATION: a real recording medium has finite
    coherence time. Three pairwise recordings at different times have
    independent phase references; their three relative phases are
    UNCONSTRAINED. The 3-beam recording, by simultaneous capture,
    enforces the triangle consistency as a *physical fact* about the
    recording. The constraint *is* the joint relation that pairs cannot
    encode. -/
theorem theorem2_irreducibility :
    ¬ Function.Surjective tripleToPairs := by
  intro h_surj
  -- Counterexample: a pair configuration violating triangle consistency.
  obtain ⟨tc, h_tc⟩ := h_surj
    { phase12 := 0, phase13 := 0, phase23 := Real.pi }
  -- h_tc : tripleToPairs tc = ⟨0, 0, π⟩
  -- Decompose the structure equality.
  simp only [tripleToPairs, IndepPairs.mk.injEq] at h_tc
  obtain ⟨h12, h13, h23⟩ := h_tc
  -- h12 : tc.phase12 = 0
  -- h13 : tc.phase12 + tc.phase23 = 0
  -- h23 : tc.phase23 = π
  rw [h12, h23, zero_add] at h13
  -- h13 : Real.pi = 0
  exact Real.pi_ne_zero h13

/- ============================================================
   PART 4 — Corollaries: characterising the image
   ============================================================ -/

/-- The image of `tripleToPairs` IS the consistency subspace.
    Combined with Theorem 2, this exactly characterises which
    pairwise configurations correspond to a coherent triple. -/
theorem tripleToPairs_image (ip : IndepPairs) :
    (∃ tc : TripleCoherent, tripleToPairs tc = ip) ↔ ip.IsConsistent := by
  constructor
  · rintro ⟨tc, rfl⟩
    exact tripleToPairs_consistent tc
  · intro h_consistent
    refine ⟨{ phase12 := ip.phase12, phase23 := ip.phase23 }, ?_⟩
    unfold tripleToPairs IndepPairs.IsConsistent at *
    cases ip
    simp_all

/-- `tripleToPairs` is INJECTIVE: different triples produce different
    pairwise representations. (No information is lost in the embedding;
    only constraints are added on the codomain side.) -/
theorem tripleToPairs_injective : Function.Injective tripleToPairs := by
  intro tc1 tc2 h_eq
  simp only [tripleToPairs, IndepPairs.mk.injEq] at h_eq
  obtain ⟨h12, _, h23⟩ := h_eq
  cases tc1; cases tc2
  congr

/-- COMBINED RESULT: `tripleToPairs` is injective but NOT surjective.

    The two encodings are formally information-distinct:
    - Triples ↪ Pairs (injective): no information lost; the consistency
      constraint is preserved.
    - Pairs ↛ Triples (not surjective): pairs can encode states that
      no triple can produce.

    SECURITY / CORRECTNESS INTERPRETATION FOR HOLOGRAPHIC STORAGE: the
    PROT-HH-001 architecture's claim that "3+ beam recordings are
    irreducible to pairwise recordings" is now a formal statement: the
    two recording modalities encode different information. The 3-beam
    recording's *constraint* (triangle consistency) IS the joint
    multi-party relation that pairwise recordings cannot capture, even
    in principle. -/
theorem tripleToPairs_inj_not_surj :
    Function.Injective tripleToPairs ∧ ¬ Function.Surjective tripleToPairs :=
  ⟨tripleToPairs_injective, theorem2_irreducibility⟩

/- ============================================================
   PART 5 — Status and integration
   ============================================================

   PROVEN:
     ✓ TripleCoherent and IndepPairs structures
     ✓ tripleToPairs embedding with simp lemmas for projections
     ✓ tripleToPairs_consistent — every triple's image satisfies
       triangle consistency
     ✓ THEOREM 2: theorem2_irreducibility — tripleToPairs is NOT
       surjective; there exist pair configurations not in its image
     ✓ tripleToPairs_image — characterises the image as exactly the
       consistency subspace (iff)
     ✓ tripleToPairs_injective — no information lost in the embedding
     ✓ tripleToPairs_inj_not_surj — combined inj-but-not-surj

   DEFERRED (forthcoming Phase 4 work):
     ◦ Quaternion-extension: replace ℝ-valued phases with ℍ-valued
       polarisation states; show analogous irreducibility holds in the
       SU(2) sub-case (PROT-HH-001 PRED-HH-09 dependency).
     ◦ Higher-arity: prove that n-beam coherent recordings are not
       reducible to (n-1)-beam recordings for any n ≥ 3.
     ◦ Information-theoretic form: state the codimension as Shannon
       entropy / mutual information units.

   The core load-bearing claim of PROT-HH-001 §3.1 is now formally
   established. PROT-HH-002 should cite this file (Wyrd corpus v1.3+)
   as Level 1 substrate.

   COMPANION DEMONSTRATION:
     `BMA/projects/holographic-hypergraph/sim_theorem2.py` provides
     numerical evidence over a phase-drift sweep, showing the
     consistency residual grows with drift_std and pairwise fidelity
     degrades correspondingly.
-/

end HolographicHypergraph
end Wyrd
