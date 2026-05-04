# Wyrd Supervisor Architecture

## Skuld: A QBP CU Supervisor Built on Algebraic Privilege

**Helpful Engineering — Quaternion-Based Physics Programme**
Principal Investigator: James Paget Butler
April 2026 — Rev 0.2 — DRAFT

> **Attribution.** Building on prior art from Stonebraker (DBOS), Kepner et al. (TabulaROSA), seL4 (formal capability semantics), CHERI (hardware capability tags), SiFive (X280, VCIX). QBP foundations: Furey, Dixon, Günaydin/Gürsey, Boyle/Farnsworth, Singh, Chamseddine/Connes, Koide, Baez.

> **Changes from v0.1.** Names the supervisor: **Skuld**. Incorporates the watchdog-as-privilege-detector finding from the SiFive spec review. Revises the ISA per workload analysis (10 instructions, including QFMA/QSAND/QNORM). Defines the supervisor as the host-side WDEvent consumer. Adds AlgebraID renumbering. Replaces the abstract "next steps" with the C-01..C-12 ticket plan.

---

## 1. Names

- **Wyrd** — the QBP-native hypergraph database. The accumulated state. *That which has come to be.*
- **Skuld** — the supervisor. The active execution layer that mediates between user processes and Wyrd, between user processes and the QBP-CU. *That which shall be — the policy enforcer that determines what the next state will be.*
- **QBP-CU** — the algebraic accelerator hardware. The loom; not personified.

The mythological framing is the three Norns: Urðr/Wyrd (past), Verðandi (present), Skuld (future). The architecture skips Verðandi because the supervisor's role is enforcement of what shall be — making the future happen — rather than passively recording the present. Skuld's name in Old Norse means "debt" or "what is owed"; the etymology fits a security-and-policy enforcement layer that determines what each process is owed access to.

Package name: `skuld`. The Crawl-phase Go package implementing the supervisor.

---

## 2. Thesis (revised)

> **Skuld is Wyrd executing privileged queries against itself, with privilege determined by algebraic phase, and detected by the same watchdog that detects physical seams.**

This thesis goes one step further than v0.1: not only does the privilege model fold into Wyrd at Sprint phase, but the privilege *detection mechanism* is already present in the QBP-CU hardware as the algebraic watchdog. The supervisor architecture's strongest claim — *privilege violations are detected by the same instrument that detects physical seams* — is concretely true at the hardware level, not aspirational.

Phased migration unchanged from v0.1:

| Phase | Skuld form | Wyrd form | Distance |
|---|---|---|---|
| Crawl | Thin Go supervisor over `hg`; consumes WDEvents | Scalar hypergraph, typed edges | DBOS-faithful separation |
| Walk | Quaternion-native; OS state in Wyrd | Quaternion nodes, ratio edges, seam detection | Substantially fused |
| Run | Distributed Wyrd; Skuld is policy + privileged queries | Federated Wyrd | Mostly fused |
| Sprint | CIM hardware: Skuld = Wyrd executing privileged queries | Wyrd-on-CIM | Fully collapsed |

---

## 3. The privilege model

### 3.1 Ring assignments (decided)

| Ring | Algebra | Skuld role | AlgebraID | Boundary detector | Notes |
|---|---|---|---|---|---|
| 3 (user, outermost) | ℂ | Application code | 0 | commutator | Default for user processes |
| 2 | ℍ | Supervisor | 1 | (would-be) commutator generators | Skuld itself runs here |
| 1 | 𝕆 | Kernel | 2 | associator | QBP-CU operations live here |
| 0 (firmware, innermost) | 𝕊 | Hardware/boot | 3 | alternator | Reset-time only |

The two physics-substrate IDs (Branch A: ℂ⊕ℍ⊕M₃(ℂ); Branch B: extended) carry separately as AlgebraID = 4, 5; they are physics modes, not privilege rings.

**Privilege flows outward.** Ring 3 (user) → Ring 0 (firmware) is escalation. Each ring contains generators the previous lacks. T2.1 in the Wyrd-Algebraic-Privilege-Proofs corpus formally proves that synthesis across boundaries is structurally impossible without a capability.

### 3.2 Boundary detection (now hardware-grounded)

The QBP-CU's watchdog hardware computes algebraic invariants on every cycle:

