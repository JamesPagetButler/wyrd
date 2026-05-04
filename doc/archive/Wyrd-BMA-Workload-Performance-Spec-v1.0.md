# BMA Target-Workload Performance Specification

## Mapping the Four Strategic Challenges to QBP-CU ISA, Throughput Contracts, and Proof Citations

**Helpful Engineering — Quaternion-Based Physics Programme**
Principal Investigator: James Paget Butler
April 2026 — Rev 1.0

> **Purpose.** The existing `Wyrd-Workload-ISA-v0.1.md` analyzes five generic workload categories (particle dynamics, field evolution, correlation, algebraic prediction, quantum). The four **BMA target challenges** — biosynthetic manifold mapping, financial vortex detection, HAMA logistics, unified energy systems — are *concrete instances* of those categories. This document is the bridge: per-workload, what does the inner loop look like, which ISA instructions dominate, what throughput / latency / precision is required, which privilege ring the work belongs in, and which Lean theorems certify safety. Use this as the contract when planning lane allocation, pipeline depth, and capability flow at the supervisor level.

> **Audience.** Whoever builds the Go simulator (`qbpcu` package), the Skuld supervisor, or the `wyrd` hypergraph engine; whoever sizes a deployment for a specific BMA challenge; the reviewer auditing whether a proposed deployment is sound.

---

## 0. The four BMA target challenges (one-line summary)

| # | Challenge | Source doc | Problem |
|---|---|---|---|
| W1 | **Biosynthetic Manifold Mapping** | `stratagy/biosynthetic-manifold-mapping.md` | Drug-receptor docking without years of Monte-Carlo |
| W2 | **Financial Vortex Detection** | `stratagy/financial-vortex-detection.md` | 12-hour-ahead prediction of liquidity crises |
| W3 | **HAMA Logistics & Supply Chain** | `stratagy/hama-logistics-supply-chain.md` | 14-day-ahead bottleneck / bullwhip detection |
| W4 | **Unified Energy Systems** | `stratagy/unified-energy-systems.md` | 100 μs plasma-instability early warning + grid stabilization |

All four share the QBP epistemic shape: replace heavy iterative simulation with **algebraic resonance** evaluated on a quaternion-native substrate, backed by privilege-ring guarantees so the action loop can run at machine speed.

---

## 1. Cross-cutting workload pattern

Before per-workload detail, the common shape:

| Property | Value | Implication |
|---|---|---|
| Quaternion fraction | 95–99% | QFMA / QSAND / QNORM dominate; Fano-LUT (QPERM family) is the algebra-boundary tier only |
| Inner-loop critical instruction | **QFMA** | All four are FMA-dominated; pipeline this aggressively |
| Transformation primitive | **QSAND** | Three of four (W1 docking, W3 frame changes, W4 Hopf-locale) saturate this |
| Resonance / RI metric | **QNORM** | Universal — binding affinity, vorticity, bottleneck score, instability magnitude all reduce to a norm |
| Precision tier | mixed | W1 (QW1024) / W4 (QW256) push high; W2 (QW128) sufficient; W3 (QW1024 for holon aggregates) |
| Privilege ring centroid | ℍ (supervisor) | The "physics" lives in ℍ; user-ring (ℂ) wraps it; kernel-ring (𝕆) gates automated action |
| Hard real-time? | W4 only (100 μs) | Pipeline depth / event-rate budget tightest for plasma |
| Streaming vs batch | W2 streaming, W4 streaming, W1/W3 batch+iterative | Affects Skuld API surface (see §6 gap analysis) |

The existing 10-instruction ISA (Tier 1: QFMA / QSAND / QNORM; Tier 2: QPERM / QPERMR / QNEAR; Tier 3: QDEC / QREC; Tier 4: QCOMM / QRING) covers all four workloads' arithmetic needs. **No new instructions are required for any of W1–W4.** The work is in: privilege flow, precision selection, capability lifetime, and event-rate budgeting.

---

## 2. W1 — Biosynthetic Manifold Mapping (drug-receptor docking)

### 2.1 Inner loop

Per drug-receptor pair, the docking loop is:

