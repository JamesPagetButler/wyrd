# Wyrd Lean — Class B Phase 2 Session Report

**Helpful Engineering — Quaternion-Based Physics Programme**
April 2026 — Class B (hypergraph reasoning) Lean session

## Result

`lake build` succeeds with **zero sorries, zero user-defined axioms** across all 10 files (7 Phase 1 + 3 Phase 2 added this session).

```
Build completed successfully (2041 jobs).
✔ Wyrd.Hypergraph    (NEW — Phase 2)
✔ Wyrd.CTH           (NEW — Phase 2)
✔ Wyrd.Bridge        (NEW — Phase 2)
✔ Wyrd.CayleyDickson
✔ Wyrd.Foundations
✔ Wyrd.Projection
✔ Wyrd.Capability
✔ Wyrd.Noise
✔ Wyrd.SedenionWitness
✔ Wyrd.OctonionAlternative
✔ Wyrd
```

113 total theorems/definitions across 10 files. Toolchain unchanged: lean v4.30.0-rc1, mathlib rev `a090f46d`.

## Closed gaps

The three Class B gaps from `Wyrd-Workload-ISA-v0.2.md` §3.8:

| Gap | Theorem | File | Status |
|---|---|---|---|
| Graph-invariant preservation under hyperedge addition | `hyperedge_preserves_incident_edges` (C-20a) | Hypergraph.lean | ✅ proven |
| CTH metric monotonicity (Tier-2) | `cth_measurement_evidence_monotonic` (C-20b) | CTH.lean | ✅ proven |
| Bridge atomicity (conservation form) | `bridge_promote_preserves_count` (C-20c) | Bridge.lean | ✅ proven |

## What was built

### `Wyrd/Hypergraph.lean` — basic hypergraph foundation (12 defs/theorems)

- `HyperEdge V` and `Graph V` structures with explicit `DecidableEq` instance (manual, not `deriving`, due to typeclass ambiguity in this Lean version — same content)
- `incident`, `incidentEdges`, `incomingEdges`, `outgoingEdges`, `addEdge` operations
- `not_incident_of_not_in_edge` — local helper
- **`hyperedge_preserves_incident_edges` (C-20a)** — non-incident edge addition leaves a node's incident-edge set untouched
- `invariant_under_nonincident_addition` — abstract corollary: any function f depending only on `incidentEdges v` is preserved
- `hyperedge_preserves_incoming_edges` and `hyperedge_preserves_outgoing_edges` — directional specializations

**Proof tactic for C-20a:** `unfold; rw [addEdge_edges, Finset.filter_insert]; simp [not_incident]`. Standard Finset-filter manipulation.

### `Wyrd/CTH.lean` — entropy and monotonicity (5 defs/theorems)

- `TrustTier` enum (axiom, proof, measurement, prediction) matching CTH paper Definition 2
- `measurementEntropy(δ) = -log(1 - δ)` for the Tier-2 entropy formula
- `log_monotone_on_positive` — local helper using `Real.log_le_log`
- **`cth_measurement_evidence_monotonic` (C-20b)** — better evidence (δ' ≤ δ) yields lower or equal entropy
- `proofEntropy = 0` (Tier-1 boundary)
- `cth_zero_error_zero_entropy` — boundary case η(0) = 0

**Proof tactic for C-20b:** show 1-δ ≤ 1-δ' from h_evidence; both positive from upper bounds; `Real.log_le_log` gives log(1-δ) ≤ log(1-δ'); negate via `linarith`.

### `Wyrd/Bridge.lean` — promotion atomicity (8 defs/theorems)

- `State` with `contextusQueue` and `cthQueue` (Finset Signal each)
- `signalCount` — observable total = sum of queue cardinalities
- `promote` — atomic transition: erase from contextus, insert into cth
- Simp lemmas for promote's effect on each queue
- **`bridge_promote_preserves_count` (C-20c)** — promotion conserves total signal count
- `bridge_promote_signal_in_cth` — post-state has signal in CTH
- `bridge_promote_signal_not_in_contextus` — post-state has signal NOT in Contextus
- `bridge_promote_exactly_one_side` — combined: signal is in exactly one queue

**Proof tactic for C-20c:** unfold signalCount; `card_erase_of_mem` (decreases by 1) and `card_insert_of_notMem` (increases by 1); `card_pos` for the natural-subtraction handling; `omega` closes.

### `Wyrd.lean` top-level updated

Imports the three new Phase 2 files alongside the seven Phase 1 files.

### `Wyrd-Proofs-Reference-v1.1.md`

Promoted from v1.0 to v1.1, adding §§12–14:
- §12: Hypergraph types and C-20a
- §13: CTH entropy and C-20b
- §14: Bridge promotion and C-20c

Phase 1 content (§§1–11) is unchanged and lives in v1.0; v1.1 is additive.

## Mathlib API drift encountered

Two API names had drifted between when the proof sketches were drafted and the current mathlib snapshot:

1. `Finset.card_insert_of_not_mem` → **`Finset.card_insert_of_notMem`** (camelCase shift; common pattern in mathlib4)
2. `Finset.not_mem_erase` doesn't exist as a direct name; derived from `Finset.mem_erase` via `intro h; exact (Finset.mem_erase.mp h).1 rfl` instead

Both fixes are in the committed code; updated names match the rev `a090f46d` snapshot.

## Honest accounting — scope choices and deviations

### Bridge atomicity — conservation form vs full state-machine form

The C-20c theorem proves the **conservation form** of atomicity: signals are conserved under promotion (no losses, no duplicates, exactly-one-side). This is the substantive content for Class B integrity claims.

