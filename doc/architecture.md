# Wyrd architecture

## Two halves

```
+----------------------------------------------------+
|                    Lean 4 corpus                   |
|                    (lean/Wyrd/)                    |
|                                                    |
|  Phase 1 — algebraic privilege   ℂ ⊂ ℍ ⊂ 𝕆 ⊂ 𝕊    |
|  Phase 2 — Class B hypergraph    Bridge / CTH      |
|  Phase 3 — Class C operational   Cart / Constit.   |
|  Phase 4 — physical instantiation HolographicHG    |
|                                                    |
|  Output: theorems with zero sorries / zero axioms  |
+----------------------------------------------------+
                          |
                          | cited by doc comment
                          v
+----------------------------------------------------+
|                    Go runtime                      |
|                                                    |
|  model/   — typed hypergraph (Node, Hyperedge, …)  |
|  compute/ — privilege, bridge, consistency         |
|  store/   — JSON (Crawl), MuninnDB (Walk)…         |
|                                                    |
|  Output: linkable Go library, JSON CLIs, tests     |
+----------------------------------------------------+
                          |
                          | imported by
                          v
+----------------------------------------------------+
|                Downstream consumers                |
|                                                    |
|  bma-systema   — cognitive architecture            |
|  confluent-trust — epistemic-health metrics        |
|  Contextus     — cross-domain pattern matching     |
+----------------------------------------------------+
```

## Why two halves

A formal-only project has no operational deployment; a runtime-only
project has no machine-checked soundness. Wyrd is both because the
downstream consumers (BMA in particular) make safety claims that need
to be inspectable: "this code cannot synthesize across the privilege
boundary" must be verifiable from outside the running process.

The contract: every Go API that has a load-bearing invariant cites the
Lean theorem it relies on. Diverging from the spec without updating the
theorem (or vice versa) is an audit failure. CI enforces:

- Lean side: `lake build` succeeds with zero `sorry` and zero
  user-defined `axiom`.
- Go side: `go test -race ./...` passes; `golangci-lint run` clean.

A planned Phase 5 (deferred from current work) will write the Lean
operational-semantics contract for the QBP-CU emulator's quaternion
arithmetic, closing the loop between hardware kernels and corpus
soundness. The Stream A→B migration plan
([peer-review-005](https://github.com/JamesPagetButler/qbp-compute-unit/blob/main/architecture/peer-review-005-stream-migration.md))
makes this concrete: at M1, qbp-compute-unit gains the
`qbp.amode` / `qbp.bsel` / `qbp.psel` CSRs that turn the algebraic
privilege ring (ℂ ⊂ ℍ ⊂ 𝕆 ⊂ 𝕊) from a software convention into a
hardware-checkable mode. Phase 5's ε-tolerance theorems specify
exactly what the QW64 / QW128(DD) kernels must compute; the lean2rom
pipeline (qbp-compute-unit issue #7) makes the sign tables shared
ground truth.

## Four-corner picture

Wyrd is the substrate corner of a four-corner architecture:

```
            QBP-CU (computes; emits WDEvent)
              /      \
             /        \
          Wyrd ─── BMA ─── CTH
       (substrate) (consumer) (epistemic measure)
```

- **QBP-CU** — quaternion-algebra computer; emits a `WDEvent` per
  algebraic op. Stream A v1.1 (Xqbp.* surface form) + Stream B
  (RV-Fano Layer 0/1/2 underlying machine) co-evolve via the M0–M3
  migration plan.
- **Wyrd** (this) — typed-hypergraph substrate; Lean-certified
  invariants; importable by the other three corners.
- **BMA** — cognitive consumer; holds the live hypergraph; sleep cycle
  measures itself via CTH and gates self-modification via `Wyrd.Constitutional`.
- **CTH** — epistemic-health library (`v0.1.0` shipped 2026-05-05);
  static `ρ_net` over a programme inventory at Crawl, live `ρ_net`
  via the WDEvent → CTH loop at Walk-α.

The cross-repo integration interface for QBP-CU ↔ Wyrd is specified
in [qbp-compute-unit/doc/wyrd-integration.md](https://github.com/JamesPagetButler/qbp-compute-unit/blob/main/doc/wyrd-integration.md)
v0.2 (typed-per-width Gearbox API; tier ⊥ width orthogonality;
`Sedenion.lean` source-of-truth lives in qbp-compute-unit per option (b)).

## Phase progression

| Phase | Wyrd corpus | Wyrd Go | Downstream gate |
|---|---|---|---|
| Crawl (current, v0.1) | Phases 1–4 closed in Lean | model + compute + JSON store + `HamiltonProduct` API | CTH v0.1.0 shipped; BMA / Contextus consume Wyrd at Walk |
| Walk (v0.2) | Phase 5 ISA semantics opens (ε-tolerance theorems for QW64 / QW128 DD) | Gearbox dispatch via `qbp-compute-unit/emulator`; MuninnDB store; WDEvent → CTH loop | qbp-compute-unit M1 (AMODE/BSEL/PSEL CSRs); BMA Walk gate |
| Run (v0.3) | Phase 6: federation theorems | SurrealDB, Skuld supervisor enforcing AMODE at hardware boundary, HAMA | BMA Sprint; full federation |
| Sprint (v1.0+) | Phase 7+: information-theoretic codimensions | Real hardware | Production deploy |
