# Skuld-Spec v1.0 — Critical Review for Hammer Use Case

**Helpful Engineering — Quaternion-Based Physics Programme**
April 2026 — Review document
Reviewer: Claude Opus 4.7 (red-team / architecture role)

> **Method.** Read *Skuld-Spec-v1.0.md* with Hammer simulation as the primary user. Hammer needs to dispatch ~10⁶ particle-pair updates per second per lane (Wyrd-Workload-ISA-v0.1.md §2.1), with quaternion FMA dominating the inner loop. Every API call in the Skuld-Spec is evaluated for: (a) does it serve Hammer well, (b) does it serve Hammer at scale, (c) does it impose unnecessary overhead. Identified gaps drive Skuld-Spec v1.1 amendments.

> **Verdict.** Skuld-Spec v1.0 is correct in its 9-call surface but has **five concrete gaps** that would force Hammer to either (a) bypass Skuld and talk directly to qbpcu (defeating the privilege model) or (b) eat unacceptable per-op overhead. All five are addressable without expanding the surface materially.

---

## 1. The Hammer workload, summarized

A single Hammer simulation step:

1. For each particle pair (i, j) — typically 10² to 10³ pairs per particle × 10⁴ particles = 10⁶-10⁷ pairs per timestep:
   1.1. Read particle i, j state from Wyrd
   1.2. Compute force quaternion via QFMA chain (~5-10 QFMAs)
   1.3. Compute pair contribution via QSAND (1-2)
   1.4. Accumulate to total force
2. For each particle:
   2.1. Integrate force → new velocity (1-2 QFMA)
   2.2. Integrate velocity → new position (1-2 QFMA)
   2.3. Compute energy via QNORM (1)
   2.4. Update particle in Wyrd
3. Compute total energy, check conservation
4. Optionally: log seam events from the watchdog stream
5. Repeat for next timestep, target ~10³ Hz update rate per particle

Per timestep at 10⁴ particles: ~10⁷ QBP-CU operations, ~10⁵ Wyrd reads, ~10⁴ Wyrd writes, ~10⁴ stream events to inspect. **Total per second at 10³ Hz update: ~10¹⁰ QBP-CU ops, ~10⁸ Wyrd reads, ~10⁷ Wyrd writes, ~10⁷ stream events.**

Skuld must mediate all of this. Per-op overhead matters enormously.

---

## 2. The five concrete gaps

### Gap 1: No batch QBPSubmit

**The problem.** `QBPSubmit(pid, op, srcs...) → reqid` is a per-operation call. Hammer would need to call it ~10¹⁰ times per second. Each call is a Go function call into Skuld, which acquires the per-process mutex, performs the capability check, forwards to `qbpcu.Accelerator.Submit`. Best case in modern Go: ~30-50 ns per call. **Per-second budget: 30-50 seconds of pure overhead per second of simulation time.** Unworkable.

**The fix.** Add a batch variant:

```go
QBPSubmitBatch(pid ProcID, ops []Op) ([]ReqID, error)

type Op struct {
    Opcode qbpcu.Opcode
    Srcs   []Operand
}
```

Hammer batches ~10³-10⁴ ops per call. Per-second QBPSubmit calls drop from 10¹⁰ to 10⁶-10⁷. Per-call overhead amortized by batch size; capability check is once per batch (the entire batch runs in the same ring or it fails).

A second batch variant for the common QFMA-chain pattern:

```go
QBPFMAChain(pid ProcID, accumulator Operand, terms [](Operand, Operand)) (ReqID, error)
```

Submits a chain `accumulator + Σ aᵢ·bᵢ` as a single dependent batch. Skuld serializes via `qbpcu.Accelerator`'s single-op interface but the batch traverses the API only once.

### Gap 2: Polling model creates roundtrip latency

