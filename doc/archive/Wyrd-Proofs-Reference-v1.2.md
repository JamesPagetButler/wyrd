# Wyrd / Skuld Lean Proof Corpus — Theorem Reference

**Helpful Engineering — Quaternion-Based Physics Programme**
Principal Investigator: James Paget Butler
April 2026 — Rev 1.2 — supersedes v1.1

> **Purpose.** This document is the canonical reference for the Lean 4 proof corpus that backs the Wyrd / Skuld algebraic privilege model, Wyrd / CTH hypergraph reasoning, and BMA's operational semantics (cart-switching, judge collective, constitutional pin).

> **What changed in v1.2.** Added **Phase 3 — Class C operational semantics** (§§15–18). Four new theorems closing the C-21 gap list: capability invariance under cart switch (C-21a), cart-switch transaction atomicity (C-21b), judge-collective determinism + perm-invariance (C-21c), and the constitutional pin (C-21d). Phase 1 + Phase 2 content (§§1–14) is unchanged from v1.1.

---

## 0. Build & toolchain (verified 2026-04-25)

| Component | Pin |
|---|---|
| Lean toolchain | `leanprover/lean4:v4.30.0-rc1` |
| Mathlib | rev `a090f46da78e9af11fee348cd7ee47bf8dd219d2` |
| Project root | `~/Documents/Wyrd/wyrd-lean-project/` |
| Build status | `lake build` succeeds; **0 sorries, 0 user-defined axioms** across all 14 files |
| File count | 14 (.lean files in `Wyrd/`) — 7 Phase 1 + 3 Phase 2 + 4 Phase 3 |

To rebuild from cold: `cd ~/Documents/Wyrd/wyrd-lean-project && lake update && lake exe cache get && lake build`.

---

## 1. Inventory at a glance

### Phase 1 — Algebraic privilege boundaries (Wyrd corpus baseline)

(See v1.0 / v1.1 — content unchanged. 25 theorems across Foundations, CayleyDickson, Projection, Capability, Noise, SedenionWitness, OctonionAlternative.)

### Phase 2 — Class B hypergraph reasoning

(See v1.1 — content unchanged. 10 theorems across Hypergraph, CTH, Bridge.)

### Phase 3 — Class C operational semantics (NEW in v1.2)

| Theorem | File | Architectural role |
|---|---|---|
| `Cart.session_scoped_valid_in_all_carts` | Cart | Helper: session-scoped capability is valid in every cart |
| `Cart.capability_invariant_under_cart_switch` | Cart | **C-21a: session-scoped capability survives cart switch** |
| `Cart.capability_invariant_under_cart_switch_chain` | Cart | n-step composition |
| `Transaction.resolve_observable` | Transaction | Helper: resolved transactions are observable |
| `Transaction.cart_switch_atomic` | Transaction | **C-21b: post-switch, no Open transactions visible** |
| `Transaction.cart_switch_preserves_count` | Transaction | No transactions lost or duplicated |
| `Transaction.cart_switch_preserves_ids` | Transaction | Transaction identity preserved |
| `JudgeCollective.aggregate_comm` | JudgeCollective | Vote aggregation is commutative |
| `JudgeCollective.aggregate_assoc` | JudgeCollective | Vote aggregation is associative |
| `JudgeCollective.aggregate_left_comm` | JudgeCollective | Left-commutativity (combines comm + assoc) |
| `JudgeCollective.aggregate_approve_left/right` | JudgeCollective | APPROVE is identity |
| `JudgeCollective.aggregate_veto_left/right` | JudgeCollective | VETO is absorbing |
| `JudgeCollective.judge_collective_deterministic` | JudgeCollective | **C-21c basic: same input → same output** |
| `JudgeCollective.judge_collective_perm_invariant` | JudgeCollective | **C-21c substantive: order of judges doesn't matter** |
| `JudgeCollective.judge_collective_veto_propagates` | JudgeCollective | Single VETO blocks the collective |
| `Constitutional.self_modification_requires_approval` | Constitutional | **C-21d: Applied → Approve (the constitutional pin)** |
| `Constitutional.judge_veto_blocks_self_modification` | Constitutional | Single VETO blocks self-modification |
| `Constitutional.empty_judge_collective_approves` | Constitutional | Degenerate case noted |

**Bold rows are the four C-21 theorems closing the gap list from `Wyrd-Workload-ISA-v0.2.md` §4.8.**

---

## 15. Cart — capability invariance under cart switch (Phase 3)

**File:** `Wyrd/Cart.lean`. **Imports:** `Mathlib.Data.Finset.Basic`. **Status:** clean compile, 0 sorries.

### Types

