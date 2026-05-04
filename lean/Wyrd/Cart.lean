/-
  Wyrd/Cart.lean

  Class C Phase 3 — Cart-as-context formal model + capability invariance.

  Helpful Engineering — Quaternion-Based Physics Programme
  April 2026 — Rev 0.1

  ============================================================
  PURPOSE
  ============================================================

  BMA operates within Systema's horse-and-cart framework. Theory
  Cart and Engineering Cart are operating modes BMA switches
  between; the optimal pattern is rapid alternation. A capability
  issued to BMA at session start should remain valid across cart
  switches if it was issued with session scope (rather than
  cart-specific scope).

  This file proves:

    (C-21a) capability_invariant_under_cart_switch:
      A session-scoped capability remains valid in the new cart
      after any cart-switch operation. Equivalently: cart-switching
      does not invalidate session-level capabilities.

  ============================================================
  MODEL
  ============================================================

  - Cart enumerates the operating modes (Theory, Engineering,
    Beekeeper, plus extensible domain-specific carts represented
    abstractly via a Nat ID for non-canonical carts).
  - Scope distinguishes capability lifetime: Session (entire beekeeper
    session, survives cart switches), Cart (specific cart only),
    or Task (specific task within a cart, narrowest).
  - Capability bundles a token with its scope.
  - CartSession represents BMA's running state with a current cart
    and a set of held capabilities.
  - switchTo is the cart-switch operation: changes the current cart;
    capabilities are NOT invalidated by the switch (they persist as
    a list, though their per-cart validity depends on their scope).
-/

import Mathlib.Data.Finset.Basic

namespace Wyrd
namespace Cart

/- ============================================================
   PART 1 — Cart, Scope, Capability types
   ============================================================ -/

/-- The operating modes BMA can switch between. Theory and Engineering
    are the canonical pair; Beekeeper handles direct hardware control;
    domain-specific carts are addressed via numeric IDs. -/
inductive Cart : Type where
  | theory
  | engineering
  | beekeeper
  | domainSpecific (id : Nat)
  deriving DecidableEq, Repr

/-- The lifetime / scope of a capability.

    - `session` : valid for the entire beekeeper session, across all
      cart switches. The default for capabilities BMA needs continuously.
    - `cart c`  : valid only when the indicated cart is active. Used
      sparingly when a specific operating mode requires a privilege
      others should not have.
    - `task n`  : valid only for the duration of the indicated task.
      Narrowest; expires automatically when the task completes.
-/
inductive Scope : Type where
  | session
  | cart (c : Cart)
  | task (taskId : Nat)
  deriving DecidableEq, Repr

/-- A capability — a token with a scope. The `token` is opaque here;
    the substantive content for our theorems is the scope. -/
structure Capability where
  token : Nat
  scope : Scope
  deriving DecidableEq, Repr

/-- A capability is valid in a given cart depending on its scope:
    - session-scoped: always valid
    - cart-scoped:    valid only if the cart matches
    - task-scoped:    we conservatively model task-scoped as not valid
      across arbitrary carts; refinement requires task-tracking model
      out of scope here. -/
def Capability.validInCart (c : Capability) (cart : Cart) : Prop :=
  match c.scope with
  | Scope.session => True
  | Scope.cart c' => c' = cart
  | Scope.task _ => False

instance (c : Capability) (cart : Cart) : Decidable (c.validInCart cart) := by
  unfold Capability.validInCart
  cases c.scope <;> simp <;> infer_instance

/- ============================================================
   PART 2 — CartSession state and switching
   ============================================================ -/

/-- BMA's running state: which cart is active, and which capabilities
    are held. -/
structure CartSession where
  currentCart : Cart
  capabilities : Finset Capability

/-- The cart-switch operation: change the active cart; capabilities
    are preserved (the switch does NOT invalidate the capability
    list — per-cart validity is determined by scope, not by the switch). -/
def CartSession.switchTo (s : CartSession) (newCart : Cart) : CartSession :=
  { currentCart := newCart, capabilities := s.capabilities }

@[simp] theorem CartSession.switchTo_currentCart (s : CartSession) (newCart : Cart) :
    (s.switchTo newCart).currentCart = newCart := rfl

@[simp] theorem CartSession.switchTo_capabilities (s : CartSession) (newCart : Cart) :
    (s.switchTo newCart).capabilities = s.capabilities := rfl

/-- A session has a valid capability if it's in the held set AND
    it's valid in the currently active cart. -/
