# Skuld

## A QBP CU Supervisor — Crawl Phase API Specification

**Helpful Engineering — Quaternion-Based Physics Programme**
Principal Investigator: James Paget Butler
April 2026 — Rev 1.0 — DRAFT

> **Definition.** Skuld is the QBP-CU supervisor — the policy and mediation layer between user processes and the QBP-CU accelerator hardware, and between user processes and Wyrd. Named after the Norn of *what shall be*, Skuld enforces what each process is owed access to, mediates capability-bearing transitions across algebraic privilege rings, and consumes the watchdog event stream that makes privilege violations structurally detectable.

> **Scope.** This specification is for the **Crawl phase** of Skuld. The Crawl supervisor is a thin Go layer running on the host CPU, reading `WDEvent`s from the QBP-CU watchdog via the host-pinned ring buffer, mediating syscalls from user processes, and exposing the 9-call user-visible API plus 2 supervisor-internal calls. The Walk/Run/Sprint phase progressions evaporate this layer into Wyrd queries; this spec does not describe those phases beyond noting the migration paths.

> **Attribution.** Per QBP standing rule. Building on prior art from DBOS (Stonebraker et al.), seL4 (Klein et al.), CHERI (Watson et al.), TabulaROSA (Kepner et al.), SiFive (Asanovic et al., for VCIX and X280 reference). QBP foundations: Furey, Dixon, Günaydin/Gürsey, Boyle/Farnsworth, Singh, Chamseddine/Connes, Koide, Baez.

---

## 1. Overview

### 1.1 What Skuld is

A Go package, `skuld`, providing process management, QBP-CU mediation, capability tracking, and Wyrd query routing for QBP research workloads. Runs on the host RV64GC core; consumes the QBP-CU watchdog's WDEvent stream; sits between user processes and both the accelerator and the database.

### 1.2 What Skuld is not

- **Not a kernel.** Skuld manages QBP-research processes; it does not handle Linux kernel responsibilities (memory paging, file system, network stack). The host Linux kernel handles those; Skuld is a user-space service layered on top.
- **Not a scheduler in the conventional sense.** Skuld dispatches work to the QBP-CU and tracks process state, but does not preempt CPU between user processes — each Skuld-managed process runs as a Linux thread, scheduled by Linux. Skuld decides who *gets the QBP-CU* next, not who gets the CPU.
- **Not a complete OS abstraction.** A Skuld-managed process can still make Linux syscalls. Skuld mediates only QBP-CU and Wyrd access.

### 1.3 The minimal mental model

A research workload (Hammer simulation, BMA inference, GRB analysis) runs as one or more `Process`es registered with Skuld. Each `Process` has a `CapabilitySet` (which algebraic rings it may operate in) and a connection to Wyrd. To run physics, the process calls `QBPSubmit`; Skuld checks the capability against the current ring and either authorizes (writing the ring into `qbp_ctl.ALGEBRA_ID` and forwarding to the accelerator) or faults. The accelerator processes the op; the watchdog emits a `WDEvent`; Skuld consumes the event, applies privilege policy, and forwards physical-seam events to Wyrd.

---

## 2. Core types

