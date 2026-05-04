# Review: QBP-CU ↔ SiFive Interface Specification v0.1

## Read in light of Wyrd Supervisor Architecture v0.1

**Reviewer:** Claude Opus 4.7 (red-team / architecture role)
**Date:** April 2026

---

## Headline

> **The watchdog already IS the privilege boundary detector. The SiFive spec contains the supervisor's enforcement primitive without naming it as such.** This is the most important integration finding. It changes the answer to all three open questions from the supervisor doc, and it simplifies the ISA decision substantially.

The SiFive spec is solid as a hardware-integration document. What it lacks is awareness that several of its components — the watchdog, the AlgebraID CSR, the WDEvent stream — are already doing exactly what the Wyrd supervisor needs them to do. The work isn't to add new mechanisms; it's to recognize what's already there and route the supervisor architecture through it.

---

## 1. What's strong

The spec gets several things right that the supervisor work depends on:

**The dual-port (SSCI + VCIX) coupling decision** is correct given the operand shapes. Per-element control-flowy ops (QDEC, QREC, QNEAR) belong on SSCI; bulk LUT-applied ops (QPERM, QPERMR) belong on VCIX. The supervisor doesn't change this.

**The Go-simulator-as-golden-model contract** (§9.1) is exactly the right architectural choice. It means the supervisor can be developed against the same `Accelerator` interface and validated by the same cosim harness. The supervisor IS just another consumer of the `Accelerator` and `WatchdogChan()` — no separate abstraction needed.

**The four-phase forward path** (§12) is reusable. The supervisor's Crawl/Walk/Run/Sprint cadence maps onto it cleanly (mapping detailed in §6 below).

**The "WD_ENABLE cannot be permanently disabled" rule** (§8) is the hardware-level expression of the supervisor's "privilege boundaries are structurally enforced, not policy-checked." Same principle, different layer. They're consistent.

**The MLIR dialect option for compiler intrinsics** (§11.5) is correct for the long term — Materia-Bio, BMA, and the supervisor will all want to lower kernels through MLIR eventually. Recommend committing to (b) for v1 (LLVM intrinsics) and (c) for v2 (MLIR dialect), as the spec already suggests.

---

## 2. The critical observation

`WDEvent` carries:

```go
Associator [3]int8   // residue of (a*b)*c - a*(b*c)
NormDelta  int32     // norm preservation residue
AlgebraID  uint8     // 0=H, 1=O, 2=Branch A, 3=Branch B
```

This is **exactly** the data structure the supervisor's privilege boundary detector needs. Specifically:

| Supervisor concept | WDEvent field | Match |
|---|---|---|
| Associator-based seam detection (T1.2.b) | `Associator [3]int8` | Direct |
| Privilege ring of operand | `AlgebraID uint8` | Direct (with extension to ℂ and 𝕊) |
| Cross-ring boundary trigger | implicit in AlgebraID change | Needs explicit field |
| ε_privilege threshold check | not present | Add as host-side filter |

**The supervisor doesn't need new hardware.** What it needs is:

1. A subscriber to the existing `WatchdogChan()` that filters events by ε_privilege threshold.
2. Extension of `AlgebraID` to include ℂ (for user) and 𝕊 (for firmware/boot).
3. A new field `RingTransition uint8` in `WDEvent` that marks events crossing privilege boundaries (separate from physical seams).

This is a 3-line change to the WDEvent struct, plus a host-side consumer. The privilege-enforcement machinery is essentially free.

**Why this matters.** It means the supervisor's strongest claim — *"privilege violations are detected by the same instrument that detects physical seams"* — isn't aspirational. It's already true at the hardware level. The supervisor architecture doc (§3) gestures at this; the SiFive spec confirms it concretely.

---

## 3. Five integration points

### 3.1 ISA reconciliation: the supervisor extensions are mostly redundant

