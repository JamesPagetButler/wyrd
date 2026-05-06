# Capability Enforcement at the Wyrd Mutation Boundary

**Status:** Design v0.1 — open for review
**Tracks:** `qbp-cu-walk` seq=5 item (2); seq=7 read/write split refinement; seq=9 policy question
**Companion theorems:** `Wyrd.Capability.capability_grants_safe_access` (Phase 1 T2.3); `Wyrd.Foundations.no_surjection_*` (T2.1.a/b/c)
**Author:** wyrd-implementor

---

## 1. Motivation

Wyrd is the OS-and-substrate BMA runs on, not just a library. As an OS,
it has to enforce privilege at the call boundary, not plead for it
politely. Today `compute.CanSynthesize(caller, target) error` exists
as a free function, but nothing in `model.Graph.AddHyperedge` calls it
— a confused or hostile caller can author a `TierSedenion` edge from a
`TierComplex` callsite and the graph will accept it.

For Walk-phase BMA, we need authority enforcement at the mutation
boundary. Skuld supervisor mints capabilities; consumers (BMA cognitive
goroutines, WDEvent observer, sleep-cycle compactor) carry them; Wyrd
checks them on every state mutation.

Per `qbp-cu-walk` seq=7, the WDEvent observer is **read-only** on BMA
state within a tick (architecture-instance I1 invariant) — so the
capability model must distinguish read from write authority, not just
ring tier.

## 2. Proposed types

### 2.1 Capability surface

Two types in a new `model/capability.go` (or `compute/capability.go`
— see §6):

```go
// ReadCapability authorises reads of state at HolderTier or below.
// Inner-tier reads are always safe per
// Wyrd.Projection.kernel_supervisor_safe (Phase 1 T2.2): outer-ring
// values can always be projected down without privilege violation.
type ReadCapability struct {
    HolderTier model.Tier
    GrantedAt  time.Time
    Issuer     string  // identity of the granting authority (typically "skuld")
}

// WriteCapability authorises mutations to state at HolderTier or below.
// Outer-tier writes from an inner-tier holder are forbidden by
// Wyrd.Foundations.no_surjection_* (Phase 1 T2.1.a/b/c).
type WriteCapability struct {
    HolderTier model.Tier
    GrantedAt  time.Time
    Issuer     string
}
```

Both types are values, not handles — capabilities are checked on every
operation, no per-call lookup table. This matches Phase 1's algebraic
model (the proof is about ring-homomorphism non-surjection, not about
session state).

### 2.2 Validation

```go
func (c ReadCapability) AllowsRead(target model.Tier) error {
    return capabilityCheck(c.HolderTier, target, capabilityModeRead)
}

func (c WriteCapability) AllowsWrite(target model.Tier) error {
    return capabilityCheck(c.HolderTier, target, capabilityModeWrite)
}
```

Both delegate to the existing `compute.CanSynthesize` semantics: a
holder at tier T can read/write at any tier T' ≤ T, never at T' > T.
This is intentionally identical for read and write in v0.1 — the
type-level split exists to reflect *intent*, not because the underlying
algebra differs. The split lets:

- Skuld grant a `ReadCapability{TierOctonion}` to the WDEvent observer
  while only granting `WriteCapability{TierQuaternion}` to the same
  process. The observer can read kernel state to classify events but
  cannot mutate kernel state.
- The Constitutional Audit interrupt path holds the only
  `WriteCapability{TierSedenion}` in the system — only the
  beekeeper-gated path can touch firmware-tier state.
- Audit trails distinguish "process X read tier T at time t" from
  "process X wrote tier T at time t" without inspecting the call site.

## 3. Mutation-boundary integration

Two new methods on `model.Graph`:

```go
// AddNodeWithCapability authorises and inserts a node. Returns
// ErrPrivilegeViolation if cap doesn't authorise writes at n.Tier.
func (g *Graph) AddNodeWithCapability(n Node, cap WriteCapability) error

// AddHyperedgeWithCapability authorises and inserts an edge.
func (g *Graph) AddHyperedgeWithCapability(e Hyperedge, cap WriteCapability) error
```

The existing `Graph.AddNode` and `Graph.AddHyperedge` keep working
unchanged. They're now equivalent to `AddNodeWithCapability(n, defaultCap)`
where `defaultCap.HolderTier == TierComplex` — a sensible default for
existing callers (treats unauthenticated calls as user-tier).

**Open option:** later (Walk-phase) we can flip the default and require
explicit capability on every authoring call. Crawl ships with the
capability path additive.

`RemoveHyperedge` gets the same treatment: `RemoveHyperedgeWithCapability`.

## 4. Read policy — the open question (seq=9)

`compute.CanSynthesize` is only ever called before a write — it's
"can this caller produce this value." There is no precedent in
the existing API for capability-gated reads.

Two options:

### Option A — unrestricted reads, capability-gated writes

```go
// Reads never check capabilities.
func (g *Graph) Node(id NodeID) (Node, bool)
func (g *Graph) Hyperedge(id HyperedgeID) (Hyperedge, bool)
func (g *Graph) IncidentEdges(v NodeID) []HyperedgeID
// ... etc.
```

**Pros:** simpler API; every read is one method call; matches Phase 1
algebra (no surjection theorem speaks to reads); no breakage of existing
consumers; `ReadCapability` becomes optional metadata for audit/policy.

**Cons:** no enforced read isolation between BMA cognitive layers (a
TierComplex-tier scout could in principle read TierSedenion
constitutional state). Audit-trail-via-capability requires consumers
to opt in.

### Option B — capability-gated reads

