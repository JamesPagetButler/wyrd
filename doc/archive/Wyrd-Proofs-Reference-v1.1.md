# Wyrd / Skuld Lean Proof Corpus — Theorem Reference

**Helpful Engineering — Quaternion-Based Physics Programme**
Principal Investigator: James Paget Butler
April 2026 — Rev 1.1 — supersedes v1.0

> **Purpose.** This document is the canonical reference for the Lean 4 proof corpus that backs the Wyrd / Skuld algebraic privilege model and Wyrd / CTH hypergraph reasoning. Every theorem in the corpus is listed with: (a) its formal statement, (b) the proof strategy used, (c) the architectural property it underwrites, and (d) where downstream code is expected to cite it. Use this when implementing `qbpcu`, `wyrd`, or `skuld` Go packages.

> **What changed in v1.1.** Added **Phase 2 — Class B hypergraph theorems** (§§12–14). Three new theorems closing the gaps identified in `Wyrd-Workload-ISA-v0.2.md` §3.8: hyperedge invariance under non-incident addition (C-20a), CTH measurement-evidence monotonicity (C-20b), and Bridge promotion conservation (C-20c). The Phase 1 algebraic-privilege content (§§1–11) is unchanged.

---

## 0. Build & toolchain (verified 2026-04-25)

| Component | Pin |
|---|---|
| Lean toolchain | `leanprover/lean4:v4.30.0-rc1` |
| Mathlib | rev `a090f46da78e9af11fee348cd7ee47bf8dd219d2` |
| Project root | `~/Documents/Wyrd/wyrd-lean-project/` |
| Build status | `lake build` succeeds; **0 sorries, 0 user-defined axioms** across all 10 files |
| File count | 10 (.lean files in `Wyrd/`) — 7 Phase 1 + 3 Phase 2 |

To rebuild from cold: `cd ~/Documents/Wyrd/wyrd-lean-project && lake update && lake exe cache get && lake build`.

---

## 1. Inventory at a glance

### Phase 1 — Algebraic privilege boundaries

| Theorem | File | Architectural role |
|---|---|---|
| `no_surjection_comm_to_noncomm` | Foundations | Abstract: ring map from CommRing onto non-commutative ring impossible |
| `no_surjection_assoc_to_nonassoc` | Foundations | Abstract: assoc → nonassoc impossible |
| `no_surjection_alt_to_nonalt` | Foundations | Abstract: alt → nonalt impossible |
| `commutator_eq_zero_of_comm` | Foundations | T1.2: commutator detector vanishes in commutative rings |
| `associator_eq_zero_of_assoc` | Foundations | T1.2: associator detector vanishes in associative rings |
| `alternator_eq_zero_of_alt` | Foundations | T1.2: alternator detector vanishes in alternative rings |
| `commutator_quaternion_witness` | Foundations | T1.2.a: ∃ a b ∈ ℍ with [a,b] ≠ 0 |
| `no_surjection_complex_to_quaternion` | Foundations | **T2.1.a: user (ℂ) → supervisor (ℍ) boundary closed** |
| `associator_octonion_witness` | CayleyDickson | T1.2.b: ∃ a b c ∈ 𝕆 with (ab)c ≠ a(bc) |
| `alternator_sedenion_witness` | SedenionWitness | T1.2.c: ∃ a b ∈ 𝕊 with (aa)b ≠ a(ab) |
| `octonion_alternative` | OctonionAlternative | 𝕆 is alternative (used for T2.1.c) |
| `quat_norm_is_real` | OctonionAlternative | q · star q lies in scalar subfield |
| `quat_real_part_is_real` | OctonionAlternative | q + star q lies in scalar subfield |
| `Projection.π_mul_of_inner` | Projection | **T2.2: outer-ring × inner-ring values, projected, equal inner-ring product** |
| `Projection.π_mul_ι` | Projection | T2.2 corollary on ι-embedded values |
| `Projection.kernel_supervisor_safe` | Projection | **Security headline: kernel computations on supervisor values, projected back, are safe** |
| `Capability.sandwich_preservation_associative` | Capability | Sandwich p·u·p⁻¹ behaves correctly under invertibility |
| `Capability.capability_grants_safe_access` | Capability | **T2.3 positive: capability holder can compute safely** |
| `Capability.no_capability_means_no_synthesis` | Capability | **T2.3 negative: no capability ⇒ no wider-ring synthesis** |
| `Capability.wider_capability_subsumes_narrower` | Capability | Capability projection: wider grants narrower |
| `Capability.hammer_capability_model` | Capability | Worked example for the Hammer simulation |
| `NoiseBound.abs_error_one_mul` | Noise | One-product fp32 error bound |
| `NoiseBound.abs_error_two_muls` | Noise | Two-product chain fp32 error bound |
| `NoiseBound.fp32_noise_unit_magnitude` | Noise | **T3.1: fp32 noise floor ~3·10⁻⁶ at unit magnitude** |
| `NoiseBound.fp32_noise_decimal_magnitude` | Noise | T3.1: fp32 noise floor ~3·10⁻³ at M=10 |

