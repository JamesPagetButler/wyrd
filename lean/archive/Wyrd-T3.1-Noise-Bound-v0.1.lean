/-
  Wyrd-T3.1-Noise-Bound-v0.1.lean

  T3.1 — Associator-noise bound for finite-precision arithmetic.

  Helpful Engineering — Quaternion-Based Physics Programme
  Principal Investigator: James Paget Butler
  April 2026 — Rev 0.1 — DRAFT

  ============================================================
  PURPOSE
  ============================================================

  The Wyrd seam detector computes associator(a, b, c) = (a·b)·c − a·(b·c)
  and compares against threshold ε. Two thresholds are relevant:
      ε_physics    — physical-seam threshold (~10⁻³ for typical sims)
      ε_privilege  — privilege-boundary threshold (must exceed noise)

  The associator's mathematical zero on associative inputs is corrupted
  by floating-point rounding. T3.1 establishes a parametric bound on the
  noise floor δ_fp as a function of:
      ε_fp  — relative rounding error per operation (2⁻²³ for fp32)
      M     — magnitude bound on operand components
      d     — operation depth (multiply-adds in a single associator)

  We prove a Wilkinson-style backward error bound:
      |computed_associator| ≤ K · ε_fp · M³  + O(ε_fp²)

  where K is a small constant determined by depth d.

  Security relevance: this bound lets us SET ε_privilege to a value
  that PROVABLY distinguishes real privilege violations from rounding
  artifacts. Without this, ε_privilege is empirical.

  ============================================================
  APPROACH
  ============================================================

  Lean's mathlib has limited IEEE-754 infrastructure. Instead of
  formalizing IEEE-754 directly, we model finite-precision arithmetic
  abstractly: a "floating-point structure" is a real-valued semiring
  equipped with a rounding map fl : ℝ → F satisfying

      ∀ x ∈ ℝ, |fl(x) − x| ≤ ε_fp · |x|

  All actual operations (add, mul, sub) are then modelled as the exact
  real operation followed by fl. This is the standard model in numerical
  analysis (Higham, "Accuracy and Stability of Numerical Algorithms").

  ============================================================
  STATUS
  ============================================================
  DRAFT. Provides the abstract framework and the parametric bound for
  one quaternion product chain. The full octonion associator (which
  involves nested products) is sketched. Specialization to fp32 with
  concrete numerical values is provided as a corollary.
-/

import Mathlib.Analysis.SpecialFunctions.Pow.Real
import Mathlib.Data.Real.Basic
import Mathlib.Tactic.Linarith
import Mathlib.Tactic.NormNum
import Mathlib.Tactic.FieldSimp

namespace Wyrd
namespace NoiseBound

/- ============================================================
   PART 1 — Abstract floating-point structure
   ============================================================ -/

/-- A rounding model. Captures the relative-error property of any
    IEEE-style floating-point format: rounding never changes a value
    by more than (1 + ε_fp) factor.

    For fp16: ε_fp ≈ 2⁻¹⁰ ≈ 10⁻³
    For fp32: ε_fp ≈ 2⁻²³ ≈ 1.2 × 10⁻⁷
    For fp64: ε_fp ≈ 2⁻⁵² ≈ 2.2 × 10⁻¹⁶ -/
structure RoundingModel where
  /-- The rounding map. Sends a real number to its representable approximation. -/
  fl : ℝ → ℝ
  /-- The relative-error parameter. -/
  ε_fp : ℝ
  /-- Positivity of ε_fp. -/
  ε_pos : 0 < ε_fp
  /-- ε_fp is small (less than 1). -/
  ε_small : ε_fp < 1
  /-- The fundamental rounding inequality. -/
  fl_error : ∀ x : ℝ, |fl x - x| ≤ ε_fp * |x|
  /-- Zero is exactly representable. -/
  fl_zero : fl 0 = 0

namespace RoundingModel
variable (R : RoundingModel)

/-- Rounded multiplication: round(x · y).
    Models a single floating-point multiply. -/
def mul (x y : ℝ) : ℝ := R.fl (x * y)

/-- Rounded addition: round(x + y). -/
def add (x y : ℝ) : ℝ := R.fl (x + y)

/-- Rounded subtraction: round(x − y). -/
def sub (x y : ℝ) : ℝ := R.fl (x - y)

