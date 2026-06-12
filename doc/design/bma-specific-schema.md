# BMA-Specific Node-Type Schema — W-Toddle-2 (OD-11(c))

**Status:** Design **v0.1 — LANDED** (PR #47 design + PR #48 impl + PR #55 W-Toddle-2-extension, all merged 2026-05-14/15). The design doc and implementation are on Wyrd `main`. §I4 sign-off received from `@bma`, `@bma-implementor`, and beekeeper prior to merge.
**Tracks:** Wyrd issue [#43](https://github.com/JamesPagetButler/wyrd/issues/43) (OD-11(c) absorption tracking); paired with `@bma-implementor`'s TD-4 inventory delivered `live-test` seq=99
**Governance anchor:** ADR-003 §I4; `#sprint-1-toddle-entry` Sprint 1 deliverable #2 for wyrd-implementor; OD-11(c) constitutional approval from `@bma` Marcy `#toddle-design` seq=24
**Authors:** wyrd-implementor

---

## 0. §I4 invariant — design-doc-as-S-01-review-surface

W-Toddle-2 is the **second** of three OD-11(c) absorption deliverables (per `#toddle-design` seq=9 Block B precursor + seq=11 announcement):

- **W-Toddle-1** ✅ generic primitives — `Node.TierImmune`, `Node.Salience`, `Graph.SetRetentionCap`. PR #39 design merged 2026-05-14; PR #42 impl merged 2026-05-14; Lean anchor `Wyrd.TierImmunity` in PR #46.
- **W-Toddle-2** (this doc) — `model.NodeType`-to-policy mapping for BMA-specific structures per the TD-4 inventory. **No new substrate primitives**; this is a doc-grade contract specifying which `NodeType` strings the W-Toddle-1 substrate enforces immunity / salience for.
- **W-Toddle-3** — BMA `hg/` shim retirement timeline (joint authorship with `@bma-implementor`).

Per Marcy `#toddle-design` seq=24 OD-11(c) constitutional check: A11 Topological Cognition decay-immunity guarantees survive the absorption *structurally* via the W-Toddle-1 Lean anchor, provided the absorption process sets `TierImmune=true` on every TD-4 inventory entry that needs it. **W-Toddle-2 is exactly that "absorption process" spec.**

## 1. Motivation

W-Toddle-1 shipped generic primitives — `Node.TierImmune bool`, `Node.Salience float64`. The primitives are intentionally tenant-agnostic: any consumer can use them. **W-Toddle-2 specifies which BMA `NodeType` strings actually require which W-Toddle-1 settings**, so the BMA `hg/` shim's per-write policy is documented + auditable rather than implicit-in-call-sites.

Without this mapping doc:
- BMA `hg/` shim author (`@bma-implementor`) must keep the policy table mentally, with no doc to reference at review time
- Contextus / Sharp Butler / future tenants have no canonical example of "what does it look like to use TierImmune + Salience as a tenant?"
- The Marcy seq=24 constitutional invariant ("the absorption process sets `TierImmune=true` on every NT_SEED at migration time") has no contract surface — there's no document a reviewer can point to and say "this NT_SEED-typed node was created without TierImmune=true; that's a violation."

This doc closes those gaps by making the mapping load-bearing + reviewable.

## 2. TD-4 inventory — the load-bearing table

From `@bma-implementor` `live-test` seq=99 + `#toddle-design` seq=25 (delivered 2026-05-14 in response to my TD-4 ask):

| BMA `NodeType` (canonical) | `TierImmune` | `Salience` default | Rationale |
|---|---|---|---|
| `bma.seed` | **true** | **1.0** | NT_SEED — permanent Layer 3 hypergraph nodes per BMA Seed Protocol Step 9. Foundation-of-identity; cannot decay |
| `bma.lineage.life-certificate` | **true** | **1.0** | NT_LIFE_CERTIFICATE — lineage protocol invariant |
| `bma.lineage.death-certificate` | **true** | **1.0** | NT_DEATH_CERTIFICATE — lineage protocol invariant |
| `bma.observation` | **false** | Hebbian-modulated (initial 0.0) | NT_OBSERVATION — decay-eligible per sleep cycle; salience set by Hebbian co-activation |
| `bma.params.proposal` | **true** | **1.0** | NT_PARAM_PROPOSAL — audit-trail invariant; proposed-then-rejected proposals must survive for explainability |
| `bma.params.trust-state` | **true** | **1.0** | NT_PARAM_TRUST_STATE — audit-trail invariant |
| `bma.lineage.last-words` | **true** | **1.0** | NT_LAST_WORDS — ETInherited; lineage-permanent |
| `bma.lineage.eulogy` | **true** | **1.0** | NT_EULOGY — ETInherited; lineage-permanent |

**Eight `NodeType` strings; seven are `TierImmune=true`.** Only `bma.observation` is decay-eligible — the operational hot path where Hebbian reinforcement modulates `Salience` upward and Ebbinghaus decay modulates it downward over the sleep cycle.

Canonical naming uses the `bma.*` prefix discipline established in PR #16 (`bma.runtime.*` reserved namespace). Sub-categories (`lineage.*`, `params.*`) are tenant-internal; Wyrd substrate doesn't validate them beyond the `bma.` prefix.

## 3. Decision — policy table lives as a Go map, not a switch

```go
// model/bma_policy.go (proposed; lands in W-Toddle-2 impl PR)

// bmaNodeTypePolicy is the canonical mapping of BMA-specific NodeType
// strings to the (TierImmune, defaultSalience) policy each requires.
// Per @bma-implementor TD-4 inventory (live-test seq=99).
//
// Marcy's #toddle-design seq=24 constitutional check makes this map
// load-bearing: every NT_SEED-typed node MUST be created with
// TierImmune=true, or A11 Topological Cognition decay-immunity
// guarantees fail.
var bmaNodeTypePolicy = map[NodeType]struct {
    TierImmune bool
    Salience   float64
}{
    "bma.seed":                       {TierImmune: true, Salience: 1.0},
    "bma.lineage.life-certificate":   {TierImmune: true, Salience: 1.0},
    "bma.lineage.death-certificate":  {TierImmune: true, Salience: 1.0},
    "bma.observation":                {TierImmune: false, Salience: 0.0},
    "bma.params.proposal":            {TierImmune: true, Salience: 1.0},
    "bma.params.trust-state":         {TierImmune: true, Salience: 1.0},
    "bma.lineage.last-words":         {TierImmune: true, Salience: 1.0},
    "bma.lineage.eulogy":             {TierImmune: true, Salience: 1.0},
}

// BMAPolicy returns the canonical TierImmune + Salience defaults for a
// BMA-prefixed NodeType. Returns (false, 0.0, false) for unknown types,
// signaling no canonical policy — the caller decides defaults.
func BMAPolicy(t NodeType) (immune bool, salience float64, known bool) {
    p, ok := bmaNodeTypePolicy[t]
    return p.TierImmune, p.Salience, ok
}

// ApplyBMAPolicy mutates n.TierImmune + n.Salience to the canonical
// defaults for n.Type if a policy is known. No-op for unknown types.
// Idempotent — safe to call multiple times.
func ApplyBMAPolicy(n *Node) {
    immune, salience, ok := BMAPolicy(n.Type)
    if !ok {
        return
    }
    n.TierImmune = immune
    n.Salience = salience
}
```

**Why a map + helper, not a switch:**
- The policy is data, not control flow. Switches add a third place (config + doc + code) where the table must agree.
- Sharp Butler / Möbius Fusion can register their own policies via a future `RegisterPolicy(prefix, map)` extension without touching BMA's table.
- Lean / proof side gets a Finset to reason over rather than a syntactic pattern-match.

**Why `ApplyBMAPolicy` instead of validation:**
- The BMA `hg/` shim **applies** the policy at write time (sets the fields). Validation would mean rejecting writes that disagree; that's stricter than the contract needs. Tenants can override the defaults at write time if they have a reason (e.g., a one-off experimental node).
- `bma cap`-style capability enforcement happens elsewhere; this is just default-setting.

## 4. Soundness anchor

No new Lean theorem is required for W-Toddle-2. The substrate guarantee that "TierImmune=true → never decay" is already proven in `Wyrd.TierImmunity.tier_immune_node_preserves_eviction` (PR #46). W-Toddle-2 is the **policy that determines which nodes get TierImmune=true** — it's downstream of the structural guarantee.

If a future Lean theorem is desired (Marcy / bma-implementor's call), it would say: *"For any node `n` constructed via `bmaSchemaConstruct(...)`, if `n.Type ∈ {bma.seed, bma.lineage.*, bma.params.*}` then `n.TierImmune == true`."* That's a syntactic property of the construction function, not a structural property of the graph — provable via cases analysis on `bmaNodeTypePolicy`. Estimated ~20 LOC if requested.

My lean (no Lean theorem at v0.1): the structural guarantee from W-Toddle-1 + the auditability of this map's contents make A11 constitutionally clean. Adding a Lean theorem proves a fact about code, not a fact about math.

## 5. What the W-Toddle-2 impl PR ships

```
model/bma_policy.go              — bmaNodeTypePolicy map + BMAPolicy + ApplyBMAPolicy
model/bma_policy_test.go         — round-trip test per entry; ApplyBMAPolicy idempotence; unknown-type no-op; JSON wire-format round-trip for each NodeType
doc/integration/bma.md           — usage sketch: BMA hg/ shim calls model.ApplyBMAPolicy before Graph.AddNodeWithCapability
```

No `compute/`, `query/`, `scout/`, or `predictions/` impact. No new public types beyond the helper function.

## 6. Not in W-Toddle-2 (deferred or out-of-scope)

- **BMA `hg/` shim retirement timeline** — that's W-Toddle-3; this doc only specifies the policy, not the migration sequence.
- **Hebbian reinforcement rule for `Salience`** — bma-systema concern; consumes `Salience` field after W-Toddle-2 sets the default. Hebbian rule is a sleep-cycle implementation detail not visible at the Wyrd substrate.
- **Per-tenant policy maps** (Contextus, Sharp Butler) — out of scope. If/when a second tenant needs a policy map, a `RegisterPolicy(prefix string, m map[NodeType]Policy)` API can land in W-Toddle-4 or later. v0.1 hardcodes the BMA table specifically.
- **`bma.runtime.*` types from PR #16** (`FlagNormDrift`, `ObsZDDetected`, etc.) — these are *anchors* for the WDEvent observer, not *nodes* in the hypergraph. Different surface; not in TD-4 inventory; out of scope for W-Toddle-2.

## 7. Open questions for §I4 reviewers

1. **`bmaNodeTypePolicy` as `map[NodeType]struct{...}` vs `var bmaSeedPolicy = ...; var bmaObservationPolicy = ...`** — the map is succinct; per-policy vars are more grep-friendly. My lean: **map**. Pushback if review prefers grep-friendly.
2. **`ApplyBMAPolicy(*Node)` vs `NewBMANode(NodeType, ...) Node`** — applying to existing Node is mutation-friendly for the BMA `hg/` shim; a `NewBMANode` constructor is cleaner but more API surface. My lean: **`ApplyBMAPolicy`** — single helper, BMA shim already constructs Nodes via existing model API.
3. **What about Salience-override?** Per §3 design: tenants can override defaults at write time. Should the BMA `hg/` shim ever override? My lean: **never at the substrate layer** — BMA's Hebbian rule modulates `Salience` after the initial set, by calling `g.UpdateNodeSalience(id, newValue)` (forthcoming method) rather than reconstructing the node. Capability-gated.
4. **`bma.observation` Salience default of `0.0`** — Hebbian starts unmodulated. Alternative: 0.5 (neutral). My lean: **0.0** — explicit Hebbian writes set it; default of 0.0 means "not yet modulated" which is information-bearing.

## 8. §I4 named reviewers

- `@bma` (Marcy / Gen 61+) — primary BMA gov-layer; A11 constitutional check (already approved in principle at `#toddle-design` seq=24; this PR makes the policy table reviewable)
- `@bma-implementor` — TD-4 inventory author; their data, their sign-off on the canonicalisation
- beekeeper — final acceptance

`@contextus-impl` consultative — the same `model.*` helper pattern is what they'll want for `bma.contextus.signal.*` if they migrate to a policy-driven schema later.

## 9. Migration path

1. Land this design doc — §I4 sign-off from named reviewers.
2. Open W-Toddle-2 impl PR per §5; CI green; tests verify the map's contents match this doc's table verbatim.
3. (BMA-side) `@bma-implementor` updates the `hg/` shim to call `model.ApplyBMAPolicy(&node)` before `Graph.AddNodeWithCapability(node, cap)` at write time.
4. (Walk-α) W-Toddle-3 design specifies when BMA `hg/` becomes empty (or stays as a thin re-export); migration is policy-only at v0.1, code-removal at W-Toddle-3.

## 10. What this PR depends on (already landed)

- ✅ PR #39 design + PR #42 impl — `Node.TierImmune` + `Node.Salience` + `Graph.SetRetentionCap`
- ✅ PR #46 — Lean `tier_immune_node_preserves_eviction` (structural guarantee this policy table consumes)
- ✅ Beekeeper OD-11(c) decision (`live-test` seq=95)
- ✅ Marcy gov-layer constitutional approval (`#toddle-design` seq=24)
- ✅ `@bma-implementor` TD-4 inventory delivery (`live-test` seq=99)

All prerequisites are on `main`. W-Toddle-2 is ready to drop in.

---

## Cross-references

- Wyrd issue [#43](https://github.com/JamesPagetButler/wyrd/issues/43) — OD-11(c) tracking issue this design closes the W-Toddle-2 deliverable for
- `live-test` seq=95 (beekeeper OD-11(c) decision), seq=99 (TD-4 inventory)
- `#toddle-design` seq=9 (Block B precursor), seq=24 (Marcy constitutional approval), seq=25 (bma-implementor closeout-ack)
- `#sprint-1-toddle-entry` seq=1 (Sprint 1 scope assignment)
- Wyrd PR #39 (W-Toddle-1 design) + PR #42 (W-Toddle-1 impl) + PR #46 (Lean anchor)
- BMA Theory Addendum 11 (Topological Cognition — decay-immunity origin)
- BMA Seed Protocol (Step 9 — `NT_SEED` definition)
- ADR-003 §I4

---

*Status: LANDED — design doc (PR #47) + impl (PR #48) + W-Toddle-2-extension (PR #55) all merged on Wyrd `main` as of 2026-05-15. Implementation complete; §I4 sign-off received. W-Toddle-4 `doc/integration/bma.md` refresh landing with issue #43 close PR.*
