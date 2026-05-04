# Integrating Wyrd with Contextus

Status: target for Contextus Walk-phase. Contextus is the cross-domain
pattern-matching layer; Wyrd is its storage substrate for InsightSignals
and the source side of the Contextus → CTH bridge.

## What Contextus gets from Wyrd

- A typed hypergraph for InsightSignals (the cross-domain pattern
  candidates produced by Edge Scout, Corpus Scout, and Bridge Agent).
- Atomic promotion to CTH via `compute.Bridge.Promote`.
- Reusable consistency primitives for the multi-source claim agreement
  case (where Contextus has detected a candidate cross-domain match
  and needs to evaluate whether the agreeing chains are consistent
  before promoting).

## Mapping Contextus types to Wyrd types

| Contextus concept | Wyrd type | Notes |
|---|---|---|
| InsightSignal | `model.Node` of `Type = "contextus.signal"` |  |
| Cross-domain match candidate | `model.Hyperedge` connecting InsightSignals from N domains | Arity ≥ 3 → Theorem 2 irreducibility applies |
| Edge Scout / Corpus Scout / Bridge Agent | not stored — runtime services that produce edges |  |

InsightSignals typically live at `TierComplex` (scalar evidence) or
`TierQuaternion` (when the signal carries a phase or polarisation
attribute from a QBP-domain source).

## Soundness citations Contextus gains

- Bridge promotion to CTH: `Wyrd.Bridge.bridge_promote_preserves_count`
  guarantees no signal is lost in transit.
- N-way agreement: when Contextus Agent flags a candidate match where
  N ≥ 3 InsightSignals from independent domains agree, the underlying
  hyperedge carries irreducible joint information per
  `Wyrd.HolographicHypergraphHigherArity.theorem2_irreducibility_n_arity`.
  This is the formal substrate for "synergy" claims in the matching
  algorithm.
- Triangle consistency on three-domain agreements:
  `compute.TriangleAdditive` and `TriangleMultiplicative`.

## Sketch

```go
import (
    "github.com/JamesPagetButler/wyrd/model"
    "github.com/JamesPagetButler/wyrd/compute"
)

// Add an InsightSignal as a node:
sig := model.Node{
    ID:   model.NodeID("contextus:signal:" + signalID),
    Type: model.NodeType("contextus.signal"),
    Tier: model.TierComplex,
    Created: signal.DetectedAt,
}
g.AddNode(sig)

// Add a 3-domain match as a hyperedge:
match := model.Hyperedge{
    ID:      model.HyperedgeID("contextus:match:" + matchID),
    Nodes:   []model.NodeID{sig1, sig2, sig3},
    Weight:  model.NewComplexWeight(score, 0),
    Created: time.Now(),
}
g.AddHyperedge(match)

// When the match is judged Bridge-eligible, promote into CTH:
br := &compute.Bridge{Source: contextusGraph, Destination: cthGraph}
if err := br.Promote(match.ID); err != nil {
    return fmt.Errorf("contextus: bridge: %w", err)
}
```

## Open questions

- Contextus's three agent types (Edge Scout, Corpus Scout, Bridge Agent
  per Theory v1.4 / Spec v1.2) — do they all author edges through the
  same `wyrd.model.Graph`, with `Node.Type` distinguishing source? Or
  does each agent have its own graph that's merged later?
- The "fish-on-the-line" detection (cross-domain InsightSignal flagged
  by Edge Scout) is currently hypothesis-tier — should it be added to
  Wyrd at all before reaching the Bridge Agent's confidence threshold?
  Tradeoff: storing every candidate gives full provenance vs. inflates
  the graph with low-quality edges. Walk-phase Contextus needs to
  decide.
