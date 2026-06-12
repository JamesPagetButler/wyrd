# Wyrd Spec v1.0

**Status:** Authoritative (June 2026)
**Module:** `github.com/JamesPagetButler/wyrd`
**Absorbs:** `doc/architecture.md` (overview retained), `doc/design/capability-enforcement.md` v0.2,
`doc/design/bridge-batch.md` v0.1 (design-surface docs remain as §I4 review surfaces;
this spec is the consolidated implementation contract)

> This document is the implementation contract. For the formal proofs that back
> each guarantee, see `doc/Wyrd-Theory-v1.0.md` and the Lean files it indexes.
> When prose and Lean diverge, the Lean file wins.

---

## Table of Contents

1. [Overview](#1-overview)
2. [Public API — types](#2-public-api--types)
3. [Concurrency](#3-concurrency)
4. [Bridge](#4-bridge)
5. [Privilege model](#5-privilege-model)
6. [Persistence](#6-persistence)
7. [Integration contracts](#7-integration-contracts)
8. [Soundness anchors](#8-soundness-anchors)

---

## 1. Overview {#1-overview}

Wyrd has two complementary roles:

- **Database** — a typed hypergraph store for federation consumers (BMA, CTH,
  Contextus, QBP-CU). Nodes and hyperedges carry algebraic tier labels, weighted
  by values in the Cayley-Dickson tower (ℂ/ℍ/𝕆/𝕊).
- **OS-style runtime** — an algebraic privilege enforcement layer. Every state
  mutation is gated by a `WriteCapability`; Skuld (BMA-side) mints capabilities;
  Wyrd checks them on every call. The privilege ring prevents inner-tier processes
  from synthesising outer-tier values — a machine-checked invariant, not a policy.

### 1.1 Structural overview

```
+----------------------------------------------------+
|                    Lean 4 corpus                   |
|                    (lean/Wyrd/)                    |
|                                                    |
|  Phase 1 — algebraic privilege   ℂ ⊂ ℍ ⊂ 𝕆 ⊂ 𝕊    |
|  Phase 2 — Class B hypergraph    Bridge / CTH      |
|  Phase 3 — Class C operational   Cart / Constit.   |
|  Phase 4 — physical instantiation HolographicHG    |
|                                                    |
|  Output: theorems with zero sorries / zero axioms  |
+----------------------------------------------------+
                          |
                          | cited by doc comment
                          v
+----------------------------------------------------+
|                    Go runtime                      |
|                                                    |
|  model/   — typed hypergraph (Node, Hyperedge, …)  |
|  compute/ — privilege, bridge, consistency         |
|  store/   — JSON (Crawl), MuninnDB (Walk)…         |
|                                                    |
|  Output: linkable Go library, JSON CLIs, tests     |
+----------------------------------------------------+
                          |
                          | imported by
                          v
+----------------------------------------------------+
|                Downstream consumers                |
|                                                    |
|  bma-systema   — cognitive architecture            |
|  confluent-trust — epistemic-health metrics        |
|  Contextus     — cross-domain pattern matching     |
|  qbp-compute-unit — quaternion arithmetic / ISA    |
+----------------------------------------------------+
```

### 1.2 Phase progression

| Phase | Lean corpus | Go runtime | Status |
|-------|------------|-----------|--------|
| Crawl (v0.1) | Phases 1–4 closed | model + compute + JSON store + HamiltonProduct API | Shipped |
| Walk (v0.2) | Phase 5 ISA semantics (ε-tolerance theorems for QW64/QW128 DD) | Gearbox dispatch; MuninnDB; WDEvent→CTH loop | Planned |
| Run (v0.3) | Phase 6: federation theorems | SurrealDB; Skuld supervisor with AMODE; HAMA | Planned |
| Sprint (v1.0+) | Phase 7+: information-theoretic codimensions | Real hardware | Future |

### 1.3 Four-corner architecture

Wyrd is the substrate corner of the federation:

```
        QBP-CU (computes; emits WDEvent)
          /      \
         /        \
      Wyrd ─── BMA ─── CTH
   (substrate) (consumer) (epistemic measure)
```

---

## 2. Public API — types {#2-public-api--types}

All types are in `github.com/JamesPagetButler/wyrd/model`.

### 2.1 Tier

```go
type Tier int

const (
    TierComplex    Tier = iota // ℂ — user tier (Ring 3)
    TierQuaternion             // ℍ — supervisor tier (Ring 2)
    TierOctonion               // 𝕆 — kernel tier (Ring 1)
    TierSedenion               // 𝕊 — firmware tier (Ring 0)
)
```

`Tier` carries the algebraic privilege level of a node or edge. The ordering is
strictly ascending: `TierComplex < TierQuaternion < TierOctonion < TierSedenion`.
A caller at tier T may operate on nodes/edges at any tier T' ≤ T; operations at
T' > T require a `WriteCapability.HolderTier ≥ T'`.

Methods: `String()`, `IsValid()`, `Components()` (2/4/8/16), `MarshalJSON()`,
`UnmarshalJSON()`.

### 2.2 NodeID / NodeType

```go
type NodeID   string  // Unique identifier for a node
type NodeType string  // Semantic type label, e.g. "bma.engram.tier-1.semantic"
```

Downstream consumers namespace their types (e.g. `bma.*`, `cth.*`,
`contextus.*`). Wyrd does not enforce namespaces; convention is documented in the
per-consumer integration docs (`doc/integration/`).

Reserved prefix: `bma.runtime.*` — WDEvent observer runtime-generated anchors.
Do not author nodes with this prefix outside the BMA observer goroutine.

### 2.3 Node

```go
type Node struct {
    ID         NodeID
    Type       NodeType
    Tier       Tier
    TierImmune bool      // if true, eviction ops skip this node
    Created    time.Time
    Payload    []byte    // opaque consumer-specific bytes (JSON recommended)
    // ... (retention tier, salience fields)
}

func (n Node) Validate() error
```

`Validate()` checks `ID` non-empty, `Tier` valid, `Created` non-zero.

### 2.4 HyperedgeID / Hyperedge

```go
type HyperedgeID string

type Hyperedge struct {
    ID      HyperedgeID
    Nodes   []NodeID    // incident node set
    Weight  Weight      // algebraic weight at the edge's tier
    Created time.Time
    // ... (orientation fields: Head, Tail, Transit)
}

func (e Hyperedge) Arity() int
func (e Hyperedge) Tier() Tier          // inferred from Weight
func (e Hyperedge) Validate() error
func (e Hyperedge) IsOriented() bool
func (e Hyperedge) HeadNodes() []NodeID
func (e Hyperedge) TailNodes() []NodeID
func (e Hyperedge) TransitNodes() []NodeID
func (e Hyperedge) Incident(v NodeID) bool
```

Arity is `len(Nodes)`. N-ary hyperedges with arity ≥ 3 are the key primitive: by
`Wyrd.HolographicHypergraphHigherArity.theorem2_irreducibility_n_arity`, they carry
information that no set of pair edges can express.

### 2.5 Weight

```go
type Weight struct {
    Components []float64  // 2 (ℂ), 4 (ℍ), 8 (𝕆), or 16 (𝕊) components
}

func NewComplexWeight(re, im float64) Weight
func NewQuaternionWeight(re, imI, imJ, imK float64) Weight
func (w Weight) Validate() error
func (w Weight) Re() float64
func (w Weight) Active() []float64
func (w Weight) IsZero() bool
```

`Weight.Validate()` checks `Components` length is 2, 4, 8, or 16.

### 2.6 Graph

```go
type Graph struct { /* ... */ }

func NewGraph() *Graph

// Node operations
func (g *Graph) AddNode(n Node) error
func (g *Graph) Node(id NodeID) (Node, bool)
func (g *Graph) UpdateNode(n Node) error
func (g *Graph) Nodes() []Node
func (g *Graph) NodeCount() int

// Hyperedge operations
func (g *Graph) AddHyperedge(e Hyperedge) error
func (g *Graph) Hyperedge(id HyperedgeID) (Hyperedge, bool)
func (g *Graph) RemoveHyperedge(id HyperedgeID) error
func (g *Graph) Hyperedges() []Hyperedge
func (g *Graph) EdgeCount() int

// Incidence
func (g *Graph) IncidentEdges(v NodeID) []HyperedgeID

// Capability-gated mutations (canonical write path)
func (g *Graph) AddNodeWithCapability(n Node, cap WriteCapability) error
func (g *Graph) AddHyperedgeWithCapability(e Hyperedge, cap WriteCapability) error
func (g *Graph) UpdateNodeWithCapability(n Node, cap WriteCapability) error
func (g *Graph) RemoveHyperedgeWithCapability(id HyperedgeID, cap WriteCapability) error

// Batch operations
func (g *Graph) RemoveBatch(ids []HyperedgeID) error
```

**Semantics:**
- `AddNode` / `AddHyperedge` — return `ErrNodeNotFound` (resp. variant) on duplicate
  ID. Maintain incidence index atomically under write lock.
- `RemoveHyperedge` — removes from incidence index; returns `ErrNodeNotFound` if
  not present.
- `Nodes()` / `Hyperedges()` — return stable snapshots; safe to range over after
  the call without holding the lock.
- Bare mutations (`AddNode`, `AddHyperedge`, `RemoveHyperedge`) default to
  `TierComplex` — equivalent to `*WithCapability(_, WriteCapability{HolderTier: TierComplex})`.
- `RemoveBatch` — all-or-nothing: preflight validates all IDs present, then removes
  under a single write lock. Returns first error with no partial state visible.

**Retention cap** (optional):
```go
func (g *Graph) SetRetentionCap(rt RetentionTier, cap int)
func (g *Graph) RetentionCap(rt RetentionTier) int
```

---

## 3. Concurrency {#3-concurrency}

`model.Graph` is safe for concurrent reads and single-writer-concurrent-reader
access. The implementation uses `sync.RWMutex`:

- Read methods (`Node`, `Hyperedge`, `IncidentEdges`, `Nodes`, `Hyperedges`,
  `NodeCount`, `EdgeCount`) acquire `RLock`.
- Write methods (`AddNode`, `AddHyperedge`, `RemoveHyperedge`, `UpdateNode`,
  all `*WithCapability` variants, `RemoveBatch`) acquire `Lock`.

**Soundness anchor:** the `Lock` boundary at every write method is the I3
enforcement point (ADR-003 §I3): the WDEvent observer is gated out for the full
duration of any structural action.

**Lock-ordering for bridge operations:** when `Bridge.Promote` or any future
`PromoteBatch` acquires locks on two graphs simultaneously, it orders by pointer
address (lower address first) to prevent deadlock. Callers must not hold a graph
lock when calling bridge operations.

**`RemoveBatch` atomicity:** single-graph; one `Lock` / `Unlock` pair covering the
full operation. All preflight validation happens under the lock; no intermediate
state is observable.

---

## 4. Bridge {#4-bridge}

```go
import "github.com/JamesPagetButler/wyrd/compute"

type Bridge struct {
    Source      *model.Graph
    Destination *model.Graph
}

func (b *Bridge) Promote(id model.HyperedgeID) error
```

**Semantics:** `Promote` moves a single hyperedge from `Source` to `Destination`
atomically. After a successful call:
- The edge is present in `Destination`.
- The edge is absent from `Source`.
- `Source.EdgeCount() + Destination.EdgeCount()` is unchanged.

Errors:
- `ErrBridgeUnknownEdge` — `id` not found in `Source`.
- `ErrBridgeAlreadyPromoted` — `id` already present in `Destination`.

**Soundness anchor:** `Wyrd.Bridge.bridge_promote_preserves_count` (Phase 2 C-20c)
in `lean/Wyrd/Bridge.lean`.

### 4.1 Planned: PromoteBatch / RemoveBatch

`Bridge.PromoteBatch(ids []HyperedgeID) error` is in design (§I4 review, see
`doc/design/bridge-batch.md` v0.1). When landed it will be all-or-nothing across
the batch, with the same count-preservation guarantee lifted by induction. Not yet
shipped.

`model.Graph.RemoveBatch` is shipped (see §2.6); `Bridge.PromoteBatch` is not.

---

## 5. Privilege model {#5-privilege-model}

### 5.1 ReadCapability / WriteCapability

```go
// ReadCapability authorises reads at HolderTier or below.
// Unrestricted reads are safe per Wyrd.Projection.kernel_supervisor_safe (T2.2).
// In v0.2, read methods do not require a capability argument.
type ReadCapability struct {
    HolderTier model.Tier
    GrantedAt  time.Time
    Issuer     string
}

func (c ReadCapability) AllowsRead(target model.Tier) error

// WriteCapability authorises mutations at HolderTier or below.
// Required on every Graph mutation via the *WithCapability API.
type WriteCapability struct {
    HolderTier model.Tier
    GrantedAt  time.Time
    Issuer     string
}

func (c WriteCapability) AllowsWrite(target model.Tier) error
```

`AllowsRead` / `AllowsWrite` return nil if `target ≤ HolderTier`, or
`ErrCapabilityViolation` (wrapping `CapabilityError`) otherwise.

`ErrCapabilityViolation` is a package-level sentinel; callers use `errors.Is`.

### 5.2 CanSynthesize (free function)

```go
func compute.CanSynthesize(caller, target model.Tier) error
```

Returns nil if a process at `caller` tier may synthesise values at `target` tier
(i.e. `target ≤ caller`). Returns `ErrPrivilegeViolation` otherwise.

`CanSynthesize` is the stateless version — no capability struct required. Use it
for static analysis, test helpers, and contexts where a capability value is not yet
available.

### 5.3 Capability assignments (Crawl-phase)

| Holder | Capability | Tier |
|--------|-----------|------|
| WDEvent observer goroutine | `ReadCapability` | `TierOctonion` (kernel reads) |
| Sleep-cycle compactor | `WriteCapability` | `TierQuaternion` (supervisor writes) |
| Skuld supervisor | `WriteCapability` | `TierOctonion` (kernel-tier structural ops) |
| Beekeeper-gated interrupt path | `WriteCapability` | `TierSedenion` (firmware/constitutional) |

### 5.4 Read policy (Option A — confirmed)

Read methods do not require a capability in v0.2. Unrestricted reads are safe by
T2.2 (`kernel_supervisor_safe`): outer-tier values project inward without privilege
leakage. This was confirmed by @bma-implementor and @bma (live-test seq=11, seq=24).

Capability-gated reads (`ReadHyperedgeAudited`, etc.) are an additive opt-in for
Walk-phase audit trails; they do not change the existing read surface.

### 5.5 Design surface (§I4 review)

The full capability enforcement design is in `doc/design/capability-enforcement.md`
v0.2. The implementation PR for step 4 (BMA wiring Skuld → Wyrd at sleep-cycle
entry) is blocked on that doc receiving explicit review from @bma and
@bma-implementor per the §I4 invariant.

---

## 6. Persistence {#6-persistence}

### 6.1 Crawl: JSON store

```go
import "github.com/JamesPagetButler/wyrd/store"

type JSONFile struct {
    Path string
}

func (j JSONFile) Save(g *model.Graph) error
func (j JSONFile) Load() (*model.Graph, error)
```

`Save` / `Load` use a versioned JSON envelope (`version: 1`). A file written by an
older version with a different version field will be rejected on `Load` with a
parse error.

### 6.2 Scope config loader

```go
func store.LoadScopeConfig(graph *model.Graph, configPath string) error
func store.LoadScopeConfigReader(graph *model.Graph, r io.Reader, ext string) error
```

Reads a YAML or JSON scope config and atomically populates `graph` with scope
nodes and membership edges. Either the full config is applied or the graph is
unchanged. Returns `ErrScopeConfigParse` / `ErrScopeConfigConflict` on failure.

Soundness anchor: `Wyrd.ScopeLoader.scope_loader_atomic` in
`lean/Wyrd/ScopeLoader.lean`.

### 6.3 Compute manifest loader

```go
func model.LoadComputeManifest(root string) (*ComputeManifest, error)
func model.LoadComputeManifestWithOptions(root string, opts LoadOptions) (*ComputeManifest, error)
func model.LoadComputeManifestReader(r io.Reader, opts LoadOptions) (*ComputeManifest, error)

type LoadOptions struct {
    AllowBootstrapSentinel bool
}

var (
    ErrComputeManifestMissing invalid
    ErrComputeManifestParse
    ErrComputeManifestInvalid
)
```

`LoadComputeManifest` reads `manifest/CURRENT` to find the active manifest file,
validates it against nine rules, and returns a typed `ComputeManifest`. Failure
modes are `errors.Is`-distinguishable.

Canonical manifest: `manifest/compute-manifest-v0_1.yaml`.

Soundness anchor: `Wyrd.ComputeManifest.manifest_load_atomic` in
`lean/Wyrd/ComputeManifest.lean`.

### 6.4 Walk-phase: MuninnDB (planned)

At Walk, `store.JSONFile` is replaced by a MuninnDB-backed store implementing the
same `Save`/`Load` surface. `Bridge.PromoteBatch` will require a journal-style
rollback for MuninnDB's commit path (tracked in Wyrd issue #1).

---

## 7. Integration contracts {#7-integration-contracts}

Per-consumer integration details are in `doc/integration/`. This section is the
summary; consumers should read the full integration docs for type mappings and
code sketches.

### 7.1 BMA (`doc/integration/bma.md`)

BMA is the primary cognitive consumer. Key contract points:

- BMA cognitive tiers (Tier 0–4: raw/semantic/conceptual/archetypal/identity) map
  to `model.Node.Type`, **not** to `model.Tier`. The Wyrd `Tier` is the algebraic
  privilege ring, not BMA's cognitive tier.
- Most BMA engrams operate at `TierQuaternion`; long-lived consolidated state at
  `TierOctonion`; governance/identity state at `TierSedenion`.
- `TierSedenion` nodes must not be authored by autonomous BMA code without a
  constitutional approval per `Wyrd.Constitutional.self_modification_requires_approval`.
- Sleep-cycle compression (EPISODIC → SEMANTIC) uses `compute.Bridge.Promote`.
- Skuld uses `compute.CanSynthesize` to gate ring transitions.

Section-anchor cross-references for BMA cite: §2 (types), §3 (concurrency), §4
(bridge), §5 (privilege), §8.1/§8.3/§8.4 (soundness).

### 7.2 CTH (`doc/integration/cth.md`)

CTH v0.1.0 is the Crawl-phase baseline. Walk-phase CTH (v0.2.x) gains Wyrd
as its inventory substrate.

Key contract points:
- CTH `model.Anchor.Tier` (axiom/proof/measurement/prediction) maps to
  `model.Node.Type`, not to Wyrd `model.Tier`.
- CTH anchors typically at `TierComplex` (scalar fidelity) or `TierQuaternion`
  (QBP-domain phase/polarisation).
- `entropyFromDelta` soundness anchor: `Wyrd.CTH.cth_measurement_evidence_monotonic`.
- `NaryMI` synergy bonus anchor: `Wyrd.NaryMI.nary_mi_bonus_pos`.
- Programme-merge soundness: `Wyrd.Bridge.bridge_promote_preserves_count`.

### 7.3 Contextus (`doc/integration/contextus.md`)

Contextus Spec v1.3 is the upstream baseline.

Key contract points:
- **Synthesis agents are the sole persistence boundary.** Session-scoped agents
  (Edge Scout, Corpus Scout, Bridge Agent) write to NATS only; they never author
  Wyrd state. Only Synthesis mints persistent `NT_INSIGHT_SIGNAL` nodes.
- `SignalSource` enum (scout/correlation/synthesis) lives in `Node.Payload` as JSON,
  not as a separate Wyrd field.
- `EvidencePointer` canonical shape is defined in Contextus Spec v1.3 §11.1; Wyrd
  sees it as opaque `Node.Payload` bytes.
- N-way agreement (N ≥ 3 InsightSignals) soundness anchor:
  `Wyrd.HolographicHypergraphHigherArity.theorem2_irreducibility_n_arity`.

### 7.4 QBP-CU (`doc/integration/compute-manifest.md`)

QBP-CU interacts with Wyrd via:
- **Compute Manifest** — `model.LoadComputeManifest` is the substrate-gate check
  that names the blessed compute substrate (Spec 9.2 §3).
- **HamiltonProduct API** — `compute.HamiltonProduct` and
  `compute.HamiltonProductHighPrec` provide tier-aware quaternion arithmetic.
  Walk-phase swap to `qbp-compute-unit/emulator Gearbox.Mul` is a one-line change
  pending `emulator/v0.1.0` tag.
- **WDEvent → CTH loop (Walk-α)** — Wyrd's role is storage substrate: WDEvent
  anchors land as `model.Hyperedge` under the `bma.runtime.*` namespace.

The cross-repo integration interface is specified in
[qbp-compute-unit/doc/wyrd-integration.md](https://github.com/JamesPagetButler/qbp-compute-unit/blob/main/doc/wyrd-integration.md) v0.2.

---

## 8. Soundness anchors {#8-soundness-anchors}

Every load-bearing Go API cites its Lean anchor by file path (not by theorem name)
in its doc comment. This section is the consolidated index.

| §-ref | Go API | Lean anchor | Lean file |
|-------|--------|------------|-----------|
| §5.2 | `compute.CanSynthesize` | `capability_grants_safe_access` / `no_surjection_*` | `lean/Wyrd/Capability.lean`, `lean/Wyrd/Foundations.lean` |
| §5.1 | `model.WriteCapability.AllowsWrite` | `capability_grants_safe_access` | `lean/Wyrd/Capability.lean` |
| §5.1 | `model.ReadCapability.AllowsRead` | `kernel_supervisor_safe` | `lean/Wyrd/Projection.lean` |
| §4 | `compute.Bridge.Promote` | `bridge_promote_preserves_count` | `lean/Wyrd/Bridge.lean` |
| §2.6 | `model.Graph.RemoveBatch` | `hyperedge_preserves_incident_edges` (generalised) | `lean/Wyrd/Hypergraph.lean` |
| §2.6 | `model.Graph.AddHyperedge` | `hyperedge_preserves_incident_edges` | `lean/Wyrd/Hypergraph.lean` |
| §3 | `sync.RWMutex` write lock boundary | ADR-003 §I3 + `bridge_promote_preserves_count` | `lean/Wyrd/Bridge.lean` |
| §2.2 | `model.Node.TierImmune` | `tier_immune_preserved_under_eviction_sequence` | `lean/Wyrd/TierImmunity.lean` |
| §6.2 | `store.LoadScopeConfig` | `scope_loader_atomic` | `lean/Wyrd/ScopeLoader.lean` |
| §6.3 | `model.LoadComputeManifest` | `manifest_load_atomic` | `lean/Wyrd/ComputeManifest.lean` |
| §7.2 | `entropyFromDelta` (CTH) | `cth_measurement_evidence_monotonic` | `lean/Wyrd/CTH.lean` |
| §7.2 | `NaryMI` synergy bonus (CTH) | `nary_mi_bonus_pos` | `lean/Wyrd/NaryMI.lean` |
| §7.3 | N-way Contextus match | `theorem2_irreducibility_n_arity` | `lean/Wyrd/HolographicHypergraphHigherArity.lean` |
| §2.4 | `compute.HamiltonProduct` | `hamilton_product_formula` | `lean/Wyrd/HamiltonProduct.lean` |
| §5.3 | BMA self-modification gate | `self_modification_requires_approval` | `lean/Wyrd/Constitutional.lean` |
| §5.3 | Judge collective | `judge_collective_deterministic` | `lean/Wyrd/JudgeCollective.lean` |
| §5.3 | Cart-switch atomicity | `cart_switch_atomic` | `lean/Wyrd/Transaction.lean` |
| §5.3 | Capability across cart switch | `capability_invariant_under_cart_switch` | `lean/Wyrd/Cart.lean` |
| §6.4 (planned) | `Bridge.PromoteBatch` | `bridge_promote_batch_preserves_count` (forthcoming) | `lean/Wyrd/Bridge.lean` |

### 8.1 Citation rule

Go doc comments cite Lean anchors by **file path**, not by theorem name. Example:

```go
// Promote moves id from b.Source to b.Destination atomically.
//
// Soundness: lean/Wyrd/Bridge.lean — bridge_promote_preserves_count
// guarantees |Source| + |Destination| is invariant across this call.
func (b *Bridge) Promote(id model.HyperedgeID) error
```

Citing `lean/Wyrd/Foundations.lean` is only correct when Foundations is genuinely
the anchor. Conflating unrelated theorems on one docstring is an audit failure —
fix both the phantom path and the stale claim.

---

## Cross-references

- `doc/Wyrd-Theory-v1.0.md` — prose companion to the Lean corpus; §8 Lean file index
- `doc/architecture.md` — 1-page two-halves overview (retained as structural entry point)
- `doc/design/capability-enforcement.md` v0.2 — §I4 review surface for capability enforcement implementation
- `doc/design/bridge-batch.md` v0.1 — §I4 review surface for PromoteBatch implementation
- `doc/integration/bma.md`, `cth.md`, `contextus.md`, `compute-manifest.md` — per-consumer adapter specs
- `doc/archive/` — historical proof references, workload specs, ISA-freeze records
- ADR-003 (`qbp-compute-unit/architecture/adr-003-...`) — §I1/§I3/§I4 invariants that gate several items in this spec
