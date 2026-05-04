# Changelog

All notable changes to Wyrd. Format: [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Added

- Repository scaffolding: Go module, `model/`, `compute/`, `store/`,
  `internal/validate/`, `schema/`, `cmd/`, `testdata/`, `tools/`.
- Lean 4 corpus moved into `lean/`. 16 files, 4 phases, 2048 lake jobs,
  zero sorries, zero user-defined axioms.
- `model.Tier` (ℂ/ℍ/𝕆/𝕊) with JSON round-trip.
- `model.Node`, `model.Hyperedge`, `model.Weight`, `model.Graph` with
  incidence-index maintenance.
- `compute.CanSynthesize` enforcing the four-tier ring-tower closure.
- `compute.Bridge.Promote` with destination-staging atomicity, citing
  `Wyrd.Bridge.bridge_promote_preserves_count` (Phase 2).
- `compute.TriangleAdditive` and `compute.TriangleMultiplicative` for
  Phase 4 consistency checks (ℝ and ℍ cases).
- `store.JSONFile` Crawl-phase persistence with versioned envelope.
- CI: `ci-go.yml` (build / vet / test / lint) and `ci-lean.yml`
  (lake build + zero-sorry / zero-axiom guard).
- Integration documents for CTH, BMA, and Contextus consumers.

### Changed

- `doc/archive/Wyrd-Proofs-Reference-v1.5.md` written, supersedes v1.4
  (closes #7). Captures: HolographicHypergraphQuaternion (Theorem 2 ℍ),
  HolographicHypergraphHigherArity (Theorem 2 n ≥ 3 ℝ), T3.2 promotion
  in Noise.lean, and the §23 audit-table closure of T3.2.