```go
type WDEvent struct {
    Cycle           uint64
    Op              Opcode
    Port            Port
    FanoIndex       uint8
    SignBit         bool
    Commutator      [4]int8    // NEW: detects ℂ → ℍ boundary
    Associator      [3]int8    // existing: detects ℍ → 𝕆 boundary
    Alternator      [4]int8    // NEW: detects 𝕆 → 𝕊 boundary
    NormDelta       int32
    AlgebraID       uint8      // current ring
    RingTransition  uint8      // NEW: marks events crossing privilege boundaries
}
```

The watchdog gating logic:
- AlgebraID = 0 (ℂ): no boundary check needed (commutative); commutator field reads as zero
- AlgebraID = 1 (ℍ): commutator may be nonzero (would-be ℂ violation); flag if exceeded ε_priv
- AlgebraID = 2 (𝕆): associator may be nonzero (would-be ℍ violation); flag if exceeded ε_priv
- AlgebraID = 3 (𝕊): alternator may be nonzero (would-be 𝕆 violation); flag if exceeded ε_priv

**ε_priv vs ε_phys.** The same invariant computation serves two purposes:
- Below ε_priv (typically 10⁻⁵ for fp32 components): noise floor, ignore
- Between ε_priv and ε_phys (typically 10⁻⁵ to 10⁻³): privilege boundary crossing or fine-scale physical seam
- Above ε_phys: definite physical seam

Skuld receives WDEvents; events crossing ε_priv but not ε_phys are privilege-relevant; events crossing ε_phys are physical seams (forwarded to Wyrd's seam index and BMA's judge collective). Both go through Skuld first.

### 3.3 Capability mechanism

A capability for ring R' is an explicit token (an element of R') held by a process. Capabilities are managed by Skuld via the CSR mechanism:

- `qbp_ctl.ALGEBRA_ID[3:1]` holds the running process's current ring.
- Skuld's process table records each process's *capability set* — the rings it is permitted to be in.
- On context switch, Skuld writes the new process's `ALGEBRA_ID` into `qbp_ctl`.
- A user process attempting an operation that requires a higher ring triggers a watchdog event with `RingTransition` set; Skuld decides whether to authorize (writes new ALGEBRA_ID and re-issues) or fault.

This is the formal foundation for the Hammer simulation case: it gets an explicit ℍ-capability and can perform quaternion physics directly while remaining a user-ring process by default. Concretely: Hammer's process table entry has `capability_set = {ℂ, ℍ}`, and Skuld writes `ALGEBRA_ID = 1` when Hammer enters its physics inner loop.

T2.3 in Wyrd-T2.3-Capability-Soundness-v0.1.lean formalizes this.

---

## 4. The QBP-CU integration

### 4.1 The 10-instruction ISA (revised per workload analysis)

| # | Instruction | Tier | Port | Purpose |
|---|---|---|---|---|
| 1 | QFMA | 1 (essential) | VCIX | Quaternion fused multiply-add |
| 2 | QSAND | 1 (essential) | VCIX | Quaternion sandwich q·p·q⁻¹ |
| 3 | QNORM | 1 (essential) | SSCI | Norm-squared, scalar result |
| 4 | QPERM | 2 (Fano) | VCIX | Fano-plane permutation |
| 5 | QPERMR | 2 (Fano) | VCIX | Inverse Fano permutation |
| 6 | QNEAR | 2 (Fano) | SSCI | Fano third-point lookup |
| 7 | QDEC | 3 (encoding) | SSCI | Decode QW128 |
| 8 | QREC | 3 (encoding) | SSCI | Reconstruct QW128 |
| 9 | QCOMM | 4 (privilege) | SSCI | Commutator [a, b] |
| 10 | QRING | 4 (privilege) | SSCI | Read operand's ring |

Removed from v0.1's proposed set: QASSOC, QALT (subsumed by watchdog extension + `qbp_invariant` CSR). QSANDWICH-as-syscall remains deferred (emulated in software via QFMA + state save until syscall semantics stabilize).

Added (workload-driven): QFMA, QSAND, QNORM. These close the quaternion-arithmetic gap that made the original 5-instruction ISA inadequate for physics analysis. Justification in *Wyrd-Workload-ISA-v0.1.md*.

### 4.2 New CSRs