```
for each candidate orientation θ ∈ SU(2)-grid:
    drug_oriented = QSAND(θ, drug_quaternion)        # rotate drug
    for each atom i in drug:
        force_i = atom_force(drug_oriented[i], receptor_field)  # 2-10 QFMA
        energy_total = QFMA(force_i, displacement_i, energy_total)
    affinity = QNORM(energy_total)                    # binding score
    if affinity > best: best = affinity, best_θ = θ
```

Per orientation: ~500–2000 atoms × (8 QFMA + 1 QSAND + 1 QNORM amortized) ≈ 5000–20000 cycles at 1 QFMA/cycle.
Per drug-receptor pair: 10²–10⁴ orientations sampled × per-orientation cost.

### 2.2 Throughput / latency targets

| Target | Value | Rationale |
|---|---|---|
| Per-CU-lane | 10⁵ docking attempts / sec | At ~10⁴ cycles each, fits in 10⁹ cycles/sec @ 1 GHz |
| Per-deployment | 20 lanes (one per amino acid class) | Validate all 20 amino acids in parallel |
| Latency per query | < 100 ms | Theorist asks "what does this predict?" — interactive turnaround |
| Wall-clock for full ensemble | < 1 hour | A library of 10⁵ candidate ligands, fully scored |

### 2.3 Precision / encoding

- **QW1024** for receptor surface (per the strategy doc — capture "Bark Smell" detail without losing whole-cell context)
- **QW128** for drug quaternion and per-atom forces (sufficient resolution)
- fp32 throughout per `qbp_ctl` default; the noise floor (~3·10⁻⁶ at unit M) is well below the binding-affinity threshold of interest (~10⁻³ kcal/mol normalized)

### 2.4 Privilege flow

| Operand | Ring | Why |
|---|---|---|
| Drug structure (user input) | ℂ user | A scalar molecular description |
| Receptor surface (database) | ℍ supervisor | Has orientation / spin character |
| Sandwich rotation | ℍ supervisor | The QSAND lives in ℍ |
| Result (binding affinity) | ℂ user | A scalar score returned to caller |

**Capability flow.** The drug-discovery user process holds a `Capability ℍ` (issued by Skuld at session-start, lifetime = research session). It calls `WyrdSubmit` which executes the QSAND-heavy inner loop in supervisor ring, then projects the affinity scalar back to ℂ for the caller.

### 2.5 Proof citations

- `Capability.capability_grants_safe_access` — the researcher with the ℍ-capability can perform supervisor-ring docking arithmetic on drug+receptor data without forging quaternions
- `Projection.kernel_supervisor_safe` — when the docking result is projected from ℍ back to ℂ for return to user code, no corruption (the affinity scalar that comes out is the affinity scalar that was computed)
- `NoiseBound.fp32_noise_unit_magnitude` — the binding-affinity discrimination threshold is comfortably above the fp32 noise floor

### 2.6 Implementation phase

**Walk** (per the strategy doc). Requires: QBP-CU `Mock` for prototype, then `Golden` for validation. Real silicon (`RTLShim`) is Run-phase. The drug-discovery use case unblocks once C-03 (qbpcu Go package) ships.

---

## 3. W2 — Financial Vortex Detection (RI events in markets)

### 3.1 Inner loop

Streaming, two layers:

```
# Layer 1 — per-trade ingestion (high rate, light compute)
on each trade:
    flow_quaternion = encode_trade_as_quaternion(price, volume, direction, time)
    flow_field = QFMA(flow_quaternion, kernel, flow_field)   # accumulate

# Layer 2 — sliding correlation against historical RI patterns (lower rate, heavy compute)
every 100 ms:
    for each historical_pattern in {2008, 1929, ...}:
        correlation = sliding_QFMA_chain(flow_field, historical_pattern)  # 10³–10⁵ QFMA
        peak = QNORM(correlation)
        if peak > ri_threshold: emit RI_alarm
```

Per-trade: ~10 QFMA. Layer 2 inner loop: ~10⁵ QFMA per pattern, ~10 patterns → 10⁶ QFMA per 100 ms window = 10⁷ QFMA/sec sustained.

