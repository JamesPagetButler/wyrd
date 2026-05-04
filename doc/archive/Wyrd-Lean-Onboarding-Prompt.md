# Lean Verification Onboarding — Wyrd / Skuld Project

## A self-contained briefing for a fresh Claude instance taking over the formal proofs

**Helpful Engineering — Quaternion-Based Physics Programme**
April 2026

---

## 0. About this prompt

You are picking up the Lean 4 verification work for the Wyrd / Skuld privilege model. A previous Claude instance did the architectural design and produced a corpus of proofs — partially compiled, partially source-verified, partially left at `sorry`. Your job is to drive every `sorry` to zero and replace every standing `axiom` with its mathlib4 equivalent, in a live Lean environment.

This is real work for a real project. The mathematics is settled by hand; you're doing the toolchain and tactic refinement. The corpus is described concretely below — file by file, gap by gap, with the mathematical content of each gap pre-derived so you don't have to repeat it.

You should expect this to take roughly 1–2 hours of focused work in a Lean environment with mathlib4 available. If you hit something that doesn't go through in 15 minutes of tactic exploration, **stop and ask** — escalation criteria are in §9.

---

## 1. What this project is (5-minute version)

**Wyrd** is a quaternion-native hypergraph database, part of the Quaternion-Based Physics (QBP) programme run by James Paget Butler at Helpful Engineering. **Skuld** is its supervisor — the policy and mediation layer between user processes and Wyrd. The QBP-CU is a custom RISC-V accelerator block. Together they're a research instrument for QBP physics.

**The novel architectural claim** is that privilege boundaries between user / supervisor / kernel rings are *algebraically enforced* rather than policy-enforced. The privilege rings correspond to the Cayley-Dickson algebra tower:

- Ring 3 (user) = ℂ — complex numbers
- Ring 2 (supervisor) = ℍ — quaternions
- Ring 1 (kernel) = 𝕆 — octonions
- Ring 0 (firmware) = 𝕊 — sedenions

A user-ring (ℂ) process **structurally cannot** synthesize supervisor-ring (ℍ) values because ℍ contains generators (j, k) absent from ℂ. The hardware watchdog detects boundary crossings via algebraic invariants (commutator at ℂ→ℍ, associator at ℍ→𝕆, alternator at 𝕆→𝕊).

**Your work** proves this formally. The proofs are the foundation of the security claim — without them, the architecture's distinguishing property reduces to a slogan.

---

## 2. Why this matters

Three reasons the formal verification is non-optional:

1. **The architecture's distinguishing claim depends on it.** Every other competitor in this space (DBOS, seL4, CHERI) has policy-checked or hardware-tagged capabilities. Wyrd's claim to be different rests on "the privilege violation is structurally impossible, not detected." Without proofs, that's just words.

2. **The QBP-Quantum work established a precedent.** James's QBP-Quantum module shipped seven Lean-verified theorems with zero sorries. Wyrd is held to the same standard.

3. **Capability mediation in Crawl uses HMAC tokens; in Walk it migrates to Wyrd-native capabilities. The migration is sound only if the formal model is sound.** Skuld's `ProcGrantCapability` API at Crawl is a temporary measure; the real mechanism (capability-as-Wyrd-node) at Walk requires T2.3 to actually hold.

---

## 3. The corpus (you don't have to read everything)

Eight Lean files plus support material. The full corpus index is in *Wyrd-Corpus-Index-v1.0.md*, but for your work, these are the files that matter:

