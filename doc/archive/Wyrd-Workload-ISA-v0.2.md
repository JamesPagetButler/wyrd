# Wyrd Workload Analysis & ISA Optimization — Three-Class Framework

## Designing the QBP-CU + Wyrd + Skuld Stack for Dense Compute, Hypergraph Reasoning, and Meta-Cognitive Control

**Helpful Engineering — Quaternion-Based Physics Programme**
April 2026 — Rev 0.2 — supersedes v0.1

> **Thesis (revised).** v0.1 analyzed five workload categories — particle dynamics, field evolution, correlation, algebraic prediction, quantum — and concluded the QBP-CU should be re-balanced for quaternion arithmetic with a 10-instruction ISA. That conclusion stands but is incomplete. **The QBP-CU is one of three substrates** the BMA stack uses; it dominates only one of three workload classes. v0.2 adds the missing two: hypergraph reasoning (Contextus, CTH) where Wyrd dominates, and meta-cognitive control (Carts) where Skuld + LLM orchestration dominate. The performance contract for each class is fundamentally different; a single ISA-level optimization cannot serve all three.

> **What changed from v0.1.** All v0.1 content is retained as **Class A — Dense Numerical Physics**. Two new classes are added: **Class B — Hypergraph Reasoning** and **Class C — Meta-Cognitive Control**. Cross-class interactions get their own section (§7). Six new entries land in the spec gap list (§8). The 10-instruction ISA from v0.1 is unchanged; v0.2's contribution is the API and architectural optimizations *outside* the CU.

---

## 1. The three workload classes — at a glance

| Property | Class A — Dense Physics | Class B — Hypergraph Reasoning | Class C — Meta-Cognitive Control |
|---|---|---|---|
| Worked examples | W1–W4 (biosynthetic, finance, supply chain, energy) + the 5 v0.1 categories | Contextus, CTH | Theory Cart, Engineering Cart, Carts in general |
| Dominant subsystem | **QBP-CU** (compute-bound) | **Wyrd / MuninnDB + NATS** (graph-bound) | **Skuld + LLM + CCB** (orchestration-bound) |
| Critical instruction / op | QFMA (95–99% of cycles) | Hypergraph traversal + typed-edge lookup | Capability scoping + cart-switch transition |
| Latency profile | Throughput-driven (10⁹+ ops/sec); W4 is hard-RT 100 μs | Interactive (< 100 ms p99); audit batch < 1 hour | Interactive (< 1 sec p95 cart switch); 10 Hz CCB hard cycle |
| Privilege ring centroid | ℍ supervisor with 𝕆 for automated action (W2/W3/W4) | ℍ for Wyrd queries; Bridge crossings between Contextus (ℂ producer) and CTH (ℍ evaluator) | Cross-ring; cart determines ring |
| Proof corpus coverage | **complete** (Capability + Projection + Noise) | **partial** — algebraic boundaries apply to Bridge crossings, but graph-invariant theorems missing | **mostly absent** — operational semantics not formalized |
| Optimization vectors | Pipelined QFMA; vector forms; precision tiers; cluster scaling | Indexing; metric caching; event-driven recompute; query batching; tiered locality | Warm-cart preload; capability scope persistence; CCB priority biasing; LLM context caching; mode-switch atomicity |

**Key non-obvious observation:** Class A optimizations don't help B or C. The QBP-CU is largely *idle* during Class B and C work; the bottleneck is elsewhere. A deployment running a balanced workload (some of each) needs distinct optimization budgets for each class.

---

## 2. Class A — Dense Numerical Physics (summary; full content in v0.1)

The five categories from v0.1 — particle dynamics, field evolution, correlation analysis, algebraic prediction, quantum computation — collectively constitute Class A. They share:

- **95–99% quaternion-arithmetic dominance** → QFMA is the central instruction.
- **Throughput-bound** → pipelined QFMA at 1 instr/cycle steady state, 4-quaternion vector lanes.
- **Compute-bound** → QBP-CU is the bottleneck; Wyrd is auxiliary (read inputs, store outputs).
- **Static privilege flow** → operands stay in one ring for the duration of a kernel; capability is checked at the syscall boundary, not per-instruction.

The detailed per-workload analysis for Class A is in `Wyrd-BMA-Workload-Performance-Spec-v1.0.md` (W1–W4) and v0.1 of this document (5 generic categories). v0.2 does not modify Class A; we cite it forward.

**Class A optimization toolkit:** pipelined QFMA, vector forms (VCIX, 4 quaternions per register), precision tiers (fp16 / fp32 / fp64 via `qbp_ctl.PRECISION`), QW128/QW256/QW1024 encoding via QDEC/QREC, cluster scaling for W4-class hard-RT.

---

## 3. Class B — Hypergraph Reasoning

### 3.1 Workload character

A Class B workload's inner loop is a **graph traversal augmented by occasional algebraic compute**. The traversal is the bottleneck; the algebraic compute is incidental. Specifically:

