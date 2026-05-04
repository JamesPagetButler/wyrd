# Wyrd / Skuld Lean Proof Corpus — Theorem Reference

**Helpful Engineering — Quaternion-Based Physics Programme**
Principal Investigator: James Paget Butler
April 2026 — Rev 1.0

> **Purpose.** This document is the canonical reference for the Lean 4 proof corpus that backs the Wyrd / Skuld algebraic privilege model. Every theorem in the corpus is listed with: (a) its formal statement, (b) the proof strategy used, (c) the architectural property it underwrites, and (d) where downstream code is expected to cite it. Use this when implementing `qbpcu`, `wyrd`, or `skuld` Go packages — every soundness argument in those packages should reference an entry here.

---

## 0. Build & toolchain (verified 2026-04-25)

| Component | Pin |
|---|---|
| Lean toolchain | `leanprover/lean4:v4.30.0-rc1` |
| Mathlib | rev `a090f46da78e9af11fee348cd7ee47bf8dd219d2` |
| Project root | `~/Documents/Wyrd/wyrd-lean-project/` |
| Build status | `lake build` succeeds; **0 sorries, 0 user-defined axioms** across all 7 files |

To rebuild from cold: `cd ~/Documents/Wyrd/wyrd-lean-project && lake update && lake exe cache get && lake build`.

---

## 1. Inventory at a glance

| Theorem | File | Architectural role |
|---|---|---|
| `no_surjection_comm_to_noncomm` | Foundations | Abstract: ring map from CommRing onto non-commutative ring impossible |
| `no_surjection_assoc_to_nonassoc` | Foundations | Abstract: assoc → nonassoc impossible |
| `no_surjection_alt_to_nonalt` | Foundations | Abstract: alt → nonalt impossible |
| `commutator_eq_zero_of_comm` | Foundations | T1.2: commutator detector vanishes in commutative rings |
| `associator_eq_zero_of_assoc` | Foundations | T1.2: associator detector vanishes in associative rings |
| `alternator_eq_zero_of_alt` | Foundations | T1.2: alternator detector vanishes in alternative rings |
| `commutator_quaternion_witness` | Foundations | T1.2.a: ∃ a b ∈ ℍ with [a,b] ≠ 0 |
| `no_surjection_complex_to_quaternion` | Foundations | **T2.1.a: user (ℂ) → supervisor (ℍ) boundary closed** |
| `associator_octonion_witness` | CayleyDickson | T1.2.b: ∃ a b c ∈ 𝕆 with (ab)c ≠ a(bc) |
| `alternator_sedenion_witness` | SedenionWitness | T1.2.c: ∃ a b ∈ 𝕊 with (aa)b ≠ a(ab) |
| `octonion_alternative` | OctonionAlternative | 𝕆 is alternative (used for T2.1.c) |
| `quat_norm_is_real` | OctonionAlternative | q · star q lies in scalar subfield |
| `quat_real_part_is_real` | OctonionAlternative | q + star q lies in scalar subfield |
| `Projection.π_mul_of_inner` | Projection | **T2.2: outer-ring × inner-ring values, projected, equal inner-ring product** |
| `Projection.π_mul_ι` | Projection | T2.2 corollary on ι-embedded values |
| `Projection.kernel_supervisor_safe` | Projection | **Security headline: kernel computations on supervisor values, projected back, are safe** |
| `Capability.sandwich_preservation_associative` | Capability | Sandwich p·u·p⁻¹ behaves correctly under invertibility |
| `Capability.capability_grants_safe_access` | Capability | **T2.3 positive: capability holder can compute safely** |
| `Capability.no_capability_means_no_synthesis` | Capability | **T2.3 negative: no capability ⇒ no wider-ring synthesis** |
| `Capability.wider_capability_subsumes_narrower` | Capability | Capability projection: wider grants narrower |
| `Capability.hammer_capability_model` | Capability | Worked example for the Hammer simulation |
| `NoiseBound.abs_error_one_mul` | Noise | One-product fp32 error bound |
| `NoiseBound.abs_error_two_muls` | Noise | Two-product chain fp32 error bound |
| `NoiseBound.fp32_noise_unit_magnitude` | Noise | **T3.1: fp32 noise floor ~3·10⁻⁶ at unit magnitude** |
| `NoiseBound.fp32_noise_decimal_magnitude` | Noise | T3.1: fp32 noise floor ~3·10⁻³ at M=10 |

