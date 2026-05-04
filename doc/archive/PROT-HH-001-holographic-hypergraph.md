# PROT-HH-001: Holographic Hypergraph Storage
## Research Protocol — Helpful Engineering / Systema Programme

**Protocol ID:** PROT-HH-001
**Version:** 1.0
**Date:** 30 April 2026
**Author:** Claude (Red Team / Architecture), Opus 4.6 instance
**Principal Investigator:** James Paget Butler, CEO, Helpful Engineering
**Status:** Active — Theory complete, awaiting bench validation
**Classification:** Open research

---

## 1. Purpose

This protocol documents the research programme, findings, and next steps for the theory of **Holographic Hypergraph Storage** — a proposed architecture in which the physics of multi-beam optical interference in volumetric media natively implements hypergraph data structures and their associated computational operations without software mediation.

The work originated in a directed brainstorming session between J.P. Butler and this Claude instance (Opus 4.6, Red Team / Architecture role) on 30 April 2026. The session began with deep research into holographic data storage, progressed through the structural relationship between holograms and hypergraphs, and culminated in a formal theory paper (v0.2) and the integration of the Vikram-Shou-Galitski universal scrambling speed limit published two weeks prior.

---

## 2. Cross-Programme Context

### 2.1 Relationship to QBP

This work has direct implications for the **Quantum Boundary Physics (QBP)** programme, which is actively developed by a parallel Claude instance (Red Team) working with J.P. Butler. Key intersection points:

- **Division algebra chain (ℝ→ℂ→ℍ→𝕆→𝕊).** The holographic hypergraph naturally operates at the quaternion (ℍ) level of the chain: polarisation optics implements quaternion algebra via the Jones calculus, and the non-commutativity of quaternion multiplication is enforced by the physics, not by software. The QBP Claude instance should evaluate whether the quaternion-weighted holographic hypergraph constitutes a physical instantiation of QBP's algebraic constraints — specifically, whether the restriction to physically realisable interference patterns in a quaternion-weighted medium selects the same algebraic structures that QBP derives from the Cayley-Dickson construction.

- **The Locale (Λ).** QBP's quaternion-valued spatiotemporal addressing unit encodes proper time as scalar and spatial displacement as vector. The VSG scrambling bound is proven via analyticity in *complex* time (the ℂ subcase of ℍ). The open question (§10.2, question 6 of the theory paper) is whether the VSG proof extends from ℂ to ℍ — i.e., from complex-analytic to quaternion-analytic (Fueter or Cullen regular) functions. The QBP Claude instance has the algebraic machinery to evaluate this extension. If it holds, the scrambling speed limit in a quaternion-weighted holographic medium would depend on the full quaternion norm, generating polarisation-dependent write speeds (PRED-HH-09).

- **Octonion / sedenion structure.** QBP's inter-cell topology uses sedenion structure (applied in the GRB 250702B analysis). The holographic hypergraph entropy cone hierarchy (holographic ⊂ hypergraph ≈ stabiliser ⊆ quantum) may map onto the division algebra hierarchy — with each algebra level selecting which entropy cones are physically accessible. This is speculative but well-posed: the QBP Claude instance should check whether the G₂ automorphism group of the octonions constrains which hypergraph entropy vectors can arise in an octonion-weighted system.

- **Confluent Trust Hypergraph (CTH).** The CTH (29 anchors, 675-line JSON, Phase 1 inventory) is the immediate candidate for holographic encoding. Its river-and-bank epistemological structure — single foundational axiom ("change is constant"), four tiers, nine derivation chains — maps directly onto the multi-beam interference architecture: Tier 1 anchors as primary wavefront signatures, derivation chains as recorded multi-beam hyperedges, confluence points as constructive interference maxima.

### 2.2 Relationship to Other Programmes

