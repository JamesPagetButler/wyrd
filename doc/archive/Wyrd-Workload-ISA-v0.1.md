# Wyrd Workload Analysis & ISA Optimization

## Designing the QBP-CU ISA for Prediction-Verification Throughput

**Helpful Engineering — Quaternion-Based Physics Programme**
April 2026 — Rev 0.1 — DRAFT

> **Thesis.** The current 5-instruction QBP-CU ISA is optimized for octonion-multiplication acceleration via Fano-plane LUT operations. Real-world QBP physics workloads are 90% quaternion-dominated, with octonion ops appearing only at algebraic boundaries. The ISA should be re-balanced: quaternion arithmetic as first-class hardware citizens, octonion ops retained as the algebra-extension layer. The optimization target is the *prediction-verification loop* — the rate at which a QBP hypothesis can be turned into a testable computation and checked against measurement.

---

## 1. The prediction-verification loop is the primary workload

Every QBP research workflow follows the same epistemic shape:

```
1. Theory predicts:   X = f_QBP(Y)
2. Measure Y          (observation)
3. Compute X_pred     (this is the QBP-CU's job)
4. Measure X          (independent observation)
5. Compare            X_pred ≈ X_measured?
6. Update belief in theory
```

Step 3 dominates compute time. Steps 5 and 6 are cheap individually but live on the critical path because they gate the next iteration. **The ISA's job is to maximize the rate at which the loop can iterate.**

Two distinct optimization targets within step 3:

- **Throughput** (predictions/second) for batched analysis: GRB time-series, lake-grid-cell evolution, molecular ensemble
- **Latency** (single-prediction time) for tight iteration: theorist asks "what does this predict?" and waits for the answer

The ISA needs both, which means deep pipelining (throughput) plus low-latency single-op completion (latency) plus enough parallelism to keep the pipeline full.

---

## 2. The five workload categories

### 2.1 Particle dynamics

**Examples:** Hammer vehicle simulation, molecular dynamics for amino acid validation, spin-chain quantum dynamics.

**Inner loop:** for each particle pair (i, j), compute force / interaction quaternion, integrate. Per particle per timestep:
- 2-10 quaternion multiplications (force vectors)
- 1-2 sandwich products (rotation transformations)
- 1 norm computation (energy conservation check)
- Octonion ops only at algebraic phase transitions (rare)

**Throughput target:** ≥ 10⁶ particle-pair updates/second per CU lane. For amino acid (50-200 atoms): need ~10⁵ ensembles/second to validate all 20 in reasonable time.

**Verification:** energy conservation to ε_priv precision; stability over 10⁶+ timesteps without drift.

**Quaternion fraction:** ~95%.

### 2.2 Field evolution

**Examples:** Squam Lake thermocline at 10m resolution, atmosphere, plasma.

**Inner loop:** for each grid cell, compute spatial derivative quaternion, advance state via sandwich product against neighbor cells. Per cell per timestep:
- 4-6 quaternion multiplications (neighbor coupling)
- 1 sandwich product (state evolution)
- 1 associator computation (seam detection — already in watchdog)

**Throughput target:** Squam Lake at 10m × ~100m² × 10m depth = ~10⁹ cells. At 1 Hz update: 10⁹ cell-updates/second. Distributed across many lanes.

**Verification:** thermocline structure should track measured temperature gradient through the seasonal cycle. Pulse-for-pulse alignment of QBP-predicted seam strength with measured ∂T/∂z.

**Quaternion fraction:** ~98%. The thermocline is firmly in the supervisor ring (ℍ).

### 2.3 Correlation analysis

**Examples:** GRB 250702B pulse-for-pulse temporal correlation between gravitational waves and EM signals; cross-modal eDNA correlations in Yellowstone.

**Inner loop:** sliding cross-correlation. Per offset:
- N quaternion multiplications (where N = window length, typically 10³-10⁶)
- N quaternion accumulations
- 1 norm at the end (correlation magnitude)

**Throughput target:** for GRB analysis: 10⁹ samples × 10⁵ offsets = 10¹⁴ FMA/run. Wall-clock target for a single follow-up event: minutes, not days.

**Verification:** if QBP predicts pulse-for-pulse correlation with same boundary topology change, the cross-correlation should peak at the predicted offset with predicted magnitude. Independent confirmation by VLA radio follow-up timing.