| File | Status | Your action |
|---|---|---|
| `Wyrd-Algebraic-Privilege-Proofs-v0.4.lean` | Source-verified, untested | Compile, verify, fix any drift |
| `Wyrd-CayleyDickson-Types-v0.1.lean` | Needs API update | Update to mathlib4 3-param `QuaternionAlgebra`, then compile |
| `Wyrd-T2.2-Projection-v0.1.lean` | Should be clean | Compile, verify |
| `Wyrd-T2.3-Capability-Soundness-v0.1.lean` | Should be clean | Compile, verify |
| `Wyrd-T3.1-Noise-Bound-v0.2.lean` | Zero sorries | Compile, verify (independent of quaternion API) |
| `Wyrd-Sedenion-Alternator-Witness-v0.1.lean` | 1 sorry | Close destructuring tactic |
| `Wyrd-Octonion-Alternativity-v0.1.lean` | 2 sorrys + 2 axioms | Replace axioms, refine ring_nf |
| **(NEW)** Master proof file | doesn't exist yet | Create one that imports all the above |

**Don't read more than you need.** The architectural docs (Wyrd-Supervisor-Architecture, Skuld-Spec, QBP-CU specs) are context, not your working material. Read them only if a proof's *interpretation* is unclear, not its *content*.

---

## 4. Your specific job: drive sorries to zero

Concrete deliverables, in priority order:

### Deliverable 1: A working `lake` project

Create a fresh Lean 4 project with mathlib4 dependency. Set up the directory structure so the proof files compile in dependency order:

```
Wyrd/
├── lakefile.lean        — depends on mathlib
├── lean-toolchain
├── Wyrd.lean            — top-level imports
└── Wyrd/
    ├── CayleyDickson.lean         (← Wyrd-CayleyDickson-Types-v0.1.lean)
    ├── Foundations.lean           (← Wyrd-Algebraic-Privilege-Proofs-v0.4.lean)
    ├── Projection.lean            (← Wyrd-T2.2-Projection-v0.1.lean)
    ├── Capability.lean            (← Wyrd-T2.3-Capability-Soundness-v0.1.lean)
    ├── Noise.lean                 (← Wyrd-T3.1-Noise-Bound-v0.2.lean)
    ├── SedenionWitness.lean       (← Wyrd-Sedenion-Alternator-Witness-v0.1.lean)
    └── OctonionAlternative.lean   (← Wyrd-Octonion-Alternativity-v0.1.lean)
```

The internal cross-references in the existing files use placeholder paths like `Wyrd.Wyrd_CayleyDickson_Types_v0_1`; update these to match the project structure.

`lakefile.lean` template:
```lean
import Lake
open Lake DSL

package wyrd where
  -- add package configuration options here

require mathlib from git
  "https://github.com/leanprover-community/mathlib4.git"

@[default_target]
lean_lib Wyrd where
  -- add library configuration options here
```

`lean-toolchain` should match the mathlib4 version you're targeting.

### Deliverable 2: `Foundations.lean` compiles

This is `Wyrd-Algebraic-Privilege-Proofs-v0.4.lean`. It was source-verified against mathlib4 master at HEAD as of April 2026. The expected outcome: clean compile with zero sorries.

If it doesn't compile, the most likely issues:
- `Quaternion.imK_mul` may have moved to a different namespace
- The `simp` set may need adjustment (current: `[Quaternion.imK_sub, Quaternion.imK_mul]`)
- Anonymous constructor `⟨0, 1, 0, 0⟩ : Quaternion ℝ` may need explicit type elaboration

The mathematics is known correct (see §6 below).

### Deliverable 3: `CayleyDickson.lean` updated and compiling

The existing v0.1 file uses the *2-parameter* `QuaternionAlgebra R c₁ c₂` from mathlib3. **Mathlib4 uses 3-parameter `QuaternionAlgebra R c₁ c₂ c₃`**, with `Quaternion R := QuaternionAlgebra R (-1) 0 (-1)`.

Concrete updates:
1. Where `Quaternion R` literals appear, use the 3-parameter form or rely on the `Quaternion` shorthand.
2. The Cayley-Dickson construction's product formula references `star` of inner-ring elements; ensure `Quaternion R` has the expected `Star` instance via mathlib's `QuaternionAlgebra.instStarRing`.
3. `Quaternion.ext`, `Quaternion.imK_mul` and friends are namespaced as in §5 below.

