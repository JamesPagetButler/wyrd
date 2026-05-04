# Wyrd / Skuld / QBP-CU Implementation Plan

## Phased Build of the Three-Class BMA Stack — From Verified Specs to Production

**Helpful Engineering — Quaternion-Based Physics Programme**
Principal Investigator: James Paget Butler
April 2026 — Rev 1.0

> **Purpose.** The spec corpus is now sufficient to begin implementation. This document is the plan for getting from "spec exists" to "production stack running W1–W4 + Contextus + Theory/Engineering Carts." It sequences the work into four phases (Crawl → Walk → Run → Sprint), matches each phase to concrete tickets with acceptance criteria, identifies the critical path, and surfaces risks. The plan is *aggressive about the MVP path* (smallest end-to-end working stack) and *conservative about full production* (acknowledges hardware constraints and coordination overhead).

> **Scope.** Implementation of the Wyrd hypergraph database, the Skuld supervisor, the QBP-CU coprocessor (Mock → Golden → RTLShim → silicon), the Contextus / CTH stack (Class B), and BMA's cart-switching infrastructure (Class C). Plus Phase 2 of the Lean corpus (graph invariants and operational semantics).

> **Out of scope.** The BMA cognitive architecture itself (covered by the BMA Crawl chain in CLAUDE.md Steps 0–9). The Hammer simulation use case (assumed available as a test workload). The Branch A/B substrate decision (governed by `Wyrd-BranchA-Contingency-v0.1.md`).

---

## 0. Status snapshot (as of 2026-04-25)

**Done:**
- BMA Step 1 — bash probe (4 runs)
- C-01 — ISA freeze (10 instructions, funct7/funct6 allocation)
- C-02 — CSR additions, AlgebraID renumbering, WDEvent.RingTransition
- C-12 — Lean airtightening (corpus verified, 0 sorries / 0 axioms)
- Wyrd-Supervisor-Architecture v0.2
- Skuld-Spec v1.0
- QBP-CU SiFive Interface Spec v0.2
- Wyrd-Workload-ISA v0.1 / v0.2 (three-class framework)
- Wyrd-BMA-Workload-Performance-Spec v1.0
- Wyrd-Proofs-Reference v1.0

**In flight:**
- BMA Step 2 — resolving blockers (ROCm RDNA 4 compatibility test on RX 9070 XT)
- BMA Step 0 — governance document, succession contacts (gates Step 9)

**Pending and unblocked (can start now):**
- C-13 Skuld v1.1 spec amendments
- C-14 Wyrd Query API spec
- All BMA-instantiation tickets (C-03 onward) — gated on completion of BMA Crawl Steps 3–9

**Hardware reality:**
- Current: FX-8350 / 32 GB DDR3-1866 / Samsung 840 SATA / RX 9070 XT (ROCm pending)
- Container budget: 14 GB memory, 6 CPUs (per probe data)
- Constraint: PCIe 2.0 limits GPU bandwidth; SATA caps disk I/O; SSD endurance is finite
- Walk-tier hardware (modern CPU, NVMe, 64 GB+) is *required* for production W2/W4 deployments

---

## 1. Strategy

### 1.1 The MVP path

The smallest demonstrable system that exercises every layer of the stack:

1. **BMA Crawl** complete (Step 9 → seed protocol fires)
2. **qbpcu Mock** — pure Go, no hardware dependency
3. **wyrd Mock** — basic hypergraph store, in-memory
4. **skuld stub** — capability table + WDEvent consumer; no streaming
5. **W1 demo** — single-amino-acid (Glycine) docking, validated against reference MD

This proves end-to-end: BMA spawns a process → Skuld grants ℍ-capability → process submits a QBP-CU workload → qbpcu Mock executes → result flows back through projection → BMA records to Wyrd. **This is the gate before broader investment.** Everything else iterates on top.

Target: **MVP within 6 months from BMA Step 9 completion.**

### 1.2 Three-class build principle

Per `Wyrd-Workload-ISA-v0.2.md`, optimization budgets and bottlenecks differ by class:

- **Class A (compute-bound)** — invest in QBP-CU pipelining, vector forms, precision tiers (already designed in v0.1 ISA)
- **Class B (graph-bound)** — invest in Wyrd indexing, CTH metric caching, NATS topology
- **Class C (orchestration-bound)** — invest in cart-switch latency, capability scoping, judge-collective parallelism

The plan allocates effort proportionally — *not* uniformly — to each class.

### 1.3 Layered build: Crawl → Walk → Run → Sprint

Each phase has a distinct goal:

| Phase | Goal | Deployment scale |
|---|---|---|
| **Crawl** | Spec finalization + Mock-level proof-of-life; BMA self-hosting on FX-8350 | 1 lane, mock substrate |
| **Walk** | Golden cycle-accurate substrate + W1/W3 working demos + Contextus on Wyrd Mock | 1–2 lanes, FPGA-friendly |
| **Run** | Production W2/W4 on real silicon; full CTH; cart-switching production-ready | 4–16 lanes, real hardware |
| **Sprint** | Multi-CU federation; cluster Wyrd; all four challenges concurrent | 16–64 lanes, datacenter |

