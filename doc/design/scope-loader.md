# Scope-Node Configuration Loader API v0.1

**Status:** Design **v0.1** — open for review per ADR-003 §I4
**Tracks:** Wyrd issue [#33](https://github.com/JamesPagetButler/wyrd/issues/33); Contextus issue #9 (companion); QBP federation tenancy §3
**Governance anchor:** ADR-003 §I4; `#toddle-design` seq=12 (qbp-implementor capability audit, #4); seq=13 (this commitment)
**Authors:** wyrd-implementor

---

## 0. §I4 invariant — design-doc-as-S-01-review-surface

This document is the §I4 review surface for the Wyrd-side scope-node configuration loader. Implementation PR blocked on explicit sign-off from named reviewers (§9).

The companion Contextus issue #9 (schema + reference loader) can stub against this API surface immediately — that's exactly why this design lands ahead of W-Toddle-2/3/4 (per `@contextus-impl` `#addendum-18-walk` seq=47 ask and my `#toddle-design` seq=13 commitment).

## 1. Motivation

Two federation requirements converge on this issue:

**1.1 Tenant declarations need a Wyrd-shaped landing place.** Each federation tenant declares its Stance × Locale focal cone as scope-node configuration (per Contextus Spec v1.3 §4.6). For QBP, that's ~10 `NT_SCOPE_PHYSICAL` + ~6 `NT_SCOPE_CONCEPTUAL` hyperedges with geometry, type-node assignments, and temporal_range metadata. Today there is **no API for declarative scope-node loading**; an implementor instance would have to hand-craft each hyperedge via raw `model.AddHyperedge` calls. That doesn't scale and isn't reproducible.

**1.2 BMA capability #4 — scope-node-tagged NT_SIGNAL emission** (per `@qbp-implementor` `#toddle-design` seq=12 audit). The QBP-tenant scope nodes from `docs/qbp-federation-tenancy.md` v0.1 (PR #403 draft) cannot populate Wyrd until the loader ships. Without this, BMA's Scout dispatch (via ScoutQuery — PR #35) has no scope cone to dispatch against; the federation-tenancy substrate stays paper-only.

**1.3 Crawl-shippable framing for the meeting cadence.** Per `@contextus-impl` seq=44: Contextus #9 schema can land standalone (JSON Schema + YAML fixtures + reference loader against a stub Wyrd API). This design fixes the API signature so contextus-impl's stub-against-real-shape work proceeds in parallel.

## 2. Decision — `store/scope_loader.go`

```go
// Package store extension: scope-node configuration loader.
//
// LoadScopeConfig reads a Wyrd-shaped scope configuration from disk
// and populates the given graph as a transactional unit. Either all
// scope nodes + membership edges land in the graph, or none do —
// on any error the graph is restored to its pre-call state.
//
// Soundness anchor: forthcoming Wyrd.Hypergraph.scope_loader_atomic
// (proven; CI Phase 2 gate). Proof structure follows the bridge-batch
// preserves_count pattern (Phase 2 C-20c).
package store

import (
    "github.com/JamesPagetButler/wyrd/model"
)

// LoadScopeConfig reads a scope-node configuration (YAML or JSON) at
// configPath, validates it against Contextus Spec v1.3 §4.6 + §11.4,
// and atomically populates graph with the resulting scope nodes +
// membership edges.
//
// Returns ErrScopeConfigInvalid if validation fails (no mutation).
// Returns ErrScopeLoadConflict if a scope node ID is already present
// in graph (no mutation; explicit upsert is a v0.2 surface decision).
// Returns wrapped fs / parse errors for I/O issues.
//
// Concurrency: holds graph.mu.Lock() across the full transaction —
// same I3 atomicity guarantee as PromoteBatch (ADR-003 §I3).
func LoadScopeConfig(graph *model.Graph, configPath string) error
```

### 2.1 Tenant-side schema (YAML)

```yaml
# qbp-tenant-scope.yaml — illustrative QBP federation tenant config

physical_scopes:
  - id: "contextus:scope:physical:cascadia"
    description: "Cascadia subduction zone (lat 40-50, lon -130 to -120, surface to 50km depth)"
    bounds:
      lat: [40.0, 50.0]
      lon: [-130.0, -120.0]
      time: ["2010-01-01T00:00:00Z", "2030-01-01T00:00:00Z"]
      height: [-50000.0, 0.0]
    type_nodes: ["bma.runtime.geophysical"]

conceptual_scopes:
  - id: "contextus:scope:conceptual:slow-slip"
    description: "Episodic Tremor and Slip phenomena"
    type_nodes: ["bma.runtime.slow-slip", "qbp.tenant.ets"]

scope_memberships:
  - scope: "contextus:scope:physical:cascadia"
    member: "contextus:scope:conceptual:slow-slip"
    weight_tier: "complex"
```

The YAML maps to Contextus Spec v1.3 §11.4 types verbatim. JSON is also accepted (same schema, different encoding). The reference YAML→struct decoder lives Contextus-side (issue #9); Wyrd consumes the struct shape.

**Note on `weight_tier`** (per `@contextus-impl` PR #40 review): the `weight_tier` field in the YAML maps to the `Weight.Tier` of the constructed `model.Hyperedge`, not a separate field on `ScopeMembership` itself. Spec v1.3 §11.4's `ScopeMembership` carries no tier of its own — the loader inspects `weight_tier` (or defaults to `complex`) and constructs `model.Weight{Tier: t}` for the hyperedge. If the YAML schema's `weight_tier` doesn't map onto a field in the Contextus Go type post-PR #12, this is a YAML-only loader concern; the resulting `model.Hyperedge` carries the tier through `model.Weight` as already specified.

### 2.2 Contextus type imports — single source of truth

Per the §4.6 mapping in `doc/integration/contextus.md`, the canonical `ScopePhysical` / `ScopeConceptual` / `ScopeMembership` Go types live in **Contextus's `pkg/types/`** (relocated from `internal/contextus/types/` in Contextus PR #12, merged 2026-05-14, **specifically to enable this federation contract**). Wyrd does NOT re-define these. The loader imports:

```go
import (
    ctypes "github.com/JamesPagetButler/contextus/pkg/types"
)
```

This requires `contextus` to be a Go module — already true. GOPRIVATE for `github.com/JamesPagetButler/*` is already configured (per CLAUDE.md workspace).

**Why not mirror?** Federation contract: Contextus owns the type. Mirroring creates two-sources-of-truth drift; Spec v1.3 §11.4 updates would require Wyrd to chase. Importing fixes drift at compile time. The cost is a cross-repo dependency, which the gonum/mat cost+benefit precedent (PR #29 / CONTRIBUTING.md Q4) already established as acceptable for `store/`.

**Why `pkg/types/` not `internal/`?** Go's internal-package rule blocks `internal/...` paths from being imported by other modules. The Contextus team relocated the types to `pkg/types/` in PR #12 to make this design's import valid. Per `@contextus-impl` `#toddle-design` seq=17 + seq=19 + seq=22: the relocation **closed PR #40's compile-block conditional finding** and is the federation contract going forward.

### 2.3 Transactional semantics

`LoadScopeConfig` is **all-or-nothing**, matching `PromoteBatch`'s ADR-003 §I3 atomicity. The implementation:

```
Phase 1 (parse + validate, no mutation):
  - decode YAML/JSON
  - validate each scope_id is non-empty + RFC-3986-ish
  - validate bounds (Min ≤ Max per-component)
  - validate type_nodes exist as model.NodeType strings
  - check no scope_id already present in graph

Phase 2 (commit under one Lock):
  - graph.mu.Lock()
  - for each physical / conceptual scope: graph.AddNode(...)
  - for each membership: graph.AddHyperedge(...)
  - graph.mu.Unlock()

On any phase-1 error: return error; no mutation.
On any phase-2 error: should not happen (phase-1 caught it); panic
  (this is an invariant violation, not a runtime error).
```

Soundness anchor (forthcoming, lands proven in impl PR): `Wyrd.Hypergraph.scope_loader_atomic` — a graph that receives a `LoadScopeConfig` call either contains all scope nodes + membership edges or contains none, post-call. The proof reduces to a finite induction over the config's scope list, where each step applies `Wyrd.Bridge.bridge_promote_preserves_count` (Phase 2 C-20c) reasoning.

## 3. Not in v0.1

- **Upsert semantics** (replace existing scope nodes). v0.1 returns `ErrScopeLoadConflict` on collision; v0.2 may add a `LoadScopeConfigUpsert` variant.
- **Cross-tenant scope merging.** Multi-tenant Wyrd graphs at Walk-phase may want scope-node deduplication across tenants; that's a Walk concern, not a Toddle-loader concern.
- **Streaming load.** For a single tenant (~16 scope nodes), atomic-load fits comfortably in one Lock window. Multi-thousand-node configs would want streaming; deferred until evidence shows we need it.
- **Schema migration** (v1.3 → v1.4 field changes). Contextus owns the type; schema-migration tooling stays Contextus-side.
- **TierImmune / Salience on scope nodes** — W-Toddle-1 (PR #39) primitives compose: tenants whose scope nodes should never decay set `TierImmune = true` in the YAML; loader threads it through. Documented as a YAML option in §2.1 (deferred to a v0.x example doc).

## 4. Soundness anchors

- **`Wyrd.Hypergraph.scope_loader_atomic`** (forthcoming, lands with impl PR). Proof structure: finite induction over the config's scope list; inductive step is one application of `Wyrd.Bridge.bridge_promote_preserves_count` (Phase 2 C-20c). ~30 LOC estimate; ships proven (no sorry, no axiom).
- **`Wyrd.Hypergraph.hyperedge_preserves_incident_edges`** (Phase 2 C-20a) — covers the membership-edge addition step.
- **ADR-003 §I3** — RWMutex `Lock` window held across the full transaction (write-atomicity invariant).

## 5. Cross-repo coordination

This issue is **paired with Contextus #9** (scope-node config schema + reference loader). The split:

| Concern | Owner | Deliverable |
|---|---|---|
| YAML / JSON schema spec | contextus-impl | `contextus/schema/scope-config.schema.json` |
| Go struct types | contextus-impl | `contextus/internal/contextus/types/scope_*.go` (already shipped at v1.3) |
| YAML fixtures (QBP tenant) | contextus-impl + qbp-implementor | `contextus/testdata/scope-configs/qbp.yaml` |
| Reference YAML→struct decoder | contextus-impl | `contextus/internal/contextus/scopeloader/decoder.go` (consumed by wyrd) |
| Wyrd `LoadScopeConfig` | wyrd-implementor | `wyrd/store/scope_loader.go` (this design) |
| Lean soundness anchor | wyrd-implementor | `wyrd/lean/Wyrd/ScopeLoader.lean` |
| BMA-side `bma scope load` reins | bma-implementor | `bma/internal/bma/cli/scope_load.go` (Toddle deliverable) |

contextus-impl can stub Contextus #9 against the API signature defined here **starting now**; no need to wait for the impl PR.

## 6. What this design PR ships

Only the design doc (this file). The impl PR ships:

```
store/scope_loader.go            — LoadScopeConfig implementation
store/scope_loader_test.go       — load happy path, conflict error, validation error, atomicity (parse-fail mid-list leaves graph untouched)
store/scope_loader_test_fixtures/ — fixtures cross-tested against Contextus YAML fixtures
lean/Wyrd/ScopeLoader.lean       — Wyrd.Hypergraph.scope_loader_atomic (proven; no sorry)
lean/Wyrd.lean                   — import ScopeLoader
doc/integration/bma.md           — usage sketch: `bma scope load <path>` reins wrapper consuming this API
doc/integration/contextus.md     — usage sketch updated to show LoadScopeConfig as the canonical load path
```

## 7. Open questions for §I4 reviewers

1. **YAML vs JSON as canonical**. Tenants tend to author YAML for human-readability; QBP federation tenancy v0.1 uses YAML in its `docs/qbp-federation-tenancy.md` §3 examples. JSON is more machine-friendly. My lean: **accept both**, dispatch on file extension; YAML is canonical for human-authored configs, JSON for tool-generated. Pushback OK.
2. **`LoadScopeConfig(graph, configPath)` vs `LoadScopeConfig(graph, reader io.Reader)`**. The path form is convenient for one-shot tenant bootstrap; the Reader form composes with HTTP / S3 / etc. My lean: **both** — path form is sugar over reader form; reader form is the primitive. Land both in v0.1; cheap.
3. **Conflict semantics** — error vs silent skip vs upsert. My lean: **error at v0.1** (`ErrScopeLoadConflict`). Silent skip is surprising; upsert is a v0.2 concern. Pushback if @contextus-impl needs upsert at Toddle for tenant re-bootstrap.
4. **`type_nodes` validation** — should the loader require type-node strings to already exist in the graph, or allow forward-references? My lean: **allow forward-references** at v0.1 (just validate string-shape). Most tenants populate type-nodes in the same load; requiring presence would force a particular ordering. v0.2 may add a strict-mode option.

## 8. Migration path

1. Land this design doc — §I4 sign-off from named reviewers (§9).
2. Open impl PR (`store/scope_loader.go` + Lean anchor); CI green; Contextus YAML fixtures pass round-trip.
3. (Parallel) contextus-impl ships Contextus #9 (schema + decoder + YAML fixtures) against the API signature defined here. Contextus #9 stub-against-real-shape work proceeds **without waiting for Wyrd's impl PR**.
4. (Walk-α) BMA `hg/` shim → `bma scope load` reins wrapper consumes `LoadScopeConfig`; QBP tenant scope-nodes populate at BMA startup.
5. (Walk-α) ScoutQuery (Wyrd PR #35) consumes scope-node `model.Node`s populated by this loader as its `Volume` / `Stance` parameters.

## 9. §I4 named reviewers

- `@contextus-impl` — primary consumer (Contextus #9 stub depends on this API signature)
- `@qbp-implementor` — primary tenant author (QBP federation tenancy §3 produces YAML for this loader)
- `@bma-implementor` — runtime consumer (BMA reins wraps this API for `bma scope load`)
- beekeeper

`@cth-implementor` consultative only (scope nodes don't directly intersect CTH inventory schema).

## 10. Items NOT decided here

- **Scope-node provenance metadata** (who declared the scope, when, which Contextus revision) — captured in the YAML's free-form metadata fields; Wyrd treats it as opaque `Node.Payload` bytes per existing convention.
- **Scope-node retention policy** (decay, eviction caps) — composes with W-Toddle-1 (PR #39) primitives; not a scope-loader concern.
- **Scope-bounded query routing** — when BMA dispatches ScoutQuery over a specific scope, the query/ subpackage handles routing; not the loader.
- **Multi-version scope-config** (scope evolves; how does it migrate?) — out of scope for v0.1; v0.2 may add a `ScopeVersion` field.

---

## Cross-references

- Wyrd issue [#33](https://github.com/JamesPagetButler/wyrd/issues/33) — this design's home issue
- Contextus issue #9 — companion (schema + reference decoder + YAML fixtures)
- QBP federation tenancy [#403](https://github.com/JamesPagetButler/QBP/pull/403) — §3 produces YAML for this loader
- BMA Theory Addendum 18.0 §2.1 / §2.2 (Stance × Locale focal cone — what scope nodes encode)
- Contextus Spec v1.3 §4.6 (Scope Nodes) + §11.4 (Go type definitions)
- `#toddle-design` seq=12 (qbp-implementor capability audit, capability #4), seq=13 (this commitment)
- `#addendum-18-walk` seq=44 / seq=47 (contextus-impl Wyrd #33 cadence ask)
- Wyrd PR #35 (ScoutQuery) — consumer of scope-nodes this loader produces
- Wyrd PR #39 (W-Toddle-1 tier-immune + salience) — composes with this loader for tenant retention policy
- ADR-003 §I3 (Lock atomicity), §I4 (design-doc-as-S-01-review-surface)

---

*Status: DRAFT v0.1 — open for §I4 review. Implementation PR blocked on explicit sign-off from `@contextus-impl`, `@qbp-implementor`, `@bma-implementor`, and the beekeeper.*
