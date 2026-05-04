# QBP-CU ISA Freeze (C-01) and CSR Additions (C-02)

## Binding Decision Document

**Helpful Engineering — Quaternion-Based Physics Programme**
Principal Investigator: James Paget Butler
April 2026 — Rev 1.0 — DECIDED

> **Purpose.** This document is the binding output of tickets C-01 (ISA freeze) and C-02 (CSR additions, AlgebraID renumbering, WDEvent extension) per *Wyrd-Supervisor-Architecture-v0.2.md* §6. It supersedes the §11.1 placeholders in *QBP-CU SiFive Interface Specification v0.1*. All subsequent work — Go simulator (C-03 onward), VexRiscv FPGA, SiFive engagement — uses these encodings.

> **Status.** DECIDED. Changes after this document is committed require a versioned amendment and explicit migration of all consumers.

---

## 1. The 10-instruction ISA — final encoding

### 1.1 SSCI (scalar) instructions — custom-0 opcode `0001011`

| Mnemonic | funct7    | funct3 | Operands                  | Result    | Purpose |
|----------|-----------|--------|---------------------------|-----------|---------|
| `QDEC`   | `0000000` | `000`  | rs1, rs2 (paired QW)      | rd (lo)   | Decode QW128 to canonical form |
| `QREC`   | `0000001` | `001`  | rs1, rs2 (paired QW)      | rd (lo)   | Reconstruct QW128 |
| `QNEAR`  | `0000010` | `010`  | rs1, rs2 (idx)            | rd (lo)   | LUT third-element lookup |
| `QNORM`  | `0000011` | `011`  | rs1, rs2 (paired QW)      | rd (scalar) | Norm-squared |
| `QCOMM`  | `0000100` | `100`  | rs1, rs2 (paired QW)      | rd (lo)   | Commutator [a, b] |
| `QRING`  | `0000101` | `101`  | rs1                       | rd (scalar AlgebraID) | Read operand's ring |

Reserved funct7 values: `0000110`, `0000111` (room for two future scalar additions).

### 1.2 VCIX (vector) instructions — funct6 block `001000`–`001111`

| Mnemonic        | VCIX form     | funct6     | Operands                   | Result | Purpose |
|-----------------|---------------|------------|----------------------------|--------|---------|
| `QPERM`         | `sf.vc.v.iv`  | `001000`   | vs2, imm5 (LUT index)      | vd     | LUT permutation |
| `QPERMR`        | `sf.vc.v.iv`  | `001001`   | vs2, imm5                  | vd     | Inverse LUT permutation |
| `QPERM.vv`      | `sf.vc.v.vv`  | `001010`   | vs1, vs2 (vector indices)  | vd     | Vector-of-indices permutation |
| `QFMA`          | `sf.vc.v.vv`  | `001011`   | va, vb, vc                 | vd     | Fused multiply-add |
| `QSAND`         | `sf.vc.v.vv`  | `001100`   | vq, vp                     | vd     | Sandwich q·p·q⁻¹ |
| `QNORM.v`       | `sf.vc.v.v`   | `001101`   | vs2                        | vd     | Per-lane norm-squared |

Reserved funct6 values: `001110`, `001111` (room for two future vector additions, e.g., a future QSANDWICH-as-syscall instruction at Walk phase if syscall semantics stabilize).

### 1.3 Why the order

funct7 / funct6 values are not arbitrary. They are ordered by frequency-of-use × privilege-level inverse, so the most-used / lowest-privilege instructions get the smallest encodings:

- SSCI: encoding-housekeeping (QDEC, QREC) → user-callable arithmetic (QNEAR, QNORM, QCOMM) → privilege check (QRING)
- VCIX: LUT (QPERM family) → quaternion arithmetic (QFMA, QSAND, QNORM.v)

The compiler can use this ordering when assigning instruction-stream priorities; it also produces visually clean disassembly for high-frequency code paths.

---

## 2. AlgebraID renumbering — final assignments