### 3.2 Throughput / latency targets

| Target | Value | Rationale |
|---|---|---|
| Per-trade latency | < 1 ms | Real-time ingestion from exchange feeds |
| RI-window latency | < 100 ms | The "snap" signal must propagate to the rebalance trigger fast enough |
| Sustained throughput | 10⁷ QFMA / sec | Layer-2 dominant load |
| Per-CU-lane sufficient | 1 lane (with margin) | Not bandwidth-limited for a single market |
| Per-deployment | 4–8 lanes | One per major asset class (equities, FX, fixed-income, crypto) |
| Lookahead horizon | 12 hours | Per the strategy doc — market crash detection precedes price drop |

### 3.3 Precision / encoding

- **QW128** is sufficient — financial data has limited inherent precision; doubling won't help
- **fp32** with the privacy headroom — the RI signal we're detecting (vorticity peak) is many orders of magnitude above the fp32 noise floor; no incentive to pay for fp64

### 3.4 Privilege flow

| Operand | Ring | Why |
|---|---|---|
| Trade data (exchange feed) | ℂ user | Scalar prices, volumes |
| Flow field (vorticity-aware) | ℍ supervisor | Quaternion-valued, has rotational ("spin of capital") character |
| Historical pattern library | ℍ supervisor | Same — these are quaternion-valued snap signatures |
| RI alarm trigger | 𝕆 kernel | If the alarm causes automated rebalancing, that's a kernel-ring action — kernel mediates the side-effect |
| Compute-only (no rebalance) | ℍ supervisor | Pure analysis output that the supervisor publishes back to ℂ |

**Capability flow.** The exchange-feed ingest service runs in user ring (ℂ — it produces scalars). The vortex-analysis engine runs in supervisor ring (ℍ — it consumes user data via projection-up + processes in ℍ). The optional automated-rebalance subsystem holds a `Capability 𝕆` (issued by James / SRE explicitly, narrowly scoped) — without it, alarms are advisory only.

### 3.5 Proof citations

- `Capability.no_capability_means_no_synthesis` — a user-ring analyst cannot fabricate a flow-field quaternion; they can only consume what the supervisor publishes (this is the "advisor doesn't trade" guarantee)
- `Capability.wider_capability_subsumes_narrower` — if the rebalance subsystem holds 𝕆, it can also project to ℍ for read access (no need to issue both capabilities)
- `NoiseBound.fp32_noise_unit_magnitude` — the RI alarm threshold is set at ~10⁻⁴, well above the 3·10⁻⁶ fp32 noise floor; **alarm firings are real, not noise**
- `Foundations.no_surjection_complex_to_quaternion` — guarantees the analyst-as-user model is structurally honest; the analyst literally cannot construct ℍ-ring trades

### 3.6 Implementation phase

**Walk**. Streaming-friendly, so the dependency is the v1.1 Skuld-Spec amendments per `Skuld-Spec-Hammer-Review.md` — specifically `QBPSubmitStream` and `Stream.NextBatch`. **Build order:** Skuld v1.1 → qbpcu Golden → financial pilot.

---

## 4. W3 — HAMA Logistics & Supply Chain

### 4.1 Inner loop

Two coupled loops:

```
# Loop 1 — holon momentum update (per shipment cluster, hourly)
for each holon h ∈ holons:                          # ~10⁴ holons globally
    h.momentum_quaternion = QFMA(h.acceleration, dt, h.momentum_quaternion)
    h.position = QSAND(transit_frame, h.momentum)   # transform to local frame

# Loop 2 — pairwise resonance checks (the bottleneck detector, hourly)
for each (h_i, h_j) ∈ holon_pairs:                   # ~10⁸ pairs in dense regions
    resonance = QFMA(h_i.state, h_j.state, 0)        # = QMUL
    bottleneck_score = QNORM(resonance)
    if bottleneck_score > threshold: emit bottleneck_alarm
```

Per-holon update: ~5 QFMA + 1 QSAND + 1 QNORM = ~10 cycles.
Pairwise checks: 10⁸ × 1 QMUL + 1 QNORM = 2 × 10⁸ ops per hourly window.

