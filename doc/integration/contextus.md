# Integrating Wyrd with Contextus

Status: target for Contextus Walk-phase, aligned with **Contextus Spec v1.2**.
Contextus is the cross-domain pattern-matching layer; Wyrd is its
storage substrate for InsightSignals and the source side of the
Contextus → CTH bridge.

## What Contextus gets from Wyrd

- A typed hypergraph for InsightSignals — the cross-domain pattern
  candidates produced from scout, correlation, and synthesis activity.
- Atomic promotion to CTH via `compute.Bridge.Promote`.
- Reusable consistency primitives for the multi-source claim agreement
  case (where Contextus has detected a candidate cross-domain match
  and needs to evaluate whether the agreeing chains are consistent
  before promoting).
- A formal soundness anchor for the "3+ domains agree" pattern via the
  higher-arity irreducibility theorem.

## Mapping Contextus types to Wyrd types

| Contextus concept | Wyrd type | Notes |
|---|---|---|
| InsightSignal | `model.Node` of `Type = "contextus.signal"` | Carries a `SignalSource` sub-attribute (see below) |
| `SignalSource` enum | embedded in `Node.Payload` as JSON | Values: `scout` \| `correlation` \| `synthesis` |
| `EvidencePointer` | `model.Node` of `Type = "contextus.evidence-pointer"` | "Where, not what" — points to a corpus locator, never embeds the evidence itself |
| Physical scope (focus area) | `model.Node` of `Type = "contextus.scope.physical"` | Spec v1.2 §4.6 (e.g., a sensor footprint, a geographic region) |
| Conceptual scope (focus area) | `model.Node` of `Type = "contextus.scope.conceptual"` | Spec v1.2 §4.6 (e.g., a topic / theory / frame) |
| Cross-domain match candidate | `model.Hyperedge` connecting InsightSignals from N domains | Arity ≥ 3 → Theorem 2 irreducibility applies |

### Ephemeral, NOT stored as Wyrd nodes

The session-scoped Contextus agents — Edge Scout, Corpus Scout, Bridge
Agent, and Synthesis — are **runtime services**, not persistent state.
They produce InsightSignals but are themselves ephemeral; they
should never be authored as `model.Node`s in the Wyrd graph.
Provenance information (which agent type produced a signal) lives on
the signal itself via `SignalSource`, not as a separate node.

InsightSignals typically live at `TierComplex` (scalar evidence) or
`TierQuaternion` (when the signal carries a phase or polarisation
attribute from a QBP-domain source).

## Soundness citations Contextus gains

- Bridge promotion to CTH: `Wyrd.Bridge.bridge_promote_preserves_count`
  guarantees no signal is lost in transit.
- N-way agreement: when Contextus flags a candidate match where
  N ≥ 3 InsightSignals from independent domains agree, the underlying
  hyperedge carries irreducible joint information per
  `Wyrd.HolographicHypergraphHigherArity.theorem2_irreducibility_n_arity`.
  This is the formal substrate for the "synergy" claim Spec v1.2 makes
  about 3+ domain confluences.
- Triangle consistency on three-domain agreements:
  `compute.TriangleAdditive` and `compute.TriangleMultiplicative`.

## Sketch

```go
import (
    "encoding/json"
    "github.com/JamesPagetButler/wyrd/model"
    "github.com/JamesPagetButler/wyrd/compute"
)

// SignalSource is the Contextus-side enum carried in Node.Payload.
type SignalSource string

const (
    SourceScout       SignalSource = "scout"
    SourceCorrelation SignalSource = "correlation"
    SourceSynthesis   SignalSource = "synthesis"
)

// EvidencePointer is the "where, not what" reference attached to a
// signal. The pointer carries an opaque locator into the corpus the
// evidence lives in; it never embeds the evidence body.
type EvidencePointer struct {
    Corpus    string `json:"corpus"`     // namespace, e.g. "edna:goa-cube-2026"
    Locator   string `json:"locator"`    // corpus-specific addressing form
    Confidence float64 `json:"confidence,omitempty"`
}

// Add an InsightSignal as a node, with provenance and evidence pointers
// in the payload.
type signalPayload struct {
    Source    SignalSource     `json:"source"`
    Evidence  []EvidencePointer `json:"evidence"`
    DetectedBy string          `json:"detected_by,omitempty"` // human-readable agent label, optional
}

payload, _ := json.Marshal(signalPayload{
    Source:   SourceScout,
    Evidence: []EvidencePointer{{Corpus: "edna", Locator: "goa-2026-week-12"}},
})

sig := model.Node{
    ID:      model.NodeID("contextus:signal:" + signalID),
    Type:    model.NodeType("contextus.signal"),
    Tier:    model.TierComplex,
    Created: signal.DetectedAt,
    Payload: payload,
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

Both prior open questions are now **answered**:

1. ~~Do agents author edges through one shared `model.Graph` or per-agent
   graphs?~~ **Answered (Spec v1.2):** agents are session-scoped and
   ephemeral; they don't author edges directly. InsightSignals authored
   by the agents are stored in a single shared `model.Graph` with
   `SignalSource` distinguishing provenance.
2. ~~Should hypothesis-tier "fish-on-the-line" candidates be persisted
   into Wyrd before the Bridge Agent's confidence threshold?~~
   **Answered (Spec v1.2 §4.4):** synthesis subscribes to
   `ctx.bridge.intervention`, `ctx.edge.boundary.*`, and
   `ctx.corpus.diversity` and mints InsightSignals only when ephemeral
   session findings warrant persistence. The "where not what"
   `EvidencePointer` discipline (Spec v1.2 §5.x) keeps the graph from
   inflating with low-quality embedded evidence.

## References

- Contextus Spec v1.2 §4.4 — synthesis subscription rule
- Contextus Spec v1.2 §4.6 — Scope Nodes (physical / conceptual)
- Contextus Spec v1.2 §5.x — Evidence Pointer Discipline
- Wyrd issue [#6](https://github.com/JamesPagetButler/wyrd/issues/6) — Contextus signal store