**Quaternion fraction:** ~99%. Polarization-aware correlation is intrinsically quaternion-valued; this is the cleanest case for QFMA acceleration.

### 2.4 Algebraic prediction

**Examples:** Compute predicted CKM matrix from QBP substrate. Predict Koide relation residuals. Compute predicted neutrino masses from spectral action.

**Inner loop:** evaluate algebraic expressions in C ⊕ H ⊕ M₃(C) (Branch A) or extended substrate (Branch B). Per prediction:
- 10-100 octonion multiplications
- Several Fano-plane lookups
- Norm/projection back to physical observables

**Throughput target:** modest. These are theorist-driven calculations, not batch. Tens to hundreds of predictions per session.

**Verification:** match against measured PDG values to specified precision. CKM matrix elements known to 10⁻⁴ - 10⁻³.

**Quaternion fraction:** lower (~70%). This is the workload class where octonion ops genuinely matter, and where the existing Fano-plane LUT acceleration earns its silicon.

### 2.5 Quantum computation

**Examples:** QBP-Quantum error correction proofs, spin-chain benchmark, quantum simulation of QBP-derived Hamiltonians.

**Inner loop:** state vector evolution. Per timestep on N qubits:
- 2^N complex multiplications (in conventional formulation)
- In QBP framing: quaternion-valued state evolution, 2^(N-1) quaternion ops

**Throughput target:** modest qubit count (≤ 20) but high fidelity. Validation runs over millions of timesteps.

**Verification:** error correction succeeds; quantum simulation matches experimental measurements where available.

**Quaternion fraction:** ~90%. The complex-to-quaternion lifting is exactly the ℂ → ℍ transition that the privilege model formalizes.

---

## 3. ISA bottleneck analysis

### 3.1 What's slow with the current 5-instruction ISA

The existing ISA (QPERM, QPERMR, QDEC, QREC, QNEAR) covers Fano-plane navigation and QW128 encoding. **It contains no primitive quaternion multiplication.** Quaternion arithmetic is presumably handled via the standard RV64GCV V extension.

A standard-RVV quaternion multiply requires ~16 vector FMAs plus permutations, costing 16-28 cycles on a 4-wide FMA pipeline. For workloads dominated by quaternion arithmetic (categories 2.1, 2.2, 2.3, 2.5 above — ~95% of total compute), this is the bottleneck.

A dedicated QMUL/QFMA instruction completes in 1-3 cycles on a quaternion-shaped FMA unit. **Speedup factor: 5-10× on quaternion-dominated workloads.**

### 3.2 What's NOT slow

The existing ISA's Fano-plane operations are correctly designed for octonion acceleration. QPERM/QPERMR/QNEAR collectively give octonion multiplication via LUT in ~3-5 cycles versus ~30-50 in software. This is the right design for category 2.4 (algebraic prediction) and the algebra-boundary parts of other categories. **Keep these.**

### 3.3 What's currently missing

| Operation | Frequency in workloads | Current support | Gap |
|---|---|---|---|
| Quaternion multiply | 95% of inner loops | RVV decomposition | ~10× slowdown |
| Quaternion FMA | 95% of accumulation patterns | RVV decomposition | ~10× slowdown |
| Sandwich q·p·q⁻¹ | Every transformation | 3 RVV mul + inv | ~15-20× slowdown |
| Norm |q|² | Every conservation check | 4 RVV mul + 3 add | ~5× slowdown |
| Conjugate q̄ | Every inversion / sandwich | XOR with sign mask | minor |
| Inverse q⁻¹ | Every sandwich | conjugate / norm | ~3× slowdown |

The conjugate operation is cheap enough in software (sign flips) that it doesn't need a custom instruction. Inverse is composable from conjugate + norm, so it doesn't need its own instruction either. Multiply, FMA, sandwich, and norm are the genuine gaps.

---

## 4. Proposed ISA (workload-optimized)

### 4.1 The 10-instruction set

