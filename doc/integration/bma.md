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
real QBP-CU silicon). The arithmetic for `Weight.Components` at
`TierQuaternion` will eventually dispatch to qbp-emulator's `Gearbox.Mul`
(QW64 fast path) or QW128 DD path (deferred per
[qbp-compute-unit#1][qcu-1]). Wyrd's Go layer remains backend-agnostic;
the dispatch happens in a future `compute/quaternion.go` file.

## Open questions

- BMA's `episodic`, `semantic`, `archetypal` tiers — does each get its
  own `model.Graph` (clean separation, expensive promotion) or are
  they distinguished only by `Node.Type` (cheap, requires filter
  predicates)? Lean's `Bridge` proves count-preservation across two
  graphs; the multi-graph approach has formal soundness for free.
- Skuld's `cart` enum (theory / engineering / beekeeper / domain-specific
  per `Wyrd.Cart`): how is it surfaced in the Wyrd Go API? Currently
  not modelled in `compute/`; deferred to a future `compute/cart.go`.

[qcu-1]: https://github.com/JamesPagetButler/qbp-compute-unit/issues/1