```
for each query q:
    candidate_set = wyrd.query(q.start_node, q.predicate_pattern)   # graph traversal
    for each path p in candidate_set:
        weight = accumulate_quaternion_weights(p.edges)              # small QFMA burst
        if weight.norm > threshold:
            results.append((p, weight))
    return rank_by_score(results)
```

The dominant cost is the `wyrd.query` step: hypergraph traversal across heterogeneous node types, predicate-matching against typed edges, possibly multi-hop with backtracking. The algebraic compute on the path weights — quaternion accumulation, norm computation — is a small constant per path.

Class B operations include:
- **Multi-modal data fusion** (Contextus): heterogeneous source adapters write to Wyrd; cross-modal queries traverse Wyrd
- **Confluence-point detection** (CTH): given a new derivation hyperedge, find pre-existing chains that reach the same target
- **Spatiotemporal queries** (QBP Locale): queries over quaternion-addressed nodes (e.g., "what happened in this region between t₀ and t₁")
- **Insight Signal emission** (Contextus): event-triggered when a metric crosses threshold
- **Information-theoretic metrics** (CTH): per-node η, per-edge μ, programme-level Δ and Re_e

### 3.2 Worked example — Contextus (ecosystem insight discovery)

Contextus stores ecosystem signals in MuninnDB / Wyrd as typed nodes (biological, geographic, temporal, agent-generated) connected by typed hyperedges. The Insight Signal pipeline:

1. **Source adapters** (Podman containers) ingest CSV / satellite imagery / time series → write to NATS subject `ctx.ingest.<source>` → Wyrd writer fans in.
2. **Agent doctrines** subscribe to NATS subjects, run pattern matching against Wyrd, emit Insight Signal nodes (`NT_INSIGHT_SIGNAL`) when patterns match.
3. **Bridge** promotes high-confidence signals to CTH evaluation via NATS subject `bridge.contextus.signal`.
4. **CTH** evaluates the signal as a candidate derivation chain, computes its trust metrics, returns Trust Receipt via NATS subject `bridge.cth.receipt`.
5. **Visualizer** renders the manifold with seam-strength overlay; updates as signals arrive.

**Per-query inner loop:** ~10 ms of Wyrd traversal, ~1 ms of QBP-CU compute, ~1 ms of NATS round-trip. The dominant cost is the traversal.

### 3.3 Worked example — CTH (Confluent Trust Hypergraph)

CTH evaluates the *epistemic health* of a research programme using five computable metrics over its trust hypergraph. Each metric has a different recompute cadence:

| Metric | Symbol | Inner cost | Recompute trigger |
|---|---|---|---|
| Anchor entropy | η(v) | O(1) per node — depends on tier | Tier change or new evidence |
| Edge fidelity | μ(e) | O(1) per edge — derivation noise | Edge added or measurement updated |
| Mutual information | I(C_a, C_b) | O(\|C_a\| + \|C_b\|) per confluence point | New derivation chain reaching existing target |
| Programme deficit | Δ(G) | O(\|V\| + \|E\|) — sum over nodes | Periodic batch (hourly / daily) |
| Reynolds-analogue | Re_e | O(\|chain\|) per chain | Periodic batch + on-incoherence-fire |

**Critical observation:** CTH metrics are mostly *static* for stable programmes. η for Tier-0 axioms doesn't change. μ for an edge changes only when its underlying measurement updates. The expensive computation is Δ, which is a whole-programme audit — but Δ is needed only periodically.

**This drives the optimization strategy below.**

### 3.4 Bottleneck analysis

For a typical Class B mixed workload (Contextus ingestion + CTH continuous evaluation):

| Layer | Time fraction | Optimization lever |
|---|---|---|
| Wyrd hypergraph traversal | 40–60% | Indexing strategy, locality, query batching |
| NATS message bus + serialization | 15–25% | Subject hierarchy, message size, queue depth |
| CTH metric computation | 10–20% | Caching, event-driven recompute, dirty-flag |
| QBP-CU compute (path weights, norms) | 5–10% | Vector batch + amortize CU round-trips |
| LLM agent inference (Contextus agents) | variable, 0–30% | Prompt caching, batched inference |

The QBP-CU is **5–10% of compute time** for typical Class B workloads. Investing in CU optimizations here is low-leverage. Investing in Wyrd indexing and CTH caching is high-leverage.

### 3.5 Optimizations for Class B

**B-OPT-1: Tier-locality indexing in Wyrd.**
Partition the hypergraph storage by CTH tier (0=Axiom, 1=Proof, 2=Measurement, 3=Prediction). Tier-0 nodes change ~never; Tier-3 nodes change continuously. Co-locate same-tier nodes on disk; the tier becomes a primary partition key. **Effect:** 5–10× reduction in cache misses on metric recompute, since most metric updates touch only one tier.