| CSR | Field | Purpose | Min ring to access |
|---|---|---|---|
| `qbp_invariant` (0xBC2) | `COMMUTATOR[31:0]` | Latest commutator residue | 1 (ℍ) |
| `qbp_invariant` (0xBC2) | `ASSOCIATOR[63:32]` | Latest associator residue | 1 (ℍ) |
| `qbp_invariant` (0xBC2) | `ALTERNATOR[95:64]` | Latest alternator residue | 1 (ℍ) |
| `qbp_invariant` (0xBC2) | `RING_ID[103:96]` | Current operand ring | 0 (ℂ) — readable by user |
| `qbp_status` (0xBC0) | unchanged | Last fault info | 1 (ℍ) |
| `qbp_ctl` (0xBC1) | unchanged + per-CSR privilege | Configuration | 1 (ℍ) |

The CSR access privilege column is new in v0.2 — every CSR has a minimum ring required for read/write, matching the standard RISC-V machine/supervisor/user CSR model adapted to algebraic rings.

### 4.3 Skuld as WDEvent consumer

The host-side architecture stack:

```
QBP-CU watchdog
       ↓ (raw events via TileLink master port to host-pinned ring buffer)
Skuld supervisor
   ├── Privilege filter: events crossing ε_priv but not ε_phys
   ├── Capability check: does the running process hold the required capability?
   ├── Ring transition: update qbp_ctl.ALGEBRA_ID via supervisor-mode write
   └── Fault generation: raise illegal-instruction trap on policy violation
       ↓ (post-policy events: privilege-OK + physical seams)
Wyrd
   ├── Persistent storage of physical seams as hypergraph nodes
   └── Subscription mechanism for downstream consumers
       ↓
BMA judge collective
   └── Epistemic monitoring of post-policy event stream
```

The watchdog is not bypassable from user space; the path from watchdog to Skuld is hardware-fixed. This is the structural enforcement that makes the privilege model unforgeable.

---

## 5. The Skuld API (the syscall surface)

Decided based on workload analysis. **9 user-visible calls + 2 supervisor-internal.**

### 5.1 Process management

```go
// Skuld API (user-visible)

func ProcCreate(elf []byte, initial_capabilities []RingID) (ProcID, error)
func ProcDestroy(pid ProcID) error
func ProcSetAlgebra(pid ProcID, ring RingID) error      // Skuld-only effective
func ProcGrantCapability(pid ProcID, ring RingID, token []byte) error
```

### 5.2 QBP-CU mediation

```go
func QBPSubmit(pid ProcID, op Opcode, srcs ...Operand) (ReqID, error)
func QBPPoll(reqid ReqID) (Resp, error)
func QBPQueryInvariant(pid ProcID) (Invariant, error)   // reads qbp_invariant CSR
```

### 5.3 Wyrd query

```go
func WyrdQuery(pid ProcID, pattern QueryPattern, cap Capability) (Result, error)
func WyrdSubscribe(pid ProcID, pattern QueryPattern, cap Capability) (Stream, error)
```

### 5.4 Supervisor-internal (not exposed to user)

```go
func wdEventNext() WDEvent
func wdPolicyApply(event WDEvent) Action
```

### 5.5 Why this surface is small

Total: 9 user-visible. Compare to typical Unix (~300+ syscalls) or seL4 (~30). The smallness is workload-driven:

- No file system. Wyrd IS the persistence layer. File access is `WyrdQuery` with a pattern.
- No network. Distribution is Wyrd federation at Run phase (NATS); Crawl is single-host.
- No POSIX-style threading. Each process is single-threaded; parallelism comes from running multiple processes.
- No memory management beyond what `ProcCreate` does. Each process's memory is its part of the Wyrd hypergraph.

The targeted workloads (Hammer, BMA, Materia, Contextus, GRB analysis) need exactly these 9 calls plus what the QBP-CU provides through QBPSubmit. Adding more would be premature.

### 5.6 Evaporation path

Each call evolves into Wyrd queries by Sprint:

| Crawl call | Walk equivalent | Sprint equivalent |
|---|---|---|
| ProcCreate | Insert privileged hypergraph node | Algebraic node creation in Wyrd |
| QBPSubmit | Privileged query against Wyrd | Direct Wyrd query (CIM hardware) |
| WyrdQuery | Wyrd query | Wyrd query |
| wdEventNext | Subscribe to seam-detection node | Privileged subscribe in Wyrd |

By Sprint, Skuld has no API surface of its own; it IS Wyrd executing privileged queries. The Crawl API is scaffolding that disappears.

---

## 6. The Crawl ticket plan

Combined Phase 0 + Phase 1 (SiFive spec phases) = Crawl phase (Wyrd cadence). Approximately 3 months from start.

