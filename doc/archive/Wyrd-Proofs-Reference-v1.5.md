# Wyrd / Skuld Lean Proof Corpus — Theorem Reference

**Helpful Engineering — Quaternion-Based Physics Programme**
Principal Investigator: James Paget Butler
May 2026 — Rev 1.5 — supersedes v1.4

> **Purpose.** The canonical reference for the Lean 4 proof corpus that backs the Wyrd / Skuld algebraic privilege model, Wyrd / CTH hypergraph reasoning, BMA's operational semantics, and the PROT-HH-001 holographic hypergraph theory.

> **What changed in v1.5.**
> - **Phase 4 ℍ extension landed** (`Wyrd/HolographicHypergraphQuaternion.lean`) — Theorem 2 over quaternion polarisation states. Witness ⟨i, i, j⟩ confirmed numerically on QBP-CU emulator at W64..W512.
> - **Phase 4 higher-arity landed** (`Wyrd/HolographicHypergraphHigherArity.lean`) — Theorem 2 generalised to all n ≥ 3 (ℝ case). Image characterised exactly as the consistency subspace.
> - **T3.2 promoted from definition to theorem** in `Wyrd/Noise.lean` — `threshold_separation_bounds_noise`, `threshold_separation_strict`, `noise_below_threshold`. Closes the v1.3 audit asterisk.
> - **Repo home opened**: `github.com/JamesPagetButler/wyrd`. The Lean corpus now lives at `lean/`; a sibling Go runtime (`model/`, `compute/`, `store/`) consumes the theorems via doc-comment citations.
>
> Phase 1 + Phase 2 + Phase 3 + Phase 4 (n=3 ℝ case) content from v1.4 is unchanged and applies; v1.5 is additive.

---

## 0. Build & toolchain (verified 2026-05-03)

| Component | Pin |
|---|---|
| Lean toolchain | `leanprover/lean4:v4.30.0-rc1` |
| Mathlib | rev `a090f46da78e9af11fee348cd7ee47bf8dd219d2` |
| Project root | `~/Documents/Wyrd/lean/` (and `github.com/JamesPagetButler/wyrd/lean/`) |
| Build status | `lake build` succeeds; **0 sorries, 0 user-defined axioms** across all 16 files |
| Total declarations | **~225** (Phase 1+2+3+4 ℝ unchanged at ~200 from v1.4, plus 16 new in Phase 4 ℍ + higher-arity + T3.2 promotion) |
| File count | **16** (.lean files in `Wyrd/`) |
| CI | `.github/workflows/ci-lean.yml` runs `lake build` + greps for `sorry` / `^axiom ` |

To rebuild from cold: `cd lean && lake update && lake exe cache get && lake build`.

---

## 1. Inventory at a glance — full corpus

### Phase 1 — Algebraic privilege boundaries (unchanged from v1.3)

See v1.3 §1 for the full Phase 1 listing — the four-tier ring-tower closure (ℂ ↛ ℍ ↛ 𝕆 ↛ 𝕊), projection, capability, noise; T2.1.a/b/c, T2.4 sandwich multiplicativity.

**v1.5 additions to Phase 1**: T3.2 abstract theorem promotion (see §33).

### Phase 2 — Class B hypergraph reasoning (unchanged from v1.3)

| Theorem | File | Status |
|---|---|---|
| `Hypergraph.hyperedge_preserves_incident_edges` | Hypergraph | C-20a |
| `CTH.cth_measurement_evidence_monotonic` | CTH | C-20b |
| `Bridge.bridge_promote_preserves_count` | Bridge | C-20c |

### Phase 3 — Class C operational semantics (unchanged from v1.3)

| Theorem | File | Status |
|---|---|---|
| `Cart.capability_invariant_under_cart_switch` | Cart | C-21a |
| `Transaction.cart_switch_atomic` | Transaction | C-21b |
| `JudgeCollective.judge_collective_perm_invariant` | JudgeCollective | C-21c (substantive) |
| `Constitutional.self_modification_requires_approval` | Constitutional | C-21d |

### Phase 4 — Physical instantiation (PROT-HH-001 holographic hypergraph)