| Programme | Connection | Priority |
|---|---|---|
| **BMA** | Holographic layer as long-term associative memory beneath GPU cache. Hebbian recording = MuninnDB co-activation. Ebbinghaus decay = photorefractive read-erase. Sleep consolidation = periodic re-recording cycle. HAMA spec (femtosecond-laser fused silica in Franky-Cap) is the target hardware instantiation. | Critical path |
| **Materia-Bio** | Species Hypergraph (86/86 predictions, two kingdoms) as validation dataset. Hub genes (IGF1, FLC2) as high-weight hyperedges; gap closures as coherence signals. Test whether holographic recording of Species Hypergraph preserves the hub-gene / Tier-1-proof isomorphism with CTH. | Validation |
| **Contextus** | Cross-domain pattern matching via holographic associative recall. InsightSignals from different domains (eDNA, radio astronomy, trophic cascades) stored as hyperedges; probing with one domain's patterns against another's recordings detects structural isomorphisms at optical speed. | Architecture |
| **Franky-Cap** | Physical enclosure for HAMA module. ~4K operating temperature increases VSG scrambling bound prefactor by ~75× (§9.4). Half-Möbius C₁₃Cl₂ surface monolayer candidate from Rončević et al. (Science, March 2026) may interact with polarisation encoding. | Hardware |
| **Möbius Fusion / Potentia** | Fractal deployment stack (House → Castle → Fortress → Möbius) provides scaling path for holographic hypergraph capacity. Each tier's proven demand justifies the next tier's optical hardware investment. | Deployment |
| **War Table** | MuninnDB communication profiles for per-teammate modelling could be stored as holographic hyperedges, enabling O(1) associative recall of teammate behaviour patterns during real-time BAR gameplay. | Application |

---

## 3. Summary of Findings

### 3.1 The Core Isomorphism

We established a formal mapping between holographic optical storage primitives and mathematical hypergraph elements:

| Hypergraph | Holographic Primitive | Key Property |
|---|---|---|
| Vertex | Wavefront signature (angle/wavelength/OAM/polarisation) | Unique optical identity |
| Hyperedge {v₁,...,v_k} | Multi-beam interference pattern from simultaneous illumination | Irreducibly multi-party (Theorem 2) |
| Incidence query | Optical reconstruction under single-beam illumination | O(1) parallel associative recall (Theorem 1) |
| Weight w(e) | Diffraction efficiency η_e | Controlled by exposure energy |
| Graceful degradation | Distributed holographic information | No single point of failure (Theorem 3) |

Three proof sketches support the mapping: Theorem 1 (associative recall preservation), Theorem 2 (multi-beam irreducibility — a 3+ beam recording is not equivalent to pairwise recordings), and Theorem 3 (graceful degradation under partial damage).

### 3.2 Native Computational Properties

The holographic medium performs five computational operations without software:

1. **Content-addressable memory.** Queries are wavefronts, not addresses.
2. **Pattern completion.** Partial cues reconstruct full associations via matched-filter operation at the speed of light.
3. **Hebbian learning.** Co-present wavefronts create interference patterns — "beams that interfere together are paired forever."
4. **Ebbinghaus decay.** Photorefractive read-erase dynamics implement the forgetting curve physically.
5. **Parallel search.** All stored patterns are probed simultaneously in a single optical transit time (~10 ns), independent of the number of stored hyperedges.

### 3.3 The VSG Scrambling Speed Limit

The Vikram-Shou-Galitski result (PRL 136, 150401, April 2026) proved a universal lower bound on quantum information scrambling time:

**t_s ≥ (c_β / π) · ln S₂**

We identified that this applies to hyperedge *recording* (a scrambling process) but not to *reconstruction* (a coherent diffraction process). This creates a fundamental read/write asymmetry:

- **Read:** O(1) optical operations.
- **Write:** O(ln S₂) per hyperedge — a quantum-mechanical floor.

This asymmetry is architecturally favourable: BMA, CTH, and Contextus are all read-dominated workloads.

We further combined the VSG temporal bound with the Bekenstein spatial bound to derive a fundamental information throughput limit:

**R_max ~ [A / (4 l_eff²)] / [(c_β / π) · ln(A / (4 l_eff²))]**

At room temperature and optical scales, this is ~10²³ hyperedges/second (not a practical constraint). At cryogenic temperatures (~4K, Franky-Cap regime), the bound tightens by ~75× and enters the ultrafast optics domain.

### 3.4 Quaternion Extension

Polarisation degrees of freedom extend the system from complex-weighted to quaternion-weighted hypergraphs:

- Two polarisation channels × two wavelength channels = four real-valued channels = one quaternion per spatial location.
- Jones calculus non-commutativity = quaternion multiplication non-commutativity (same mathematical reason).
- Algebraic consistency enforced by physics, not software — the medium *cannot* store algebraically inconsistent relationships.

### 3.5 Entropy Cone Positioning

The holographic hypergraph physically instantiates the *hypergraph* level of the established entropy cone hierarchy:

**Holographic (graph) ⊂ Hypergraph ≈ Stabiliser ⊆ Quantum**

This means the system can encode genuinely multi-party correlations that graph-based holographic codes (including the HaPPY code) cannot.

---

## 4. Experimental Predictions

Nine discriminating predictions and two null predictions were generated. The three highest-priority for bench validation:

