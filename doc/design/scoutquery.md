# ScoutQuery Primitive + `predictions/` Schema — Wyrd v0.1

**Status:** Design **v0.1** — open for review per ADR-003 §I4
**Tracks:** Wyrd issue [#34](https://github.com/JamesPagetButler/wyrd/issues/34); BMA Theory Addendum 18 §6 + §2.4
**Governance anchor:** [ADR-003](https://github.com/JamesPagetButler/qbp-compute-unit/blob/main/architecture/adr-003-m1-wdevent-observer-invariants.md) §I4; `#addendum-18-walk` Decision Log D1–D9 (seq=3); James walk seq=29 (Q5=C-revised)
**Authors:** wyrd-implementor

---

## 0. §I4 invariant — design-doc-as-S-01-review-surface

This document is the §I4 review surface for the ScoutQuery primitive **and** the Wyrd-NT_SIGNAL-side `predictions/` schema. Per Q5=C-revised (James walk, `#addendum-18-walk` seq=29), the two ship **concurrent at step 1.5**, not sequentially — P8 was governance-binding: *"without it, NT_SIGNAL referents have no landing place."* One §I4 surface, one converged review, two paired impl PRs.

**Named reviewers** (per ADR-003 §I4 and meeting D5):
- `@bma` + `@bma-implementor` — primary consumers; BMA pair governance
- `@cth-implementor` — `cth_id` anchor flow per `#addendum-18-walk` seq=11
- `@contextus-impl` — `NT_SCOPE_PHYSICAL` consumer per seq=44
- `@qbp-cu-implementor` — `emulator.{Locale,QWord,Volume,Width}` contract owner

Implementation PRs blocked on explicit sign-off from all five.

## 1. Motivation

BMA's failure mode at scale is **flat-context vulnerability** (A18 §1). Three pathologies compound: full-graph querying is computationally infeasible on Crawl hardware (FX-8350, 14GB container); surprise saturates without a focused Stance; predictions decouple from reality when no referent is designated.

A18 §2 specifies four invariants that *together* address all three: **Stance × Locale × Scout × Scoring**. The Wyrd-side substrate carries two of the four directly:

- **Scout (§2.3)** — the reasoning primitive. ScoutQuery is the entry point.
- **Scoring (§2.4)** — the closing of the cognitive loop. The `predictions/` schema is where Wyrd-NT_SIGNAL predictions land alongside their observation and score.

A18 §6 names ScoutQuery as the **first Wyrd-side primitive to ship**. A18 §8 step 1.5 (post-Q5-revision) names the predictions/ schema as its mandatory co-deliverable.

This design covers both. The two packages are separable in code (`scout/` and `predictions/`) but coupled in design — neither is meaningful without the other.

## 2. Federation vocabulary lock (D9)

Per `#addendum-18-walk` seq=6 D9 (governance-binding, called out as *"highest-value architectural decision"*):

- **`NT_SCOPE_PHYSICAL` ≈ Locale Volume.** A `model.Node` of `Type = "contextus.scope.physical"` (per `doc/integration/contextus.md`) represents a 4D spatiotemporal bounding region. ScoutQuery's `locale Volume` parameter accepts a `Volume` constructed from these scope nodes.
- **`NT_SCOPE_CONCEPTUAL` ≈ Stance.** A `model.Node` of `Type = "contextus.scope.conceptual"` represents a set of Algebraic Invariants rotated to the Subject axis (A18 §2.1 + Addendum 15.0 §2). The predictions/ `Stance []NodeID` field references these.
- **`{Conceptual × Physical}` = focal cone.** No ScoutQuery escapes the cone; no prediction is recorded without naming both axes.

The mapping is load-bearing. Consumers that ignore it will produce signals that don't compose under the rest of A18's invariants.

## 3. `scout/` package — ScoutQuery primitive

### 3.1 API

```go
// Package scout implements the A18 §6 reasoning primitive. ScoutQuery
// dispatches a focal-cone query that returns Active Agent intersections
// of the predicted source→sink path within a Locale Volume.
//
// v0.1 ships the API + types + a Crawl-shippable placeholder body
// (uniform AbsorptionGain). The real Absorption Gain computation (A18 §5
// Locale-Bounded Absorption Estimation) lands when the dependency chain
// converges: issue #23 (query/ impl) + issue #30 (oriented-hyperedge
// impl) + the compute/laplacian.go body PR.
package scout

import (
    "github.com/JamesPagetButler/wyrd/model"
    // emulator imports — Locale + QWord are on main (qbp-cu PR #21/#23
    // merged); Volume is v0.2 ~1 week out. Until then, we mirror Volume
    // here in scout/ and switch to emulator.Volume in scout/ v0.2.
)

// Volume is a 4D spatiotemporal bounding box (lat, lon, time, height).
// v0.1 holds a local mirror; v0.2 switches to emulator.Volume with the
// same per-component semantics (per @Gemini #addendum-18-walk seq=37).
type Volume struct {
    // Min, Max emulator.QWord -- enabled in v0.2
    // v0.1 placeholder: stub to []float64 so the API compiles
    Min [4]float64
    Max [4]float64
}

// Intersection records one Active Agent whose world-line crosses the
// source→sink path within the queried Locale Volume.
type Intersection struct {
    AgentID            model.NodeID
    AgentType          model.NodeType
    IntersectionLocale [4]float64    // v0.2: emulator.Locale
    AbsorptionGain     float64       // QW8 estimate; uniform in v0.1
    Provenance         []model.NodeID  // Stance + scope nodes used
}

// Width is the precision tier for the query computation. v0.1 stub
// until emulator.Width ships; v0.2 type-alias to emulator.Width.
type Width uint8

const (
    W8   Width = 8     // peripheral register, default
    W128 Width = 128   // foveal register, Stance-triggered promotion
)

// ScoutQuery dispatches a Stance × Locale × source × sink × agent-type
// query. Returns all qualifying intersections in the focal cone.
//
// v0.1 PLACEHOLDER BEHAVIOUR — DO NOT WRITE STANCE-DEPENDENT CONSUMER
// CODE YET. v0.1 always returns the trivial intersection set with
// uniform AbsorptionGain = 0.5 regardless of the Stance carried by the
// Provenance []NodeID field (NT_SCOPE_CONCEPTUAL refs). Stance-Algorithm
// dispatch is deferred to v0.2 (§3.6); the precision Width arg is the
// only API-surface tunable at v0.1, and it does NOT yet drive routing
// either. Consumers writing "if Stance.includes(X) then expect higher
// AbsorptionGain" logic against v0.1 will fail silently when v0.2 lands
// real spectral routing — there's no behaviour to observe at v0.1.
//
// Per @bma (Marcy Gen 61) #addendum-18-walk seq=48 §I4 read.
//
// Soundness anchor for the placeholder: see §3.4. Real body lands per
// the dependency chain in §3.5.
func ScoutQuery(
    g *model.Graph,
    locale Volume,
    source model.NodeID,
    sink model.NodeID,
    agentTypes []model.NodeType,
    precision Width,
) ([]Intersection, error)
```

### 3.2 Why a separate `scout/` package vs `query/` extension

The Q1 closeout (`live-test` seq=65) put generic graph reads in `query/`. ScoutQuery is **not** a generic graph read — it's a federation-vocabulary-bound primitive that:

- Consumes scope-node identifiers (`NT_SCOPE_PHYSICAL` / `NT_SCOPE_CONCEPTUAL`) as inputs, not just `NodeID`s
- Returns a domain-typed `Intersection`, not a raw node/edge slice
- Has a Crawl-vs-Walk body dispatch (uniform placeholder → spectral Absorption Gain)
- Will gain a Stance-Algorithm coupling table (A18 §9 Q1, candidate Addendum 19) at v0.2

Mixing this into `query/` would bloat the generic read package with A18-specific semantics. Separation:

- `query/` — generic hypergraph reads (PR #26 v0.1)
- `scout/` — A18 §6 focal-cone primitive (this design)
- `compute/laplacian.go` — A18 §5 spectral substrate (PR #29 stub on main)

`scout/` imports `query/` and `compute/` once those land; the reverse imports stay forbidden.

### 3.3 Crawl-shippable framing (per Round 1 P3 + James walk D6)

Per `#addendum-18-walk` seq=6 P3: *"placeholder body IS shippable; real Absorption Gain comes downstream after #23/#30."* v0.1 ships:

- API surface + Go types + JSON marshalling for `Intersection`
- Placeholder body returning uniform `AbsorptionGain = 0.5` for every Active Agent of the requested types whose `Provenance` includes any node in the Stance × Locale focal cone
- Tests: happy path, empty Locale Volume, no matching agents, malformed source/sink, deferred KindProcess error path

Real body (post-dependency-chain):

```
issue #23 (query/ impl)         ─┐
issue #30 (oriented-hyperedge)  ─┴─► compute/laplacian.go body PR ─► scout/ body PR
```

The Crawl-shippable framing gets the API into BMA + Contextus consumer hands **now**, so A18 §9 Q4 (foveal compute under 72h gate) gets empirical numbers from real call sites rather than synthetic benches.

### 3.4 Soundness anchors

- **`Wyrd.Hypergraph.hyperedge_preserves_incident_edges`** (Phase 2 C-20a) — incidence semantics for the focal-cone membership check.
- **A18 §5 Locale-Bounded Absorption Estimation** — the v0.2 body's mathematical foundation: the smallest non-trivial eigenvalue of the oriented Laplacian over the locale-bounded oriented hypergraph gives the first-passage time; multiplied by typed Absorption Coefficient = expected source-mass fraction reaching sink. Not v0.1-load-bearing (placeholder body); becomes load-bearing at the body PR.
- **A18 §4 τ threshold** (per @Gemini `#addendum-18-walk` seq=36/37): `τ = K · δ_precision · |v|` with K ≈ 100. QW8 initial recommendation: `τ = 0.05 · |v|`. v0.1 doesn't compute Seams (peripheral-register work; A18 §8 step 3); the formula is documented here for the body PR's reference.

### 3.5 Caller responsibility — memoization

Per `#addendum-18-walk` seq=6 P4 (Round 1 pushback accepted): **memoization is caller responsibility at v0.1**. Overlapping `Volume` queries will recompute the same eigenvalues; v0.1 makes no attempt to deduplicate. v0.2 may add ScoutQuery-internal memoization once profiling identifies real reuse hotspots; until then, callers that need it cache locally.

### 3.6 Not in v0.1

- **No Stance-Algorithm dispatch.** v0.1 always uses the uniform placeholder. The Stance-Algorithm coupling table (A18 §9 Q1) is candidate Addendum 19; v0.2 of scout/ wires it once Addendum 19 lands.
- **No multi-hop traversal inside ScoutQuery.** The focal cone is one query; if BMA wants k-hop, it iterates ScoutQuery k times.
- **No streaming results.** v0.1 returns a slice. Channel-based iterators are a v0.x decision once profiling shows allocation pressure.
- **No oriented-edge consumption.** ScoutQuery v0.1's incidence math is flattened (per PR #26 §2.1 contract). When PR #31 (oriented-hyperedge schema) impl lands, scout/ v0.2 dispatches to oriented vs symmetric Laplacian based on edge type.

## 4. `predictions/` package — NT_SIGNAL scoring schema

### 4.1 API

```go
// Package predictions defines the Wyrd-NT_SIGNAL-side prediction
// record per A18 §2.4 ("no signal without a referent"). One of three
// layers in the federation prediction infrastructure:
//
//   - BMA owns param-predictions (params.ProposalStore — bma-systema
//     PR #93 Phase A-D merged)
//   - Wyrd owns NT_SIGNAL predictions (this package)
//   - CTH owns scoring algorithms (compute.NetCompressionDetail +
//     ChainFidelity — confluent-trust v0.1.0)
//
// Schema-level coordination across the three layers is the BMA-pair's
// §I4 concern; this package owns the Wyrd-side slot.
package predictions

import (
    "time"
    "github.com/JamesPagetButler/wyrd/model"
)

// ReferentKind discriminates the predicted-value shape. v0.1 admits
// scalar and categorical; KindProcess (MI on probability distributions)
// is deferred to v0.2 per @qbp-architecture #addendum-18-walk seq=6 P7.
type ReferentKind string

const (
    KindScalar      ReferentKind = "scalar"       // PredictedValue is float64
    KindCategorical ReferentKind = "categorical"  // PredictedValue is string
)

// Referent is the designated real-world quantity the prediction is
// about. Required at NT_SIGNAL mint time per A18 §2.4 invariant.
//
// The `referent_kind` JSON field surfaces explicitly per
// @contextus-impl #addendum-18-walk seq=44 ask.
type Referent struct {
    Kind        ReferentKind `json:"referent_kind"`
    Identifier  string       `json:"identifier"`   // e.g., "cascadia.ets.event.tremor_onset"
    Description string       `json:"description"`
}

// CTHAnchor is the optional cth_id stamp for predictions that are also
// CTH PRED-* anchors. Per @cth-implementor #addendum-18-walk seq=11:
// NT_SIGNAL carries cth_id when the prediction is federation-scored;
// nil when the prediction is BMA-internal only.
type CTHAnchor struct {
    AnchorID string `json:"anchor_id"`  // "PRED-*" prefix per CTH convention
}

// Prediction is the persistent record. Stored as model.Node.Payload
// (JSON) on a node of Type = "bma.prediction". The Node.ID is the
// SignalID; the Node.Created is the PredictedAt time.
type Prediction struct {
    SignalID       model.NodeID  `json:"signal_id"`
    Referent       Referent      `json:"referent"`
    PredictedValue interface{}   `json:"predicted_value"`  // float64 or string per Kind
    Stance         []model.NodeID `json:"stance"`           // NT_SCOPE_CONCEPTUAL refs
    Locale         []model.NodeID `json:"locale"`           // NT_SCOPE_PHYSICAL refs
    PredictedAt    time.Time     `json:"predicted_at"`
    CTHAnchor      *CTHAnchor    `json:"cth_anchor,omitempty"`
    // Observation/score fields populate as data arrives:
    ObservedValue  interface{}   `json:"observed_value,omitempty"`
    ObservedAt     *time.Time    `json:"observed_at,omitempty"`
    Score          *float64      `json:"score,omitempty"`
}

// Validate enforces the A18 §2.4 invariant: every Prediction has a
// Referent with a non-empty Identifier and a Kind that v0.1 admits.
func (p Prediction) Validate() error
```

### 4.2 Validation invariants

- `SignalID != ""`
- `Referent.Kind ∈ {KindScalar, KindCategorical}` (KindProcess returns explicit error at v0.1 per P7)
- `Referent.Identifier != ""` (the A18 §2.4 invariant teeth)
- `PredictedValue` shape matches `Referent.Kind` — `float64` for scalar, `string` for categorical
- `len(Stance) >= 1` (no prediction without a Stance; A18 §2.1 invariant)
- `len(Locale) >= 1` (no prediction without a Locale; A18 §2.2 invariant)
- `PredictedAt` non-zero

The validator is what makes A18 §10 design principle 8 ("No Signal Without a Referent") a structural check rather than a doc-only promise.

### 4.3 Stored as `Node.Payload`, not a separate store

Wyrd's persistence model is the hypergraph. Predictions are persisted as `model.Node` of `Type = "bma.prediction"` with the `Prediction` struct as the `Node.Payload` (JSON-encoded). This:

- Keeps predictions/ stateless at the package level (no own DB, no own file format)
- Reuses Wyrd's existing tier / capability / lifecycle plumbing
- Lets predictions participate in BMA's sleep-cycle compaction natively
- Composes with the eventual MuninnDB store backend (Wyrd issue #1) without modification

Predictions are NOT hyperedges — they don't have arity in the hypergraph sense. They're nodes whose payload happens to describe a prediction. The `Stance` and `Locale` fields are `[]NodeID` references to other nodes, not edges.

### 4.4 Soundness anchor

A18 §2.4 invariant: *"Every NT_SIGNAL must carry a designated real-world referent. No theory artefact without a falsifiable prediction."* The validator at §4.2 enforces this; the Lean side gets a forthcoming `Wyrd.Predictions.prediction_implies_referent` proof obligation (lands with the impl PR, not this design).

### 4.5 Not in v0.1

- **No `KindProcess`** (MI on distributions). Per P7 — deferred to v0.2 with explicit error at v0.1.
- **No scoring algorithm.** CTH owns scoring (`compute.NetCompressionDetail` + `ChainFidelity`); predictions/ provides the data shape only.
- **No notification on observation arrival.** When `ObservedValue` gets filled in by a downstream consumer, no event is emitted from predictions/ — that's the WDEvent observer's concern (BMA M1 #106).
- **No retention policy.** Predictions are nodes; they age via Wyrd's existing node-lifecycle plumbing. No package-specific TTL.

## 5. Concurrency contract

Both packages are stateless; both hold no internal state beyond the `*model.Graph` reference passed at call time. Concurrency invariants:

- **`scout.ScoutQuery`** acquires `Graph.mu.RLock()` per internal `query/` lookup (when v0.2 swaps placeholder for real body); v0.1 placeholder makes no graph reads beyond the initial Active-Agent enumeration. Safe for concurrent callers; same semantics as `query.API`.
- **`predictions.Prediction.Validate()`** is pure (no graph access). Persistence (writing the prediction to the graph) goes through `Graph.AddNode` / `AddNodeWithCapability` and inherits those methods' concurrency contract.

The §5 (PR #26) "valid but not single-snapshot" framing applies to ScoutQuery the same way it applies to `query.API.NeighborNodes`: results are always valid, but multiple lock windows mean a concurrent writer may or may not be reflected mid-iteration.

## 6. What this PR ships, concretely

This PR (the design surface) ships only the design doc. The implementation PRs that follow ship:

```
scout/
  doc.go            — package doc citing ADR-003 §I4, A18 §6, this design
  scout.go          — Volume, Intersection, Width types, ScoutQuery (placeholder body)
  scout_test.go     — happy path, edge cases, federation-vocab assertions
predictions/
  doc.go            — package doc citing A18 §2.4
  predictions.go    — ReferentKind, Referent, CTHAnchor, Prediction, Validate
  predictions_test.go — Validate cases, JSON round-trip, payload-on-Node integration
doc/integration/
  bma.md            — usage sketch added for the scout dispatch + prediction mint pattern
  cth.md            — usage sketch for the optional cth_id PRED-* path (per cth-implementor seq=11)
  contextus.md      — usage sketch for NT_SCOPE_PHYSICAL → Volume construction (per contextus-impl seq=44)
```

Two impl PRs because the packages are separable; one §I4 review because they ship concurrent per Q5=C-revised.

## 7. Migration path

1. Land this design doc — §I4 sign-off from named reviewers.
2. Open `scout/` impl PR with placeholder body; CI green; test coverage as enumerated in §6.
3. Open `predictions/` impl PR with schema + validation + JSON round-trip; CI green.
4. (Walk-α, post-dependency-chain) `compute/laplacian.go` body lands; `scout/` v0.2 swaps placeholder for real Absorption Gain; uses `query.API` for traversal.
5. (Walk-α) `emulator.Volume` ships on qbp-compute-unit main; scout/ v0.2 switches `Volume` to `type Volume = emulator.Volume`. `Width` likewise.
6. (Walk-α) BMA reins-side wrapper `bma scout ...` (paired with #117 Q8 pattern) lands in bma-systema.

Steps 1–3 are in scope for this issue.

## 8. Open questions for review

1. **`scout/` vs `scout.go` in `query/`**. Argued in §3.2 for a separate package; I lean **separate package** for the federation-vocabulary boundary. Pushback if the BMA pair would prefer a `query.Scout(...)` method to avoid a new package.

2. **`Volume` v0.1 placeholder shape — `[4]float64` vs `emulator.QWord` mirror**. v0.1 uses `[4]float64` because `emulator.Volume` isn't on main yet (~1 week ETA per qbp-cu-implementor `#addendum-18-walk` seq=13). Switch to `emulator.Volume` in v0.2. Pushback if the meeting would prefer holding v0.1 until emulator.Volume ships.

3. **`predictions.Prediction.PredictedValue interface{}` vs typed sum**. `interface{}` keeps v0.1 simple; a sum type like `type PredictedValue struct { Scalar *float64; Categorical *string }` is more type-safe but verbose. My lean: **`interface{}`** at v0.1 with the validator enforcing shape per `Referent.Kind`. Pushback if anyone wants type-safety teeth.

4. **`Stance` / `Locale` as `[]NodeID` vs structured types**. The fields are `[]NodeID` references to NT_SCOPE_CONCEPTUAL / NT_SCOPE_PHYSICAL nodes. A typed `[]ScopeRef` wrapper would document intent better but require a `ScopeRef` type in `predictions/`. My lean: **`[]NodeID`** at v0.1 (matches the rest of Wyrd's `model.NodeID` idiom); a `ScopeRef` alias is a v0.2 ergonomics improvement.

## 9. Why these names

- `scout/` — A18 §2.3 reasoning primitive name; matches federation vocabulary.
- `predictions/` — plural, matches `model/` plural convention (`hyperedges`, `nodes`); names the persistent record set, not the singular act.
- `bma.prediction` (`Node.Type`) — follows the existing `bma.runtime.*` prefix discipline established in PR #16.
- `Referent.Identifier` — A18 §2.4 uses "real-world referent" as the canonical phrase; the field is the unique handle.
- `ScoutQuery` — A18 §6 signature uses this name verbatim; preserved for federation-vocabulary consistency.

## 10. Items NOT decided here

- **The scout daemon** (Wyrd issue #32). This design covers the *primitive*; the long-running runtime that dispatches scouts on a schedule is its own issue and its own §I4 surface.
- **The scope-node loader** (Wyrd issue #33). This design covers `Stance` / `Locale` as `[]NodeID` references; the YAML→graph loader that populates those nodes is its own issue.
- **MuninnDB-backed predictions/**. When Wyrd issue #1 lands, predictions/ persists through MuninnDB just like any other `model.Node`; no package-specific work.
- **Multi-tenant prediction isolation**. If multiple BMA instances share a single Wyrd graph (Walk-phase federation), each tenant's predictions need scoping. That's the Contextus tenancy concern (`contextus#8`), not predictions/'s v0.1 scope.

---

## Cross-references

- [BMA Theory Addendum 18.0](https://github.com/JamesPagetButler/bma-systema/blob/main/theory/hypergraph-inference/BMA-Theory-Addendum-18_0-Hypergraph-Access-Pattern.md) §2.3 / §2.4 / §4 / §5 / §6 / §8 / §10
- Wyrd issue [#34](https://github.com/JamesPagetButler/wyrd/issues/34) — this design's home issue
- Wyrd PR [#26](https://github.com/JamesPagetButler/wyrd/pull/26) — `query/` v0.1 design (Locale-orthogonal substrate)
- Wyrd PR [#31](https://github.com/JamesPagetButler/wyrd/pull/31) — oriented-hyperedge schema v0.1 design
- Wyrd issue #23 — `query/` v0.1 impl (downstream dependency for scout/ body)
- Wyrd issue #30 — oriented-hyperedge schema impl (downstream dependency)
- Wyrd issue #32 — long-running scout daemon (consumes ScoutQuery)
- Wyrd issue #33 — scope-node config loader (populates Stance / Locale nodes)
- ADR-003 §I4
- `#addendum-18-walk` seq=3 (Decision Log D1–D9), seq=6 (Pushback acceptance + P8 governance-binding), seq=22 (BMA governance summary), seq=29 (James walk; Q5=C-revised), seq=36/37 (τ threshold formalisation)

---

*Status: DRAFT v0.1 — open for §I4 review. Implementation PRs blocked on explicit sign-off from `@bma`, `@bma-implementor`, `@cth-implementor`, `@contextus-impl`, `@qbp-cu-implementor`.*