**Bold rows are the load-bearing claims of PROT-HH-001 §3.1.**

| Theorem | File | Status | Phase 4 rev |
|---|---|---|---|
| `tripleToPairs` (def) | HolographicHypergraph | embedding | v1.4 |
| `tripleToPairs_consistent` | HolographicHypergraph | image consistent | v1.4 |
| **`theorem2_irreducibility`** | HolographicHypergraph | **n=3, ℝ — not surjective** ⭐ | v1.4 |
| `tripleToPairs_image` | HolographicHypergraph | n=3, ℝ — image = consistency subspace | v1.4 |
| `tripleToPairs_injective` | HolographicHypergraph | n=3, ℝ — no info loss | v1.4 |
| `tripleToPairs_inj_not_surj` | HolographicHypergraph | n=3, ℝ — combined | v1.4 |
| `tripleToPairsH` (def) | HolographicHypergraphQuaternion | embedding | **v1.5** |
| `tripleToPairsH_consistent` | HolographicHypergraphQuaternion | image satisfies multiplicative triangle | **v1.5** |
| **`theorem2_irreducibility_quaternion`** | HolographicHypergraphQuaternion | **n=3, ℍ — not surjective** ⭐ | **v1.5** |
| `tripleToPairsH_image` | HolographicHypergraphQuaternion | n=3, ℍ — image = consistency | **v1.5** |
| `tripleToPairsH_injective` | HolographicHypergraphQuaternion | n=3, ℍ — injective | **v1.5** |
| `tripleToPairsH_inj_not_surj` | HolographicHypergraphQuaternion | n=3, ℍ — combined | **v1.5** |
| `nToAllPairs` (def) | HolographicHypergraphHigherArity | n-arity embedding | **v1.5** |
| `AllPairs.diag_zero` | HolographicHypergraphHigherArity | consistency consequence | **v1.5** |
| `AllPairs.antisym` | HolographicHypergraphHigherArity | consistency consequence | **v1.5** |
| `nToAllPairs_consistent` | HolographicHypergraphHigherArity | image lies in consistency subspace | **v1.5** |
| **`theorem2_irreducibility_n_arity`** | HolographicHypergraphHigherArity | **all n ≥ 3, ℝ — not surjective** ⭐ | **v1.5** |
| `nToAllPairs_image` | HolographicHypergraphHigherArity | n-arity image = consistency subspace | **v1.5** |

---

## 2 through 14 — Phase 1 / Phase 2 details (unchanged from v1.2)

(Sections 2–14 covering original Phase 1 / Phase 2 content remain in v1.2 / v1.3 / v1.4.)

---

## 21 through 26 — v1.3 content (unchanged)

(Sections 21–26 — Phase 1 follow-ups [T2.1.b, T2.1.c, T2.4 substantive form], the supervisor-architecture theorem-index audit, and updated downstream-code citation patterns — are unchanged from v1.3.)

---

## 27. HolographicHypergraph — Phase 4 v1.4 (unchanged)

See v1.4 §27 for the full coverage of the n=3 ℝ case (`tripleToPairs`, `theorem2_irreducibility`, `tripleToPairs_image`, `tripleToPairs_injective`, `tripleToPairs_inj_not_surj`).

---

## 31. HolographicHypergraphQuaternion — Phase 4 v1.5 ℍ extension (NEW)

**File:** `lean/Wyrd/HolographicHypergraphQuaternion.lean`. **Imports:** `Mathlib.Algebra.Quaternion`, `Mathlib.Data.Complex.Basic`, `Mathlib.Tactic.NormNum`. **Status:** clean compile, 0 sorries, 0 axioms.

### Background — the multiplicative triangle

The composition law for polarisation states is multiplicative (Jones calculus): when beam A's polarisation is rotated to B and then to C, the cumulative rotation is `q_AB · q_BC` (Hamilton product), not `q_AB + q_BC`. The triangle constraint becomes:

```
q_AC = q_AB · q_BC
```

The embedding `tripleToPairsH : (q_AB, q_BC) ↦ (q_AB, q_AB · q_BC, q_BC)` enforces this multiplicative constraint on its image.