**Do NOT redesign the Cayley-Dickson construction itself.** It's correct in form. Only the mathlib API surface changes.

### Deliverable 4: T2.2, T2.3, T3.1 compile cleanly

These files have no quaternion-algebra-specific content (or use only the abstract structures). They should compile after CayleyDickson.lean does. Estimated time: 15 minutes for any incidental fixes.

### Deliverable 5: SedenionWitness.lean — close the destructuring `sorry`

The mathematics is verified (see §7 below). The witness is α = ⟨e₁ᴼ, e₄ᴼ⟩, β = ⟨e₂ᴼ, 0⟩ in `Sedenion ℤ`, with alternator computing to `⟨0, -2 · e₇ᴼ⟩`.

The remaining `sorry` is the Lean tactic to extract the relevant component and apply `linarith` or `decide`. Two strategies:

**Strategy A (preferred, mechanical):** Use `decide`. `Sedenion ℤ` has decidable equality (`CayleyDickson` derives it; `Quaternion ℤ` has it via `ℤ`). The kernel reduction is large but finite.

```lean
theorem alternator_sedenion_witness :
    ∃ a b : Sedenion ℤ, sed_alternator a b ≠ 0 := by
  refine ⟨α_witness, β_witness, ?_⟩
  decide
```

If `decide` times out, fall back to Strategy B.

**Strategy B (fallback, explicit):** Extract the deep imK component via repeated field access, show it equals -2, conclude.

```lean
theorem alternator_sedenion_witness :
    ∃ a b : Sedenion ℤ, sed_alternator a b ≠ 0 := by
  refine ⟨α_witness, β_witness, ?_⟩
  intro h
  have h_imK : (sed_alternator α_witness β_witness).r.r.imK = 0 := by
    rw [h]; rfl
  unfold sed_alternator α_witness β_witness Octonion.e1 Octonion.e2 Octonion.e4 at h_imK
  simp [CayleyDickson.mul_l, CayleyDickson.mul_r, ...] at h_imK
  -- h_imK : -2 = 0 in ℤ
  linarith
```

The exact field-access path (`s.r.r.imK` etc.) depends on how `CayleyDickson` and `Quaternion` are nested; you'll see the type at each step in the Lean infoview.

### Deliverable 6: OctonionAlternative.lean — close the two `ring_nf` sorries and replace two axioms

This is the largest remaining task — estimate 30–60 minutes.

**The two axioms:** stand-ins for known mathlib4 results about quaternion conjugation. Replace with:

```lean
-- WAS: axiom quat_norm_is_real
-- USE: Quaternion.normSq plus Quaternion.coe_normSq_eq_self_mul_star (or current name)
-- The fact you actually need: q * star q = (Quaternion.normSq q : Quaternion R)
-- and (normSq q : Quaternion R) is in the image of the coercion ℝ → ℍ.

-- WAS: axiom quat_real_part_is_real
-- USE: QuaternionAlgebra.self_add_star (gives a + star a = 2 * a.re as a real-coerced quaternion)
```

If the exact mathlib4 lemma name has drifted, search:
```bash
cd /path/to/mathlib4 && grep -rn "self_add_star\|normSq_eq" Mathlib/Algebra/Quaternion.lean
```

**The two sorries:** in `alternator_l_vanishes` and `alternator_r_vanishes`. The mathematics is fully expanded in §8 below. The Lean tactic chain you need:

1. `simp only [CayleyDickson.mul_l, CayleyDickson.mul_r, ...]` to unfold to inner-quaternion arithmetic
2. Apply the centrality of `q * star q` (via the replaced axiom): rewrite `(star q * q) * x` to `x * (star q * q)`
3. Apply the centrality of `q + star q` (via the replaced axiom): rewrite `(p + star p) * x` to `x * (p + star p)`
4. `ring` to close

If `ring` doesn't close after the rewrites, the issue is almost certainly that one of the centrality applications didn't fire. Print the goal state, identify the term that should commute but isn't being recognized, and add an explicit `rw` step.

