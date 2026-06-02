# Wyrd

**Helpful Engineering — Quaternion-Based Physics Programme**
Principal Investigator: James Paget Butler

A quaternion-native typed hypergraph database whose runtime contracts are
formally verified in Lean 4.

> **Status:** Crawl (v0.1.0-alpha). Lean Phase 1–4 closed; `model` /
> `compute` / `store` shipped in main; CTH integration tracked at
> [confluent-trust v0.1.0](https://github.com/JamesPagetButler/confluent-trust/releases/tag/v0.1.0);
> QBP-CU integration interface specified in
> [wyrd-integration.md v0.2](https://github.com/JamesPagetButler/qbp-compute-unit/blob/main/doc/wyrd-integration.md).

## What it is

Wyrd is the storage and compute substrate underneath three Helpful
Engineering programmes:

- **[bma-systema][bma]** — the cognitive architecture (BMA). Wyrd hosts
  the typed hypergraph for the autonomic / subconscious / conscious layers,
  with quaternion-weighted hyperedges, Hebbian co-activation, and
  Ebbinghaus decay (Walk-phase).
- **[confluent-trust][cth]** — the epistemic-health engine (CTH). Wyrd is
  the persistence layer for trust-anchor inventories at the Walk gate.
- **Contextus** — cross-domain pattern matching. InsightSignals are
  Wyrd hyperedges; the Bridge between Contextus and CTH promotes signals
  per `Wyrd.Bridge` (Phase 2).

Wyrd has two halves:

1. A **Lean 4 proof corpus** at `lean/` that machine-checks the
   load-bearing invariants: algebraic privilege (ℂ ⊂ ℍ ⊂ 𝕆 ⊂ 𝕊), graph
   invariants under hyperedge addition, bridge-promotion atomicity,
   judge-collective determinism, holographic-hypergraph irreducibility.
2. A **Go runtime** at the repo root: typed hypergraph data model, JSON
   persistence (Crawl), and compute primitives (privilege checks,
   bridge promotion, triangle-consistency). Each Go API that has a
   formal soundness claim cites the Lean theorem in its doc comment.

Wyrd is self-sufficient: the storage layer, engram subsystem, and
hardware supervisor are all native and Lean-verified across phases.
No third-party database is required at any phase.

## Phases

| Phase | Storage | Capabilities |
|---|---|---|
| Crawl (v0.1.x) | JSON files | Core types, all algebraic-privilege checks, bridge promotion, consistency, CLI |
| Walk  (v0.2.x) | Wyrd native DB + MuninnDB engrams + NATS events | Hebbian co-activation, Ebbinghaus decay, branch-locked vaults, QBP-CU emulator-accelerated quaternion arithmetic |
| Run   (v0.3.x) | Wyrd native DB + Skuld supervisor + HAMA Tier-N | Skuld supervisor enforcing privilege at the hardware boundary; HAMA Tier-N memory |

## Quick start

```bash
git clone git@github.com:JamesPagetButler/wyrd.git
cd wyrd
go build ./...
go test -race ./...
```

To build the Lean corpus:

```bash
cd lean
lake exe cache get   # download mathlib oleans (~3 GB)
lake build           # ~12 s warm, ~3 min cold
```

## Lean corpus overview

| File | Phase | What it proves |
|---|---|---|
| `Wyrd/Foundations.lean` | 1 | Four-tier no-surjection (T2.1.a/b/c) — ℂ ↛ ℍ ↛ 𝕆 ↛ 𝕊 |
| `Wyrd/Projection.lean` | 1 | Outer→inner projection well-definedness (T2.2) |
| `Wyrd/Capability.lean` | 1 | Capability soundness + sandwich multiplicativity (T2.3, T2.4) |
| `Wyrd/Noise.lean` | 1 | fp32 associator noise floor + threshold separation (T3.1, T3.2) |
| `Wyrd/Hypergraph.lean` | 2 | Incident-edge invariance under non-incident addition (C-20a) |
| `Wyrd/CTH.lean` | 2 | Measurement-evidence entropy monotonicity (C-20b) |
| `Wyrd/Bridge.lean` | 2 | Promotion atomicity / count conservation (C-20c) |
| `Wyrd/Cart.lean` | 3 | Capability-scope persistence across cart switches (C-21a) |
| `Wyrd/Transaction.lean` | 3 | Cart-switch atomicity (C-21b) |
| `Wyrd/JudgeCollective.lean` | 3 | Judge determinism + permutation invariance (C-21c) |
| `Wyrd/Constitutional.lean` | 3 | Self-modification requires approval (C-21d) |
| `Wyrd/HolographicHypergraph.lean` | 4 | Triple-vs-pairs irreducibility, ℝ case (Theorem 2) |
| `Wyrd/HolographicHypergraphQuaternion.lean` | 4 | Theorem 2, ℍ case |
| `Wyrd/HolographicHypergraphHigherArity.lean` | 4 | Theorem 2 generalised to all n ≥ 3 |

Build status: 16 files in `Wyrd/`, 2048 lake jobs, **zero sorries, zero
user-defined axioms**. Toolchain: `lean v4.30.0-rc1`, mathlib `a090f46d`.

## Soundness pattern

Go API with a formal claim cites its Lean anchor in the doc comment:

```go
// Promote moves the hyperedge with the given ID from Source to Destination.
//
// Soundness — `Wyrd.Bridge` (Phase 2 v1.1):
//   - bridge_promote_preserves_count (C-20c): total edge count is conserved
//   - bridge_promote_exactly_one_side: signal in exactly one queue post-promotion
func (b *Bridge) Promote(id model.HyperedgeID) error { ... }
```

Diverging from the spec without updating the theorem (or vice versa)
is an audit failure.

## Theory anchors

Wyrd's substrate-level authority model and its cross-tenant interactions
are governed by theory documents that live in the federation
[`inter/theory/`](https://github.com/JamesPagetButler/inter/tree/main/theory)
directory. The canonical cross-references for Wyrd consumers:

| Theory | Location | What it governs |
|---|---|---|
| Verðandi Authority Theory v0.2 | `inter/theory/Verdandi-Authority-Theory-v0.2.md` | Ring-algebra authority model (ℂ/ℍ/𝕆/𝕊); three-gap check; delegation algebra; succession pacts |
| Verðandi Addendum A | `inter/theory/Verdandi-Authority-Theory-v0.2-addendum-A.md` | Grantable vs constructed caps (§A.4); witnessed vs declared provenance (§A.5); tamper-evident timestamps (§A.8) |
| BMA Theory v3.0 | `inter/theory/BMA-Theory-Consolidated-v3_0-DRAFT.md` | Pentagon Pod architecture (§2.1); ring-tier cognitive layer assignments (A20.0) |

The Lean substrate-tier engine for Verðandi (`lean/Wyrd/Verdandi.lean`) is
Walk-phase scope; it will import these theory documents as design anchors when
filed. The tracking issue is [wyrd#71](https://github.com/JamesPagetButler/wyrd/issues/71).

## Documentation

- [`doc/integration/cth.md`](doc/integration/cth.md) — how CTH consumes Wyrd
- [`doc/integration/bma.md`](doc/integration/bma.md) — how BMA consumes Wyrd
- [`doc/integration/contextus.md`](doc/integration/contextus.md) — how Contextus consumes Wyrd
- [`doc/archive/`](doc/archive/) — design papers, theory documents, prior proof-reference revisions
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — workflow, conventions, signed commits

## Attribution

The Cayley-Dickson construction follows Schafer; Baez 2002, *The Octonions*,
Bull. AMS. The QBP framework stands on Furey, Dixon, Günaydin / Gürsey,
Boyle / Farnsworth, Singh, Chamseddine / Connes, Koide, Baez. CTH framework:
Shannon, Dempster-Shafer, Pearl, Newman, Huet, Berge, Jirousek-Shenoy.

## Licence

To be determined. Until set, treat as "all rights reserved, Helpful Engineering."

[bma]: https://github.com/JamesPagetButler/bma-systema
[cth]: https://github.com/JamesPagetButler/confluent-trust
