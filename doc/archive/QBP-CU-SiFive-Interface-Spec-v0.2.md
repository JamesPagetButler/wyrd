# QBP-CU ↔ SiFive Interface Specification

**Version:** 0.2 (incorporating Skuld supervisor integration and workload-driven ISA)
**Author:** James Paget Butler
**Editor:** Claude Opus 4.7 (red-team / architecture role)
**Status:** Working document. Pre-RTL. The contract between the Go cycle-accurate simulator, the eventual SiFive integration, the Gemini theory side, and the Skuld supervisor (per *Wyrd-Supervisor-Architecture-v0.2.md*).

---

## Changes from v0.1

1. **ISA expanded to 10 instructions** per *Wyrd-Workload-ISA-v0.1.md*. New: QFMA, QSAND, QNORM (quaternion arithmetic for prediction-verification throughput), QCOMM, QRING (privilege-related). QASSOC and QALT explicitly NOT added — see §6 for why.
2. **Watchdog as privilege boundary detector**, not just algebraic invariant monitor. Watchdog now computes commutator, associator, AND alternator, gated by AlgebraID. New CSR `qbp_invariant` exposes the latest invariant tuple.
3. **AlgebraID renumbered** to align with Wyrd ring tower: 0=ℂ, 1=ℍ, 2=𝕆, 3=𝕊. Branch A (4) and Branch B (5) reserved for physics modes.
4. **WDEvent struct extended** with `Commutator`, `Alternator`, and `RingTransition` fields.
5. **CSR access privilege column** — every CSR gets a minimum AlgebraID required for read/write, matching the algebraic rings.
6. **Watchdog event consumer is Skuld**, not BMA's judge collective. Section 10 revised. BMA receives Skuld's post-policy stream.
7. **funct6/funct7 bit allocations** committed (formerly TBD).

---

## 1. Purpose

Define how the QBP Compute Unit — comprising the ten custom-0 instructions, the algebraic LUT, the quaternion FMA unit, and the algebraic watchdog — couples to a SiFive RISC-V host core in a way that:

1. Preserves single-toolchain operation (upstream LLVM, no proprietary SDK).
2. Lets the QBP-Go cycle-accurate simulator stand in for the accelerator block during architectural exploration.
3. Provides observable hooks at every algebra-crossing for the watchdog to verify quaternion / octonion invariants per cycle.
4. **NEW:** Exposes the watchdog's privilege boundary detection to the Skuld supervisor via a host-pinned ring buffer, with deterministic event ordering for cosim reproducibility.
5. **NEW:** Maximizes throughput on the prediction-verification loop, the dominant compute pattern in QBP physics workloads.
6. Survives changes to QW128 internals, algebraic LUT representation, or the substrate decision (Branch A vs. Branch B per *Wyrd-BranchA-Contingency-v0.1.md*).

## 2. Target host

Unchanged from v0.1. Primary X280 Gen 2; secondary X160 Gen 2 for FPGA validation.

## 3. Coupling decision (revised for 10-instruction ISA)

| Op | Shape | Port | Tier | Rationale |
|---|---|---|---|---|
| `QFMA`   | Quaternion fused multiply-add across N lanes | **VCIX** | 1 (essential) | Vector-native; the dominant workload op |
| `QSAND`  | Sandwich q·p·q⁻¹ across N lanes | **VCIX** | 1 (essential) | Vector-native; transformation-heavy workloads |
| `QNORM`  | Norm-squared, scalar result per lane | **SSCI** | 1 (essential) | Reduce-to-scalar; energy/probability checks |
| `QPERM`  | Algebraic LUT permutation across N lanes | **VCIX** | 2 (Fano) | Vector-native; LUT-driven |
| `QPERMR` | Inverse permutation across N lanes | **VCIX** | 2 (Fano) | Same shape, inverse table |
| `QNEAR`  | Algebraic LUT third-element lookup | **SSCI** | 2 (Fano) | Single-element, control-flowy |
| `QDEC`   | Reduce QW to canonical form | **SSCI** | 3 (encoding) | Branchy, scalar, watchdog-heavy |
| `QREC`   | Reconstruct QW from canonical form | **SSCI** | 3 (encoding) | Inverse of QDEC |
| `QCOMM`  | Commutator [a, b] | **SSCI** | 4 (privilege) | Boundary detection ℂ→ℍ; software-visible |
| `QRING`  | Read operand's AlgebraID | **SSCI** | 4 (privilege) | Single-cycle privilege check |