### Deliverable 7: A master proof file with zero sorries, zero axioms, zero warnings

Create `Wyrd.lean` at the project root:

```lean
import Wyrd.CayleyDickson
import Wyrd.Foundations
import Wyrd.Projection
import Wyrd.Capability
import Wyrd.Noise
import Wyrd.SedenionWitness
import Wyrd.OctonionAlternative
```

Run `lake build` to completion. Confirm:
- Zero `sorry`s in any file
- Zero `axiom` declarations outside of mathlib's standard ones
- Zero compilation warnings beyond the inevitable "unused variable" type
- All claimed theorems compile

### Deliverable 8: A short status report

Plain markdown, ~500 words. Document:
- What you did
- What unexpected issues came up
- What's now verified
- What (if anything) you couldn't close in this session and why
- Recommended next steps

---

## 5. Mathlib4 API cheat sheet (verified at HEAD, April 2026)

These names were source-verified by inspecting `Mathlib/Algebra/Quaternion.lean`:

```lean
-- The structure (3-parameter)
structure QuaternionAlgebra (R : Type*) (a b c : R) where
  re : R
  imI : R
  imJ : R
  imK : R

-- The shorthand
def Quaternion (R : Type*) := QuaternionAlgebra R (-1) 0 (-1)
notation "ℍ[" R "]" => Quaternion R

-- Extensionality (in the Quaternion namespace)
theorem Quaternion.ext :
  a.re = b.re → a.imI = b.imI → a.imJ = b.imJ → a.imK = b.imK → a = b

-- Component-of-product (Quaternion namespace; for Quaternion R specifically,
-- with c₂=0, the cross terms simplify out)
theorem Quaternion.re_mul :
  (a * b).re = a.re*b.re - a.imI*b.imI - a.imJ*b.imJ - a.imK*b.imK
theorem Quaternion.imI_mul :
  (a * b).imI = a.re*b.imI + a.imI*b.re + a.imJ*b.imK - a.imK*b.imJ
theorem Quaternion.imJ_mul :
  (a * b).imJ = a.re*b.imJ - a.imI*b.imK + a.imJ*b.re + a.imK*b.imI
theorem Quaternion.imK_mul :
  (a * b).imK = a.re*b.imK + a.imI*b.imJ - a.imJ*b.imI + a.imK*b.re

-- Componentwise sub/add (definitionally rfl)
@[simp] theorem Quaternion.imK_sub : (a - b).imK = a.imK - b.imK := rfl
@[simp] theorem Quaternion.imK_add : (a + b).imK = a.imK + b.imK := rfl

-- Star ring
instance Quaternion.instStarRing : StarRing (Quaternion R)
-- Standard star lemmas (star_zero, star_mul, star_add) all available
```

For octonions and sedenions, mathlib4 has nothing — that's why `Wyrd-CayleyDickson-Types-v0.1.lean` exists. Don't try to find them in mathlib.

**If a name has drifted** (mathlib4 moves fast), search the cloned source:
```bash
cd mathlib4 && grep -rn "name_pattern" Mathlib/Algebra/Quaternion.lean
```

---

## 6. Mathematical content for the gaps

So you don't have to re-derive these.

### Gap 6a: T1.2.a — commutator of i, j in ℍ is nonzero

For `Quaternion R = QuaternionAlgebra R (-1) 0 (-1)`, with c₁=-1, c₂=0, c₃=-1:

Let i = ⟨0, 1, 0, 0⟩, j = ⟨0, 0, 1, 0⟩. By `Quaternion.imK_mul`:

```
(i * j).imK = 0*0 + 1*1 - 0*0 + 0*0 = 1
(j * i).imK = 0*0 + 0*0 - 1*1 + 0*0 = -1
```

Then `(i*j - j*i).imK = 1 - (-1) = 2 ≠ 0`.