It does **NOT** prove the **full state-machine form** of atomicity (no partial-state observable to external observers, modeled with explicit transitions and visibility). That would require a process-calculus or temporal-logic formalism — substantially more setup than the conservation proof.

**Tracked as Phase 3 deferred work** in Bridge.lean §5 status block and Proofs Reference v1.1 §14 scope note.

### CTH not yet integrated with Hypergraph at the metric level

The `measurementEntropy` function is defined on raw real numbers (δ → ℝ), not on hypergraph nodes. To make full CTH evaluator soundness claims, we'd want `eta : Graph V → V → ℝ` indexed by tier — and theorems like "η is invariant under non-incident edge addition for non-incident nodes" composed from C-20a + C-20b.

This composition is a one- to two-day extension and not blocking. C-20b alone closes the "evidence cannot raise η" gap that was the load-bearing claim for Class B production.

### Cosmetic warnings (2)

```
warning: Wyrd/CTH.lean:114:5: unused variable `h_δ_lower`
warning: Wyrd/CTH.lean:115:5: unused variable `h_δ'_lower`
```

Both are range-bound hypotheses that document the function's domain (δ ∈ [0, 1)) but aren't strictly needed by the proof body — `linarith` doesn't use them because the proof routes through `1 - δ' ≤ 1 - δ` which only needs `h_evidence`. Could rename to `_h_δ_lower` to silence; left as-is to keep the domain assumption visible to readers.

### `deriving DecidableEq` failed for `HyperEdge`

When `HyperEdge` was declared with `deriving DecidableEq`, the build hit:
```
synthesized type class instance is not definitionally equal to expression
inferred by typing rules, synthesized inst✝ inferred inst✝¹
```

This is a typeclass-instance ambiguity (two `DecidableEq V` instances in scope, the structure parameter and an inherited variable). Workaround: declared the instance manually with explicit pattern-matching on the structure constructors. Same content, slightly more verbose. Tracked as a low-priority cleanup if a future mathlib bump resolves it.

## What this unblocks

Per `Wyrd-Implementation-Plan-v1.0.md` C-20a/b/c are listed as Phase 2 acceptance gates. With them proven, the Phase 2 acceptance gate for Class B Lean is now achievable — depends on the rest of Phase 2 (qbpcu Golden, Hammer integration, W1/W3 prototypes, Contextus prototype) catching up.

The proven theorems are immediately citable from any Go code. The citation pattern (per Proofs Reference v1.1 §16):

```go
// cth.evaluator.AddDerivation appends a new derivation hyperedge.
// Soundness: incremental updates have local effect by
// Hypergraph.hyperedge_preserves_incident_edges. Bounded blast radius.
```

```go
// cth.evaluator.UpdateEvidence updates a Tier-2 measurement node.
// Soundness: if δ' ≤ δ, post-update entropy η' ≤ η by
// CTH.cth_measurement_evidence_monotonic. Evidence is monotonically beneficial.
```

```go
// bridge.Promote moves a signal from Contextus to CTH atomically.
// Soundness: signal count is preserved by Bridge.bridge_promote_preserves_count;
// signal is in exactly one queue post-promotion by
// Bridge.bridge_promote_exactly_one_side. No losses, no duplicates.
```

## What's next

**Phase 3 — Class C operational semantics.** Four theorems still open per the Implementation Plan §2.4:

| ID | Theorem | Blocker |
|---|---|---|
| C-21a | `capability_invariant_under_cart_switch` | needs `Cart-as-Context` spec (C-16, James-direct) |
| C-21b | `cart_switch_atomic` | needs Wyrd transaction model spec (C-17, James-direct) |
| C-21c | `judge_collective_deterministic` | independent — can start any time |
| C-21d | `self_modification_requires_approval` | depends on C-21c judge model |

Phase 3 is operationally more complex (state machines, observability, vote aggregation) than Phase 2's algebraic content. Not blocking implementation work; runs in parallel.

## Summary table of session deliverables

| Artifact | Path | Size |
|---|---|---|
| Hypergraph proofs | `Wyrd/Hypergraph.lean` | 12 defs/theorems |
| CTH proofs | `Wyrd/CTH.lean` | 5 defs/theorems |
| Bridge proofs | `Wyrd/Bridge.lean` | 8 defs/theorems |
| Top-level imports | `Wyrd.lean` (updated) | 10 files imported |
| Proofs reference | `~/Documents/Wyrd/Archive/Wyrd-Proofs-Reference-v1.1.md` | supersedes v1.0 |
| This report | `~/Documents/Wyrd/wyrd-lean-project/SESSION-REPORT-Class-B.md` | this file |

`lake build` time on warm cache: ~10 sec. Cold rebuild: ~3 min.

## Honest claim of what was verified

- **Compiled with `lake build` to completion.** Yes.
- **Zero sorries in any proof body.** Yes (verified by `grep -rEn "^\s*sorry\b"`).
- **Zero user-defined axioms.** Yes (verified by `grep -rEn "^axiom "`).
- **Theorems match the architectural claims they're stated to prove.** Yes — each theorem statement is in this report; readers can cross-check against the Lean source.
- **Mathematics matches the CTH theory paper / Workload-ISA spec.** Yes — theorem statements are direct formalizations of the gap descriptions in `Wyrd-Workload-ISA-v0.2.md` §3.8.

What was NOT done in this session:
- Full state-machine atomicity for the Bridge (conservation form only)
- CTH metric integration with Hypergraph (entropy is on ℝ, not on graph nodes)
- Phase 3 / Class C theorems
- Branch A sensitivity for the new theorems

These are tracked deferrals, not silent omissions.

---

*End of Wyrd Lean Class B Phase 2 Session Report.*