```go
package skuld

import (
    "context"
    "errors"
    "io"
    "time"

    "github.com/JamesPagetButler/QBP/qbpcu"
    "github.com/JamesPagetButler/QBP/wyrd"
)

// ProcID uniquely identifies a Skuld-managed process across the Skuld instance's lifetime.
// IDs are not reused after process destruction.
type ProcID uint64

// RingID names a Wyrd algebraic privilege ring.
// The values match qbp_ctl.ALGEBRA_ID encoding.
type RingID uint8

const (
    RingC RingID = 0 // ℂ — user (default)
    RingH RingID = 1 // ℍ — supervisor (Skuld itself)
    RingO RingID = 2 // 𝕆 — kernel (QBP-CU operations)
    RingS RingID = 3 // 𝕊 — firmware (boot only)
    // 4, 5 reserved for Branch A (ℂ⊕ℍ⊕M₃(ℂ)) and Branch B physics modes
)

// CapabilitySet records which rings a process may operate in.
// A user process holds {RingC} by default; a process granted an ℍ-capability
// holds {RingC, RingH}; etc. Capabilities are inclusive — a process holding
// RingO implicitly holds RingC and RingH (per T2.2 projection lemma).
type CapabilitySet struct {
    rings    map[RingID]bool
    tokens   map[RingID][]byte // capability tokens per ring (Walk-phase use)
}

// Process is the unit of Skuld management.
type Process struct {
    ID            ProcID
    Name          string
    Capabilities  CapabilitySet
    CurrentRing   RingID
    CreatedAt     time.Time
    State         ProcessState
    // Internal — not part of public API
    wyrdConn      *wyrd.Connection
    pendingReqs   map[ReqID]chan Resp
}

type ProcessState uint8

const (
    StateInit ProcessState = iota
    StateRunning
    StateBlocked
    StateFaulted
    StateTerminated
)

// ReqID identifies an outstanding QBP-CU operation.
type ReqID uint64

// Operand is a typed wrapper over QW128 or vector-of-QW128 for VCIX ops.
// The actual encoding is defined in the qbpcu package; Skuld passes them
// through without inspection (Skuld is not a quaternion arithmetic engine).
type Operand = qbpcu.Operand

// Resp is the result of a completed QBP-CU operation.
// Status indicates OK, watchdog fault, or decode error.
type Resp struct {
    Result    qbpcu.QW128
    Status    qbpcu.Status
    FaultCode uint32
    InvariantSnapshot Invariant // populated on every op; reads qbp_invariant CSR
}

// Invariant is the qbp_invariant CSR snapshot at the moment of completion.
type Invariant struct {
    Commutator [4]int8
    Associator [3]int8
    Alternator [4]int8
    RingID     RingID
    Cycle      uint64
}

// QueryPattern is a Wyrd query — actual structure defined in the wyrd package.
type QueryPattern = wyrd.Pattern
type Result = wyrd.Result
type Stream = wyrd.Subscription

// Capability is a token authorizing a wider-ring operation. At Crawl phase,
// these are bytes that Skuld interprets via its capability table; at Walk
// phase they become Wyrd hypergraph nodes.
type Capability []byte
```

---

## 3. Errors

Skuld defines a flat error set. All errors implement Go's standard `error` interface and can be tested with `errors.Is`.

```go
var (
    // Process-management errors
    ErrProcNotFound       = errors.New("skuld: process not found")
    ErrProcAlreadyExists  = errors.New("skuld: process with that name already registered")
    ErrInvalidELF         = errors.New("skuld: invalid ELF binary")
    ErrProcLimitReached   = errors.New("skuld: maximum process count reached")

    // Capability errors
    ErrInsufficientCapability = errors.New("skuld: process lacks capability for requested ring")
    ErrInvalidCapability      = errors.New("skuld: malformed or expired capability token")
    ErrRingNotPermitted       = errors.New("skuld: ring not in process capability set")

    // QBP-CU errors
    ErrQBPBusy            = errors.New("skuld: QBP-CU request queue full")
    ErrQBPInvalidOp       = errors.New("skuld: unrecognized QBP-CU opcode")
    ErrQBPDecodeError     = errors.New("skuld: QBP-CU decode error")
    ErrWatchdogFault      = errors.New("skuld: QBP-CU watchdog raised fault")
    ErrInvalidReqID       = errors.New("skuld: request ID not recognized")

    // Wyrd errors  
    ErrWyrdQueryRefused   = errors.New("skuld: Wyrd query refused due to capability check")
    ErrWyrdUnavailable    = errors.New("skuld: Wyrd database unreachable")

    // Lifecycle errors
    ErrSkuldShutdown      = errors.New("skuld: supervisor is shutting down")
)
```

Watchdog faults carry a `FaultCode` field providing additional detail; Skuld defines the fault code constants in `skuld/faults.go` (mirroring `qbpcu/faults.go`).

---

## 4. The Skuld interface

The supervisor exposes a single interface, `Supervisor`, that user code obtains via `New()`. The interface contains the 9 user-visible methods plus 2 supervisor-internal methods (used by the WDEvent consumer goroutine).