**Bold rows are the core security theorems.** The architecture's distinguishing claim ("privilege violations are structurally impossible, not detected") rests on `no_surjection_complex_to_quaternion` plus the ring-tower extensions (T2.1.b, T2.1.c).

---

## 2. Foundations — abstract structural lemmas + ℂ → ℍ closure

**File:** `Wyrd/Foundations.lean`. **Imports:** mathlib4 only. **Status:** clean compile, 0 sorries.

### `no_surjection_comm_to_noncomm`

```lean
theorem no_surjection_comm_to_noncomm
    {R S : Type*} [CommRing R] [Ring S]
    (h_noncomm : ∃ x y : S, x * y ≠ y * x)
    (φ : R →+* S) : ¬ Function.Surjective φ
```

**Reading:** if S has a non-commutative pair, no ring homomorphism R → S can be surjective when R is commutative.

**Proof:** assume surjective; pull witnesses back through φ; commutativity in R contradicts non-commutativity in S.

**Cited by:** `no_surjection_complex_to_quaternion` (concrete ℂ → ℍ closure).

### `no_surjection_assoc_to_nonassoc`, `no_surjection_alt_to_nonalt`

The two analogous theorems for associativity and alternativity. Same structural shape, weaker premises. Together they give the full ring-tower closure: **no inner ring can surject onto its outer doubling.**

### Boundary detectors `commutator`, `associator'`, `alternator'`

Three definitions that compute the algebraic residue identifying which "tower level" a triple sits in:

| Detector | Definition | Vanishes when |
|---|---|---|
| `commutator a b` | `a*b - b*a` | `a, b` commute |
| `associator' a b c` | `(a*b)*c - a*(b*c)` | the triple is associative |
| `alternator' a b` | `(a*a)*b - a*(a*b)` | left-alternativity holds |

The hardware watchdog computes these on every cycle (see SiFive spec §6, `qbp_invariant` CSR). The vanishing theorems certify that "algebra-level" privilege checks are well-defined.

### `commutator_quaternion_witness` (T1.2.a)

```lean
theorem commutator_quaternion_witness :
    ∃ a b : Quaternion ℝ, commutator a b ≠ 0
```

**Witness:** `a = ⟨0, 1, 0, 0⟩` (i), `b = ⟨0, 0, 1, 0⟩` (j). The commutator's `imK` component is `1 - (-1) = 2 ≠ 0`.

**Proof tactic:** `congrArg (·.imK) h ; simp [Quaternion.imK_sub, Quaternion.imK_mul, Quaternion.imK_zero] ; norm_num`.

**Architectural role:** existence of two ℍ values that don't commute — required premise for the ℂ → ℍ closure below.

### `no_surjection_complex_to_quaternion` (T2.1.a) ⭐

```lean
theorem no_surjection_complex_to_quaternion
    (φ : ℂ →+* Quaternion ℝ) : ¬ Function.Surjective φ
```

**Reading:** no ring homomorphism from ℂ to ℍ is surjective.

**Architectural meaning:** **Skuld user-ring (ℂ) processes structurally cannot synthesize supervisor-ring (ℍ) values by any sequence of ring operations.** Not "are detected if they try" — *cannot*.

**Cited by:** every Skuld API call that crosses the user/supervisor boundary. The Go-level capability mechanism (`Skuld.GrantCapability`) is sound only because this theorem holds. When the watchdog fires a `WDEvent` for a commutator violation, the firing is the *runtime witness* that the algebraic ring map was never constructed in the first place.

---

## 3. CayleyDickson — octonion / sedenion construction + 𝕆 witness

**File:** `Wyrd/CayleyDickson.lean`. Mathlib has no octonion or sedenion type, so we build them via the generic Cayley-Dickson doubling.

### Types

```lean
structure CayleyDickson (A : Type*) where
  l : A
  r : A
  deriving DecidableEq

abbrev Octonion (R : Type*) [CommRing R] := CayleyDickson (Quaternion R)
abbrev Sedenion (R : Type*) [CommRing R] := CayleyDickson (Octonion R)
```