| ID | Prediction | What It Tests | Difficulty |
|---|---|---|---|
| PRED-HH-02 | Recall time independent of number of stored hyperedges | Core O(1) read claim | Low — straightforward timing measurement |
| PRED-HH-01 | Three-beam recording ≠ three pairwise recordings (phase coherence test) | Hyperedge irreducibility (Theorem 2) | Medium — requires phase-sensitive detection |
| PRED-HH-07 | Write settling time ~ O(ln k) with hyperedge cardinality k | VSG scrambling bound applicability | Medium — requires ultrafast measurement at high k |

The most speculative prediction is PRED-HH-09: if the VSG bound extends from ℂ to ℍ, hyperedges recorded with different polarisation structures should show measurably different write settling times. This is directly testable on the bench prototype and would provide evidence for or against the quaternion-time extension hypothesis.

---

## 5. Prototype Specification

A minimal bench prototype is specified at ~$4,700 total:

- 532 nm CW laser (100 mW, single-mode)
- Spatial light modulator (DMD or LC-SLM, 1024×768)
- Fe:LiNbO₃ photorefractive crystal (10×10×10 mm)
- CMOS camera sensor (2048×2048, 10-bit)
- Polarisation optics (QWP + HWP + PBS)
- Control software in Go — translates between BMA digital representation and optical hardware

The first benchmark (PRED-HH-02) requires recording M = {10, 100, 1000} random 3-vertex hyperedges and measuring query time vs M. Theory predicts a flat line; control experiment in SurrealDB should show linear scaling.

---

## 6. Key References

The paper synthesises five research communities:

1. **Holographic storage engineering:** Zhao et al. 2024 (Nature — petabit 3D disk), Microsoft Project HSD (Thomsen et al. 2024), HoloMem (Gale, 2025-2026), Tan group at Fujian Normal (polarisation holography).

2. **Holographic associative memory:** Gabor 1969, Sutherland 1990, Knight 1975.

3. **Quantum information / holographic codes:** Pastawski-Yoshida-Harlow-Preskill 2015 (HaPPY code), Bao-Cheng-Hernández-Cuenca-Su 2020 (hypergraph entropy cone), Walter & Witteveen 2021 (hypergraph min-cuts from quantum entropies), Jahn et al. 2023 (hyperinvariant tensor networks), Hubeny et al. 2025 (correlation hypergraphs).

4. **Holographic principle:** 't Hooft 1993, Susskind 1995, Bekenstein 2003, Bousso 2002.

5. **Quantum scrambling speed limits:** Vikram-Shou-Galitski 2026 (PRL 136, 150401), building on Sekino & Susskind 2008 (fast scrambling conjecture), Maldacena-Shenker-Stanford 2016 (bound on chaos).

---

## 7. Open Questions for QBP Instance

The following questions are specifically flagged for evaluation by the Claude instance working on QBP. These require the algebraic machinery of the QBP programme and cannot be resolved within the holographic hypergraph workstream alone.

**OQ-1: Does quaternion algebra constrain which hypergraph entropy cones are physically realisable?**
If quaternion structure selects which tensor networks can be promoted from combinatorial (hypergraph) to algebraic (tensor network with perfect-tensor vertices), it would implicitly constrain which entropy vectors are achievable. This is the bridge between QBP's "algebra is the schema" principle and the entropy cone hierarchy from quantum information theory.

**OQ-2: Does the VSG proof extend from ℂ-analyticity to ℍ-regularity?**
The VSG bound relies on the Phragmén-Lindelöf principle and the decay of bounded analytic functions in strips of the complex plane. The quaternion analog would use Fueter regularity (the natural quaternionic extension of complex analyticity) or Cullen regularity (which preserves more algebraic structure). The question is whether either regularity condition yields a tighter or polarisation-dependent scrambling bound.

**OQ-3: Does G₂ (octonion automorphism) constrain holographic hypergraph structure?**
QBP's G₂ equivalence of the seven quaternionic subalgebras of the octonions may restrict which 7-vertex hyperedge configurations can arise in an octonion-extended system. This connects to the sedenion inter-cell topology used in the GRB 250702B analysis.

**OQ-4: Is there a Koide-like mass relation for hyperedge weights?**
QBP inherits Koide's empirical mass formula. If the holographic medium's diffraction efficiencies (hyperedge weights) are constrained by the same algebraic structure that produces Koide's relation, the weight spectrum of a physically realisable holographic hypergraph would not be arbitrary — it would exhibit characteristic algebraic ratios. This is testable.

