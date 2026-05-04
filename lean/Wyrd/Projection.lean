/-
  Wyrd-T2.2-Projection-v0.1.lean

  T2.2 — Projection well-definedness for the Cayley-Dickson tower.

  Helpful Engineering — Quaternion-Based Physics Programme
  Principal Investigator: James Paget Butler
  April 2026 — Rev 0.1 — DRAFT

  ============================================================
  PURPOSE
  ============================================================

  Each adjacent pair in the Wyrd privilege ring tower
       ℂ ⊂ ℍ ⊂ 𝕆 ⊂ 𝕊
  has a canonical projection from outer to inner ring: drop the higher
  generators and keep only the inner-ring component. In Cayley-Dickson
  coordinates, the projection of (a, b) ∈ CD A is just `a ∈ A`.

  T2.2 has three parts to prove:

    (a) The projection is a well-defined linear map.
    (b) The projection commutes with operations on values that are
        already in the inner ring (i.e., values with zero outer half).
    (c) The composite projection 𝕊 → 𝕆 → ℍ → ℂ → ℝ is well-defined.

  SECURITY INTERPRETATION:
    A kernel-ring (𝕆) process producing a result that happens to lie
    entirely in the supervisor sub-ring (ℍ) can be returned to a
    supervisor caller WITHOUT corruption: the projection is the identity
    on values already in the inner ring. This is what makes the
    layered privilege model viable in practice — kernels can return
    legitimate results to supervisors without supervisors having to
    re-validate every byte.

  ============================================================
  STATUS
  ============================================================
  DRAFT. The projections are defined and the linearity properties
  are mechanical (componentwise). The "commutes with inner-ring ops"
  property is the substantive content and proves cleanly.
-/

import Wyrd.CayleyDickson

namespace Wyrd
namespace Projection

/- ============================================================
   PART 1 — Generic Cayley-Dickson projection
   ============================================================ -/

variable {A : Type*}

/-- The canonical projection CD A → A: take the "left half".
    Geometrically: drop the higher generators; algebraically: project
    onto the inner subalgebra. -/
def π : CayleyDickson A → A := CayleyDickson.l

/-- The inclusion A → CD A: pad with zero in the right half.
    This is the splitting of the projection. -/
def ι [Zero A] (a : A) : CayleyDickson A := ⟨a, 0⟩

/- ============================================================
   PART 2 — T2.2.a: projection is a linear map
   ============================================================ -/

/-- π preserves zero. -/
@[simp] theorem π_zero [Zero A] : π (0 : CayleyDickson A) = 0 := rfl

/-- π preserves addition. -/
@[simp] theorem π_add [Add A] (x y : CayleyDickson A) :
    π (x + y) = π x + π y := rfl

/-- π preserves negation. -/
@[simp] theorem π_neg [Neg A] (x : CayleyDickson A) :
    π (-x) = -π x := rfl

/-- π preserves subtraction. -/
@[simp] theorem π_sub [Sub A] (x y : CayleyDickson A) :
    π (x - y) = π x - π y := rfl

/- ι is a section of π: applying π after ι is the identity.
   This is the round-trip property. -/
@[simp] theorem π_ι [Zero A] (a : A) : π (ι a : CayleyDickson A) = a := rfl

/- ============================================================
   PART 3 — T2.2.b: the substantive content

   Projection commutes with multiplication on inner-ring inputs.

   Let x, y ∈ CD A with x.r = y.r = 0 (i.e., both in the image of ι).
   Then π(x * y) = π(x) * π(y).

   This is the theorem that makes "kernel returns supervisor-safe
   value" actually safe: if both factors are supervisor-ring values,
   the kernel-ring product equals the supervisor-ring product.
   ============================================================ -/

/-- T2.2.b: projection commutes with multiplication on inner-ring values.
    If x and y are both in the image of ι (i.e., have zero outer halves),
    then π(x * y) = π(x) * π(y). -/
theorem π_mul_of_inner [NonUnitalNonAssocRing A] [StarAddMonoid A]
    {x y : CayleyDickson A}
    (hx : x.r = 0) (hy : y.r = 0) :
    π (x * y) = π x * π y := by
  unfold π
  rw [CayleyDickson.mul_l, hy, star_zero, zero_mul, sub_zero]

/-- COROLLARY: applying π to a product where both factors come from the
    inner ring via ι yields the inner-ring product.

    This is the form most useful for security reasoning: when kernel
    code wraps two supervisor values, multiplies them in kernel-ring
    arithmetic, and projects back, the result equals the supervisor-ring
    product. The kernel-ring "lift" leaves no residue. -/
