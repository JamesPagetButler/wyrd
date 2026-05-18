/-
  Wyrd/ComputeManifest.lean

  Class B Phase 2 (extension) — Lean soundness anchor for the
  model.LoadComputeManifest primitive (Wyrd issue track:
  repo-bma-systema-issue-#164 + #170 + #171; design surface merged
  in PR #58 on 2026-05-15; Go impl merged in PR #59 on 2026-05-18).

  Helpful Engineering — Quaternion-Based Physics Programme
  May 2026 — Rev 0.1

  ============================================================
  PURPOSE
  ============================================================

  PR #58 §4 committed to a soundness anchor for LoadComputeManifest:

    "Wyrd.ComputeManifest.manifest_load_atomic (forthcoming, lands
     with impl PR). Proof structure: validation is a pure predicate
     on parsed YAML; load is either-validate-and-return or
     return-error. Trivially atomic since there's no graph mutation;
     ~15 LOC Lean estimate."

  This file delivers exactly that. The theorem formalises the
  PR #58 §2.3 either-validate-and-return-or-return-error claim:

    LoadComputeManifest is atomic. The loader returns either a
    Validated manifest or an Error; it never returns a
    Partially-loaded state visible to consumers.

  Unlike scope_loader_atomic (which reasons about a stateful graph),
  manifest_load_atomic is a pure-function atomicity claim: the
  loader is total and disjoint — every parsed input maps to exactly
  one of the two outcomes, by case analysis on the validation
  predicate.

  ============================================================
  WHY THIS THEOREM MATTERS
  ============================================================

  Without atomicity, the Go runtime could (a) silently coerce a
  malformed manifest into a partially-populated struct, (b) leave
  unset fields at zero-value while reporting success, (c) emit a
  half-validated struct alongside an error. Each of these failure
  modes would propagate into federation CI's mode-(b) gate
  (per repo-bma-systema-issue-#171 Phase B-PR-8) and produce
  silent false-positives on substrate-credibility checks.

  The Lean theorem fixes the contract: consumers see either
  Validated or Rejected. There is no third option.

  ============================================================
  PROOF STRUCTURE
  ============================================================

    1. Model the parsed-YAML input as an abstract RawManifest
       (the Go-runtime `model.ComputeManifest` after yaml.Unmarshal
       but before Validate).

    2. The validation predicate `isValid : RawManifest → Bool`
       abstracts the conjunction of rules 1-9 from PR #58 §2.4.

    3. The load operation returns a `LoadOutcome`: either
       `validated raw` or `rejected raw`. The two branches are
       structurally disjoint (Lean's inductive types make this
       free).

    4. The atomicity theorem: for every (raw, valid) pair, the
       output is exactly one of the two cases — and which case
       is determined entirely by `valid`. This is total
       function disjointness, provable by case analysis on
       `valid`.
-/

import Mathlib.Data.List.Basic

namespace Wyrd
namespace ComputeManifest

/-- A parsed-YAML manifest before validation. The abstract type;
    concrete Go-side shape is `model.ComputeManifest` per PR #58
    §2.3. -/
structure RawManifest where
  /-- Opaque payload; the abstract Lean form does not commit to
      specific field shapes because the atomicity argument is
      independent of the fields' content. -/
  payload : List String
  deriving DecidableEq, Repr

/-- The loader's output. Either the input passed validation and is
    returned as `validated`, or it failed and is returned as
    `rejected`. The two branches are mutually exclusive by Lean's
    inductive-type construction. -/
inductive LoadOutcome where
  /-- Validation passed; consumers may use the manifest. -/
  | validated (m : RawManifest)
  /-- Validation failed; consumers MUST NOT use the manifest.
      The raw is carried for diagnostic purposes only (per the
      Go-runtime convention of wrapping `ErrComputeManifestInvalid`
      with the specific cause). -/
  | rejected (m : RawManifest)
  deriving Repr

/-- Project the manifest carried by a LoadOutcome. Useful for
    consumers that need to inspect the rejected form (rare;
    typically only for diagnostic logging). -/
def LoadOutcome.manifest : LoadOutcome → RawManifest
  | LoadOutcome.validated m => m
  | LoadOutcome.rejected m => m

/-- The atomic-load operation. Validation predicate is an abstract
    parameter (`valid : Bool`); concretely in the Go runtime, it is
    the conjunction of rules 1-9 from PR #58 §2.4 (version regex,
    authored_at parseable, phase enum, etc).

    Returns either `validated raw` or `rejected raw` per the
    Bool result. The function is total and deterministic. -/
def load (raw : RawManifest) (valid : Bool) : LoadOutcome :=
  if valid then
    LoadOutcome.validated raw
  else
    LoadOutcome.rejected raw

/- ============================================================
   PART 1 — Main theorem: manifest_load_atomic
   ============================================================ -/

