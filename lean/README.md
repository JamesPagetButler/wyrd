# Wyrd Lean Verification Project

**Helpful Engineering — Quaternion-Based Physics Programme**
Principal Investigator: James Paget Butler
April 2026

## What this is

The formal verification corpus for the Wyrd / Skuld algebraic privilege model. Wyrd is a quaternion-native hypergraph database; Skuld is its supervisor. The privilege boundaries between user / supervisor / kernel / firmware rings correspond to subalgebras in the Cayley-Dickson tower (ℂ ⊂ ℍ ⊂ 𝕆 ⊂ 𝕊). These proofs establish that:

1. **Boundary detectors are sound** (T1.2): the commutator vanishes in commutative rings, the associator vanishes in associative rings, the alternator vanishes in alternative rings.
2. **Generators cannot be synthesized across boundaries** (T2.1): no ring homomorphism from an inner ring to an outer ring is surjective.
3. **Projections are well-defined** (T2.2): outer-ring computations on inner-ring values, projected back, equal inner-ring computations.
4. **Capabilities are sound** (T2.3): a process holding a capability for ring R' can perform R'-ring operations safely; without one, no synthesis.
5. **Noise floors are bounded** (T3.1): fp32 representation of the associator gives a noise floor distinguishable from privilege thresholds.

These five facts together formalize the security claim of the Wyrd / Skuld architecture.

## Structure

```
.
├── README.md                  — this file
├── lakefile.lean              — Lake build configuration
├── lean-toolchain             — Lean version
├── Wyrd.lean                  — top-level module (imports everything)
├── Wyrd/
│   ├── CayleyDickson.lean         — Octonion and Sedenion types via Cayley-Dickson
│   ├── Foundations.lean           — T1.2 (boundary detectors) + T2.1.a (ℂ→ℍ)
│   ├── Projection.lean            — T2.2 (projection well-definedness)
│   ├── Capability.lean            — T2.3 (capability soundness)
│   ├── Noise.lean                 — T3.1 (associator noise bound, fp32)
│   ├── SedenionWitness.lean       — Concrete sedenion alternator witness
│   └── OctonionAlternative.lean   — Octonions are alternative
└── archive/                       — historical versions for audit trail
    ├── Wyrd-Algebraic-Privilege-Proofs-v0.1.lean
    ├── Wyrd-Algebraic-Privilege-Proofs-v0.2.lean
    ├── Wyrd-Algebraic-Privilege-Proofs-v0.3.lean
    └── Wyrd-T3.1-Noise-Bound-v0.1.lean
```

## Build

```bash
lake exe cache get      # download precompiled mathlib oleans
lake build              # build the Wyrd library
```

If `lake exe cache get` fails, mathlib will build from source — budget 30–60 minutes.

## Status as of April 2026

| File | Status | Sorries | Axioms |
|---|---|---|---|
| `CayleyDickson.lean` | Needs API update for mathlib4 3-param `QuaternionAlgebra` | 0 (octonion witness deferred to comments) | 0 |
| `Foundations.lean` | Source-verified against mathlib4 master | 0 | 0 |
| `Projection.lean` | Should compile cleanly | 0 | 0 |
| `Capability.lean` | Should compile cleanly | 0 | 0 |
| `Noise.lean` | Zero sorries | 0 | 0 |
| `SedenionWitness.lean` | One mechanical sorry | 1 | 0 |
| `OctonionAlternative.lean` | Two `ring_nf` sorries, two axioms standing in for mathlib lemmas | 2 | 2 |

**Total to close:** 3 sorries, 2 axioms. Estimated work in a live Lean environment: 1–2 hours. Detailed instructions in the `Wyrd-Lean-Onboarding-Prompt.md` document (in the parent corpus).

## Theorem inventory

### Tier 1 — foundational (in Foundations.lean)

- `no_surjection_comm_to_noncomm` — abstract commutativity-preservation theorem
- `no_surjection_assoc_to_nonassoc` — abstract associativity-preservation theorem
- `no_surjection_alt_to_nonalt` — abstract alternativity-preservation theorem
- `commutator_eq_zero_of_comm` (T1.2.a vanishing)
- `associator_eq_zero_of_assoc` (T1.2.b vanishing)
- `alternator_eq_zero_of_alt` (T1.2.c vanishing)
- `commutator_quaternion_witness` (T1.2.a witness — i*j vs j*i in ℍ)
- `no_surjection_complex_to_quaternion` (T2.1.a — user→supervisor boundary closed)

### Tier 2 — privilege model

- `Projection.kernel_supervisor_safe` (T2.2 main payload)
- `Capability.capability_grants_safe_access` (T2.3 positive)
- `Capability.no_capability_means_no_synthesis` (T2.3 negative)
- `Capability.hammer_capability_model` (worked example: Hammer simulation)
- `Capability.wider_capability_subsumes_narrower`
- `Capability.sandwich_preservation_associative`

### Tier 3 — precision (in Noise.lean)

- `RoundingModel` (abstract IEEE-style floating-point model)
- `abs_error_one_mul`, `abs_error_two_muls` (parametric error bounds)
- `fp32_noise_unit_magnitude` (~3e-6 noise floor)
- `fp32_noise_decimal_magnitude` (~3e-3 noise floor at M=10)
- `threshold_separation_safe` (T3.2 statement)

### Witness theorems

- `associator_octonion_witness` (in CayleyDickson.lean, witnesses T1.2.b)
- `alternator_sedenion_witness` (in SedenionWitness.lean, witnesses T1.2.c)
- `octonion_alternative` (in OctonionAlternative.lean, used in T2.1.c)

## Attribution

Per QBP standing rule: this work stands on the shoulders of Furey, Dixon, Günaydin/Gürsey, Boyle/Farnsworth, Singh, Chamseddine/Connes, Koide, and Baez. The Cayley-Dickson construction follows standard references (Schafer; Baez 2002, *The Octonions*).

## Documentation

The full corpus index is in `Wyrd-Corpus-Index-v1.0.md` (parent project). Specific design rationale:

- **Why these proofs?** See `Wyrd-Supervisor-Architecture-v0.2.md`.
- **What if Branch A wins?** See `Wyrd-BranchA-Contingency-v0.1.md`.
- **How to drive sorries to zero?** See `Wyrd-Lean-Onboarding-Prompt.md`.

## License

To be determined. Until set, treat as "all rights reserved, Helpful Engineering."