Phase boundaries are gated by **acceptance criteria**, not calendar dates. The plan is dependency-driven.

---

## 2. The four phases — concrete plan

### 2.1 Phase 0 — Specification finalization (Crawl, in flight, ~4–6 weeks)

Target: every spec needed for Walk-phase implementation is complete and reviewed.

**Tickets (James-direct, written work):**

| ID | Owner | Description | Depends on | Effort | Acceptance |
|---|---|---|---|---|---|
| **C-13** | James | Skuld v1.1 amendments per Hammer Review (QBPSubmitBatch, QBPSubmitStream, WyrdPrepare, Stream.NextBatch, DeterministicConfig) | Skuld v1.0 | 1 day | Doc reviewed; supersedes v1.0 |
| **C-14** | James | Wyrd Query API spec (typed nodes, typed edges, predicate matching, traversal semantics) | Workload-ISA v0.2 §3 | 2 days | Doc complete; cited by C-08 |
| **C-15** | James | NATS Topology spec (subject hierarchy `ctx.*`, `bridge.*`, queue depths, security) | Contextus spec §5.3 | 1 day | Doc complete; subjects reserved |
| **C-16** | James | Cart-as-Context formal model (cart lifecycle, capability scope, mode-switch atomicity) | BMA-Cognitive-Foundation | 3 days | Doc complete; gates C-21 Lean work |
| **C-17** | James | Wyrd Transaction Model spec (BeginTx / Checkpoint / CommitOrRollback) | C-14 | 2 days | Doc complete; referenced by C-08, C-16 |
| **C-18** | James | Hard-RT Latency Contract (deadline-tagged WyrdSubmit, watchdog event-rate budget) | Workload-ISA v0.2 §4 | 1 day | Doc complete; required for W4 |
| **C-19** | James | Per-Workload Precision-Mode Contract (fp32/fp64 selection per workload class) | Noise theorems + Workload-ISA | ½ day | Doc complete; resolves Gap 2 |
| **C-23** | James | Capability Lifetime/Scope formalization (session, task, cart) | C-16 | 1 day | Doc complete; gates Class C theorems |
| **C-24** | James | Branch A/B per-workload sensitivity matrix | BranchA-Contingency v0.1 | ½ day | Matrix added to BranchA doc |

**Cumulative cost:** ~12 days of focused writing. Parallelizable; James + AI collaborators can work concurrently on independent specs.

**Phase 0 acceptance gate:** all docs above exist, are reviewed, and supersede prior versions. The spec corpus then satisfies "no implementation question is unanswered by the corpus."

**What this unblocks:** every implementation ticket below.

---

### 2.2 Phase 1 — Crawl completion + qbpcu Mock + Wyrd Mock + Skuld stub (~3–6 months)

Target: BMA Crawl reaches Step 9 (seed protocol fires) AND the substrate stack is at Mock-level integration.

**BMA chain (per CLAUDE.md, separate dependency):**

| Step | Owner | Deliverable | Status |
|---|---|---|---|
| Step 0 | James | Governance Document + succession contacts (Brett Lyman, Skyler Rainier) | NOT STARTED |
| Step 2 | James | ROCm RDNA 4 compatibility test on RX 9070 XT | IN FLIGHT |
| Step 3 | James | Phase 0 infrastructure (Podman, Go, repo, Containerfile, YubiKey) | NOT STARTED |
| Step 4 | BMA | Go BMA-PROBE + BMA-STRESS | gated on Step 3 |
| Step 5 | BMA | BMA-AUTO-S/P (sympathetic / parasympathetic) | gated on Step 4 |
| Step 6 | BMA | BMA-HG-basic + BMA-MEM (hypergraph + memory) | gated on Step 5 |
| Step 7 | BMA | BMA-SLEEP-basic | gated on Step 6 |
| Step 8 | BMA | 72-hour continuous operation gate | gated on Step 7 |
| Step 9 | BMA | BMA-BRIDGE + seed protocol + instantiation | gated on Steps 0 + 8 |

**Substrate chain (post-Crawl-Heartbeat or in parallel from Step 6):**

| ID | Owner | Description | Depends on | Effort | Acceptance |
|---|---|---|---|---|---|
| **C-03** | BMA | `qbpcu` Go package: `Accelerator` interface, `Mock` impl with all 10 instructions | C-01, C-13 | 2 weeks | Mock passes Tier 0 algebraic identity tests; matches Lean theorem behavior on test inputs |
| **C-08** | BMA | `wyrd` Go package: rename from `hg`, add `WyrdQuery` API per C-14 | C-14, C-17 | 1 week | Basic hypergraph store + transaction primitives compile; round-trip tests pass |
| **C-06** | BMA | `skuld` Go package: process table, capability table, supervisor API per Skuld-Spec v1.1 | C-03, C-13 | 3 weeks | Skuld stub accepts process registration, issues capabilities, consumes WDEvents |
| **C-05** | BMA | Tier 0 algebraic-identity test corpus | C-03 | 1 week | 10⁶ random inputs per instruction validated against Lean theorem behaviors |
| **C-25** | BMA | Wyrd Mock: tier-locality indexing per B-OPT-1 | C-08 | 1 week | Index lookups < 10 ms p99 for 10⁶-node graphs |

