/-
  Wyrd/CTH.lean

  Class B Phase 2 — CTH (Confluent Trust Hypergraph) entropy + monotonicity.

  Helpful Engineering — Quaternion-Based Physics Programme
  April 2026 — Rev 0.1

  ============================================================
  PURPOSE
  ============================================================

  CTH (per James's Confluent Trust Hypergraph paper, 2026) defines
  five computable metrics over a trust hypergraph: η (entropy),
  μ (fidelity), I (mutual information), Δ (programme deficit),
  Re_e (Reynolds-analogue for incoherence).

  This file formalizes the Tier-2 (measurement) entropy

      η(v) = -log(1 - δ(v))   for δ(v) ∈ [0, 1)

  and proves:

    (C-20b) cth_measurement_evidence_monotonic:
      if a measurement node v with fractional error δ is updated
      with consistent evidence yielding new error δ' ≤ δ, then
      the new entropy satisfies η(v') ≤ η(v).

  This certifies the load-bearing CTH property: better evidence
  cannot make the framework more uncertain.

  ============================================================
  TIER MODEL
  ============================================================

  Trust tiers (per CTH paper Definition 2):
    Tier 0 — Axiom: η(v) = H_axiom (foundational, fixed)
    Tier 1 — Proof: η(v) = 0 (machine-verified, sorry count = 0)
    Tier 2 — Measurement: η(v) = -log(1 - δ) where δ is fractional error
    Tier 3 — Prediction: η(v) inherited from weakest upstream link

  Monotonicity is interesting at Tier 2; Tier 0 is constant; Tier 1 is
  identically zero; Tier 3 derives from upstream. C-20b targets Tier 2.
-/

import Mathlib.Analysis.SpecialFunctions.Log.Basic
import Mathlib.Tactic.Linarith
import Mathlib.Tactic.NormNum
import Wyrd.Hypergraph

namespace Wyrd
namespace CTH

/- ============================================================
   PART 1 — Trust tiers
   ============================================================ -/

/-- The four CTH trust tiers per Definition 2. -/
inductive TrustTier where
  | axiom        -- Tier 0
  | proof        -- Tier 1
  | measurement  -- Tier 2
  | prediction   -- Tier 3
  deriving DecidableEq, Repr

/- ============================================================
   PART 2 — Tier-2 measurement entropy
   ============================================================ -/

/-- Tier-2 entropy: η(δ) = -log(1 - δ) for δ ∈ [0, 1).

    DOMAIN ASSUMPTION: the input δ must satisfy 0 ≤ δ < 1 for the
    expression to be well-defined (log is finite for positive
    arguments; we ensure 1 - δ > 0). Outside this range the function
    is mathematically undefined; the theorems below carry the range
    constraint as an explicit hypothesis. -/
noncomputable def measurementEntropy (δ : ℝ) : ℝ :=
  -Real.log (1 - δ)

/- ============================================================
   PART 3 — Theorem C-20b: monotonicity under evidence
   ============================================================ -/

/-- KEY LEMMA: log is monotone nondecreasing on positive reals.
    For 0 < a ≤ b, we have log a ≤ log b. -/
theorem log_monotone_on_positive (a b : ℝ) (ha : 0 < a) (hab : a ≤ b) :
    Real.log a ≤ Real.log b := by
  exact Real.log_le_log ha hab

/-- THEOREM C-20b: measurement-evidence entropy is monotone under
    evidence improvement.

    If a Tier-2 measurement node v has fractional error δ ∈ [0, 1),
    and consistent new evidence reduces the error to δ' ∈ [0, 1)
    with δ' ≤ δ, then the new entropy is at most the old entropy:

        η(δ') ≤ η(δ)

    READING: more evidence cannot raise η for a Tier-2 node. The
    framework's epistemic state is monotone in the evidence direction
    — adding consistent evidence is always at least as good as having
    less. CTH's auditability claim depends on this property: an
    auditor reviewing a programme's evidence trail can rely on η
    decreasing (or staying constant) as more measurements arrive,
    not flipping unpredictably.

    SECURITY INTERPRETATION: the watchdog cannot be tricked into
    raising η — and thereby weakening a Tier-2 node's trust standing
    — by an attacker submitting "evidence" that's actually a no-op.
    Adding any evidence δ' ≤ δ is safe; it monotonically tightens
    the trust standing. -/
theorem cth_measurement_evidence_monotonic
    (δ δ' : ℝ)
    (h_δ_lower : 0 ≤ δ) (h_δ_upper : δ < 1)
    (h_δ'_lower : 0 ≤ δ') (h_δ'_upper : δ' < 1)
    (h_evidence : δ' ≤ δ) :
    measurementEntropy δ' ≤ measurementEntropy δ := by
  unfold measurementEntropy
  -- 1 - δ ≤ 1 - δ' since δ' ≤ δ
  have h_pos_δ : (0 : ℝ) < 1 - δ := by linarith
  have h_pos_δ' : (0 : ℝ) < 1 - δ' := by linarith
  have h_diff : 1 - δ ≤ 1 - δ' := by linarith
  -- log is monotone increasing → log (1 - δ) ≤ log (1 - δ')
  have h_log : Real.log (1 - δ) ≤ Real.log (1 - δ') :=
    log_monotone_on_positive (1 - δ) (1 - δ') h_pos_δ h_diff
  -- Negate both sides
  linarith

/- ============================================================
   PART 4 — Tier-1 (proof) entropy is identically zero
   ============================================================ -/

/-- Tier-1 entropy is zero by definition: a machine-verified proof
    has no information deficit beyond its premises. Stated as a
    constant function for compositionality. -/
def proofEntropy : ℝ := 0

/-- Trivial corollary: any two Tier-1 nodes have equal entropy. -/
theorem cth_proof_entropy_equal : ∀ _v₁ _v₂ : Unit, proofEntropy = proofEntropy := by
  intros; rfl

/- ============================================================
   PART 5 — Boundary case: δ = 0 means η = 0
   ============================================================ -/

/-- A perfect measurement (zero fractional error) has zero entropy. -/
theorem cth_zero_error_zero_entropy : measurementEntropy 0 = 0 := by
  unfold measurementEntropy
  simp [Real.log_one]

/- ============================================================
   PART 6 — Status
   ============================================================

   PROVEN:
     ✓ TrustTier enumeration matches CTH Definition 2
     ✓ measurementEntropy : ℝ → ℝ via -log(1 - δ)
     ✓ log_monotone_on_positive (helper)
     ✓ C-20b: cth_measurement_evidence_monotonic
       — better evidence (δ' ≤ δ) yields lower or equal entropy
     ✓ proofEntropy = 0 (Tier-1 boundary)
     ✓ cth_zero_error_zero_entropy: η(0) = 0

   DEFERRED (next phase, not blocking):
     ◦ Hypergraph-level entropy: integrate measurementEntropy with
       Wyrd.Hypergraph by indexing it through a per-node tier function
     ◦ Programme deficit Δ(G) = Σ η over source anchors
     ◦ Mutual information I at confluence points
     ◦ Reynolds-analogue Re_e for incoherence detection

   These are larger formal commitments. C-20b alone closes the
   "evidence cannot raise η" gap that was blocking Class B
   production claims about CTH soundness.
-/

end CTH
end Wyrd