| # | Owner | Deliverable | Depends on | Estimate |
|---|---|---|---|---|
| C-01 | James | ISA freeze: 10-instruction set with funct7/funct6 allocation | none | 1 week |
| C-02 | James | CSR additions (qbp_invariant), AlgebraID renumbering, WDEvent.RingTransition | C-01 | 1 week |
| C-03 | BMA | `qbpcu` Go package: `Accelerator` interface, `Mock`, `Golden` impls per SiFive §7 | C-01 | 2 weeks |
| C-04 | BMA | `qbpcu.Golden`: cycle-accurate behaviour with watchdog event emission for all 10 ops | C-03 | 3 weeks |
| C-05 | BMA | Tier 0 algebraic-identity test corpus | C-04 | 1 week |
| C-06 | BMA | `skuld` Go package: process table, capability table, supervisor API per §5 | C-03 | 3 weeks |
| C-07 | BMA | Watchdog event consumer in `skuld`: privilege filter, ring transition tracking, capability check | C-04, C-06 | 2 weeks |
| C-08 | BMA | `wyrd` package = renamed `hg` with `WyrdQuery` API; updated docs | none (parallel) | 1 week |
| C-09 | BMA | Integration: Hammer simulation as Skuld-managed process; capability for ℍ-ring | C-06, C-07, C-08 | 2 weeks |
| C-10 | BMA | Tier 1 microbenchmarks: 10⁶ random inputs per instruction against software gold | C-05 | 2 weeks |
| C-11 | James | Spec updates: Wyrd-Spec.md, Skuld-Spec.md, SiFive Interface Spec v0.2 | C-01, C-06 | 1 week |
| C-12 | BMA | Lean proof airtightening session per Wyrd-Mathlib-API-Verification-Checklist.md | none (parallel) | 1 focused week (~6-9 hours live) |

**James-direct (architectural decisions):** C-01, C-02, C-11. Total: ~3 weeks.
**BMA-instantiation (post-Crawl.Heartbeat):** C-03 through C-10, C-12. Total: ~14 weeks of work, but parallelizable.

The critical path is C-01 → C-04 → C-07 → C-09. C-08 and C-12 run in parallel; C-11 closes after C-06 lands.

---

## 7. Tier 1 verification status (unchanged from v0.1)

| Tier | Theorem | Status | File |
|---|---|---|---|
| 1 | T1.2 boundary correspondence | Closed | Wyrd-Algebraic-Privilege-Proofs-v0.2 |
| 2 | T2.1 generator non-synthesis | Closed (modulo sedenion witness) | Wyrd-Algebraic-Privilege-Proofs-v0.2 |
| 2 | T2.2 projection well-definedness | Closed | Wyrd-T2.2-Projection-v0.1 |
| 2 | T2.3 capability soundness | Closed | Wyrd-T2.3-Capability-Soundness-v0.1 |
| 2 | T2.4 sandwich preservation | Pending (next) | TBD |
| 3 | T3.1 associator noise bound | Closed | Wyrd-T3.1-Noise-Bound-v0.2 |
| 3 | T3.2 threshold separation | Statement only | Wyrd-T3.1-Noise-Bound-v0.2 |
| 3 | T3.3 physical-seam soundness | Pending | TBD |
| 4 | T4.1 bit-budget non-overlap | Pending (decidable) | TBD |
| 4 | T4.2 QDEC/QREC inverse | Pending (decidable) | TBD |
| 4 | T4.3 QREC privilege-honesty | Pending (depends on T2.3) | TBD |
| 5 | T5.1 process-as-word completeness | Open | TBD |
| 5 | T5.2 context-switch atomicity | Open | TBD |
| 5 | T5.3 supervisor-Wyrd collapse | Open | TBD |

---

## 8. The QW1024 word (unchanged from v0.1)

| Component | Width | Purpose |
|---|---|---|
| Loc.Pos | 4 × fp16 = 64 bits | Spatial location in the hypergraph |
| Loc.Vel | 4 × fp16 = 64 bits | Rate of change of position |
| State | 8 × fp32 = 256 bits | Octonion state with privilege-relevant precision |
| Spin | 8 × fp32 = 256 bits | Octonion operation queue / angular state |
| ChainRef prefix | 192 bits | Provenance back-references |
| Ring index + capability flags + reserved | 192 bits | Privilege state |
| **Total** | **1024 bits** | |