### Phase 2 — Class B hypergraph reasoning (NEW in v1.1)

| Theorem | File | Architectural role |
|---|---|---|
| `Hypergraph.hyperedge_preserves_incident_edges` | Hypergraph | **C-20a: incremental hypergraph updates have local effect** |
| `Hypergraph.invariant_under_nonincident_addition` | Hypergraph | Abstract corollary — any local property of a node is preserved under non-incident edge addition |
| `Hypergraph.hyperedge_preserves_incoming_edges` | Hypergraph | Specialization for incoming edges (incoming ⊆ incident) |
| `Hypergraph.hyperedge_preserves_outgoing_edges` | Hypergraph | Specialization for outgoing edges (outgoing ⊆ incident) |
| `CTH.cth_measurement_evidence_monotonic` | CTH | **C-20b: better evidence cannot raise η for a Tier-2 node** |
| `CTH.cth_zero_error_zero_entropy` | CTH | Boundary case: η(δ=0) = 0 |
| `Bridge.bridge_promote_preserves_count` | Bridge | **C-20c: Bridge promotion conserves total signal count** |
| `Bridge.bridge_promote_signal_in_cth` | Bridge | Post-promote, signal is in CTH queue |
| `Bridge.bridge_promote_signal_not_in_contextus` | Bridge | Post-promote, signal is not in Contextus queue |
| `Bridge.bridge_promote_exactly_one_side` | Bridge | Combined: signal in exactly one queue (no partial state) |

**Bold rows are the core security/correctness theorems.** Phase 1 establishes that privilege violations are structurally impossible. Phase 2 establishes that hypergraph state evolution is well-behaved: incremental updates have local effect, evidence monotonically tightens trust, and bridge promotions conserve signals.

---

## 2 through 11 — Phase 1 theorems

> Phase 1 sections (Foundations, CayleyDickson, Projection, Capability, Noise bound, SedenionWitness, OctonionAlternative, ring-tower closure, downstream API, versioning) are unchanged from v1.0. See `Wyrd-Proofs-Reference-v1.0.md` for those entries (or read directly from the `.lean` files; the section content has been verified equivalent).

For brevity, this v1.1 document focuses on the new Phase 2 content. The Phase 1 content remains in v1.0 as a permanent reference; v1.1 is additive, not replacement.

---

## 12. Hypergraph — Class B foundation (Phase 2)

**File:** `Wyrd/Hypergraph.lean`. **Imports:** `Mathlib.Data.Finset.Basic`, `Finset.Card`, `Finset.Insert`. **Status:** clean compile, 0 sorries.

### Types

```lean
structure HyperEdge (V : Type*) [DecidableEq V] where
  premises : Finset V
  target : V

structure Graph (V : Type*) [DecidableEq V] where
  vertices : Finset V
  edges : Finset (HyperEdge V)
```

A hyperedge has a finite premise set and a single target — matches CTH's notion of a directed hyperedge `e = (S_e, t_e)` per the theory paper Definition 1. The `Graph` carries vertex and edge sets explicitly. `DecidableEq V` is required throughout because Finset operations need it.

> **Implementation note.** `DecidableEq (HyperEdge V)` is declared explicitly (not via `deriving`) because the auto-deriving path triggers a typeclass-instance ambiguity in this Lean version. The manual instance is structurally identical.

### Operations

```lean
def HyperEdge.incident (e : HyperEdge V) (v : V) : Prop :=
  v ∈ e.premises ∨ v = e.target

def Graph.incidentEdges (G : Graph V) (v : V) : Finset (HyperEdge V) :=
  G.edges.filter (·.incident v)

def Graph.incomingEdges (G : Graph V) (v : V) : Finset (HyperEdge V) :=
  G.edges.filter (·.target = v)

def Graph.outgoingEdges (G : Graph V) (v : V) : Finset (HyperEdge V) :=
  G.edges.filter (v ∈ ·.premises)

def Graph.addEdge (G : Graph V) (e : HyperEdge V) : Graph V :=
  { vertices := G.vertices, edges := insert e G.edges }
```