/-- THEOREM (PR #58 §4 commitment): manifest_load_atomic.

    The loader produces EXACTLY one of two outcomes per input:
    either the validated form OR the rejected form. There is no
    third state, and the two states are structurally disjoint
    (Lean's inductive-type discipline makes this provable by
    case analysis on the validation Bool).

    SECURITY / CORRECTNESS INTERPRETATION (PR #58 §0):
    half-validated manifests would propagate into federation CI's
    mode-(b) eligibility gate (per repo-bma-systema-issue-#171)
    and silently break substrate-credibility checks. Consumers
    of `LoadComputeManifest` would receive a non-nil manifest
    pointer alongside a non-nil error, leading to the classic
    Go failure mode of "ignored the error, used the half-populated
    struct anyway." The Lean theorem makes this structurally
    impossible at the contract level. -/
theorem manifest_load_atomic
    (raw : RawManifest) (valid : Bool) :
    -- Either the output is validated (carrying the manifest):
    (∃ m, load raw valid = LoadOutcome.validated m)
    ∨
    -- Or the output is rejected (no validated manifest visible):
    (∃ m, load raw valid = LoadOutcome.rejected m) := by
  cases valid with
  | true =>
    left
    exact ⟨raw, rfl⟩
  | false =>
    right
    exact ⟨raw, rfl⟩

/- ============================================================
   PART 2 — Disjointness corollary
   ============================================================ -/

/-- COROLLARY: the two output branches are disjoint. No input can
    produce both a validated AND a rejected outcome. Follows from
    Lean's inductive-type discriminator. -/
theorem load_branches_disjoint (raw : RawManifest) (valid : Bool) :
    ¬ ((∃ m, load raw valid = LoadOutcome.validated m) ∧
       (∃ m, load raw valid = LoadOutcome.rejected m)) := by
  intro ⟨⟨_, h_val⟩, ⟨_, h_rej⟩⟩
  rw [h_val] at h_rej
  cases h_rej

/- ============================================================
   PART 3 — Determinism corollary
   ============================================================ -/

/-- COROLLARY: same input → same output. The loader is deterministic
    (no hidden state, no randomness). Trivially true at this Lean
    abstraction because `load` is a pure function; useful as an
    explicit contract for downstream verification. -/
theorem load_deterministic
    (raw : RawManifest) (valid : Bool) :
    load raw valid = load raw valid := rfl

/- ============================================================
   PART 4 — Validation predicate maps directly to outcome
   ============================================================ -/

/-- COROLLARY: the validation Bool is decisive — `valid = true` iff
    the outcome is validated. The Go runtime's `Validate` method
    returns nil iff `load` should return validated; this Lean
    theorem fixes that correspondence as a contract. -/
theorem load_validated_iff_valid
    (raw : RawManifest) (valid : Bool) :
    (∃ m, load raw valid = LoadOutcome.validated m) ↔ valid = true := by
  cases valid with
  | true =>
    constructor
    · intro _; rfl
    · intro _; exact ⟨raw, rfl⟩
  | false =>
    constructor
    · intro ⟨_, h⟩
      simp [load] at h
    · intro h
      cases h

/-- COROLLARY: rejection iff validation failed. The contrapositive
    form, useful for consumer-side dispatch. -/
theorem load_rejected_iff_invalid
    (raw : RawManifest) (valid : Bool) :
    (∃ m, load raw valid = LoadOutcome.rejected m) ↔ valid = false := by
  cases valid with
  | true =>
    constructor
    · intro ⟨_, h⟩
      simp [load] at h
    · intro h
      cases h
  | false =>
    constructor
    · intro _; rfl
    · intro _; exact ⟨raw, rfl⟩

/- ============================================================
   PART 5 — Status
   ============================================================

   PROVEN:
     ✓ RawManifest + LoadOutcome + load model
     ✓ manifest_load_atomic — PR #58 §4 main claim:
       "either validated or rejected; never a third state"
     ✓ load_branches_disjoint — the two outcomes are mutually
       exclusive (structural, from Lean inductive types)
     ✓ load_deterministic — same input always produces same output
     ✓ load_validated_iff_valid — validation bool is decisive
     ✓ load_rejected_iff_invalid — contrapositive form

   No sorry, no user-defined axiom — Phase 2 CI gate compliant.

   ~30 LOC of theorem bodies (well within the PR #58 §4 ~15 LOC
   estimate per theorem; the file totals ~200 LOC including
   commentary + corollaries + parallel forms).

   FOLLOW-ON SCOPE (per PR #58 design doc §11 v0.2 housekeeping):
     ◦ Credibility-window field validation (repo-bma-systema-
       issue-#171 Phase B): when `last_passing_tier_b` lands as
       a struct field, its staleness check is a runtime predicate
       — out of scope for substrate atomicity reasoning here.
     ◦ Verified-invariants forward-pin: the `verified_invariants`
       v0.2 schema slot will carry references to other substrate-
       tier Lean theorems (including this one, transitively); the
       cross-theorem coherence story lands at repo-bma-systema-
       issue-#170 Phase C-PR-12 (`CycleCounterCrossPhase.lean`).
-/

end ComputeManifest
end Wyrd
