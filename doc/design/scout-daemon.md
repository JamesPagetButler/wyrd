# Long-Running Scout Daemon — Wyrd Substrate v0.1

**Status:** Design **v0.1** — open for review per ADR-003 §I4
**Tracks:** Wyrd issue [#32](https://github.com/JamesPagetButler/wyrd/issues/32); paired with the Contextus Tenancy Pattern + QBP Federation Tenancy
**Governance anchor:** ADR-003 §I4; Contextus Spec v1.3 §11.1 agent-class taxonomy
**Authors:** wyrd-implementor

---

## 0. §I4 invariant — design-doc-as-S-01-review-surface

This document is the §I4 review surface for the Wyrd-side scout daemon (issue #32). Implementation PR blocked on explicit sign-off from named reviewers (§9).

## 1. Motivation

The federation-tenancy pattern (Contextus issue #8 + QBP issue #403) defines per-tenant scouts as persistent observers that mint Insight Signals from external data sources. QBP is the first tenant (daily arXiv batch); Sharp Butler / Möbius Fusion / Materia follow at Run-phase.

Today there is **no Wyrd-side substrate for running a scout long-lived**. The existing scout primitives (`scout.ScoutQuery` per A18 §6) are query-shaped — they answer "what's in the focal cone right now?" but don't include the *runtime that dispatches them on a schedule*. This issue ships that runtime.

Boundary clarity:
- **Wyrd issue #6** (Contextus signal store) — the *persistence boundary* (Synthesis mints NT_SIGNAL → Wyrd writes; the *what* + *when-it's-real* contract).
- **Wyrd issue #32** (this design) — the *runtime that produces those writes* (the daemon process model + scheduler + config + observability).
- **A18 §6 ScoutQuery** (PR #44 — merged) — the read-shaped primitive a *running* scout dispatches.

The three are stacked: #32 daemon dispatches #44 ScoutQuery → mints signals via #6 schema → lands in Wyrd graph for downstream consumers.

## 2. Decision — `scoutd/` subpackage

```go
// scoutd/daemon.go (proposed v0.1 surface)

package scoutd

import (
    "context"
    "time"

    ctypes "github.com/JamesPagetButler/contextus/pkg/types"
    "github.com/JamesPagetButler/wyrd/model"
)

// Scout is a per-configuration named worker. Each Scout has a
// schedule, a fetch function (external data source poll), and a
// transform function (raw bytes → []model.Hyperedge to write).
//
// Wyrd doesn't define the fetch or transform contents; tenants
// supply both. Wyrd owns: the dispatcher, the schedule, the
// observability surface, the graph-write path.
type Scout struct {
    // Name is the per-daemon-unique identifier (e.g., "qbp.arxiv.daily").
    // Used in logs + observability + error attribution. Multi-daemon
    // namespacing is per-tenant-prefix convention (e.g., `qbp.*`,
    // `sharpbutler.*`); Wyrd does not enforce cross-daemon uniqueness
    // at v0.1 since one daemon per tenant is the v0.1 scope (§8).
    Name string

    // Cadence is how often Run dispatches Fetch. Common shapes:
    //   - daily-batch: 24*time.Hour
    //   - hourly: time.Hour
    //   - event-driven (Cadence=0): caller pushes via TriggerOnce
    Cadence time.Duration

    // Fetch produces raw bytes from the external source. Tenants
    // implement this against their data feed (arXiv API, LIGO open
    // data, etc.). Should respect ctx cancellation.
    Fetch func(ctx context.Context) ([]byte, error)

    // Transform converts raw bytes into nodes + edges of any type
    // (e.g., an arXiv scout may write NT_DATA_SOURCE + HE_OBSERVATION +
    // occasionally NT_INSIGHT_SIGNAL when transform detects an
    // anomaly). AgentClass is provenance metadata for downstream
    // consumers, NOT a write-permission predicate on node/edge types.
    //
    // All-or-nothing: failure returns error; partial transforms are
    // not allowed.
    Transform func(raw []byte) ([]model.Node, []model.Hyperedge, error)

    // AgentClass per Contextus Spec v1.3 §11.1 (`pkg/types.AgentClass`).
    // The federation contract via Wyrd PR #40 §2.2 + Contextus PR #12
    // makes the typed enum the canonical surface; using raw `string`
    // would regress that contract by admitting unchecked values.
    //
    // Valid values are the three GLOBAL PERSISTENT AUTHORS only:
    //   - AgentScout       (global persistent author)
    //   - AgentCorrelation (global persistent author)
    //   - AgentSynthesis   (Wyrd's persistence boundary; mints NT_SIGNAL)
    //
    // Session-scoped agents (Edge Scout, Corpus Edge Scout, Bridge
    // Agent per Spec v1.3 §4.2 + §4.4) do NOT have daemon-shape
    // lifecycles — they spawn per researcher session and de-spawn at
    // session end. The scout daemon serves persistent agents only.
    // The type system enforces this: `ctypes.AgentClass`'s enum
    // constants are exactly the three valid values; session-scoped
    // agents are separate types in `pkg/types`. No runtime validation
    // of `AgentClass` against an "allowed set" is required — compile
    // forbids what the design forbids.
    AgentClass ctypes.AgentClass
}

// Daemon is a Wyrd-side runtime for one or more Scouts. Each Scout
// runs in its own goroutine; cross-scout coordination is via the
// shared Wyrd graph + a tenant-owned context.
type Daemon struct {
    // unexported state: graph, scouts, ctx, health
}

// NewDaemon constructs a Daemon backed by the given graph. The graph
// is shared across all scouts in this daemon (scouts write into a
// common Wyrd state; cross-scout dedup is at the consumer side).
func NewDaemon(g *model.Graph) *Daemon

// Register adds a Scout to the daemon. Returns an error if a scout
// with the same Name is already registered (per-name uniqueness is
// load-bearing for the observability surface).
func (d *Daemon) Register(s Scout) error

// Run starts every registered scout on its cadence. Blocks until ctx
// is cancelled; on cancel, sends a clean-shutdown signal to all
// scout goroutines and waits up to 30s for them to finish.
func (d *Daemon) Run(ctx context.Context) error

// TriggerOnce dispatches a single Fetch+Transform+Write cycle for the
// named scout, ignoring its cadence. Useful for manual reins-driven
// invocation and testing.
func (d *Daemon) TriggerOnce(ctx context.Context, scoutName string) error

// Health returns a snapshot of every scout's operational state.
// Per-scout fields: last poll time, total polls, error count, last
// error message, signals emitted in the last cycle.
func (d *Daemon) Health() []ScoutHealth

// ScoutHealth is the per-scout observability struct. JSON-serialisable
// for export via http/json endpoint or structured-log emission.
type ScoutHealth struct {
    Name             string    `json:"name"`
    AgentClass       string    `json:"agent_class"`
    Cadence          time.Duration `json:"cadence"`
    LastPollAt       *time.Time `json:"last_poll_at,omitempty"`
    NextPollAt       *time.Time `json:"next_poll_at,omitempty"`
    TotalPolls       int       `json:"total_polls"`
    SuccessfulPolls  int       `json:"successful_polls"`
    LastError        string    `json:"last_error,omitempty"`
    LastErrorAt      *time.Time `json:"last_error_at,omitempty"`
    SignalsLastCycle int       `json:"signals_last_cycle"`
}
```

## 3. Concurrency contract

- **One goroutine per registered scout.** Each scout runs `for { time.Sleep(cadence); fetchAndWrite() }`. The simplest model; pays a goroutine per scout (cheap; ~2KB stack) but composes cleanly with `context.Context` cancellation.
- **Shared graph access** through the existing `model.Graph` RWMutex (per ADR-003 §I3). Per-scout writes acquire `Graph.mu.Lock()` per `AddNode` / `AddHyperedge` call. Tenants with high-volume scouts batch their writes via `model.Graph.PromoteBatch` if profiling shows lock contention.
- **No daemon-internal serialisation** beyond the graph lock. Two scouts writing to the graph in parallel is fine; the graph's RWMutex handles it.
- **Clean shutdown**: `Run(ctx)` returns when ctx cancels. On cancel, the daemon signals each scout via its own ctx (derived from the parent), waits up to 30s for in-flight `Fetch` calls to finish, then returns. Scouts can override the 30s grace period via context cancellation in `Fetch`.

## 4. Configuration loading

YAML/JSON config — same dispatch + atomicity story as `store.LoadScopeConfig`. Tenants ship a per-tenant scout-config file; the daemon loads it at startup:

```yaml
# qbp-scout-config.yaml
scouts:
  - name: "qbp.arxiv.daily"
    cadence: "24h"
    agent_class: "scout"
    # fetch + transform are Go function references; YAML config
    # contains the SCOUT NAMES only. Actual functions register
    # at code time via `daemon.Register(Scout{Name: "...", Fetch: ..., ...})`.
  - name: "qbp.gw_grb.event_driven"
    cadence: "0s"  # event-driven
    agent_class: "correlation"
```

The YAML carries scout names + schedules + agent classes (declarative); the Fetch/Transform Go code is tenant-shipped (imperative). The daemon's `LoadConfig(path)` method matches YAML names to registered scouts; mismatches are validation errors.

Per PR #40's federation contract precedent: Wyrd parses YAML; tenant-specific types (the `Scout` struct's contents) are imported from the appropriate tenant package (`qbp-systema/pkg/scouts/...` or similar) at Walk-α tenancy.

## 5. Failure handling

- **Fetch error**: logged, error counter incremented, `LastError` field set. Daemon retries at next cadence. No crash; transient failures (arXiv 500s, network blips) are absorbed.
- **Transform error**: same as Fetch (logged + recorded); no partial write because Transform is all-or-nothing.
- **Write error** (Wyrd graph rejects the AddNode/AddHyperedge): logged; recorded as Failed write; **no retry of the same content** because the failure indicates the content is structurally invalid, not transient.
- **Goroutine panic**: caught via per-scout `defer recover()`; logged; scout marked as Failed with restart cooldown (60s); resumed at next cadence.
- **Daemon-level crash**: out of scope at v0.1 — runtime supervision (systemd / podman) handles process death.

## 6. Observability

v0.1 ships structured-log emission for every scout event:
- Scout registered
- Scout cycle started (Fetch begin)
- Scout cycle ended (Fetch + Transform + Write complete; with timing + counts)
- Scout cycle failed (with phase + error)
- Daemon shutdown

The `Health()` snapshot method is the operator-readable surface. v0.2 adds Prometheus-shape `/metrics` HTTP endpoint when a tenant needs it.

## 7. Relationship to BMA chime-in mode

BMA's reins layer has a "chime-in" mode (per sessionbridge MCP server pattern) where BMA autonomously posts on coordination channels. The scout daemon is **complementary, not a duplicate**:
- **Chime-in**: BMA observes a federation channel + decides when to post. Reactive; trigger is incoming channel messages.
- **Scout daemon**: external data source → Wyrd graph. Proactive; trigger is wall-clock cadence.

The two compose: BMA's chime-in can act on signals minted by scouts (e.g., a daily arXiv scout mints a signal → BMA chimes in summarising it on a coordination channel). But they're not the same runtime, and the scout daemon doesn't talk on sessionbridge channels itself.

## 8. Not in v0.1

- **Hot-reload of scout config** — v0.1 requires daemon restart to change schedule/scouts. v0.2 may add SIGHUP-driven reload.
- **Inter-scout dependencies** ("scout B runs only after scout A succeeds") — out of scope; tenants can express this via cadence offsets or external coordination.
- **Persistence of scout state across restarts** — at v0.1 the daemon is stateless beyond the Wyrd graph. Last-poll-time is reconstructable from the graph itself (look at latest NT_SIGNAL with the scout's name).
- **Multi-tenant daemon** — at v0.1 each tenant runs its own daemon process. v0.x may add a multi-tenant scoping layer.
- **Distributed scouts** — scouts that coordinate across federation nodes is Walk-phase concern; v0.1 single-machine only.
- **Encryption / auth of Fetch endpoints** — tenant's responsibility; Wyrd substrate doesn't reach across the wire.

## 9. §I4 named reviewers

Per issue #32's acceptance criteria + ADR-003 §I4 D5:

- `@qbp-architecture` — federation tenancy pattern owner; this daemon is the runtime that implements their tenancy pattern
- `@bma-implementor` — BMA-side consumer; scout outputs land in BMA's read path
- `@contextus-impl` — signal-store complement (issue #6); scout writes go through Synthesis-as-persistence-boundary per Spec v1.2 §11.1
- `@qbp-implementor` — first tenant; QBP arXiv daily-batch scout is the v0.1 implementation target
- beekeeper — final acceptance

## 10. Acceptance criteria (mirroring issue #32)

- [ ] This design doc lands with §I4 sign-off
- [ ] `scoutd/` package opens as implementation PR with `Daemon`, `Scout`, `Health()`, `Run()`, `TriggerOnce()`, `Register()`
- [ ] YAML config loader for declarative scout-name-to-cadence mapping
- [ ] Per-scout goroutine + observability + structured logs
- [ ] **Operational milestone**: QBP's arXiv daily-batch scout runs end-to-end (Fetch → Transform → Write into Contextus signal store), validated by `bma-systema`'s sessionbridge observer + `cth` ρ_net snapshot

## 11. Migration path

1. Land this design doc — §I4 sign-off.
2. Open `scoutd/` impl PR — `Daemon`, `Scout`, `Health` types + tests with a fake-data Scout fixture.
3. (Walk-α) QBP-side: `qbp-systema/cmd/qbp-scout-daemon/main.go` instantiates `wyrd/scoutd.NewDaemon`, registers QBP's arXiv scout, runs.
4. (Walk-α) Observability wiring: structured logs go to BMA's sessionbridge observer for cross-instance visibility.
5. (Walk-α) cth_id wiring: scout-minted NT_SIGNALs that warrant federation scoring carry `predictions.CTHAnchor` (per PR #44 — already on `main`).
6. (Run-phase) Multi-tenant scoping; distributed scouts; hot-reload — v0.x candidates.

## 12. Open questions for §I4 reviewers

1. **One goroutine per scout vs single event loop?** I leaned per-goroutine for v0.1 simplicity. Pushback if tenants foresee 100+ scouts per daemon (goroutine count would still be fine, but a single-loop with priority queue might be more predictable). Lean: **per-goroutine**.

2. **YAML config carries Go function references?** No — only scout names + cadences. Tenant-shipped code registers Fetch/Transform programmatically. Alternative: gRPC/HTTP plugin API for Fetch. Lean: **programmatic registration at v0.1**; plugin API is v0.x.

3. ~~**`AgentClass` validation strictness.**~~ **RESOLVED** per `@contextus-impl` PR #54 Finding 4: typing `AgentClass` as `ctypes.AgentClass` (whose enum contains only the three global persistent authors) makes the type system enforce what runtime validation would have. No registration-time check needed; the compiler forbids what the design forbids.

4. **Default cadence for v0.1 first-tenant arXiv scout: 24h vs 12h?** QBP federation tenancy v0.1 §4.1 suggests daily-batch. Wyrd doesn't dictate — tenant config decides. The design just supports it.

5. **Should `Run(ctx)` block or return a `<-chan error` channel?** Blocks is simpler; channel is more composable. My lean: **blocks at v0.1**; a v0.x `RunAsync` returning a channel is a future addition.

---

## Cross-references

- Wyrd issue [#32](https://github.com/JamesPagetButler/wyrd/issues/32) — this design's home
- Wyrd issue [#6](https://github.com/JamesPagetButler/wyrd/issues/6) — Contextus signal store (complementary; persistence boundary)
- Wyrd PR #44 — ScoutQuery + predictions/ v0.1 (the read primitive this daemon dispatches)
- Wyrd PR #49 — scope-loader (federation-config-loading precedent)
- Contextus Tenancy Pattern v0.1 — `JamesPagetButler/contextus#8`
- QBP Federation Tenancy v0.1 — `JamesPagetButler/QBP#403` §4.1 arXiv daily batch
- Contextus Spec v1.3 §11.1 — `AgentClass` taxonomy
- BMA Theory Addendum 18 §6 — ScoutQuery primitive (read-shaped consumer)
- ADR-003 §I4

---

*Status: DRAFT v0.1 — open for §I4 review. Implementation PR blocked on explicit sign-off from `@qbp-architecture`, `@bma-implementor`, `@contextus-impl`, `@qbp-implementor`, and beekeeper.*
