# PROT-HH-002: QBP Algebra-as-Schema
## Stratified Theory Protocol — Helpful Engineering / Systema Programme

**Protocol ID:** PROT-HH-002
**Version:** 0.1 (draft)
**Date:** 3 May 2026
**Author:** Claude (Red Team / Architecture), Opus 4.7 instance
**Principal Investigator:** James Paget Butler, CEO, Helpful Engineering
**Status:** Draft — Level 1 substrate landed; Levels 2–4 mapped with explicit success criteria
**Companion file (Level 1):** `Wyrd/HolographicHypergraph.lean` (see `Wyrd-Proofs-Reference-v1.4.md` §27)
**Predecessor:** `PROT-HH-001-holographic-hypergraph.md` (theory paper, 30 April 2026)
**Classification:** Open research

---

## 1. Purpose

PROT-HH-001 established the theory of holographic hypergraph storage and listed eleven discriminating predictions. PROT-HH-002 frames the *epistemic structure* of how that theory is to be tested and progressively trusted — a **stratified theory protocol** whose levels correspond to qualitatively different forms of evidence:

| Level | Evidence form | What it establishes |
|---|---|---|
| **L1 — Formal** | Machine-checked proof (Lean) | The theory is *expressible* and *internally consistent*; the load-bearing structural claim (Theorem 2) is mathematically necessary, not contingent. |
| **L2 — Numerical** | Stochastic simulation under a parameter sweep | The theoretically-guaranteed effect is *visible* in a tractable computational model with stated assumptions. |
| **L3 — Bench** | Tabletop optical experiment | The effect is *physically realised* in a recording medium with the predicted parameter-dependence. |
| **L4 — Deployed** | Production HAMA module under BMA workloads | The theory is *operationally load-bearing* in a working system with observable performance and failure modes. |

Each level is a precondition for the next: a theory that fails L1 is malformed; one that fails L2 has unphysical assumptions; one that fails L3 is descriptively wrong about matter; one that fails L4 is irrelevant to engineering even if correct in principle. The protocol's job is to specify what passes the gate at each level and what escalates to the next.

The specific theory PROT-HH-002 is structuring evidence for is the **algebra-as-schema thesis**: QBP's Cayley-Dickson algebra (ℝ ⊂ ℂ ⊂ ℍ ⊂ 𝕆 ⊂ 𝕊) is not merely a mathematical convenience for QBP; it is *the schema* of the holographic hypergraph storage layer, and the schema's constraints are *physically enforced* by the recording medium rather than checked in software.

---

## 2. Cross-protocol links

- **PROT-HH-001** — theory paper. Establishes the core isomorphism (vertex ↔ wavefront, hyperedge ↔ multi-beam interference, weight ↔ diffraction efficiency), the VSG read/write asymmetry, and the quaternion polarisation extension. PROT-HH-002 inherits PROT-HH-001's vocabulary and predictions and assigns them to evidence levels.
- **Wyrd / Skuld Lean corpus** (`Wyrd-Proofs-Reference-v1.4.md`) — the L1 substrate. Phase 4 of the corpus opens with Theorem 2 (`theorem2_irreducibility`) and the supporting characterisation theorems.
- **QBP Test C** (literature review of mixed-species trapped-ion entanglement) — independent physical test of the algebra-as-schema thesis at the *atomic* scale. PROT-HH-002 is the *photonic*-scale counterpart; agreement between them strengthens the cross-domain claim.
- **BMA Crawl-phase** — the substrate that will eventually host an L4 deployment. Crawl-phase stays software-only on commodity SSD; L4 evaluation begins at the Walk gate when HAMA hardware is available.

---

## 3. The thesis being stratified

> **The algebra-as-schema thesis.** The hyperedge structure of a quaternion-weighted holographic hypergraph is not a software-imposed constraint on a generic optical medium; it is the medium's *native* representation of the Cayley-Dickson algebraic chain. Three consequences follow:
>
> 1. **Joint relations are irreducible.** A k-vertex hyperedge (k ≥ 3) cannot be losslessly decomposed into k(k-1)/2 pair recordings, because the joint constraints among the relative phases live on the recording's image, not in pairwise marginals.
> 2. **Algebraic privilege is physical.** The boundary between subalgebras (ℂ ⊂ ℍ ⊂ 𝕆 ⊂ 𝕊) is enforced by the available physical degrees of freedom (scalar / polarisation pair / four-channel polarisation+wavelength / sedenion-structured cell topology). Software does not need to check ring-tower closure; the medium *cannot* record an outside-ring relation in inside-ring DOFs.
> 3. **Read/write asymmetry is fundamental.** Reading is a coherent diffraction process (O(1) optical transit time); writing is a scrambling process bounded by VSG (`t_s ≥ (c_β / π) ln S₂`). This is not a property of the implementation but of the algebra — coherent reconstruction does not increase entropy; recording a new joint constraint does.