/-- Bound on the error of a single rounded multiply.
    |R.mul x y − x*y| ≤ ε_fp · |x*y|

    This is just an instance of the rounding-model axiom. -/
theorem mul_error (x y : ℝ) :
    |R.mul x y - x * y| ≤ R.ε_fp * |x * y| := by
  unfold mul
  exact R.fl_error (x * y)

/-- Bound on the error of a single rounded subtract. -/
theorem sub_error (x y : ℝ) :
    |R.sub x y - (x - y)| ≤ R.ε_fp * |x - y| := by
  unfold sub
  exact R.fl_error (x - y)

end RoundingModel

/- ============================================================
   PART 2 — Composition of errors

   When operations are chained, errors compose multiplicatively but
   to first order add linearly. The standard result: a chain of n
   operations with relative error ε each has total relative error
   bounded by n·ε + O(ε²).

   We prove the linear composition for two operations, which is the
   building block for arbitrary depth.
   ============================================================ -/

/-- ABSOLUTE-ERROR COMPOSITION (one product, magnitude bound).

    If |x|, |y| ≤ M, then the rounded product fl(x*y) deviates from
    x*y by at most ε_fp · M².

    This converts the relative-error axiom into an absolute-error
    bound, which is what we actually need for associator-magnitude
    reasoning. -/
theorem abs_error_one_mul (R : RoundingModel) (x y : ℝ) (M : ℝ)
    (hM : 0 ≤ M) (hx : |x| ≤ M) (hy : |y| ≤ M) :
    |R.mul x y - x * y| ≤ R.ε_fp * M^2 := by
  calc |R.mul x y - x * y|
      ≤ R.ε_fp * |x * y| := R.mul_error x y
    _ = R.ε_fp * (|x| * |y|) := by rw [abs_mul]
    _ ≤ R.ε_fp * (M * M) := by
        apply mul_le_mul_of_nonneg_left
        · exact mul_le_mul hx hy (abs_nonneg _) hM
        · exact le_of_lt R.ε_pos
    _ = R.ε_fp * M^2 := by ring

/-- COMPOSED ERROR for a chain of two products: fl(fl(a*b) * c).

    Bound: |fl(fl(a*b)*c) − a*b*c| ≤ 2·ε_fp · M³ + ε_fp² · M³

    The linear-in-ε_fp coefficient is 2 (one error per multiply).
    The quadratic term is small enough to bound by ε_fp · M³ when
    ε_fp is small. -/
theorem abs_error_two_muls (R : RoundingModel) (a b c : ℝ) (M : ℝ)
    (hM : 1 ≤ M)
    (ha : |a| ≤ M) (hb : |b| ≤ M) (hc : |c| ≤ M) :
    |R.mul (R.mul a b) c - (a * b) * c| ≤ 2 * R.ε_fp * M^3 + R.ε_fp^2 * M^3 := by
  -- The intermediate result fl(a*b) has magnitude ≤ M^2 + ε_fp · M^2
  -- Strategy: triangle inequality through the intermediate exact product (a*b)*c.
  -- |fl(fl(a*b)*c) − (a*b)*c|
  --   ≤ |fl(fl(a*b)*c) − fl(a*b)*c|  +  |fl(a*b)*c − (a*b)*c|
  --   ≤ ε_fp · |fl(a*b)| · |c|        +  |fl(a*b) − a*b| · |c|
  --   ≤ ε_fp · (|a*b| + ε_fp·|a*b|) · M  +  ε_fp · |a*b| · M
  --   ≤ ε_fp · M³ + ε_fp² · M³ + ε_fp · M³
  -- The proof structure is sound; the explicit Lean tactic chain is
  -- mechanical (chain of `calc` steps with abs_mul, mul_le_mul, linarith).
  -- Stated as a target; full proof body is straightforward bookkeeping.
  sorry  -- TODO: replace with explicit calc chain

/- ============================================================
   PART 3 — Quaternion associator noise bound (sketch)

   The full quaternion associator involves ~12 multiplications and
   several additions/subtractions. Following the same composition
   pattern, the total error is bounded by:

       |fl_associator(a, b, c) − associator(a, b, c)|
           ≤ K_quat · ε_fp · M³ + O(ε_fp²)

   where K_quat is a small constant (≤ 24) representing the maximum
   number of error-introducing operations in the associator chain.

   The catastrophic-cancellation case: when associator(a, b, c) = 0
   exactly (i.e., (a, b, c) is a triple in an associative subalgebra),
   the noise floor is K_quat · ε_fp · M³ — pure numerical noise with
   no signal. This is the "ε_privilege must be set above this floor"
   constraint.
   ============================================================ -/

