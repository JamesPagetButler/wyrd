/-
  Wyrd/NaryMI.lean

  CTH NaryMI synergy positivity — Phase 4 v1.5 CTH lift.

  Helpful Engineering — Quaternion-Based Physics Programme
  May 2026 — Wyrd corpus v1.5 (CTH-side companion)

  ============================================================
  PURPOSE
  ============================================================

  Provides the formal soundness anchor for `confluent-trust`'s
  `compute/mutual_info.go::NaryMI` synergy bonus. The bonus, defined as

      synergyPerExtraPath = 0.5 * log2(1 + n / (chiSq + ε))
      bonus               = (n - 2) * synergyPerExtraPath

  is the operational form of `Wyrd.HolographicHypergraph.theorem2_irreducibility`
  (Phase 4 v1.4) and its higher-arity generalisation
  `Wyrd.HolographicHypergraphHigherArity.theorem2_irreducibility_n_arity`
  (Phase 4 v1.5): an N-way agreement at a confluence carries information
  beyond the pairwise sum.

  This file proves the *strict positivity* of the synergy term and
  total bonus under the conditions CTH's NaryMI always satisfies
  (n ≥ 3, ε > 0 from `epsilonRegularization`, chiSq ≥ 0 by construction).

  ============================================================
  CONNECTION TO PHASE 4 IRREDUCIBILITY
  ============================================================

  Theorem 2 (HolographicHypergraph) shows that an n-beam coherent
  recording carries joint constraints not in any pair-only
  decomposition. NaryMI is a numerical estimator of joint information
  for N independent estimators of a common target; its synergy term
  measures how much "extra" information is carried by the joint
  agreement beyond the sum of pairwise agreements.

  The PHILOSOPHICAL connection: both express that joint information
  is not pair-decomposable. The FORMAL connection in this file is
  weaker — we prove the synergy term is strictly positive under
  realistic preconditions, which is the *necessary* condition for
  NaryMI to behave as a non-trivial joint-information estimator.
  We do NOT prove (here) that NaryMI numerically tracks Theorem 2's
  irreducibility statement — the latter is a structural claim about
  embeddings; the former is a numerical claim about Gaussian channel
  capacities.

  Tracks: github.com/JamesPagetButler/confluent-trust#35,
          github.com/JamesPagetButler/wyrd#3.

  See `Wyrd-Proofs-Reference-v1.5.md` §31 (Phase 4 ℍ) and the
  forthcoming v1.6 §28 (NaryMI lift, this file) for the corpus map.
-/

import Mathlib.Analysis.SpecialFunctions.Log.Base
import Mathlib.Tactic.Linarith
import Mathlib.Tactic.Positivity

namespace Wyrd
namespace NaryMI

/- ============================================================
   PART 1 — The synergy term
   ============================================================ -/

/-- The per-extra-path synergy term in CTH's NaryMI:
    `0.5 * log₂(1 + n / (chiSq + ε))`.

    This is the quantity multiplied by `(N − 2)` in NaryMI to give
    the total bonus over the pairwise sum. -/
noncomputable def synergyTerm (n : ℕ) (chiSq ε : ℝ) : ℝ :=
  (1 / 2) * Real.logb 2 (1 + (n : ℝ) / (chiSq + ε))

/- ============================================================
   PART 2 — Strict positivity of the synergy term
   ============================================================ -/

/-- The argument `1 + n / (chiSq + ε)` is strictly greater than 1 when
    `n > 0`, `chiSq ≥ 0`, and `ε > 0`. -/
private theorem synergy_arg_gt_one (n : ℕ) (chiSq ε : ℝ)
    (h_n : 0 < n) (h_chiSq : 0 ≤ chiSq) (h_ε : 0 < ε) :
    1 < 1 + (n : ℝ) / (chiSq + ε) := by
  have h_denom_pos : 0 < chiSq + ε := by linarith
  have h_n_pos : (0 : ℝ) < (n : ℝ) := by exact_mod_cast h_n
  have h_quot_pos : 0 < (n : ℝ) / (chiSq + ε) := div_pos h_n_pos h_denom_pos
  linarith

/-- The synergy term is strictly positive for `n > 0`, `chiSq ≥ 0`,
    `ε > 0`. CTH's NaryMI always satisfies these: `n` is the chain
    count (positive whenever the function is called with ≥ 1 chain),
    `chiSq` is a sum of squares (non-negative), and `ε` is
    `epsilonRegularization = 1e-12 > 0`. -/