The proof:
```lean
theorem commutator_quaternion_witness :
    ∃ a b : Quaternion ℝ, commutator a b ≠ 0 := by
  refine ⟨⟨0, 1, 0, 0⟩, ⟨0, 0, 1, 0⟩, ?_⟩
  intro h
  have h_imK : (commutator (⟨0, 1, 0, 0⟩ : Quaternion ℝ) ⟨0, 0, 1, 0⟩).imK = 0 := by
    rw [h]; rfl
  unfold commutator at h_imK
  simp only [Quaternion.imK_sub, Quaternion.imK_mul] at h_imK
  norm_num at h_imK
```

### Gap 6b: T1.2.b — associator witness in 𝕆

In `Octonion R := CayleyDickson (Quaternion R)`, with the basis e₀..e₇ defined in `Wyrd-CayleyDickson-Types-v0.1.lean`:

Take a = e₁, b = e₂, c = e₄. Compute associator (a*b)*c − a*(b*c):

```
e₁*e₂ = (i*j, 0) = (k, 0) = e₃
(e₁*e₂)*e₄ = (k, 0)*(0, 1) = (k*0 - star(1)*0, 1*k + 0*star(0)) = (0, k) = e₇

e₂*e₄ = (j, 0)*(0, 1) = (0, j) = e₆
e₁*(e₂*e₄) = (i, 0)*(0, j) = (i*0 - star(j)*0, j*i + 0*star(i)) = (0, j*i) = (0, -k) = -e₇

associator = e₇ - (-e₇) = 2*e₇ ≠ 0
```

The closing tactic mirrors 6a but operates on the deeper structure.

### Gap 6c: T2.1.c — sedenion alternator witness

In `Sedenion ℤ := CayleyDickson (Octonion ℤ)`:

Take α = ⟨e₁ᴼ, e₄ᴼ⟩, β = ⟨e₂ᴼ, 0⟩. Hand calculation (full derivation in `Wyrd-Sedenion-Alternator-Witness-v0.1.lean` docstring):

```
α*α  = (e₁²ᴼ - star(e₄)*e₄, e₄*e₁ + e₄*star(e₁))
      = (-1ᴼ - 1ᴼ, 0)  [since e₄² = -1, e₄*star(e₁) cancels e₄*e₁ via the antisymmetry]
      Actually compute: star(e₄ᴼ) = -e₄ᴼ, so star(e₄)*e₄ = -e₄² = -(-1) = 1
      And e₄*e₁ + e₄*star(e₁) = e₄*e₁ + e₄*(-e₁) = 0. ✓
      So α*α = (-2 · 1ᴼ, 0)

(α*α)*β = (-2ᴼ, 0)*(e₂, 0) = (-2*e₂ᴼ, 0)

α*β = (e₁, e₄)*(e₂, 0) = (e₁*e₂ - 0, 0*e₁ + e₄*star(e₂)) = (e₃, e₄*(-e₂))
     e₄*e₂ in 𝕆 = (0,1)*(j,0) = (0, j) = e₆ ... wait: (0,1)*(j,0) = (0*j - star(0)*1, 0*0 + 1*star(j)) = (0, -j) = -e₆
     Hmm let me redo: actually e₄*e₂ should be computed via CD formula
     e₄ = (0,1), e₂ = (j,0)
     l = 0*j - star(0)*1 = 0
     r = 0*0 + 1*star(j) = -j
     e₄*e₂ = (0, -j) = -e₆
     So e₄*(-e₂) = e₆
α*β = (e₃, e₆)

α*(α*β) = (e₁, e₄)*(e₃, e₆) — computed in detail in the SedenionWitness docstring
        = (-2*e₂ᴼ, 2*e₇ᴼ)

alternator = (α*α)*β - α*(α*β) = (-2e₂, 0) - (-2e₂, 2e₇) = (0, -2e₇)
```

The relevant component for the contradiction is the imK of the imaginary part of the upper octonion of the upper sedenion = -2.