The accelerator block presents both ports; the LUT, watchdog, and quaternion FMA unit are shared resources arbitrated per §6.

## 4. Instruction encoding (FROZEN)

### 4.1 SSCI (scalar)

Custom-0 opcode space `0001011` (RISC-V `custom-0`). funct3 distinguishes operand classes; funct7 selects the specific instruction.

| Mnemonic | funct7      | rs2 / aux  | rs1       | funct3 | rd          | opcode    |
|----------|-------------|------------|-----------|--------|-------------|-----------|
| `QDEC`   | `0000000`   | qw_in[hi]  | qw_in[lo] | `000`  | qw_out[lo]  | custom-0  |
| `QREC`   | `0000001`   | qw_in[hi]  | qw_in[lo] | `001`  | qw_out[lo]  | custom-0  |
| `QNEAR`  | `0000010`   | idx        | qw_in[lo] | `010`  | qw_out[lo]  | custom-0  |
| `QNORM`  | `0000011`   | qw_in[hi]  | qw_in[lo] | `011`  | rd (scalar) | custom-0  |
| `QCOMM`  | `0000100`   | qw_b[lo]   | qw_a[lo]  | `100`  | qw_out[lo]  | custom-0  |
| `QRING`  | `0000101`   | (unused)   | qw_in[lo] | `101`  | rd (scalar) | custom-0  |

(QW128 is two-register on RV64; the high half is delivered via paired-register convention. funct7 values are now FROZEN at the values shown.)

### 4.2 VCIX (vector)

Use SiFive's published `sf.vc.*` family with a contiguous block of QBP funct6 allocations:

| QBP op | VCIX form | QBP funct6 | Notes |
|---|---|---|---|
| `QPERM(vd, vs2, idx)`  | `sf.vc.v.iv` | `001000` | imm = LUT index 0..6 (Fano) or 0..15 (extended) |
| `QPERMR(vd, vs2, idx)` | `sf.vc.v.iv` | `001001` | inverse permutation |
| `QPERM.vv(vd, vs1, vs2)` | `sf.vc.v.vv` | `001010` | vector-of-indices form |
| `QFMA(vd, va, vb, vc)` | `sf.vc.v.vv` (3-source variant) | `001011` | vd ← va·vb + vc |
| `QSAND(vd, vq, vp)`    | `sf.vc.v.vv` | `001100` | vd ← vq · vp · vq⁻¹ |
| `QNORM(vd, vs2)`       | `sf.vc.v.v`  | `001101` | vd[i] ← \|vs2[i]\|² (per-lane scalar) |

QBP funct6 block: `001000` through `001101` allocated; `001110`-`001111` reserved.

This is the second-generation vectorized counterpart of the scalar SSCI ops. Both forms must compute byte-identical results given equivalent inputs (verified by the cosim harness, §9.2).

## 5. Pipeline timing model

### 5.1 Host stages (X280, abstracted)

Unchanged from v0.1.

### 5.2 Accelerator stages

```
A_RECV → A_DECODE → A_DISPATCH → [LUT | QFMA_UNIT | WD_UNIT] → A_WD_GATE → A_RESP
                                                    │
                       watchdog tap on every dispatch
```

Per-instruction cycle budgets:

