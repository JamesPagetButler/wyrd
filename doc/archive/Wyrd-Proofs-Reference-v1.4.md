# Wyrd / Skuld Lean Proof Corpus — Theorem Reference

**Helpful Engineering — Quaternion-Based Physics Programme**
Principal Investigator: James Paget Butler
May 2026 — Rev 1.4 — supersedes v1.3

> **Purpose.** The canonical reference for the Lean 4 proof corpus that backs the Wyrd / Skuld algebraic privilege model, Wyrd / CTH hypergraph reasoning, BMA's operational semantics, and the PROT-HH-001 holographic hypergraph theory.

> **What changed in v1.4.**
> - **Phase 4 opened** (`Wyrd/HolographicHypergraph.lean`) — physical-instantiation theorems for the holographic hypergraph storage architecture (PROT-HH-001).
> - **Theorem 2 proven** (`theorem2_irreducibility`) — `tripleToPairs` is not surjective: a 3-beam coherent recording is information-distinct from three independent pair recordings (the triangle constraint cannot be encoded by independent pairs).
> - Image of the embedding characterised exactly (`tripleToPairs_image`): the image is the consistency subspace `phase13 = phase12 + phase23`.
> - Embedding shown to be injective (`tripleToPairs_injective`): no information is lost going from triple to consistent-pair representation; only constraints are added on the codomain.
> - Combined inj-but-not-surj statement (`tripleToPairs_inj_not_surj`).
> - Numerical companion `sim_theorem2.py` shipped at `BMA/projects/holographic-hypergraph/`.
>
> Phase 1 + Phase 2 + Phase 3 content from v1.3 is unchanged and applies; v1.4 is additive — opens the physical-instantiation phase of the corpus and lands the load-bearing claim of PROT-HH-001 §3.1 as machine-checked Lean.

---

## 0. Build & toolchain (verified 2026-05-03)

| Component | Pin |
|---|---|
| Lean toolchain | `leanprover/lean4:v4.30.0-rc1` |
| Mathlib | rev `a090f46da78e9af11fee348cd7ee47bf8dd219d2` |
| Project root | `~/Documents/Wyrd/wyrd-lean-project/` |
| Build status | `lake build` succeeds; **0 sorries, 0 user-defined axioms** across all 15 files |
| Total declarations | **~200** (Phase 1+2+3 unchanged at 190 from v1.3, plus 10 new in Phase 4) |
| File count | **15** (.lean files in `Wyrd/`) |

To rebuild from cold: `cd ~/Documents/Wyrd/wyrd-lean-project && lake update && lake exe cache get && lake build`.

---

## 1. Inventory at a glance — full corpus

### Phase 1 — Algebraic privilege boundaries (unchanged from v1.3)

See v1.3 §1 for the full Phase 1 listing (the four-tier ring-tower closure ℂ ↛ ℍ ↛ 𝕆 ↛ 𝕊; projection, capability, noise; T2.1.a/b/c, T2.4 sandwich multiplicativity).

### Phase 2 — Class B hypergraph reasoning (unchanged from v1.3)

| Theorem | File | Status |
|---|---|---|
| `Hypergraph.hyperedge_preserves_incident_edges` | Hypergraph | C-20a |
| `CTH.cth_measurement_evidence_monotonic` | CTH | C-20b |
| `Bridge.bridge_promote_preserves_count` | Bridge | C-20c |
| (and supporting lemmas) | | |

### Phase 3 — Class C operational semantics (unchanged from v1.3)

| Theorem | File | Status |
|---|---|---|
| `Cart.capability_invariant_under_cart_switch` | Cart | C-21a |
| `Transaction.cart_switch_atomic` | Transaction | C-21b |
| `JudgeCollective.judge_collective_perm_invariant` | JudgeCollective | C-21c (substantive) |
| `Constitutional.self_modification_requires_approval` | Constitutional | C-21d |
| (and supporting lemmas) | | |

### Phase 4 — Physical instantiation (PROT-HH-001 holographic hypergraph) — *NEW v1.4*

**Bold rows are the load-bearing claims of PROT-HH-001 §3.1.**