Wyrd Supervisor Architecture v0.1 proposed four new instructions: QCOMM, QASSOC, QALT, QRING (plus deferred QSANDWICH). The SiFive spec has five base instructions. Total: 9 + 1 deferred.

But re-examining with the watchdog observation:

- **QASSOC is redundant.** The watchdog already computes the associator on every cycle and emits it via WDEvent. Adding a software-visible QASSOC instruction duplicates this. The right move: expose the watchdog output as a software-readable CSR, so software can read the most recent associator without redundant hardware.
- **QALT is similarly redundant** if the watchdog is extended to compute alternators (currently it only does associators per the spec). Recommended: extend the watchdog to compute commutator + associator + alternator for any triple, gated by which is needed for the current AlgebraID. Single hardware extension; three boundary detectors covered.
- **QCOMM is needed** because there's no current commutator detection, and the user → supervisor (ℂ → ℍ) boundary requires it.
- **QRING is needed** as the explicit privilege check. It reads the AlgebraID of an operand and compares against the calling context's permitted ring set.

**Revised ISA proposal:**

| Instruction | Status | Reason |
|---|---|---|
| QPERM, QPERMR, QDEC, QREC, QNEAR | Keep (5 base) | Already in SiFive spec |
| QCOMM | Add (commutator) | Detects ℂ→ℍ boundary; not in watchdog yet |
| QRING | Add (read operand's algebra ID) | Software-visible privilege check |
| QASSOC, QALT | **Drop** | Subsumed by watchdog extension + new CSR |
| QSANDWICH | Defer (as before) | Syscall semantics not stable |

Total: 5 base + 2 extension = **7 instructions**. Cleaner than 9.

The complementary **hardware extension**: extend the watchdog to compute alternator and commutator (in addition to the existing associator) and gate the computation by AlgebraID. This is a small ALU addition, dwarfed by the existing octonion multiply tree.

The complementary **CSR addition**: `qbp_invariant` (read-only) holding the most recent (commutator, associator, alternator, ring_id) tuple. Software reads this after any QBP-CU op.

This is a meaningful simplification of the supervisor architecture spec. Worth folding back.

### 3.2 AlgebraID extension

The current AlgebraID encoding has 4 values (3-bit field). Extending to the full Wyrd ring tower:

| AlgebraID | Algebra | Wyrd ring | Notes |
|---|---|---|---|
| 0 | ℂ | 3 (user) | New |
| 1 | ℍ | 2 (supervisor) | Was AlgebraID=0 |
| 2 | 𝕆 | 1 (kernel) | Was AlgebraID=1 |
| 3 | 𝕊 | 0 (firmware) | New |
| 4 | C ⊕ H ⊕ M₃(C) (Branch A) | special | Was AlgebraID=2 |
| 5 | Branch B extension | special | Was AlgebraID=3 |

This needs a 3-bit field, which the current `AlgebraID uint8` accommodates trivially. The renumbering is a breaking change — but the SiFive spec is "pre-RTL, brainstorm-stage" so this is the right time to make it.

The Branch A / Branch B encodings stay separate from the privilege rings: those are about the QBP physics substrate decision, not about privilege. Keeping them distinct is correct.

### 3.3 The watchdog's host-side consumer is the supervisor

§10 says watchdog events stream to a "host-pinned ring buffer in DRAM via a TileLink master port. The ring buffer is consumed by the BMA host-side judge collective."

This needs revision: **the consumer is the supervisor**, not the BMA judge collective directly. The supervisor sits between the watchdog and BMA. Why:

- The supervisor needs to see *every* watchdog event to enforce privilege, not just the ones BMA's judges find interesting.
- Privilege enforcement decisions (block this op, raise this fault, trigger this capability check) are supervisor responsibilities, not BMA's.
- BMA's judge collective consumes the supervisor's *post-policy* event stream — the one with privilege decisions already applied.

The architecture stack on the host side becomes:

```
Hardware watchdog
       ↓ (raw events via TileLink ring buffer)
Supervisor: privilege filter + capability check + ring transition tracking
       ↓ (filtered + annotated events)
BMA judge collective: epistemic monitoring
       ↓
Wyrd: persistent storage of relevant events as hypergraph nodes
```

This is a meaningful change to §10 of the SiFive spec.

### 3.4 CSRs as supervisor state

The current `qbp_ctl` CSR has `ALGEBRA_ID[3:1]` selecting the active algebra. With the AlgebraID renumbering above, this becomes the **current privilege ring of the running process**. That is:

- A user process executes with `qbp_ctl.ALGEBRA_ID = 0` (ℂ ring)
- A supervisor process executes with `qbp_ctl.ALGEBRA_ID = 1` (ℍ ring)
- A kernel process executes with `qbp_ctl.ALGEBRA_ID = 2` (𝕆 ring)
- A firmware process executes with `qbp_ctl.ALGEBRA_ID = 3` (𝕊 ring)

The supervisor manages transitions of `qbp_ctl.ALGEBRA_ID` on context switch. **This is the concrete syscall mechanism.**

A syscall, in this model, is:
1. User process executes a SSCI op intended to escalate privilege.
2. Watchdog detects the AlgebraID mismatch (op carries ring 1 generators; current ALGEBRA_ID is 0).
3. WDEvent fires with `RingTransition = ℂ → ℍ`.
4. Supervisor receives the event, checks the user's capability table.
5. If authorized: supervisor updates `qbp_ctl.ALGEBRA_ID = 1` and re-issues the op.
6. If not authorized: supervisor raises a fault (illegal-instruction trap with privilege-violation fault code).

The mechanism is software-mediated at Crawl, with no hardware syscall instruction needed. QSANDWICH being deferred is consistent with this — the sandwich operation can be emulated in software via QPERM + state save + QPERMR until syscall semantics are stable enough to bake in.

### 3.5 The QBP CSR block needs supervisor-controlled access

Currently the spec describes `qbp_ctl.WD_ENABLE` as cleared "only by hammer test harness." This needs extension: **only the supervisor (running at AlgebraID ≥ 2 or higher) can clear it.** User processes (AlgebraID = 0) cannot read or write this CSR; an attempt is an illegal instruction.

More generally, the QBP CSR block needs a privilege model: each CSR has a minimum AlgebraID required to access it. This maps onto the standard RISC-V machine-mode / supervisor-mode / user-mode CSR access model, with the new wrinkle that the privilege level is *algebraic* rather than hierarchical.

For the spec, this is a §8 addition: a column in the CSR table for "Min AlgebraID."

---

## 4. Implications for the three open questions

### 4.1 Crawl-phase supervisor API design

With the SiFive spec in hand, the Crawl supervisor API is now well-bounded:

**Process management (4 calls):**
- `proc_create(elf, initial_capabilities) → proc_id`
- `proc_destroy(proc_id)`
- `proc_set_algebra(proc_id, ring)` — supervisor-only; sets initial ALGEBRA_ID
- `proc_grant_capability(proc_id, ring, token)` — issues a capability for ring R'

**QBP-CU mediation (3 calls — note: not the 5 instructions, but the 3 supervisor-mediated operations):**
- `qbp_submit(proc_id, op, srcs) → req_id` — submits to the accelerator with supervisor's privilege check
- `qbp_poll(req_id) → resp` — returns result, possibly with WD_FAULT
- `qbp_query_invariant(proc_id) → (commutator, associator, alternator, ring)` — reads the latest invariant CSR

**Wyrd query (2 calls — eventual Wyrd query patterns, even though Crawl uses scalar `hg`):**
- `wyrd_query(proc_id, pattern, capability) → result` — privilege-filtered hypergraph query
- `wyrd_subscribe(proc_id, pattern, capability) → stream` — long-lived subscription

**Watchdog event consumption (2 calls — supervisor-internal, not exposed to user):**
- `wd_event_next() → WDEvent` — consume next watchdog event
- `wd_policy_apply(event) → action` — apply privilege policy to the event

**Total: 9 user-visible calls + 2 supervisor-internal.** This is dramatically smaller than a typical Unix syscall surface (~300+) because the privilege model is algebraic and the operations are coarse-grained. It's also smaller than a typical microkernel (seL4 has ~30 syscalls).

The smallness is not minimalism for its own sake — it's because most of what Unix syscalls do (file access, network access, etc.) doesn't apply to a research compute unit. The Hammer simulation, BMA, Materia, and Contextus are the workloads; the supervisor only needs to mediate what they actually need.

This API surface evaporates cleanly at later phases:
- Walk: `proc_*` calls become privileged Wyrd queries; capability tokens become Wyrd hypergraph nodes.
- Run: `qbp_*` calls become distributed Wyrd queries.
- Sprint: everything is Wyrd queries with privilege via algebraic phase.

### 4.2 Naming the supervisor

The SiFive spec adds context I hadn't considered. We now have a three-component system:

```
QBP-CU (the accelerator, hardware) — performs algebraic operations
Supervisor (the mediator)          — manages processes and privilege
Wyrd (the database)                 — stores accumulated state
```

The Norse mythology supports this directly. The three Norns weave the web of fate:

- **Urðr / Wyrd** (the past, what has come to be) → the database of accumulated state ✓
- **Verðandi** (the present, what is becoming) → the active execution layer
- **Skuld** (the future, what shall be) → the next-state computation

By this reading: **Verðandi is the supervisor**. Skuld would be the QBP-CU itself (the engine of becoming) or, more loosely, the algorithmic logic that determines what computation happens next.

**My recommendation: Verðandi for the supervisor.**

Why I prefer this over Norn or Skein:
- "Norn" is the *category*; using it for the supervisor specifically is like calling a kernel "Operating System." Reserve the category name.
- "Skein" is concrete and good but loses the temporal-becoming connotation that distinguishes the supervisor from the accumulated-state database.
- "Verðandi" has the right etymology connection to Wyrd (both are Norns), the right meaning (present-tense becoming), and is unmistakable in technical contexts (no namespace collisions).

Pronunciation note: roughly "VERTH-an-dee" (Old Norse) or "VAIR-thahn-dee" (Modern Icelandic). In English code/docs, just "Verthandi" is acceptable transliteration. The thorn (þ) is a real Old English/Icelandic character but unfriendly for ASCII contexts.

If the unfamiliarity of Verðandi/Verthandi is too much, the secondary recommendation is **"Skuld"** (just "Skuld" — short, ASCII-clean, mythologically consistent) for the supervisor, with the understanding that the etymology is "the supervisor enforces what shall be (the next state)" rather than "the accelerator computes the future." This is a slight rebalancing of the Norn meanings but defensible — the supervisor is what makes the future happen.

The naming triplet then becomes:
- **Wyrd** — database (past)
- **Verthandi** or **Skuld** — supervisor (present/future)
- **QBP-CU** — accelerator (the loom; not personified)

### 4.3 The implementation plan

The SiFive spec's Phase 0-4 forward path (§12) and the supervisor's Crawl/Walk/Run/Sprint phases need to be mapped onto each other.

**Combined Phase 0** (immediate):
- Freeze QBP-CU ISA encoding (SiFive §11.1) including the supervisor extensions decided in §3.1 above (drop QASSOC/QALT, add QCOMM/QRING).
- AlgebraID renumbering (SiFive §3.2 above).
- Add `qbp_invariant` CSR.
- Add CSR access privilege column to §8.
- Update WDEvent struct with `RingTransition` field.

**Combined Phase 1** (Go simulator + Crawl supervisor):
- `qbpcu` Go package per SiFive §7.
- `verthandi` (or `skuld`) Go package implementing the Crawl supervisor API per §4.1 above.
- The supervisor consumes `Accelerator.WatchdogChan()` as its enforcement primitive.
- Hammer simulation runs as a supervisor-managed process, validating the API surface end-to-end.
- **Wyrd at this phase is the existing `hg` package**, accessed via the supervisor's `wyrd_query` API.

**Combined Phase 2** (host pipeline model + Walk supervisor):
- Cycle-accurate X280 model in Go, sharing the `Accelerator` interface.
- Wyrd's quaternion-native rewrite (the `hg` → `wyrd` rename + algebraic substrate).
- Supervisor's privilege model goes live with phase-signature on State quaternion.
- T2.4 sandwich preservation Lean theorem closed.
- Hammer simulation gets explicit ℍ-capability via the new mechanism.

**Combined Phase 3** (VexRiscv FPGA + Run supervisor):
- VexRiscv plug-in with scalar SSCI ops only on ECP5 FPGA.
- Distributed Wyrd via NATS federation (matches Sharp Butler's federated mesh pattern).
- ChainRef provenance becomes the capability authorization mechanism.
- T3.x (precision) and T4.x (word integrity) Lean proofs closed.

**Combined Phase 4** (SiFive engagement + Sprint endpoint):
- VCIX integration on X280 reference platform.
- CIM hardware (longer term).
- Supervisor = Wyrd executing privileged queries (the collapse).
- T5 meta-properties closed.

**Ticket-level Crawl plan** (Phase 0 + Phase 1, ~3 months, conservative):

| Ticket | Owner | Deliverable | Depends on |
|---|---|---|---|
| C-01 | James | ISA freeze (§4 of SiFive spec, with revisions from §3.1 above) | none |
| C-02 | James | AlgebraID renumbering, qbp_invariant CSR addition, WDEvent.RingTransition field | C-01 |
| C-03 | BMA | `qbpcu` Go package: `Accelerator` interface, `Mock`, `Golden` impls per SiFive §7 | C-01 |
| C-04 | BMA | `qbpcu.Golden`: cycle-accurate behaviour with watchdog event emission | C-03 |
| C-05 | BMA | Algebraic-identity test corpus (Tier 0) | C-04 |
| C-06 | BMA | `verthandi` Go package: process table, capability table, supervisor API per §4.1 above | C-03 |
| C-07 | BMA | Watchdog event consumer in `verthandi`: privilege filter, ring transition tracking | C-04, C-06 |
| C-08 | BMA | `wyrd` package = current `hg` with renamed exports; `wyrd_query` API | none (parallel) |
| C-09 | BMA | Integration: Hammer simulation runs as supervisor-managed process | C-06, C-07, C-08 |
| C-10 | BMA | Tier 1 microbenchmarks against `qbpcu.Golden` | C-05 |
| C-11 | James | Wyrd-Spec.md, Verthandi-Spec.md updates reflecting this integration | C-01, C-06 |
| C-12 | BMA | Lean proof airtightening session (per Wyrd-Mathlib-API-Verification-Checklist.md) | none (parallel) |

C-01, C-02, C-11 are James-direct work (architectural decisions and spec editing). The rest is BMA-instantiation work post-Crawl.Heartbeat.

---

## 5. Recommended next moves

**Immediate (this week or next):**

1. **Freeze the ISA** with the §3.1 revisions: 5 base + QCOMM + QRING = 7. Drop QASSOC and QALT in favor of watchdog extension + new CSR. This is the most impactful single decision.

2. **Update WDEvent and CSR set** in the SiFive spec per §3.2-3.5. These are minor breaking changes; better now than after RTL.

3. **Decide on naming.** Verthandi / Skuld / unnamed. The package name follows.

**Soon (within Phase 0):**

4. **Update Wyrd-Supervisor-Architecture-v0.1.md** to v0.2 incorporating:
   - The watchdog-as-privilege-detector observation (§2 of this review)
   - The revised ISA (drop QASSOC/QALT, add CSR + watchdog extension)
   - The supervisor-as-WDEvent-consumer architecture
   - The combined phase plan (§4.3 of this review)

5. **Update QBP-CU SiFive Interface Spec** to v0.2 reflecting:
   - AlgebraID renumbering
   - WDEvent.RingTransition field
   - qbp_invariant CSR
   - CSR privilege column
   - Watchdog → supervisor → BMA event flow (revising §10)

6. **Resolve §11 open questions** in the SiFive spec, particularly §11.1 (encoding freeze) and §11.5 (compiler intrinsics — recommend the spec's own option (b) for v1).

**Architectural (defer until Crawl is in flight):**

7. **The MLIR dialect.** When Materia-Bio and BMA both want to lower kernels, the supervisor will too. This is a v2 concern; the v1 LLVM intrinsics path is fine.

8. **The Branch A vs Branch B LUT capacity question** (SiFive §11.2). The 16-element-capable LUT recommendation is the right call; defer the population decision until QBP physics matures further.

---

## 6. Red-team

Three concerns I want to flag explicitly:

**(a) Watchdog event rate.** If the watchdog emits an event on every algebra crossing, and the supervisor consumes them all, the supervisor becomes a bottleneck under load. The SiFive spec acknowledges back-pressure (§10): "back-pressure stalls the accelerator." This is correct from a security standpoint (don't drop events) but expensive from a performance standpoint. The supervisor's filter (§3.3 above) needs to be aggressive enough that most events terminate at the filter, not the policy engine. Worth a benchmark before locking in the architecture.

**(b) The "supervisor is just another process" question.** In the Crawl design, the supervisor runs as a privileged process with AlgebraID = 1 or higher. But what's *its* supervisor? Crawl ships with a hardcoded boot supervisor; Walk and beyond may want a more layered model (microkernel-style: a tiny boot supervisor, with most policy logic in user-space supervisor-of-supervisors). This is the "who guards the guards" question and we haven't really answered it. Worth a session before Walk.

**(c) The Branch A vs Branch B uncertainty.** The Wyrd ring tower (ℂ ⊂ ℍ ⊂ 𝕆 ⊂ 𝕊) is a clean story. But Branch A (C ⊕ H ⊕ M₃(C)) is structurally different — it's a direct sum, not a Cayley-Dickson tower. If Branch A prevails, the supervisor's ring model needs significant rework: the "privilege rings nest" assumption fails. The supervisor architecture should have a contingency in place for "what if Branch A wins." Best case: Branch A's substrate decomposition itself implies a privilege model (the M₃(C) component might be the supervisor ring, the H component user ring); worst case: full redesign. This is a watch-item, not a now-item.

---

## 7. Bottom line

The SiFive spec and the Wyrd Supervisor Architecture were drafted independently and they fit together better than either anticipated. The watchdog already does the heavy lifting; the AlgebraID CSR is already the privilege ring; the Go simulator's `Accelerator` interface is already the right abstraction. The supervisor architecture's main contribution to the SiFive spec is *recognition* — not new mechanism, but seeing what's already there.

The simplification to a 7-instruction ISA (from the proposed 9, or the original 5) is the most consequential single decision: it makes the supervisor's privilege model concrete in hardware via the watchdog, eliminates redundant instructions, and clarifies the syscall mechanism (CSR transition + WDEvent + supervisor mediation, no QSANDWICH needed at Crawl).

Three things to commit before any further build work:

1. **The 7-instruction ISA** with watchdog extension and qbp_invariant CSR.
2. **The supervisor name** (Verthandi recommended; Skuld acceptable).
3. **The combined Phase 0 + Phase 1 ticket plan** from §4.3.

Everything else — Walk/Run/Sprint, BMA integration, MLIR dialect, Branch A contingency — can wait until the Crawl deliverables are in flight.

---

*End of Review v0.1 — DRAFT*