**B-OPT-2: Event-driven confluence-point detection.**
Currently the CTH spec implies confluence points are found by O(N²) chain-pair enumeration. Replace with: when a new hyperedge `e` is added with target `t`, check whether `t` already has another derivation chain. If yes, fire a confluence-check job *only on the matching pair*. This is **O(1) at write time, O(N) total over a programme's lifetime** instead of O(N²) per audit. Wyrd needs an inverted index `target → derivation_chains`; cheap to maintain.

**B-OPT-3: Cached static metrics with dirty-flag invalidation.**
For each Wyrd node, store the cached η alongside the node. Maintain a `dirty: bool` flag set when any incoming evidence changes. On query, if `!dirty`, return cached value; if `dirty`, recompute and cache. **Effect:** η lookup goes from O(\|chains_to_v\|) to O(1) for stable nodes, which is the common case.

**B-OPT-4: Programme-Δ as incremental update, not full recompute.**
Δ(G) = Σ η(v) over sources. If only one source-anchor's η changes, Δ updates by exactly the delta — no need to re-sum. Track Δ as a running aggregate; update on each cached-metric invalidation. **Effect:** Δ recompute drops from O(\|V\|) to O(1) for typical incremental updates. Full recompute reserved for programme-structure changes (rare).

**B-OPT-5: NATS subject partitioning by access tier.**
Contextus has access tiers (per spec §7.1). Partition NATS subjects to match: `ctx.public.*`, `ctx.internal.*`, `ctx.restricted.*`. Subscribers self-select; routing is constant-time. **Effect:** removes an authorization check from the hot path; security check happens once at subject subscription, not per message.

**B-OPT-6: Bridge as broadcast channel, not RPC.**
A single Insight Signal often goes to multiple CTH evaluators (different domain adapters). Use NATS pub/sub on `bridge.contextus.signal` instead of point-to-point RPC. **Effect:** O(consumer_count) → O(1) producer-side cost; consumers paginate at their own rate.

**B-OPT-7: Quaternion path-weight memoization.**
Hot paths in the hypergraph (e.g., the most-queried derivation chain in QBP) get queried repeatedly. Memoize the accumulated quaternion weight for the path; invalidate on any edge change. **Effect:** the small QBP-CU burst per query is amortized; for hot paths it disappears entirely.

**B-OPT-8: Vector batching of CU round-trips.**
When a single query returns N candidate paths, gather all N path-weight computations into one VCIX vector instruction (4 paths per register). **Effect:** reduces CU round-trip overhead from O(N) to O(N/4); for batches < 4, falls through to scalar.

**B-OPT-9: Tier-based query rejection (predicate pushdown).**
A Class B query that requires Tier-0 axioms should never traverse Tier-3 prediction nodes. Push tier filters into Wyrd's query planner. **Effect:** removes irrelevant traversal early; saves 30–50% on traversal time for tier-specific queries.

**B-OPT-10: Confluence-point eager firing + cache.**
When a confluence point is detected (B-OPT-2), eagerly compute its mutual information I(C_a, C_b) and cache as a property on both chains' targets. The next query asking "is this chain coherent?" returns the cached answer. **Effect:** I-metric latency drops from per-query computation to cached lookup; the eager cost is amortized over future queries.

### 3.6 Throughput / latency contract for Class B

| Operation | Target | Rationale |
|---|---|---|
| Insight Signal emission | 10²–10⁴ signals/sec | Contextus ingestion rate, dependent on data volume |
| Wyrd query latency (interactive) | < 100 ms p99 | UI/agent responsiveness |
| Wyrd query latency (batch) | < 10 sec p99 | CTH audit |
| Bridge crossing rate | 10² Trust Receipts/sec | Sustained Contextus → CTH |
| CTH Δ-recompute (incremental) | < 100 ms | Per cached-metric invalidation |
| CTH Δ-recompute (full audit) | < 1 hour for 10⁴-anchor programme | Programme-level epistemic audit |
| Confluence-point detection (per write) | < 1 ms | Constant-time at write time |
| QBP-CU utilization | 5–10% (intentional) | Class B is graph-bound, not CU-bound |

### 3.7 Privilege flow for Class B

The privilege ring assignment for Class B subsystems:

| Subsystem | Ring | Why |
|---|---|---|
| Contextus source adapters | ℂ user | Producers of scalar-typed events |
| Wyrd hypergraph store | ℍ supervisor | The privileged data layer; Skuld mediates access |
| NATS message bus | ℍ supervisor | Routing infrastructure |
| Contextus agents | ℂ → ℍ via capability | Agents need ℍ-capability to write Insight Signals |
| Bridge layer | ℍ → 𝕆 transition | The Bridge is where Contextus claims become CTH-evaluable, crossing into kernel ring |
| CTH evaluators | 𝕆 kernel | CTH evaluates the trustworthiness of the entire programme — including BMA's own claims; must be in kernel ring to prevent self-modification |
| Trust Receipts (CTH output) | ℍ supervisor (read-only to ℂ) | Published back to Contextus consumers as immutable records |