| Op | Dispatch | Compute | Cycles total | Pipelined? |
|---|---|---|---|---|
| QFMA   | 1 | 1-3 | 1-3  | Yes (1 op/cycle steady) |
| QSAND  | 1 | 3-5 | 3-5  | Partially (2 ops/cycle peak) |
| QNORM  | 1 | 1   | 1    | Yes |
| QPERM  | 1 | 1-2 | 1-2  | Yes |
| QPERMR | 1 | 1-2 | 1-2  | Yes |
| QNEAR  | 1 | 1   | 1    | Yes |
| QDEC   | 1 | 3-5 | 3-5  | No (branchy) |
| QREC   | 1 | 3-5 | 3-5  | No (branchy) |
| QCOMM  | 1 | 2   | 2    | Yes |
| QRING  | 1 | 1   | 1    | Yes |

QFMA-dominated workloads (the 95% case per §3 of *Wyrd-Workload-ISA-v0.1.md*) achieve ~1 quat-FMA / cycle / lane at steady state on X280 with 4-wide vector → ~4 quat-FMA/cycle aggregate.

### 5.3 Hazards

- **RAW on vd (vector):** standard X280 vector hazard logic.
- **Algebra fault:** if the watchdog detects an invariant violation exceeding ε_priv, response carries `WD_FAULT`. Software handler decodes via custom CSR (see §8).
- **Cross-ring operation without capability:** if AlgebraID encoded in the op exceeds the running process's permitted set per `qbp_ctl.ALGEBRA_ID`, watchdog event with `RingTransition` flag is forwarded to Skuld for policy decision; the in-flight op completes but its result is gated until Skuld responds (see §10).

## 6. Internal arbitration

- **The algebraic LUT** is a single hardware structure with two read ports (one per ingress: SSCI, VCIX). Concurrent reads; serialized writes (table reload, e.g., for substrate switch). Sized for 16 elements per *Wyrd-Workload-ISA-v0.1.md* §11.2 to accommodate either Cayley-Dickson (8 octonion basis) or Branch A (9 matrix-unit basis).
- **The QFMA unit** is dedicated silicon (16 multipliers + 13 adders per quaternion product, fully-pipelined). One per QBP-CU lane.
- **The algebraic watchdog** has one input event channel per ingress, merged into a single ordered stream by a deterministic round-robin scheduler. Watchdog computes:
  - Commutator [a, b] (when ALGEBRA_ID = 1, ℍ, or higher)
  - Associator (a, b, c) (when ALGEBRA_ID = 2, 𝕆, or higher)
  - Alternator [a, a, b] (when ALGEBRA_ID = 3, 𝕊)
  - All three are computed in parallel; gating logic selects which is *checked* against thresholds based on AlgebraID.

**Why no QASSOC / QALT instructions.** The watchdog already computes these on every cycle and exposes via the new `qbp_invariant` CSR. Software-visible instructions for these would duplicate hardware while running on the slower instruction-issue path. The CSR-read path is faster, unforgeable, and parallel with computation.

## 7. Go simulator interface

Updated `WDEvent` and `Req`/`Resp` structs:

```go
package qbpcu

// Port discriminator
type Port uint8
const (
    PortSSCI Port = iota
    PortVCIX
)

// Opcode for the 10-instruction ISA
type Opcode uint8
const (
    OpQDEC Opcode = iota
    OpQREC
    OpQNEAR
    OpQNORM
    OpQCOMM
    OpQRING
    OpQPERM     // VCIX
    OpQPERMR    // VCIX
    OpQPERM_VV  // VCIX, vector-indices
    OpQFMA      // VCIX
    OpQSAND     // VCIX
    OpQNORM_V   // VCIX
)

// AlgebraID matches qbp_ctl.ALGEBRA_ID encoding
type AlgebraID uint8
const (
    AlgebraC      AlgebraID = 0  // ℂ — user
    AlgebraH      AlgebraID = 1  // ℍ — supervisor
    AlgebraO      AlgebraID = 2  // 𝕆 — kernel
    AlgebraS      AlgebraID = 3  // 𝕊 — firmware
    AlgebraBranchA AlgebraID = 4 // ℂ⊕ℍ⊕M₃(ℂ)
    AlgebraBranchB AlgebraID = 5 // extended Cayley-Dickson
)

// A request from host to accelerator
type Req struct {
    Cycle    uint64
    Port     Port
    Op       Opcode
    VL       uint32      // VCIX only
    SrcA     QW128       // or vector-lane buffer for VCIX
    SrcB     QW128       // for VCIX vv form
    SrcC     QW128       // for QFMA (3-source)
    Imm      uint8       // for VCIX iv form
    DestTag  uint16
}

// A response from accelerator to host
type Resp struct {
    Cycle             uint64
    DestTag           uint16
    Result            QW128
    Status            Status      // OK | WD_FAULT | DECODE_ERROR | CAP_GATED
    FaultCode         uint32
    InvariantSnapshot Invariant   // qbp_invariant CSR snapshot at completion
}

// The watchdog event, tapped at every algebraic crossing
type WDEvent struct {
    Cycle           uint64
    Op              Opcode
    Port            Port
    LUTIndex        uint8       // 0..15 (renamed from FanoIndex for substrate-agnosticism)
    SignBit         bool
    Commutator      [4]int8     // NEW v0.2: residue per axis (re, imI, imJ, imK)
    Associator      [3]int8     // residue per axis (renumbered from v0.1)
    Alternator      [4]int8     // NEW v0.2
    NormDelta       int32
    AlgebraID       AlgebraID
    RingTransition  uint8       // NEW v0.2: 0 = none, else encodes from→to ring transition
}

// The accelerator interface the host model talks to
type Accelerator interface {
    Submit(r Req)
    Poll() (Resp, bool)
    WatchdogChan() <-chan WDEvent
    Tick(cycle uint64)
}

// The Invariant tuple read from qbp_invariant CSR
type Invariant struct {
    Commutator [4]int8
    Associator [3]int8
    Alternator [4]int8
    Cycle      uint64
    AlgebraID  AlgebraID
}
```

The host-side cycle-accurate model treats VCIX/SSCI as functional units with the latency model of §5; the result comes from `Accelerator`. Swapping implementations (Mock, Golden, RTLShim) changes only which `Accelerator` the host model is wired to.

## 8. Custom CSRs (with privilege column)

| CSR                      | Field             | Purpose                          | Min AlgebraID |
|--------------------------|-------------------|----------------------------------|---------------|
| `qbp_status` (0xBC0)     | `WD_LAST_FAULT[15:0]` | Last watchdog fault code     | 1 (ℍ)         |
| `qbp_status` (0xBC0)     | `PORT[17:16]`     | Which port raised the fault      | 1 (ℍ)         |
| `qbp_status` (0xBC0)     | `RING_TRANSITION[19:18]` | Last ring transition       | 1 (ℍ)         |
| `qbp_ctl` (0xBC1)        | `WD_ENABLE[0]`    | Master watchdog enable           | 1 (ℍ)         |
| `qbp_ctl` (0xBC1)        | `ALGEBRA_ID[3:1]` | Active algebra (Wyrd ring)       | 1 (ℍ)         |
| `qbp_ctl` (0xBC1)        | `FP_MODE[4]`      | 0=fp32, 1=fp16 (QFMA precision)  | 1 (ℍ)         |
| `qbp_invariant` (0xBC2)  | `COMMUTATOR[31:0]` | Latest commutator (4×int8)     | 1 (ℍ)         |
| `qbp_invariant` (0xBC2)  | `ASSOCIATOR[55:32]` | Latest associator (3×int8)    | 1 (ℍ)         |
| `qbp_invariant` (0xBC2)  | `ALTERNATOR[87:56]` | Latest alternator (4×int8)    | 1 (ℍ)         |
| `qbp_invariant` (0xBC2)  | `RING_ID[95:88]`   | Last operation's ring          | 0 (ℂ — readable by user) |
| `qbp_invariant` (0xBC2)  | `CYCLE[127:96]`    | Cycle at which residues captured | 0 (ℂ)        |