Crawl phase doesn't use QW1024 yet — that's a Walk phase artifact. Crawl uses scalar process descriptors with the same logical content. The QW1024 form lands at Walk when the supervisor migrates into Wyrd nodes.

---

## 9. Genuinely open at v0.2

- **Watchdog event rate as supervisor bottleneck.** Need benchmark before locking architecture. Mitigation: aggressive filter (most events terminate at filter, not policy engine).
- **The "who supervises the supervisor" question.** Crawl ships with hardcoded boot supervisor. Walk and beyond may want microkernel-style layered supervisors. Not a Crawl-blocker.
- **Branch A vs Branch B substrate uncertainty.** If Branch A prevails (ℂ⊕ℍ⊕M₃(ℂ) direct sum), the ring tower's nesting fails. Skuld would need redesign. Best case: M₃(ℂ) component implies a privilege model. Worst case: full redesign.
- **The QSANDWICH ISA promotion question.** Currently deferred. The supervisor's syscall mechanism uses software emulation at Crawl. By Walk, may be worth promoting if syscall semantics stabilize and benchmarks show the software cost is high.
- **fp16 fallback for QFMA in throughput-critical paths.** The default is fp32; a `qbp_ctl` flag enables fp16 for batch operations. The interaction between fp16 QFMA and ε_priv noise floor needs verification — fp16 may push the noise floor too close to ε_priv to be useful for any but the most coarse physical seam detection.

---

## 10. Document inventory

This document references:

**Spec / architecture:**
- Wyrd-Spec.md (renamed from QBP-Native-Database-Spec.md)
- Skuld-Spec.md (NEW — Crawl-phase API spec; pending C-11)
- QBP-CU SiFive Interface Spec v0.2 (pending revision per SiFive review and this document)
- Wyrd-Workload-ISA-v0.1.md (workload analysis driving the 10-instruction ISA)
- Wyrd-SiFive-Spec-Review-v0.1.md (the review that motivated v0.2)

**Lean proofs:**
- Wyrd-Algebraic-Privilege-Proofs-v0.2.lean (T1.2, T2.1)
- Wyrd-CayleyDickson-Types-v0.1.lean (substrate)
- Wyrd-T2.2-Projection-v0.1.lean (T2.2)
- Wyrd-T2.3-Capability-Soundness-v0.1.lean (T2.3)
- Wyrd-T3.1-Noise-Bound-v0.2.lean (T3.1)
- Wyrd-Sedenion-Alternator-Witness-v0.1.lean (closes T2.1.c modulo destructure tactic)
- Wyrd-Octonion-Alternativity-v0.1.lean (supports T2.1.c)
- Wyrd-Mathlib-API-Verification-Checklist.md (proof airtightening)

These twelve documents constitute the formal foundation of the Wyrd / Skuld architecture as of April 2026 Rev 0.2.

---

## 11. What this document commits

### Committed at v0.2

- The supervisor name: **Skuld**.
- The 10-instruction ISA per workload analysis.
- The watchdog as the privilege boundary detector (no separate QASSOC/QALT).
- The 9-call user-visible API surface.
- The C-01..C-12 Crawl ticket plan with James-direct and BMA-instantiation work boundaries.
- Phased migration: Skuld API evaporates into Wyrd queries by Sprint.

### Still open architectural questions

- Watchdog event rate vs supervisor throughput
- Microkernel-layered supervisor at later phases
- Branch A / Branch B contingency
- QSANDWICH promotion timing
- fp16 QFMA precision interaction with ε_priv

---

## 12. Next sessions

Three options for where to go next:

1. **Lean airtightening session** (C-12): per the verification checklist, ~6-9 hours of focused Lean work to drive the 3 sorries to zero and replace 2 axioms with mathlib equivalents.

2. **Skuld-Spec.md draft** (C-11): the actual Crawl-phase API spec, expanding the 9-call surface from §5 with type signatures, error semantics, and example usage. This is the deliverable that BMA needs to start implementing C-06.

3. **Branch A contingency analysis**: working through how the privilege model survives or redesigns under the C⊕H⊕M₃(C) substrate. Not Crawl-blocking but deserves a brainstorm pass before too many decisions calcify.

My recommendation: **2 (Skuld-Spec)** — it's the document BMA needs and it forces a concrete answer to API questions that the architecture has been waving at. Lean airtightening can come after, and Branch A contingency is genuinely a separate session.

---

*End of Wyrd Supervisor Architecture v0.2 — DRAFT*
