/-
  Wyrd/HamiltonProduct.lean

  Wyrd-local Hamilton product formula theorem.

  Helpful Engineering — Quaternion-Based Physics Programme
  Principal Investigator: James Paget Butler
  May 2026 — Rev 0.1

  ============================================================
  PURPOSE
  ============================================================

  Notary Phase 1 Cycle 1 (dispatch 2026-05-20) returned
  `INCONCLUSIVE_UNREACHED_GOAL` because `lean/Wyrd/Foundations.lean`
  does not contain a *named* HamiltonProduct theorem — it imports
  `Quaternion.mul` from mathlib4 and uses it as a dependency in
  ring-tower closure proofs, but Wyrd does not re-prove the formula.

  The Go runtime docstring in `compute/quaternion.go` cited
  `"Quaternion.mul in lean/Wyrd/Foundations.lean"` as the Lean
  anchor. That citation was documentation-level cross-formalism
  drift: an import citation ≠ Wyrd proof ownership. This file
  resolves it.

  Two seam records resolved here (wyrd-issue-#68 companion fix):
    NT_SEAM_RECORD_001 — phantom theorem citation in Go docstring
    NT_SEAM_RECORD_002 — sandwich_mul / HamiltonProduct docstring
                         conflation (separated in quaternion.go)

  ============================================================
  THEOREM SHAPE
  ============================================================

  `hamilton_product_formula` states the explicit 16-multiply /
  12-add component expansion of `Quaternion.mul` (mathlib4,
  `Mathlib.Algebra.Quaternion`) and proves equality component-wise:

    (a * b).re  = a.re*b.re   - a.imI*b.imI - a.imJ*b.imJ - a.imK*b.imK
    (a * b).imI = a.re*b.imI  + a.imI*b.re  + a.imJ*b.imK - a.imK*b.imJ
    (a * b).imJ = a.re*b.imJ  - a.imI*b.imK + a.imJ*b.re  + a.imK*b.imI
    (a * b).imK = a.re*b.imK  + a.imI*b.imJ - a.imJ*b.imI + a.imK*b.re

  This matches the Go implementation in `compute/quaternion.go`
  `HamiltonProduct` at `TierQuaternion` (dispatches to
  `emulator.Gearbox.QMul64`), which implements the same formula.

  The theorem is stated over any `CommRing α` — the quaternion
  multiplication formula holds regardless of the coefficient ring.
  Over `ℝ` this is the standard Hamilton product.

  NOTARY TARGET: this theorem is the Lean side of Notary Cycle 1
  Competency #1 (Lean→Coq cross-prover correspondence). The Coq
  port target is the `hamilton_product_formula` theorem body, which
  states four ring-arithmetic equalities provable from the mathlib4
  definition of `Quaternion.mul`.
-/

import Mathlib.Algebra.Quaternion

namespace Wyrd
namespace HamiltonProduct

/-- `hamilton_product_formula`: the explicit 16-multiply / 12-add
    component expansion of `Quaternion.mul` (mathlib4).

    For quaternions `a = (a₀, a₁, a₂, a₃)` and `b = (b₀, b₁, b₂, b₃)`
    (with components `.re`, `.imI`, `.imJ`, `.imK` respectively), the
    Hamilton product `a * b` has components:

      w = a₀b₀ - a₁b₁ - a₂b₂ - a₃b₃
      x = a₀b₁ + a₁b₀ + a₂b₃ - a₃b₂
      y = a₀b₂ - a₁b₃ + a₂b₀ + a₃b₁
      z = a₀b₃ + a₁b₂ - a₂b₁ + a₃b₀

    This is the Wyrd-local named theorem that resolves `NT_SEAM_RECORD_001`
    (phantom theorem citation in `compute/quaternion.go`). The Go
    implementation (`HamiltonProduct` at `TierQuaternion`, delegating to
    `emulator.Gearbox.QMul64`) is the runtime realisation of this formula.

    Stated over `CommRing α` for full generality; specialises to the
    standard Hamilton product over `ℝ`. -/
theorem hamilton_product_formula {α : Type*} [CommRing α] (a b : Quaternion α) :
    a * b = ⟨a.re * b.re - a.imI * b.imI - a.imJ * b.imJ - a.imK * b.imK,
             a.re * b.imI + a.imI * b.re + a.imJ * b.imK - a.imK * b.imJ,
             a.re * b.imJ - a.imI * b.imK + a.imJ * b.re + a.imK * b.imI,
             a.re * b.imK + a.imI * b.imJ - a.imJ * b.imI + a.imK * b.re⟩ := by
  ext
  · simp only [Quaternion.re_mul]
  · simp only [Quaternion.imI_mul]
  · simp only [Quaternion.imJ_mul]
  · simp only [Quaternion.imK_mul]

/- ============================================================
   STATUS
   ============================================================

   PROVEN (wyrd-issue-#68, no holes, zero user-defined axiom):
     ✓ hamilton_product_formula — 16-mul/12-add Hamilton product
       formula equals mathlib4 Quaternion.mul; stated over CommRing;
       resolves NT_SEAM_RECORD_001 + NT_SEAM_RECORD_002

   NOTARY FOLLOW-ON:
     Notary Cycle 1 Competency #1 Lean→Coq cross-prover port now
     has a named Wyrd-local anchor theorem. Notary dispatch re-run
     per wyrd-issue-#68 acceptance criterion AC-6.

   THIS FILE IS RESEARCH-TIER (not substrate-tier).
   Substrate-tier promotion is a separate PR + §I4 + Spec 9.2 §2
   four-criteria gate. Adding to lean/Wyrd/Substrate.lean is the
   promotion action; that is NOT part of this PR.
-/

end HamiltonProduct
end Wyrd