theorem π_mul_ι [NonUnitalNonAssocRing A] [StarAddMonoid A] (a b : A) :
    π ((ι a : CayleyDickson A) * (ι b : CayleyDickson A)) = a * b := by
  rw [π_mul_of_inner rfl rfl]
  rfl

/- ============================================================
   PART 4 — Concrete projections for the Wyrd tower
   ============================================================ -/

variable {R : Type*} [CommRing R]

/-- Projection 𝕆 → ℍ: drop the upper quaternion. -/
def π_O_to_H : Octonion R → Quaternion R := π

/-- Projection 𝕊 → 𝕆: drop the upper octonion. -/
def π_S_to_O : Sedenion R → Octonion R := π

/-- Projection ℍ → ℂ: drop the j and k components.
    This uses mathlib's quaternion structure directly rather than
    Cayley-Dickson form, since ℍ comes from mathlib pre-built. -/
def π_H_to_C (q : Quaternion R) : R × R := (q.re, q.imI)
-- Note: using R × R here as a stand-in for ℂ-over-R; mathlib's
-- Complex type is over ℝ specifically. For privilege-model purposes,
-- the structure-preserving content is the dropping of imJ and imK,
-- which this captures.

/-- Composite projection 𝕊 → 𝕆 → ℍ. -/
def π_S_to_H : Sedenion R → Quaternion R := π_O_to_H ∘ π_S_to_O

/-- Composite projection 𝕊 → ℂ-component. -/
def π_S_to_C (s : Sedenion R) : R × R := π_H_to_C (π_S_to_H s)

/- ============================================================
   PART 5 — T2.2.c: composite projection well-definedness

   The composite of well-defined projections is well-defined.
   Each individual projection is linear; their composition is linear.
   ============================================================ -/

/-- The composite π_S_to_O preserves zero. -/
@[simp] theorem π_S_to_O_zero : π_S_to_O (0 : Sedenion R) = 0 := rfl

/-- The composite π_S_to_O preserves addition. -/
@[simp] theorem π_S_to_O_add (x y : Sedenion R) :
    π_S_to_O (x + y) = π_S_to_O x + π_S_to_O y := rfl

/-- The composite π_S_to_H preserves zero. -/
@[simp] theorem π_S_to_H_zero : π_S_to_H (0 : Sedenion R) = 0 := rfl

/-- The composite π_S_to_H preserves addition. -/
@[simp] theorem π_S_to_H_add (x y : Sedenion R) :
    π_S_to_H (x + y) = π_S_to_H x + π_S_to_H y := by
  unfold π_S_to_H Function.comp
  rw [π_S_to_O_add]
  rfl

/- ============================================================
   PART 6 — Security interpretation as a theorem
   ============================================================ -/

/-- SECURITY THEOREM (T2.2 main payload):

    "Kernel computations on supervisor-ring values, projected back,
     are equivalent to supervisor-ring computations."

    Concretely: if a kernel-ring (𝕆) process receives two supervisor-ring
    (ℍ) values, multiplies them as octonions, and projects the result
    back to ℍ, the answer is the same as if the multiplication had been
    performed directly in ℍ.

    This is the formal foundation for "kernel can return values to
    supervisor without corruption" — the bedrock of the layered
    privilege model. -/
theorem kernel_supervisor_safe (a b : Quaternion R) :
    π_O_to_H ((⟨a, 0⟩ : Octonion R) * (⟨b, 0⟩ : Octonion R)) = a * b := by
  unfold π_O_to_H
  exact π_mul_ι a b

/- ============================================================
   PART 7 — Outstanding work
   ============================================================

   COMPLETE:
     ✓ π and ι defined for generic CD
     ✓ π preserves all componentwise additive structure
     ✓ T2.2.b: π_mul_of_inner — projection commutes with mul on inner inputs
     ✓ Concrete projections π_O_to_H, π_S_to_O, π_H_to_C, composites
     ✓ kernel_supervisor_safe — security interpretation as theorem

   DEFERRED:
     ◦ ℍ → ℂ projection at the level of mathlib's Complex type (not just R × R)
       Requires a coercion or specialized treatment because mathlib's
       Complex is over ℝ specifically. Not needed for the abstract security
       property.
     ◦ Full LinearMap or RingHom packaging — currently π is a bare function.
       Should be promoted to a structured map for downstream consumers.
       Mechanical work.

   This file delivers the security-relevant content of T2.2:
     "Outer-ring computations on inner-ring values, when projected back,
      equal inner-ring computations."

   That property, plus T2.1 (generator non-synthesis), gives the full
   bidirectional safety claim of the layered privilege model.
-/

end Projection
end Wyrd