**Capability flow.** Contextus adapters hold ℂ→ℍ-capabilities (issued by the Helpful Engineering admin, scoped to their data source). The Bridge holds an ℍ→𝕆-capability (long-lived, project-scoped). CTH evaluators are kernel-ring services with no capability needed — they're inside 𝕆.

**The Bridge is the load-bearing privilege boundary** for Class B. It must be specified rigorously (see §8 gap list).

### 3.8 Proof citations + gaps for Class B

**Proven, applies to Class B:**
- `Capability.capability_grants_safe_access` — Contextus adapters with ℂ→ℍ-capability can write to Wyrd safely
- `Capability.no_capability_means_no_synthesis` — adapters without capability can't fabricate Wyrd writes
- `Projection.kernel_supervisor_safe` — CTH-computed Trust Receipts (in 𝕆) project back to ℍ for publication without corruption
- `Capability.wider_capability_subsumes_narrower` — the Bridge's 𝕆-capability subsumes ℍ-capability for read access

**Gap — Class B-specific theorems missing:**
- **Graph-invariant preservation under hyperedge addition.** When CTH adds a confluence-point edge, does it preserve the algebraic invariants of surrounding nodes? Currently no theorem certifies this. *Proposed:* `theorem hyperedge_preserves_node_invariants` — add an edge `e = (S, t)`; prove that for all `v ∉ S ∪ {t}`, the algebraic properties of `v` (specifically η(v), μ for v's incoming edges) are unchanged.
- **CTH metric monotonicity.** Some operations should be monotonic (e.g., adding evidence cannot *decrease* η for a Tier-2 node). Currently no theorem certifies this. *Proposed:* `theorem cth_evidence_monotonic`.
- **Bridge atomicity.** When the Bridge promotes an Insight Signal to CTH, the promotion is atomic (the signal is either fully promoted or not at all; no partial state visible). Currently no theorem certifies this. *Proposed:* `theorem bridge_promotion_atomic` — requires Wyrd transaction model formalized.

These gaps are not blocking — Class B can be implemented using the existing proven theorems for the algebraic boundaries, with the new theorems added incrementally as the corpus matures.

---

## 4. Class C — Meta-Cognitive Control

### 4.1 Workload character

A Class C workload is BMA's *own internal operation* — the meta-cognitive layer that decides which other workloads to run, how to allocate attention across them, and how to switch between modes (carts) over time.

The Carts are BMA's operating modes:

- **Theory Cart** — understanding problems. Sense → Analyse → Respond. Active when domain is "Complicated" (known unknowns, model needs fleshing out). Operations: hypergraph traversal, claim evaluation, judge-collective vote, document drafting.
- **Engineering Cart** — building solutions. Probe → Sense → Respond. Active when domain is "Complex" (unknown unknowns, no reliable predictive model). Operations: code construction, simulation runs, validation, hardware probes.
- **Domain-specific carts** — extensible, e.g., "Beekeeper Cart" for hardware control, "Audit Cart" for governance review.

**Cart switching is the critical-path operation** because BMA's optimal pattern is *rapid alternation*: theory proposes a micro-hypothesis → engineering builds a micro-experiment → results feed back to theory. The shorter the loop, the faster progress (per BMA-Theory-Consolidated v2.0). **Cart-switch latency is on the loop's critical path.**

### 4.2 Worked example — Theory Cart

Theory Cart's inner loop:

```
on beekeeper_input:
    interpret_input → identify domain (Complicated)
    classify_problem → match against known patterns in Wyrd
    explore_solution_space → traverse hypergraph; query CTH for trust metrics
    evaluate_candidates → judge-collective parallel evaluation
    draft_spec → produce Theory Cart artifact
    return spec to beekeeper
```

Per-cycle compute:
- Interpret + classify: ~1–2 sec LLM inference, ~50 ms Wyrd lookups
- Explore: ~10–100 ms Wyrd hypergraph traversal (Class B work spawned from C)
- Evaluate: ~5–15 sec parallel judge-collective evaluation (5–15 judges)
- Draft: ~2–10 sec LLM inference

**Total: 10–30 sec per Theory Cart turn.** Long-context LLM inference dominates.

### 4.3 Worked example — Engineering Cart

Engineering Cart's inner loop:

```
on theory_spec_arrives:
    plan_implementation → decompose into work items
    progressive_hardening → Reference design → Guidance → Requirement
    sandbox_first → Podman / Jupiter 2 / test instance
    run_simulation → spawn Class A workload on QBP-CU (e.g., qbpcu Mock or Golden)
    validate_against_spec → CTH verifies measurement matches prediction
    if real_world_action_needed: request beekeeper_permission (training wheels)
    apply_action_or_iterate
```

Per-cycle compute:
- Plan: ~1–5 sec LLM inference
- Sandbox spin-up: ~10–30 sec (Podman + warm cache)
- Simulation: variable; can be hours (W4 plasma reactor sim) or seconds (W1 single-amino-acid validation)
- Validation: ~100 ms–10 sec depending on Class B volume
- Action: blocked on beekeeper permission during Crawl/Walk; near-instant during Run

**Total: highly variable.** The cart can spawn arbitrarily large Class A workloads; the cart's *own* overhead is small.

### 4.4 Cart switching as the critical path

The performance contract for Class C centers on **how fast can BMA switch carts**? Because the optimal pattern is rapid Theory ↔ Engineering alternation, switch overhead directly limits progress.

What happens during a cart switch:
1. Current cart marks in-flight work as checkpointed (Wyrd writes flushed; in-progress Class A workloads suspended)
2. Capability scope is renegotiated with Skuld (some capabilities valid in one cart but not another; e.g., real-world action capability is Engineering-only)
3. LLM context is swapped (Theory Cart's prompt → Engineering Cart's prompt; cart-specific persona / doctrine loaded)
4. Wyrd cursor positions are restored (each cart maintains its own "open" graph queries)
5. CCB priority is updated (the new cart bumps to high priority; old cart drops)
6. New cart begins responding to beekeeper

**Cold cart switch:** all of the above happens. Worst-case ~3–5 sec.
**Warm cart switch:** state for the target cart is preserved from the last visit; only steps 2, 3, 5 happen. Best-case < 500 ms.

Optimization budget:
- **Warm-cart preservation** is the biggest lever — sub-second target only achievable if state isn't fully torn down at switch
- **LLM context caching** (Anthropic prompt caching) saves on the largest constant — the cart-specific system prompt
- **CCB priority biasing** ensures the active cart gets compute resources; idle carts are throttled
- **Capability-scope persistence** (B-OPT-7 in spirit) — if a capability is valid across all carts, don't re-issue on switch

### 4.5 Optimizations for Class C

**C-OPT-1: Warm-cart state preservation.**
Don't tear down a cart's state on switch-out; mark it dormant. On switch-in, restore from dormant state instead of cold-loading. **Effect:** cart-switch latency drops from 3–5 sec (cold) to < 500 ms (warm). Memory cost: ~100 MB per dormant cart (LLM context + Wyrd cursors). Acceptable for 2–4 active carts.

**C-OPT-2: Anthropic prompt caching for cart-specific system prompts.**
Theory Cart and Engineering Cart have stable, large system prompts (~5–15K tokens of cart doctrine). Cache via Anthropic's `cache_control` mechanism. **Effect:** first call after cache miss is full cost; subsequent calls within 5-min cache window are 90% cheaper and faster. Compatible with cart-switching pattern (rapid alternation = high cache hit rate).

**C-OPT-3: Capability scope persistence across cart switches.**
A capability issued in Theory Cart should remain valid when BMA switches to Engineering Cart, *if its scope is session-level rather than cart-level*. Skuld supports session-scoped capabilities; cart-scoped (narrower) capabilities are rare and explicitly marked. **Effect:** cart switch doesn't trigger Skuld round-trips for capability re-issue.

**C-OPT-4: CCB priority bumping on cart switch.**
The cross-channel-bus runs at 10 Hz (per CLAUDE.md). On cart switch, the new cart's priority bumps to top of the queue; the old cart drops to background. **Effect:** response latency for the new cart is bounded by 1 CCB cycle = 100 ms; old cart's in-flight work continues at low priority but doesn't starve.

**C-OPT-5: Judge-collective parallel evaluation.**
A judge-collective vote is N independent judge evaluations (per CLAUDE.md, "domain-weighted approval"). Run them in parallel via N parallel LLM calls + an aggregation step. **Effect:** latency drops from O(N) to O(1) (in wall-clock terms) at the cost of N× peak compute. Worth it for the 5–15-judge case.

**C-OPT-6: Pre-flight veto detection.**
Before running the full judge-collective evaluation, run a cheap heuristic "would any judge veto this?" check. If yes, short-circuit to MAJOR_CONCERN without paying for the full vote. **Effect:** ~30% of proposals get the cheap path (since vetoes are designed to be detectable from rule violations).

**C-OPT-7: Vote caching.**
The same proposal evaluated twice in the same beekeeper session returns the cached vote (unless context changed). **Effect:** Theory Cart re-evaluating an Engineering-Cart-rejected approach doesn't pay the vote cost twice.

**C-OPT-8: Mode-switch atomicity via Wyrd transactions.**
Cart switches happen mid-work. Wyrd needs `BeginTransaction` / `Checkpoint` / `CommitOrRollback` semantics so a switch doesn't leave Wyrd in a partial state. Spec the transaction API in Skuld v1.1 (currently a gap). **Effect:** zero dropped Wyrd writes across cart switches.

**C-OPT-9: Tier-aware memory management across carts.**
BMA's memory tiers (Tier 0–4 per CLAUDE.md) interact with cart switching. Hot tier (Tier 0/1) for current cart; warm tier (Tier 2/3) for other dormant carts; compressed (Tier 4) for fully cold carts. **Effect:** memory footprint per dormant cart is bounded; cart-switch cost amortizes the tier transition.

**C-OPT-10: Affect-signal continuity.**
The four affect signals (curiosity, confidence, anxiety, satisfaction per Theory Addendum 9.0) are part of BMA's identity, not its current task. They persist across cart switches; do not snapshot/restore. **Effect:** simpler state model + behavioral coherence (BMA doesn't "forget" it was anxious about a topic when switching carts).