```go
// Supervisor is the Skuld API as exposed to user processes and internal goroutines.
type Supervisor interface {
    // 4.1 Process management (4 methods)
    ProcCreate(elf []byte, name string, initialCaps []RingID) (ProcID, error)
    ProcDestroy(pid ProcID) error
    ProcSetAlgebra(pid ProcID, ring RingID) error
    ProcGrantCapability(pid ProcID, ring RingID, token Capability) error

    // 4.2 QBP-CU mediation (3 methods)
    QBPSubmit(pid ProcID, op qbpcu.Opcode, srcs ...Operand) (ReqID, error)
    QBPPoll(reqid ReqID) (Resp, error)
    QBPQueryInvariant(pid ProcID) (Invariant, error)

    // 4.3 Wyrd query (2 methods)
    WyrdQuery(pid ProcID, pattern QueryPattern, cap Capability) (Result, error)
    WyrdSubscribe(pid ProcID, pattern QueryPattern, cap Capability) (Stream, error)

    // 4.4 Supervisor-internal (2 methods, used by Skuld's own goroutines
    //     — not part of the user-facing surface but exposed for testing
    //     and for the WDEvent consumer)
    wdEventNext(ctx context.Context) (qbpcu.WDEvent, error)
    wdPolicyApply(event qbpcu.WDEvent) (Action, error)
}
```

The supervisor-internal methods are accessible via `SupervisorInternal` (a separate interface that embeds `Supervisor`); production user code only sees `Supervisor`.

---

## 5. API specification

### 5.1 ProcCreate

```go
ProcCreate(elf []byte, name string, initialCaps []RingID) (ProcID, error)
```

Creates a new Skuld-managed process from an ELF binary.

**Parameters:**
- `elf` — RV64GC ELF binary; Skuld validates the magic bytes and architecture but does not parse symbols. Maximum size 64 MiB at Crawl.
- `name` — human-readable identifier; must be unique across living processes. Maximum 256 bytes UTF-8.
- `initialCaps` — list of rings the process is born with. Must include `RingC`. Skuld returns `ErrInsufficientCapability` if the caller (currently always the human-operator at Crawl phase) is not authorized to grant the requested rings.

**Returns:**
- A fresh `ProcID` on success.
- `ErrInvalidELF` if validation fails.
- `ErrProcAlreadyExists` if `name` is in use.
- `ErrProcLimitReached` if Skuld has reached its process cap (default 256 at Crawl).

