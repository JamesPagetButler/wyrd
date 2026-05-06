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
- `lean/Wyrd/NaryMI.lean` adds the CTH-domain lift `nary_mi_bonus_pos`
  (closes #3). Bonus strictly positive for `n ≥ 3` with bounded inputs.
- `compute/quaternion.go` adds the canonical `HamiltonProduct` /
  `HamiltonProductHighPrec` API (closes #11; refs #2). Tier-aware
  dispatch on the inline `math/big.Float` path; the Gearbox swap is a
  one-line follow-up pending the upstream `emulator/v0.1.0` tag.
- Integration docs refreshed for CTH `v0.1.0` (shipped 2026-05-05),
  BMA triangle/WDEvent loop architecture, and Contextus Spec v1.2
  alignment (`SignalSource` enum: scout/correlation/synthesis;
  `EvidencePointer` discipline; physical/conceptual scope nodes).
- `doc/architecture.md` updated with the four-corner picture
  (QBP-CU / Wyrd / BMA / CTH) and the Stream A→B migration plan
  reference (peer-review-005).