| Theorem | File | Status |
|---|---|---|
| `TripleCoherent` (struct) | HolographicHypergraph | 2-DOF triple recording |
| `IndepPairs` (struct) | HolographicHypergraph | 3-DOF pair recordings |
| `tripleToPairs` (def) | HolographicHypergraph | the natural embedding |
| `tripleToPairs_phase12 / 13 / 23` | HolographicHypergraph | simp lemmas for projections |
| `IndepPairs.IsConsistent` (def) | HolographicHypergraph | triangle predicate |
| `tripleToPairs_consistent` | HolographicHypergraph | image always consistent |
| **`theorem2_irreducibility`** | HolographicHypergraph | **Theorem 2 — `tripleToPairs` not surjective** ⭐ |
| `tripleToPairs_image` | HolographicHypergraph | image = consistency subspace (iff) |
| `tripleToPairs_injective` | HolographicHypergraph | no information lost going triple → pairs |
| `tripleToPairs_inj_not_surj` | HolographicHypergraph | combined: injective ∧ ¬ surjective |

---

## 2 through 14 — Phase 1 / Phase 2 details (unchanged from v1.2)

(Sections 2–14 covering Foundations original content, CayleyDickson, Projection, Capability, Noise, SedenionWitness, OctonionAlternative, Hypergraph, CTH, Bridge are unchanged from v1.2 and remain accessible there.)

---

## 21 through 26 — v1.3 content (unchanged)

(Sections 21–26 covering Phase 1 follow-ups [T2.1.b, T2.1.c, T2.4 substantive form], the supervisor-architecture theorem-index audit, and updated downstream-code citation patterns are unchanged from v1.3.)

---

## 27. HolographicHypergraph — Phase 4 physical-instantiation theorems (*NEW v1.4*)

**File:** `Wyrd/HolographicHypergraph.lean`. **Imports:** `Mathlib.Analysis.SpecialFunctions.Trigonometric.Basic` (for `Real.pi_ne_zero`), `Mathlib.Tactic.Linarith`, `Mathlib.Tactic.NormNum`. **Status:** clean compile, 0 sorries, 0 axioms. **Companion:** `~/Documents/BMA/projects/holographic-hypergraph/sim_theorem2.py`.

### Background — the physical claim

PROT-HH-001 §3.1 asserts:

> A 3-beam coherent recording is NOT equivalent to three independent pair recordings, even in a fully linear-response medium.

The architectural significance: the holographic hypergraph storage layer is irreducibly multi-party. Its hyperedges encode joint relations among k ≥ 3 vertices that no decomposition into k(k-1)/2 pair recordings can capture — even in principle. This is the formal core of "the medium *is* the schema" — quaternion-weighted volumetric interference physically enforces a category-theoretic constraint that software-mediated databases must enforce at runtime.

### The encodings

```lean
structure TripleCoherent where
  phase12 : ℝ
  phase23 : ℝ

structure IndepPairs where
  phase12 : ℝ
  phase13 : ℝ
  phase23 : ℝ
```

A triple-coherent recording carries 2 DOF (the third relative phase is determined by the triangle constraint `phase13 = phase12 + phase23`). Three independent pair recordings carry 3 DOF (each pair has its own absolute phase reference; the recovered relative phases are unconstrained).

### The embedding

```lean
def tripleToPairs (tc : TripleCoherent) : IndepPairs :=
  { phase12 := tc.phase12
    phase13 := tc.phase12 + tc.phase23
    phase23 := tc.phase23 }

def IndepPairs.IsConsistent (ip : IndepPairs) : Prop :=
  ip.phase13 = ip.phase12 + ip.phase23
```

By construction the embedding lands in the consistency subspace.

### `theorem2_irreducibility` ⭐

```lean
theorem theorem2_irreducibility :
    ¬ Function.Surjective tripleToPairs
```

**Reading:** there exist independent-pair configurations that no triple-coherent recording can produce.