**The problem.** `QBPSubmit` returns immediately, caller calls `QBPPoll(reqid)` which blocks. For tight-loop workloads this is fine — submit a batch, poll, repeat. But the spec doesn't define a *streaming* completion model, which Hammer's pipelined inner loop wants. With pipeline depth ~10, Hammer is filling the QBP-CU's queue and consuming results in FIFO order.

**The fix.** Expose a result stream:

```go
QBPSubmitStream(pid ProcID) (Stream[Resp], chan<- Op, error)
```

Returns two channels: a result channel (receive-only) and an op channel (send-only). Hammer pumps ops into the send channel, reads results from the receive channel. Skuld manages the queue depth, capability checks on each op (or once-per-stream if Hammer guarantees ring stability), and back-pressures Hammer if the QBP-CU is saturated.

This is strictly more powerful than the QBPSubmitBatch pattern from Gap 1; QBPSubmitBatch can be implemented in terms of QBPSubmitStream. Both surface APIs exist for ergonomics — batch for "I have N ops ready," stream for "I'll feed you ops continuously."

### Gap 3: Wyrd query latency on hot read paths

**The problem.** Hammer reads particle state from Wyrd ~10⁸ times per second. `WyrdQuery(pid, pattern, cap)` is a one-shot call returning a `Result`. Each call has the same per-call overhead as QBPSubmit (~30-50 ns) plus the actual Wyrd query cost. **Per-second budget: another 5+ seconds of overhead.**

**The fix.** Add prepared queries:

```go
WyrdPrepare(pid ProcID, template QueryTemplate, cap Capability) (PreparedQuery, error)
PreparedQuery.Execute(args ...Operand) (Result, error)
PreparedQuery.Close() error
```