The `addEdge` operation models the canonical Wyrd write: append a new hyperedge. Vertex set is left unchanged (well-formedness assumed handled upstream).

### `hyperedge_preserves_incident_edges` (C-20a) ⭐

```lean
theorem hyperedge_preserves_incident_edges
    (G : Graph V) (e : HyperEdge V) (v : V)
    (hv_p : v ∉ e.premises) (hv_t : v ≠ e.target) :
    (G.addEdge e).incidentEdges v = G.incidentEdges v
```

**Reading:** if v is neither a premise nor the target of e, the set of edges incident to v in `G.addEdge e` equals the set of edges incident to v in G.

**Proof tactic:** unfold `incidentEdges`; apply `Finset.filter_insert`; the conditional reduces because `v` is not incident to `e`.

**Architectural meaning:** **incremental updates to the Wyrd hypergraph have local effect.** When CTH adds a confluence-point edge connecting two existing chains, the local properties of unrelated nodes are unchanged. **Without this theorem, every insertion would require a full graph audit** — production CTH would be O(N²) per write. With this theorem, the audit is O(touched_nodes).

**Cited by:** `cth.evaluator.AddDerivation` (Class B implementation), Bridge promotion logic, any incremental Δ-recompute path.

### `invariant_under_nonincident_addition` (abstract corollary)

```lean
theorem invariant_under_nonincident_addition
    {α : Type*} (G : Graph V) (e : HyperEdge V) (v : V)
    (hv_p : v ∉ e.premises) (hv_t : v ≠ e.target)
    (f : Graph V → V → α)
    (h_local : ∀ G₁ G₂ : Graph V, G₁.incidentEdges v = G₂.incidentEdges v → f G₁ v = f G₂ v) :
    f (G.addEdge e) v = f G v
```

**Reading:** any function `f` that depends only on `incidentEdges v` is preserved under non-incident edge addition.

**Architectural meaning:** the *abstraction* of C-20a. CTH's η, μ-incoming, and any other locally-defined metric all instantiate this pattern by providing `h_local` (a proof that the metric is local).

### `hyperedge_preserves_incoming_edges` and `hyperedge_preserves_outgoing_edges`

Specializations of C-20a for the directional cases. Used when the metric depends specifically on incoming edges (e.g., η for Tier-2 derivations from a single confluence) or outgoing edges (e.g., support-set size for axiom anchors).

---

## 13. CTH — entropy + monotonicity (Phase 2)

**File:** `Wyrd/CTH.lean`. **Imports:** `Mathlib.Analysis.SpecialFunctions.Log.Basic`, `Wyrd.Hypergraph`. **Status:** clean compile, 0 sorries (2 cosmetic unused-variable warnings).

### Trust tiers

```lean
inductive TrustTier where
  | axiom        -- Tier 0
  | proof        -- Tier 1
  | measurement  -- Tier 2
  | prediction   -- Tier 3
  deriving DecidableEq, Repr
```

Matches CTH paper Definition 2 exactly.

### Tier-2 entropy

```lean
noncomputable def measurementEntropy (δ : ℝ) : ℝ := -Real.log (1 - δ)
```

**Domain:** δ ∈ [0, 1) for the formula to be well-defined (`Real.log` is finite for positive arguments). Theorems below carry the range as explicit hypotheses.

### `cth_measurement_evidence_monotonic` (C-20b) ⭐

```lean
theorem cth_measurement_evidence_monotonic
    (δ δ' : ℝ)
    (h_δ_lower : 0 ≤ δ) (h_δ_upper : δ < 1)
    (h_δ'_lower : 0 ≤ δ') (h_δ'_upper : δ' < 1)
    (h_evidence : δ' ≤ δ) :
    measurementEntropy δ' ≤ measurementEntropy δ
```

**Reading:** if a Tier-2 measurement node has fractional error δ, and consistent new evidence reduces the error to δ' ≤ δ, then the new entropy is at most the old entropy.

**Proof:** `1 - δ ≤ 1 - δ'` (from `h_evidence`); both positive (from upper bounds); `Real.log` is monotone increasing on positive reals (`log_monotone_on_positive`); negate both sides; `linarith` closes.

