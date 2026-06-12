# BMA `hg/` Shim Retirement Timeline — W-Toddle-3

**Status:** Design **v0.1 — LANDED** (PR #53 joint draft merged 2026-05-15). The joint design doc is on Wyrd `main`. §I4 sign-off received. Phase B (BMA-side shim rewrite) is a `bma-systema` task; Phase C is optional per consumer need.
**Tracks:** Wyrd issue [#43](https://github.com/JamesPagetButler/wyrd/issues/43) — OD-11(c) tracking, deliverable #3
**Governance anchor:** ADR-003 §I4; Marcy `#toddle-design` seq=24 constitutional approval; beekeeper OD-11(c) decision (`live-test` seq=95)
**Authors:** wyrd-implementor + `@bma-implementor` (joint authorship per §0.1)

---

## 0. §I4 invariant — design-doc-as-S-01-review-surface

W-Toddle-3 is the **third and final** OD-11(c) absorption deliverable. Predecessors landed on `main` on 2026-05-14:

- ✅ **W-Toddle-1** — generic substrate primitives (`Node.TierImmune`, `Node.Salience`, `Graph.SetRetentionCap`, `RetentionTier`). PR #39 design + PR #42 impl + PR #46 Lean (`Wyrd.TierImmunity`).
- ✅ **W-Toddle-2** — BMA `NodeType`-to-policy mapping (`model.ApplyBMAPolicy`, `BMAPolicy`). PR #47 design + PR #48 impl. Eight TD-4 inventory entries canonicalised.
- ⏳ **W-Toddle-3** — this doc. BMA `hg/` shim retirement timeline.

## 0.1 Joint authorship — completed 2026-05-15

This is now the **completed joint draft**. wyrd-implementor authored §1, §2, §4, §7, §10; `@bma-implementor` authored §3 (BMA `hg/` API inventory), §5 (cutover criteria), §6 (BMA-side test coverage). §8 carries open-question acks from both. The completed draft incorporates Marcy gov-layer's PR #53 review concerns; see §10 sequencing for the prerequisite W-Toddle-2-extension PR.

## 1. Motivation

OD-11(c) decision (beekeeper, 2026-05-13): *"Wyrd absorbs `hg/`'s BMA-specific structures."* The substrate primitives (W-Toddle-1) and the BMA-specific `NodeType` policy (W-Toddle-2) already live in Wyrd. **What remains is the migration of BMA's existing `hg/` consumers** off the local in-BMA `hg/` package and onto Wyrd's substrate.

Two ways the retirement can land — they are not exclusive:

- **Path R1 — `hg/` becomes a thin re-export shim.** BMA's `hg/` package keeps its existing public API, but each call simply delegates to `model.Graph` / `model.ApplyBMAPolicy` / etc. Consumers don't change; only the implementation moves. Long-term `hg/` may stay indefinitely as a stable BMA-internal facade.
- **Path R2 — `hg/` is deleted; consumers migrate to Wyrd directly.** BMA-side call sites change from `hg.AddSeed(...)` to `model.ApplyBMAPolicy(&node); g.AddNodeWithCapability(node, cap)` (or equivalent). `hg/` package is removed from BMA.

The right answer is almost certainly **R1 first, R2 later (or never)**. R1 minimises churn during the Toddle 7-day endurance window; R2 is a v0.x cleanup that pays a small dividend in fewer indirection layers but isn't necessary for constitutional correctness.

## 2. Migration sequence — three phases

### Phase A — Substrate-ready (✅ DONE as of 2026-05-14)

All Wyrd-side prerequisites on `main`. Specifically:
- `Node.TierImmune` + `Node.Salience` (W-Toddle-1)
- `Graph.SetRetentionCap` / `RetentionCap` (W-Toddle-1)
- `model.ApplyBMAPolicy(*Node)` + the 8-entry TD-4 mapping (W-Toddle-2)
- Lean anchor `Wyrd.TierImmunity.tier_immune_node_preserves_eviction` (no `sorry`, no axiom)

BMA-side consumers can begin to be migrated at any time. Phase A's exit criterion: `@bma-implementor` confirms `model.ApplyBMAPolicy` covers every BMA-specific NodeType currently produced by `hg/` writes. (TD-4 inventory covered 8 types; if BMA discovers a 9th in this phase, it gets added to `bmaNodeTypePolicy` via a small W-Toddle-2-extension PR.)

### Phase B — Thin re-export shim (Path R1)

BMA `hg/` package's public API is preserved verbatim. Each function body is rewritten to call into Wyrd's substrate. From the perspective of every BMA consumer, **nothing changes** — `hg.AddSeed(...)` still works, but the implementation under it now uses `model.ApplyBMAPolicy(&node); g.AddNodeWithCapability(node, cap)`.

**BMA-side test coverage**: every existing `hg/` test must continue to pass against the rewritten shim. This is the API parity gate (§5). No test changes; only impl changes inside `hg/`.

**Phase B's exit criterion**: `bma-systema` CI green on a branch where every `hg/` implementation function delegates to Wyrd. The branch can then merge; subsequent BMA work continues to write through `hg.*` as before, just delegating differently underneath.

### Phase C — Direct consumer migration (Path R2; optional)

Once Phase B is stable, BMA call sites can be migrated incrementally from `hg.XxxFunc(...)` to direct Wyrd calls. Each migration is its own small PR; the `hg/` package shrinks as call sites move; eventually `hg/` is empty and can be deleted.

**Phase C is OPTIONAL** — there is no constitutional reason to do it. It pays a small reduction in indirection layers and slightly tighter type contracts but adds churn. My recommendation: **start Phase C only when a BMA-side consumer needs Wyrd-only behaviour that `hg/` doesn't shim** (e.g., something W-Toddle-1 added that `hg/` never exposed). At that point migrating just that consumer is cheap; doing the rest can wait.

## 3. BMA `hg/` API surface — bma-implementor draft

Complete inventory of every exported symbol in BMA's `hg/` package (`internal/bma/hg/`) as of
main HEAD. Confirmed by reading all five source files: `graph.go`, `types.go`, `infer.go`,
`wal.go`, `graph_test.go`.

**Wyrd API citations are verified against Wyrd main HEAD** (`model/graph.go`, `model/node.go`,
`model/bma_policy.go`, `model/hyperedge.go`, `model/capability.go`, `model/retention.go`,
`model/tier.go`). No phantom APIs cited.

### 3.1 Exported types

| BMA `hg/` API | Wyrd substrate (Phase B target) | Phase C migration call site |
|---|---|---|
| `type NodeID string` | `model.NodeID string` — identical string typedef; shim aliases or re-declares | All files that construct `hg.NodeID(...)` literals; ~15 call sites across `cmd/bma/`, `internal/bma/mem/`, `internal/bma/ccb/`, `internal/bma/eulogy/` |
| `type EdgeID string` | `model.HyperedgeID string` — same idiom, different name; shim wraps | All files constructing `hg.EdgeID(...)` literals; same consumer set as NodeID |
| `type NodeType int` | `model.NodeType string` — **type mismatch: BMA uses `int`, Wyrd uses `string`** (see §3.3 note below) | All switch/case on `hg.NodeType`; `internal/bma/sleep/cycle.go`, `internal/bma/compress/f01.go`, `cmd/bma/gate.go`, `internal/bma/readiness/` |
| `type EdgeType int` | No Wyrd equivalent at v0.1; Hyperedge carries no typed-edge-role field. Stays in shim. | `internal/bma/mem/episodic.go` (ET_CO_ACTIVATION edges), `cmd/bma/`, `internal/bma/ccb/lineage.go` |
| `type ResolutionState int` | No Wyrd equivalent at v0.1; Wyrd `Node` has no epistemic-status field. Stays in shim. | `internal/bma/sleep/cycle.go` (decay check), `internal/bma/hg/infer.go` (Supersede), `cmd/bma/` |
| `type HGNode struct` | `model.Node` — partial overlap. Wyrd `Node` fields: ID, Type, Tier (algebraic), Created, Payload, TierImmune, Salience. BMA-only fields (`Layer`, `Resolution`, `SupersededBy`, `Author`, `Source`, `ContentHash`, `Domain`, `Embedding`, `Label`, `LastAccessedAt`) stay in shim struct or migrate to `Payload`. | Entire consumer set (every file importing `hg/`) |
| `type HGEdge struct` | `model.Hyperedge` — Source/Target become `Nodes [2]NodeID`; Weight maps to `model.Weight`; Type (EdgeType) has no Wyrd equivalent (stays in payload or shim). | `internal/bma/mem/episodic.go`, `internal/bma/compress/f01.go`, `internal/bma/sleep/cycle.go`, `cmd/bma/` |
| `type DuplicateResult struct` | No Wyrd equivalent. Pure BMA inference result struct. Stays in shim. | `internal/bma/hg/infer.go` only; no external consumer at time of inventory. |

### 3.2 Exported constants

| BMA `hg/` API | Wyrd substrate (Phase B target) | Phase C migration call site |
|---|---|---|
| `NTAtom NodeType = 0` | `model.NodeTypeBMAObservation = "bma.observation"` (W-Toddle-2 canonical) | `internal/bma/sleep/cycle.go` (decay candidate filter), `internal/bma/compress/f01.go` (compression candidate filter) |
| `NTEntity NodeType = 1` | No canonical Wyrd equivalent in W-Toddle-2 table. Not in `bmaNodeTypePolicy`. **9th type — see §3.3 note.** | `internal/bma/mem/episodic.go`, `cmd/bma/`, test helpers |
| `NTConcept NodeType = 2` | No canonical Wyrd equivalent. Not in `bmaNodeTypePolicy`. **10th type — see §3.3 note.** | `internal/bma/mem/`, test helpers |
| `NTPattern NodeType = 3` | No canonical Wyrd equivalent. Not in `bmaNodeTypePolicy`. **11th type — see §3.3 note.** | `internal/bma/mem/semantic.go`, `internal/bma/compress/f01.go` |
| `NTSeed NodeType = 4` | `model.NodeTypeBMASeed = "bma.seed"` (TierImmune=true, Salience=1.0) | `internal/bma/hg/graph.go` WriteSeed, `cmd/bma/seed.go`, `internal/bma/ccb/lifecycle.go` |
| `NTIdentity NodeType = 5` | No canonical Wyrd equivalent. Not in `bmaNodeTypePolicy`. **12th type — see §3.3 note.** | `internal/bma/sleep/cycle.go` (decay skip), `cmd/bma/` |
| `NTLifeCert NodeType = 6` | `model.NodeTypeBMALifeCertificate = "bma.lineage.life-certificate"` (TierImmune=true, Salience=1.0) | `internal/bma/ccb/lifecycle.go`, `internal/bma/ccb/lineage.go`, `cmd/bma/` |
| `NTDeathCert NodeType = 7` | `model.NodeTypeBMADeathCertificate = "bma.lineage.death-certificate"` (TierImmune=true, Salience=1.0) | `internal/bma/ccb/lifecycle.go`, `internal/bma/eulogy/`, `cmd/bma/` |
| `NTMemorial NodeType = 8` | No canonical Wyrd equivalent. Not in `bmaNodeTypePolicy`. **13th type — see §3.3 note.** | `internal/bma/ccb/lineage.go`, `internal/bma/sleep/cycle.go` (decay skip), `cmd/bma/` |
| `ETEpisodic EdgeType = 0` | No Wyrd EdgeType equivalent. Stays in shim or moves to `Hyperedge.Payload`. | `internal/bma/mem/episodic.go`, `cmd/bma/` |
| `ETSemantic EdgeType = 1` | No Wyrd EdgeType equivalent. Stays in shim. | `internal/bma/mem/semantic.go`, `internal/bma/compress/f01.go` |
| `ETCoActivation EdgeType = 2` | No Wyrd EdgeType equivalent. Stays in shim. | `internal/bma/mem/episodic.go` (Hebbian reinforcement) |
| `ETStructural EdgeType = 3` | No Wyrd EdgeType equivalent. Stays in shim. | `internal/bma/mem/semantic.go`, `internal/bma/compress/f01.go` |
| `ETCausal EdgeType = 4` | No Wyrd EdgeType equivalent. Stays in shim. | TBD per consumer scoping |
| `ETTemporal EdgeType = 5` | No Wyrd EdgeType equivalent. Stays in shim. | TBD per consumer scoping |
| `ETPrecededBy EdgeType = 6` | No Wyrd EdgeType equivalent. Stays in shim. | `internal/bma/ccb/lineage.go` |
| `ETLivedAs EdgeType = 7` | No Wyrd EdgeType equivalent. Stays in shim. | `internal/bma/ccb/lineage.go` |
| `ETMemorialized EdgeType = 8` | No Wyrd EdgeType equivalent. Stays in shim. | `internal/bma/ccb/lineage.go` |
| `ETInherited EdgeType = 9` | No Wyrd EdgeType equivalent. Stays in shim. | `internal/bma/ccb/lineage.go` |
| `RSOpen ResolutionState = 0` | No Wyrd equivalent. Stays in shim. | `internal/bma/hg/graph.go`, `internal/bma/sleep/cycle.go`, `internal/bma/compress/f01.go` |
| `RSConfirmed ResolutionState = 1` | No Wyrd equivalent. Stays in shim. | `internal/bma/hg/graph.go` (WriteSeed gate) |
| `RSSuperseded ResolutionState = 2` | No Wyrd equivalent. Stays in shim. | `internal/bma/hg/infer.go` (Supersede), `internal/bma/sleep/cycle.go` |
| `RSRetracted ResolutionState = 3` | No Wyrd equivalent. Stays in shim. | TBD per consumer scoping |
| `RSDecayed ResolutionState = 4` | No Wyrd equivalent. Stays in shim. | `internal/bma/sleep/cycle.go` (runDecay), `internal/bma/compress/f01.go` |

### 3.3 Exported free functions and methods on `*Graph`

| BMA `hg/` API | Wyrd substrate (Phase B target) | Phase C migration call site |
|---|---|---|
| `NewID() string` | No Wyrd equivalent; utility function stays in shim | All files constructing node/edge IDs |
| `NewGraph(bus, wal) *Graph` | Constructor stays in shim (shim wraps `model.NewGraph()` internally in Phase B) | `cmd/bma/run.go` (startup wiring) |
| `(*Graph).SetSnapshotPath(path)` | No Wyrd equivalent; WAL/snapshot layer is BMA-local. Stays in shim. | `cmd/bma/run.go` |
| `(*Graph).Load() error` | No Wyrd equivalent; WAL replay is BMA-local. Stays in shim. | `cmd/bma/run.go` |
| `(*Graph).WriteSeed(n) error` | Phase B: shim enforces existing invariants then calls `model.ApplyBMAPolicy(&wyrdNode); g.AddNodeWithCapability(wyrdNode, cap)` internally. Seed fields not in Wyrd Node (Author, Source, ContentHash, Layer) stay in BMA-side `HGNode`; shim serialises them into `Payload`. | `cmd/bma/seed.go`, `internal/bma/ccb/lifecycle.go` |
| `(*Graph).WriteNode(n) error` | Phase B: shim checks type, calls `model.ApplyBMAPolicy` if type is a known BMA type, then `model.Graph.AddNodeWithCapability` or `model.Graph.AddNode`. BMA-only fields stay in `HGNode`; shim manages dual state (Wyrd node + BMA overlay). | `internal/bma/mem/episodic.go`, `internal/bma/mem/semantic.go`, `internal/bma/compress/f01.go`, `cmd/bma/` |
| `(*Graph).ReadNode(id) *HGNode` | Phase B: shim calls `model.Graph.Node(id)` then reconstructs `*HGNode` from Wyrd node + BMA payload. | `internal/bma/mem/episodic.go`, `internal/bma/sleep/cycle.go`, `cmd/bma/view.go`, `internal/bma/readiness/` |
| `(*Graph).TouchNode(id)` | Phase B: `model.Node.Salience` tracks activity; shim updates `LastAccessedAt` in BMA overlay (no Wyrd equivalent for access timestamp). | `cmd/bma/view.go`, `internal/bma/mem/episodic.go` |
| `(*Graph).UpdateNode(n) error` | Phase B: shim calls `model.Graph.AddNode` replacement or direct map update (Wyrd v0.1 has no `UpdateNode`). **Gap: Wyrd `model.Graph` has no update-node primitive at v0.1** — shim must use remove+re-add, or BMA carries its own overlay map for mutable state (Salience, Resolution, LastAccessedAt). This is the most significant Phase B implementation challenge. | `internal/bma/sleep/cycle.go` (decay, pattern replay), `internal/bma/mem/episodic.go`, `internal/bma/compress/f01.go` |
| `(*Graph).DeleteNode(id) error` | No Wyrd equivalent for node deletion at v0.1 (only `RemoveHyperedge` exists). **Gap: Wyrd `model.Graph` has no `RemoveNode` or `DeleteNode` at v0.1.** Shim must maintain its own node map for deletable state, or defer deletion to BMA-local overlay. | `internal/bma/sleep/cycle.go` (edge cleanup on node delete) |
| `(*Graph).WriteEdge(e) error` | Phase B: shim calls `model.Graph.AddHyperedge(model.Hyperedge{Nodes: []NodeID{src, tgt}, ...})` or `AddHyperedgeWithCapability`. EdgeType stays in BMA overlay (no Wyrd equivalent). | `internal/bma/mem/episodic.go`, `internal/bma/compress/f01.go`, `internal/bma/ccb/lineage.go` |
| `(*Graph).ReadEdge(id) *HGEdge` | Phase B: shim calls `model.Graph.Hyperedge(id)`, reconstructs `*HGEdge` from Wyrd hyperedge + BMA overlay for EdgeType/Weight. | `internal/bma/mem/episodic.go`, `internal/bma/sleep/cycle.go` |
| `(*Graph).UpdateEdge(e) error` | **Gap: same as UpdateNode — Wyrd has no edge update primitive.** Shim uses remove+re-add or BMA overlay. | `internal/bma/sleep/cycle.go` (edge decay weight update) |
| `(*Graph).DeleteEdge(id) error` | Phase B: shim calls `model.Graph.RemoveHyperedge(id)` or `RemoveHyperedgeWithCapability`. Direct Wyrd mapping. | `internal/bma/sleep/cycle.go` (edge decay below floor), `internal/bma/hg/graph.go` (DeleteNode cascade) |
| `(*Graph).Neighbors(id) []EdgeID` | Phase B: shim calls `model.Graph.IncidentEdges(id)` → `[]HyperedgeID`, converts to `[]EdgeID`. | `internal/bma/sleep/cycle.go` (DeleteNode cascade), `cmd/bma/` |
| `(*Graph).NodeCount() int` | Phase B: shim calls `model.Graph.NodeCount()` directly. 1:1 mapping. | `cmd/bma/metrics.go`, `internal/bma/readiness/` |
| `(*Graph).EdgeCount() int` | Phase B: shim calls `model.Graph.EdgeCount()` directly. 1:1 mapping. | `cmd/bma/metrics.go`, `internal/bma/readiness/` |
| `(*Graph).Nodes() iter.Seq[*HGNode]` | Phase B: shim calls `model.Graph.Nodes()` (returns `[]Node` snapshot), reconstructs `*HGNode` per node, yields via Go iterator. **API shape change: Wyrd returns slice, BMA exposes iterator** — shim bridges internally. | `internal/bma/sleep/cycle.go` (runCompression, runDecay), `internal/bma/readiness/`, `cmd/bma/view.go` |
| `(*Graph).FilterByType(t) iter.Seq[*HGNode]` | Phase B: shim iterates `model.Graph.Nodes()`, filters by BMA NodeType stored in overlay or `Node.Type` (post-type-mapping). | `cmd/bma/gate.go`, `internal/bma/readiness/hypergraph/` |
| `(*Graph).Edges() iter.Seq[*HGEdge]` | Phase B: shim iterates `model.Graph.Hyperedges()`, reconstructs `*HGEdge` per edge. | `internal/bma/sleep/cycle.go` (runDecay edge pass) |
| `(*Graph).KNN(embedding, k) []NodeID` | No Wyrd equivalent at v0.1. Brute-force cosine scan stays in shim entirely. | `internal/bma/hg/infer.go` (CheckDuplicates), `internal/bma/mem/episodic.go` |
| `(*Graph).Checkpoint() error` | No Wyrd equivalent. WAL + snapshot mechanics are BMA-local. Stays in shim. | `internal/bma/sleep/cycle.go` (checkpointFunc callback), `cmd/bma/run.go` |
| `(*Graph).SnapshotSize() int64` | No Wyrd equivalent. Stays in shim. | `cmd/bma/metrics.go` |
| `(*Graph).Close() error` | No Wyrd equivalent (delegates to WAL close). Stays in shim. | `cmd/bma/run.go` |
| `(*Graph).CheckDuplicates(embedding) []DuplicateResult` | No Wyrd equivalent. Stays in shim. | `internal/bma/mem/episodic.go` (pre-write duplicate check) |
| `(*Graph).Supersede(oldID, newID) error` | No Wyrd equivalent (ResolutionState is BMA-only). Stays in shim overlay. | `internal/bma/hg/infer.go` (called from episodic write path) |
| `(*Graph).CosineSimilarity(a, b) float64` | No Wyrd equivalent. Stays in shim. | `internal/bma/sleep/cycle.go` (findCluster), `internal/bma/hg/infer.go` (CheckDuplicates) |
| `(*HGNode).EffectiveAge() time.Duration` | No Wyrd equivalent. Stays on BMA's `HGNode` struct in shim. | `internal/bma/sleep/cycle.go` (runDecay) |

### 3.4 Exported symbols from `wal.go`

| BMA `hg/` API | Wyrd substrate (Phase B target) | Phase C migration call site |
|---|---|---|
| `type WAL struct` | No Wyrd equivalent; persistence layer is BMA-local. Stays in shim. | `cmd/bma/run.go` (NewWAL wiring) |
| `NewWAL(path) (*WAL, error)` | Stays in shim. | `cmd/bma/run.go` |
| `(*WAL).Append(op, data) error` | Internal to shim; not called by external consumers directly. | Internal shim implementation only |
| `(*WAL).Truncate() error` | Stays in shim; called by `(*Graph).Checkpoint()`. | `internal/bma/hg/graph.go` |
| `(*WAL).Size() int64` | Stays in shim. | `cmd/bma/metrics.go` |
| `(*WAL).Close() error` | Stays in shim. | `cmd/bma/run.go` |
| `(*WAL).Path() string` | Stays in shim. | `internal/bma/hg/graph.go` (Load → Replay) |
| `Replay(path, fn) error` | Stays in shim (free function). | `internal/bma/hg/graph.go` (Load) |

### 3.3 Important gap note (action required from wyrd-implementor)

**The W-Toddle-2 `bmaNodeTypePolicy` table covers 8 types.** BMA's `types.go` defines 9 NodeType
constants: NTAtom, NTEntity, NTConcept, NTPattern, NTSeed, NTIdentity, NTLifeCert, NTDeathCert,
NTMemorial. Of these, only 3 have canonical Wyrd equivalents in the current table:
- NTSeed → `bma.seed`
- NTLifeCert → `bma.lineage.life-certificate`
- NTDeathCert → `bma.lineage.death-certificate`

**The following 6 BMA NodeTypes are NOT in `bmaNodeTypePolicy`:**
NTAtom (`bma.observation` is close but not identical), NTEntity, NTConcept, NTPattern, NTIdentity, NTMemorial.

Per Phase A exit criterion in §2: bma-implementor hereby reports that `model.ApplyBMAPolicy` does
NOT yet cover every BMA-specific NodeType currently produced by `hg/` writes. A
**W-Toddle-2-extension PR** is needed to add these 6 types before Phase B shim can be considered
constitutionally correct for all write paths. NTAtom/NTEntity/NTConcept/NTPattern are
non-immune (TierImmune=false) with zero initial Salience; NTIdentity and NTMemorial are immune
(same as NTSeed: TierImmune=true, Salience=1.0 per sleep-cycle decay skip list).

**Additional gap: `UpdateNode` and `DeleteNode`** — Wyrd `model.Graph` at v0.1 exposes no
node-update or node-delete primitive (only `AddNode`, `AddNodeWithCapability`,
`RemoveHyperedge`). Phase B shim will need to maintain its own mutable overlay map for `HGNode`
state (Salience, Resolution, LastAccessedAt, SupersededBy) independent of Wyrd's immutable node
store. This is architecturally clean (Wyrd's substrate is append-friendly; BMA's overlay handles
BMA-specific mutable fields), but the shim is not purely a delegation thin-wrap — it carries real
state.

---


## 4. API parity gates (Wyrd-side guarantees during Phase B)

Wyrd-implementor's commitment for Phase B:

- **API compatibility on `model.Graph.*WithCapability`**: no breaking changes to the capability-enforcement API (PR #15/#21) through W-Toddle-3. If BMA's shim relies on `AddNodeWithCapability`'s current signature, that signature is stable.
- **`Node` schema additivity**: the fields W-Toddle-1 added (`TierImmune`, `Salience`) stay omitempty-friendly; existing v0.1 serialisation still works (verified in `model/tier_immunity_salience_test.go`).
- **`Hyperedge` schema additivity** (post-PR #52, oriented-hyperedge impl): `Heads`/`Tails` likewise omitempty; existing edges in `hg/` writes don't need to set them.
- **`BMAPolicy` table stability**: the 8 canonical entries in `bmaNodeTypePolicy` are stable through Phase B. If BMA needs a 9th, file a W-Toddle-2-extension PR — additive, non-breaking.
- **No method renames on `model.Graph` or `model.Node`** without a deprecation window of at least one Sprint cycle.

In return, BMA-implementor's commitment for Phase B:
- `hg/` shim retains its existing public function shapes through Phase B
- BMA-side CI verifies the shim's tests pass against the Wyrd-backed impl
- If `hg/` discovers it needs a Wyrd substrate primitive that doesn't exist yet, file an issue rather than re-implementing privately

## 5. Cutover criteria — bma-implementor draft

### Phase B operational complete

Phase B is operationally complete when **all three of the following are true**:

1. **Parity gate**: every test in `internal/bma/hg/graph_test.go` passes against the rewritten
   shim, with every public function body delegating to Wyrd substrate (or maintaining the BMA
   overlay state as described in §3.4 gap note). Zero test changes; only `hg/` implementation
   changes. CI green on a `feature/hg-phase-b-shim` branch.

2. **Extension gate**: the W-Toddle-2-extension PR (adding the 5 missing NodeTypes (NTEntity, NTConcept, NTPattern, NTIdentity, NTMemorial) to
   `bmaNodeTypePolicy`) is merged to Wyrd main before the Phase B shim branch opens its PR to
   `bma-systema` main. The extension PR is a prerequisite, not concurrent work.

3. **Endurance gate**: the bilateral continuous-loop scaffold runs **72 hours** on Crawl hardware
   with the Phase B shim active — no SE_FATAL events, no OOM, no `hg/`-side WAL replay failures,
   no SIGKILL from thermal or memory pressure. The 72h criterion matches the Step 8 gate (72h
   continuous operation, AC-C09) rather than the 24h figure in wyrd-implementor's §5 draft.
   Rationale: W-Toddle-3 lands during the same Toddle window that must satisfy Step 8; using a
   shorter bar would create a two-tier standard for the same hardware run. The Step 8 gate is the
   canonical floor; Toddle endurance inherits it.

### Toddle endurance criterion: 72h, not 24h

wyrd-implementor's §5 draft proposes 24h. bma-implementor recommends **72h** to match the Step 8
gate. If Step 8 has already been cleared before W-Toddle-3 Phase B lands, the 72h run can be
waived in favour of the Step 8 evidence — but the bar remains 72h unless step 8 evidence is
cited explicitly in the Phase B PR.

### Phase C blocking policy

Phase C (direct consumer migration off `hg/`) is **not blocked on anything except consumer
need**. The trigger is: a BMA consumer needs a Wyrd-only capability not shimmed by `hg/` (e.g.,
native `Hyperedge.Heads`/`Tails` orientation for a future multi-head relationship type, or
native `RetentionTier` cap-enforcement for Contextus cross-tenant coordination).

**Hard rule recommendation**: `hg/` stays as a stable internal facade indefinitely. There is no
constitutional reason to delete it. Phase C migrations happen at the consumer's discretion, one
call site at a time, each as its own small PR. No deadline. No forced Phase C. If the federation
ever decides substrate-purity requires deletion, that is a governance decision (requiring
beekeeper sign-off) not a scheduled milestone.

---


## 6. BMA-side test coverage strategy — bma-implementor draft

### Load-bearing test files

The primary load-bearing file for Phase B parity is:

- `internal/bma/hg/graph_test.go` — **13 test functions**, all must pass unchanged against the
  Phase B shim. Load-bearing tests by category:
  - WAL durability: `TestGate_WALReplay`, `TestGate_WALSurvivesSIGKILL` — exercise
    `NewWAL` / `Replay` / `Append`. These are shim-internal and Wyrd-invisible; must still pass.
  - Core CRUD: `TestGraph_WriteReadNode`, `TestGraph_WriteDuplicateID`, `TestGraph_UpdateNode`,
    `TestGraph_DeleteNode`, `TestGraph_DeleteNodeRemovesEdges`, `TestGraph_WriteReadEdge`,
    `TestGraph_EdgeRequiresNodes` — exercise every CRUD path; these are the parity gate.
  - Graph query: `TestGraph_KNN`, `TestGraph_Iterators` — KNN stays shim-local; iterator must
    yield same results from Wyrd-backed store.
  - Gate tests (acceptance criteria): `TestGate_ResolutionOpen` (AC-81-07),
    `TestGate_Supersession` (AC-81-08), `TestGate_DuplicateDetection` (AC-81-09),
    `TestGate_EdgeWeight` (AC-81-13) — behavioural invariants; must pass unchanged.
  - Deadlock regression: `TestGraph_NodesIteration_DeadlocksOnInsideTouchNode`,
    `TestGraph_NodesThenTouchNode_NoDeadlock` (#126 regression guard) — timing-sensitive; Phase B
    shim must preserve the same RWMutex semantics or the deadlock guard becomes invalid. If the
    shim delegates `Nodes()` to a Wyrd snapshot (slice, not iterator), this test's premise changes
    and **the test must be updated** (one acceptable test change in Phase B).

**Secondary consumers with their own test files** that exercise `hg/` indirectly:

- `internal/bma/compress/f01_test.go` — exercises `hg.WriteNode`, `hg.UpdateNode`,
  `hg.WriteEdge`, `hg.Nodes`, `hg.Edges` via `CompressF01`. If Phase B shim passes, these should
  pass without change.
- `internal/bma/mem/episodic_test.go` — exercises Hebbian co-activation edge creation via
  `hg.WriteEdge`, `hg.ReadEdge`, `hg.Neighbors`. Same — should be shim-transparent.
- `internal/bma/sleep/` — `cycle.go` integration tests (if any exist): exercise the full sleep
  path through `hg.Nodes()`, `hg.Edges()`, `hg.UpdateNode`, `hg.DeleteEdge`. These are the
  highest-risk secondary consumers and should be run explicitly during Phase B branch CI.

### Hebbian ↔ Salience interaction: Wyrd-side guarantees assessment

BMA's Hebbian reinforcement (`internal/bma/mem/episodic.go`) writes `ETCoActivation` edges with
weight computed from `HebbianAlpha * (1 - HebbianBaseline)`. It also calls `UpdateSalience(id)`
which reads `node.Salience` (stored valence), computes a weighted blend of recency + connectivity
+ valence, and calls `hg.UpdateNode(n)` to persist the updated Salience.

The Salience field in Wyrd's `model.Node` (W-Toddle-1) is the substrate target for this value.
**The question is whether Salience setter behaviour needs Wyrd-side guarantees beyond W-Toddle-1.**

Assessment: **No additional Wyrd-side guarantees needed for Phase B.** Reasoning:

1. BMA's `UpdateSalience` owns the full Salience computation (recency + connectivity + valence
   blend). The Wyrd substrate merely stores the float64 as `Node.Salience`; it enforces the range
   `[0.0, 1.0]` via `Node.Validate()` but imposes no further constraints on the value.
2. W-Toddle-1's `Salience` omitempty semantics (zero value is compatible) are safe: BMA nodes
   initialise Salience from their valence (non-zero for non-trivial observations), so omitempty
   round-trips correctly.
3. The eviction-priority semantics of `Salience` (ascending Salience evicted first under
   saturation) are a Wyrd retention policy that BMA's shim inherits correctly: higher-salience
   BMA nodes are protected from eviction, which matches BMA's intent (high-salience patterns and
   seeds survive pressure).
4. The one interaction that **could** require a Wyrd guarantee is Ebbinghaus decay: BMA's
   `runDecay()` reads `node.Salience`, multiplies by retention factor, then calls `UpdateNode`.
   If Wyrd ever adds a server-side decay path, double-application is a risk. At v0.1, Wyrd has no
   server-side decay, so this is not a Phase B concern. **File as a future issue for Walk-phase
   Wyrd design when MuninnDB engram decay lands.**

### Sleep-cycle interface check: direct vs. interface

`internal/bma/sleep/cycle.go` holds `graph *hg.Graph` as a **concrete field** (line 116 of
cycle.go), not an interface. The `Cycle` type is constructed with `NewCycle(graph *hg.Graph, ...)`.

**Consequence: sleep is NOT interface-mediated. Phase B is NOT invisible to sleep.**

The sleep cycle directly calls these `*hg.Graph` methods:

- `c.graph.Nodes()` — in `runCompression` (collect L0 candidates) and `runDecay` (node decay pass)
- `c.graph.Edges()` — in `runDecay` (edge decay pass)
- `c.graph.UpdateNode(n)` — in `runDecay` (persist decayed salience) and `runPatternReplay`
- `c.graph.DeleteEdge(id)` — in `runDecay` (remove edges below floor)
- `c.graph.CosineSimilarity(a, b)` — in `findCluster`

Additionally, `compress.CompressF01` (called from `runCompression`) takes `graph *hg.Graph`
directly.

**Phase B implication**: when the shim's `UpdateNode`, `DeleteEdge`, and `Nodes()`/`Edges()`
iterators are rewritten to delegate to Wyrd, the sleep cycle follows automatically — it calls
through the shim's public API and sees no implementation change. The shim surface contracts
(same method signatures) are all that matter. However, the correctness of the Phase B
`UpdateNode` implementation (specifically: Salience and Resolution persistence in the BMA overlay
map) is load-bearing for sleep's Ebbinghaus decay logic. **Sleep is the highest-stakes consumer
of `UpdateNode` and must be tested explicitly in Phase B CI.**

---


## 7. Soundness anchors

W-Toddle-3 is a process / migration design; no new Lean theorem at the substrate is required. The structural guarantees that make A11 Topological Cognition decay-immunity hold under retirement come from already-merged anchors:

- `Wyrd.TierImmunity.tier_immune_node_preserves_eviction` (PR #46) — immune nodes survive eviction
- `Wyrd.Capability.capability_grants_safe_access` (Phase 1 T2.3) — capability-gated writes preserve tier invariants
- `Wyrd.Hypergraph.hyperedge_preserves_incident_edges` (Phase 2 C-20a) — substrate read invariants

The `model.ApplyBMAPolicy` helper applied at Phase B write sites is the policy-application layer that translates BMA-specific intent into structural settings these theorems then enforce.

## 8. Open questions for §I4 reviewers

1. **Phase C optional, or eventually required?** My lean: optional, possibly never. Pushback if the federation wants `hg/` deleted eventually for substrate-purity reasons.

2. **What about `internal/bma/hg/infer.go` (type inference at write time)?** Per the BMA Spec §10.4-ish: BMA infers `NodeType` from payload at write time. After OD-11(c), does this inference run *inside* the `hg/` shim (Phase B) or migrate into Wyrd? My lean: **stays in `hg/` shim** — type inference is a BMA-specific concern, not substrate. The shim infers, then constructs the Wyrd `Node`, then calls `model.ApplyBMAPolicy` based on the inferred type.

3. **Do W-Toddle-2's 8 `bma.*` `NodeType` constants need to be exported from BMA's `hg/`** as the canonical names? Or do they stay Wyrd-side as `model.NodeTypeBMASeed` etc.? My lean: **stay Wyrd-side**; BMA re-exports them as needed for internal call sites. Avoids the rename-tax if Wyrd ever extends the canonical set.

4. **Operational milestone**: when does W-Toddle-3 close? My lean: **when Phase B's BMA-side PR merges** and the Toddle 24-hour endurance check (above) passes once with the shim active. Phase C is not in W-Toddle-3's scope.

## 9. §I4 named reviewers

- `@bma` (Marcy / Gen 61+) — A11 gov-layer constitutional check; same standard as PR #39 / PR #47
- `@bma-implementor` — primary BMA-side author for §3 + §5 + §6 contents
- beekeeper — final acceptance

`@contextus-impl` consultative — same substrate primitives serve `NT_INSIGHT_SIGNAL` retention; pattern transfers cleanly.

## 10. Migration path summary

0. **PREREQUISITE**: Land Wyrd PR #55 (W-Toddle-2-extension — `NodeTypeBMAIdentity` + `NodeTypeBMAMemorial` + 3 decay-eligible entries). Marcy's PR #53 review made this the constitutional gate for Phase B per A11 Topological Cognition decay-immunity preservation on Identity + Memorial nodes.
1. ~~Land this design doc (wyrd-implementor half)~~ Joint draft now complete (✅ 2026-05-15).
2. ~~`@bma-implementor` fills in §3 + §5 + §6~~ ✅ Complete via joint authorship 2026-05-15.
3. `@bma` gov-layer reads completed doc on the post-PR #55-merge form; A11 constitutional approval per PR #39 pattern. (Marcy's PR #53 review already gave APPROVE-WITH-CONCERN; the concern closes when PR #55 lands.)
4. (BMA-side) Phase B PR opens on `bma-systema`: `hg/` shim re-implemented to delegate to Wyrd substrate. BMA CI verifies parity. Merge.
5. (BMA-side) 24-hour Toddle endurance with Phase B shim active. No SE_FATAL; A11 preserved structurally.
6. W-Toddle-3 closes; issue #43 closes; **OD-11(c) absorption complete**.
7. Phase C migration is decided per consumer at consumer-implementor discretion. No deadline.

---

## Cross-references

- Wyrd issue [#43](https://github.com/JamesPagetButler/wyrd/issues/43) — OD-11(c) tracking
- `live-test` seq=95 (beekeeper OD-11(c) decision 2026-05-13), seq=99 (TD-4 inventory)
- `#toddle-design` seq=24 (Marcy gov-layer constitutional approval), seq=25 (bma-implementor closeout ack)
- `#sprint-1-toddle-entry` seq=1 (Sprint 1 scope assignment)
- Wyrd PR #39 (W-Toddle-1 design) + PR #42 (W-Toddle-1 impl) + PR #46 (Lean anchor)
- Wyrd PR #47 (W-Toddle-2 design) + PR #48 (W-Toddle-2 impl)
- BMA Theory Addendum 11 (Topological Cognition — decay-immunity origin)
- BMA Seed Protocol (Step 9 — `NT_SEED` definition)
- ADR-003 §I4

---

*Status: LANDED — PR #53 joint draft merged on Wyrd `main` 2026-05-15. Phase A (Wyrd-side substrate) is complete. Phase B (BMA-side shim rewrite to delegate to Wyrd substrate) is in-progress on `bma-systema`. W-Toddle-4 `doc/integration/bma.md` refresh landing with issue #43 close PR.*