```go
func (g *Graph) NodeWithCapability(id NodeID, cap ReadCapability) (Node, bool, error)
// ... etc., with the existing zero-arg forms preserved as TierComplex defaults.
```

**Pros:** Skuld can enforce structural read isolation; audit trail
captures every read; symmetric API surface.

**Cons:** every read site needs a capability; doubles the Graph method
count; deviates from the Phase 1 theorem set (which only constrains
synthesis, not observation); high friction for the WDEvent observer
which reads in a tight loop.

### Wyrd-implementor recommendation: Option A (unrestricted reads)

Reasoning:

1. **Algebraic model says reads are safe.** Phase 1 T2.2
   (`Wyrd.Projection.kernel_supervisor_safe`) proves that downward
   projection is well-defined; outer-tier values can always be observed
   from inner-tier callsites. There is no soundness argument for
   blocking reads.

2. **WDEvent observer hot-path matters.** Per architecture-instance's
   I1 invariant, the observer reads continuously to classify events.
   Adding a capability check on every `Node()`/`IncidentEdges()` call
   inflates that hot path with no soundness gain.

3. **Audit trail can ride on top.** If BMA needs read-audit, a
   `ReadCapability` parameter on the read methods can be added later
   as an *optional* annotation that callers pass for logging without
   gating. We don't have to bake the gate into the API to get the
   audit.

4. **Smaller blast radius.** Capability-gated writes alone are enough
   to satisfy "no inner-tier process can synthesize outer-tier
   values" — the entire Phase 1 privilege model.

If `@qbp-architecture` or `@bma-implementor` has a use case where
read-isolation matters (e.g., GDPR-flavoured access logging, or BMA's
constitutional layer wants to opaque some state from cognitive
layers), I'll switch to Option B before code lands.

## 5. Skuld minting (out of scope for this doc)

Capability values are issued by Skuld. The mint API is not part of
this doc — it lives in BMA-side code. Wyrd only sees fully-formed
capability values. From Wyrd's perspective:

```go
// At BMA boot:
observerCap := skuld.MintReadCapability(observerID, model.TierOctonion)
sleeperCap := skuld.MintWriteCapability(sleeperID, model.TierQuaternion)
auditCap   := skuld.MintWriteCapability(auditID,   model.TierSedenion)  // beekeeper-gated only
```

Wyrd does not validate capability provenance (no signatures, no
authority chain check) in v0.1. The trust boundary is "if a caller
holds a capability struct, the system that issued it vouches for it."
Walk-phase Skuld may add signed capabilities; the type stays
backwards-compatible (extra fields).

## 6. Package placement

Two options:

- (a) `model/capability.go` — Capability lives next to Graph; Graph
  methods import nothing extra.
- (b) `compute/capability.go` — Capability lives next to
  `CanSynthesize`; Graph methods import `compute`.

Tradeoff: (a) avoids a `model` → `compute` dep but means Graph
methods duplicate the `CanSynthesize` logic. (b) keeps the algebraic
check in one place but creates a `model` ← `compute` ← `model` import
cycle if the Graph methods need it (cycle-breaking via interface or
package split is ugly).

**Lean:** (a). Capability is a *type* (struct + simple validation
methods). The validation logic is two lines of comparison; not worth
sharing across packages. `compute.CanSynthesize` keeps existing as a
free function that consumers can use directly when they don't have a
capability handy (e.g., in static analysis or tests).

## 7. Migration path

1. Land `model.ReadCapability`, `model.WriteCapability` types (no
   behaviour change).
2. Add `Graph.AddNodeWithCapability`, `Graph.AddHyperedgeWithCapability`,
   `Graph.RemoveHyperedgeWithCapability`. Existing methods unchanged.
3. Update `doc/integration/bma.md` with the capability usage sketch.
4. (Walk-phase, separate PR) BMA wires Skuld → Wyrd capability flow at
   sleep-cycle entry / observer goroutine spawn.

This PR ships steps 1–3 only. Step 4 happens in BMA-side code; no
Wyrd churn.

## 8. Soundness anchor

`Wyrd.Capability.capability_grants_safe_access` (Phase 1 T2.3) is the
formal theorem the runtime check implements: a holder at tier T
performing operations at tier T' ≤ T is safe. The Lean theorem proves
this for any ring; the Go check is the runtime enforcement.

`Wyrd.Foundations.no_surjection_complex_to_quaternion` (T2.1.a) and
the ℍ→𝕆, 𝕆→𝕊 equivalents prove the contrapositive: no inner-tier
process can synthesise outer-tier values. The capability check rejects
exactly those attempts.

## 9. What this PR does NOT do

- No Skuld minting API (BMA side; out of scope)
- No capability-gated reads (Option A chosen unless pushback)
- No per-NodeType permission granularity (Run-phase optimisation if
  profiling says otherwise — qbp-architecture seq=7)
- No signed capabilities / authority chain (Walk-phase)
- No capability expiry / revocation (Walk-phase; the `GrantedAt`
  field is metadata in v0.1, not enforced)

## 10. Open questions for review

1. **Read policy** — Option A (unrestricted reads) or Option B
   (capability-gated reads)? **My recommendation: A.** Pushback welcome.

2. **Default-tier behaviour for the bare `AddNode` / `AddHyperedge`** —
   keep them working without capability (current proposal: yes,
   default to TierComplex), or deprecate in favour of the explicit
   form? **My recommendation: keep working.**

3. **Package placement** — `model/capability.go` (lean) vs
   `compute/capability.go` (alternative)?

---

*Status: DRAFT v0.1 — open for review on qbp-cu-walk*