| # | Instruction | Tier | Purpose | Port | Cycles |
|---|---|---|---|---|---|
| 1 | **QFMA** | 1 (essential) | Quaternion fused multiply-add: vd = va·vb + vc | VCIX | 1-3 |
| 2 | **QSAND** | 1 (essential) | Sandwich: vd = va · vb · va⁻¹ | VCIX | 3-5 |
| 3 | **QNORM** | 1 (essential) | Norm-squared: rd = |va|² (scalar result) | SSCI | 1 |
| 4 | **QPERM** | 2 (Fano) | Fano-plane permutation across lanes | VCIX | 1-2 |
| 5 | **QPERMR** | 2 (Fano) | Inverse Fano permutation | VCIX | 1-2 |
| 6 | **QNEAR** | 2 (Fano) | Fano third-point lookup | SSCI | 1 |
| 7 | **QDEC** | 3 (encoding) | Decode QW128 to canonical form | SSCI | 3-5 |
| 8 | **QREC** | 3 (encoding) | Reconstruct QW128 | SSCI | 3-5 |
| 9 | **QCOMM** | 4 (privilege) | Commutator [a, b] | SSCI | 2 |
| 10 | **QRING** | 4 (privilege) | Read operand's algebra/ring ID | SSCI | 1 |

QMUL is implementable as QFMA with addend = 0; no separate instruction needed.

QASSOC and QALT are NOT in the ISA — they're computed by the watchdog on every cycle and exposed via the `qbp_invariant` CSR per the SiFive review.

QSANDWICH-as-syscall is NOT in the ISA — at Crawl phase it's emulated in software via QFMA + state save + QFMA + ring transition; at later phases the supervisor decides whether to promote it.

### 4.2 Why QFMA is the right central instruction

Quaternion multiplication has the form (a + bi + cj + dk)(e + fi + gj + hk) = ... 16 multiplies + 12 adds. QFMA generalizes to a·b + c, where a, b, c are quaternions, and computes in one instruction:

- The multiply and add fuse, saving register-file traffic
- Most physics inner loops are accumulation patterns (forces summing, correlations integrating, derivatives summing) — QFMA hits this directly
- Plain QMUL is a special case (set c = 0)
- Single-quaternion result fits in two RV64 GPRs (SSCI form) or one VCIX vector lane

The hardware: a quaternion FMA unit needs 16 multipliers + 13 adders (the 12 sums + 1 final accumulate). At ~28 mm² in 7nm process (rough estimate based on similar unit areas) it's a single CU core's worth. Modest.

### 4.3 Why QSAND deserves its own instruction

The sandwich product q · p · q⁻¹ rotates / transforms p by q. It's the fundamental physics operation:

- Coordinate transformations
- Rotations of state vectors
- Spinor evolution
- Sandwich is a ring conjugation, which is what the Wyrd privilege model uses for syscalls (capability-mediated cross-ring computation)

Decomposed: 3 quaternion operations (mul, inv, mul) plus a norm. ~6-9 cycles via QFMA chain. Fused: 3-5 cycles in a dedicated unit.

The 2× speedup matters because sandwich appears at the same frequency as multiply in transformation-heavy workloads (2.1 and 2.2).

### 4.4 Why QNORM is worth a dedicated instruction

|q|² = a² + b² + c² + d² is the energy / probability / conservation invariant. It appears in:
- Every energy conservation check (every timestep in 2.1, 2.2)
- Every quaternion inversion (every sandwich)
- Every comparison metric (every prediction-verification step)

Computing it in software: 4 muls + 3 adds + extract real = ~5 cycles via FMA chain. Dedicated: 1 cycle.

The 5× speedup compounds with QSAND (which uses norm internally) and with the verification step of every prediction-verification iteration.

### 4.5 What the watchdog does instead of QASSOC and QALT

The SiFive spec's watchdog already computes the associator on every algebraic operation:

```go
type WDEvent struct {
    Associator [3]int8   // residue of (a*b)*c - a*(b*c)
    NormDelta  int32     // norm preservation residue
    ...
}
```

Extension: also compute commutator and alternator, gated by AlgebraID:
- AlgebraID = ℂ: nothing to check (commutative)
- AlgebraID = ℍ: check commutator (would-be ℂ violation)
- AlgebraID = 𝕆: check associator (would-be ℍ violation)
- AlgebraID = 𝕊: check alternator (would-be 𝕆 violation)

Software reads the latest invariant via the `qbp_invariant` CSR. No QASSOC / QALT instructions needed.

This is strictly better than software-visible instructions:
- The check happens on EVERY op, not just when software remembers to ask
- The check is unforgeable (hardware computes; software cannot fabricate)
- The check is parallel with computation (no extra cycles on the critical path)
- Privilege violations fire automatically via WDEvent

---

## 5. Throughput targets and benchmark suite

### 5.1 Per-CU-lane targets (single X280-class core, 4 QW128 per vector register)