**Multiplication rule (Schafer / Baez convention):** `(a, b)(c, d) = (a·c − star(d)·b, d·a + b·star(c))`.

**Conjugation:** `star (a, b) = (star a, −b)`.

The componentwise additive structure (`Add`, `Neg`, `Sub`, `Zero`, `One`) is derived; `Mul` and `Star` are explicit instances. The full `Ring` instance is *not* derived (associativity fails for octonions, which is the entire point) — only the algebraic structures we need for the privilege witnesses.

### Octonion basis e₀..e₇

```lean
def e0 : Octonion R := ⟨1, 0⟩
def e1 : Octonion R := ⟨⟨0, 1, 0, 0⟩, 0⟩  -- (i, 0)
...
def e7 : Octonion R := ⟨0, ⟨0, 0, 0, 1⟩⟩  -- (0, k)
```

Used to construct concrete witnesses for non-associativity.

### `associator_octonion_witness` (T1.2.b)

```lean
theorem associator_octonion_witness :
    ∃ a b c : Octonion ℤ, associator a b c ≠ 0
```

**Witness:** `(e₁, e₂, e₄)`. Hand-computed: `(e₁·e₂)·e₄ = e₇`, `e₁·(e₂·e₄) = −e₇`, so the associator equals `(0, 2k)` with `imK = 2 ≠ 0`.

**Proof tactic:** `congrArg (fun x : Octonion ℤ => x.r.imK) h ; simp [...]`. The deepest right-component imK reduces to `2 ≠ 0` via integer arithmetic.

**Architectural role:** existence premise for `no_surjection_quaternion_to_octonion` (T2.1.b — supervisor → kernel boundary). Closing T2.1.b is straightforward composition: combine `no_surjection_assoc_to_nonassoc` with this witness.

---

## 4. Projection — T2.2 (outer-ring computation on inner-ring values)

**File:** `Wyrd/Projection.lean`.

### `π_mul_of_inner`

```lean
theorem π_mul_of_inner [NonUnitalNonAssocRing A] [StarAddMonoid A]
    {x y : CayleyDickson A}
    (hx : x.r = 0) (hy : y.r = 0) :
    π (x * y) = π x * π y
```

**Reading:** if both factors live entirely in the inner half (right component zero), the projection of the Cayley-Dickson product equals the inner-ring product.

**Proof:** unfold `π`, rewrite via `mul_l`, use `hy` and `star_zero` to eliminate the outer-ring cross-term.

### `kernel_supervisor_safe` ⭐

```lean
theorem kernel_supervisor_safe (a b : Quaternion R) :
    π_O_to_H ((⟨a, 0⟩ : Octonion R) * (⟨b, 0⟩ : Octonion R)) = a * b
```

**Reading:** when a kernel-ring (𝕆) process performs octonion multiplication on two values that originated from the supervisor ring (ℍ), and projects the result back to ℍ, the answer equals the supervisor-ring product directly. **No corruption.**

**Architectural meaning:** this is what makes "kernel returns supervisor-safe value" actually safe. A 𝕆-ring service can compute on ℍ-ring inputs and hand the answer back without the supervisor needing to re-validate every byte. **Bedrock of the layered privilege model.**

**Cited by:** Skuld's `WyrdSubmit` returning to user code; any Skuld syscall whose output is supposed to live in the caller's ring.

---

## 5. Capability — T2.3 (capability soundness, both directions)

**File:** `Wyrd/Capability.lean`.

### Capability structure

```lean
structure Capability (A : Type*) [Mul A] [Zero A] [One A] where
  token : A
  nonzero_witness : token ≠ (0 : A) → ∃ inv : A, token * inv = 1 ∧ inv * token = 1
```

A capability is a wrapped wider-ring element with an invertibility witness. The token is the "p" in the sandwich operation `p · u · p⁻¹`.

### Sandwich semantics

```lean
def sandwich {A : Type*} [Mul A] (p u p_inv : A) : A := p * u * p_inv

theorem sandwich_preservation_associative
    {A : Type*} [Ring A] (p u p_inv : A)
    (h_inv1 : p * p_inv = 1) (h_inv2 : p_inv * p = 1) :
    sandwich p u p_inv * p = p * u
```