**C-OPT-11: Sleep-cycle-aware cart eviction.**
During BMA-SLEEP (the consolidation phase), all carts can be safely evicted from warm state to compressed state. Wake-up restores only the cart most recently active. **Effect:** sleep-cycle memory footprint drops dramatically; only active cart needs warm preservation through the cycle.

**C-OPT-12: Beekeeper-channel bypass for instinct responses.**
Per BMA-Cognitive-Foundation §0.6, "instincts and muscle memory" are fast-path responses for familiar situations (sub-second). These should bypass cart loading entirely — if the beekeeper says "status?" BMA responds directly without consulting Theory or Engineering Cart. **Effect:** beekeeper-perceived latency for trivial questions drops from cart-switch + cart-process to a constant ~200 ms.

### 4.6 Throughput / latency contract for Class C

| Operation | Target | Rationale |
|---|---|---|
| Cart switch (warm) | < 500 ms p95 | Rapid alternation pattern |
| Cart switch (cold) | < 3 sec p95 | First entry to a cart; acceptable bootstrap cost |
| Beekeeper turn (text response) | < 2 sec p95 | Interactive responsiveness |
| Beekeeper turn (instinct) | < 200 ms p95 | Sub-second status / instinct responses |
| Judge-collective vote (parallel) | < 10 sec p95 | 5–15 judges in parallel |
| Pre-flight veto detection | < 500 ms p95 | Cheap check before expensive vote |
| CCB negotiation cycle | exactly 10 Hz, 100 ms hard cycle | Per architecture spec |
| Wyrd transaction commit | < 100 ms p99 | Cart-switch atomicity gate |
| Sleep-cycle entry / wake | < 30 sec | Consolidation overhead amortized over hours-long sleep |

