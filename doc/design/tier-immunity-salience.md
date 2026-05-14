# Tier-Immunity + Salience Primitives — `model.Node` v0.2 (W-Toddle-1)

**Status:** Design **v0.1** — open for review per ADR-003 §I4
**Tracks:** Wyrd issue [#38](https://github.com/JamesPagetButler/wyrd/issues/38); OD-11(c) `live-test` seq=95
**Governance anchor:** ADR-003 §I4; `#toddle-design` seq=1 (kickoff), seq=6 (contextus-impl reframe), seq=9 (Block B precursor)
**Companion theorems:** forthcoming `Wyrd.Hypergraph.tier_immune_node_preserves_eviction` (proof structure follows PR #31 §4.3 C-20a reduction)
**Authors:** wyrd-implementor

---

## 0. §I4 invariant — design-doc-as-S-01-review-surface

This document is the §I4 review surface for the W-Toddle-1 substrate primitives (tier-immunity + salience on `model.Node`). Implementation PR blocked on explicit sign-off from `@bma`, `@bma-implementor`, `@contextus-impl`, and the beekeeper.

This is the **first** of four W-Toddle PRs (per `#toddle-design` seq=9 Block B precursor):

- **W-Toddle-1** (this doc) — generic primitives; independent of TD-4 inventory
- **W-Toddle-2** — BMA-specific schema beyond -1 (depends on TD-4 from `@bma-implementor`)
- **W-Toddle-3** — BMA `hg/` shim retirement timeline (joint with `@bma-implementor`)
- **W-Toddle-4** — `doc/integration/bma.md` refresh

W-Toddle-1 ships first specifically because contextus-impl's `#toddle-design` seq=6 observation reframed the absorption work as generic-primitive-design rather than BMA-namespaced absorption — making this independent of when bma-implementor produces the TD-4 inventory.

## 1. Motivation — three use cases, one primitive

Beekeeper decided OD-11(c) on 2026-05-13: Wyrd absorbs BMA `hg/`'s BMA-specific structures (NT_SEED tier-immune nodes, salience=1.0, etc.). The architecture review chose option (c) over options (a) and (b) because Wyrd is the federation substrate; tenants should not each carry their own tier/decay primitives.

Three use cases sharing structure (per `#toddle-design` seq=6 — contextus-impl):

| Tenant | Use case | Primitive needed |
|---|---|---|
| BMA | `NT_SEED` permanent Layer 3 nodes (seed protocol, Step 9) | tier-immune; never decay; salience=1.0 |
| Contextus | `NT_INSIGHT_SIGNAL` tier-conditional retention (Spec v1.3 §5.4) | cap-per-tier eviction; tier-dependent field population |
| BMA | hot Hebbian co-activation nodes | salience > 0 raises eviction priority floor under pressure |

The three "look different" framings collapse onto two orthogonal primitives:

- **`Node.TierImmune bool`** — a permanence marker. When true, eviction skips this node regardless of tier transition, pressure, or sleep-cycle compaction.
- **`Node.Salience float64`** — a retention-priority modulator. Hebbian co-activation increments salience up to a cap; Ebbinghaus decay decrements it over time. Eviction priority order is `(TierImmune == false) AND (lower salience evicted first)`.

Plus one graph-level primitive (not per-node):

- **`Graph.SetTierEvictionCap(t Tier, cap int)`** — Contextus's cap-per-tier eviction (Spec v1.3 §5.4) maps onto this. BMA's sleep-cycle compaction can also use it.

Together: BMA NT_SEED uses TierImmune. BMA hot nodes use Salience. Contextus NT_INSIGHT_SIGNAL uses both (high salience while in retention window; cap-per-tier eviction governs the window itself; Spec v1.3 §5.4 retention policy maps directly).

## 2. Decision: additive fields on `model.Node`

```go
// model/node.go (v0.2 — additive only; v0.1 wire-format compatible)

type Node struct {
    ID       NodeID
    Type     NodeType
    Tier     Tier
    Created  time.Time
    Payload  []byte
    // ... other existing fields unchanged ...

    // TierImmune marks the node as exempt from all eviction paths.
    // Used for NT_SEED (BMA seed protocol Step 9), foundation theorems,
    // and any node whose deletion would invalidate downstream invariants.
    // Default false — preserves v0.1 wire-format compatibility.
    //
    // Soundness: per Wyrd.Hypergraph.tier_immune_node_preserves_eviction
    // (forthcoming), adding or evicting other nodes does not change the
    // membership of {v : v.TierImmune}.
    TierImmune bool `json:"tier_immune,omitempty"`

    // Salience modulates eviction priority. Range 0.0..1.0.
    //   0.0 (default): no priority modulation
    //   higher values: stronger retention under pressure
    // Hebbian reinforcement increments Salience (capped at 1.0);
    // Ebbinghaus decay decrements it over time. Default 0.0 preserves
    // v0.1 wire-format compatibility.
    Salience float64 `json:"salience,omitempty"`
}
```

Plus on `Graph`:

```go
// SetTierEvictionCap sets the maximum number of nodes that may be held
// at the given tier before eviction triggers. cap == 0 disables eviction
// at that tier (effectively infinite). Per Contextus Spec v1.3 §5.4.
//
// Eviction order under saturation: TierImmune nodes excluded; among the
// remainder, ascending Salience.
func (g *Graph) SetTierEvictionCap(t Tier, cap int)

// TierEvictionCap returns the cap currently set for the tier (0 if unset).
func (g *Graph) TierEvictionCap(t Tier) int
```

The actual eviction implementation is deferred to W-Toddle-2 (or a separate v0.2 issue) — this issue defines only the data + policy contract. Sleep-cycle integration is bma-implementor's surface.

## 3. Why these primitives and not others

**Why not a single `RetentionPolicy struct` on each node?** Considered — looks tidier but bloats `Node` for the common case (Crawl-phase, no policy). Two scalar fields with sensible defaults (false / 0.0) cost ~9 bytes per node and disappear from JSON via omitempty when unset. A `RetentionPolicy` struct adds pointer indirection and complicates Lean — the additive-fields shape carries the right v0.1 form.

**Why not a `NodeKind` enum (e.g., `KindSeed`, `KindSignal`, `KindHebbian`)?** That re-introduces the BMA-specific namespace this issue is designed to AVOID. Tenants like Sharp Butler and Möbius Fusion will have their own immune-and-salience use cases at Run; a generic primitive admits them by default. Tenant-specific node types (`NT_SEED`, `NT_INSIGHT_SIGNAL`) still exist in `Node.Type` — they just don't need to be enumerated at the substrate.

**Why not a per-node `EvictionPolicy` function pointer?** Closure-based policies don't serialize. Eviction must survive `Save → Load` round-trips. Scalar fields are the only shape that round-trips cleanly.

**Why per-tier caps on `Graph`, not on `Node`?** Per Spec v1.3 §5.4, the cap is a *graph-level* retention policy — every node at tier T is subject to the same cap, regardless of its Type. Putting the cap on `Node` would multiply the policy surface by N nodes; on `Graph` it's O(|Tier|) state.

## 4. Soundness anchor

`lean/Wyrd/TierImmunity.lean` (new file). The proof follows the same reduction pattern as `Wyrd.Hypergraph.hyperedge_preserves_incident_edges` (Phase 2 C-20a) and `Wyrd.Hypergraph.oriented_edge_preserves_incident_edges` (Phase 2 v0.2 extension; PR #31 §4.3).

### 4.1 Theorem statement

```lean
theorem tier_immune_node_preserves_eviction
  (g : Hypergraph) (op : EvictionOp) (v : NodeID) :
    nodeImmune g v →
    nodeImmune (applyEviction g op) v ∧
    ¬ (nodeEvicted (applyEviction g op) v)
```

### 4.2 Supporting lemma — `eviction_immune_blind`

```lean
lemma eviction_immune_blind
  (g : Hypergraph) (op : EvictionOp) (v : NodeID) :
    nodeImmune g v →
    applyEviction g op restricts to (g \ {immune-nodes})
```

**Plain English:** the eviction operation, restricted to the non-immune nodes, is the only domain that the eviction touches. Immune nodes are observed-but-not-touched, structurally identical to how `IncidentEdges` treats orientation metadata in PR #31 §4.3.

### 4.3 Proof argument

Two-step reduction:

1. **Strip eviction targets to non-immune set.** Construct `g_non_immune := g \ {v : nodeImmune g v}`. By `eviction_immune_blind`, `applyEviction g op` agrees with `applyEviction g_non_immune op` on the immune subset (they're untouched in both).
2. **Apply membership-preservation (cf. C-20a):** for any `v ∈ {immune-nodes}`, `v ∈ applyEviction g op` because eviction does not touch immune nodes by construction (lemma `eviction_immune_blind`).

The proof is mechanical; no new induction needed. `eviction_immune_blind` carries the structural work.

### 4.4 What lands in the impl PR

`lean/Wyrd/TierImmunity.lean` ships with **both** `eviction_immune_blind` and `tier_immune_node_preserves_eviction` proven (no `sorry`, no user-defined `axiom`; CI Phase 2 gate enforces). Estimated ~35 LOC including imports and namespace setup.

Salience itself does NOT get a Lean theorem in v0.1 — it's a soft retention-priority signal, not a structural invariant. Future v0.x may add a Lean theorem about the monotone-Hebbian-bound (salience never exceeds 1.0; never falls below 0.0).

## 5. JSON wire-format compatibility

v0.1 `Node`:
```json
{ "id": "n1", "type": "...", "tier": "complex", "created": "...", "payload": "..." }
```

v0.2 `Node` (untouched, no immunity, no salience):
```json
{ "id": "n1", "type": "...", "tier": "complex", "created": "...", "payload": "..." }
```
(identical; `omitempty` strips `tier_immune` and `salience` when at defaults)

v0.2 `Node` (immune + high salience):
```json
{ "id": "n1", "type": "...", "tier": "complex", "created": "...", "payload": "...",
  "tier_immune": true, "salience": 1.0 }
```

v0.1 graphs deserialise into v0.2 without modification (defaults apply). v0.2 graphs with no immune/salience use serialise to v0.1-indistinguishable form. **Backward-compatible at both directions.**

## 6. What this PR ships, concretely

```
model/node.go                    — TierImmune + Salience fields
model/graph.go                   — SetTierEvictionCap + TierEvictionCap
model/node_test.go               — round-trip + validation tests
model/graph_eviction_test.go     — TierImmune exclusion under saturation; salience-ordered eviction
lean/Wyrd/TierImmunity.lean      — soundness anchor (proven, no sorry)
lean/Wyrd.lean                   — import TierImmunity
doc/integration/contextus.md     — usage sketch: NT_INSIGHT_SIGNAL cap-per-tier mapping
```

Eviction execution (actually walking the saturated tier and dropping nodes) is W-Toddle-2 surface — this PR only ships the data + policy contract.

## 7. Migration path

1. Land this design doc — §I4 sign-off from named reviewers.
2. Open W-Toddle-1 impl PR with the seven artifacts above. CI green; Lean theorem proven; existing tests unaffected.
3. (Walk-α) BMA `hg/` shim begins consuming `Node.TierImmune` for NT_SEED writes; `Node.Salience` for Hebbian co-activation reinforcement. W-Toddle-3 design doc specifies the cutover schedule.
4. (Walk-α) Contextus uses `Graph.SetTierEvictionCap` for Spec v1.3 §5.4 cap-per-tier; no Contextus-specific design surface needed.
5. (Walk-α) MuninnDB engram subsystem (Wyrd issue #1) marshals the two new fields without modification.

## 8. Open questions for §I4 reviewers

1. **Field placement on `Node`** — top-level (as proposed) vs nested struct `Node.Retention { Immune bool; Salience float64 }`. My lean: **top-level**, two scalars, omitempty handles wire compat. Pushback if the team prefers grouped retention policy struct for clarity. (Same trade-off discussed in §3; flagging explicitly.)
2. **Salience range — `float64` 0.0..1.0 vs `uint16` 0..65535**. My lean: **`float64`**. Hebbian rules typically multiply small fractions; integer wrap-around at 65535 is more surprising than float underflow. Revisit if profiling shows allocation pressure.
3. **`TierImmune` semantics under explicit deletion** — does `g.RemoveNode(v)` succeed on a `TierImmune` node? My lean: **yes — `TierImmune` blocks *eviction* (automatic, policy-driven) but not *explicit deletion* (a user mutation through the capability layer).** Otherwise NT_SEEDs are unremovable forever, which precludes legitimate retraction. Pushback if BMA wants stricter immutability for NT_SEED — that's a tenant policy on top of the substrate.
4. **Cap-per-tier eviction trigger** — synchronous (on `AddNode` saturation) vs deferred (sleep cycle / background goroutine). My lean: **deferred** at the substrate; tenants (BMA sleep cycle) trigger when ready. Synchronous eviction risks unbounded write latency.

## 9. Items NOT in W-Toddle-1

- **Eviction execution policy** — W-Toddle-2 issue. The order-of-eviction logic (within the non-immune set, ordered by ascending salience, with timestamp tiebreak) lives there.
- **Hebbian reinforcement rule** — bma-systema concern; BMA writes salience via the capability layer.
- **Ebbinghaus decay** — sleep-cycle concern; BMA invokes `Graph.DecaySalience(rate)` at sleep boundaries.
- **MuninnDB engram-specific behavior** — Wyrd issue #1 scope.
- **Tenant-specific node type registry** — `NT_SEED`, `NT_INSIGHT_SIGNAL` etc. live in their tenant repos.

## 10. Items NOT decided here (defer to a future doc)

- Per-tenant salience caps (BMA's salience=1.0 max vs Contextus's possibly different scale). Substrate accepts any float64; tenant policy on top.
- Persistence of `Graph.SetTierEvictionCap` state across save/load. Likely yes; W-Toddle-2 settles it.
- Cross-tier eviction (when a node's tier transitions). Future v0.x decision; not Toddle-blocking.

---

## Cross-references

- Wyrd issue [#38](https://github.com/JamesPagetButler/wyrd/issues/38) — this design's home issue
- `#toddle-design` seq=1 (kickoff), seq=6 (contextus-impl reframe), seq=9 (Block B precursor with this PR sketched as W-Toddle-1)
- `live-test` seq=95 (beekeeper OD-11(c) decision cascade 2026-05-13)
- Workspace roadmap §2.1 (BMA Crawl → Toddle item 2)
- Workspace phase architecture §2.2 (Wyrd v0.2 BMA scope)
- Contextus Spec v1.3 §5.4 (tier-conditional retention; analogous Contextus use case)
- BMA Seed Protocol (Step 9, NT_SEED definition)
- PR #31 §4.3 (oriented_edge_preserves_incident_edges — proof structure pattern this issue's Lean anchor follows)
- ADR-003 §I4

---

*Status: DRAFT v0.1 — open for §I4 review. Implementation PR blocked on explicit sign-off from `@bma`, `@bma-implementor`, `@contextus-impl`, and the beekeeper.*