/-- Statement of the quaternion-associator noise bound.

    For triples (a, b, c) in ℍ with all components bounded by M, the
    fp-computed associator differs from the exact associator by at
    most K · ε_fp · M³ for K bounded by a small constant.

    For triples in a sub-associative algebra (where the exact
    associator is zero), the fp-computed associator is bounded
    purely by the noise floor K · ε_fp · M³. -/
def associator_noise_bound (R : RoundingModel) (M : ℝ) : ℝ :=
  24 * R.ε_fp * M^3

/-- T3.1 specialized to fp32.

    For fp32 (ε_fp ≈ 2⁻²³ ≈ 1.2e-7) and components bounded by M = 10:
    noise floor ≤ 24 · 1.2e-7 · 1000 ≈ 2.9e-3.

    For M = 1: noise floor ≤ 2.9e-6.

    Therefore ε_privilege should be at least ~10⁻⁵ for M = 1 components,
    or ~10⁻² for M = 10 components, to provably distinguish real
    privilege boundaries from rounding artifacts. -/
def fp32_noise_floor (M : ℝ) : ℝ := 24 * 1.2e-7 * M^3

theorem fp32_noise_unit_magnitude :
    fp32_noise_floor 1 ≤ 3e-6 := by
  unfold fp32_noise_floor
  norm_num

theorem fp32_noise_decimal_magnitude :
    fp32_noise_floor 10 ≤ 3e-3 := by
  unfold fp32_noise_floor
  norm_num

/- ============================================================
   PART 4 — Threshold separation theorem (T3.2 preview)

   Given the noise floor, the privilege threshold ε_privilege must
   be strictly greater than the noise floor for the privilege detector
   to be sound. This is T3.2.

   Equivalent: ε_privilege / (ε_fp · M³) ≥ K · (some safety factor).

   For fp32 with M = 1 and a safety factor of 100×:
       ε_privilege ≥ 100 · 24 · 1.2e-7 ≈ 3 × 10⁻⁴

   This is a recommendation, not a tight bound — but it's a
   defensible numerical floor for the privilege model.
   ============================================================ -/

/-- Threshold separation: the privilege threshold must exceed the
    noise floor by some safety factor. -/
def threshold_separation_safe
    (ε_priv : ℝ) (R : RoundingModel) (M : ℝ) (safety_factor : ℝ) : Prop :=
  ε_priv ≥ safety_factor * associator_noise_bound R M

/- ============================================================
   PART 5 — Outstanding work
   ============================================================

   COMPLETE:
     ✓ Abstract RoundingModel with the standard relative-error axiom
     ✓ Single-operation absolute error bound
     ✓ Specialization to fp32 with concrete numerical bounds
     ✓ Statement of associator_noise_bound
     ✓ Statement of threshold_separation_safe (T3.2 preview)
     ✓ fp32_noise_unit_magnitude and fp32_noise_decimal_magnitude:
       concrete numerical claims, machine-checked

   DEFERRED (the one `sorry`):
     ◦ abs_error_two_muls — the proof body. The bound is correct;
       the closing tactic is a calc chain (triangle inequality through
       intermediate exact products). Mechanical work — committing
       this proof body is the next sub-task.

   GENUINELY OPEN:
     ◦ The full octonion associator depth count. We claimed K ≤ 24
       on intuition; the actual operation count for octonion-product
       associators needs to be verified. If the correct K is
       substantially larger (e.g., 200), the safety factor recommendation
       in Part 4 changes proportionally.
     ◦ The catastrophic-cancellation analysis when (a*b)*c ≈ a*(b*c).
       This is where Wilkinson backward analysis gets sharp; the simple
       forward bound here is loose. Sharpening could make ε_privilege
       smaller (and thus more discriminating).

   This file is the FRAMEWORK. With abs_error_two_muls completed and
   the K constant nailed down, T3.1 stands as a fully formal noise
   bound that ε_privilege must exceed.
-/

end NoiseBound
end Wyrd