### 4.7 Privilege flow for Class C

BMA itself crosses rings as it operates:

| Cart / Mode | Effective ring | Why |
|---|---|---|
| Theory Cart | ℂ user | Pure thought; no actions; reads from Wyrd; doesn't write |
| Engineering Cart (sandbox) | ℂ user | Code-in-sandbox; no real-world effect |
| Engineering Cart (real action) | ℍ supervisor + capability | Real-world action requires explicit beekeeper permission (training wheels protocol) |
| Beekeeper Cart | ℍ supervisor or 𝕆 kernel | Hardware control; fine-grained capability per device |
| BMA-SLEEP cycle | 𝕆 kernel | Sleep cycle modifies the hypergraph wholesale (compression); requires kernel ring |
| Self-modification (Code Update Lifecycle) | 𝕊 firmware | Per Theory 9.0 — code updates require firmware ring |

**Capability flow.** The beekeeper authorizes BMA at session start with a session capability (typically ℍ-scoped). Cart-switch within the session reuses this capability (C-OPT-3). Specific real-world actions require the beekeeper to grant a narrower task-scoped capability ("yes, push this commit", "yes, fire the SLURM job"). Self-modification requires the judge-collective to vote APPROVE on the change before BMA can self-write to its own code repository — this is the constitutional pin per CLAUDE.md.

### 4.8 Proof citations + gaps for Class C

**Proven, applies to Class C:**
- `Capability.capability_grants_safe_access` — when BMA-with-capability acts in Engineering Cart, the action is sound
- `Capability.no_capability_means_no_synthesis` — without the beekeeper's explicit grant, BMA cannot perform real-world actions in user-ring carts
- `Capability.wider_capability_subsumes_narrower` — beekeeper's session capability covers all carts within the session

**Gap — Class C-specific theorems missing (deeper than Class B):**