### `theorem2_irreducibility_quaternion` ⭐

```lean
theorem theorem2_irreducibility_quaternion :
    ¬ Function.Surjective tripleToPairsH
```

**Reading:** there exist independent quaternion pair recordings that no triple-coherent recording can produce.

**Proof:** counterexample ⟨i, i, j⟩. Any preimage triple would force `i · j = i`, but in `Quaternion ℝ` we have `i · j = k` (Hamilton), and `k.imK = 1` while `i.imK = 0`. Closed via `Quaternion.imK_mul + norm_num`.

**Numerical witness verification:** `BMA/projects/holographic-hypergraph/quaternion-witness/main.go` uses qbp-emulator's `Gearbox.Mul` at four precision tiers (W64, W128, W256, W512) and reports the residual norm `‖q_AC − q_AB·q_BC‖² = 2.0` exactly across all tiers (60+ digits identical). This exactness is the structural-vs-noise discriminator: an integer-component residual is exact in big.Float arithmetic at any precision. Validates the witness as algebraically structural before the Lean proof was committed.

### Other v1.5 theorems on this file

- `tripleToPairsH_image` — image is exactly the multiplicative-consistency subspace.
- `tripleToPairsH_injective` — different triples → different pair representations.
- `tripleToPairsH_inj_not_surj` — combined statement.

### Architectural meaning

This theorem closes the algebra-as-schema thesis (PROT-HH-002 §3 consequence 2, Level 1 substrate) for the quaternion / SU(2) sub-case: the joint constraint encoded by `q_AC = q_AB · q_BC` is irreducible to pair-only data even when polarisation is the encoding. Sets the pattern for Sprint-phase extensions to octonion polarisation cases.

---

## 32. HolographicHypergraphHigherArity — Phase 4 v1.5 generalised (NEW)

**File:** `lean/Wyrd/HolographicHypergraphHigherArity.lean`. **Imports:** `Mathlib.Analysis.SpecialFunctions.Trigonometric.Basic`, `Mathlib.Tactic.{Linarith, NormNum}`. **Status:** clean compile, 0 sorries, 0 axioms.

### Background

Generalises the n=3 ℝ case to all n ≥ 3:

```lean
def nToAllPairs {n : ℕ} (nc : NCoherent n) : AllPairs n :=
  { relPhase := fun i j => nc.absPhase i - nc.absPhase j }

def AllPairs.IsConsistent {n : ℕ} (ap : AllPairs n) : Prop :=
  ∀ i j k : Fin n, ap.relPhase i k = ap.relPhase i j + ap.relPhase j k
```

### `theorem2_irreducibility_n_arity` ⭐

```lean
theorem theorem2_irreducibility_n_arity {n : ℕ} (h_n : 3 ≤ n) :
    ¬ Function.Surjective (@nToAllPairs n)
```

**Reading:** for any n ≥ 3, the embedding from coherent recordings to all-pairs recordings is not surjective.

**Proof:** witness `relPhase 0 2 = π, others = 0`. Triangle on (0, 1, 2) forces `π = 0 + 0 = 0`, contradiction via `Real.pi_pos`.

**Strength of the result:** "all-pairs" is a strictly finer decomposition than "(n-1)-beam leave-one-out." So this theorem implies the originally-deferred phrasing ("n-beam ↛ (n-1)-beam"): if you can't reconstruct from all pairs, you can't reconstruct from any decomposition that factors through pairs.

### Other v1.5 theorems on this file

- `AllPairs.diag_zero` — consistency forces `relPhase i i = 0`.
- `AllPairs.antisym` — consistency forces `relPhase i j = -relPhase j i`.
- `nToAllPairs_consistent` — image lies in the consistency subspace.
- `nToAllPairs_image` — image is **exactly** the consistency subspace; the reverse direction reconstructs `absPhase i := relPhase i 0` from a consistent all-pairs recording.

### Not in v1.5 (deliberate)