| AlgebraID | Algebra | Wyrd ring | Skuld role | Boundary detector |
|-----------|---------|-----------|------------|-------------------|
| `000` (0) | ℂ       | 3 (user, outermost) | Application | commutator |
| `001` (1) | ℍ       | 2 (supervisor) | Skuld itself | (commutator from below) |
| `010` (2) | 𝕆       | 1 (kernel) | QBP-CU operations | associator |
| `011` (3) | 𝕊       | 0 (firmware, innermost) | Hardware/boot | alternator |
| `100` (4) | ℂ⊕ℍ⊕M₃(ℂ) | (special) | Branch A physics mode | trace |
| `101` (5) | extended Cayley-Dickson | (special) | Branch B physics mode | TBD |
| `110` (6) | reserved | — | — | — |
| `111` (7) | reserved | — | — | — |

The 3-bit field `qbp_ctl.ALGEBRA_ID[3:1]` accommodates this with two reserved values for future expansion.

**Default initialization:** at hardware reset, `ALGEBRA_ID = 011` (𝕊, firmware) for boot. Skuld transitions to `001` (ℍ) immediately on initialization, then per-process to `000` (ℂ) when dispatching user processes. User processes can request `001` only with a Skuld-mediated capability presentation.

---

## 3. CSR final layout

### 3.1 `qbp_status` (CSR address `0xBC0`)

| Bits     | Field             | R/W | Purpose                                  | Min AlgebraID |
|----------|-------------------|-----|------------------------------------------|---------------|
| `[15:0]` | `WD_LAST_FAULT`   | R   | Last watchdog fault code                 | 1 (ℍ)         |
| `[17:16]` | `PORT`           | R   | 0=SSCI, 1=VCIX (which port faulted)      | 1 (ℍ)         |
| `[19:18]` | `RING_TRANSITION` | R  | 00=none, 01=ℂ→ℍ, 10=ℍ→𝕆, 11=𝕆→𝕊         | 1 (ℍ)         |
| `[31:20]` | reserved          | R   | reads as zero                            | —             |
| `[63:32]` | reserved          | R   | reads as zero                            | —             |

### 3.2 `qbp_ctl` (CSR address `0xBC1`)