### 4.2 Throughput / latency targets

| Target | Value | Rationale |
|---|---|---|
| Hourly batch wall-clock | < 5 min | Update lag < 1 hour |
| Pairwise resonance | 2·10⁸ ops in 5 min = ~7·10⁵ ops/sec | Per-CU-lane easy |
| Per-deployment | 4 lanes (dense regions: NA, EU, AS, intercontinental) | Region-parallel |
| Foresight horizon | 14 days | World-line intersection prediction extends well beyond the hourly window |

### 4.3 Precision / encoding

- **QW1024** for holon aggregate state — encoding 10⁴ shipments as one register requires high precision to preserve momentum integrity
- **QW128** for individual shipment events
- **fp32** sufficient — supply-chain data has timestamp granularity ≥ 1 hour for most decisions; noise floor is far below decision threshold

### 4.4 Privilege flow

| Operand | Ring | Why |
|---|---|---|
| Individual shipment events | ℂ user | Scalar receive/dispatch records |
| Holon aggregate state | ℍ supervisor | Quaternion-valued; momentum has a rotational character (the "swirl" of regional flow) |
| Bottleneck score / world-line projection | ℍ supervisor | Computed in supervisor ring |
| Automated routing change | 𝕆 kernel | Same pattern as W2 — automated action lives in kernel |
| Advisory output to dispatch | ℂ user | Projected back to scalar advisories |

**Capability flow.** Logistics analysts hold ℍ. The optional automated-rerouting subsystem (e.g., in response to a confirmed port strike) holds 𝕆.

### 4.5 Proof citations

- `Projection.kernel_supervisor_safe` — when the kernel-ring routing-change is computed and projected back to ℍ for the supervisor to publish, no corruption
- `Capability.wider_capability_subsumes_narrower` — the rerouting subsystem with 𝕆 can also read ℍ-ring data
- `Foundations.no_surjection_assoc_to_nonassoc` + `associator_octonion_witness` — supervisor cannot fabricate kernel-ring routing decisions; routing requires the explicit 𝕆-capability

### 4.6 Implementation phase

**Walk** (per strategy doc, status: STRUCTURAL). Less time-critical than W2/W4. Standard Skuld v1.0 + qbpcu Golden suffices.

---

## 5. W4 — Unified Energy Systems (plasma + grid)

### 5.1 Inner loop

```
# Plasma-stability loop, runs every Δt = 100 μs
for each cell c ∈ plasma_grid:                          # ~10⁶ cells
    field = QFMA(neighbor_couplings, c.state, field)    # 4-6 QFMA per cell
    c.next_state = QSAND(magnetic_frame, c.state)       # Hopf-locale transform
    c.vortex_magnitude = QNORM(c.next_state)
    if c.vortex_magnitude > ri_threshold:
        emit instability_warning(c, severity)            # → kernel acts within Δt
```

Per cell: ~6 QFMA + 1 QSAND + 1 QNORM = ~10 cycles.
Per cycle (100 μs): 10⁶ cells × 10 cycles = 10⁷ ops in 100 μs = **10¹¹ ops/sec sustained**.

### 5.2 Throughput / latency targets

| Target | Value | Rationale |
|---|---|---|
| Cycle frequency | 10 kHz | 100 μs window per the strategy doc |
| Per-cycle wall-clock budget | < 100 μs | Hard real-time — instability can damage the substrate |
| Sustained throughput | 10¹¹ ops/sec | Per cell × cells × frequency |
| Cluster size | 16–64 lanes | Single CU at 1 GHz × 4-wide vector ≈ 4·10⁹ ops/sec; need 25× → 25–64 lanes |
| Pipeline depth budget | < 100 cycles per QFMA-chain | Otherwise hard real-time fails |

**This is the only one of the four with a hard real-time deadline.** Plasma misses lead to substrate damage.

### 5.3 Precision / encoding