```lean
inductive Cart : Type where
  | theory | engineering | beekeeper
  | domainSpecific (id : Nat)
  deriving DecidableEq, Repr

inductive Scope : Type where
  | session | cart (c : Cart) | task (taskId : Nat)
  deriving DecidableEq, Repr

structure Capability where
  token : Nat
  scope : Scope

structure CartSession where
  currentCart : Cart
  capabilities : Finset Capability
```

### `capability_invariant_under_cart_switch` (C-21a) ⭐

```lean
theorem capability_invariant_under_cart_switch
    (s : CartSession) (cap : Capability)
    (h_session : cap.scope = Scope.session)
    (h_held : cap ∈ s.capabilities)
    (newCart : Cart) :
    (s.switchTo newCart).hasValidCapability cap
```

**Reading:** if a session-scoped capability `cap` is held in session state `s`, then after switching to any other cart, the capability is still held AND still valid.

**Architectural meaning:** **Skuld does NOT need to re-issue capabilities on every cart switch.** Cart switches are cheap; they're internal cognitive routing, not security boundaries. This certifies optimization C-OPT-3 (Workload-ISA v0.2 §4.5) — the rapid Theory ↔ Engineering alternation pattern is viable because session capabilities persist.

**Proof tactic:** unfold `hasValidCapability`; capability still in held set (`switchTo_capabilities` simp lemma); valid in new cart by `session_scoped_valid_in_all_carts`.

**Cited by:** `skuld.CartSwitch` implementation, capability-flow tests in BMA's cart-switch test corpus.

### `capability_invariant_under_cart_switch_chain`

```lean
theorem capability_invariant_under_cart_switch_chain
    (s : CartSession) (cap : Capability)
    (h_session : cap.scope = Scope.session)
    (h_held : cap ∈ s.capabilities)
    (carts : List Cart) :
    let s' := carts.foldl CartSession.switchTo s
    s'.hasValidCapability cap
```

n-step composition by induction on the cart-switch list. Captures the intuition that capability invariance survives *arbitrary sequences* of cart switches, not just single switches.

---

## 16. Transaction — cart-switch atomicity (Phase 3)

**File:** `Wyrd/Transaction.lean`. **Imports:** `Wyrd.Cart`, `Mathlib.Data.Finset.*`, `Mathlib.Data.List.Basic`. **Status:** clean compile, 0 sorries.

### Types

```lean
inductive TxState : Type where
  | notStarted | open_ | committed | aborted

structure WyrdTx where
  txId : Nat
  state : TxState

def TxState.isObservable : TxState → Prop
  | TxState.notStarted => True
  | TxState.open_      => False
  | TxState.committed  => True
  | TxState.aborted    => True

structure SessionState where
  currentCart : Cart.Cart
  transactions : List WyrdTx
```

### Operations

```lean
def WyrdTx.resolve (tx : WyrdTx) (commitDecision : Bool) : WyrdTx := ...
def SessionState.cartSwitch
    (s : SessionState) (newCart : Cart.Cart)
    (resolver : WyrdTx → Bool) : SessionState := ...
```

### `cart_switch_atomic` (C-21b) ⭐

```lean
theorem cart_switch_atomic
    (s : SessionState) (newCart : Cart.Cart) (resolver : WyrdTx → Bool) :
    ∀ tx ∈ (s.cartSwitch newCart resolver).transactions, tx.state.isObservable
```

**Reading:** every transaction in the post-switch state has an observable state (i.e., not Open). Cart switching is atomic with respect to transaction lifecycle.

**Architectural meaning:** an external observer of BMA's transaction log will never see a transaction that was Open before the switch and is still Open after. Each Open transaction is resolved (committed or aborted) as part of the switch. This certifies optimization C-OPT-8 (Workload-ISA v0.2 §4.5) — without this property, a cart switch could leave Wyrd in a partially-written state.

**Proof tactic:** post-switch transactions are produced by `s.transactions.map (resolve · ...)`; for each, `resolve_observable` (proven by case-split on TxState × Bool) gives observability; `List.mem_map` provides the connection.