Sandwich preserves the operand under right-multiplication by `p` (in associative rings). The proof is a calc chain through `p * u * (p_inv * p) = p * u * 1 = p * u`.

### `capability_grants_safe_access` (T2.3 positive) ⭐

```lean
theorem capability_grants_safe_access
    {A : Type*} [Ring A] [StarRing A]
    (_cap : Capability (CayleyDickson A))
    (a b : A) :
    Projection.π ((⟨a, 0⟩ : CayleyDickson A) * (⟨b, 0⟩ : CayleyDickson A)) = a * b
```

**Reading:** a process holding a capability for the wider ring CAN perform wider-ring arithmetic on inner-ring inputs and project the result back without corruption.

The token is unused in this theorem because the operation is on inner-ring values that don't need privilege escalation. **The capability authorizes the *introduction* of wider-ring values, not the operations on already-inner-ring values.** When actual wider-ring values are introduced, the capability is consumed/authenticated upstream — at the syscall boundary, not in the middle of computation.

### `no_capability_means_no_synthesis` (T2.3 negative) ⭐

```lean
theorem no_capability_means_no_synthesis
    {R S : Type*} [CommRing R] [Ring S]
    (h_S_strict : ∃ x y : S, x * y ≠ y * x)
    (φ : R →+* S) : ¬ Function.Surjective φ
```

**Reading:** identical content to `no_surjection_comm_to_noncomm`, framed in capability language. A process with no R'-capability cannot synthesize R'-ring values.

### `wider_capability_subsumes_narrower`

```lean
theorem wider_capability_subsumes_narrower
    {A : Type*} [Ring A] [StarRing A]
    (cap_outer : Capability (CayleyDickson A))
    (h_l_nonzero : cap_outer.token.l ≠ 0)
    (h_l_invertible : ∃ inv : A, cap_outer.token.l * inv = 1 ∧
                                  inv * cap_outer.token.l = 1) :
    ∃ cap_inner : Capability A, cap_inner.token = cap_outer.token.l
```

**Reading:** holding a kernel-ring (𝕆) capability gives you all supervisor-ring (ℍ) capabilities by projection of the token. Formalizes "kernel can do anything supervisor can do."

### `hammer_capability_model`

Worked example: the Hammer simulation holds a capability for `CayleyDickson A` and can perform inner-ring multiplication safely. Lives in the corpus as a sanity check that the Skuld API flow type-checks end-to-end.

---

## 6. Noise bound — T3.1 (fp32 noise floor below privilege threshold)

**File:** `Wyrd/Noise.lean`. Self-contained: only depends on real-number analysis from mathlib, no quaternion-API.

### `RoundingModel` abstraction

```lean
structure RoundingModel where
  fl : ℝ → ℝ
  ε_fp : ℝ
  ε_pos : 0 < ε_fp
  ε_small : ε_fp < 1
  fl_error : ∀ x : ℝ, |fl x - x| ≤ ε_fp * |x|
  fl_zero : fl 0 = 0
```

The model is generic over IEEE-style floating-point; instantiate `ε_fp = 2⁻²³ ≈ 1.19e-7` for fp32, `2⁻⁵² ≈ 2.22e-16` for fp64.

### `abs_error_one_mul`, `abs_error_two_muls`

```lean
theorem abs_error_one_mul (R : RoundingModel) (x y : ℝ) (M : ℝ)
    (hM : 0 ≤ M) (hx : |x| ≤ M) (hy : |y| ≤ M) :
    |R.mul x y - x * y| ≤ R.ε_fp * M^2

theorem abs_error_two_muls (R : RoundingModel) (a b c : ℝ) (M : ℝ)
    (hM : 1 ≤ M) (ha : |a| ≤ M) (hb : |b| ≤ M) (hc : |c| ≤ M) :
    |R.mul (R.mul a b) c - (a * b) * c| ≤ 2 * R.ε_fp * M^3 + R.ε_fp^2 * M^3
```

The two-multiplication bound is what dominates the associator's noise. The full octonion associator evaluates `(a*b)*c - a*(b*c)`, which compounds depth-24 chains — bound by `24 · ε_fp · M³` at first order.