**Architectural meaning:** **better evidence cannot raise η for a Tier-2 node.** The framework's epistemic state is monotone in the evidence direction — adding consistent evidence is always at least as good. CTH's auditability claim depends on this property: an auditor reviewing a programme's evidence trail can rely on η decreasing (or staying constant) as more measurements arrive.

**Security interpretation:** the watchdog cannot be tricked into raising η — and weakening a Tier-2 node's trust standing — by an attacker submitting "evidence" that's a no-op. Adding any δ' ≤ δ is safe; it monotonically tightens the trust standing.

**Cited by:** `cth.evaluator.UpdateEvidence`, audit log validation, programme-Δ recompute.

### `cth_zero_error_zero_entropy`

```lean
theorem cth_zero_error_zero_entropy : measurementEntropy 0 = 0
```

Boundary case: a perfect measurement (δ = 0) has zero entropy. Provable by `Real.log_one` and `simp`. Useful as a sanity check and for bridging to Tier-1 (which always has η = 0).

### Phase 2 — what's deferred from CTH

The full CTH evaluator needs more theorems eventually (covered in Phase 2.5+):
- Hypergraph-level entropy: integrate `measurementEntropy` with `Wyrd.Hypergraph` by indexing it through a per-node tier function
- Programme deficit Δ(G) = Σ η over source anchors
- Mutual information I at confluence points
- Reynolds-analogue Re_e for incoherence detection

These are deferred. **C-20b alone closes the "evidence cannot raise η" gap that was blocking Class B production claims.**

---

## 14. Bridge — promotion atomicity (Phase 2)

**File:** `Wyrd/Bridge.lean`. **Imports:** `Mathlib.Data.Finset.Basic`, `Finset.Card`, `Finset.Insert`. **Status:** clean compile, 0 sorries.

### State model

```lean
structure State (Signal : Type*) [DecidableEq Signal] where
  contextusQueue : Finset Signal
  cthQueue : Finset Signal

def State.signalCount (b : State Signal) : ℕ :=
  b.contextusQueue.card + b.cthQueue.card

def State.promote (b : State Signal) (s : Signal)
    (h_in : s ∈ b.contextusQueue) (h_out : s ∉ b.cthQueue) : State Signal :=
  { contextusQueue := b.contextusQueue.erase s
    cthQueue := insert s b.cthQueue }
```

The `promote` operation is atomic by construction: it produces the post-state directly without any intermediate (partial) state being observable.

### `bridge_promote_preserves_count` (C-20c) ⭐

```lean
theorem bridge_promote_preserves_count (b : State Signal) (s : Signal)
    (h_in : s ∈ b.contextusQueue) (h_out : s ∉ b.cthQueue) :
    (b.promote s h_in h_out).signalCount = b.signalCount
```

**Reading:** Bridge promotion preserves the total signal count. Signals are neither created nor destroyed; they only move between queues.

**Proof tactic:** unfold `signalCount`; rewrite via `card_erase_of_mem` (cardinality decreases by 1) and `card_insert_of_notMem` (cardinality increases by 1); use `card_pos` to handle the natural-number subtraction; `omega` closes.

**Architectural meaning:** the **conservation form of atomicity**. An attacker (or bug) cannot use promotion to inflate the signal count (creating phantoms) or deflate it (silent drops). Combined with the precondition (`s ∈ contextusQueue`, `s ∉ cthQueue`), this rules out both Class B integrity-failure modes:

1. **Lost signal:** post-promote, signal would be in neither queue
2. **Duplicated signal:** post-promote, signal would be in both queues

Both are excluded by the type signature plus this theorem.

**Cited by:** `bridge.Promote()` implementation, NATS message-bus delivery semantics, audit trail validation.

### Corollaries — exactly-one-side property

```lean
theorem bridge_promote_signal_in_cth (...) : s ∈ (b.promote s ...).cthQueue
theorem bridge_promote_signal_not_in_contextus (...) : s ∉ (b.promote s ...).contextusQueue
theorem bridge_promote_exactly_one_side (...) :
    s ∈ (b.promote s ...).cthQueue ∧ s ∉ (b.promote s ...).contextusQueue
```

These give the "no partial state" claim in conservation form: post-promote the signal is in exactly one queue (the CTH side), never both, never neither.

### Scope note — what this does NOT cover

This is the **conservation form** of atomicity, sufficient for Class B integrity claims. It does NOT model:

- **Full state-machine atomicity** with explicit observers, transitions, and visibility (would require process-calculus or temporal-logic formalism). The conservation form is the substantive content; the state-machine form would add formal modeling of partial-state non-observability.
- **Multi-signal batched promotion.** Currently single-signal; batches are a downstream optimization (B-OPT-6 in Workload-ISA v0.2).
- **Failure / abort semantics.** Currently we model successful atomic promotion only. Aborts that revert state would extend with a complementary `abort` operation and a conservation-under-abort theorem.

These gaps are tracked in the Phase 3 Lean work (per Implementation Plan v1.0).

---

## 15. Phase 2 summary and Phase 3 outlook

**Phase 2 is complete** for the Class B gaps identified in `Wyrd-Workload-ISA-v0.2.md` §3.8:

| Gap | Theorem | Status |
|---|---|---|
| Graph-invariant preservation under hyperedge addition | `hyperedge_preserves_incident_edges` (C-20a) | ✅ proven |
| CTH metric monotonicity (Tier-2) | `cth_measurement_evidence_monotonic` (C-20b) | ✅ proven |
| Bridge atomicity (conservation form) | `bridge_promote_preserves_count` (C-20c) | ✅ proven |

**Phase 3 — Class C operational semantics** is still ahead (per Implementation Plan v1.0 §2.4 / C-21a-d):

| Gap | Status |
|---|---|
| `capability_invariant_under_cart_switch` (C-21a) | not started — depends on `Cart-as-Context` formal model (C-16) |
| `cart_switch_atomic` (C-21b) | not started — depends on Wyrd transaction model (C-17) |
| `judge_collective_deterministic` (C-21c) | not started — straightforward once judge model formalized |
| `self_modification_requires_approval` (C-21d) | not started — the constitutional pin formalization |

These are Phase 3 (Run-phase) work in the implementation plan. Class C is operationally more complex than Class B; the formal model requires state machines and observability, not just hypergraph algebra.

---

## 16. Downstream code — citation pattern (updated)

The citation pattern from v1.0 §10 extends to Phase 2 theorems. Skeleton:

```go
// cth.evaluator.AddDerivation appends a new derivation hyperedge to the
// trust hypergraph. Soundness: when this new edge is non-incident to an
// existing node v, v's local properties (η, μ-incoming) are preserved by
// Hypergraph.hyperedge_preserves_incident_edges. This is the formal
// foundation for incremental updates with bounded blast radius.
//
// See Wyrd-Proofs-Reference-v1.1.md §12.
func (e *Evaluator) AddDerivation(...) error { ... }
```

```go
// cth.evaluator.UpdateEvidence updates a Tier-2 measurement node's
// fractional error δ in response to new evidence. Soundness: if δ' ≤ δ,
// the post-update entropy η' ≤ η by CTH.cth_measurement_evidence_monotonic.
// This guarantees evidence is monotonically beneficial; the auditor can
// rely on η non-increasing across consistent updates.
//
// See Wyrd-Proofs-Reference-v1.1.md §13.
func (e *Evaluator) UpdateEvidence(node, newDelta) error { ... }
```

```go
// bridge.Promote moves an Insight Signal from the Contextus queue to
// the CTH evaluation queue. Soundness: signal count is preserved by
// Bridge.bridge_promote_preserves_count; the signal is in exactly one
// queue post-promotion by Bridge.bridge_promote_exactly_one_side. No
// partial state, no lost signals, no duplicates.
//
// See Wyrd-Proofs-Reference-v1.1.md §14.
func (b *Bridge) Promote(signal Signal) error { ... }
```

---

## 17. Versioning

This is v1.1, additive over v1.0 (Phase 1 content). Convention:
- v1.0 → v1.1: Phase 2 theorems added (this revision)
- v1.1 → v2.0: Phase 3 (Class C operational semantics) — anticipated 2026 H2 / 2027 H1
- v2.x: incremental additions and refinements

Older versions remain in the Archive for audit; the latest is canonical.

---

## Attribution

QBP framework: Furey, Dixon, Günaydin / Gürsey, Boyle / Farnsworth, Singh, Chamseddine / Connes, Koide, Baez. CTH framework: Shannon, Dempster-Shafer (with Pearl's critique), Newman / Huet on confluence, Berge on hypergraphs, Jirousek-Shenoy entropy. The Cayley-Dickson construction follows Schafer; Baez 2002, *The Octonions*, Bull. AMS.

---

*End of Wyrd Proofs Reference v1.1.*
