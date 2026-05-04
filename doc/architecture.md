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
soundness.

## Phase progression

| Phase | Wyrd corpus | Wyrd Go | Downstream gate |
|---|---|---|---|
| Crawl (current, v0.1) | Phases 1–4 closed in Lean | model + compute + JSON store | CTH/BMA/Contextus *can* import; not yet integrated |
| Walk (v0.2) | Phase 5 ISA semantics opens | MuninnDB store, NATS events, QBP-CU dispatch | BMA Walk gate; CTH Walk-phase issues |
| Run (v0.3) | Phase 6: federation theorems | SurrealDB, Skuld supervisor, HAMA | BMA Sprint; full federation |
| Sprint (v1.0+) | Phase 7+: information-theoretic codimensions | Real hardware | Production deploy |
