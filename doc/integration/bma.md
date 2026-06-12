# Integrating Wyrd with BMA

Status: Toddle-phase primitives landed (W-Toddle-1 through W-Toddle-3,
Wyrd main as of 2026-05-15). BMA Walk-phase uses `model.ApplyBMAPolicy`
+ `Node.TierImmune` / `Node.Salience` for all hg/ writes; the `hg/` shim
delegates to Wyrd substrate (Phase B retirement per
`doc/design/hg-shim-retirement.md`).

## What BMA gets from Wyrd

- A typed hypergraph for the cognitive layers (autonomic, subconscious
  L/R, conscious A/B). Nodes are engrams; hyperedges are co-activations
  and derivation chains.
- Algebraic-privilege enforcement at the cognitive-layer boundary.
  Skuld supervisor uses `compute.CanSynthesize` to gate ring transitions.
- Bridge atomicity (`compute.Bridge.Promote`) for tier-to-tier engram
  promotion (e.g., EPISODIC → SEMANTIC during sleep cycles).

## Mapping BMA types to Wyrd types

BMA's hypergraph is the central data structure; Wyrd is its substrate.
The mapping is mostly direct:

| BMA concept | Wyrd type | Notes |
|---|---|---|
| Engram (Tier 0–4) | `model.Node` | `Type` carries BMA tier; `Tier` carries algebraic level |
| Co-activation hyperedge (Hebbian) | `model.Hyperedge` | Weight = Hebbian strength as quaternion |
| Derivation chain | `model.Hyperedge` of arity ≥ 3 | Soundness via Phase 4 irreducibility |
| Capability (per Skuld) | constructed from `compute.CanSynthesize` calls | not a stored type |
| Sleep-cycle compression | `model.Graph.RemoveHyperedge` + `AddHyperedge` | Phase 4 v1.5+ may add a dedicated compression API |

The BMA tiers (0=raw, 1=semantic, 2=conceptual, 3=archetypal, 4=identity)
are encoded in `model.Node.Type` (e.g., `bma.engram.tier-1.semantic`),
not in `model.Tier`. The Wyrd `model.Tier` corresponds to the algebraic
privilege ring used for arithmetic on the engram's weight, not to BMA's
cognitive tier.

Most BMA engrams operate at `TierQuaternion` (Hamilton arithmetic for
hyperedge weights) with select long-lived consolidated state at
`TierOctonion` (kernel-level invariants). `TierSedenion` is reserved
for firmware-level identity / governance state and should never be
authored by autonomous BMA code without a constitutional approval
(see `Wyrd.Constitutional.self_modification_requires_approval`, Phase 3).

## Tier-immunity and salience (W-Toddle-1/2 — OD-11(c))

BMA-specific node types now carry substrate-enforced tier-immunity and
salience defaults via `model.ApplyBMAPolicy`. This is the Wyrd-side
delivery of OD-11(c) (beekeeper decision, `live-test` seq=95).

### Policy table — canonical TD-4 inventory

The full 13-entry canonical mapping lives in `model/bma_policy.go`
(`bmaNodeTypePolicy`). Summary of the load-bearing entries:

| `model.NodeType` constant | String value | `TierImmune` | `Salience` |
|---|---|---|---|
| `NodeTypeBMASeed` | `bma.seed` | **true** | **1.0** |
| `NodeTypeBMALifeCertificate` | `bma.lineage.life-certificate` | true | 1.0 |
| `NodeTypeBMADeathCertificate` | `bma.lineage.death-certificate` | true | 1.0 |
| `NodeTypeBMAObservation` | `bma.observation` | false | Hebbian-modulated (initial 0.0) |
| `NodeTypeBMAParamProposal` | `bma.params.proposal` | true | 1.0 |
| `NodeTypeBMAParamTrustState` | `bma.params.trust-state` | true | 1.0 |
| `NodeTypeBMALastWords` | `bma.lineage.last-words` | true | 1.0 |
| `NodeTypeBMAEulogy` | `bma.lineage.eulogy` | true | 1.0 |
| `NodeTypeBMAIdentity` | `bma.lineage.identity` | true | 1.0 |
| `NodeTypeBMAMemorial` | `bma.lineage.memorial` | true | 1.0 |
| `NodeTypeBMAEntity` | `bma.entity` | false | 0.0 |
| `NodeTypeBMAConcept` | `bma.concept` | false | 0.0 |
| `NodeTypeBMAPattern` | `bma.pattern` | false | 0.0 |

