# Integrating Wyrd with CTH

Status: target for `confluent-trust` **v0.2.x (Walk-phase)**. CTH **v0.1.0**
shipped 2026-05-05 ("Crawl Complete") with stdlib-only `model/` and
`compute/`; this document is the contract for the v0.2 Walk-phase
integration.

> **Repo state cross-link.** CTH `v0.1.0` is the upstream stable
> baseline; semver promises hold for that minor version. A formal
> deprecation policy is the one outstanding ask from CTH (see
> [confluent-trust#35](https://github.com/JamesPagetButler/confluent-trust/issues/35)
> and the architecture-instance handoff dated 2026-05-05).

## What CTH gets from Wyrd

- A typed hypergraph for trust-anchor inventories. CTH `model.Anchor`,
  `model.Chain`, and `model.Confluence` map to `wyrd.model.Node` and
  `wyrd.model.Hyperedge`.
- Atomic Contextus-to-CTH bridge promotion (`compute.Bridge.Promote`)
  with formal count-preservation soundness (Wyrd `Bridge` Phase 2 C-20c).
- Triangle / N-ary consistency primitives (Wyrd `compute.TriangleAdditive`,
  `TriangleMultiplicative`) for the same algebraic substrate that
  motivates the `NaryMI` synergy bonus in `compute/mutual_info.go`.

## Mapping CTH types to Wyrd types

| CTH | Wyrd | Notes |
|---|---|---|
| `model.Anchor.ID` | `model.NodeID` | Use `cth:anchor:<id>` namespace |
| `model.Anchor.Tier` (TierAxiom/Proof/Measurement/Prediction) | `model.Node.Type` | Free-form; tier-prefixed |
| `model.Chain.ID` | `model.HyperedgeID` | Hyperedge connects all anchors in the chain |
| `model.Chain.StepTypes` | embedded in `model.Hyperedge.Weight` (encoded as bytes) | Or kept side-table |
| `model.Confluence` | `model.Hyperedge` of arity ≥ 3 | The non-pair-decomposable case |

The `model.Anchor.Tier` (CTH-internal: axiom/proof/measurement/prediction)
is **not** the same as Wyrd `model.Tier` (algebraic ℂ/ℍ/𝕆/𝕊). CTH
anchors usually live at Wyrd's `TierComplex` (scalar values) or
`TierQuaternion` (anchors carrying phase / polarisation predictions
from QBP).

## Soundness citations CTH gains

Once CTH consumes Wyrd (Walk-phase):

- `compute/entropy.go::entropyFromDelta` already implements the
  monotonicity property proven by `Wyrd.CTH.cth_measurement_evidence_monotonic`.
  Add a doc comment cite.
- `compute/mutual_info.go::NaryMI` synergy bonus is the operational
  form of `Wyrd.HolographicHypergraph.theorem2_irreducibility`, with
  the CTH-domain lift `Wyrd.NaryMI.nary_mi_bonus_pos` (Phase 4 v1.5,
  landed 2026-05-04). The bonus being strictly positive for `n ≥ 3`
  with bounded chi-squared is now formally certified — a doc-comment
  citation in `mutual_info.go` is the v0.2 follow-up. Tracked at
  [confluent-trust#35][cth-issue-35].
- Programme-merge soundness (CTH issue #14, Walk-phase) lifts from
  `Wyrd.Bridge.bridge_promote_preserves_count`.

## Crawl → Walk migration sketch

CTH `v0.1.0` Crawl uses its own `store/json.go` for inventories. Walk-phase CTH would:

```go
import (
    "github.com/JamesPagetButler/wyrd/model"
    "github.com/JamesPagetButler/wyrd/store"
    "github.com/JamesPagetButler/wyrd/compute"
)

// Open a Wyrd graph that backs the CTH inventory.
g, err := store.JSONFile{Path: "cth-inventory.wyrd.json"}.Load()
// or, Walk-phase: open a MuninnDB-backed store

// Adding a CTH anchor.
node := model.Node{
    ID:   model.NodeID("cth:anchor:" + anchor.ID),
    Type: model.NodeType("cth.anchor." + anchor.Tier.String()),
    Tier: model.TierComplex,
    Created: anchor.CreatedAt,
}
if err := g.AddNode(node); err != nil { ... }

// Adding a CTH chain as a hyperedge.
edge := model.Hyperedge{
    ID:    model.HyperedgeID("cth:chain:" + chain.ID),
    Nodes: anchorsToNodeIDs(chain.Anchors),
    Weight: model.NewComplexWeight(float64(chain.Fidelity), 0),
    Created: chain.CreatedAt,
}
if err := g.AddHyperedge(edge); err != nil { ... }

// Promoting a Contextus signal into CTH:
br := &compute.Bridge{Source: contextusGraph, Destination: cthGraph}
if err := br.Promote(model.HyperedgeID("contextus:signal:" + sig.ID)); err != nil { ... }
```

## Open questions

- Where does CTH's `chain.Fidelity` (real-valued in [0,1]) live in the
  Wyrd weight? `TierComplex` with imaginary part 0 works; alternatively
  a 2-dim weight encoding (μ, σ) for fidelity-with-uncertainty.
- Does the CTH `step_type` enum need its own Wyrd `NodeType` per type,
  or is a single edge attribute sufficient?
- For confluences with >3 chains, **use one large hyperedge.** This is
  no longer ambiguous after `theorem2_irreducibility_n_arity` (Phase 4
  v1.5): the n-ary hyperedge carries information no pair-decomposition
  can encode. Splitting into pair edges is information-lossy.

## Triangle context

CTH consumption sits inside the four-corner architecture:

```
        QBP-CU (computes; emits WDEvent)
          /      \
         /        \
      Wyrd ─── BMA ─── CTH
   (substrate) (consumer) (epistemic measure)
```

At Walk-α, the WDEvent → CTH ρ_net loop (BMA implementor handoff §5)
turns ρ_net into a **live** signal of runtime algebraic integrity, not
just a static measure of the QBP programme inventory. CTH consumers
embedded in BMA gain that loop for free; standalone CTH consumers
(non-BMA) keep the Crawl-style static-inventory model.

[cth-issue-35]: https://github.com/JamesPagetButler/confluent-trust/issues/35