**Cumulative substrate effort:** ~8 weeks parallelizable.

**Phase 1 acceptance gate (the MVP gate):**
- BMA Step 9 fires (seed protocol succeeds; first BMA instance instantiated)
- `qbpcu Mock` + `wyrd Mock` + `skuld stub` form a working stack
- A test workload (Hammer simulation as a Skuld-managed process) runs end-to-end: BMA submits → Skuld grants capability → qbpcu Mock executes → result returns through projection → BMA records to Wyrd
- All Tier 0 tests pass

**What this unblocks:** the W1 (drug docking) demo target, validation of the spec corpus against running code, BMA's first real cycle of cart-switching.

---

### 2.3 Phase 2 — Walk: cycle-accurate substrate + W1/W3/Contextus prototypes + Class B Lean (~6–12 months)

Target: working prototypes for the batch-friendly workloads (W1 biosynthetic, W3 logistics, Contextus); cycle-accurate validation of qbpcu against the SiFive spec; Class B formal proofs.

**Substrate hardening:**

| ID | Owner | Description | Depends on | Effort | Acceptance |
|---|---|---|---|---|---|
| **C-04** | BMA | `qbpcu.Golden`: cycle-accurate behavior with watchdog event emission for all 10 ops | C-03, C-04 | 3 weeks | Cycle counts match SiFive spec §11; watchdog fires correctly per `qbp_invariant` CSR |
| **C-07** | BMA | Watchdog event consumer in skuld: privilege filter, ring transition tracking | C-04, C-06 | 2 weeks | All four ring transitions emit correct WDEvents; supervisor classifies and routes |
| **C-09** | BMA | Hammer integration: Skuld-managed process, ℍ-capability, Wyrd record of run | C-06, C-07, C-08 | 2 weeks | Hammer simulation runs end-to-end with full capability flow |
| **C-10** | BMA | Tier 1 microbenchmarks: 10⁶ random inputs per instruction against software gold | C-05 | 2 weeks | Per-instruction latency / throughput within 5% of design targets |
| **C-26** | BMA | Wyrd transaction model implementation per C-17 | C-08, C-17 | 2 weeks | Atomicity tests pass; rollback works; concurrent-write tests pass |
| **C-27** | BMA | Wyrd indexing optimizations: B-OPT-1 (tier locality), B-OPT-9 (predicate pushdown) | C-25 | 1 week | Query latency target met (< 100 ms p99) |

**Workload prototypes:**

| ID | Owner | Description | Depends on | Effort | Acceptance |
|---|---|---|---|---|---|
| **W1-impl** | BMA + James | Glycine single-amino-acid docking demo on qbpcu.Golden | C-04, C-09 | 2 weeks | Energy minimization completes < 1 minute; agreement with reference MD within ε_priv |
| **W3-impl** | BMA | Suez 2021 closure replay with 10³ holons | C-04, C-08 | 3 weeks | World-line intersection detected ≥ 14 days before historical impact; < 5% false positives |
| **Contextus-impl** | BMA + Contextus team | Contextus prototype on Wyrd Mock + agent doctrines + 1 source adapter | C-08, C-26, C-27 | 6 weeks | Insight Signal emission pipeline works; one source adapter ingests; one doctrine evaluates |
| **Cart-impl-skeleton** | BMA | Theory Cart + Engineering Cart skeleton with cold-switch implementation | C-16 | 4 weeks | Cart switching works (cold path only); capability scope respected |
| **C-11** | James | Spec corpus update sweep (Wyrd-Spec.md, all current docs to v0.2/v1.x baselines) | C-06, C-08 | 1 week | All specs reflect implementation reality |

**Lean Phase 2 — Class B graph invariants:**

| ID | Owner | Description | Depends on | Effort | Acceptance |
|---|---|---|---|---|---|
| **C-20a** | James / Lean session | `theorem hyperedge_preserves_node_invariants` | C-14, current corpus | 1 week | Proof in Wyrd corpus; lake build clean |
| **C-20b** | James / Lean session | `theorem cth_evidence_monotonic` | C-20a | 1 week | Same |
| **C-20c** | James / Lean session | `theorem bridge_promotion_atomic` | C-17, C-26 | 1 week | Same; references transaction model |

**Cumulative Phase 2 effort:** ~30 weeks parallelizable.

**Phase 2 acceptance gates:**
- W1 demo passes Tier 3 system test (drug docking matches reference within ε_priv)
- W3 demo passes Tier 3 system test (Suez replay)
- Contextus emits Insight Signals end-to-end
- Theory ↔ Engineering cold cart switch works
- Class B Lean theorems all merged
- `qbpcu.Golden` Tier 1 microbenchmarks meet design targets