Nodes with `TierImmune=true` survive all automatic eviction paths
(cap-per-retention-tier saturation, sleep-cycle compaction). This is
the substrate enforcement of A11 Topological Cognition decay-immunity —
the BMA hg/ shim calls `model.ApplyBMAPolicy` at every write site to
ensure constitutional invariants hold structurally.

Soundness: `Wyrd.TierImmunity.tier_immune_node_preserves_eviction`
(`lean/Wyrd/TierImmunity.lean`, PR #46) proves that a node with
`TierImmune=true` survives any `EvictionOp`, and
`tier_immune_preserved_under_eviction_sequence` extends this to
arbitrary sequences of eviction operations.

### Usage pattern — BMA hg/ shim write site

```go
import (
    "time"
    "github.com/JamesPagetButler/wyrd/model"
)

// At every BMA hg/ write site that constructs a Wyrd Node:
n := model.Node{
    ID:      model.NodeID("bma:" + localID),
    Type:    model.NodeTypeBMASeed,          // or any bma.* constant
    Tier:    model.TierComplex,
    Created: time.Now(),
    Payload: payload,
}
// Apply canonical TierImmune + Salience defaults for this NodeType.
// Idempotent; safe to call multiple times.
model.ApplyBMAPolicy(&n)

// Capability-gated add: passes I1/I3 boundary (ADR-003 §I3).
if err := g.AddNodeWithCapability(n, cap); err != nil {
    return fmt.Errorf("bma: hg: write seed: %w", err)
}
```

For NT_SEED (and all other `TierImmune=true` types), after
`ApplyBMAPolicy` the node has `TierImmune=true` and `Salience=1.0`.
No eviction path can remove it; the substrate enforces permanence
structurally per the Lean anchor above.

### Salience modulation (NT_OBSERVATION, Hebbian path)

NT_OBSERVATION nodes (`NodeTypeBMAObservation`) start at `Salience=0.0`
(decay-eligible baseline). The BMA hg/ shim updates `Salience` via
`Graph.UpdateNodeWithCapability` after Hebbian co-activation events:

```go
// Read the existing node, bump Salience, write it back.
existing, ok := g.Node(nodeID)
if !ok {
    return fmt.Errorf("bma: salience: node %s not found", nodeID)
}
updated := existing
updated.Salience = min(existing.Salience+hebbianDelta, 1.0)
if err := g.UpdateNodeWithCapability(updated, cap); err != nil {
    return fmt.Errorf("bma: salience: update: %w", err)
}
```

Ebbinghaus decay (sleep cycle) calls `UpdateNodeWithCapability` in the
same pattern, multiplying `Salience` by the retention factor. Salience
is a BMA-owned value; Wyrd stores it but does not compute Hebbian or
Ebbinghaus dynamics.

### Retention caps (NT_INSIGHT_SIGNAL cross-tenant pattern)

BMA may set per-retention-tier caps via `Graph.SetRetentionCap`:

```go
g.SetRetentionCap(model.RetentionCore,   50)
g.SetRetentionCap(model.RetentionNear,   20)
g.SetRetentionCap(model.RetentionPeripheral, 5)
```

Eviction under saturation excludes `TierImmune` nodes; among the
remainder, ascending `Salience` is evicted first. The cap-per-tier
eviction *execution* (walking the saturated tier) is a W-Toddle-2
delivery; the policy contract is set by `SetRetentionCap`.

**Cross-tenant note:** Contextus uses the same `RetentionTier` axis for
`NT_INSIGHT_SIGNAL` retention (Spec v1.3 §5.4). The `RetentionTier`
type (`lean/Wyrd/model/retention.go`) is intentionally separate from
the algebraic `model.Tier` to prevent API-level axis confusion. See
`doc/design/tier-immunity-salience.md` §1 (tier-axis disambiguation).

### Design docs and migration plan

- `doc/design/tier-immunity-salience.md` — W-Toddle-1 design, generic primitives
- `doc/design/bma-specific-schema.md` — W-Toddle-2 design, TD-4 policy table
- `doc/design/hg-shim-retirement.md` — W-Toddle-3 joint design, Phase B/C migration
  sequence; co-authored with `@bma-implementor`

Phase B (BMA `hg/` shim rewritten to delegate to Wyrd) is a BMA-side
task (`bma-systema`). Phase C (optional direct consumer migration off
`hg/`) happens at consumer discretion. See
`doc/design/hg-shim-retirement.md` §2 for the full sequence.

## Soundness citations BMA gains

- Hyperedge addition does not change the incident set of non-incident
  nodes — `Wyrd.Hypergraph.hyperedge_preserves_incident_edges` (C-20a,
  Phase 2). BMA's local-update guarantees flow from this.
- Sleep-cycle compression promotion: count is preserved per
  `Wyrd.Bridge.bridge_promote_preserves_count` (C-20c, Phase 2).
- Cart-switch atomicity: when BMA switches operating mode (theory /
  engineering / beekeeper / domainSpecific), capabilities scoped to
  the cart persist correctly per `Wyrd.Cart.capability_invariant_under_cart_switch`
  (C-21a, Phase 3).
- Self-modification gate: code updates require judge-collective
  approval per `Wyrd.Constitutional.self_modification_requires_approval`
  (C-21d, Phase 3). BMA's `judge_veto_blocks_self_modification` is the
  load-bearing safety theorem.
- **Tier-immunity permanence** (W-Toddle-1, PR #46):
  `Wyrd.TierImmunity.tier_immune_node_preserves_eviction` — NT_SEED and
  all other `TierImmune=true` nodes survive any automatic eviction
  operation, making A11 Topological Cognition decay-immunity structural.
- **Update-node capability safety** (wyrd-issue-#57, PR #82):
  `Graph.UpdateNodeWithCapability` checks both the existing node tier
  and the replacement tier inside one lock critical section (TOCTOU-free),
  so Hebbian salience bumps and Ebbinghaus decay writes cannot escape
  the I1/I3 mutation boundary.

## Crawl → Walk migration sketch

BMA Crawl-phase MuninnDB is the engram store today. Walk-phase BMA
wraps MuninnDB behind a Wyrd `store.Backend` interface (deferred
abstraction; not yet defined) so that:

```go
// Skuld can perform a privilege check before authoring an engram edge:
if err := compute.CanSynthesize(callerTier, edge.Tier()); err != nil {
    return fmt.Errorf("bma: skuld: %w", err)
}

// Sleep-cycle promotion (EPISODIC → SEMANTIC):
br := &compute.Bridge{Source: episodicGraph, Destination: semanticGraph}
if err := br.Promote(engramID); err != nil {
    return err
}
```

## Hardware backend

BMA Walk-phase will run on QBP-CU emulator hardware (or, post-Walk, on
real QBP-CU silicon).

The Wyrd-side dispatch surface is now in main as of PR #11
(2026-05-04):

- `compute.HamiltonProduct(a, b model.Weight) (model.Weight, error)` —
  Tier-aware dispatch (TierComplex / TierQuaternion inline; TierOctonion
  / TierSedenion → `ErrTierUnsupported`).
- `compute.HamiltonProductHighPrec(a, b model.Weight, prec uint)` —
  arbitrary-precision path. Currently uses `math/big.Float`; the swap
  to qbp-emulator's `Gearbox.Mul` is a one-line change pending the
  `emulator/v0.1.0` tag (waiting on lean2rom #7) and `QBP_PAT`
  cross-repo CI access. Tracked in
  [Wyrd issue #2](https://github.com/JamesPagetButler/wyrd/issues/2).

The QBP-CU integration interface is specified in
[qbp-compute-unit/doc/wyrd-integration.md](https://github.com/JamesPagetButler/qbp-compute-unit/blob/main/doc/wyrd-integration.md)
v0.2 (typed-per-width Gearbox API, tier ⊥ width orthogonality, Lean
source-of-truth pinned to qbp-compute-unit per option (b)).

## Triangle architecture

BMA sits at the consumer apex of a four-corner graph:

```
            QBP-CU (computes; emits WDEvent)
              /      \
             /        \
          Wyrd ─── BMA ─── CTH
       (substrate) (this) (epistemic measure)
```

- **QBP-CU** computes; emits `WDEvent` per algebraic op (passive in M0,
  active in M1).
- **Wyrd** is the typed-hypergraph substrate BMA holds.
- **CTH** measures epistemic health (`ρ_net`, fidelity, sediment).
- **BMA** is the consumer; sleep cycle uses Wyrd for state and CTH for
  self-monitored ρ_net.

### WDEvent → CTH ρ_net loop (Walk-α)

At M1 (Walk-α), BMA gains an active observer that drains
`cpu.WatchdogChan` and classifies each event into a CTH input type:

```
QBP-CU op execution
  └─→ WDEvent {AlgebraID, NormDelta, ZDClass, ZDIndices, …}
       └─→ BMA observer goroutine
            └─→ classify:
                 |NormDelta| > ε   →  FLAG-norm-drift-{nodeID}
                 ZDClass != NotZD  →  OBS-zd-detected-{i,j,k,l}
                 successful op     →  (no anchor; pure runtime)
            └─→ inject anchors into BMA's *Inventory
                 └─→ CTH compute.NetCompressionDetail
                      └─→ ρ_net measurement now reflects runtime
                          algebraic health
```

Wyrd's role in this loop is purely as the storage substrate: anchors
land as `model.Hyperedge` instances under a reserved namespace.

### Reserved `bma.runtime.*` namespace

Runtime-generated anchors from the WDEvent observer use the reserved
`bma.runtime.*` `Node.Type` prefix to keep them clearly separate from
authored hypergraph types:

| Anchor class | `Node.Type` |
|---|---|
| Norm-drift flag | `bma.runtime.flag-norm-drift` |
| Zero-divisor observation | `bma.runtime.obs-zd-detected` |
| (additional WDEvent classes) | `bma.runtime.<event-type>` |

A Wyrd-side constant for the `bma.runtime.` prefix is a candidate for
the next `model/` change so the namespace is enforceable rather than a
convention.

## Open questions

- BMA's `episodic`, `semantic`, `archetypal` tiers — does each get its
  own `model.Graph` (clean separation, expensive promotion) or are
  they distinguished only by `Node.Type` (cheap, requires filter
  predicates)? Lean's `Bridge` proves count-preservation across two
  graphs; the multi-graph approach has formal soundness for free.
  *(BMA implementor handoff 2026-05-05 §3 confirmed in-process Go API
  pattern; the multi-vs-single-graph choice remains open and is BMA's
  call.)*
- Skuld's `cart` enum (theory / engineering / beekeeper / domain-specific
  per `Wyrd.Cart`): how is it surfaced in the Wyrd Go API? Currently
  not modelled in `compute/`; deferred to a future `compute/cart.go`
  alongside the M1 `qbp.amode/bsel/psel` CSR work (peer-review-005).

[qcu-1]: https://github.com/JamesPagetButler/qbp-compute-unit/issues/1