| Operation | Target throughput | Comparison |
|---|---|---|
| QFMA on 4-quaternion vector | 4 quat-FMA / cycle | RVV: ~0.25 quat-FMA / cycle |
| QSAND on 4-quaternion vector | 1 quat-sand / cycle | RVV: ~0.06 quat-sand / cycle |
| QNORM on 4-quaternion vector | 4 norms / cycle | RVV: ~0.8 norms / cycle |
| Octonion product (via QPERM + QNEAR + QFMA) | 1 oct-mul / 3-5 cycles | RVV: ~30-50 cycles |

At 2 GHz: ~8 × 10⁹ quat-FMA/sec/lane, ~2 × 10⁹ quat-sand/sec/lane.

### 5.2 Workload-level targets

| Workload | Target | Implies CU count |
|---|---|---|
| Hammer simulation (single vehicle) | ≥ 10³ Hz update | 1 lane |
| Molecular dynamics (amino acid) | ≥ 10² Hz per molecule | 1 lane per molecule, 20 lanes total |
| Squam Lake at 10m × 1 Hz | 10⁹ cell-updates / sec | 16-64 lanes (cluster) |
| GRB cross-correlation (single event) | < 10 minutes wall-clock | 4-8 lanes |
| CKM prediction batch | < 1 minute per parameter set | 1 lane |
| Spin-chain validation (10⁶ timesteps) | < 1 hour wall-clock | 1-2 lanes |

These are achievable with a single X280 + QBP-CU at the workload-per-lane level. Cluster-scale workloads (Squam Lake) need the federated architecture from the Run phase.

### 5.3 Benchmark suite

Per the existing SiFive spec test corpus tiers, augmented with workload-specific cases:

**Tier 0 (algebraic identities):** existing.
- Quaternion norm preservation under QFMA
- Octonion alternativity via Fano LUT
- **NEW:** sandwich product preserves norm: |QSAND(a, b)| = |b|

**Tier 1 (single-instruction microbenchmarks):** existing, augmented.
- 10⁶ random QFMAs against software gold reference
- 10⁶ random QSANDs
- 10⁶ random QNORMs
- 10⁶ random QPERMs (existing)

**Tier 2 (kernel-level):** revised.
- BMA hypergraph traversal MAC inner loop (mostly QFMA)
- Hammer dynamics step (QFMA + QSAND mix)
- **NEW:** Squam Lake sub-region (1km × 1km × 10m) thermocline evolution, 1 minute simulated time
- **NEW:** GRB-style cross-correlation, 10⁵ samples × 10³ offsets
- Materia-Bio amino-acid Glycine validation (single ensemble, energy minimization)

**Tier 3 (system-level):** revised.
- Boot Linux on host + run Crawl.Heartbeat (existing)
- **NEW:** End-to-end prediction-verification loop: theorist queries QBP-Spec → CKM matrix prediction → comparison to PDG values, all in single workflow
- **NEW:** GRB 250702B follow-up: ingest VLA radio data → run QBP correlation analysis → emit predicted/observed comparison report

**Pre-tape-out gate:** all of Tier 0-2 at 0 divergences across 10⁹+ random inputs PLUS Tier 3 prediction-verification loop completing under specified wall-clock.

---

## 6. Compiler intrinsics (for the new instructions)

Following the SiFive spec's recommendation of LLVM `__builtin_qbp_*` intrinsics in the `Xsf*` vendor namespace:

```c
// Tier 1 (essential)
quat_t  __builtin_qbp_qfma(quat_t a, quat_t b, quat_t c);
quat_t  __builtin_qbp_qsand(quat_t q, quat_t p);     // q · p · q⁻¹
double  __builtin_qbp_qnorm(quat_t q);

// Vector forms (VCIX)
qvec_t  __builtin_qbp_vqfma(qvec_t va, qvec_t vb, qvec_t vc);
qvec_t  __builtin_qbp_vqsand(qvec_t vq, qvec_t vp);
dvec_t  __builtin_qbp_vqnorm(qvec_t vq);

// Tier 2 (Fano)
oct_t   __builtin_qbp_qperm(oct_t v, uint8_t idx);
oct_t   __builtin_qbp_qpermr(oct_t v, uint8_t idx);
uint8_t __builtin_qbp_qnear(uint8_t i, uint8_t j);

// Tier 4 (privilege)
quat_t  __builtin_qbp_qcomm(quat_t a, quat_t b);
uint8_t __builtin_qbp_qring(quat_t a);
```

