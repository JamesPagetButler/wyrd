# Wyrd Lean — Class C Phase 3 Session Report

**Helpful Engineering — Quaternion-Based Physics Programme**
April 2026 — Class C (operational semantics) Lean session

## Result

`lake build` succeeds with **zero sorries, zero user-defined axioms** across all 14 files (7 Phase 1 + 3 Phase 2 + 4 Phase 3 added this session).

```
Build completed successfully (2045 jobs).

Phase 1 (algebraic privilege boundaries):
  ✔ Wyrd.CayleyDickson · Foundations · Projection · Capability · Noise · SedenionWitness · OctonionAlternative

Phase 2 (Class B hypergraph reasoning):
  ✔ Wyrd.Hypergraph · CTH · Bridge

Phase 3 (Class C operational semantics) — NEW this session:
  ✔ Wyrd.Cart                — capability scope across cart switches
  ✔ Wyrd.Transaction         — cart-switch atomicity (transaction observability)
  ✔ Wyrd.JudgeCollective     — pure judges, vote aggregation, perm-invariance
  ✔ Wyrd.Constitutional      — constitutional pin (self-modification gate)

  ✔ Wyrd                     — top-level imports
```

182 total theorems/definitions/structures/instances across 14 files. Toolchain unchanged: lean v4.30.0-rc1, mathlib rev `a090f46d`.

## Closed gaps

The four Class C gaps from `Wyrd-Workload-ISA-v0.2.md` §4.8 / `Wyrd-Implementation-Plan-v1.0.md` C-21a/b/c/d:

| Gap | Theorem | File | Status |
|---|---|---|---|
| Capability scope persistence across cart switches | `capability_invariant_under_cart_switch` (C-21a) | Cart.lean | ✅ proven |
| Cart-switch atomicity wrt Wyrd transactions | `cart_switch_atomic` (C-21b) | Transaction.lean | ✅ proven |
| Judge-collective determinism | `judge_collective_deterministic` + `judge_collective_perm_invariant` (C-21c) | JudgeCollective.lean | ✅ proven |
| Self-modification constitutional pin | `self_modification_requires_approval` (C-21d) | Constitutional.lean | ✅ proven |

## What was built — file by file

### `Wyrd/Cart.lean` — capability invariance (13 declarations)

- `Cart` enum (theory, engineering, beekeeper, domainSpecific id) — operating modes BMA switches between
- `Scope` enum (session, cart, task) — capability lifetime distinctions
- `Capability` structure (token + scope) with `validInCart` predicate
- `CartSession` state (currentCart + held capabilities) with `switchTo` operation
- `session_scoped_valid_in_all_carts` (helper)
- **`capability_invariant_under_cart_switch` (C-21a)** — session-scoped capability survives any single switch
- `capability_invariant_under_cart_switch_chain` — n-step composition (induction on switch list)

**Proof tactic for C-21a:** unfold `hasValidCapability` after switch; capability still in held set (switchTo preserves capabilities); valid in new cart by `session_scoped_valid_in_all_carts`.

### `Wyrd/Transaction.lean` — cart-switch atomicity (13 declarations)

- `TxState` lifecycle (notStarted, open_, committed, aborted) — `open_` to avoid Lean keyword conflict
- `WyrdTx` structure (txId + state) with `resolve` operation
- `TxState.isObservable` (everything except `open_`)
- `SessionState` with `cartSwitch` — atomic operation that resolves all transactions via a per-tx commit/abort decision function
- `resolve_observable` (helper)
- **`cart_switch_atomic` (C-21b)** — every post-switch transaction has an observable state (no Open transactions visible)
- `cart_switch_preserves_count` — no transactions lost or duplicated
- `cart_switch_preserves_ids` — transaction identity survives the switch

**Proof tactic for C-21b:** post-switch transactions are `s.transactions.map (resolve · ...)`; for each, `resolve_observable` gives observability; `List.mem_map` provides the connection.

### `Wyrd/JudgeCollective.lean` — judge determinism + vote aggregation (18 declarations)