def CartSession.hasValidCapability (s : CartSession) (cap : Capability) : Prop :=
  cap ∈ s.capabilities ∧ cap.validInCart s.currentCart

/- ============================================================
   PART 3 — Theorem C-21a: capability invariance under cart switch
   ============================================================ -/

/-- KEY LEMMA: a session-scoped capability is valid in every cart. -/
theorem session_scoped_valid_in_all_carts (cap : Capability)
    (h_session : cap.scope = Scope.session) (cart : Cart) :
    cap.validInCart cart := by
  unfold Capability.validInCart
  rw [h_session]
  trivial

/-- THEOREM C-21a: a session-scoped capability that was held in
    a cart session remains held AND remains valid after any
    cart-switch operation.

    READING: switching from Theory to Engineering (or any cart pair)
    does not invalidate session-level capabilities. The beekeeper's
    session token granted at instantiation stays valid throughout
    the beekeeper-conversation lifetime regardless of how BMA
    alternates between carts.

    SECURITY / CORRECTNESS INTERPRETATION: Skuld does NOT need to
    re-issue capabilities on every cart switch. Cart switches are
    cheap; they're internal cognitive routing, not security
    boundaries. This is what makes the rapid Theory ↔ Engineering
    alternation pattern viable — the optimal pattern (per
    BMA-Theory-Consolidated v2.0 §8.6) is short loops, and short
    loops are only short if cart switching is free of expensive
    capability roundtrips.

    OPTIMIZATION CONNECTION: this theorem certifies C-OPT-3 from
    Workload-ISA v0.2 §4.5 ("Capability scope persistence across
    cart switches"). Without this proof, the C-OPT-3 claim was
    asserted; with it, it is demonstrably sound. -/
theorem capability_invariant_under_cart_switch
    (s : CartSession) (cap : Capability)
    (h_session : cap.scope = Scope.session)
    (h_held : cap ∈ s.capabilities)
    (newCart : Cart) :
    (s.switchTo newCart).hasValidCapability cap := by
  unfold CartSession.hasValidCapability
  refine ⟨?_, ?_⟩
  · -- capability is still held
    rw [CartSession.switchTo_capabilities]
    exact h_held
  · -- capability is valid in the new cart (because it's session-scoped)
    rw [CartSession.switchTo_currentCart]
    exact session_scoped_valid_in_all_carts cap h_session newCart

/-- COROLLARY: capability invariance composes — multiple cart switches
    in sequence do not invalidate session-scoped capabilities. -/
theorem capability_invariant_under_cart_switch_chain
    (s : CartSession) (cap : Capability)
    (h_session : cap.scope = Scope.session)
    (h_held : cap ∈ s.capabilities)
    (carts : List Cart) :
    let s' := carts.foldl CartSession.switchTo s
    s'.hasValidCapability cap := by
  induction carts generalizing s with
  | nil =>
    simp [List.foldl]
    refine ⟨h_held, ?_⟩
    exact session_scoped_valid_in_all_carts cap h_session s.currentCart
  | cons cart rest ih =>
    simp [List.foldl]
    have h_intermediate : (s.switchTo cart).hasValidCapability cap :=
      capability_invariant_under_cart_switch s cap h_session h_held cart
    have h_held' : cap ∈ (s.switchTo cart).capabilities := h_intermediate.1
    exact ih (s.switchTo cart) h_held'

/- ============================================================
   PART 4 — Status
   ============================================================

   PROVEN:
     ✓ Cart enumeration (theory, engineering, beekeeper, domain-specific)
     ✓ Scope enumeration (session, cart, task)
     ✓ Capability with token + scope
     ✓ Capability.validInCart (decidable)
     ✓ CartSession state with switchTo operation
     ✓ session_scoped_valid_in_all_carts (helper)
     ✓ C-21a: capability_invariant_under_cart_switch
     ✓ capability_invariant_under_cart_switch_chain (n-step composition)

   DEFERRED (not blocking):
     ◦ Task-scoped capability lifetimes (currently modeled as "not
       valid in arbitrary carts"; refinement requires per-task tracking)
     ◦ Capability revocation semantics (the held set never shrinks
       in our minimal model; revocation would extend with a removeCap
       operation and corresponding theorems)
     ◦ Cross-instance capability sharing (multi-BMA federation; Sprint phase)

   This file is COMPLETE for the C-21a deliverable. The minimal Cart
   model proven here is consistent with what the broader Cart-as-Context
   spec (C-16, James-direct) is expected to specify; if the full spec
   refines or extends this model, the theorems remain valid (or extend
   straightforwardly).
-/

end Cart
end Wyrd