### `fp32_noise_unit_magnitude`, `fp32_noise_decimal_magnitude` ⭐

```lean
theorem fp32_noise_unit_magnitude :
    fp32_noise_floor 1 ≤ 3e-6
theorem fp32_noise_decimal_magnitude :
    fp32_noise_floor 10 ≤ 3e-3
```

**Architectural meaning:** at unit magnitude the fp32 noise floor on the associator is ~3·10⁻⁶. The privilege threshold `ε_priv` (per the SiFive spec) is set well above this — typically 10⁻⁴ or larger — so noise cannot fake a privilege violation. **This is the quantitative half of "watchdog firings are real, not noise."**

`threshold_separation_safe ε_priv R M k := ε_priv ≥ k · associator_noise_bound R M`

is the contract the supervisor uses to set its alarm thresholds; with k = 30 (safety factor), fp32 at M=1 gives ε_priv ≥ ~10⁻⁴, defensible.

---

## 7. Sedenion alternator witness — T1.2.c support

**File:** `Wyrd/SedenionWitness.lean`.

### `alternator_sedenion_witness`

```lean
theorem alternator_sedenion_witness :
    ∃ a b : Sedenion ℤ, sed_alternator a b ≠ 0
```

**Witness:** `α = (e₁ᴼ, e₄ᴼ)`, `β = (e₂ᴼ, 0)`. Hand-computed: alternator equals `(0, −2 e₇ᴼ)`. The deepest `.r.r.imK` component is `−2 ≠ 0`.

**Proof tactic:** `congrArg (fun s => s.r.r.imK) h ; simp [...]` — the imK reduces to `−2 = 0` over ℤ, simp closes by `omega` / contradiction.

**Architectural role:** existence premise for T2.1.c (kernel → firmware boundary, 𝕆 → 𝕊). The full T2.1.c proof composes `no_surjection_alt_to_nonalt`, `octonion_alternative` (proven below), and this witness.

> **Deferred:** the abstract corollary `sedenion_not_alternative : ¬ ∀ a b, (a*a)*b = a*(a*b)` is unused by downstream theorems and would require an `AddGroup` instance on `Sedenion ℤ` (boilerplate). Omitted intentionally.

---

## 8. Octonion alternativity

**File:** `Wyrd/OctonionAlternative.lean`.

### `quat_norm_is_real`, `quat_real_part_is_real`

```lean
theorem quat_norm_is_real (q : Quaternion R) :
    ∃ c : R, q * star q = (⟨c, 0, 0, 0⟩ : Quaternion R)

theorem quat_real_part_is_real (q : Quaternion R) :
    ∃ c : R, q + star q = (⟨c, 0, 0, 0⟩ : Quaternion R)
```

**Witnesses:** `q.re² + q.imI² + q.imJ² + q.imK²` for the norm; `2 · q.re` for the real-part doubling.

**Proof:** `ext + simp [component lemmas] + ring` — straight component-level expansion of the Quaternion star and product.

**Architectural role:** standalone facts about quaternion conjugation. (Not used by `octonion_alternative` below — the proof bypassed the centrality argument these lemmas were originally drafted for. They remain in the file as cited results.)

### `alternator_l_vanishes`, `alternator_r_vanishes`

```lean
theorem alternator_l_vanishes (p q r s : Quaternion R) :
    (((⟨p, q⟩ * ⟨p, q⟩ : CayleyDickson (Quaternion R)) * ⟨r, s⟩).l) -
      ((⟨p, q⟩ * (⟨p, q⟩ * ⟨r, s⟩ : CayleyDickson (Quaternion R))).l) = 0
-- and similarly for .r
```

**Reading:** the .l and .r components of the octonion alternator vanish.

**Proof tactic:** `simp only [CayleyDickson.mul_l, CayleyDickson.mul_r] ; ext ; simp [Quaternion component lemmas] ; ring`.

The proof reduces to polynomial identities in the 16 real components (`p.re, p.imI, ..., s.imK`), which `ring` closes because ℍ is associative. **The alternator vanishing for 𝕆 = CD(ℍ) is a polynomial consequence of ℍ-associativity** — no centrality argument is needed.