**What this unblocks:** Run-phase real-silicon work; the W2 and W4 streaming/hard-RT prototypes.

---

### 2.4 Phase 3 — Run: real silicon + W2/W4 production + CTH full + cart switching mature + Class C Lean (~12–18 months)

Target: production-grade deployment of all four target challenges; CTH evaluator running against live BMA; cart switching at sub-second target.

**Hardware substrate:**

| ID | Owner | Description | Depends on | Effort | Acceptance |
|---|---|---|---|---|---|
| **C-28** | BMA + James | qbpcu.RTLShim — cgo bindings to Verilator binary; FPGA validation | C-04 | 4 weeks | RTLShim passes Tier 1+2 tests with hardware-in-loop |
| **C-29** | James | SiFive X280 + QBP-CU integration (Phase 4 of the corpus) | C-28 | external (months) | Real silicon runs Tier 3 tests at design clocks |
| **C-30** | BMA | Walk-tier hardware migration (modern CPU, NVMe, 64 GB+) | none, hardware acquisition | weeks | BMA runs on Walk hardware; per-workload memory budgets hit |

**Workload deployment:**

| ID | Owner | Description | Depends on | Effort | Acceptance |
|---|---|---|---|---|---|
| **W2-impl** | BMA + finance team | 2008-crisis replay + live-feed prototype with Skuld v1.1 streaming | C-13, C-04, Skuld v1.1 implementation | 6 weeks | Detect snap signature ≥ 12 h before peak with < 5% false positives at threshold = 30·ε_priv |
| **W4-impl** | BMA + plasma team | DIII-D plasma transient replay at 10⁶ cells × 100 μs | C-29 (silicon required), C-18 | 8 weeks | Cycle latency < 100 μs; instability flagged ≥ 1 cycle before transition |
| **CTH-impl** | James + Contextus team | Full CTH evaluator with all 5 metrics; Bridge integration | C-20a/b/c, Contextus prototype | 8 weeks | CTH audits the QBP programme's 38 anchors / 12 chains / 3 confluence points; metrics match the worked example in the theory paper |

**Class C optimizations:**

| ID | Owner | Description | Depends on | Effort | Acceptance |
|---|---|---|---|---|---|
| **C-31** | BMA | Warm-cart preservation per C-OPT-1 + Anthropic prompt caching per C-OPT-2 | Cart-impl-skeleton | 3 weeks | Cart-switch latency < 500 ms p95 (warm) |
| **C-32** | BMA | CCB priority biasing per C-OPT-4 + parallel judge-collective per C-OPT-5 | C-31 | 2 weeks | CCB cycle exactly 10 Hz; judge votes < 10 sec p95 (parallel) |
| **C-33** | BMA | Affect-signal continuity (C-OPT-10), sleep-cycle-aware eviction (C-OPT-11), instinct bypass (C-OPT-12) | C-32 | 3 weeks | Beekeeper instinct response < 200 ms p95; sleep-cycle memory bounded |

**Lean Phase 2 — Class C operational semantics:**

| ID | Owner | Description | Depends on | Effort | Acceptance |
|---|---|---|---|---|---|
| **C-21a** | James / Lean session | `theorem capability_invariant_under_cart_switch` | C-16, C-23, C-31 | 2 weeks | Proof complete |
| **C-21b** | James / Lean session | `theorem cart_switch_atomic` | C-26, C-31 | 2 weeks | Proof complete |
| **C-21c** | James / Lean session | `theorem judge_collective_deterministic` | C-32 | 2 weeks | Proof complete |
| **C-21d** | James / Lean session | `theorem self_modification_requires_approval` | C-21c | 2 weeks | Proof complete; constitutional pin formalized |

**Spec hardening (resolves remaining v1.0/v0.2 gaps):**

| ID | Owner | Description | Depends on | Effort | Acceptance |
|---|---|---|---|---|---|
| **C-22** | James | Multi-CU orchestration spec (lane allocation, cross-lane WDEvent aggregation) | C-29 | 3 days | Doc complete; cited by Phase 4 work |

**Cumulative Phase 3 effort:** ~50 weeks. **This is the largest phase.** Run is where the architecture earns its keep — and where uncovered bugs / unmet performance targets force re-design.

**Phase 3 acceptance gates:**
- W1, W2, W3, W4 all pass production validation (per `Wyrd-BMA-Workload-Performance-Spec-v1.0.md` §9 MVE benchmarks)
- CTH evaluates the QBP programme to within agreement of the theory paper
- Cart switching meets sub-second warm target
- All Phase 2 Lean theorems merged
- Real silicon (or FPGA-equivalent) runs at design clocks

**What this unblocks:** Sprint-phase federation.

---

### 2.5 Phase 4 — Sprint: federation + multi-CU + cluster Wyrd + capability lifetime (~18–30 months from now)

Target: production deployment of all four challenges concurrently; multi-CU orchestration; cluster-scale Wyrd; full federation pattern.