- **Injectivity**: `nToAllPairs` is not injective (NCoherent values differing by a global phase shift produce identical AllPairs outputs). The quotient form `nToAllPairs_injective_mod_shift` is deferred to v1.6.
- **Quaternion higher-arity** (`HolographicHypergraphQuaternion` × n ≥ 3): deferred to v1.6.
- **(n-1)-beam coherent leave-one-out variant**: harder to model (requires cross-recording phase-reference independence); deferred to v1.6+.

---

## 33. Noise — T3.2 abstract theorem promotion (NEW v1.5)

**File:** `lean/Wyrd/Noise.lean` (extended at end). **Status:** clean compile, 0 sorries, 0 axioms.

The v1.3 audit (§23) flagged T3.2 as "definition + numerical bounds present, no abstract theorem" and recommended a v1.4/v1.5 promotion. The promotion lands here.

### Helpers

- `associator_noise_bound_nonneg` — non-negative for non-negative magnitude.
- `associator_noise_bound_pos` — strictly positive for strictly positive magnitude.

### `threshold_separation_bounds_noise`

```lean
theorem threshold_separation_bounds_noise
    (ε_priv : ℝ) (R : RoundingModel) (M : ℝ) (k : ℝ)
    (h_k : 1 ≤ k) (h_M : 0 ≤ M)
    (h_sep : threshold_separation_safe ε_priv R M k) :
    associator_noise_bound R M ≤ ε_priv
```

Under threshold separation with safety factor `k ≥ 1`, the associator noise bound is bounded above by `ε_priv`.

### `threshold_separation_strict`

```lean
theorem threshold_separation_strict
    (ε_priv : ℝ) (R : RoundingModel) (M : ℝ) (k : ℝ)
    (h_k : 1 < k) (h_M : 0 < M)
    (h_sep : threshold_separation_safe ε_priv R M k) :
    associator_noise_bound R M < ε_priv
```

Strict version: with `k > 1` and positive magnitude, the noise bound is *strictly* below `ε_priv`.

### `noise_below_threshold` (operational form)

```lean
theorem noise_below_threshold
    (ε_priv : ℝ) (R : RoundingModel) (M : ℝ) (k : ℝ) (noise : ℝ)
    (h_k : 1 < k) (h_M : 0 < M)
    (h_sep : threshold_separation_safe ε_priv R M k)
    (h_noise : |noise| ≤ associator_noise_bound R M) :
    |noise| < ε_priv
```

Any actual noise sample within the associator bound is strictly below `ε_priv`. **This is the citation Skuld and downstream callers use** when claiming "no noise event can reach the privilege threshold from the noise side."

### Audit table update

The v1.3 §23 audit table now shows T3.2 as ✅ **Closed** (was ⚠️). The remaining forward-looking entries (T3.3 / T4.x / T5.x) are unchanged.

---

## 34. Downstream code citation pattern — v1.5 additions

### From `github.com/JamesPagetButler/wyrd` (Go runtime)

```go
// compute.CanSynthesize enforces the four-tier ring-tower closure.
//
// Soundness — Wyrd.Foundations:
//   no_surjection_complex_to_quaternion (T2.1.a):  ℂ ↛ ℍ
//   no_surjection_quaternion_to_octonion (T2.1.b): ℍ ↛ 𝕆
//   no_surjection_octonion_to_sedenion (T2.1.c):   𝕆 ↛ 𝕊
func CanSynthesize(caller, target model.Tier) error { ... }
```

```go
// compute.TriangleMultiplicative checks q_ik = q_ij · q_jk consistency
// (Hamilton product). Image of the embedding lies in this subspace per
// Wyrd.HolographicHypergraphQuaternion.tripleToPairsH_consistent (Phase 4 v1.5).
func TriangleMultiplicative(qIJ, qJK, qIK model.Weight) float64 { ... }
```

```go
// compute.Bridge.Promote moves a hyperedge from Source to Destination atomically.
//
// Soundness — Wyrd.Bridge (Phase 2 v1.1):
//   - bridge_promote_preserves_count (C-20c)
//   - bridge_promote_exactly_one_side
func (b *Bridge) Promote(id model.HyperedgeID) error { ... }
```

### From `github.com/JamesPagetButler/confluent-trust` (CTH)

After the issue #35 lift lands:

```go
// compute.NaryMI's synergy bonus is the operational form of Wyrd's
// holographic-hypergraph irreducibility for confluences of arity ≥ 3.
// The bonus being non-zero for N ≥ 3 with bounded chi-squared is
// formalised by Wyrd.HolographicHypergraphHigherArity.theorem2_irreducibility_n_arity
// (Phase 4 v1.5) and its CTH lift `nary_mi_synergy_strict` (Wyrd v1.6 candidate).
func NaryMI(predictions, sigmas []float64) float64 { ... }
```

---

## 35. Versioning chain (updated)

- **v1.0**: Phase 1 baseline (algebraic privilege, T2.1.a only)
- **v1.1**: Phase 2 added (Class B: Hypergraph, CTH, Bridge)
- **v1.2**: Phase 3 added (Class C: Cart, Transaction, JudgeCollective, Constitutional)
- **v1.3**: Phase 1 follow-ups (T2.1.b, T2.1.c explicit; T2.4 substantive form; theorem-index audit)
- **v1.4**: Phase 4 opened (HolographicHypergraph: Theorem 2 + image + injectivity, ℝ n=3 case)
- **v1.5** (current): Phase 4 ℍ extension, Phase 4 higher-arity ℝ, T3.2 promotion; Wyrd-the-repo opens at `github.com/JamesPagetButler/wyrd`
- **v1.6** (anticipated): NaryMI Lean lift (CTH); higher-arity ℍ; quotient-by-shift injectivity for higher-arity; (n-1)-beam coherent decomposition variant
- **v2.0** (long-term): Phase 5 ISA semantics — operational-semantics contracts for QBP-CU emulator kernels

All older versions retained in `doc/archive/`; **v1.5 is the canonical reference as of 2026-05-03**.

---

## 36. Theorem inventory — final tally

**16 files, ~225 declarations, 0 sorries, 0 user-defined axioms.**

| Phase | Theorems / lemmas / structures | Architectural role |
|---|---|---|
| Phase 1 (algebraic privilege) | ~35 (with T3.2 promotion) | Four-tier ring tower closure + projection + capability + noise + threshold separation |
| Phase 2 (Class B hypergraph) | ~20 | Graph invariants + CTH metrics + Bridge atomicity |
| Phase 3 (Class C operational) | ~25 | Cart capabilities + transaction atomicity + judge determinism + constitutional pin |
| Phase 4 (physical instantiation) | ~25 | HH irreducibility: ℝ n=3, ℍ n=3, ℝ n≥3 |
| Helpers / corollaries / instances | ~120 | Supporting infrastructure |

The corpus now provides:

- **End-to-end formal foundations** for all three workload classes (Phase 1+2+3),
- **Constitutional safety properties** for autonomous BMA operation (Phase 3),
- **Physical-instantiation soundness** for the holographic hypergraph storage layer over both ℝ and ℍ, all arities n ≥ 3 (Phase 4),
- **Threshold-separation bounds** as theorems, not just predicates (T3.2 promotion).

---

## Attribution

QBP framework: Furey, Dixon, Günaydin / Gürsey, Boyle / Farnsworth, Singh, Chamseddine / Connes, Koide, Baez. CTH framework: Shannon, Dempster-Shafer, Pearl, Newman, Huet, Berge, Jirousek-Shenoy. Cynefin domain framing: Snowden. The Cayley-Dickson construction follows Schafer; Baez 2002, *The Octonions*, Bull. AMS. The judge-collective design with VETO-absorbing aggregation follows the BMA Governance Document; the constitutional-pin pattern is consistent with Ethics v1.1. The VSG scrambling speed limit cited in PROT-HH-001 is Vikram-Shou-Galitski (PRL 136, 150401, 2026); the Bekenstein bound is the canonical 1981 result. The holographic hypergraph theory was developed by J. Butler & Claude (Opus 4.6 Red Team / Architecture instance), 30 April 2026 (PROT-HH-001) and 3 May 2026 (Phase 4 Lean substrate, by J. Butler & Claude Opus 4.7).

---

*End of Wyrd Proofs Reference v1.5.*
