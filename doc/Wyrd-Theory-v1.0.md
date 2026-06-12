# Wyrd Theory v1.0

**Status:** Authoritative (June 2026)
**Lean corpus:** `lean/Wyrd/` — 25 files, zero sorries, zero user-defined axioms
**Go runtime:** `github.com/JamesPagetButler/wyrd` (module root)

> The Lean files are authoritative. This document is the prose entry point that
> points back at them. When the two diverge, the Lean file wins.

---

## Table of Contents

1. [What Wyrd is](#1-what-wyrd-is)
2. [The Cayley-Dickson hierarchy](#2-the-cayley-dickson-hierarchy)
3. [Phase 1 — Algebraic privilege](#3-phase-1--algebraic-privilege)
4. [Phase 2 — Hypergraph substrate](#4-phase-2--hypergraph-substrate)
5. [Phase 3 — Operational / constitutional](#5-phase-3--operational--constitutional)
6. [Phase 4 — Physical instantiation and holographic storage](#6-phase-4--physical-instantiation-and-holographic-storage)
7. [Substrate tier](#7-substrate-tier)
8. [Lean file index](#8-lean-file-index)

---

## 1. What Wyrd is

Wyrd is a typed hypergraph substrate with machine-checked algebraic invariants. It
has two halves:

- **Lean 4 corpus** (`lean/Wyrd/`) — 25 files of formally-verified theorems covering
  algebraic privilege, hypergraph properties, operational semantics, holographic
  storage, and substrate-tier physical invariants. All theorems carry zero `sorry`
  and zero user-defined axioms beyond mathlib4.
- **Go runtime** (`model/`, `compute/`, `store/`) — importable library that
  implements the same invariants at runtime, with every load-bearing API citing
  its Lean anchor by file path.

Downstream consumers (BMA, CTH, Contextus, QBP-CU) rely on Wyrd for:

- **Privilege enforcement** — algebraic ring-tower prevents inner-tier processes from
  synthesising outer-tier values.
- **Graph semantics** — typed hyperedge operations with provable count/incidence
  preservation.
- **Atomicity** — bridge promotion and batch operations are all-or-nothing.
- **Governance** — judge collective determinism and constitutional self-modification
  gates.
- **Physical grounding** — holographic-hypergraph irreducibility and noise bounds
  that connect the algebra to hardware.

---

## 2. The Cayley-Dickson hierarchy

The Wyrd privilege ring is the Cayley-Dickson tower:

```
  ℂ  (complex)     — Ring 3 / user tier
  ℍ  (quaternion)  — Ring 2 / supervisor tier
  𝕆  (octonion)    — Ring 1 / kernel tier
  𝕊  (sedenion)    — Ring 0 / firmware tier
```

Each step in the tower doubles the dimension and loses an algebraic property:

| Step | Gain | Loss |
|------|------|------|
| ℝ → ℂ | imaginary unit _i_ | — |
| ℂ → ℍ | units _j_, _k_ | commutativity |
| ℍ → 𝕆 | units _e₄_–_e₇_ | associativity |
| 𝕆 → 𝕊 | units _e₈_–_e₁₅_ | alternativity |

Each property loss is formalised as a non-surjection theorem (Phase 1, §3 below):
no ring homomorphism can map outer-tier values into an inner tier, because the
inner tier lacks the algebraic structure to represent them. This is the formal
foundation of Wyrd's privilege model.

**Lean file:** `lean/Wyrd/CayleyDickson.lean` — Cayley-Dickson type definition,
basis vectors (e₀–e₁₅), octonion/sedenion embedding maps, and the
`associator_octonion_witness` + `alternator_sedenion_witness_exists` lemmas that
feed Phase 1.

**Lean file:** `lean/Wyrd/Foundations.lean` — the core no-surjection theorems
(T2.1.a/b/c). **Read-only substrate; do not modify.**

---

## 3. Phase 1 — Algebraic privilege

Phase 1 establishes that the ring tower is strictly stratified: no process at one
tier can synthesise values that belong to a wider tier.

### T2.1 — No-surjection trilogy

Three theorems in `lean/Wyrd/Foundations.lean`:

| Theorem | Statement | Lean identifier |
|---------|-----------|-----------------|
| T2.1.a | No ring homomorphism ℂ → ℍ is surjective | `no_surjection_complex_to_quaternion` |
| T2.1.b | No ring homomorphism ℍ → 𝕆 is surjective | `no_surjection_quaternion_to_octonion` |
| T2.1.c | No ring homomorphism 𝕆 → 𝕊 is surjective | `no_surjection_octonion_to_sedenion` |

The proofs use concrete witnesses: `commutator_quaternion_witness` (ℂ is
commutative, ℍ is not), `octonion_assoc_witness_explicit` (ℍ is associative, 𝕆 is
not), and the sedenion alternator witness from `SedenionWitness.lean`.

The abstract structural lemmas (`no_surjection_comm_to_noncomm`,
`no_surjection_assoc_to_nonassoc`, `no_surjection_alt_to_nonalt`) generalize
these to any rings with/without the relevant property — the ring-specific proofs
instantiate these.

**Lean file:** `lean/Wyrd/Foundations.lean` (abstract lemmas + ℂ→ℍ and ℍ→𝕆 witnesses)
**Lean file:** `lean/Wyrd/SedenionWitness.lean` (𝕆→𝕊 alternator witness — explicit
`α_witness` and `β_witness` sedenion values)
**Lean file:** `lean/Wyrd/OctonionAlternative.lean` (𝕆 alternativivity — the
`octonion_alternative` theorem: (a·a)·b = a·(a·b) for all a, b ∈ 𝕆; used as
the contrapositive witness for 𝕊 loss)

### T2.2 — Projection well-definedness

`lean/Wyrd/Projection.lean`: the canonical projection π from outer ring to inner
ring (dropping the Cayley-Dickson "second half") is well-defined. The key theorem:

- `kernel_supervisor_safe` — for any a, b : ℍ, projecting the product π(a · b) in
  𝕆 back to ℍ equals the product computed purely in ℍ. Outer-ring computations on
  inner-ring values projected back equal inner-ring computations.

This is why unrestricted reads are safe in the privilege model: any outer-tier value
can be projected to an inner tier without information leakage in the upward direction.

**Lean file:** `lean/Wyrd/Projection.lean`

### T2.3 — Capability soundness

`lean/Wyrd/Capability.lean`: the practical mechanism for the privilege model.

Key theorems:
- `capability_grants_safe_access` — a holder at tier T performing operations at tier
  T' ≤ T is sound (the ring-tower non-surjection properties hold).
- `no_capability_means_no_synthesis` — without a capability at tier T, no synthesis
  of T-values is possible.
- `sandwich_mul` — sandwich conjugation preserves products: p·(u₁·u₂)·p⁻¹ =
  (p·u₁·p⁻¹)·(p·u₂·p⁻¹). The "sandwich preservation" property that secures
  capability delegation.
- `wider_capability_subsumes_narrower` — capability at tier T subsumes capability at
  any T' ≤ T.
- `hammer_capability_model` — the full privilege model theorem consolidating T2.1–T2.3.

**Lean file:** `lean/Wyrd/Capability.lean`

### Noise bound (T3.1)

`lean/Wyrd/Noise.lean`: floating-point arithmetic on ring-tower values introduces
rounding error. Phase 1 includes the noise bound:

- `abs_error_two_muls` — absolute error bound for two floating-point multiplications
  in sequence: |fl(fl(a·b)·c) − (a·b)·c| ≤ 2·ε_mach·M³ (where M bounds |a|,|b|,|c|
  and ε_mach is the machine epsilon).
- `threshold_separation_strict` — if the algebraic separation between ring tiers
  exceeds the noise floor, tier comparisons remain correct in floating-point.
- `noise_below_threshold` — for fp32 with unit-magnitude operands, the noise floor
  is safely below the tier separation threshold.

This connects the abstract ring tower to the concrete arithmetic used in the Go
runtime (`compute/quaternion.go`).

**Lean file:** `lean/Wyrd/Noise.lean`

### Hamilton product (Phase 1 companion)

`lean/Wyrd/HamiltonProduct.lean`: `hamilton_product_formula` — the explicit
formula for quaternion multiplication (re/imI/imJ/imK components), verified against
mathlib4's `Quaternion` definitions. This is the soundness anchor for
`compute.HamiltonProduct` and `compute.HamiltonProductHighPrec` in Go.

**Lean file:** `lean/Wyrd/HamiltonProduct.lean`

---

## 4. Phase 2 — Hypergraph substrate

Phase 2 establishes the graph-theoretic properties that make Wyrd's typed
hypergraph safe for concurrent use by downstream consumers.

### C-20a — Hyperedge non-interference

`lean/Wyrd/Hypergraph.lean`: adding a hyperedge e to graph G does not change
any property of vertex v that depends only on edges incident on v, as long as v is
not in e.

Key theorems:
- `hyperedge_preserves_incident_edges` (C-20a) — for any vertex v not in e,
  `incidentEdges(G ∪ {e}, v) = incidentEdges(G, v)`.
- `invariant_under_nonincident_addition` — any predicate on v's incident edges is
  invariant under adding a non-incident edge.
- `hyperedge_preserves_incoming_edges`, `hyperedge_preserves_outgoing_edges` —
  directed variants of C-20a.

BMA's local-update guarantees (engram neighbours are stable under remote edge adds)
flow from C-20a. Contextus's scout locality property does too.

**Lean file:** `lean/Wyrd/Hypergraph.lean`

### C-20b — CTH entropy monotonicity

`lean/Wyrd/CTH.lean`: the CTH entropy function η(v) = −log(1 − δ(v)) for
δ(v) ∈ [0, 1) is monotone in the measurement evidence δ.

- `cth_measurement_evidence_monotonic` — if δ₁ ≤ δ₂ then η(δ₁) ≤ η(δ₂).

This is the soundness anchor for `confluent-trust`'s `compute/entropy.go::entropyFromDelta`.

**Lean file:** `lean/Wyrd/CTH.lean`

### C-20c — Bridge promotion count preservation

`lean/Wyrd/Bridge.lean`: a signal promoted from a Contextus-side graph to a
CTH-side graph is not duplicated or lost.

Key theorems:
- `bridge_promote_preserves_count` (C-20c) — promoting signal s from Source to
  Destination leaves `signalCount(Source) + signalCount(Destination)` unchanged.
- `bridge_promote_signal_in_cth` — after promotion, s is in Destination.
- `bridge_promote_signal_not_in_contextus` — after promotion, s is not in Source.
- `bridge_promote_exactly_one_side` — s is in exactly one of Source/Destination
  at all times.

C-20c is the formal basis for BMA's sleep-cycle count conservation, Contextus→CTH
bridge atomicity, and the planned `PromoteBatch` batch primitive.

**Lean file:** `lean/Wyrd/Bridge.lean`

### Scope loader atomicity (Phase 2 extension)

`lean/Wyrd/ScopeLoader.lean`: the `LoadScopeConfig` Go primitive either applies
a full scope configuration or leaves the graph unchanged.

Key theorems:
- `scope_loader_atomic` — `atomicLoad` produces either the updated graph or the
  original, never a partial state.
- `scope_loader_count_preservation` — node count is unchanged when a load is
  rejected.
- `scope_loader_rejection_preserves_state` — a rejected load returns the original
  graph exactly.

**Lean file:** `lean/Wyrd/ScopeLoader.lean`

### Compute manifest atomicity (Phase 2 extension)

`lean/Wyrd/ComputeManifest.lean`: `LoadComputeManifest` either produces a valid
manifest or returns an error; it never silently accepts a structurally invalid manifest.

Key theorems:
- `manifest_load_atomic` — `load` produces exactly one of `Valid` or `Invalid`.
- `load_branches_disjoint` — `Valid` and `Invalid` outcomes are mutually exclusive.
- `load_deterministic` — same raw input + same `valid` flag → same outcome.
- `load_validated_iff_valid`, `load_rejected_iff_invalid` — the outcome matches
  the `valid` flag exactly.

**Lean file:** `lean/Wyrd/ComputeManifest.lean`

### Tier immunity (Phase 2 extension)

`lean/Wyrd/TierImmunity.lean`: nodes marked `TierImmune` are not affected by
eviction operations that would otherwise remove them.

Key theorems:
- `immune_not_in_effective_removal` — an immune node v is never in the effective
  removal set of any eviction.
- `eviction_immune_blind` — applying an eviction op to a graph containing only immune
  nodes leaves the graph unchanged.
- `tier_immune_preserved_under_eviction_sequence` — a node that is immune at step k
  of an eviction sequence remains in the graph at step k+1 and beyond.

**Lean file:** `lean/Wyrd/TierImmunity.lean`

### NaryMI synergy positivity (Phase 2/4 bridge, v1.5)

`lean/Wyrd/NaryMI.lean`: the CTH NaryMI synergy bonus is strictly positive for
n ≥ 3 with bounded chi-squared inputs.

Key theorems:
- `synergyTerm_pos` — the per-path synergy term is positive when n ≥ 1 and
  chiSq ≥ 0, ε > 0.
- `nary_mi_bonus_pos` (C-22) — for n ≥ 3, chiSq ≥ 0, ε > 0, the bonus is
  strictly positive.
- `nary_mi_bonus_zero_at_two` — for n = 2, the bonus is zero (pair case collapses
  to standard MI).

This is the formal justification for `confluent-trust`'s `NaryMI` synergy term and
connects to Theorem 2 irreducibility: n-ary hyperedges carry strictly more
information than all pairwise decompositions.

**Lean file:** `lean/Wyrd/NaryMI.lean`

---

## 5. Phase 3 — Operational / constitutional

Phase 3 formalises the governance and session-management properties of Wyrd as a
cognitive operating system.

### C-21a — Capability invariance under cart switch

`lean/Wyrd/Cart.lean`: a capability issued at session scope (rather than
cart-specific scope) remains valid across Systema cart switches
(Theory Cart ↔ Engineering Cart ↔ Domain-Specific Cart).

Key theorems:
- `session_scoped_valid_in_all_carts` — a session-scoped capability is valid in
  every cart.
- `capability_invariant_under_cart_switch` (C-21a) — switching carts does not
  invalidate a session-scoped capability.
- `capability_invariant_under_cart_switch_chain` — invariance holds across a
  chain of arbitrary cart switches.

**Lean file:** `lean/Wyrd/Cart.lean`

### C-21b — Transaction atomicity across cart switch

`lean/Wyrd/Transaction.lean`: any open Wyrd transaction must be resolved (committed
or aborted) before a cart switch returns. The system never observes "open transaction
crossing a cart boundary."

Key theorems:
- `resolve_observable` — after calling `resolve(tx, decision)`, the transaction state
  is observable (committed or aborted, never pending).
- `cart_switch_atomic` — `cartSwitch` is only possible from a state where no
  transaction is pending.
- `cart_switch_preserves_count`, `cart_switch_preserves_ids` — the set of resolved
  transactions is unchanged by a cart switch.

**Lean file:** `lean/Wyrd/Transaction.lean`

### C-21c — Judge collective determinism

`lean/Wyrd/JudgeCollective.lean`: vote aggregation is deterministic and
order-independent.

Key theorems:
- `judge_collective_deterministic` (C-21c) — same judges, same proposal, same
  context → same aggregate vote.
- `judge_collective_perm_invariant` — the aggregate vote is invariant under
  permutation of the judge list.
- `judge_collective_veto_propagates` — if any judge votes VETO, the collective
  result is VETO.
- `aggregate_comm`, `aggregate_assoc` — the vote aggregation operator is
  commutative and associative.

**Lean file:** `lean/Wyrd/JudgeCollective.lean`

### C-21d — Constitutional self-modification gate

`lean/Wyrd/Constitutional.lean`: BMA code updates require unanimous APPROVE from
the judge collective.

Key theorems:
- `self_modification_requires_approval` (C-21d) — `tryApplyCodeUpdate` succeeds if
  and only if the judge collective returns APPROVE (no VETO, no MAJOR_CONCERN).
- `judge_veto_blocks_self_modification` — if any judge votes VETO, the code update
  is rejected.
- `empty_judge_collective_approves` — an empty judge collective trivially returns
  APPROVE; callers are responsible for ensuring the collective is non-empty in
  production.

**Lean file:** `lean/Wyrd/Constitutional.lean`

---

## 6. Phase 4 — Physical instantiation and holographic storage

Phase 4 connects the algebraic and graph-theoretic machinery to physical storage
and hardware properties.

### Theorem 2 — 3-beam holographic irreducibility (ℝ case)

`lean/Wyrd/HolographicHypergraph.lean`: a 3-beam coherent recording is not
equivalent to three independent pairwise recordings. The irreducibility resides in
phase coherence: three beams bind phases φ₁₂, φ₂₃, φ₁₃ under the triangle
constraint φ₁₃ = φ₁₂ + φ₂₃. Three independent pair recordings have no such
constraint.

Key theorems:
- `theorem2_irreducibility` — the map `tripleToPairs` from `TripleCoherent` to
  `IndepPairs` is injective but not surjective. Its image is exactly the consistent
  pairs (those satisfying the triangle constraint).
- `tripleToPairs_consistent` — the image of `tripleToPairs` always satisfies the
  triangle constraint.
- `tripleToPairs_inj_not_surj` — the map is injective but not surjective, proven
  by explicit counterexample.

**Lean file:** `lean/Wyrd/HolographicHypergraph.lean`

### Theorem 2ℍ — Quaternion-valued irreducibility

`lean/Wyrd/HolographicHypergraphQuaternion.lean`: lifts Theorem 2 to the
quaternion case. Polarisation states compose multiplicatively (Jones calculus):
q_AC = q_AB · q_BC. The triangle constraint becomes a quaternion product
constraint. The non-surjection result holds in ℍ.

**Lean file:** `lean/Wyrd/HolographicHypergraphQuaternion.lean`

### Theorem 2 n-ary — Higher-arity generalisation

`lean/Wyrd/HolographicHypergraphHigherArity.lean`: generalises Theorem 2 from
n = 3 to all n ≥ 3.

Key theorems:
- `theorem2_irreducibility_n_arity` — for n ≥ 3, the map from n-coherent recordings
  to all-pairs representations is injective but not surjective.
- `nToAllPairs_consistent` — n-coherent recordings always map to consistent all-pairs
  configurations.

This is the formal foundation for Wyrd's claim that n-ary hyperedges (n ≥ 3) carry
irreducible information not expressible by any collection of pair edges. It is the
soundness anchor for `compute.TriangleAdditive` / `compute.TriangleMultiplicative`
and for Contextus's "3+ domain agreement" claim.

**Lean file:** `lean/Wyrd/HolographicHypergraphHigherArity.lean`

---

## 7. Substrate tier

The substrate tier is the constitutionally-frozen layer of theorems that the
federation has committed to as permanent. Substrate-tier theorems may not be edited
post-promotion; amendments require deprecate-and-replace.

**`lean/Wyrd/Substrate.lean`** is the import aggregator. Adding an `import
Wyrd.<Module>` line here is the promotion act. It is the canonical index of every
federation-pinned theorem.

**`lean/Wyrd/SubstrateTrace.lean`** provides the substrate-trace structure for
Phase-driven substrate-tier invariants. Key concepts:

- `Monotonic` — a trace is monotonic if no phase is a strict predecessor of a later
  one (monotone-non-decreasing).
- `AdvanceByOne` — a trace advances by exactly one phase per step.

These properties are the substrate for the cycle-counter cross-phase invariant.

**`lean/Wyrd/CycleCounterCrossPhase.lean`** is the first substrate-tier theorem:
the instruction-retire cycle counter advances by 1 per retired instruction,
monotonically, across all compute-manifest phases. This theorem is constitutionally
pinned per Spec 9.2 §5.

**Lean file:** `lean/Wyrd/Substrate.lean` (import aggregator — never edit promoted theorems)
**Lean file:** `lean/Wyrd/SubstrateTrace.lean`
**Lean file:** `lean/Wyrd/CycleCounterCrossPhase.lean`

---

## 8. Lean file index

All 25 Lean files in `lean/Wyrd/`:

| File | Phase | What it proves |
|------|-------|----------------|
| `Foundations.lean` | Phase 1 | T2.1 no-surjection trilogy (ℂ→ℍ, ℍ→𝕆, 𝕆→𝕊) — **read-only** |
| `CayleyDickson.lean` | Phase 1 | Cayley-Dickson type + basis vectors; associator/alternator witnesses |
| `Projection.lean` | Phase 1 | T2.2 — kernel_supervisor_safe; π projection maps |
| `Capability.lean` | Phase 1 | T2.3 — capability_grants_safe_access; sandwich_mul; privilege model |
| `Noise.lean` | Phase 1 | T3.1 — abs_error_two_muls; threshold_separation_strict; fp32 noise floor |
| `HamiltonProduct.lean` | Phase 1 | hamilton_product_formula — quaternion multiplication |
| `OctonionAlternative.lean` | Phase 1 | octonion_alternative — (a·a)·b = a·(a·b) ∀ a, b ∈ 𝕆 |
| `SedenionWitness.lean` | Phase 1 | Explicit 𝕆→𝕊 non-alternativity witness (α_witness, β_witness) |
| `Hypergraph.lean` | Phase 2 | C-20a — hyperedge_preserves_incident_edges; non-interference lemmas |
| `CTH.lean` | Phase 2 | C-20b — cth_measurement_evidence_monotonic |
| `Bridge.lean` | Phase 2 | C-20c — bridge_promote_preserves_count; exactly_one_side |
| `ScopeLoader.lean` | Phase 2 ext | scope_loader_atomic; count/state preservation under rejection |
| `ComputeManifest.lean` | Phase 2 ext | manifest_load_atomic; load_deterministic; disjointness |
| `TierImmunity.lean` | Phase 2 ext | immune_not_in_effective_removal; tier_immune_preserved_under_eviction |
| `NaryMI.lean` | Phase 2/4 | C-22 — nary_mi_bonus_pos; bonus zero at n=2 |
| `Cart.lean` | Phase 3 | C-21a — capability_invariant_under_cart_switch |
| `Transaction.lean` | Phase 3 | C-21b — cart_switch_atomic; resolve_observable |
| `JudgeCollective.lean` | Phase 3 | C-21c — judge_collective_deterministic; veto_propagates |
| `Constitutional.lean` | Phase 3 | C-21d — self_modification_requires_approval; veto_blocks |
| `HolographicHypergraph.lean` | Phase 4 | Theorem 2 — tripleToPairs injective, not surjective (ℝ) |
| `HolographicHypergraphQuaternion.lean` | Phase 4 | Theorem 2ℍ — quaternion-valued irreducibility |
| `HolographicHypergraphHigherArity.lean` | Phase 4 | Theorem 2 n-ary — irreducibility for n ≥ 3 |
| `Substrate.lean` | Substrate | Import aggregator — constitutionally-pinned theorems |
| `SubstrateTrace.lean` | Substrate | Monotonic / AdvanceByOne trace properties |
| `CycleCounterCrossPhase.lean` | Substrate | Cycle-counter cross-phase monotone invariant (first substrate-tier theorem) |

---

## Cross-references

- `lean/Wyrd/Foundations.lean` — read-only; the substrate ground truth for T2.1
- `lean/Wyrd/Substrate.lean` — the canonical substrate-tier import aggregator
- `doc/Wyrd-Spec-v1.0.md` — the implementation contract; §8 "Soundness anchors"
  maps each Go API to its Theory section
- `doc/architecture.md` — 1-page structural overview (two-halves diagram)
- `doc/design/capability-enforcement.md` — v0.2 design surface for capability
  enforcement at the Go mutation boundary (§I4 review in progress)
- `doc/design/bridge-batch.md` — v0.1 design surface for PromoteBatch
  (§I4 review in progress)
- Mathlib4 pin: commit `a090f46d` — do not update without @qbp-architecture approval