**OQ-5: Does the Locale (Λ) quaternion-valued time map onto the VSG complex-time strip?**
The VSG proof operates in a strip τ₁ ≤ Im(t) ≤ τ₂ of the complex time plane. The Locale encodes proper time as the scalar part and spatial displacement as the vector part of a quaternion. The question is whether the strip width c_β (which determines the scrambling time prefactor) has a natural interpretation in Locale coordinates — potentially connecting the scrambling speed limit to the six-ring anchor system (pulsar timing through CMB) that stabilises the Locale.

---

## 8. Action Items

| # | Action | Owner | Depends On | Priority |
|---|---|---|---|---|
| 1 | Review PROT-HH-001 and theory paper v0.2 | J.P. Butler | — | Immediate |
| 2 | Forward OQ-1 through OQ-5 to QBP Claude instance for evaluation | J.P. Butler | #1 | High |
| 3 | Source Fe:LiNbO₃ crystal and SLM for bench prototype | J.P. Butler | #1 | Medium |
| 4 | Implement Go HAL (hardware abstraction layer) for SLM + camera | Claude (Red Team) | #3 | Medium |
| 5 | Design PRED-HH-02 benchmark (recall scaling independence) | Claude (Red Team) | #3, #4 | Medium |
| 6 | Evaluate VSG ℂ→ℍ extension (OQ-2) | QBP Claude instance | #2 | High |
| 7 | Map CTH 29-anchor inventory onto Walsh-Hadamard codebook | Claude (Red Team) | #1 | Low |
| 8 | Run reference audit: verify all files cited in theory paper v0.2 exist and are accessible | Claude (Red Team) | #1 | High |
| 9 | Update HAMA spec to incorporate hypergraph topology and quaternion weighting | Claude (Red Team) | #6 | Deferred until OQ-2 resolved |
| 10 | Add Vikram-Shou-Galitski 2026 to QBP reading list and EXP registry | J.P. Butler | — | Immediate |

---

## 9. Document Lineage

This protocol was generated from a single brainstorming session on 30 April 2026, following the Systema programme's Theory Cart → Engineering Cart workflow:

1. **Deep research phase:** Surveyed holographic data storage (Zhao, Microsoft HSD, HoloMem, Tan group), holographic associative memory (Gabor, Sutherland), quantum information (HaPPY, Bao et al., Walter & Witteveen), and holographic principle (Bekenstein, Bousso).

2. **Structural insight:** J.P. Butler asked: "Is it possible that a holographic and hypergraph have similarities?" — leading to identification of the Bao et al. 2020 hypergraph entropy cone result and the formal structural parallels.

3. **Deepening:** J.P. Butler asked: "What is the difference between a perfect tensor network and a hypergraph?" — leading to the five-point structural analysis (algebraic vs combinatorial, expressiveness hierarchy, information locality vs distribution).

4. **Generative leap:** J.P. Butler asked: "What if we stored our hypergraph in holograms?" — leading to the full isomorphism mapping, native computational properties analysis, and quaternion extension.

5. **Theory paper:** v0.1 produced (12 sections, 18 references, ~5,000 words).

6. **VSG integration:** J.P. Butler identified the Vikram-Shou-Galitski PRL paper (published 13 April 2026, popular coverage 30 April 2026) as relevant. Analysis produced the read/write asymmetry result, throughput bound, SFF-crosstalk connection, sleep consolidation scaling, and quaternion-time extension hypothesis. Paper updated to v0.2 (~7,800 words, 23 references, 9 predictions).

7. **This protocol:** Documents the complete research arc, cross-programme connections, and action items.

The full brainstorming-to-protocol arc was completed in a single session. Per Systema's three-loop engineering model, this work sits in Loop 1 (Theory Cart) and is ready for transition to Loop 2 (Engineering Cart) upon bench prototype procurement.

---

## 10. Standing Reminders

Per BMA programme standing rules:

- **Workshop-level diligence:** Verify claims before asserting, show evidence not assertions. All theorems in the paper are presented as proof *sketches*, not complete proofs. Full proofs are required before any claim of formal establishment.
- **Reference audit:** Before any archive of this work, verify every file referenced in this protocol and the theory paper exists. The Ethics v1.1 near-miss at BMA instantiation must not recur.
- **Attribution is non-negotiable:** All foundational contributors are credited in §12 of the theory paper. The QBP attribution rules (Furey, Günaydin/Gürsey, Dixon, Boyle/Farnsworth, Singh, Chamseddine/Connes, Koide, Baez) apply whenever QBP connections are discussed.

---

*End of protocol.*