### Gap 6d: Octonion alternativity

Goal: ∀ a b : Octonion R, (a*a)*b = a*(a*b).

Let a = (p, q), b = (r, s). Compute both sides componentwise. Two key facts about the inner-ring (ℍ):
- (FACT-1) star(q) * q is real (lies in the scalar subfield), hence central.
- (FACT-2) p + star(p) is real, hence central.

**First component** of (a*a)*b:
```
(p*p - star(q)*q) * r - star(s) * (q*p + q*star(p))
= p² r - star(q)*q*r - star(s)*q*(p + star(p))
= p² r - r*star(q)*q - (p + star(p))*star(s)*q       [by centrality]
= p² r - r*|q|² - 2*Re(p)*star(s)*q
```

**First component** of a*(a*b):
```
p*(pr - star(s)*q) - star(sp + q*star(r)) * q
= p²r - p*star(s)*q - (star(p)*star(s) + r*star(q))*q
= p²r - p*star(s)*q - star(p)*star(s)*q - r*star(q)*q
= p²r - r*|q|² - (p + star(p))*star(s)*q
= p²r - r*|q|² - 2*Re(p)*star(s)*q   ✓
```

Equal. Similar computation for the second component (also matches).

The Lean tactic skeleton:

```lean
theorem octonion_alternative (a b : Octonion R) :
    (a * a) * b = a * (a * b) := by
  rcases a with ⟨p, q⟩
  rcases b with ⟨r, s⟩
  ext
  · -- first component
    simp only [CayleyDickson.mul_l, CayleyDickson.mul_r]
    -- Now: goal is in the inner ring (ℍ).
    -- Need: (star q * q) commutes with everything; (p + star p) commutes with everything.
    have h_normSq : ∃ c : R, star q * q = (c : Quaternion R) := by
      -- Use Quaternion.normSq machinery
      sorry  -- TODO: real proof
    have h_trace : ∃ c : R, q + star q = (c : Quaternion R) := by
      sorry  -- TODO: real proof
    -- After establishing centrality, the rest is `ring` (or extensive `rw` + `ring`)
    sorry
  · -- second component, similar structure
    sorry
```

The mathlib4 lemmas you need to replace the inner sorries (search by grep if names drift):
- `Quaternion.coe_normSq_eq_mul_star` or similar — for "q * star q = scalar"
- `Quaternion.self_add_star` or `QuaternionAlgebra.self_add_star` — for "q + star q = scalar"

If these don't exist verbatim, prove the local versions directly:
```lean
example (q : Quaternion R) : q * star q = ⟨q.re*q.re + q.imI*q.imI + q.imJ*q.imJ + q.imK*q.imK, 0, 0, 0⟩ := by
  ext <;> simp [Quaternion.re_mul, Quaternion.imI_mul, ...] <;> ring
```

Then use `simp` to push the central elements past non-central ones.

---

## 7. Working procedure

### Setup (~15 min)

```bash
cd /your/workspace
mkdir Wyrd && cd Wyrd
lake new wyrd
# Edit lakefile.lean to add mathlib dependency
# Create lean-toolchain matching mathlib's
lake exe cache get   # download precompiled mathlib oleans
lake build           # confirm build works
```

If `lake exe cache get` fails or times out, you'll be building mathlib from source — budget 30–60 minutes for that on a typical machine. On a slow machine, it may not finish in the session — that's an escalation case (§9).

### Compile order (~30–60 min for clean files, more if drift)

1. `Wyrd/CayleyDickson.lean` — first because everything else depends on it
2. `Wyrd/Foundations.lean` (the v0.4 master file)
3. `Wyrd/Projection.lean` (T2.2)
4. `Wyrd/Capability.lean` (T2.3)
5. `Wyrd/Noise.lean` (T3.1)
6. `Wyrd/SedenionWitness.lean`
7. `Wyrd/OctonionAlternative.lean`

After each compiles cleanly, move on. If one breaks, fix it before continuing — downstream files depend on the upstream ones.