theorem synergyTerm_pos (n : ℕ) (chiSq ε : ℝ)
    (h_n : 0 < n) (h_chiSq : 0 ≤ chiSq) (h_ε : 0 < ε) :
    0 < synergyTerm n chiSq ε := by
  unfold synergyTerm
  have h_arg : 1 < 1 + (n : ℝ) / (chiSq + ε) :=
    synergy_arg_gt_one n chiSq ε h_n h_chiSq h_ε
  have h_log_pos : 0 < Real.logb 2 (1 + (n : ℝ) / (chiSq + ε)) := by
    apply Real.logb_pos (by norm_num) h_arg
  linarith

/- ============================================================
   PART 3 — Strict positivity of the total bonus
   ============================================================ -/

/-- The total NaryMI bonus over the pairwise sum:
    `(N − 2) * synergyTerm`.

    For `N = 2`, the bonus is zero (NaryMI reduces to pairwise sum).
    For `N ≥ 3`, the bonus is strictly positive. -/
noncomputable def totalBonus (n : ℕ) (chiSq ε : ℝ) : ℝ :=
  ((n : ℝ) - 2) * synergyTerm n chiSq ε

/-- THE LIFT: For `N ≥ 3` chains with non-negative chi-squared and
    positive regularisation, the NaryMI synergy bonus over the pairwise
    sum is strictly positive.

    READING: Joint information at a confluence of arity ≥ 3 strictly
    exceeds the pairwise-sum estimator, formalising the necessary
    condition for NaryMI to behave as a non-trivial joint estimator
    in the regime CTH actually uses (n ≥ 3 chains, finite precision-
    weighted residual).

    Cited from confluent-trust `compute/mutual_info.go::NaryMI`. -/
theorem nary_mi_bonus_pos (n : ℕ) (chiSq ε : ℝ)
    (h_n : 3 ≤ n) (h_chiSq : 0 ≤ chiSq) (h_ε : 0 < ε) :
    0 < totalBonus n chiSq ε := by
  unfold totalBonus
  have h_n_pos : 0 < n := by omega
  have h_synergy : 0 < synergyTerm n chiSq ε :=
    synergyTerm_pos n chiSq ε h_n_pos h_chiSq h_ε
  have h_coeff : (0 : ℝ) < (n : ℝ) - 2 := by
    have : (3 : ℝ) ≤ (n : ℝ) := by exact_mod_cast h_n
    linarith
  exact mul_pos h_coeff h_synergy

/-- Boundary case: at exactly `N = 2` chains, the bonus is zero.
    NaryMI degenerates to the pairwise sum; no synergy. -/
theorem nary_mi_bonus_zero_at_two (chiSq ε : ℝ) :
    totalBonus 2 chiSq ε = 0 := by
  unfold totalBonus
  simp

/- ============================================================
   PART 4 — Status and integration
   ============================================================

   PROVEN:
     ✓ synergyTerm definition
     ✓ synergyTerm_pos — synergy term strictly positive in the
       realistic CTH regime (n > 0, chiSq ≥ 0, ε > 0)
     ✓ totalBonus definition
     ✓ nary_mi_bonus_pos — total bonus strictly positive for n ≥ 3
     ✓ nary_mi_bonus_zero_at_two — boundary case at n = 2

   NOT PROVEN (deliberate scope, defer to v1.6+):
     ◦ A direct logical bridge from theorem2_irreducibility_n_arity
       (HolographicHypergraphHigherArity) to nary_mi_bonus_pos. The
       two are conceptually aligned (both express "joint info not in
       pair marginals") but live over different mathematical objects
       (structural embeddings vs Gaussian channel capacity). A
       formal bridge would require modelling NaryMI's Gaussian setup
       inside the holographic-hypergraph framework, or vice versa,
       and is a Sprint-phase research question.
     ◦ Tight upper bound on totalBonus as a function of (n, chiSq, ε).
       Useful for CTH's "synergy capacity" reporting; deferred.
     ◦ Faithful Lean reproduction of NaryMI's full implementation
       (pairwise_sum + bonus). Possible but verbose; the present
       theorems target the load-bearing claim CTH actually needs to
       cite.

   The CTH lift gap from issue #35 is closed in the form CTH
   asked for (synergy strictly positive for n ≥ 3 with bounded
   inputs).
-/

end NaryMI
end Wyrd