`qbp_invariant` is the headline new CSR. Its existence is what eliminates the need for QASSOC and QALT software-visible instructions — software reads the CSR after any QBP-CU op and gets the most recent algebraic invariant residues.

The "Min AlgebraID" column makes the privilege model concrete: a user process (ALGEBRA_ID = 0) attempting to read `qbp_status` raises an illegal-instruction trap. The exception is `qbp_invariant.RING_ID` and `qbp_invariant.CYCLE`, which any process may read — these are diagnostic, not privileged.

Per the QBP standing rule: `WD_ENABLE` is reset-to-1 with sticky bit gate clearable only by the test harness AND only at AlgebraID = 3 (𝕊, firmware mode). Production silicon cannot reach 𝕊 except at boot; effectively the watchdog is permanently enabled.

## 9. Validation (revised test corpus)

### 9.1 Golden model contract

Unchanged from v0.1.

### 9.2 Cosim harness

Unchanged from v0.1.

### 9.3 Test corpus tiers

**Tier 0 — algebraic identities.**
- Quaternion norm preservation under QFMA
- Quaternion conjugate-via-sandwich is identity for unit quaternions
- Octonion alternativity via LUT
- Fano-plane Moufang identities (when LUT is Fano-populated)
- M₃(ℂ) associativity (when LUT is matrix-units populated; Branch A only)
- **NEW:** Sandwich preservation: |QSAND(q, p)| = |p| for unit q
- **NEW:** Commutator antisymmetry: QCOMM(a, b) = -QCOMM(b, a)

**Tier 1 — single-instruction microbenchmarks.**
- 10⁶ random QFMAs against software gold reference
- 10⁶ random QSANDs
- 10⁶ random QNORMs
- 10⁶ random QPERMs (each LUT mode)
- 10⁶ random QCOMMs

**Tier 2 — kernel-level (workload-specific).**
- BMA hypergraph traversal MAC (mostly QFMA)
- Hammer dynamics step (QFMA + QSAND + QNORM mix)
- **NEW:** Squam Lake sub-region (1km × 1km × 10m) thermocline evolution, 1 minute simulated time
- **NEW:** GRB-style cross-correlation, 10⁵ samples × 10³ offsets
- **NEW:** Materia-Bio amino-acid Glycine validation single ensemble

**Tier 3 — system-level.**
- Boot Linux on host + run Crawl.Heartbeat (existing)
- **NEW:** End-to-end prediction-verification: theorist queries QBP-Spec → CKM matrix prediction → comparison to PDG values, all in single workflow
- **NEW:** GRB 250702B follow-up: ingest VLA radio data → QBP correlation → emit predicted/observed comparison
- **NEW:** Skuld supervisor enforces privilege correctly under adversarial load — process A cannot read process B's quaternion state without capability

Pre-tape-out gate: Tiers 0–2 at 0 divergences across 10⁹+ random inputs PLUS Tier 3 prediction-verification under wall-clock targets per *Wyrd-Workload-ISA-v0.1.md* §5.2.

## 10. Memory model and event flow

### 10.1 Cache coherence and ordering

Unchanged from v0.1.

### 10.2 Watchdog event flow (REVISED)