- **Capability scope persistence across cart switches.** *Currently:* C-OPT-3 above is *asserted*, not *proven*. We have no formal model of cart-as-context. *Proposed:* `theorem capability_invariant_under_cart_switch` — given a capability `c` issued at session start with session-level scope, prove that for any sequence of cart switches `s₁, s₂, ..., sₙ`, the capability `c` remains valid in every cart.
- **Cart-switch atomicity.** *Currently:* C-OPT-8 specifies a Wyrd transaction model that doesn't yet exist formally. *Proposed:* `theorem cart_switch_atomic` — given a Wyrd transaction `T` open in cart `A` at the moment of switch to cart `B`, the post-switch state is either "T fully committed" or "T not started"; no partial state is observable.
- **Judge-collective vote determinism (modulo time).** Two votes on the same proposal in the same context should return the same result. *Proposed:* `theorem judge_collective_deterministic` — formalize judge as pure function of (proposal, context); aggregation as commutative monoid.
- **Self-modification constitutional pin.** *Currently:* the constitutional pin is a *runtime check* (judge-collective must vote APPROVE before code-update), not a *formal proof*. *Proposed:* `theorem self_modification_requires_approval` — for any code update U applied to BMA's own repository, the precondition `judge_collective.vote(U) == APPROVE` is enforced unforgeable.

**These gaps are substantively new mathematical work** — they require formalizing BMA's operational semantics (state machines, transaction logs, vote aggregation), not just continuing the algebraic-privilege program. This is **Phase 2 of the Lean corpus**, distinct from the Phase 1 (algebraic boundaries) that just shipped.

---

## 5. Cross-class interactions

The three classes interact heavily in real BMA operation. Three patterns are common enough to deserve their own analysis.

### 5.1 Class C drives Class A

Engineering Cart spawns a W4 plasma simulation. Pattern:
1. Engineering Cart issues `WyrdSubmit(plasma_grid_init, qbpcu_mock)` at the kernel boundary
2. QBP-CU (Class A) executes the dense numerical workload for hours
3. Cart polls or subscribes for completion via NATS
4. On completion, cart resumes with the simulation results

**Optimization implication:** The cart should NOT block during the long-running Class A workload. Cart-switch is the natural "yield" — when the simulation is running, cart goes dormant; on completion, cart re-warms.

### 5.2 Class C drives Class B

Theory Cart explores the QBP hypergraph for a derivation chain. Pattern:
1. Theory Cart issues `wyrd.query(start, predicate)` (Class B work)
2. Wyrd traverses the hypergraph, returns ranked candidates
3. Theory Cart consults CTH on each candidate's trust metric
4. Theory Cart proposes; judge-collective evaluates; result returned to beekeeper

**Optimization implication:** Class B latency dominates the Theory Cart's response time. The B-OPT-* optimizations directly improve C-OPT performance. The Wyrd query latency budget for Class B (< 100 ms p99) IS the cart-turn budget for Class C work that uses Wyrd.

### 5.3 Class B feeds Class C

CTH detects an incoherence (Re_e spikes; a confluence-point fails). Pattern:
1. CTH (Class B) detects the incoherence
2. Trust Receipt is published to Bridge
3. The receipt has high priority — it's a programme-health alert
4. Engineering Cart subscribes to high-priority CTH alerts; receives the receipt
5. Engineering Cart kicks off a diagnostic workflow (which is itself Class C work, possibly spawning Class B queries)

**Optimization implication:** Bridge-to-Cart latency matters. Use a NATS subject reserved for Class C consumption (e.g., `bridge.cth.receipt.priority_high`); subscribers self-select on priority; cart switches to Engineering Cart on priority alert *bypasses* normal cart-switch latency and goes directly into action mode. **This is C-OPT-12 generalized — Class B alerts are also "instincts."**

---

## 6. Optimization summary — total budget

A balanced BMA deployment running W1, Contextus, Theory Cart, and Engineering Cart concurrently has the following optimization budget:

| Class | Optimization investment | Where the gains compound |
|---|---|---|
| A — dense compute | ISA features, pipelining (already locked in v0.1 ISA freeze) | W1–W4 throughput; sets ceiling for what BMA can simulate |
| B — hypergraph reasoning | Indexing, caching, event-driven recompute, query batching | Contextus query latency; CTH audit cadence; Class C cart-turn time when consulting Wyrd |
| C — meta-cognitive control | Warm-cart preservation, prompt caching, CCB priority, judge parallelism | Cart-switch latency; beekeeper-perceived responsiveness; rapid-alternation rate |