A `PreparedQuery` is checked for capability once at prepare time; subsequent `Execute` calls skip the capability check (the template's ring is fixed at prepare). Hammer prepares a "read particle by id" query at startup, then calls `Execute(particle_id)` ~10⁸ times per second. Per-call overhead drops to ~10 ns (just the capability-set lookup is gone).

This is a standard pattern from databases (PostgreSQL prepared statements, ORM compiled queries). Worth importing.

### Gap 4: Subscription ergonomics for watchdog seam events

**The problem.** `WyrdSubscribe(pid, pattern, cap) → Stream` is a long-lived subscription. The Stream's `Next(ctx) (Event, error)` is per-event blocking. For the rate of seam events Hammer cares about (~10⁷ per second), this works — the cost is in the channel send/receive on the Go side, which is ~50 ns. But there's no way to *batch* event consumption.

**The fix.** Add a batched read:

```go
Stream.NextBatch(ctx, max int) ([]Event, error)
```

Returns up to `max` events that are currently available, blocking only if zero are available. Hammer drains the stream every ~1 ms, processing batches of ~10³-10⁴ events. Per-second overhead drops accordingly.

### Gap 5: Determinism for replay/debug

**The problem.** Skuld's WDEvent consumer goroutine processes events *in order of receipt* per §7.3. But Hammer's pipelined ops complete in nondeterministic order if multiple lanes are used. For research workloads, replayability matters — given the same input, two runs should produce the same output, including the same WDEvent stream order. The current spec doesn't promise this for multi-lane operation.

**The fix.** Two additions:

```go
// Session-level determinism
type DeterministicConfig struct {
    Seed       int64
    LaneOrder  LaneOrderingPolicy  // FIFOPerLane | RoundRobin | DeterministicScheduled
}
ProcCreate(elf, name, caps, det *DeterministicConfig) → ProcID

// Per-process replay
ProcReplay(pid ProcID, fromCycle uint64, toCycle uint64) error  // re-issues ops from log
```

Without these, debugging a Hammer-detected anomaly requires re-running the full simulation and hoping the bug reproduces. With them, a specific cycle-range can be replayed deterministically with the same WDEvent emission order.

This is a Crawl-acceptable complexity to add. It avoids a trap where Walk-phase debugging requires retroactive determinism support.

---

## 3. Things Skuld-Spec gets right

I want to flag the design decisions that pay off well for Hammer:

**`Resp.InvariantSnapshot` is already there.** Excellent. Hammer's energy-conservation check after each timestep is `resp.InvariantSnapshot.AlgebraID + Norm`. No separate `QBPQueryInvariant` call needed. Saves ~10⁶ syscalls per second.

**`ProcGrantCapability` is one-shot, not per-op.** Right call. Hammer gets its ℍ-capability once at startup and runs at AlgebraID=1 for the duration. Capability-check cost is "is RingH in CapabilitySet" which is a single map lookup (~10 ns).

**WyrdSubscribe streams events asynchronously.** Right model for the watchdog seam events Hammer wants to monitor. A polling-based event mechanism would have been worse.

**No CPU scheduling.** Skuld doesn't try to be a kernel scheduler — Linux handles that. This means Hammer can run as a normal Linux thread, get pinned via `runtime.LockOSThread`, and have the OS scheduler do the right thing. Skuld stays out of the way of CPU performance.

**The 9-call surface is genuinely small.** Every call is justifiable. No "and we should add X for completeness" syndrome. Hammer uses 5 of the 9 calls (ProcCreate, ProcGrantCapability, QBPSubmit, QBPPoll, WyrdSubscribe), uses none of the supervisor-internals, and has no need for the remaining 4. This is the right scope.

---

## 4. Skuld-Spec v1.1 amendments — the proposal

Three new types:

```go
// Op is a single queued QBP-CU operation, used in batch and stream APIs.
type Op struct {
    Opcode qbpcu.Opcode
    Srcs   []Operand
}

// PreparedQuery is a Wyrd query with capability pre-checked.
type PreparedQuery interface {
    Execute(args ...Operand) (Result, error)
    Close() error
}

// QueryTemplate is a parameterized query pattern.
type QueryTemplate struct {
    Pattern wyrd.Pattern
    Params  []ParamSpec
}

// DeterministicConfig configures replayability for a process.
type DeterministicConfig struct {
    Seed      int64
    LaneOrder LaneOrderingPolicy
}

type LaneOrderingPolicy uint8
const (
    LaneOrderFIFOPerLane LaneOrderingPolicy = iota
    LaneOrderRoundRobin
    LaneOrderDeterministicScheduled
)
```

Four new methods (added to the `Supervisor` interface):

```go
QBPSubmitBatch(pid ProcID, ops []Op) ([]ReqID, error)
QBPSubmitStream(pid ProcID) (<-chan Resp, chan<- Op, error)
WyrdPrepare(pid ProcID, template QueryTemplate, cap Capability) (PreparedQuery, error)
ProcReplay(pid ProcID, fromCycle, toCycle uint64) error
```

One method extended (the existing `ProcCreate` gains an optional config):

```go
// Old: ProcCreate(elf []byte, name string, initialCaps []RingID) (ProcID, error)
// New: ProcCreate(elf []byte, name string, initialCaps []RingID, det *DeterministicConfig) (ProcID, error)
```

The `det` parameter accepts `nil` for backward compatibility — non-deterministic behavior is the default.

One method extended on Stream (adds a batched form):

```go
// In Stream interface:
NextBatch(ctx context.Context, max int) ([]Event, error)
```

**Net change to the user-visible API surface:** 9 → 13 methods. Still small; still under any conceivable kernel's syscall surface.

---

## 5. Performance projection with v1.1 amendments

Recomputing Hammer's per-second overhead budget:

| Operation pattern | v1.0 cost | v1.1 cost | Improvement |
|---|---|---|---|
| 10¹⁰ QBP-CU ops via single submission | ~30-50 sec | via batch (~30 ns / batch * 10⁷ batches) ≈ 0.3 sec | ~100× |
| 10⁸ Wyrd reads via WyrdQuery | ~3-5 sec | via PreparedQuery (~10 ns / call) ≈ 1 sec | ~3-5× |
| 10⁷ stream events via Next | ~0.5 sec | via NextBatch ≈ 0.05 sec | ~10× |

Aggregate Hammer overhead in a 1-second simulation: ~33-55 seconds → ~1.4 seconds. Still significant but workable.

For an X280-class core at 2 GHz this is the right rough-cut. The actual bottleneck shifts from Skuld API overhead to QBP-CU throughput — which is where it should be.

The v1.1 amendments are not "premature optimization." They are "the difference between Hammer being practical and Hammer being a benchmark exercise." The v1.0 spec is too conservative for the actual workloads it's meant to serve.

---

## 6. What still doesn't work even after v1.1

Some workload patterns I haven't covered, each potentially needing further attention:

**Direct QBP-CU access for tight loops.** Even with batches, the per-batch Skuld overhead is ~30 ns. For absolutely-hottest inner loops, Hammer might want to bypass Skuld and call `qbpcu.Accelerator` directly. The spec should explicitly forbid this for capability-bearing processes (security violation) and explicitly allow it for capability-checked-up-front contexts. Suggested addition: `Process.GetCheckedAccelerator(token) (qbpcu.Accelerator, error)` returning a ring-restricted accelerator handle that Hammer can call directly, bypassing the Skuld API but still capability-gated.

**Cross-process communication.** Hammer-with-coupled-particles might want to communicate between processes. Currently no mechanism. Either NATS (overkill at Crawl) or shared Wyrd queries (the canonical answer). The Wyrd-as-IPC pattern needs documentation but no API change.

**Latency-sensitive control loops.** A research workflow that wants <100 μs response time from QBP-CU completion to next dispatch may not be served by the goroutine-based API at all — Go's scheduler can introduce 100+ μs latencies under load. The right answer is probably a "real-time mode" for Skuld that uses busy-polling instead of channel-based blocking. Defer to Walk; no Crawl-blocker.

---

## 7. Recommendation

**Adopt the v1.1 amendments as a Skuld-Spec v1.1 update.** They are:

- Small (4 new methods, 1 method extension, 1 stream-method extension)
- Well-motivated (real workload analysis, not speculation)
- Backward-compatible (nil-defaulted optional config)
- Testable (each amendment has a clear performance target)

**Validate against second workload.** This review used Hammer. Repeat with a different primary workload — recommend Squam Lake thermocline (very different access pattern, field-evolution rather than particle-pair). If Squam Lake reveals additional gaps, fold those into v1.1 before locking.

**Avoid bypass mechanisms in v1.1.** The "Process.GetCheckedAccelerator" suggestion in §6 is a Walk-phase question. At Crawl, keep all access through Skuld; if performance is unacceptable, learn that empirically before adding bypass.

**Update C-06 (the BMA-instantiation ticket for the skuld package) to track v1.1 as the target.** This adds maybe a week of implementation work but saves a Skuld-Spec v2 redesign cycle later.

---

## 8. Actionable updates

Three concrete artifacts to produce:

1. **Skuld-Spec v1.1** — full updated spec. ~1 day to draft.
2. **Hammer Use Case Reference** — a worked example showing Hammer's full inner loop using v1.1 APIs. Becomes test case for C-09 integration ticket. ~½ day.
3. **Squam Lake Use Case Reference** — same exercise with field-evolution workload, for v1.1 validation. ~½ day.

Order: 1, then 3 (validate with second workload), then 2 (the canonical example).

---

## 9. Acknowledgments to current Skuld-Spec

The v1.0 design got the privilege model, capability mechanism, and API shape correct. The gaps identified are about *throughput*, not *correctness*. This is a much better failure mode than the alternatives. The amendments are extending a sound design, not replacing a flawed one.

---

*End of Skuld-Spec v1.0 Critical Review for Hammer Use Case — DRAFT*
