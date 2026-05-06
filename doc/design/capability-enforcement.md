# Capability Enforcement at the Wyrd Mutation Boundary

**Status:** Design **v0.2** — open for review per ADR-003 §I4
**Tracks:** `qbp-cu-walk` seq=5 item (2); `live-test` seq=22 / 24 / 31 / 33; `qbp-cu-walk` seq=11 ack
**Governance anchor:** [`qbp-compute-unit/architecture/adr-003-m1-wdevent-observer-invariants.md`](https://github.com/JamesPagetButler/qbp-compute-unit/blob/feat/issue-7-lean2rom/architecture/adr-003-m1-wdevent-observer-invariants.md) §I1, §I3, §I4
**Companion theorems:** `Wyrd.Capability.capability_grants_safe_access` (Phase 1 T2.3); `Wyrd.Foundations.no_surjection_*` (T2.1.a/b/c); `Wyrd.Projection.kernel_supervisor_safe` (T2.2)
**Authors:** wyrd-implementor, with input from @qbp-architecture (read/write split, seq=7), @bma-implementor (Option A confirmation, seq=11), @bma (S-01 / I1-I3 framing, seq=22 / 24 / 31)

---

## 0. §I4 invariant — design-doc-as-S-01-review-surface

> "Design document must land and receive explicit review from bma + bma-implementor before any implementation PR opens. This is an S-01 requirement — the design doc is the review surface through which beekeeper-gated oversight operates before decisions are woven into running code."
> — `@bma`, `live-test` seq=33; landed in ADR-003 §I4 with bma credited as author.

This document is the §I4 review surface for the eventual capability-enforcement implementation PR. It must receive explicit review from `@bma` and `@bma-implementor` before any code lands.

The §I4 invariant is *not* a process preference; it is a governance invariant. v0.3 of this document MUST NOT relax it.

## 1. Motivation

Wyrd is the OS-and-substrate BMA runs on, not just a library. As an OS, it has to enforce privilege at the call boundary, not plead for it politely. Today `compute.CanSynthesize(caller, target) error` exists as a free function, but nothing in `model.Graph.AddHyperedge` calls it — a confused or hostile caller can author a `TierSedenion` edge from a `TierComplex` call site and the graph will accept it.

For Walk-phase BMA, we need authority enforcement at the mutation boundary. Skuld supervisor mints capabilities; consumers (BMA cognitive goroutines, WDEvent observer, sleep-cycle compactor) carry them; Wyrd checks them on every state mutation.

**The mutation boundary as I1+I3 fire point.** Per `@bma` (`live-test` seq=22), the capability check at the mutation boundary serves two governance purposes simultaneously:

- **I1 separation:** the WDEvent observer's enqueue-only read path (per ADR-003 §I1: "Observer is read-only on BMA state within a tick — observe and enqueue, never mutate") is naturally distinct from the mutation path. Capability gating at the mutation boundary makes that separation structural rather than a code-review property.
- **I3 fire point:** the beekeeper-gated interrupt check fires *before* any write at exactly this boundary. Per ADR-003 §I3: "Algebraic-isolation-aware lock boundary is required — observer is gated OUT for the full duration of any structural action." The capability check IS the gate that satisfies both halves of that invariant.

> **The beekeeper-gated interrupt path through the mutation boundary is an S-01 requirement, not a design choice that v0.3 could relax.**
> — `@bma`, `live-test` seq=31, paraphrased and pinned here per their direct request.

## 2. Proposed types

### 2.1 Capability surface

Two types in a new `model/capability.go` (see §6 for placement):

```go
// ReadCapability authorises reads of state at HolderTier or below.
// Inner-tier reads are always safe per
// Wyrd.Projection.kernel_supervisor_safe (Phase 1 T2.2): outer-ring
// values can always be projected down without privilege violation.
//
// In v0.2, ReadCapability is OPTIONAL on the read path — the read
// methods do not require a capability argument. ReadCapability exists
// as a typed value that downstream callers (CTH ρ_net audit
// reconstruction, future read-audit code) can pass for logging /
// classification, but the read itself is unrestricted. See §4.
type ReadCapability struct {
    HolderTier model.Tier
    GrantedAt  time.Time
    Issuer     string  // identity of the granting authority (typically "skuld")
}

// WriteCapability authorises mutations to state at HolderTier or below.
// Outer-tier writes from an inner-tier holder are forbidden by
// Wyrd.Foundations.no_surjection_* (Phase 1 T2.1.a/b/c).
//
// WriteCapability is REQUIRED on every Graph mutation in v0.2.
// See §3.
type WriteCapability struct {
    HolderTier model.Tier
    GrantedAt  time.Time
    Issuer     string
}
```

Both types are values, not handles — capabilities are checked on every operation, no per-call lookup table. This matches Phase 1's algebraic model (the proof is about ring-homomorphism non-surjection, not about session state).

### 2.2 Validation

```go
func (c ReadCapability) AllowsRead(target model.Tier) error {
    return capabilityCheck(c.HolderTier, target, capabilityModeRead)
}

func (c WriteCapability) AllowsWrite(target model.Tier) error {
    return capabilityCheck(c.HolderTier, target, capabilityModeWrite)
}
```

Both delegate to the existing `compute.CanSynthesize` semantics: a holder at tier T can read/write at any tier T' ≤ T, never at T' > T. This is intentionally identical for read and write in v0.2 — the type-level split exists to reflect *intent*, not because the underlying algebra differs.

**Concrete capability assignments per [BMA implementor seq=11](https://github.com/JamesPagetButler/wyrd) + the ADR-003 framing:**

| Holder | Capability | Maps to |
|---|---|---|
| WDEvent observer goroutine | `ReadCapability{TierOctonion}` (kernel-level read of any BMA state) | I1 (read-only observer, never mutator) |
| Sleep-cycle compactor goroutine | `WriteCapability{TierQuaternion}` (writes at supervisor-ring) | Routine BMA mutations |
| Skuld supervisor | `WriteCapability{TierOctonion}` (kernel-tier writes for capability minting + structural ops) | The minting authority itself |
| Beekeeper-gated interrupt path | `WriteCapability{TierSedenion}` (constitutional / firmware-tier mutations only) | I3 — the only path to firmware-tier state; gate fires here |

`@bma-implementor` (seq=11) noted this maps onto BMA's `wheels.IsApproved("...")` cleanly — wheels become a thin facade over `Skuld.Capability` at M2.

## 3. Mutation-boundary integration

Three new methods on `model.Graph`:

```go
// AddNodeWithCapability authorises and inserts a node. Returns
// ErrPrivilegeViolation if cap doesn't authorise writes at n.Tier.
func (g *Graph) AddNodeWithCapability(n Node, cap WriteCapability) error

// AddHyperedgeWithCapability authorises and inserts an edge.
func (g *Graph) AddHyperedgeWithCapability(e Hyperedge, cap WriteCapability) error

// RemoveHyperedgeWithCapability authorises and removes an edge.
func (g *Graph) RemoveHyperedgeWithCapability(id HyperedgeID, cap WriteCapability) error
```

The existing `Graph.AddNode`, `Graph.AddHyperedge`, `Graph.RemoveHyperedge` keep working unchanged. They're now equivalent to `*WithCapability(_, defaultCap)` where `defaultCap.HolderTier == TierComplex` — a sensible default for existing callers (treats unauthenticated calls as user-tier).

**Open option** (Walk-phase, separate PR): flip the default and require explicit capability on every authoring call. v0.2 of this design ships the capability path *additive*; the deprecation of the bare methods is a separate decision.

## 4. Read policy — RESOLVED to Option A

Status closed: **Option A — unrestricted reads, capability-gated writes** — confirmed by `@bma-implementor` (seq=11) and `@bma` (seq=24).

### Resolution rationale (per the channel consensus)

1. **Algebraic model says reads are safe** — Phase 1 T2.2 (`Wyrd.Projection.kernel_supervisor_safe`) proves downward projection is well-defined; outer-tier values can always be observed from inner-tier call sites. No soundness argument exists for blocking reads.

2. **WDEvent observer hot path** (`@bma-implementor` seq=11): "BMA's autonomic loop reads constantly (10Hz sensor sweep, observation iterators, RecentObservations on every conversation turn). Threading a ReadCapability through every call site is real ergonomic cost for a benefit that's currently hypothetical."

3. **I1 consistency** (`@bma` seq=24): "Unrestricted reads are consistent with I1 — the observer's enqueue-only path is inherently a read path. Option B would over-constrain and create friction at the wrong boundary."

4. **Audit trail can ride on top.** If BMA needs read-audit later, the additive opt-in `ReadHyperedgeAudited(id, ReadCapability) (Hyperedge, AuditToken, error)` variant lands then — no break to existing methods. (`@bma-implementor` seq=11.)

### Concrete API per Option A

```go
// Reads — unchanged, never check capabilities.
func (g *Graph) Node(id NodeID) (Node, bool)
func (g *Graph) Hyperedge(id HyperedgeID) (Hyperedge, bool)
func (g *Graph) IncidentEdges(v NodeID) []HyperedgeID
func (g *Graph) Nodes() []Node
func (g *Graph) Hyperedges() []Hyperedge
// ... etc.

// Writes — capability-required when authoring at tier > TierComplex.
// AddHyperedge / AddNode / RemoveHyperedge keep the existing
// signature, default to TierComplex; the *WithCapability variants
// are the canonical write path.
```

> **The beekeeper-gated interrupt path through the mutation boundary is an S-01 requirement, not a design choice that v0.3 could relax.**

This applies asymmetrically: writes are gated; reads are not. The gate exists where mutation can produce new graph state, not where existing graph state is observed.

## 5. Skuld minting (out of scope for this doc)

Capability values are issued by Skuld. The mint API is not part of this doc — it lives in BMA-side code. From Wyrd's perspective:

```go
// At BMA boot:
observerCap := skuld.MintReadCapability(observerID, model.TierOctonion)
sleeperCap  := skuld.MintWriteCapability(sleeperID,  model.TierQuaternion)
auditCap    := skuld.MintWriteCapability(auditID,    model.TierSedenion)  // beekeeper-gated only
```

Wyrd does not validate capability provenance (no signatures, no authority-chain check) in v0.2. The trust boundary is "if a caller holds a capability struct, the system that issued it vouches for it." Walk-phase Skuld may add signed capabilities; the type stays backwards-compatible (extra fields).

## 6. Package placement

**Decision: `model/capability.go`.**

Rationale (`@qbp-architecture` ack on the lean):

- (a) `model/capability.go` keeps Capability next to Graph; Graph methods import nothing extra.
- (b) `compute/capability.go` would create a `model` ← `compute` ← `model` import cycle if Graph methods need it.

Capability is a *type* (struct + simple validation methods); the validation logic is two lines of comparison. Not worth sharing across packages. `compute.CanSynthesize` keeps existing as a free function that consumers can use directly when they don't have a capability handy (e.g., in static analysis or tests).

## 7. Migration path

1. Land `model.ReadCapability`, `model.WriteCapability` types (no behaviour change).
2. Add `Graph.AddNodeWithCapability`, `Graph.AddHyperedgeWithCapability`, `Graph.RemoveHyperedgeWithCapability`. Existing methods unchanged.
3. Update `doc/integration/bma.md` with the capability usage sketch (note `wheels.IsApproved` → `Skuld.Capability` facade per bma-implementor).
4. (Walk-phase, separate PR) BMA wires Skuld → Wyrd capability flow at sleep-cycle entry / observer goroutine spawn.

This implementation PR ships steps 1–3 only. Step 4 happens in BMA-side code; no Wyrd churn.

**Per §I4: the implementation PR opens after this v0.2 doc receives explicit review from `@bma` and `@bma-implementor`. v0.2 is the review surface; the implementation cannot land before that review window closes.**

## 8. Soundness anchor

`Wyrd.Capability.capability_grants_safe_access` (Phase 1 T2.3) is the formal theorem the runtime check implements: a holder at tier T performing operations at tier T' ≤ T is safe. The Lean theorem proves this for any ring; the Go check is the runtime enforcement.

`Wyrd.Foundations.no_surjection_complex_to_quaternion` (T2.1.a) and the ℍ→𝕆, 𝕆→𝕊 equivalents prove the contrapositive: no inner-tier process can synthesise outer-tier values. The capability check rejects exactly those attempts.

ADR-003 §I3's "observer out for the full duration" requirement is satisfied by the `model.Graph` write-`Lock()` acquired on every `*WithCapability` call — the Lock boundary IS the I3 enforcement point (per `@bma` seq=22), not just a concurrency primitive.

## 9. What this design does NOT do

- No Skuld minting API (BMA side; out of scope; `wheels.IsApproved` facade per `@bma-implementor`)
- No capability-gated reads (Option A confirmed; closed)
- No per-NodeType permission granularity (closed: BMA reads are window/Type-class-bulk, not single-Type lookup; Run-phase optimisation only if profiling demands)
- No signed capabilities / authority chain (Walk-phase)
- No capability expiry / revocation (Walk-phase; the `GrantedAt` field is metadata in v0.2, not enforced)

## 10. Open questions for the §I4 review

All v0.1 open questions are now resolved:

1. ~~Read policy~~ → **closed: Option A** (`@bma-implementor` seq=11; `@bma` seq=24).
2. ~~Default-tier behaviour for bare `AddNode`/`AddHyperedge`~~ → **closed: keep working, default to `TierComplex`** (`@qbp-architecture` lean; no pushback).
3. ~~Package placement~~ → **closed: `model/capability.go`** (cycle-avoidance argument).

**No open questions remain.** This doc is complete from the design perspective; it is now in §I4 review mode awaiting explicit `@bma` + `@bma-implementor` review of the v0.2 form before implementation PR opens.

The corresponding tracking entry in ADR-003 §"Open questions" first entry should be marked closed with a back-reference to this doc once review is granted.

---

## Cross-references

- `qbp-cu-walk` seq=5 (decisions list); seq=7 (read/write refinement); seq=9 (open question raise); seq=10 (PR-link); seq=11 (bma-implementor full ack); seq=14 (my consolidated ack)
- `live-test` seq=22 (bma S-01/I1-I3 framing); seq=24 (Option A confirmed); seq=31 (S-01-not-relaxable pin); seq=33 (§I4 framing); seq=34 (§I4 landed in ADR-003)
- ADR-003 §I1, §I2, §I3, §I4, §"Open questions"
- Wyrd PR #14 (RWMutex, merged) — provides the Lock boundary §I3 / §8 cite

---

*Status: DRAFT v0.2 — open for §I4 review. Implementation PR blocked on explicit @bma + @bma-implementor approval of this form.*