**Side effects:**
- Allocates a Linux thread for the process (via Go's `runtime.LockOSThread`).
- Initializes the process's Wyrd connection.
- Registers the process in Skuld's process table.
- Does NOT invoke the ELF's entry point — the caller subsequently calls `ProcSetAlgebra` to set the initial ring and then triggers execution via the process's communication channel.

**Crawl-phase implementation note.** The "ELF binary" abstraction is preserved for compatibility with the eventual VexRiscv FPGA implementation (Phase 3). At Crawl, the ELF is loaded by Go's runtime as a normal Go binary; the QBP-CU custom-0 instructions are intercepted by the `qbpcu.Mock` or `qbpcu.Golden` accelerator. The "ELF" framing is correct; the *execution model* is Go-native.

### 5.2 ProcDestroy

```go
ProcDestroy(pid ProcID) error
```

Terminates a process, cancels all its outstanding QBP-CU and Wyrd operations, and removes it from Skuld's process table.

**Parameters:**
- `pid` — the process to destroy.

**Returns:**
- `nil` on success.
- `ErrProcNotFound` if the process does not exist or has already been destroyed.

**Side effects:**
- Cancels all `pendingReqs` for this process; their channels receive a `Resp{Status: qbpcu.StatusCancelled}`.
- Closes the Wyrd subscription streams owned by this process.
- Releases the Linux thread.
- The `ProcID` is not reused.

**Idempotency.** Calling `ProcDestroy` on an already-destroyed process returns `ErrProcNotFound` rather than silently succeeding; this prevents masking double-free bugs in user code.

### 5.3 ProcSetAlgebra

```go
ProcSetAlgebra(pid ProcID, ring RingID) error
```

Changes the process's current ring. Writes to `qbp_ctl.ALGEBRA_ID` if the process is currently dispatched to the QBP-CU.

**Parameters:**
- `pid` — the target process.
- `ring` — the ring to enter. Must be in the process's `CapabilitySet`.

**Returns:**
- `nil` on success.
- `ErrProcNotFound`, `ErrRingNotPermitted` as appropriate.

**Side effects:**
- Updates `Process.CurrentRing` in Skuld's process table.
- If the process has a QBP-CU operation in flight, the change takes effect on the next op; in-flight ops complete in their original ring.
- WDEvent stream may emit a `RingTransition` event reflecting the change.

**Skuld-internal use.** This method is called by `wdPolicyApply` when authorizing a privilege escalation triggered by a watchdog event. User code can also call it directly to descend into a lower ring.

### 5.4 ProcGrantCapability

```go
ProcGrantCapability(pid ProcID, ring RingID, token Capability) error
```

Adds a ring to the process's `CapabilitySet`, authorized by `token`. Capability tokens are opaque bytes at Crawl (Skuld validates against its capability table); at Walk, they become Wyrd hypergraph nodes.

**Parameters:**
- `pid` — the target process.
- `ring` — the ring to grant.
- `token` — Skuld-defined capability token. At Crawl, this is a HMAC-signed blob derived from a Skuld master key plus the process's identity.

**Returns:**
- `nil` on success.
- `ErrInvalidCapability` if the token does not validate.

**Side effects:**
- Updates `Process.Capabilities` in Skuld's process table.
- Stores the token in `Process.Capabilities.tokens[ring]` for future audit.

**Crawl-phase capability tokens.** At Crawl, capability minting is a privileged operation performed by the human operator via a separate command-line tool (`skuldctl mint --pid X --ring H`). The tool requires access to the Skuld master key, stored encrypted in a file readable only by the Skuld process owner. This is intentionally crude; Walk phase replaces it with Wyrd-native capabilities.

### 5.5 QBPSubmit

```go
QBPSubmit(pid ProcID, op qbpcu.Opcode, srcs ...Operand) (ReqID, error)
```

Submits a QBP-CU operation on behalf of a process. Skuld performs:

1. Capability check: the operation's required ring must be in the process's `CapabilitySet`.
2. Ring transition: if the operation requires a different ring than the process's current ring, Skuld writes the new ring to `qbp_ctl.ALGEBRA_ID`.
3. Forwarding: the operation goes to the underlying `qbpcu.Accelerator` via `Submit(qbpcu.Req)`.

**Parameters:**
- `pid` — the calling process.
- `op` — the QBP-CU opcode (one of QFMA, QSAND, QNORM, QPERM, QPERMR, QNEAR, QDEC, QREC, QCOMM, QRING).
- `srcs` — operand list. Length and shapes depend on `op`; see the QBP-CU spec.

**Returns:**
- A fresh `ReqID` on success.
- `ErrProcNotFound`, `ErrInsufficientCapability`, `ErrQBPBusy`, `ErrQBPInvalidOp` as appropriate.

**Synchrony.** This call returns immediately after the request is queued. The result is retrieved via `QBPPoll(reqid)`. Skuld does not block the caller's goroutine waiting for the QBP-CU; this is essential for throughput on the prediction-verification loop.

**Op→ring mapping.** The required ring for each opcode at Crawl:
- QPERM, QPERMR, QNEAR, QDEC, QREC: `RingO` (𝕆 — kernel). Octonion ops; Fano-LUT.
- QFMA, QSAND, QNORM, QCOMM: `RingH` (ℍ — supervisor) for ℍ-typed operands; `RingC` (ℂ — user) for ℂ-typed operands. Operand type drives the requirement.
- QRING: `RingC` (ℂ — user). Read-only ring inquiry.

A user process (`RingC` only) can issue QFMA/QSAND/QNORM/QCOMM on ℂ-typed operands without capability escalation. Quaternion-typed operands trigger the `RingH` requirement, which the user process must satisfy via an ℍ-capability (the Hammer-simulation case).

### 5.6 QBPPoll

```go
QBPPoll(reqid ReqID) (Resp, error)
```

Retrieves the result of a previously-submitted QBP-CU operation. Blocks until the result is available or the context is cancelled (a context-aware variant `QBPPollCtx` exists for cancellation).

**Parameters:**
- `reqid` — the request ID returned by `QBPSubmit`.

**Returns:**
- `Resp` with the computed result and an `Invariant` snapshot.
- `ErrInvalidReqID` if the ID was never issued or has already been polled.

**The `Invariant` field.** Every Resp carries the `qbp_invariant` CSR snapshot at the moment of completion. This is what makes per-op privilege monitoring possible; user code that wants to observe the latest commutator/associator/alternator residue reads `Resp.InvariantSnapshot` instead of issuing a separate `QBPQueryInvariant` call. For most workloads this is the right pattern — physics analysis routines naturally interleave compute and check.

### 5.7 QBPQueryInvariant

```go
QBPQueryInvariant(pid ProcID) (Invariant, error)
```

Reads the `qbp_invariant` CSR for the QBP-CU lane currently associated with `pid`. The state reflects the most recently completed operation on that lane.

**Parameters:**
- `pid` — the calling process.

**Returns:**
- The `Invariant` snapshot.
- `ErrProcNotFound` if the process is not registered.

**Use case.** Out-of-band invariant inspection. Most user code uses `Resp.InvariantSnapshot` (returned by `QBPPoll`); this method exists for cases where the invariant is needed before the next op (e.g., adaptive precision algorithms that change strategy based on observed associator drift).

### 5.8 WyrdQuery

```go
WyrdQuery(pid ProcID, pattern QueryPattern, cap Capability) (Result, error)
```

Performs a one-shot query against Wyrd, with capability-based filtering applied by Skuld.

**Parameters:**
- `pid` — the calling process.
- `pattern` — Wyrd query pattern; structure defined in the `wyrd` package.
- `cap` — capability token authorizing access to the requested data; ignored at Crawl, used at Walk and beyond.

**Returns:**
- `Result` containing the matched data.
- `ErrWyrdQueryRefused` if Skuld's policy denies the query (e.g., user process attempting to read supervisor-ring nodes).
- `ErrWyrdUnavailable` if the Wyrd backend is unreachable.

**Capability filtering.** At Crawl, every query is filtered by `pid`'s `CapabilitySet`: nodes tagged with rings outside the set are removed from results. This is the database analog of the QBP-CU privilege check — a process cannot read nodes that exist in rings it has no capability for. Walk phase replaces this with native ring-typed Wyrd queries.

### 5.9 WyrdSubscribe

```go
WyrdSubscribe(pid ProcID, pattern QueryPattern, cap Capability) (Stream, error)
```

Establishes a long-lived subscription that streams matching events as they occur.

**Parameters:**
- Same as `WyrdQuery`.

**Returns:**
- `Stream` is an interface providing `Next(ctx) (Event, error)` and `Close() error`. Iterating yields events as they occur; closing terminates the subscription server-side.
- Errors as for `WyrdQuery`.

**Lifecycle.** Subscriptions persist until `Stream.Close()` is called, the process is destroyed, or Skuld shuts down. Each subscription holds resources in Wyrd; user code should close streams it no longer needs.

**Use case.** Watchdog-event-derived seam events forwarded by Skuld are accessed via this mechanism. A BMA judge subscribes to `pattern: { type: "physical_seam", strength: > 1e-3 }` and receives events as they occur.

### 5.10 wdEventNext (supervisor-internal)

```go
wdEventNext(ctx context.Context) (qbpcu.WDEvent, error)
```

Reads the next available `WDEvent` from the QBP-CU's host-pinned ring buffer.

**Parameters:**
- `ctx` — cancellable context; if cancelled before an event is available, returns `ctx.Err()`.

**Returns:**
- The next `WDEvent`, or an error.

**Internal use only.** Not exposed via `Supervisor`. The Skuld implementation runs a goroutine that calls this in a loop, applies `wdPolicyApply` to each event, and dispatches the result.

### 5.11 wdPolicyApply (supervisor-internal)

```go
wdPolicyApply(event qbpcu.WDEvent) (Action, error)
```

Applies privilege policy to a WDEvent and returns an `Action` indicating what to do.

```go
type Action struct {
    Type      ActionType
    TargetPid ProcID    // populated for ActionFault, ActionRingChange
    NewRing   RingID    // populated for ActionRingChange
    Forward   bool      // forward this event to Wyrd as a physical seam
}

type ActionType uint8

const (
    ActionNoOp ActionType = iota   // event below all thresholds; ignore
    ActionFault                     // event implies privilege violation; raise fault
    ActionRingChange                // authorize a ring transition
    ActionForwardPhysical           // forward to Wyrd as physical seam
)
```

**Policy logic at Crawl phase:**

1. Read `event.AlgebraID` and the current process's `CapabilitySet`.
2. Compute the relevant invariant magnitude (commutator for ℂ→ℍ, associator for ℍ→𝕆, alternator for 𝕆→𝕊).
3. If magnitude < ε_priv: `ActionNoOp`.
4. Else if magnitude < ε_phys and the process holds capability for the implied ring: `ActionRingChange` to that ring.
5. Else if magnitude < ε_phys and the process does NOT hold the capability: `ActionFault`.
6. Else (magnitude ≥ ε_phys): `ActionForwardPhysical` to Wyrd; capability check still applies.

**Internal use only.** Not exposed via `Supervisor`.

---

## 6. Construction and lifecycle

### 6.1 New

```go
func New(cfg Config) (Supervisor, error)
```

Creates a new Skuld supervisor. The `Config` struct holds:

```go
type Config struct {
    // The Accelerator implementation Skuld talks to. Typically a
    // qbpcu.Golden or qbpcu.RTLShim instance, or qbpcu.Mock for testing.
    Accelerator qbpcu.Accelerator

    // The Wyrd connection.
    Wyrd *wyrd.Connection

    // Privilege thresholds.
    EpsilonPriv float64 // typically 1e-5 for fp32
    EpsilonPhys float64 // typically 1e-3 for typical sims

    // Process limits.
    MaxProcesses uint32 // default 256

    // Capability master key (for Crawl-phase HMAC capability minting).
    MasterKey []byte // 32 bytes; required if any process uses non-RingC capabilities

    // Logger; if nil, logs to discard.
    Logger Logger
}
```

`New` validates the configuration, starts the WDEvent consumer goroutine, registers itself with the Wyrd connection (so seam events flow back), and returns a ready `Supervisor`.

### 6.2 Shutdown

```go
type ShutdownableSupervisor interface {
    Supervisor
    Shutdown(ctx context.Context) error
}
```

`Shutdown` triggers an orderly stop:

1. Refuse new `ProcCreate` and `QBPSubmit` calls (return `ErrSkuldShutdown`).
2. Wait for in-flight QBP-CU operations to complete or `ctx` to expire.
3. Send `ProcessState{StateTerminated}` to all active processes.
4. Close all Wyrd subscriptions.
5. Stop the WDEvent consumer goroutine.
6. Return.

If `ctx` expires before step 2 completes, `Shutdown` returns `ctx.Err()` and may leave processes in `StateBlocked`. Subsequent calls to the supervisor return `ErrSkuldShutdown`.

---

## 7. Concurrency model

### 7.1 Goroutines spawned by Skuld

A live Skuld instance runs the following goroutines:

1. **WDEvent consumer.** Single goroutine running `wdEventNext` → `wdPolicyApply` → action dispatch in a loop. Reads from the QBP-CU's ring buffer; applies policy; updates process state, emits faults, or forwards seam events to Wyrd.

2. **One goroutine per `Process`.** Each process has a goroutine that owns its Linux thread and serves its `pendingReqs` map. User code interacts with the process via the supervisor's API; the per-process goroutine is internal.

3. **One goroutine per active `WyrdSubscribe` stream.** Each subscription has a goroutine that reads from Wyrd and writes to the subscriber's channel.

The total goroutine count at typical workload (10 processes, 5 subscriptions): ~16. At maximum (256 processes, ~256 subscriptions): ~513. Comfortable for Go's runtime.

### 7.2 Synchronization

Skuld's process table and capability table are protected by a single read-write mutex. The WDEvent consumer holds the read lock during policy evaluation; `ProcGrantCapability` and similar mutators acquire the write lock briefly. Lock contention is not a known issue at expected workload — measure during C-09 integration.

The QBP-CU request map (`pendingReqs` per process) is protected by a per-process mutex; per-process goroutines do not block each other.

### 7.3 Ordering guarantees

- WDEvents are processed in order of receipt from the QBP-CU ring buffer. The QBP-CU guarantees deterministic round-robin order across SSCI and VCIX ports (per the SiFive spec §6).
- Within a single process, `QBPSubmit` calls return `ReqID`s in increasing order, and `QBPPoll(reqid_n)` is guaranteed to return after `QBPPoll(reqid_m)` for `n > m` only if the user polls them in that order (Skuld does not impose ordering on completion within a single op stream — that is a QBP-CU property).
- Between processes, no ordering guarantee. Each process is independent.

---

## 8. Failure semantics

### 8.1 Watchdog faults

A watchdog fault represents an algebraic invariant violation detected by the QBP-CU hardware. Skuld receives the fault via a `WDEvent` with a non-zero invariant residue and a corresponding `RingTransition` exceeding the policy threshold.

**Skuld's handling:**

1. The pending QBP-CU request that triggered the fault gets `Status = qbpcu.StatusWDFault`.
2. The fault code identifies which invariant fired (commutator, associator, alternator) and at what magnitude.
3. The process's `State` transitions to `StateFaulted`.
4. The process's per-process goroutine surfaces the fault via the next `QBPPoll`.

**Recovery.** A faulted process is not automatically destroyed; the user can inspect the fault, call `ProcSetAlgebra` to descend to a safer ring, and resume. Some workloads (Hammer simulation under aggressive timestep) may legitimately hit watchdog faults as a signal of physical seam crossings; these are not security violations.

The distinction between security violations and physical seam crossings is in the magnitude: ε_priv < |residue| < ε_phys → privilege event (security); |residue| ≥ ε_phys → physical seam (research signal). Skuld emits both; the user interprets.

### 8.2 Process crashes

If a Skuld-managed process panics (Go's panic mechanism), Skuld's per-process goroutine recovers and:

1. Marks the process `StateFaulted` with a special "host panic" fault code.
2. Cancels all `pendingReqs` for the process.
3. Does not destroy the process automatically; the user can inspect and decide.

Skuld itself does not panic except on unrecoverable invariant violations within its own state. Any Skuld panic is a bug.

### 8.3 QBP-CU unavailability

If the underlying `qbpcu.Accelerator` returns errors (e.g., the RTLShim loses cgo communication with Verilator), Skuld:

1. Returns `ErrQBPBusy` to new submissions.
2. Cancels in-flight requests with `Status = qbpcu.StatusCancelled`.
3. Logs the failure.
4. Attempts re-initialization on a backoff schedule (defined by the Accelerator implementation).

User code should treat `ErrQBPBusy` as transient and retry with backoff.

### 8.4 Wyrd unavailability

If the Wyrd connection fails, Skuld:

1. Returns `ErrWyrdUnavailable` to all `WyrdQuery` and `WyrdSubscribe` calls.
2. QBP-CU operations continue to function; only Wyrd access is gated.
3. Subscriptions in flight terminate with the error.

The QBP-CU and Wyrd are independent; the system degrades gracefully when one is down.

---

## 9. Examples

### 9.1 A simple Hammer simulation step

```go
package main

import (
    "context"
    "log"

    "github.com/JamesPagetButler/QBP/qbpcu"
    "github.com/JamesPagetButler/QBP/skuld"
    "github.com/JamesPagetButler/QBP/wyrd"
)

func main() {
    ctx := context.Background()

    sup, _ := skuld.New(skuld.Config{
        Accelerator:  qbpcu.NewGolden(),
        Wyrd:         wyrd.MustOpen("/var/lib/wyrd"),
        EpsilonPriv:  1e-5,
        EpsilonPhys:  1e-3,
        MaxProcesses: 16,
    })
    defer sup.(skuld.ShutdownableSupervisor).Shutdown(ctx)

    elf, _ := os.ReadFile("hammer.elf")
    pid, err := sup.ProcCreate(elf, "hammer-1", []skuld.RingID{skuld.RingC, skuld.RingH})
    if err != nil {
        log.Fatal(err)
    }
    defer sup.ProcDestroy(pid)

    // Hammer's inner loop: compute force, integrate.
    for step := 0; step < 1_000_000; step++ {
        // Force computation: quaternion FMA.
        reqid, err := sup.QBPSubmit(pid, qbpcu.OpQFMA, va, vb, vc)
        if err != nil {
            log.Fatalf("step %d: submit: %v", step, err)
        }
        resp, err := sup.QBPPoll(reqid)
        if err != nil || resp.Status != qbpcu.StatusOK {
            log.Fatalf("step %d: poll: %v %v", step, err, resp.Status)
        }
        // Check energy conservation via the invariant snapshot.
        if absI8(resp.InvariantSnapshot.Commutator) > 100 {
            log.Printf("step %d: commutator drift %v", step, resp.InvariantSnapshot.Commutator)
        }
        // ... integrate, update va/vb/vc ...
    }
}
```

### 9.2 GRB cross-correlation analysis

```go
// Subscribe to GRB-related seam events from BMA's ingestion pipeline.
stream, _ := sup.WyrdSubscribe(pid, wyrd.Pattern{
    Type:    "grb_pulse",
    Source:  "VLA_radio",
    Offset:  wyrd.Range{Min: -1.0, Max: 1.0},
}, nil) // capability nil at Crawl
defer stream.Close()

for {
    event, err := stream.Next(ctx)
    if err != nil { break }
    // Compute QBP correlation prediction for this pulse pair.
    pred, _ := computeQBPCorrelation(sup, pid, event)
    // Compare against measured.
    if math.Abs(pred - event.Measured) < 1e-4 {
        log.Printf("MATCH at offset %v", event.Offset)
    }
}
```

### 9.3 Granting a capability for the Hammer process

From `skuldctl`, the Crawl-phase capability minting tool:

```bash
$ skuldctl mint --pid 12345 --ring H --master-key /etc/skuld/master.key
Minted capability for process 12345, ring H
Token: 0xa3f8...d29c
```

The token is then passed to the Hammer process via its standard input or a file; the process calls `ProcGrantCapability` with the token at startup.

---

## 10. What this spec does NOT cover

- **Walk/Run/Sprint phase API.** This is the Crawl spec; later phases evaporate this API into Wyrd queries.
- **Performance tuning.** The 9-call surface is small enough that performance is dominated by the underlying QBP-CU and Wyrd, not by Skuld itself. Benchmarking is a C-09 deliverable.
- **The wdEventNext / wdPolicyApply implementation details.** These are exposed for testability but the actual implementation in `skuld` is internal to that package.
- **Integration with BMA's judge collective.** That is downstream — BMA subscribes to Wyrd, not directly to Skuld. The architecture stack from §3 of the v0.2 supervisor doc applies.
- **Multi-host Skuld.** Crawl is single-host. Run-phase Skuld federates via NATS, but that's a Run spec, not this one.

---

## 11. References

**Internal:**
- *Wyrd-Supervisor-Architecture-v0.2.md* — the architectural document this spec implements.
- *Wyrd-Workload-ISA-v0.1.md* — the workload analysis driving the 10-instruction QBP-CU ISA.
- *QBP-CU SiFive Interface Specification v0.2* — the hardware interface (pending revision per the spec review).
- *Wyrd-Spec.md* — the database spec that Skuld queries.
- *Wyrd-T2.3-Capability-Soundness-v0.1.lean* — the formal foundation of the capability mechanism.

**External:**
- DBOS papers (Stonebraker, Zaharia, et al. 2020-2024).
- seL4 capability semantics (Klein et al.).
- SiFive VCIX Software Specification v1.0.

---

## 12. Status

**Rev 1.0 — DRAFT.** This is the inaugural Skuld API spec. Not yet implemented. Implementation is ticket C-06 in the Crawl plan.

The spec is intended to be a binding contract between architecture (decided) and implementation (forthcoming). Changes prior to implementation are expected and welcome; changes after implementation lands in C-06 should be versioned (Rev 1.1, 1.2 ...) with a clear migration path.

---

*End of Skuld-Spec v1.0 — DRAFT*