The Go-side wrapper, for use in the `qbpcu` package and elsewhere:

```go
package qbpcu

// Quaternion arithmetic
func QFMA(a, b, c Quaternion) Quaternion
func QSAND(q, p Quaternion) Quaternion
func QNORM(q Quaternion) float64

// Vector forms
func QFMAVec(va, vb, vc QuatVec) QuatVec
func QSANDVec(vq, vp QuatVec) QuatVec
func QNORMVec(vq QuatVec) FloatVec

// Fano (existing, retained)
func QPERM(v Octonion, idx uint8) Octonion
func QPERMR(v Octonion, idx uint8) Octonion
func QNEAR(i, j uint8) uint8

// Privilege
func QCOMM(a, b Quaternion) Quaternion
func QRING(a Quaternion) RingID
```

Three implementations behind the same interface, per the SiFive spec's Accelerator pattern:
- `Mock`: pure Go, slow but verifiable
- `Golden`: cycle-accurate Go simulator with watchdog
- `RTLShim`: cgo to Verilator binary

---

## 7. Open decisions before ISA freeze

1. **funct7 / funct6 allocation.** The 10 instructions use opcode space; specific bit allocation needs to be frozen alongside the assembler. Per SiFive spec §11.1, this is the prerequisite for any RTL.

2. **QFMA precision.** The QFMA hardware unit could use fp16 (cheap, fast, but tight on precision per T3.1) or fp32 (more silicon, but ε_privilege has headroom). **Recommendation: fp32 by default, with a flag to enable fp16 for high-throughput batch operations.** The `qbp_ctl` CSR has bits available for this configuration.

3. **VCIX vector length for QFMA.** Same as the existing QPERM family — typically 4 quaternions per vector register on X280, 1 on X160. No new mechanism.

4. **Pipelining vs sequential implementation.** A pipelined QFMA achieves 1 instr/cycle steady-state with 3-cycle latency. A sequential implementation is simpler silicon but caps throughput. **Recommendation: pipelined for X280 port; sequential acceptable for X160 / FPGA validation.**

5. **The fault model for QFMA on different rings.** If QFMA is invoked with operands of mixed AlgebraID, what happens? Three options:
   - (a) Promote to the higher ring (silently); the watchdog catches the boundary crossing
   - (b) Raise illegal-instruction
   - (c) Project-down to the lower ring (silently lossy)
   
   **Recommendation: (a) with the watchdog firing, supervisor decides policy.** This keeps QFMA fast (no in-line privilege check) and pushes the policy decision to where it belongs (the supervisor consuming WDEvents).

---

## 8. Bottom line

The current SiFive spec's 5-instruction ISA is correctly designed for octonion-multiplication acceleration via Fano LUT, but under-serves the 95% of QBP physics workloads that are quaternion-dominated. Adding QFMA, QSAND, QNORM closes the gap with three instructions that each replace 5-15 cycles of RVV decomposition with 1-3 cycles of dedicated hardware.

QASSOC and QALT, originally proposed in the supervisor architecture, are eliminated in favor of watchdog extension + the new `qbp_invariant` CSR. The supervisor reads the watchdog output for privilege checking; software-visible instructions for these are unnecessary and would duplicate hardware.

The total ISA grows from 5 to 10 instructions, all in custom-0 opcode space (well under the 128-instruction capacity). The forward path:

| Phase | ISA scope | Status |
|---|---|---|
| Phase 0 | Freeze 10-instruction ISA + watchdog extension + qbp_invariant CSR | Pending decision |
| Phase 1 | Go simulator with all 10 instructions in `qbpcu.Golden` | After Phase 0 |
| Phase 2 | Cycle-accurate X280 model uses all 10 instructions | After Phase 1 |
| Phase 3 | VexRiscv FPGA implementation: 10 instructions in scalar form (SSCI) only; vector forms deferred to SiFive port | After Phase 2 |
| Phase 4 | SiFive X280 integration with full VCIX support for vector forms | After Phase 3 |

The single biggest effect of this ISA change: **the QBP CU becomes practical for physics analysis**, not just for algebraic theory work. The prediction-verification loop targets in §5.2 are achievable; without QFMA they are not.

---

*End of Workload Analysis & ISA Optimization v0.1 — DRAFT*