| Bits     | Field             | R/W | Purpose                                  | Min AlgebraID |
|----------|-------------------|-----|------------------------------------------|---------------|
| `[0]`    | `WD_ENABLE`       | R/W | Master watchdog enable (default 1)       | 3 (𝕊)         |
| `[3:1]`  | `ALGEBRA_ID`      | R/W | Active algebra ring                      | 1 (ℍ) write; 0 (ℂ) read |
| `[4]`    | `FP_MODE`         | R/W | 0=fp32, 1=fp16 (QFMA precision)          | 1 (ℍ)         |
| `[7:5]`  | reserved          | R/W | reads as zero                            | —             |
| `[15:8]` | `EPS_PRIV`        | R/W | ε_priv exponent (signed, 2's complement) | 1 (ℍ)         |
| `[23:16]` | `EPS_PHYS`       | R/W | ε_phys exponent (signed, 2's complement) | 1 (ℍ)         |
| `[63:24]` | reserved          | R/W | reads as zero                            | —             |

`EPS_PRIV` and `EPS_PHYS` are stored as exponents (e.g., `-5` means ε = 10⁻⁵) to keep the CSR field small. Skuld writes these at startup based on `Config.EpsilonPriv` and `Config.EpsilonPhys`.

### 3.3 `qbp_invariant` (CSR address `0xBC2`) — NEW

| Bits        | Field        | R/W | Purpose                              | Min AlgebraID |
|-------------|--------------|-----|--------------------------------------|---------------|
| `[31:0]`    | `COMMUTATOR` | R   | 4 × int8 commutator residues         | 1 (ℍ)         |
| `[55:32]`   | `ASSOCIATOR` | R   | 3 × int8 associator residues         | 1 (ℍ)         |
| `[87:56]`   | `ALTERNATOR` | R   | 4 × int8 alternator residues         | 1 (ℍ)         |
| `[91:88]`   | `RING_ID`    | R   | AlgebraID of operation that produced these | 0 (ℂ — readable by user) |
| `[95:92]`   | reserved     | R   | reads as zero                        | —             |
| `[127:96]`  | `CYCLE`      | R   | Cycle at which residues captured     | 0 (ℂ)         |

**Why mostly ℍ-restricted but with user-readable RING_ID and CYCLE.** A user process can ask "what ring was that operation in?" without privilege; this is diagnostic information needed for adaptive algorithms (e.g., a user process retrying after a precision-driven ring transition). The actual residue values are sensitive — they leak information about supervisor-ring computations the user shouldn't observe.

This is a 128-bit CSR; on RV64 it occupies two MSCRATCH-style register pairs. The decode logic accepts paired CSR reads `csrr rd_lo, qbp_invariant_lo` / `csrr rd_hi, qbp_invariant_hi` with adjacent CSR addresses `0xBC2` and `0xBC3`.

### 3.4 Reserved CSR addresses

`0xBC4` through `0xBCF` reserved for future QBP-CU CSRs (e.g., per-lane statistics, performance counters). Allocations require an amendment to this document.

---

## 4. WDEvent struct — final layout

The `WDEvent` struct format that streams through the host-pinned ring buffer:

```
struct WDEvent {                        // 64 bytes
    u64 cycle;                          // 8B  bytes 0-7
    u8  op;                             // 1B  byte  8     (Opcode enum, 0..11)
    u8  port;                           // 1B  byte  9     (0=SSCI, 1=VCIX)
    u8  lut_index;                      // 1B  byte  10    (0..15)
    u8  sign_bit;                       // 1B  byte  11
    i8  commutator[4];                  // 4B  bytes 12-15  (residues per axis)
    i8  associator[3];                  // 3B  bytes 16-18
    u8  _pad1;                          // 1B  byte  19    (alignment)
    i8  alternator[4];                  // 4B  bytes 20-23
    i32 norm_delta;                     // 4B  bytes 24-27
    u8  algebra_id;                     // 1B  byte  28    (AlgebraID, 0..7)
    u8  ring_transition;                // 1B  byte  29    (0..3, see §3.1)
    u16 _pad2;                          // 2B  bytes 30-31 (alignment)
    u64 dest_tag;                       // 8B  bytes 32-39 (matches Req.DestTag)
    u8  _pad3[24];                      // 24B bytes 40-63 (reserved for future fields)
};
```

64 bytes per event, cacheline-aligned. The ring buffer holds N events; with a typical N=4096, the buffer is 256 KiB. Sized to absorb burst rates of >10⁶ events/sec without back-pressure under typical workloads.

The 24 bytes of trailing padding are intentional — they reserve space for additional fields (e.g., second-order residues, user-tag fields) without requiring a struct layout change.

---

## 5. Compiler intrinsics — committed names

### 5.1 LLVM `__builtin_qbp_*` intrinsics

```c
// Tier 1 — quaternion arithmetic (essential)
quat_t   __builtin_qbp_qfma(quat_t a, quat_t b, quat_t c);
quat_t   __builtin_qbp_qsand(quat_t q, quat_t p);
double   __builtin_qbp_qnorm(quat_t q);

// Vector forms (VCIX)
qvec_t   __builtin_qbp_vqfma(qvec_t va, qvec_t vb, qvec_t vc);
qvec_t   __builtin_qbp_vqsand(qvec_t vq, qvec_t vp);
dvec_t   __builtin_qbp_vqnorm(qvec_t vq);

// Tier 2 — algebraic LUT (Fano under Cayley-Dickson, M₃(ℂ) under Branch A)
oct_t    __builtin_qbp_qperm(oct_t v, uint8_t idx);
oct_t    __builtin_qbp_qpermr(oct_t v, uint8_t idx);
oct_t    __builtin_qbp_vqperm(oct_t v, uoctvec_t idx_vec);
uint8_t  __builtin_qbp_qnear(uint8_t i, uint8_t j);

// Tier 3 — encoding
qw128_t  __builtin_qbp_qdec(qw128_t qw);
qw128_t  __builtin_qbp_qrec(qw128_t qw);

// Tier 4 — privilege
quat_t   __builtin_qbp_qcomm(quat_t a, quat_t b);
uint8_t  __builtin_qbp_qring(quat_t a);  // returns AlgebraID
```

### 5.2 Header file location

`<sf_qbp.h>` in the SiFive vendor namespace, following Xsf* convention.

### 5.3 Type definitions

```c
typedef struct { float re, imI, imJ, imK; } quat_t;       // 16 bytes
typedef struct { quat_t lo, hi; } qw128_t;                // 32 bytes
typedef struct { float c[8]; } oct_t;                     // 32 bytes (or M₃(ℂ) under Branch A)
typedef quat_t qvec_t __attribute__((vector_size(64)));   // 4 × quat = vector
typedef double dvec_t __attribute__((vector_size(32)));
typedef oct_t  uoctvec_t __attribute__((vector_size(32)));
```

These are the C-side names. Go bindings follow in the `qbpcu` package per *Wyrd-Workload-ISA-v0.1.md* §6.

---

## 6. Migration table — old encoding to new

For any existing code or specs that reference the v0.1 SiFive spec:

| v0.1 reference | v0.2 (this doc) reference | Action |
|---|---|---|
| `AlgebraID` 0=ℍ | `AlgebraID` 1=ℍ | Renumber |
| `AlgebraID` 1=𝕆 | `AlgebraID` 2=𝕆 | Renumber |
| `AlgebraID` 2=Branch A | `AlgebraID` 4=Branch A | Renumber |
| `AlgebraID` 3=Branch B | `AlgebraID` 5=Branch B | Renumber |
| `WDEvent.FanoIndex` | `WDEvent.lut_index` | Rename (substrate-agnostic) |
| `funct7` "tbd" | as in §1.1 | Use values from §1.1 |
| `funct6` "tbd" | as in §1.2 | Use values from §1.2 |
| 5-instruction ISA | 10-instruction ISA | Add QFMA, QSAND, QNORM, QCOMM, QRING |
| (no `qbp_invariant` CSR) | `qbp_invariant` at 0xBC2 | Add |

Existing test corpus from v0.1 (Tier 0-2) survives unchanged — the algebraic identities and microbenchmarks are all defined per-instruction. The 5 new instructions need new microbenchmark cases per *QBP-CU SiFive Interface Spec v0.2* §9.3.

---

## 7. The "freeze" promise and amendment process

**This document is binding.** The ISA encoding, CSR layout, AlgebraID values, and WDEvent struct are FROZEN. Subsequent work uses these values verbatim.

**To amend:** any change to this document requires:

1. A versioned successor (Rev 1.1, 1.2, ...) with a clear diff against this Rev 1.0.
2. James (PI) explicit sign-off.
3. Migration plan for any existing consumer code (Go simulator, VexRiscv RTL, FPGA bitstream, SiFive engagement materials).
4. Re-running the test corpus against the new encoding.

**Will NOT amend:** trivial clarifications, typo fixes, additions to "Reserved" slots that don't conflict with existing allocations. These can be made in-line with a "Rev 1.0 (clarified)" note.

**Likely to amend:** when Branch A vs Branch B is decided. Branch A would require the AlgebraID 4 slot be activated in the watchdog logic; Branch B activation would require the AlgebraID 5 slot. The encoding accommodates both without re-numbering — only the watchdog hardware changes its computed-invariant logic, gated by AlgebraID.

---

## 8. Sign-off

This document represents the intended binding decisions of tickets C-01 and C-02 per the Crawl ticket plan in *Wyrd-Supervisor-Architecture-v0.2.md* §6.

| Ticket | Description | Status |
|---|---|---|
| C-01 | ISA freeze | DECIDED in §1 of this document |
| C-02 | CSR additions, AlgebraID renumbering, WDEvent extension | DECIDED in §2-§4 of this document |

Required for unblocking:
- C-03 (qbpcu Go package) — can begin
- C-04 (qbpcu.Golden cycle-accurate impl) — needs C-03
- C-06 (skuld Go package) — can begin in parallel with C-03
- C-08 (wyrd renamed package) — can begin (independent)
- C-12 (Lean airtightening) — can begin (independent)

---

## 9. Companion documents and cross-references

This document is one of:

- *QBP-CU SiFive Interface Specification v0.2* — the full hardware interface spec; this document is its §11.1 freeze
- *Wyrd-Supervisor-Architecture-v0.2.md* — the supervisor architecture; references the encodings here
- *Wyrd-Workload-ISA-v0.1.md* — the analysis that justified the 10-instruction selection
- *Skuld-Spec-v1.0.md* — uses these encodings via the `qbpcu` package
- *Wyrd-BranchA-Contingency-v0.1.md* — argues for the AlgebraID 4 / 5 reservation

A change to this document propagates to all of the above.

---

*End of QBP-CU ISA Freeze (C-01) and CSR Additions (C-02) — DECIDED Rev 1.0*