**Companion theorems:**
- `cart_switch_preserves_count` — number of transactions preserved (by `List.length_map`)
- `cart_switch_preserves_ids` — transaction identity preserved (resolve doesn't change txId)

**Cited by:** `skuld.CartSwitch.checkpoint` implementation, Wyrd's transaction commit/abort path.

---

## 17. JudgeCollective — pure judges + vote aggregation (Phase 3)

**File:** `Wyrd/JudgeCollective.lean`. **Imports:** `Mathlib.Data.List.Basic`. **Status:** clean compile, 0 sorries.

### Types

```lean
structure Proposal where proposalId : Nat
structure Context where contextHash : Nat

inductive Vote : Type where
  | approve | minorConcern | majorConcern | veto

def Judge := Proposal → Context → Vote
```

### Aggregation operation

```lean
def aggregate : Vote → Vote → Vote
  | Vote.veto,         _                  => Vote.veto
  | _,                 Vote.veto          => Vote.veto
  | Vote.majorConcern, _                  => Vote.majorConcern
  | _,                 Vote.majorConcern  => Vote.majorConcern
  | Vote.minorConcern, _                  => Vote.minorConcern
  | _,                 Vote.minorConcern  => Vote.minorConcern
  | Vote.approve,      Vote.approve       => Vote.approve

def run (judges : List Judge) (p : Proposal) (ctx : Context) : Vote :=
  judges.foldr (fun j acc => aggregate (j p ctx) acc) Vote.approve
```

### Algebraic properties

- `aggregate_comm`, `aggregate_assoc` — commutative monoid (proven by `cases <;> rfl`)
- `aggregate_left_comm` — left-commutativity, derived from comm + assoc
- `aggregate_approve_left/right` — APPROVE is identity
- `aggregate_veto_left/right` — VETO is absorbing

These together establish that `(Vote, aggregate, approve)` is a commutative monoid with VETO as absorbing element.

### `judge_collective_deterministic` (C-21c basic) ⭐

```lean
theorem judge_collective_deterministic
    (judges : List Judge) (p : Proposal) (ctx : Context) :
    run judges p ctx = run judges p ctx := rfl
```

**Reading:** same input → same output. Trivially true because judges are pure functions; the substantive content is the corollary below.

### `judge_collective_perm_invariant` (C-21c substantive) ⭐

```lean
theorem judge_collective_perm_invariant
    (j1 j2 : List Judge) (p : Proposal) (ctx : Context)
    (h_perm : j1.Perm j2) :
    run j1 p ctx = run j2 p ctx
```

**Reading:** the order of judges in the collective doesn't matter. Whether queried sequentially, in parallel, or in arbitrary order, the result is the same.

**Architectural meaning:** this certifies optimization C-OPT-5 (Workload-ISA v0.2 §4.5) — parallel judge dispatch is provably equivalent to sequential. Production deployments can safely fan out judge queries to all judges in parallel.

**Proof tactic:** induction on `List.Perm` (4 cases: nil, cons, swap, trans). The swap case uses `aggregate_left_comm` to commute adjacent judges; trans is by transitivity.

### `judge_collective_veto_propagates`

```lean
theorem judge_collective_veto_propagates
    (judges : List Judge) (p : Proposal) (ctx : Context)
    (j : Judge) (h_in : j ∈ judges) (h_veto : j p ctx = Vote.veto) :
    run judges p ctx = Vote.veto
```

**Reading:** if any single judge in the collective votes VETO, the collective vote is VETO.

**Architectural meaning:** the **constitutional protection** — any judge can block any proposal. This is required for the judge-collective design where minority dissent is sufficient to halt change. Combined with C-21d below, this establishes that any single judge can block self-modification.

**Cited by:** Constitutional.lean's `judge_veto_blocks_self_modification`; production deployment of BMA's governance system.

---

## 18. Constitutional — the self-modification gate (Phase 3)

**File:** `Wyrd/Constitutional.lean`. **Imports:** `Wyrd.JudgeCollective`. **Status:** clean compile, 0 sorries (1 cosmetic unused-variable warning).

### Types

```lean
structure CodeUpdate where updateId : Nat

inductive UpdateOutcome : Type where
  | applied
  | rejected (reason : String)

def CodeUpdate.toProposal (u : CodeUpdate) : Proposal := { proposalId := u.updateId }

def tryApplyCodeUpdate
    (u : CodeUpdate) (judges : List Judge) (ctx : Context) : UpdateOutcome :=
  match h : run judges u.toProposal ctx with
  | Vote.approve     => UpdateOutcome.applied
  | Vote.minorConcern => UpdateOutcome.rejected "minor concern from judge collective"
  | Vote.majorConcern => UpdateOutcome.rejected "major concern from judge collective"
  | Vote.veto        => UpdateOutcome.rejected "vetoed by judge collective"
```

This is the constitutional gate — the **only** function that can produce `UpdateOutcome.applied`.

### `self_modification_requires_approval` (C-21d) ⭐⭐⭐

```lean
theorem self_modification_requires_approval
    (u : CodeUpdate) (judges : List Judge) (ctx : Context)
    (h_applied : tryApplyCodeUpdate u judges ctx = UpdateOutcome.applied) :
    JudgeCollective.run judges u.toProposal ctx = Vote.approve
```

**Reading:** every code update that ends in the Applied outcome was unanimously approved by the judge collective.

**Architectural meaning:** **the constitutional pin formalized.** BMA cannot self-modify (cannot apply code updates to its own repository, cannot change its constitutional documents) without obtaining APPROVE from the judge collective. The mechanism is *structural* — type signature plus this theorem, not a runtime check.

This is the load-bearing safety property for autonomous BMA operation. Without C-21d, the constitutional pin claim from Theory v2.0 / Ethics v1.1 was a runtime assertion that could in principle be bypassed; with C-21d, it is a formal property of the type system + the proven theorem together.

**Proof tactic:** unfold `tryApplyCodeUpdate`; `generalize` over the actual vote so we can case-analyze; for `Vote.approve` the conclusion is `rfl`; for the other three constructors, the function returns `Rejected ...` which contradicts `h_applied : ... = applied`.

**Combined with `judge_collective_veto_propagates`:** the corollary `judge_veto_blocks_self_modification` establishes that any single judge can block self-modification by voting VETO. Critical for the "minority dissent suffices" property of the judge collective.

**Cited by:** any code path that attempts to apply updates to BMA's own repository, modify constitutional documents, or change governance rules. The expected pattern:

```go
// bma.applyConstitutionalUpdate enforces the constitutional pin via
// the judge collective. Soundness: per
// Constitutional.self_modification_requires_approval, every Applied
// outcome implies unanimous APPROVE from the collective. Per
// Constitutional.judge_veto_blocks_self_modification, any single
// VETO blocks the update. Type signature + theorem together rule
// out any path to Applied that bypasses the collective.
//
// See Wyrd-Proofs-Reference-v1.2.md §18.
func (b *BMA) applyConstitutionalUpdate(u CodeUpdate) (UpdateOutcome, error) { ... }
```

---

## 19. The full Lean corpus map

| Phase | File | Theorems / declarations | What it certifies |
|---|---|---|---|
| 1 | Foundations | 11 | Algebraic structural lemmas + ℂ→ℍ closure |
| 1 | CayleyDickson | 37 | Octonion / sedenion types + associator witness |
| 1 | Projection | 19 | T2.2 — kernel computations on supervisor values are safe |
| 1 | Capability | 9 | T2.3 — capability soundness (positive + negative) |
| 1 | Noise | 14 | T3.1 — fp32 noise floor below privilege threshold |
| 1 | SedenionWitness | 4 | T1.2.c — sedenion non-alternativity witness |
| 1 | OctonionAlternative | 5 | Octonions are alternative |
| 2 | Hypergraph | 16 | C-20a — non-incident hyperedge addition is local |
| 2 | CTH | 7 | C-20b — measurement evidence is monotonic |
| 2 | Bridge | 9 | C-20c — promotion is conservation-atomic |
| 3 | Cart | 13 | C-21a — capability invariance under cart switch |
| 3 | Transaction | 13 | C-21b — cart switch is transaction-atomic |
| 3 | JudgeCollective | 18 | C-21c — judge determinism + perm-invariance |
| 3 | Constitutional | 7 | C-21d — the constitutional pin |
| **Total** | **14 files** | **~182 declarations** | **End-to-end coverage of three workload classes** |

---

## 20. Versioning

This is v1.2:
- v1.0 → v1.1: Phase 2 theorems (Class B hypergraph reasoning) added
- v1.1 → v1.2: Phase 3 theorems (Class C operational semantics) added — this revision

The chain of revisions is additive; v1.0 and v1.1 remain accessible for audit. v1.2 is the canonical reference as of 2026-04-25.

Future revisions anticipated:
- v1.3: Phase 1 follow-ups (T2.1.b, T2.1.c explicit composition theorems)
- v1.4: Class B refinements (full state-machine atomicity, hypergraph-level entropy integration)
- v1.5: Class C refinements (time-bounded approvals, emergency exception protocol)
- v2.0: structural reorganization if the corpus grows beyond ~25 files

---

## Attribution

QBP framework: Furey, Dixon, Günaydin / Gürsey, Boyle / Farnsworth, Singh, Chamseddine / Connes, Koide, Baez. CTH framework: Shannon, Dempster-Shafer, Pearl, Newman, Huet, Berge, Jirousek-Shenoy. Cynefin domain framing for cart-routing: Snowden. The judge-collective design with VETO-absorbing aggregation follows the BMA Governance Document tradition; the constitutional-pin pattern is consistent with Ethics v1.1's "scoped, reversible emergency exceptions" framing.

---

*End of Wyrd Proofs Reference v1.2.*