**Where to invest first:** Class C optimizations have the highest user-visible impact (cart switches are on the beekeeper's critical path, every conversation). Class B optimizations have the highest aggregate compute impact (most BMA cycles are graph-bound, not CU-bound). Class A is already paid for by the v0.1 ISA design.

---

## 7. Updated spec gap list (extends `Wyrd-BMA-Workload-Performance-Spec-v1.0.md` §6)

Carrying forward the 6 gaps from v1.0 (streaming API, precision-mode contract, hard-RT latency, WDEvent budget, multi-CU orchestration, capability lifetime/scope) and adding:

**Gap 7 — Wyrd query API specification.**
Class B operates against Wyrd, but the *query language and semantics* are not specified anywhere in the corpus. Contextus assumes them (typed nodes, typed edges, predicate matching), but they're not contracts. **Action:** add `Wyrd-Query-API-v1.0.md`. ~2 days.

**Gap 8 — NATS subject hierarchy contract.**
Class B and the cross-class interactions rely on NATS subjects (`ctx.*`, `contextus.*`, `bridge.*`). The hierarchy is partially specified in Contextus spec §5.3 but not as a system-wide contract. **Action:** consolidate into `Wyrd-NATS-Topology-v1.0.md`. ~1 day.

**Gap 9 — Cart-as-context formal model.**
Class C optimizations C-OPT-3 (capability scope) and C-OPT-8 (mode-switch atomicity) require a formal model of "cart as a context" that doesn't yet exist. **Action:** add `BMA-Cart-Operational-Semantics-v1.0.md`. ~3 days spec + 1 week Lean formalization.

**Gap 10 — Wyrd transaction model.**
Required by C-OPT-8. Should specify `BeginTx / Checkpoint / CommitOrRollback` semantics, with read/write set tracking and concurrency rules. **Action:** add `Wyrd-Transaction-Model-v1.0.md`. ~2 days spec + Lean theorem `transaction_atomicity`.

**Gap 11 — Class B graph-invariant theorems (3 proposed in §3.8).**
**Action:** Lean corpus Phase 2 work; ~1 week per theorem.

**Gap 12 — Class C operational-semantics theorems (4 proposed in §4.8).**
**Action:** Lean corpus Phase 2 work; the meatiest of the gaps; ~2–4 weeks.

---

## 8. Build order implications (revised from v0.1 §7)

| Phase | New deliverables for v0.2 |
|---|---|
| **Crawl** (current) | Lean corpus done. Proofs reference + workload spec done. **Next:** Skuld v1.1 (gap 1), Wyrd query API spec (gap 7), NATS topology spec (gap 8) |
| **Walk** | qbpcu Mock + Skuld v1.1 → unblocks W1, W3 (Class A) prototypes. Wyrd Mock + transaction model (gap 10) → unblocks Contextus (Class B) prototype. Cart-as-context spec (gap 9) → unblocks BMA's own cart switching (Class C) |
| **Run** | qbpcu Golden + RTLShim (Class A real-time validation). Wyrd Golden (Class B production substrate). Class B graph-invariant theorems (gap 11). |
| **Sprint** | Multi-CU orchestration (gap 5). Class C operational-semantics theorems (gap 12). Cluster-scale Wyrd. All four W1–W4 in production + Contextus + Theory/Engineering Carts in mature state. |

The Class B prototype (Contextus on Wyrd Mock + Crawl-phase BMA) is **achievable in the Walk phase**, in parallel with Class A prototypes. Class C polish — sub-second cart switches, fully optimized judge-collective — is **Run phase**.

---

## 9. Key insights surfaced by the three-class analysis

1. **The QBP-CU is not the whole story.** v0.1 framed BMA's performance bottleneck as quaternion arithmetic; that's true for Class A (W1–W4) but only ~5–10% of total compute time across a balanced workload. Class B (graph) and Class C (orchestration) dominate elsewhere.

2. **Cart switching is the user-visible bottleneck.** The beekeeper experiences BMA primarily through cart switches. C-OPT-1 (warm-cart preservation) and C-OPT-2 (prompt caching) are the highest-leverage user-facing optimizations.

3. **Class B's CTH is mostly cacheable.** The "expensive" information-theoretic computation is mostly *one-time* (B-OPT-3 + B-OPT-4); steady-state cost is low if caching is correct.

4. **The Bridge is a load-bearing privilege boundary** that hasn't been rigorously specified. Bridge atomicity (gap 7 + new theorem in §3.8) is required before production CTH evaluation.

5. **The Lean corpus has more work ahead.** Phase 1 (algebraic privilege boundaries) is complete. Phase 2 (graph invariants for Class B; operational semantics for Class C) is 6–12 weeks of formal work. **It's not blocking Walk-phase implementation, but it's gating Run-phase production.**

6. **Cross-class interactions deserve their own design layer.** The three patterns in §5 (C drives A, C drives B, B feeds C) are common enough that BMA's runtime should treat them as first-class flow patterns, not ad-hoc compositions.

7. **Per-class optimization budgets should be tracked separately.** A "20% performance improvement" budget for the BMA stack should be allocated across the classes proportional to where time is actually spent, not where the most exciting hardware lives.

---

## Attribution

QBP framework: Furey, Dixon, Günaydin / Gürsey, Boyle / Farnsworth, Singh, Chamseddine / Connes, Koide, Baez. CTH framework: builds on Shannon, Dempster-Shafer (Pearl's critique), Newman / Huet on confluence, Berge on hypergraphs, Jirousek-Shenoy entropy. Cynefin domain framing for cart selection: Snowden. Cayley-Dickson construction: Schafer; Baez 2002, *The Octonions*.

---

*End of Wyrd Workload Analysis & ISA Optimization v0.2.*