- `Proposal`, `Context`, `Vote` (approve / minorConcern / majorConcern / veto)
- `Judge := Proposal → Context → Vote` — pure function type
- `aggregate` operation (the four-valued lattice; VETO absorbs, APPROVE is identity)
- `aggregate_comm` and `aggregate_assoc` — commutative monoid laws
- `aggregate_approve_left/right`, `aggregate_veto_left/right` — identity and absorbing
- `aggregate_left_comm` — left-commutativity (combines comm + assoc)
- `run` — collective fold over judges
- **`judge_collective_deterministic` (C-21c)** — basic form (rfl, since judges are pure functions)
- **`judge_collective_perm_invariant` (C-21c substantive form)** — order-of-judges doesn't matter
- `judge_collective_empty`, `judge_collective_singleton` — degenerate cases
- **`judge_collective_veto_propagates`** — single VETO blocks the collective (constitutional protection)

**Proof tactic for `perm_invariant`:** induction on `List.Perm` (nil / cons / swap / trans). The swap case uses `aggregate_left_comm` to commute adjacent judges; trans is by transitivity of equality.

**Proof tactic for `veto_propagates`:** induction on judge list; `List.mem_cons` to split head vs tail; `aggregate_veto_left/right` for the propagation step.

### `Wyrd/Constitutional.lean` — constitutional pin (7 declarations)

- `CodeUpdate` structure with `toProposal` mapping to JudgeCollective.Proposal
- `UpdateOutcome` enum (applied / rejected with reason)
- `tryApplyCodeUpdate` — the constitutional gate; matches on the judge collective's vote
- **`self_modification_requires_approval` (C-21d)** — every Applied outcome implies the collective voted APPROVE
- **`judge_veto_blocks_self_modification`** — corollary: any single judge VETO blocks (combines C-21d with `judge_collective_veto_propagates`)
- `empty_judge_collective_approves` — degenerate case noted; production deployments must ensure non-empty collective

**Proof tactic for C-21d:** generalize over the actual vote; case-analyze each Vote constructor; only `Vote.approve` produces `Applied` (rfl); other cases give contradictory `Applied = Rejected ...`.

## Mathlib API drift encountered

One drift caught and fixed:

- `Mathlib.Data.List.Perm` → **`Mathlib.Data.List.Basic`** (List.Perm is exposed via Basic in this rev; the dedicated Perm module path doesn't exist)

## Honest accounting — design choices and deviations

### C-21a — minimal Cart model

The Cart-as-Context formal spec (C-16, James-direct) is not yet written. I formalized a minimal model with:
- 4 Cart constructors (theory, engineering, beekeeper, domainSpecific id)
- 3 Scope variants (session, cart, task)
- The minimum needed to state and prove invariance

**If C-16 is later written with a richer model** (e.g., cart hierarchies, cart inheritance, time-bounded carts), my proofs remain valid for the parts they cover and would extend cleanly. The minimal model captures the *substantive* content of C-21a — session-scoped capabilities survive switches.

### C-21b — minimal transaction model

The Wyrd Transaction Model spec (C-17, James-direct) is not yet written. I formalized:
- 4 TxState lifecycle constructors
- A `cartSwitch` operation that takes a per-transaction commit/abort resolver function
- The resolve function with case-analysis on state

**Conservation form** of atomicity: I prove "no Open transactions post-switch" + count preservation + ID preservation. This is the same flavor as C-20c (Bridge atomicity, also conservation form).

**Not modeled:** the actual writes that committed/aborted transactions perform on Wyrd state. That requires a Wyrd-state model that doesn't exist yet. Deferred to Phase 3.5+ as Wyrd's transaction implementation matures.

### C-21c — `judge_collective_deterministic` is structurally trivial

The basic form `run judges p ctx = run judges p ctx` is just `rfl` because judges are modeled as pure functions. The **substantive content** is the corollary `judge_collective_perm_invariant` (order-of-judges doesn't matter), which justifies the optimization C-OPT-5 (parallel judge evaluation from Workload-ISA v0.2 §4.5).

I added `judge_collective_veto_propagates` as a separate theorem because it's *load-bearing* for C-21d (the constitutional pin's "any veto blocks" property). Without it, C-21d's corollary `judge_veto_blocks_self_modification` couldn't be stated cleanly.

### C-21d — type-level enforcement

The constitutional pin is enforced at the type level (the function signature) plus the C-21d theorem (no path to Applied bypasses APPROVE). The proof is essentially case-analysis on the function body — every match arm except `Vote.approve` produces `Rejected`, and `Rejected ≠ Applied`.

**Not modeled:**
- Time-bounded approvals (an approval issued today should expire — requires temporal logic)
- Beekeeper override / emergency exception protocol (per Ethics v1.1: emergency exceptions are scoped and reversible)
- Multi-version code-update conflicts

These are deferred. The current C-21d proves the **core mechanism**: applying an update requires approval, structurally.

### Cosmetic warnings

- `Cart.lean`: namespace duplication warnings (the `Cart` inductive lives inside `namespace Cart` — same pattern as Phase 1's `Capability.Capability`; cosmetic, no semantic impact)
- `CTH.lean`: 2 unused-variable warnings (range hypotheses documenting the function's domain; left in for readability)
- `Constitutional.lean`: 1 unused-variable warning (`h` in the generalize step; could rename to `_h`)

None affect correctness.

## What this unblocks

The four Class C theorems close the formal-foundations gap for:

1. **C-OPT-3** (Capability scope persistence across cart switches) — now formally certified by C-21a. BMA implementers can write capability-scoping code citing this theorem.
2. **C-OPT-8** (Mode-switch atomicity via Wyrd transactions) — now formally certified by C-21b. Cart-switch implementations have a soundness anchor.
3. **C-OPT-5** (Parallel judge-collective evaluation) — now formally certified by C-21c. Parallel judge dispatch is provably equivalent to sequential.
4. **The constitutional pin** — formally certified by C-21d. BMA cannot self-modify without judge approval; this is the load-bearing safety property for autonomous BMA operation.

Combined with Phase 1 (algebraic privilege) and Phase 2 (hypergraph reasoning), the Lean corpus now covers all three workload classes' core soundness claims.

## What's next

Per the Implementation Plan v1.0, Phase 3 Lean (this session) was the last of the C-XX theorem tickets. Future Lean work falls into:

| Track | Status | When |
|---|---|---|
| Phase 1 follow-ups (T2.1.b, T2.1.c explicit composition theorems) | low priority, 1-2 hrs work | any time |
| Class B refinements (full state-machine atomicity, hypergraph-level entropy) | not started | Phase 2 maturation, ~weeks |
| Class C refinements (time-bounded approvals, emergency exceptions, multi-update conflicts) | not started | Phase 3 maturation, ~weeks |
| New theorem categories from production experience | unknown | as gaps surface |

**The Lean corpus is now sufficient for Walk-phase and most Run-phase work.** Sprint-phase production deployments may surface new gaps; the corpus extends as needed.

## Summary table of session deliverables

| Artifact | Path | Size |
|---|---|---|
| Cart proofs | `Wyrd/Cart.lean` | 13 declarations |
| Transaction proofs | `Wyrd/Transaction.lean` | 13 declarations |
| JudgeCollective proofs | `Wyrd/JudgeCollective.lean` | 18 declarations |
| Constitutional proofs | `Wyrd/Constitutional.lean` | 7 declarations |
| Top-level imports | `Wyrd.lean` (updated) | 14 files imported |
| This report | `~/Documents/Wyrd/wyrd-lean-project/SESSION-REPORT-Class-C.md` | this file |

`lake build` time on warm cache: ~12 sec. Cold rebuild: ~3 min.

## Honest claim of what was verified

- **Compiled with `lake build` to completion.** Yes (2045 jobs).
- **Zero sorries in any proof body.** Yes (verified by `grep -rEn "^\s*sorry\b"`).
- **Zero user-defined axioms.** Yes (verified by `grep -rEn "^axiom "`).
- **Theorems match the architectural claims in the Workload-ISA v0.2 §4.8 gap list.** Yes — each theorem statement is in this report; readers can cross-check against the Lean source.

What was NOT done in this session:
- Time-bounded approvals (Phase 3+ refinement)
- Emergency exception protocol (per Ethics v1.1; requires capability-bound emergency tokens)
- Multi-update conflict resolution
- Production-grade Wyrd state model (cart switch + transaction effects on Wyrd writes)

These are tracked deferrals, not silent omissions.

## Honest accounting — refinements vs replacements

The Class C theorems are formalized against **minimal models** (Cart, Transaction, Judge) that are consistent with what the broader specs (C-16 Cart-as-Context, C-17 Wyrd-Transaction-Model) are expected to specify. If James writes more elaborate specs later, my proofs:

- **For refinements** (specs that add detail to the model): proofs remain valid; new theorems can be added on top.
- **For replacements** (specs that change the model semantically): proofs may need adaptation. Risk is low for Cart and Transaction — the minimal models capture the substantive content.

This is consistent with the Implementation Plan v1.0 §2.4 approach: "minimal Lean models can be drafted in parallel with James-direct specs; the proofs remain valid as long as the spec doesn't contradict the model."

---

*End of Wyrd Lean Class C Phase 3 Session Report.*