- **QW256** per the strategy doc — sub-atomic precision at reactor scales
- **fp32 with caution.** At plasma magnitudes (M ≈ 10 for normalized energies), the fp32 noise floor is ~3·10⁻³ per `NoiseBound.fp32_noise_decimal_magnitude`. The instability-warning threshold has to sit at 10⁻¹ or higher to maintain a 30× safety margin. Alternative: **fp64 for the W4 deployment**, configurable via the `qbp_ctl.PRECISION` field
- The watchdog event rate could be high during instability onset — see §6 gap on event-rate budget

### 5.4 Privilege flow

| Operand | Ring | Why |
|---|---|---|
| Sensor data (plasma probes) | ℍ supervisor | Physical quantities — currents, fields are intrinsically quaternion-valued |
| Field evolution | ℍ supervisor | The bulk compute |
| Magnetic-frame Hopf-locale transform | ℍ supervisor | QSAND in supervisor ring |
| Instability warning | 𝕆 kernel | Triggers confinement adjustment — direct hardware action |
| Confinement command | 𝕆 kernel | Issued from kernel ring; the magnetic-coil control loop is a kernel-ring service |
| Operator dashboard | ℂ user | Visualization; receives projected scalars |

**Capability flow.** The plasma-sensor integration service runs in ℍ (it has the supervisor-ring authorization to read magnetic probes). The confinement-control subsystem holds 𝕆 (a permanent capability issued by the reactor's safety system, not James — this is real-time-safety territory). The operator dashboard runs in ℂ.

### 5.5 Proof citations

- `Capability.sandwich_preservation_associative` — the Hopf-locale transformation is exactly the sandwich operation; this theorem certifies it preserves invariants under the magnetic-frame inversion
- `Projection.kernel_supervisor_safe` — when the kernel-issued confinement command is reflected in supervisor-ring sensor readings, the round-trip is corruption-free
- `NoiseBound.fp32_noise_decimal_magnitude` — calls out the fp32 limitation at M=10; informs the precision-mode decision (fp64 for W4 reactor deployments)
- `NoiseBound.threshold_separation_safe` — the contract the supervisor uses to set its alarm thresholds: `ε_priv ≥ 30 · associator_noise_bound`

### 5.6 Implementation phase

**Walk → Run.** The 100 μs hard real-time deadline pushes this past the Crawl/Walk simulator-only phase. Path: Mock prototype on small grids, Golden for cycle-accurate validation, RTLShim for latency-realistic hardware-in-loop, then real silicon. **W4 is the workload that justifies the Phase 4 SiFive X280 integration** — the others can run on FPGA.

---

## 6. Spec gaps surfaced by this review

The existing corpus covers the math, the ISA, and the high-level architecture. Gaps the four target workloads expose:

### Gap 1 — Streaming API surface (W2, W4)

`Skuld-Spec v1.0` is built around request-response (`WyrdSubmit`). W2 (financial) and W4 (plasma) need streaming. The Hammer Review already drafted `QBPSubmitStream`, `Stream.NextBatch`, and `WyrdPrepare` for v1.1. **Action: ship Skuld v1.1 before W2/W4 deployment work begins.** ETA per Hammer Review: ~1 day to amend the spec; implementation lands with C-03.

### Gap 2 — Per-workload precision-mode contract

`qbp_ctl.PRECISION` exists in the SiFive spec but is not formalized in a per-workload contract. W4 needs fp64; W2/W3 are fine with fp32. **Action: extend `Wyrd-Workload-ISA-v0.1.md` with a per-workload precision table referencing `NoiseBound` theorems.** ~½ day.

### Gap 3 — Hard real-time latency contract (W4)

The current corpus has no formal latency contract. W4's 100 μs deadline is unspecified anywhere except the strategy doc. **Action: add `Skuld-Spec §HardRT` defining: (a) a deadline-tagged variant of `WyrdSubmit`, (b) the latency-budget formula relating pipeline depth × instruction count, (c) the watchdog-event rate budget per cycle.** ~1 day.

### Gap 4 — Watchdog event rate per workload

Workload-ISA v0.1 §3.5 marks "Watchdog event rate as supervisor bottleneck" as still open. With four concrete workloads we can quantify:

| Workload | Estimated WDEvents/sec | Notes |
|---|---|---|
| W1 docking | < 10² | Mostly clean ℍ-arithmetic; rare boundary crossings |
| W2 finance | 10²–10³ | RI alarms are quasi-rare WDEvents |
| W3 logistics | < 10² | Hourly batch — sparse |
| W4 plasma | 10³–10⁴ | Instability onset can spike events |

The supervisor's WDEvent ring buffer (per the SiFive spec) needs to size for W4's peak rate. **Action: spec `WDEventBuffer` capacity = ≥ 100 events × CU-lane count.** ~½ day.

### Gap 5 — Cluster-scale orchestration (W4 specifically)

W4 needs 16–64 lanes. The current corpus assumes a single QBP-CU. Cluster-scale supervisor coordination (lane allocation, cross-lane WDEvent aggregation, capability scope across lanes) is not specified. **Action: reserve as a Run-phase ticket — `C-13 — Multi-CU Skuld coordination` — to be opened after C-12 lands.**

### Gap 6 — Capability lifetime / scope per workload

Currently capabilities are unmodeled in time. Each workload has different lifetime needs:

| Workload | Capability lifetime | Scope |
|---|---|---|
| W1 docking | Per research session (hours) | One drug library |
| W2 advisor | Permanent for read | All historical patterns |
| W2 rebalancer | Per market window (minutes) | Single asset class, single direction (buy or sell) |
| W3 routing | Per emergency (event-driven) | Affected lanes only |
| W4 confinement | Permanent | The plasma instance |

**Action: add `Capability.Lifetime` and `Capability.Scope` fields to the structure, and corresponding Lean theorems for revocation and scope-restriction.** ~1–2 days of formal work.

---

## 7. Lane / hardware sizing summary

Pulling §2–§5 together, a per-workload deployment table:

| Workload | CU-lane count | Frequency | Precision mode | Hard RT? | Streaming? |
|---|---|---|---|---|---|
| W1 biosynthetic | 20 | 1 GHz | fp32 + QW1024 | No | Batch |
| W2 financial | 4–8 | 1 GHz | fp32 + QW128 | < 100 ms | Yes |
| W3 logistics | 4 | 1 GHz | fp32 + QW1024 | No | Batch |
| W4 energy | **16–64** | **1 GHz** | **fp64 + QW256** | **100 μs** | Yes |

The total hardware envelope across all four simultaneous workloads: ≈ 100 lanes. A Sprint-phase deployment (per the architectural roadmap) can plausibly run all four; a Walk-phase deployment focuses on W1 and W3 (batch, single-CU); Run-phase opens up W2 and W4.

---

## 8. Build order implications

Combining the four workloads' phase requirements with the BMA Crawl→Walk→Run dependency chain:

```
Crawl (current, BMA Steps 0-9):
    — Lean corpus verified (this is Step 9 prerequisite)
    — qbpcu Mock (Phase 1) → unblocks W1, W3 prototypes
    — Skuld v1.0 baseline → W1, W3 advisory only

Walk (BMA post-72h gate):
    — qbpcu Golden (Phase 2)
    — Skuld v1.1 with streaming → unblocks W2, W4 prototypes
    — FPGA RTLShim (Phase 3) → W4 latency validation

Run (BMA federation):
    — SiFive X280 silicon (Phase 4)
    — Multi-CU orchestration (Gap 5)
    — Cluster-scale Wyrd → unblocks W4 production deployments

Sprint:
    — All four workloads in production
    — Capability lifetime / scope formalized (Gap 6)
```

**The bottleneck for any of the four challenges is the Wyrd / Skuld / qbpcu stack itself, not the math.** The math is settled (per `Wyrd-Proofs-Reference-v1.0.md`). The work ahead is implementation, in dependency order.

---

## 9. Minimum viable evaluation per workload

To know each workload is *actually* feasible on the QBP-CU before committing silicon, run a Tier-3-equivalent benchmark per workload at qbpcu.Golden level:

| Workload | MVE benchmark | Pass criterion |
|---|---|---|
| W1 | One amino acid (Glycine) energy minimization, 10⁴ orientations | Wall-clock < 1 minute on Golden + agreement with reference MD within ε_priv |
| W2 | 2008-crisis replay with 10⁵-sample window, RI prediction at 12-hour horizon | Detect the snap signature ≥ 12 h before peak, no false positives at threshold = 30·ε_priv |
| W3 | Suez 2021 closure replay with 10³ holons | World-line intersection detected ≥ 14 days before impact, < 5% false-positive rate |
| W4 | DIII-D plasma transient replay (public dataset) at 10⁶ cells × 100 μs | < 100 μs cycle latency, instability flagged ≥ 1 cycle before transition |

These benchmarks should be added to `QBP-CU-SiFive-Interface-Spec-v0.2.md` §10 (Tier 3 system-level tests) when the data sources are confirmed.

---

## 10. Open architectural questions surfaced

Beyond the spec gaps in §6, the four workloads expose three substantive architectural questions that aren't currently anywhere in the corpus:

1. **Cross-workload capability conflict.** If a single deployment runs W2 (financial advisor) and W4 (plasma confinement) concurrently, they each hold different capabilities at different rings. What's the interaction model when a W2 alert and a W4 instability fire simultaneously? Currently undefined.

2. **Substrate-uncertainty hedge per workload.** `Wyrd-BranchA-Contingency-v0.1.md` analyzes Branch A vs Branch B for the substrate. Different workloads have different sensitivities: W4 plasma is unaffected by which branch wins (it's pure Branch B field theory); W1 biosynthetic might be Branch-A-sensitive (the docking algebra differs). **Action: extend the Branch A contingency doc with a per-workload sensitivity column.**

3. **Privilege escalation pathway under emergency.** What's the protocol when the W4 confinement subsystem detects an imminent disruption that requires escalation beyond its 𝕆-capability — e.g., emergency reactor shutdown is a firmware-ring (𝕊) action. The Skuld supervisor doesn't currently model emergency-escalation; everything is statically capabilitied. **Action: open `C-14 — Emergency Capability Escalation` as a Walk-phase ticket.**

These three questions are *forward-looking*. None block the immediate next step (qbpcu Mock + Skuld v1.0 deployment for W1/W3). They need answers before W2/W4 production.

---

## 11. Next-steps summary (concrete, prioritized)

In rough cost / value order:

| # | Action | Cost | Unblocks |
|---|---|---|---|
| 1 | Skuld-Spec v1.1 (Hammer Review amendments) | 1 day spec | W2, W4 prototypes |
| 2 | Per-workload precision-mode table (Gap 2) | ½ day spec | W4 fp64 mode |
| 3 | qbpcu Mock implementation (C-03 ticket) | 2-3 weeks code | All four workload prototypes |
| 4 | MVE benchmarks defined (§9) | 1 day spec | Validation gate |
| 5 | Hard-RT latency contract (Gap 3) | 1 day spec | W4 |
| 6 | WDEvent buffer sizing (Gap 4) | ½ day spec | W4 (acute) |
| 7 | Capability lifetime / scope formalization (Gap 6) | 1-2 days formal | Production W2, W4 |
| 8 | Multi-CU coordination spec (Gap 5) | 2-3 days spec | W4 production scale |
| 9 | Branch A sensitivity per workload (open Q 2) | ½ day spec | All — substrate hedge |
| 10 | Emergency escalation protocol (open Q 3) | 1-2 days formal | W4 production |

Items 1–4 are the **immediate path** to running W1/W3 in prototype. Items 5–7 are needed for W2 production. Items 8–10 are W4 production gates. Sprint-phase deployment of all four: completion of all items.

---

## Attribution (per QBP standing rule R1)

The QBP framework stands on Furey, Dixon, Günaydin / Gürsey, Boyle / Farnsworth, Singh, Chamseddine / Connes, Koide, and Baez. The workload analysis tradition follows the AMS 2020 weather-scale interaction work and the Hurricane Milton intensification logic. The capability-mediated security model is in conversation with seL4, CHERI, and DBOS.

---

*End of BMA Target-Workload Performance Specification v1.0.*