The host-side architecture stack (replacing v0.1 §10's BMA-direct flow):

```
QBP-CU watchdog hardware
       ↓ (raw events via TileLink master port to host-pinned ring buffer)
Skuld supervisor (Wyrd-Supervisor-Architecture-v0.2.md)
   ├── Privilege filter: events with |residue| > ε_priv but < ε_phys
   ├── Capability check: does running process hold required capability?
   ├── Ring transition: write new ring to qbp_ctl.ALGEBRA_ID
   └── Fault generation: raise illegal-instruction on policy violation
       ↓ (post-policy events: privilege-OK and physical seams)
Wyrd database
   ├── Persistent storage of physical seams as hypergraph nodes
   └── Subscription mechanism for downstream consumers
       ↓
BMA judge collective
   └── Epistemic monitoring of post-policy event stream
```

This is a meaningful change from v0.1: BMA does not consume raw watchdog events. Skuld sits in between, applies privilege policy, and passes a filtered annotated stream onward.

Back-pressure stalls the accelerator, preferable to dropping events. Skuld is responsible for keeping up; if the policy filter is too slow, Skuld is the bottleneck and must be optimized.

## 11. Open questions (revised)

1. ~~**funct7 / funct6 freeze.**~~ **DONE** in §4. funct7 values 0000000–0000101 allocated; funct6 values 001000–001101 allocated.
2. **LUT capacity.** Provisioned 16 elements per substrate-agnostic recommendation (v0.1's 8-element Fano fits; Branch A's 9 matrix-units fits; Branch B's extension up to 16 fits). Cost: ~2× LUT silicon area; acceptable.
3. **VCIX VL semantics for QPERM/QFMA/QSAND.** When VL is not a multiple of 4 (not whole QW128s per register), zero-fill (recommendation (a) from v0.1).
4. **Chain-of-trust on `WD_ENABLE` clear.** §8 specifies AlgebraID = 3 (𝕊) requirement. Concrete BMC-visible signal definition still needed before silicon.
5. **Compiler intrinsics.** **DECISION:** option (b) — Clang/LLVM `__builtin_qbp_*` intrinsics in `Xsf*` vendor namespace — for v1. MLIR dialect deferred to v2 (Phase 4+).
6. **SiFive licensing path.** Same as v0.1; recommend in-house against published `sf.vc.*` interface for Phases 1-3, SiFive engagement at Phase 4.
7. **NEW:** **fp16 QFMA fallback.** `qbp_ctl.FP_MODE` flag enables fp16 for batch/throughput workloads. Open: does fp16 push the noise floor too close to ε_priv to be useful? Resolution: per-workload empirical, not architectural decision now.
8. **NEW:** **Substrate selection.** Currently Cayley-Dickson default per *Wyrd-BranchA-Contingency-v0.1.md* H3+H1 hedge. LUT and watchdog are configurable; selection happens at boot via `qbp_ctl.ALGEBRA_ID` initialization. Default: AlgebraID = 0 (ℂ), with Skuld responsible for migrating to higher rings on capability presentation.

## 12. Forward path (combined with Skuld phases)

| Phase | SiFive milestone | Skuld milestone | Wyrd milestone |
|---|---|---|---|
| 0 | Spec freeze (this doc) | Skuld-Spec.md draft | Wyrd-Spec.md update |
| 1 | Go simulator implementation | `skuld` package | `wyrd` = renamed `hg` |
| 2 | Host pipeline model | Walk: quaternion-native | Walk: Wyrd quaternion rewrite |
| 3 | VexRiscv FPGA | Run: federation | Run: distributed Wyrd via NATS |
| 4 | SiFive engagement | Sprint: collapse | Sprint: CIM hardware |

This unifies the SiFive Phase 0–4 with the Wyrd / Skuld Crawl–Sprint cadence into a single timeline.

---

## References

- *Wyrd-Supervisor-Architecture-v0.2.md* — the supervisor architecture
- *Wyrd-Workload-ISA-v0.1.md* — workload analysis driving the 10-instruction ISA
- *Wyrd-SiFive-Spec-Review-v0.1.md* — the review document that motivated this v0.2
- *Wyrd-BranchA-Contingency-v0.1.md* — substrate-uncertainty hedge
- *Skuld-Spec-v1.0.md* — the Crawl-phase API specification
- SiFive Vector Coprocessor Interface (VCIX) Software Specification v1.0
- LLVM RISC-V User Guide §SiFive Vendor Extensions
- MLIR `vcix` Dialect documentation

**Attribution (per QBP standing rule):** Furey, Günaydin/Gürsey, Dixon, Boyle/Farnsworth, Singh, Chamseddine/Connes, Koide, Baez. SiFive (Asanovic et al.) for VCIX and the X280 reference architecture.

---

*End of QBP-CU SiFive Interface Specification v0.2 — DRAFT*