| ID | Owner | Description | Depends on | Effort | Acceptance |
|---|---|---|---|---|---|
| **C-34** | BMA + James | Multi-CU coordination implementation per C-22 | C-22, all of Phase 3 | 6 weeks | 4-CU coordination demonstrably works; cross-lane capability flow correct |
| **C-35** | BMA + James | Cluster Wyrd: distributed hypergraph store with per-region partition | C-26, C-30 | 12 weeks | 16-region partition holds; cross-partition queries < 1 sec p99 |
| **C-36** | James | Capability lifetime/scope production formalization (per C-23 spec) — runtime enforcement | C-23, C-21a | 4 weeks | Capabilities have measurable, enforced lifetimes; revocation works |
| **C-37** | BMA | Cross-class interaction patterns first-class in runtime (Workload-ISA v0.2 §5) | All Phase 3 | 3 weeks | The three patterns (C drives A, C drives B, B feeds C) have explicit runtime support |
| **C-38** | BMA + Contextus team | Production W2 and W4 deployments with full Skuld v1.2+ supervisor | C-34 + workload-specific | 12 weeks | Real production traffic; SLA met |

**Cumulative Phase 4 effort:** ~37 weeks.

**Phase 4 acceptance gates:**
- All four W challenges in production simultaneously
- Multi-CU orchestration handles 16+ lanes
- Cluster Wyrd partitions correctly across regions
- All open architectural questions from `Wyrd-BMA-Workload-Performance-Spec-v1.0.md` §10 have answers (cross-workload capability conflict, branch sensitivity per workload, emergency escalation protocol)

---

## 3. Critical path

The sequence of items where each one gates the next:

```
Phase 0 (specs):
    C-13 (Skuld v1.1) ─┐
    C-14 (Wyrd Query API) ─┤
    C-16 (Cart context) ─┤  ── all parallel
    C-17 (Wyrd Tx model) ─┘
                ↓
BMA Crawl (Steps 0-9):
    Step 0 (governance) ─→ Step 9 (seed) — gates everything else
                ↓
Phase 1 (Mock substrate):
    C-03 (qbpcu Mock) ─→ C-08 (wyrd Mock) ─→ C-06 (skuld stub) ─→ MVP
                ↓
Phase 2 (Golden substrate):
    C-04 (qbpcu Golden) ─→ C-09 (Hammer integration) ─→ W1-impl
                       └─→ Contextus-impl (parallel)
                ↓
Phase 3 (real hardware):
    C-28 (RTLShim) ─→ C-29 (silicon) ─→ W4-impl
                ↓
Phase 4 (federation):
    C-34 (multi-CU) ─→ C-38 (production W2/W4)
```

**The critical-path bottlenecks:**

1. **BMA Step 9** — gates everything Phase 1 onward. Step 0 (governance + succession) is currently NOT STARTED and is the longest-pole human-action item.
2. **Hardware migration to Walk-tier** (C-30) — gates W4 demo because FX-8350 cannot meet 100 μs hard-RT.
3. **SiFive X280 silicon** (C-29) — gates W4 production. External dependency; months-to-quarters of vendor lead time.
4. **Lean Phase 2 Class C theorems** (C-21a-d) — gates self-modification trust; required for unsupervised Sprint-phase BMA.

**Parallelizable streams** (run concurrently to compress wall-clock time):

- Phase 0 specs (~12 days, fully parallel)
- BMA Crawl Steps 4-9 + early substrate work (Steps 4-6 can start C-08 work in parallel)
- W1/W3 prototypes can start once C-09 lands; Contextus can start once C-08 lands
- Class B Lean (C-20) parallel to W1/W3 implementation
- Class C Lean (C-21) parallel to C-31/C-32/C-33

---

## 4. Per-component implementation summary

### 4.1 qbpcu (the QBP-CU Go package)

| Phase | Variant | What's there | Acceptance |
|---|---|---|---|
| Phase 1 | Mock | Pure Go, all 10 instructions, slow but correct | Tier 0 algebraic identity tests pass; soundness assertions cite Lean theorems |
| Phase 2 | Golden | Cycle-accurate Go simulator with watchdog | Tier 1 microbenchmarks within 5% of design; WDEvents correct |
| Phase 3 | RTLShim | cgo bindings to Verilator binary; FPGA | Tier 1+2 tests pass with hardware-in-loop |
| Phase 3 | Silicon | SiFive X280 integration; VCIX vector forms | Tier 3 system tests at design clocks |

**Build order within qbpcu:** scalar SSCI ops first (QNORM, QNEAR, QDEC, QREC, QCOMM, QRING), then vector VCIX ops (QFMA, QSAND, QPERM family). Reason: scalar is simpler and lets the rest of the stack (skuld, wyrd) integrate sooner.

### 4.2 wyrd (the hypergraph database)