**Proof:** counterexample `⟨0, 0, π⟩`. If a triple `tc` mapped to it, the embedding would force `tc.phase12 + tc.phase23 = 0`, but the witness has `phase23 = π` and `phase12 = 0`, so the sum is `π ≠ 0` (closed by `Real.pi_ne_zero`).

**Architectural meaning:** the joint relation expressed by the triangle constraint is *not in the image* of any pair-decomposition; therefore pairwise hyperedges cannot encode the same information as a 3-edge. Sets the pattern for higher arity (Phase 4+ deferred work — to be proven for n ≥ 3 by analogous counterexample on the `(n-1)` simplex of constraints).

### `tripleToPairs_image`

```lean
theorem tripleToPairs_image (ip : IndepPairs) :
    (∃ tc : TripleCoherent, tripleToPairs tc = ip) ↔ ip.IsConsistent
```

**Reading:** the image is *exactly* the consistency subspace — neither smaller nor larger.

**Architectural meaning:** sharpens Theorem 2. A pair configuration is reachable from some triple iff it satisfies the triangle constraint. The "missing" configurations are *exactly* those that violate consistency. There are no other obstructions.

### `tripleToPairs_injective`

```lean
theorem tripleToPairs_injective : Function.Injective tripleToPairs
```

**Reading:** different triples → different pair representations.

**Architectural meaning:** the embedding loses no information. Going from triple to consistent-pair adds a derivable coordinate (`phase12 + phase23`) without folding any state. The two encodings differ only in their *constraints*, not their *resolvable detail*.

### `tripleToPairs_inj_not_surj`

```lean
theorem tripleToPairs_inj_not_surj :
    Function.Injective tripleToPairs ∧ ¬ Function.Surjective tripleToPairs
```

The combined statement: **information-distinct encodings**. Triples are an embedded but proper subset of pair recordings.

### Numerical companion (`sim_theorem2.py`)

