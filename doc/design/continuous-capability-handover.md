# Continuous-Capability Handover: Durability, MVCC Staging/CAS, and the Write-Fence

**Status:** Theory addendum — DRAFT, open for §I4 review
**Origin:** beekeeper GO, `live-test` seq=5539/5540 (2026-11-19), following ratification of inter#301 v1.7 + v1.7.1 (beekeeper go recorded on inter#301, issuecomment-7022163369)
**Ratified record covered:** `JamesPagetButler/inter#301`, v1.6.3 through v1.7.1 (issuecomment-7017850231 and its round 11–16 dispositions; ratification issuecomment-7022163369)
**Companion issue:** `JamesPagetButler/wyrd#108` (Skuld, the Run-phase hardware-boundary reference monitor — reading A, no new execution model)
**Authors:** wyrd-implementor, with design-surface input adopted from `qbp-architecture` (fence-timeout scoping, `live-test` seq 5533–5536)

---

## 0. What this addendum is, and isn't

This is a **design surface**, not an implementation PR. Per the §I4 convention established in `doc/design/capability-enforcement.md` and `doc/design/bridge-batch.md`, nothing here authorizes code to land — it names the primitives the ratified scheduler design (inter#301) requires of Wyrd, grounds them against what exists today, and records open questions for the implementation PR(s) that follow.

**Scope.** Three S1 prerequisites inter#301 v1.7/v1.7.1 name for Wyrd, all serving one mechanism — a continuous capability's state surviving the handover from one finite lease to its successor (v1.7 §C2, §C5; r3/r4 round dispositions R11-B2, R12-B2, R13-M2):

1. **Durability** — Wyrd state survives a crash.
2. **MVCC staging + CAS-promote** — a replacement can build up a candidate state next to the live one, and switch to it atomically or not at all.
3. **The write-fence** — the brief window in which writes pause so the replacement's tail catch-up and the promotion decision see a consistent cutoff.

These were named as a *single* prerequisite in the original v1.7 draft and split into independent ones after review (`live-test` seq 5533–5536): durability is "can Wyrd not lose data," staging/CAS is "can Wyrd hold two candidate states and switch between them," and neither implies the other.

**Not in scope here:** the scheduler's own refresh/lane logic (Nidavellir's), the `cap(boxrun, ...)` contract shape (edda's), and the pod-layout question (`inter#303`, separate design issue).

---

## 1. What exists today (checked against `origin/main`)

`model/graph.go`'s `Graph` is a single flat structure:

```go
type Graph struct {
	mu        sync.RWMutex
	nodes     map[NodeID]Node
	edges     map[HyperedgeID]Hyperedge
	incidence map[NodeID]map[HyperedgeID]struct{}
	retentionCaps map[RetentionTier]int
}
```

Load-bearing facts for this addendum:

- **No durability.** `Graph` is purely in-memory. There is no WAL, no snapshot-to-disk, no replication. A process restart loses all state. (Crawl-phase storage, per the README, is JSON files written by the separate Crawl pipeline — not a `Graph`-level durability story.)
- **No versioning.** There is no notion of a generation number, content hash of the whole graph, or "head." Individual `Hyperedge`/`Node` values are addressed by ID, not by a graph-wide version.
- **No branching.** `PromoteBatch` (`model/graph_batch.go`) is the nearest existing multi-step-atomic primitive: phase-1 preflight validates a batch of hyperedge IDs against a destination graph, phase-2 applies them, one lock window, lock-ordered by pointer address to avoid deadlock when two `PromoteBatch` calls race on the same pair of graphs. But it promotes *into* an existing graph — there is no concept of two graphs where one is "staging" and the other "committed," nor of atomically swapping which one callers see.
- **Purely reactive.** Every exported `Graph`/`capability.go` method is invoked by a caller and returns; nothing in `model/` runs a background goroutine, a timer, or any self-initiated state transition. The single `sync.RWMutex` is held only for the duration of a call.

This grounds why durability and staging/CAS are independent, nontrivially-sized primitives rather than small extensions: durability is an orthogonal persistence layer underneath the existing `Graph`; staging/CAS is a new structural concept (two coexisting graph states, one visible, one not, with an atomic swap) that `PromoteBatch` does not provide.

---

## 2. Primitive 1 — Durability

**Requirement, from v1.7 §C2:** "A lane is disposable only if its state lives in Wyrd... The shutdown contract is mandatory: drain, then persist the working state to Wyrd, then release. The replacement restores from Wyrd before it counts as ready." Tests 8 (R11 round), 12 and 13 (R12-B2: a write during startup survives the handover; a write past the fence aborts the promotion) all presuppose that a write Wyrd acknowledges as committed is still there after a crash or a process boundary.

**What this is not:** a request for Wyrd to interpret *what* state means to BMA, Nidavellir, or any other consumer. Wyrd's job is the same as everywhere else in this codebase — structural guarantees (a committed write is retained; a read after a crash sees the last committed state) — never semantic ones (whether the restored state is a *correct* checkpoint is the consumer's claim, per the wyrd#94 "Wyrd never re-checks" precedent).

**Open questions for the implementation surface:**
- Write-ahead log vs. snapshot-plus-log vs. an external replicated store — a separate design decision, likely its own PR, since it interacts with the flip-gate item (S3, authority moving to Wyrd) already named in the scheduler's progressive-bootstrap design.
- Whether durability is graph-wide or scoped to specific NodeType/Hyperedge namespaces (e.g., only `fed.grant.*`/`fed.lease.*` once those land at S2) — the continuous-capability state a lane persists is presumably scoped, not the whole graph, and scoping durability narrows the engineering problem considerably.

---

## 3. Primitive 2 — MVCC staging and CAS-promote

**Requirement, from v1.7 r2/r3 (R11-B2, R12-B2) and r4 (R13-M2's reference to "the staging head"):**

- "State promotion is transactional. The replacement writes only to a **staging branch** in Wyrd. A schema migration happens on that branch too."
- "The committed state path changes in one atomic step, on the replacement's ready signal... Each state record carries a schema version. A lane refuses to restore a version it doesn't support, failing closed instead of misreading."
- "Promotion is a **compare-and-set** on the committed head, from o_f to the staging head; if the head moved, it aborts and the outgoing lane resumes."

**Shape this implies for Wyrd:**

1. A **staging area** per continuous capability (or per lane-chain), isolated from readers of the committed state, where a replacement lane can write its restored-and-caught-up state without being visible until promoted. This is a copy-on-write/MVCC concept new to `Graph` — not a variant of `PromoteBatch`, which has no isolation step; its phase-1/phase-2 split is preflight-then-apply within a single visible graph, not visible-vs-staging.
2. A **monotonic marker on the committed state** — the "head" the CAS compares against. Candidates: a counter incremented on every promotion, or a content hash of the committed state at the moment of handover (the design record's own language, "matched by Wyrd record hash," in test 8, suggests the latter is already assumed by the scheduler side). Whichever is chosen, it must be read-and-compared atomically with the swap — the CAS's entire point is that a second writer cannot move the head between the replacement's read of it and its promote call.
3. **Schema-version refusal.** Each staged state record names a schema version; promotion (or restore from staging by a standby) that targets a version the reading lane doesn't support is refused, never silently misread. This is a structural check Wyrd can own outright — it is exactly the reference-monitor shape (check a tag against a capability, refuse or permit) the capability layer already implements in `model/capability.go`, just checking a schema tag instead of a tier.

**Division of labor, explicit:** Wyrd verifies that the CAS's precondition held (the head hadn't moved) and that the schema version is supported; Wyrd does **not** verify that the staged content is semantically a correct migration, a correct catch-up replay, or anything else about *what* the bytes mean. That stays the scheduler/consumer's claim, exactly as wyrd#94 established for the Edda-checker-result case.

**Open questions:**
- Whether staging lives as a second `Graph` instance, a tagged overlay on one `Graph`, or something else — an implementation choice, not a theory one, but it determines whether the existing `sync.RWMutex` boundary (already load-bearing for ADR-003 §I3, per `doc/design/capability-enforcement.md` §0) needs to change shape or can stay as-is with staging as an orthogonal namespace.
- Who mints the schema-version tag and where it's validated against a registry (if one is needed) — likely answered once this interacts with the `fed.grant.*`/`fed.lease.*` Node-type family planned for S2.

---

## 4. Primitive 3 — The write-fence

**Requirement, from v1.7 r3 (R12-B2) and r4 (R13-M2), as adopted after review:**

- "A short Wyrd write-fence follows: writes pause, reads continue, and the replacement applies the tail to o_f."
- "The fence is a **Wyrd-held lease** with a hard, non-extendable timeout = k_f × the measured p95 tail-apply time (derived, PROVISIONAL). On timeout, [the fence lapses and] the promotion [is] abort[ed], and AMBER fires."

**Adopted scoping (seq 5536):** the timeout is enforced **lazily**, not by a Wyrd-internal timer. Wyrd stores the fence's deadline as a field at mint time; it does not run a goroutine to proactively tear the fence down. Any call that would act on an expired fence — a tail-apply write, or the CAS-promote itself — is simply refused once the deadline has passed, as if the fence had already lapsed. This keeps Wyrd's first fence primitive consistent with every other primitive in `model/`: checked on a call, never self-initiated. The scheduler side (Nidavellir) still raises AMBER and resumes the outgoing lane on its own clock, per R12-B2 — it does not need Wyrd to tell it the fence expired; it already knows its own deadline.

**Shape this implies for Wyrd:**

1. A fence is a lease — mint it with an owner, a deadline, and the scope it covers (the capability/lane-chain whose committed state it's fencing).
2. **While a live (unexpired) fence is held, the tail-apply writes go to the staging state, never to committed.** The fence holder (the replacement) may write *staging* only. The *committed* state is writable by nobody under a live fence — the holder included — except through the single CAS-promote call that ends the fence. "Applies the tail" (v1.7 r3) means the replacement's catch-up writes land in its own staging area, to be made visible only by the eventual promote; it is not a carve-out letting the holder write committed directly.
   - **Test:** under a live fence, the holder's staging write succeeds and its committed write is refused.
   - **Mutant:** the holder is exempted from the committed-state refusal. KILLED.
3. **The fence scope is per-capability** — pinned by `qbp-architecture` ruling (`live-test` seq 5552), since the record doesn't say it explicitly. A fence covers exactly the committed-state scope of the capability (lane-chain) it names; writes outside that scope are unaffected. C2 and C5 describe one capability's handover, and a graph-wide fence would let one capability's refresh stall every other writer — the same starvation class R13 closed for admission.
   - **Test:** a write to a different capability's scope succeeds during the fence.
   - **Mutant:** the fence blocks graph-wide. KILLED.
4. **A refusal under a live fence is a distinct, retryable error — `ErrFenced` — never the generic capability-refusal error.** The record says writes *pause*; this addendum's §4.2 says refused, not queued. For those to be consistent in practice, the caller has to be able to tell a fence refusal apart from an authorization refusal: the outgoing lane (R12-B2) retries the write once the fence lapses or the CAS-promote aborts, and it can only do that correctly if it can distinguish "try again, the fence will clear" from "you were never allowed to write this."
   - **Test:** a fenced write returns `ErrFenced`; the same write succeeds once the fence lapses or the promotion aborts.
   - **Mutant:** a fenced write returns the generic capability-refused error. KILLED.
5. The CAS-promote call itself only succeeds while the fence it names is both held and unexpired, and succeeds at most once per fence (a second promote attempt against an already-consumed or lapsed fence is refused with `ErrFenced`).

---

## 5. Cross-references

- `model/graph.go`, `model/graph_batch.go`, `model/capability.go` — existing primitives this addendum builds on or contrasts with.
- `doc/design/capability-enforcement.md` — the existing design-surface convention this addendum follows (status header, §I4 review gate, division-of-labor framing).
- `doc/design/bridge-batch.md` — nearest existing precedent for a multi-step atomic primitive (`PromoteBatch`).
- `JamesPagetButler/wyrd#94` — "Wyrd never re-checks" precedent for the division-of-labor principle applied throughout §2–4.
- `JamesPagetButler/wyrd#108` — Skuld design surface; this addendum's primitives are scheduler/state-handover concerns and are explicitly out of Skuld's scope (reading A).
- `JamesPagetButler/inter#301` — the ratified scheduler design (v1.6.3 RATIFIED issuecomment-6997840093; v1.7/v1.7.1 RATIFIED issuecomment-7022163369) this addendum is the Wyrd-side theory record for.
- `JamesPagetButler/inter#303` — pod-layout design issue; input requested from wyrd-implementor, not yet drafted.