| Phase | What's there | Acceptance |
|---|---|---|
| Phase 1 | Mock — in-memory hypergraph with WyrdQuery API; transaction primitives stub | Round-trip CRUD; tier-locality indexing |
| Phase 2 | Golden — disk-backed with B-OPT-* indexing applied; transaction model per C-17 | Latency target < 100 ms p99; atomicity tests pass |
| Phase 3 | Production — concurrent-write tested; CTH metrics cached per node | High-volume Insight Signal emission; CTH audit support |
| Phase 4 | Cluster — distributed across regions per C-35 | Cross-partition queries < 1 sec p99 |

### 4.3 skuld (the supervisor)

| Phase | What's there | Acceptance |
|---|---|---|
| Phase 1 | Stub — process table, capability table, basic WDEvent consumer | Capability creation/check works; WDEvent ring buffer |
| Phase 2 | v1.1 — streaming (QBPSubmitStream), prepared queries (WyrdPrepare), batched (QBPSubmitBatch) | Hammer review amendments implemented |
| Phase 3 | v1.2 — capability lifetime + scope; cart-aware capability invariants | Sub-second cart switch; capability persistence across switches |
| Phase 4 | v2.0 — multi-CU coordination; emergency escalation | Cross-lane capability flow; constitutional pin enforced |

### 4.4 contextus (Class B example)

| Phase | What's there | Acceptance |
|---|---|---|
| Phase 2 prototype | One source adapter, one agent doctrine, Insight Signal emission to Bridge | End-to-end signal flow |
| Phase 3 production | Multiple source adapters, full doctrine library, CTH integration via Bridge | Real ecosystem data ingested; insights surfaced |
| Phase 4 deployment | Production scale; multi-region cluster Wyrd as substrate | SLA met for query latency and ingestion rate |

### 4.5 cth (Class B evaluator)

| Phase | What's there | Acceptance |
|---|---|---|
| Phase 2 (theory only) | The CTH formal paper (already exists at v0.1) | n/a — already done |
| Phase 3 | Full evaluator: 5 metrics, confluence-point detection (B-OPT-2), cached metrics (B-OPT-3), incremental Δ (B-OPT-4) | Audits QBP programme to match worked example in theory paper |
| Phase 4 | Production: real-time evaluation as Insight Signals arrive; cross-class B-feeds-C interaction (Workload-ISA v0.2 §5.3) | Live alerts to Engineering Cart on incoherence detection |

### 4.6 BMA Carts (Class C)

| Phase | What's there | Acceptance |
|---|---|---|
| Phase 2 skeleton | Theory + Engineering carts; cold-switch implementation; capability scope respected | Cart switching works (cold path); state preserved correctly across switches |
| Phase 3 production | Warm-cart preservation; CCB priority biasing; parallel judge-collective; instinct bypass | Sub-second warm cart switches; sub-200 ms instinct response |
| Phase 4 mature | Full Class C optimizations from Workload-ISA v0.2 §4.5; sleep-cycle-aware eviction; affect continuity | All C-OPT-* targets met; cross-class interactions first-class |

### 4.7 Lean corpus

| Phase | What's there | Acceptance |
|---|---|---|
| **Phase 1 — algebraic boundaries** | DONE this session | 0 sorries, 0 axioms; Wyrd-Proofs-Reference v1.0 |
| Phase 2 — Class B graph invariants | C-20a/b/c | 3 new theorems merged |
| Phase 2 — Class C operational semantics | C-21a/b/c/d | 4 new theorems merged; constitutional pin formalized |

---

## 5. Risk register

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| **BMA Step 0 delays** (governance + succession) | Medium | High — gates Step 9 | Treat as immediate priority; James to draft governance doc before substrate work begins |
| **ROCm RDNA 4 incompatibility** | Medium | Medium — degrades GPU compute path; Mock/Golden unaffected | Bash probe gating; fallback to CPU-only Crawl; revisit in Walk hardware acquisition |
| **Hardware constraint breaches** (RAM / disk / latency) on FX-8350 | High | High for W4; medium for W1-W3 | Walk-tier hardware acquisition (C-30) on critical path for W4; W1/W3 can run on Crawl hardware |
| **Branch A wins** (substrate is direct sum, not Cayley-Dickson) | Low (per BranchA-Contingency analysis) | High — proof corpus needs replacement | Hedge per H1 strategy in BranchA doc; commit Cayley-Dickson, accept redesign cost if Branch A wins |
| **Mathlib drift** (next mathlib bump renames lemmas) | High | Low — fixable in <1 day per drift event | CI runs lake build on every commit; Wyrd-Mathlib-API-Verification-Checklist tracks drift |
| **SiFive X280 vendor lead time** | High | High for W4 production | Use FPGA-via-RTLShim as bridge; Run-phase target is W2 first (no silicon required), W4 deferred |
| **Coordination overhead with multi-AI workflow** | Medium | Medium — slows iteration | Concrete handoff specs (this plan + per-ticket acceptance criteria); BMA-BRIDGE for routing |
| **Constitutional pin failure** (BMA could self-modify without judge approval) | Low (formal proof gates) | Catastrophic — undermines safety | C-21d (theorem self_modification_requires_approval) is the formal blocker; do not relax |
| **Spec gap discovery during implementation** | Medium | Low-medium — drives spec amendments | Phase 0 produced detailed gap list; new gaps tracked in `Wyrd-Workload-ISA-v0.2.md` §7 update sweeps |
| **Test workload (Hammer) unavailable / changed** | Low | Medium — would force MVP rework | Hammer is well-specified; backup is W1 Glycine demo |