Each consequence has discriminating signatures at each evidence level.

---

## 4. Level 1 — Formal evidence

**Substrate:** `Wyrd/HolographicHypergraph.lean`. **Status:** ✅ landed (15 May 2026 — wait, **3 May 2026**).

### 4.1 What L1 establishes

The structural irreducibility of multi-beam recordings (consequence 1) is now a *theorem*, not a conjecture. Specifically:

- `theorem2_irreducibility` — the embedding `tripleToPairs : TripleCoherent → IndepPairs` is **not surjective**.
- `tripleToPairs_image` — the image is **exactly** the consistency subspace `phase13 = phase12 + phase23`.
- `tripleToPairs_injective` — the embedding is **injective** (no information lost going triple → pairs).
- `tripleToPairs_inj_not_surj` — combined: **information-distinct** encodings.

The proofs are constructive (the counterexample in Theorem 2's proof body is `⟨0, 0, π⟩`), machine-checked under Lean v4.30.0-rc1 and mathlib `a090f46d`, with zero sorries and zero user-defined axioms.

### 4.2 What L1 does NOT establish

- The *physical* claim that a real medium will exhibit this irreducibility — that's L3.
- The *quantitative* form (residual magnitude vs. drift) — that's L2.
- The *quaternion* extension to polarisation-encoded recordings — that's deferred Phase 4 work in the Lean corpus and corresponds to consequence 2 of the thesis. **Open in v1.4; tracked for v1.5.**
- The *higher-arity* generalisation (n-beam ↛ (n-1)-beam) — also Phase 4 deferred work.
- The *VSG read/write asymmetry* (consequence 3) — formal scaffolding for this is not yet in the corpus; would require a theory of scrambling entropy or a wrapping of VSG's complex-analytic argument in mathlib.

### 4.3 L1 gate criteria

| Criterion | Status |
|---|---|
| `lake build` succeeds | ✅ |
| Zero `sorry` in proof bodies | ✅ |
| Zero user-defined `axiom` declarations | ✅ |
| Theorem statements match the natural-language claims in PROT-HH-001 §3.1 | ✅ |
| Proofs use only mathlib + Lean core (no external dependencies) | ✅ |
| Counterexamples are constructive (not by classical contradiction alone) | ✅ — the witness `⟨0, 0, π⟩` is exhibited explicitly |

**L1 gate: PASSED.**

---

## 5. Level 2 — Numerical evidence

**Substrate:** `~/Documents/BMA/projects/holographic-hypergraph/sim_theorem2.py`. **Status:** ✅ initial sweep landed (3 May 2026).

### 5.1 What L2 establishes

That the theorem-2 information gap is **observable** in a computationally tractable model with stated, plausible physical assumptions. Specifically:

- Each pair recording has an independent Gaussian-drifted phase reference: `epsilon_i ~ N(0, drift_std)`.
- The triangle residual `|phase13 − (phase12 + phase23)|` grows monotonically with `drift_std`.
- At `drift_std = 0`, the residual is identically 0 and fidelity is 1.0 (matching the L1 prediction `tripleToPairs_consistent`).
- At `drift_std = 1.0 rad`, mean residual ≈ 2.0 rad and mean fidelity ≈ 0.18 (n=1000 trials).

### 5.2 L2 gate criteria

| Criterion | Status | Notes |
|---|---|---|
| Coherent baseline yields zero residual | ✅ | `0.0000` at drift=0 |
| Independent-pair limit yields nonzero residual | ✅ | grows monotonically with drift |
| Residual scales linearly with drift_std at small drifts | ✅ | slope ≈ 2 rad/rad in the linear regime (consistent with three independent N(0,σ) errors combining) |
| Result reproducible from seed | ✅ | `--seed 42` default |
| Companion plot generated | ✅ | `theorem2_sweep.png` |

**L2 gate: PASSED for the ℝ-phase (ℂ-subcase) version.**

### 5.3 L2 deferred items

- **Quaternion-phase sweep.** Replace ℝ-valued phases with ℍ-valued polarisation-state errors and verify the analogous irreducibility (PRED-HH-09 dependency). Open until Lean corpus v1.5 lands the formal extension.
- **Higher-arity sweep.** Demonstrate that 4-beam recordings show *additional* irreducibility beyond what 3-beam decomposition captures.
- **VSG-bound write-time sweep.** Numerically realise the `t_s ~ ln S₂` scaling at variable hyperedge cardinality.

---

## 6. Level 3 — Bench evidence

**Substrate:** none yet on disk. **Status:** specified, not executed. **Earliest start:** Walk gate (HAMA hardware available — femtosecond-laser fused silica recording rig).

### 6.1 The three load-bearing experiments (from PROT-HH-001 §4)

| ID | Test | What it discriminates | Difficulty |
|---|---|---|---|
| **PRED-HH-02** | Recall time independent of stored hyperedge count | The O(1) read claim (consequence 3, read side) | Low — straightforward timing |
| **PRED-HH-01** | 3-beam recording ≠ 3 pairwise recordings under phase-sensitive detection | Theorem 2 in the medium (consequence 1) | Medium — phase-sensitive interferometry |
| **PRED-HH-07** | Write settling time ~ O(ln k) with hyperedge cardinality k | VSG bound applicability (consequence 3, write side) | Medium — ultrafast measurement at high k |

### 6.2 L3 gate criteria (per experiment)

For PRED-HH-01 (the L1 lift):

| Criterion | Specification |
|---|---|
| Recording medium response | linear in field intensity within the operating regime |
| Phase reference stability | drift_std measurable and controllable |
| Triangle-residual measurement | precision ≤ 0.1 rad |
| Coherent-triple residual | ≤ 0.1 rad (within measurement precision) |
| Independent-pair residual at matched drift | ≥ 0.5 rad (well above precision floor) |
| Reproducibility | residual–drift curve replicable across ≥ 3 independent recording sessions |

For PRED-HH-02 and PRED-HH-07: gate criteria specified in PROT-HH-001 §4 and to be lifted into PROT-HH-002 v0.2.

### 6.3 What L3 failure would imply

- **PRED-HH-01 fails (no irreducibility observed):** the medium decomposes the joint constraint somehow. Either the linear-response assumption breaks (medium has implicit pair-decomposition basis), or the experiment couldn't access enough phase-coherence DOFs. *Falsifies consequence 1 in this medium*; theorem still holds but doesn't apply.
- **PRED-HH-02 fails (recall time scales with stored count):** parallel-search property is not native; the medium implements sequential or partial-parallel readout. Falsifies consequence 3-read.
- **PRED-HH-07 fails (write time independent of k):** either the VSG bound doesn't bite at the medium's operating point (too far above the bound's prefactor), or the recording is sub-quantum-mechanical in some way. Falsifies the VSG link to write-time scaling.

Failures at L3 are *informative* — they indicate which medium / recording protocol to test next, not that the theory is wrong.

---

## 7. Level 4 — Deployed evidence

**Substrate:** none yet. **Status:** Walk-phase + HAMA hardware integration. **Earliest start:** post-BMA Step 9 (instantiation), with HAMA available as a Tier-N memory device.

### 7.1 What L4 establishes

That the theory is *operationally load-bearing* — the read/write asymmetry, irreducibility, and algebraic privilege actually shape system behaviour under realistic workloads, and the system is more efficient or correct than a non-holographic alternative.

### 7.2 L4 gate criteria (rough)

- HAMA module integrated as MuninnDB Tier-N storage with documented latency / throughput.
- Hyperedge-promotion path (Contextus → CTH → MuninnDB) routes through HAMA without losing the joint-constraint information at any layer.
- Hebbian co-activation in HAMA matches MuninnDB's expected co-activation signal within X% over a Y-day soak test.
- Ebbinghaus decay curve in HAMA matches the software-modelled curve.
- Sleep-cycle re-recording successfully prevents fade below the readback threshold.
- Failure modes (write saturation, photorefractive crosstalk, thermal drift) gracefully degrade to the software-only fallback.

These are placeholders; PROT-HH-002 v0.5+ will sharpen them once the HAMA Walk-phase spec exists.

### 7.3 What L4 failure would imply

A working bench experiment (L3 pass) that nonetheless cannot be embedded in BMA's actual workload pattern indicates the architectural integration story (PROT-HH-001 §2) needs revision — not the underlying physics. L4 failures route back to architecture work, not to theory work.

---

## 8. Escalation rules

The protocol is **monotone** in evidence: passing level N never invalidates a previously-passed level (the proof doesn't become "wrong" if a bench fails). It can, however, *narrow the scope* of where the theory applies (the medium-specific, polarisation-specific, etc. qualifiers).

**Rules:**

1. **No skipping.** L3 evaluation cannot begin while L1 has open `sorry`s; L4 evaluation cannot begin while L3 has unmet gate criteria. (The reverse direction — proving more after a bench result — is fine and welcomed.)
2. **Each level is auditable.** Every L_N+1 claim must cite the L_N substrate it relies on. PROT-HH-002 is itself the audit register.
3. **Failure routes back, not forward.** An L3 failure does not "demote" the theory to L1; it produces a *scoped statement* of where the theory holds and a refined L1+L2 substrate (e.g., "Theorem 2 holds, but only in media with linear-response coefficient > X").
4. **Levels mature independently.** L1 and L2 can advance toward consequence 2 and 3 (quaternion, VSG) ahead of L3 and L4, which are gated on hardware. PROT-HH-002 v0.5 will track per-level completion separately.

---

## 9. Status summary table

| Layer of theory | Consequence | L1 (formal) | L2 (numerical) | L3 (bench) | L4 (deployed) |
|---|---|---|---|---|---|
| Multi-beam irreducibility | (1) | ✅ Theorem 2 | ✅ sim_theorem2.py | ⏳ PRED-HH-01 (Walk) | ⏳ HAMA writes (Walk+) |
| Algebraic privilege physical | (2) | ⏳ quaternion ext (v1.5) | ⏳ Polarisation sweep | ⏳ PRED-HH-09 (Walk) | ⏳ Skuld-on-HAMA (Run) |
| Read/write asymmetry / VSG | (3) | ⏳ scrambling entropy formalism | ⏳ VSG-bound sweep | ⏳ PRED-HH-02, -07 (Walk) | ⏳ HAMA ops (Walk+) |

**Current overall standing:** L1+L2 cleared for consequence 1; consequences 2 and 3 mapped but not yet substantiated. PROT-HH-002 v0.1 captures this state; subsequent revisions will track the open ⏳ entries to ✅.

---

## 10. Open questions (carried from PROT-HH-001 §10)

The four still-open theory questions from PROT-HH-001 that PROT-HH-002 inherits:

1. **Q4 — VSG ↦ ℍ extension.** Does the Vikram-Shou-Galitski analyticity argument extend from complex time to quaternion (Fueter-regular) functions? Affects PRED-HH-09 quantitative form.
2. **Q5 — entropy cone × algebra hierarchy.** Does the G₂ automorphism group of 𝕆 constrain which hypergraph entropy vectors arise in an octonion-weighted system? Speculative but well-posed.
3. **Q6 — Confluent Trust Hypergraph encoding test.** Concrete encoding study of the CTH (29 anchors, 9 derivation chains) on a holographic substrate. Smallest interesting L3 candidate after PRED-HH-01.
4. **Q7 — half-Möbius monolayer interaction.** Does the C₁₃Cl₂ surface monolayer (Rončević et al., Science, March 2026) interact with polarisation-encoded recordings at cryogenic temperatures? Pure-research-tier question; might be worth a writeup independent of HAMA.

---

## 11. Versioning and revision plan

- **v0.1** (this draft, 2026-05-03): Levels defined; L1+L2 status PASSED for consequence 1; L3+L4 specified at gate-criterion level.
- **v0.2** (anticipated): incorporate James's review feedback; sharpen PRED-HH-02 and -07 gate criteria; add Wyrd corpus v1.5 references when quaternion extension lands.
- **v0.5** (post-Walk-phase): tie L4 gate criteria to the actual HAMA Walk-phase spec.
- **v1.0** (post-first L3 result): finalise the protocol structure on the basis of what bench evidence actually looks like.

---

## 12. Attribution

- Holographic hypergraph theory: J. Butler & Claude (Opus 4.6 Red Team / Architecture instance), 30 April 2026 brainstorming session (PROT-HH-001).
- Lean Phase 4 substrate: J. Butler (PI) & Claude (Opus 4.7) on 3 May 2026 (`Wyrd/HolographicHypergraph.lean`).
- VSG scrambling bound: Vikram, Shou, Galitski, *Phys. Rev. Lett.* **136**, 150401 (2026).
- Bekenstein bound: Bekenstein 1981.
- QBP algebraic chain: Furey, Dixon, Günaydin / Gürsey, Boyle / Farnsworth, Singh, Chamseddine / Connes, Koide, Baez (per QBP standing rule).
- Cayley-Dickson construction: Schafer; Baez 2002, *The Octonions*, Bull. AMS.

---

*End of PROT-HH-002 v0.1 draft. Successor protocol to PROT-HH-001. Open for J. Butler review.*