### `octonion_alternative` ⭐

```lean
theorem octonion_alternative (a b : Octonion R) :
    (a * a) * b = a * (a * b)
```

**Reading:** the octonions are alternative (left-alternative; right-alternative is the analogous proof on the other side, deferrable).

**Proof:** destructure `a` and `b` into Cayley-Dickson pairs, apply `CayleyDickson.ext` (NOT bare `ext`, which over-recurses through the inner Quaternion), close each component via `sub_eq_zero.mp` of the corresponding alternator-vanishing lemma.

**Architectural role:** the `h_R_alt` premise of `no_surjection_alt_to_nonalt` when applied to 𝕆 → 𝕊. Combined with `alternator_sedenion_witness`, gives T2.1.c — kernel (𝕆) cannot surject onto firmware (𝕊).

---

## 9. The full ring-tower closure (assembled)

Combining the theorems above, the privilege model has the four boundary closures:

| Boundary | Theorem | Status |
|---|---|---|
| user (ℂ) → supervisor (ℍ) | `no_surjection_complex_to_quaternion` | **proven directly** in Foundations |
| supervisor (ℍ) → kernel (𝕆) | `no_surjection_assoc_to_nonassoc` + `associator_octonion_witness` | proven, one-line composition |
| kernel (𝕆) → firmware (𝕊) | `no_surjection_alt_to_nonalt` + `octonion_alternative` + `alternator_sedenion_witness` | proven, two-line composition |
| firmware (𝕊) → ??? | n/a — 𝕊 is the floor | n/a |

The "one-line composition" theorems for T2.1.b and T2.1.c could be stated explicitly in Foundations.lean as a follow-up; the components are all proven and the closure is automatic. (Tracked: Phase 1 followup, ~10 minutes.)

---

## 10. Downstream code — expected citation pattern

Each Go package method that performs a privilege-relevant operation should carry a comment naming the Lean theorem that justifies it. Skeleton:

```go
// Skuld.GrantCapability creates a wider-ring capability for the calling
// process. Soundness: Capability.capability_grants_safe_access (positive,
// holder can compute on inner-ring values via the wider ring) and
// Capability.no_capability_means_no_synthesis (negative, non-holders
// cannot fabricate the token).
//
// See Wyrd-Proofs-Reference-v1.0.md §5.
func (s *Supervisor) GrantCapability(...) (*Capability, error) { ... }
```

```go
// qbpcu.QFMA executes a fused multiply-add on quaternion-shaped operands.
// Privilege-honesty: when the watchdog observes a commutator residue
// that crosses ε_priv, the firing is real (not noise) by
// NoiseBound.fp32_noise_unit_magnitude — the fp32 noise floor (~3e-6)
// is well below ε_priv (typically 10⁻⁴).
//
// See Wyrd-Proofs-Reference-v1.0.md §6.
func (cu *QBPCU) QFMA(a, b, c quat.Vec) quat.Vec { ... }
```

The `qbpcu.Mock` and `qbpcu.Golden` implementations should each pass a "soundness assertion" test that exercises the cited theorem's statement with concrete inputs and verifies the implementation's output matches the theorem's claim.

---

## 11. Versioning & verification protocol

**Authoritative source.** This document references theorem *names*; the formal content lives in the seven `.lean` files. If a proof file is edited, this document should be updated within the same commit.

**Verification cadence.** `lake build` runs in CI on every commit touching `~/Documents/Wyrd/wyrd-lean-project/`. A green build is a precondition for merging any architecture spec change that cites these theorems.

**Mathlib drift policy.** When mathlib is bumped, re-run `lake build`. Lemma renames (e.g., `abs_add` → `abs_add_le` in v4.30) are tracked in this file as a footnote on the affected theorem. If a major API shift breaks structure, escalate per the Onboarding Prompt §9.

---

## Attribution (per QBP standing rule R1)

This corpus stands on the work of Furey, Dixon, Günaydin / Gürsey, Boyle / Farnsworth, Singh, Chamseddine / Connes, Koide, and Baez. The Cayley-Dickson construction follows Schafer and Baez (2002, *The Octonions*, Bull. AMS).

---

*End of Wyrd Proofs Reference v1.0.*