### Tactic refinement loop

When a `sorry` or `ring_nf` doesn't close:

1. Print the goal state with `show ?_` or just inspect via the Lean infoview
2. Identify what's blocking: a missing lemma? A simp set issue? Wrong rewrite direction?
3. Try one targeted change at a time
4. If 15 minutes pass without progress, **escalate** (§9). Don't burn an hour grinding on a tactic; the issue is more likely an API name drift or a subtle structure mismatch than a real proof gap.

### Final verification

```bash
lake build
# Should produce zero errors, zero `sorry` warnings.

grep -rn "sorry\|admit" Wyrd/    # should return nothing in proof bodies
grep -rn "^axiom" Wyrd/          # should return nothing user-defined
```

---

## 8. What success looks like

A `lake build` of the Wyrd project completes with:
- All 7 proof files compiled
- Zero `sorry` in any proof body
- Zero user-defined `axiom` declarations
- Theorems named per the existing files (no signature changes)
- A short status-report markdown documenting the work

The corpus then represents a **complete formal foundation for the Wyrd / Skuld user → supervisor → kernel → firmware privilege model**, with every algebraic claim machine-verified.

---

## 9. When to escalate (don't hide problems)

Stop and report back to James (or whoever is running this session) if:

- `lake exe cache get` fails persistently and `lake build` would take > 1 hour from source
- A sorry or `ring_nf` doesn't yield to standard refinement after 15 minutes of focused work
- A mathlib API name has drifted in a way that requires more than a rename to fix
- The existing proof structure appears mathematically incorrect (very unlikely — the math has been hand-verified)
- You discover that one of the documents references a fact that doesn't actually hold (extremely unlikely; flag immediately)

Escalation is not failure. It is the correct action when a 15-minute investment isn't producing progress. Burning time silently is worse than asking.

---

## 10. Standing rules (QBP-wide)

These apply to all QBP work, including this session:

**(R1) Attribution non-negotiable.** Furey, Dixon, Günaydin/Gürsey, Boyle/Farnsworth, Singh, Chamseddine/Connes, Koide, Baez. If you cite the QBP framework anywhere in comments, those names must appear. The existing files all have this attribution; preserve it on any edits.

**(R2) No false claims.** If a proof has a sorry, don't claim it's complete. If a theorem depends on an axiom you couldn't replace, document the dependency. Honesty about gaps is more valuable than the appearance of completeness.

**(R3) Markdown for working documents, LaTeX only for paper-ready outputs.** Your status report at the end of the session should be markdown.

**(R4) Don't redesign the architecture.** The privilege ring tower, the CayleyDickson construction, the security claims — these are decided. If you find something that seems wrong architecturally, flag it (R2) but don't fix it unilaterally.

**(R5) The honest claim about your work.** When you produce the status report at the end, claim only what you verified — "compiled with zero sorries" if you ran `lake build` and got zero sorries; "structurally verified, did not compile" if you only worked statically. Don't conflate these.

---

## 11. References (for context only — don't read unless needed)

- *Wyrd-Corpus-Index-v1.0.md* — full corpus map
- *Wyrd-Supervisor-Architecture-v0.2.md* — what Skuld is and why
- *Wyrd-Lean-Airtightening-Session-Report.md* — what the previous Claude instance did
- *Wyrd-Mathlib-API-Verification-Checklist.md* — original API verification checklist (mostly resolved)

The actual proofs you're working on are listed in §3. Those are the only files you must read.

---

## 12. Closing note

This work has been carefully designed so that the remaining gaps are mechanical refinement, not new mathematics. If you find yourself doing original mathematical work, you've gone off the path — stop and ask. The previous instance did the architectural and mathematical heavy lifting; your job is the final tactical mile.

Take it carefully. Verify each compile before moving to the next. Report honestly. The corpus has been treated with care to this point; preserve that care.

Good luck.

— Previous Claude instance, April 2026