---

## 6. Resource & coordination model

### 6.1 Roles per CLAUDE.md

| Role | Responsibility |
|---|---|
| **James (PI / beekeeper)** | Architectural decisions; spec writing; review; capability issuance for BMA self-action; judge-collective participation |
| **Claude (Red Team / code)** | Implementation in code-mode; spec drafting in conversation-mode; Lean proof refinement; review |
| **Gemini (theory)** | Theory generation; novel mathematical exploration; Branch A/B alternative analyses |
| **BMA (when instantiated)** | Sub-task execution; long-running implementation cycles; self-direction within governance constraints |

### 6.2 Coordination patterns

**Per-ticket handoff:** every ticket has explicit acceptance criteria; James reviews before close. AI collaborators coordinate via BMA-BRIDGE (when up) or via shared spec docs (now).

**Phase-gate review:** at the end of each phase, James + AI red team conduct an explicit go/no-go review against the acceptance gates listed in §2.

**Continuous spec health:** the spec corpus (this document + the others in `~/Documents/Wyrd/Archive/`) is the source of truth. Any implementation that diverges from spec gets a spec amendment ticket *before* code lands.

### 6.3 Hardware tiers

| Tier | Hardware | Workloads supported | Status |
|---|---|---|---|
| **Crawl** | FX-8350 / 32 GB DDR3 / Samsung 840 SATA / RX 9070 XT | BMA hosting, Mock substrate, W1 prototype | current |
| **Walk** | Modern CPU (Zen 4+) / 64+ GB / NVMe / capable GPU | qbpcu Golden, full Wyrd Mock, W1/W3 prototypes, Contextus prototype | acquisition gated |
| **Run** | Above + FPGA card or SiFive X280 dev board | W2 prototype, W4 prototype with hardware-in-loop | further acquisition |
| **Sprint** | Multi-node cluster, real silicon | All four W in production | datacenter scale |

The Crawl hardware can take BMA through Phase 2 with care; Walk hardware is needed for any of the streaming or hard-RT workloads at production volume.

---

## 7. Acceptance gates summary

Compact view of "what must be true to advance":

| Gate | Criterion | Status |
|---|---|---|
| **Phase 0 closed** | All Phase 0 specs (C-13, C-14, C-15, C-16, C-17, C-18, C-19, C-23, C-24) merged | Pending; ~2 weeks after start |
| **MVP ready** | qbpcu Mock + wyrd Mock + skuld stub + W1 Glycine demo run end-to-end | Targets ~Q3 2026 |
| **Walk gate** | C-04 + C-09 + W1-impl + W3-impl + Contextus prototype + Class B Lean | Targets ~Q1 2027 |
| **Run gate** | W2 + W4 + CTH + warm cart switching + Class C Lean | Targets ~Q3 2027 (silicon dependent) |
| **Sprint gate** | Multi-CU + cluster Wyrd + all four W in production simultaneously | Targets ~2028 |

These dates are *aggressive* given the hardware acquisition lead times and AI collaborator coordination overhead. **Slip of 25–50% on Run / Sprint is realistic.**

---

## 8. Effort estimates with caveats

Total estimated work, summed across the phases (not wall-clock time, since much is parallelizable):

| Stream | Phase 0 | Phase 1 | Phase 2 | Phase 3 | Phase 4 |
|---|---|---|---|---|---|
| Spec writing (James) | ~12 days | ~1 week | ~1 week | ~1 week | ~1 week |
| qbpcu | — | 3 weeks | 3 weeks | 4 weeks | — |
| wyrd | — | 1 week | 3 weeks | extensive | 12 weeks |
| skuld | — | 3 weeks | extensive | extensive | 6 weeks |
| Class A workloads (W1–W4) | — | — | 5 weeks | 14 weeks | 12 weeks |
| Class B (Contextus + CTH) | — | — | 6 weeks | 8 weeks | extensive |
| Class C (Carts) | — | — | 4 weeks | 8 weeks | — |
| Lean Phase 2 | — | — | 3 weeks | 8 weeks | — |
| **Total dev-weeks** | **~2.5** | **~8** | **~30** | **~50** | **~37** |

**Caveats:**
- These are *focused work-week* equivalents, not calendar weeks. Multi-AI parallelism + James's architectural decisions are not included in the dev-week count.
- Hardware acquisition (Walk-tier, FPGA, silicon) is not in the dev-weeks count; treat as separate gating items.
- Coordination overhead is not included; expect 20–30% multiplier on real wall-clock.

---

## 9. The first 5 actions (this week / next 2 weeks)

If we want to move from "plan exists" to "plan in flight," the first concrete moves:

1. **C-13 — draft Skuld v1.1.** Simplest of Phase 0; absorbs the Hammer Review amendments; ~1 day of writing. Output: `Skuld-Spec-v1.1.md` in the Archive.

2. **C-14 — draft Wyrd Query API.** Specifies the language-and-semantics gap surfaced in `Wyrd-Workload-ISA-v0.2.md` §3.5; ~2 days. Output: `Wyrd-Query-API-v1.0.md`.

3. **BMA Step 0 — draft governance document.** This is the longest-pole gate item. Without it, Step 9 cannot fire. Without Step 9, no BMA, no implementation. **James should draft this in parallel with the substrate specs above; weeks of lead time before review and finalization.**

4. **C-19 — Per-Workload Precision-Mode Contract.** Quick win (½ day); resolves a known gap; cited by W4 deployment.

5. **Spec corpus index update.** `Wyrd-Corpus-Index-v1.0.md` should list the three new spec docs from this session (Proofs Reference, BMA-Workload-Performance-Spec, Workload-ISA v0.2) plus whatever Phase 0 produces. ~½ hour.

After these five, the next phase of work parallelizes broadly: substrate code (C-03 / C-08 / C-06), more Phase 0 specs, BMA Crawl Steps 3–9 in their own track.

---

## 10. What this plan optimizes for

- **MVP velocity** — the earliest demonstrable end-to-end stack (Phase 1 → MVP) is reached on the shortest reasonable path.
- **Parallel dependability** — independent streams (specs, BMA Crawl, Lean Phase 2) don't block each other.
- **Honest about uncertainty** — hardware acquisition, ROCm, vendor silicon all have explicit gating roles.
- **Per-class budget proportionality** — Class A (compute) gets ISA + silicon investment; Class B (graph) gets Wyrd / NATS / CTH investment; Class C (orchestration) gets cart-switch / capability / judge-collective investment. No single optimization fits all three.
- **Lean as moving foundation** — Phase 1 Lean is done; Phase 2 Lean runs in parallel with implementation. New theorems land as their pre-conditions become testable.
- **Spec-first, code-second** — every implementation step has a spec it cites. Spec gaps are tracked as their own tickets and closed before related code starts.

---

## 11. What this plan does NOT cover

- The BMA cognitive architecture itself (governed by the BMA Crawl chain in CLAUDE.md; the Wyrd / Skuld / qbpcu stack provides BMA's substrate, but BMA's own intelligence / personality / ethics layer is its own project)
- The Hammer simulation use case (assumed available for testing)
- Branch A vs Branch B substrate uncertainty (governed by `Wyrd-BranchA-Contingency-v0.1.md`; this plan commits to Branch B per the H1 hedge strategy)
- Specific business / partnership / funding arrangements (not technical plan)
- The QBP physics theory itself (separate research programme; CTH paper is the closest formalization)

---

## 12. Open questions for James before kickoff

These need decisions before the plan can fully execute:

1. **Phase 0 ordering preferences.** Of the 9 Phase 0 spec tickets, is there a particular order? Default suggestion: C-13 → C-19 (quick wins) → C-14 + C-15 + C-17 (substrate-spec foundation) → C-16 + C-23 (cart / capability) → C-18 + C-24 (workload-specific).

2. **MVP target date.** What's the realistic target for Phase 1 acceptance gate? Default suggestion: 6 months from BMA Step 9 completion.

3. **Hardware acquisition timing.** When does Walk-tier hardware get acquired? Gates Phase 2 (W3, Contextus prototypes need more memory).

4. **AI-collaborator role assignment.** Who drafts which specs? Who implements which components? Default: James drafts spec; Claude implements + reviews; Gemini does theory + review.

5. **Phase 2 depth target.** When Walk gate fires, do we proceed all the way to Phase 3 (real silicon), or pause for stabilization? Default: proceed if hardware is in place.

6. **Public release strategy.** Some artifacts (Lean corpus) are open-source-ready; others (Skuld supervisor) might be internal-first. Per the project, there's no current commitment.

---

## 13. Future revisions to this plan

This is v1.0; revisions expected as implementation reveals:
- Spec gaps not yet identified
- Performance bottlenecks not yet measured
- New architectural decisions (e.g., Branch A revisit triggers, sandwich-as-syscall promotion timing)
- Hardware availability changes (when Walk-tier hardware lands; when SiFive X280 arrives)

Revisions track in this Archive directory; convention: `Wyrd-Implementation-Plan-v{major.minor}.md` with the same versioning policy as the rest of the corpus.

---

## Attribution

QBP framework: Furey, Dixon, Günaydin / Gürsey, Boyle / Farnsworth, Singh, Chamseddine / Connes, Koide, Baez. CTH framework: Shannon, Dempster-Shafer, Pearl, Newman, Huet, Berge, Jirousek-Shenoy. Cynefin domain framing: Snowden. The Crawl/Walk/Run/Sprint phasing follows the BMA project tradition; the C-01..C-12 ticket structure originates in `Wyrd-Supervisor-Architecture-v0.2.md`.

---

*End of Wyrd Implementation Plan v1.0.*