The Lean theorem is the *load-bearing claim*; the simulation is *Level 1 evidence*. The script sweeps a phase-drift parameter (each pair recording's absolute phase reference is drawn from N(0, drift_std)) and reports the mean triangle residual `|phase13 − (phase12 + phase23)|` and the fidelity to the coherent-triple ground truth. At drift = 0, the residual is 0 and fidelity = 1; the residual grows with drift, fidelity decays. This is the bench-experimentalist-facing version of the theorem.

```
drift_std (rad)     mean |residual|     mean fidelity
       0.0000              0.0000            1.0000   ← coherent triple
       0.5000              0.9986            0.5481
       1.0000              2.0365            0.1768   ← independent pairs
```

Path: `~/Documents/BMA/projects/holographic-hypergraph/sim_theorem2.py`. Output: `theorem2_sweep.png`.

### Deferred Phase 4 work

| Item | Description | Priority |
|---|---|---|
| Quaternion extension | Replace ℝ-valued phases with ℍ-valued polarisation states; show analogous irreducibility for the SU(2) sub-case (PRED-HH-09 dependency) | Required for PROT-HH-002 Level 2 |
| Higher arity | Prove n-beam recordings ↛ (n-1)-beam recordings for n ≥ 3 by induction on simplex codimension | Required for the general "k-edge irreducibility" claim |
| Information-theoretic codimension | Express the gap between triple-image and full-pairs as Shannon entropy / mutual information | Forward-looking; Sprint phase |
| HAMA hardware bench protocol | Tie residual decay curve to femtosecond-laser fused silica recording; PRED-HH-01 / PRED-HH-07 verification | Walk phase, on hardware |

Phase 4 lands the *theory-substrate* for these — the deferred items extend the proof; they do not invalidate it.

---

## 28. Downstream code citation pattern — Phase 4 additions

```go
// muninndb.AddHyperedge inserts a k-ary hyperedge atomically.
// Soundness — a 3+ vertex hyperedge is irreducible to its pair-projections
// by HolographicHypergraph.theorem2_irreducibility (Wyrd-Proofs-Reference-v1.4 §27).
// The store must NOT silently decompose a k-edge into k(k-1)/2 pair edges:
// the joint constraint is information that pair edges cannot encode.
func (db *MuninnDB) AddHyperedge(verts []NodeID, weight Quaternion) error { ... }
```

```go
// hama.RecordTriple writes a 3-beam coherent recording. The triangle
// constraint phase13 = phase12 + phase23 is enforced by the medium
// (no software gating required) and matches HolographicHypergraph.tripleToPairs_consistent.
// Soundness of the read path — coherent recordings always lie in the
// consistency subspace by HolographicHypergraph.tripleToPairs_image.
func (h *HAMA) RecordTriple(beams [3]Beam, exposure_J float64) error { ... }
```

```go
// contextus.PromoteHyperedge forbids "split-and-restore" passes that
// would round-trip a hyperedge through pair edges. Per
// HolographicHypergraph.tripleToPairs_inj_not_surj (Phase 4):
//   triple ↪ pair (lossless)  but  pair ↛ triple  (information missing).
// A pass that pair-decomposes is one-way; we must keep the original triple
// or accept information loss. Honest-rounding policies live here.
func (c *Contextus) PromoteHyperedge(e *Hyperedge) (*Hyperedge, error) { ... }
```

---

## 29. Versioning chain (updated)

- **v1.0**: Phase 1 baseline (algebraic privilege, T2.1.a only)
- **v1.1**: Phase 2 added (Class B: Hypergraph, CTH, Bridge)
- **v1.2**: Phase 3 added (Class C: Cart, Transaction, JudgeCollective, Constitutional)
- **v1.3**: Phase 1 follow-ups (T2.1.b, T2.1.c explicit; T2.4 substantive form; theorem-index audit)
- **v1.4** (current): **Phase 4 opened** (HolographicHypergraph: Theorem 2 + image characterisation + injectivity + numerical companion)
- **v1.5** (anticipated): T3.2 abstract theorem promotion (deferred from v1.3); higher-arity HH irreducibility (n ≥ 3); quaternion extension of Theorem 2
- **v2.0** (long-term): structural reorganization if corpus grows beyond ~25 files

All older versions retained in the Archive for audit; **v1.4 is the canonical reference as of 2026-05-03**.

---

## 30. Theorem inventory — final tally

**15 files, ~200 declarations, 0 sorries, 0 user-defined axioms.**

| Phase | Theorems / lemmas / structures | Architectural role |
|---|---|---|
| Phase 1 (algebraic privilege) | ~30 | Four-tier ring tower closure + projection + capability + noise |
| Phase 2 (Class B hypergraph) | ~20 | Graph invariants + CTH metrics + Bridge atomicity |
| Phase 3 (Class C operational) | ~25 | Cart capabilities + transaction atomicity + judge determinism + constitutional pin |
| Phase 4 (physical instantiation) | ~10 | Holographic hypergraph multi-beam irreducibility (PROT-HH-001) |
| Helpers / corollaries / instances | ~115 | Supporting infrastructure |

The corpus now provides:
- **End-to-end formal foundations** for all three workload classes (Phase 1+2+3),
- **Constitutional safety properties** for autonomous BMA operation (Phase 3),
- **Physical-instantiation soundness** for the holographic hypergraph storage layer (Phase 4).

---

## Attribution

QBP framework: Furey, Dixon, Günaydin / Gürsey, Boyle / Farnsworth, Singh, Chamseddine / Connes, Koide, Baez. CTH framework: Shannon, Dempster-Shafer, Pearl, Newman, Huet, Berge, Jirousek-Shenoy. Cynefin domain framing: Snowden. The Cayley-Dickson construction follows Schafer; Baez 2002, *The Octonions*, Bull. AMS. The judge-collective design with VETO-absorbing aggregation follows the BMA Governance Document; the constitutional-pin pattern is consistent with Ethics v1.1. The VSG scrambling speed limit cited in PROT-HH-001 is Vikram-Shou-Galitski (PRL 136, 150401, 2026); the Bekenstein bound is the canonical 1981 result. The holographic hypergraph theory (PROT-HH-001) was developed in a 2026-04-30 directed brainstorming session between J. Butler and the Opus 4.6 Red Team / Architecture instance.

---

*End of Wyrd Proofs Reference v1.4.*
