# Integrating Wyrd with BMA

Status: target for BMA Walk-phase. Crawl-phase BMA uses MuninnDB
directly; Walk-phase wraps MuninnDB behind the Wyrd model so that BMA
operations cite Wyrd theorems.

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
