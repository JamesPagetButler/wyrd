/-
  Wyrd/TierImmunity.lean

  Class B Phase 2 (extension) — Lean soundness anchor for the W-Toddle-1
  tier-immunity substrate primitives.

  Helpful Engineering — Quaternion-Based Physics Programme
  May 2026 — Rev 0.1

  ============================================================
  PURPOSE
  ============================================================

  W-Toddle-1 (Wyrd PR #39 design + PR #42 impl, both merged
  2026-05-14) adds `Node.TierImmune bool` to the Wyrd model. The
  semantic claim made by the design (PR #39 §4):

    "Adding or evicting other nodes does not change the membership
     of {v : v.TierImmune} in the graph."

  This file formalises that claim via a minimal eviction model and
  proves:

    (1) eviction_immune_blind — the effective domain of any
        eviction operation excludes immune nodes; immune nodes are
        observed-but-not-touched, structurally.

    (2) tier_immune_node_preserves_eviction — an immune node stays
        in the graph after any EvictionOp.

  The proof structure mirrors the C-20a reduction pattern from
  `Wyrd/Hypergraph.lean` and the v0.2 oriented-edge extension
  (`Wyrd/HypergraphOriented.lean`, forthcoming): construct a
  supporting lemma that names the "observed-but-not-touched"
  invariant, then reduce the main theorem to a one-step Finset
  arithmetic on it.

  ============================================================
  MODEL
  ============================================================

  We extend the existing `Wyrd.Hypergraph.Graph` model with an
  immunity predicate carried as a `Finset V` (the set of immune
  vertex identifiers). An EvictionOp is a finite set of candidate
  victims; the eviction applies the difference set-minus the
  immune set, so immune nodes are never removed.

  At this Lean abstraction layer the model is intentionally
  generic — no Wyrd-specific NodeID type, no Salience modulation
  (Salience is a soft retention-priority signal, not a structural
  invariant, per PR #39 §4.4). Salience modulation will land its
  own Lean theorem when the W-Toddle-2 eviction-execution
  primitive is specified.
-/

import Wyrd.Hypergraph
import Mathlib.Data.Finset.Basic
import Mathlib.Data.Finset.SDiff

namespace Wyrd
namespace TierImmunity

open Wyrd.Hypergraph

variable {V : Type*} [DecidableEq V]

/-- A graph extended with a per-vertex immunity predicate. The
    `immune` Finset is the set of vertices that cannot be evicted
    regardless of eviction-policy pressure.

    Models `model.Node.TierImmune == true` in the Wyrd Go runtime
    (Wyrd PR #42 commit `4e20abb`). -/
structure ImmuneGraph (V : Type*) [DecidableEq V] extends Graph V where
  immune : Finset V

/-- An eviction operation: a finite set of vertices that the policy
    layer (sleep-cycle, cap-per-retention-tier saturation) has marked
    as candidates for removal.

    Note that `toRemove` is the candidate set; the effective removal
    set is `toRemove \ immune`, computed by `applyEviction`. -/
structure EvictionOp (V : Type*) [DecidableEq V] where
  toRemove : Finset V

/-- Apply an eviction. The vertex set has the candidate-minus-immune
    set removed; the edge set is left unmodified at this abstraction
    layer (edge pruning is a downstream W-Toddle-2 concern, not part
    of the immunity-preservation invariant). The immune set itself is
    invariant under eviction. -/
def applyEviction (G : ImmuneGraph V) (op : EvictionOp V) : ImmuneGraph V :=
  { vertices := G.vertices \ (op.toRemove \ G.immune)
    edges := G.edges
    immune := G.immune }

@[simp] theorem applyEviction_immune (G : ImmuneGraph V) (op : EvictionOp V) :
    (applyEviction G op).immune = G.immune := rfl

@[simp] theorem applyEviction_edges (G : ImmuneGraph V) (op : EvictionOp V) :
    (applyEviction G op).edges = G.edges := rfl

@[simp] theorem applyEviction_vertices (G : ImmuneGraph V) (op : EvictionOp V) :
    (applyEviction G op).vertices = G.vertices \ (op.toRemove \ G.immune) := rfl

/- ============================================================
   PART 1 — Supporting lemma: eviction_immune_blind
   ============================================================ -/

/-- LEMMA: an immune vertex is excluded from the effective removal set
    `toRemove \ immune` by construction. -/
theorem immune_not_in_effective_removal
    (G : ImmuneGraph V) (op : EvictionOp V) (v : V)
    (h_immune : v ∈ G.immune) : v ∉ op.toRemove \ G.immune := by
  intro h_in
  rw [Finset.mem_sdiff] at h_in
  exact h_in.2 h_immune

/-- KEY LEMMA — eviction_immune_blind: applying an eviction leaves an
    immune vertex's membership in the graph unchanged. The eviction
    operation is "blind" to immune vertices by construction.

    Plain English: any operation that removes vertices from the graph
    by computing `vertices \ (toRemove \ immune)` is structurally
    incapable of removing immune vertices, regardless of what's in
    `toRemove`. The immune set is observed-but-not-touched. -/
theorem eviction_immune_blind
    (G : ImmuneGraph V) (op : EvictionOp V) (v : V)
    (h_immune : v ∈ G.immune) :
    v ∈ (applyEviction G op).vertices ↔ v ∈ G.vertices := by
  rw [applyEviction_vertices]
  constructor
  · intro h
    exact Finset.mem_sdiff.mp h |>.1
  · intro h
    rw [Finset.mem_sdiff]
    exact ⟨h, immune_not_in_effective_removal G op v h_immune⟩

/- ============================================================
   PART 2 — Main theorem: tier_immune_node_preserves_eviction
   ============================================================ -/

/-- A vertex v is in the graph G if it's in the vertex set. -/
def nodeInGraph (G : ImmuneGraph V) (v : V) : Prop := v ∈ G.vertices

/-- THEOREM (PR #39 §4 commitment): tier_immune_node_preserves_eviction.

    If a vertex v is immune AND in the graph, then after applying any
    EvictionOp, v is still immune AND still in the graph.

    This is the formal version of the W-Toddle-1 design claim
    (PR #39 §4): "Adding or evicting other nodes does not change the
    membership of {v : v.TierImmune}."

    SECURITY / CORRECTNESS INTERPRETATION (PR #39 §1): consumers
    setting `Node.TierImmune = true` on NT_SEED (BMA seed protocol
    Step 9), NT_LIFE_CERTIFICATE / NT_DEATH_CERTIFICATE (lineage
    protocol invariants), NT_PARAM_TRUST_STATE (audit-trail
    invariants), and similar foundation anchors are guaranteed that
    no automatic eviction (cap-per-retention-tier saturation, sleep-
    cycle compaction) can drop those nodes — the substrate enforces
    permanence structurally, not via a runtime convention. -/
theorem tier_immune_node_preserves_eviction
    (G : ImmuneGraph V) (op : EvictionOp V) (v : V)
    (h_immune : v ∈ G.immune) (h_in : v ∈ G.vertices) :
    v ∈ (applyEviction G op).immune ∧
    v ∈ (applyEviction G op).vertices := by
  refine ⟨?_, ?_⟩
  · -- Immunity is invariant under eviction (applyEviction_immune simp).
    simp [h_immune]
  · -- Membership reduces to eviction_immune_blind plus h_in.
    exact (eviction_immune_blind G op v h_immune).mpr h_in

/- ============================================================
   PART 3 — Corollaries useful for downstream consumers
   ============================================================ -/

/-- COROLLARY: an immune vertex survives any sequence of EvictionOps. -/
theorem tier_immune_preserved_under_eviction_sequence
    (G : ImmuneGraph V) (ops : List (EvictionOp V)) (v : V)
    (h_immune : v ∈ G.immune) (h_in : v ∈ G.vertices) :
    let G' := ops.foldl applyEviction G
    v ∈ G'.immune ∧ v ∈ G'.vertices := by
  induction ops generalizing G with
  | nil => exact ⟨h_immune, h_in⟩
  | cons op rest ih =>
    have h_step := tier_immune_node_preserves_eviction G op v h_immune h_in
    exact ih (applyEviction G op) h_step.1 h_step.2

/-- COROLLARY (form most useful for W-Toddle-2 eviction-execution):
    any property of the graph that depends only on the immune set is
    preserved under eviction. -/
theorem invariant_under_eviction_on_immune
    {α : Type*} (G : ImmuneGraph V) (op : EvictionOp V)
    (f : ImmuneGraph V → α)
    (h_immune_only : ∀ G₁ G₂ : ImmuneGraph V, G₁.immune = G₂.immune → f G₁ = f G₂) :
    f (applyEviction G op) = f G := by
  apply h_immune_only
  exact applyEviction_immune G op

/- ============================================================
   PART 4 — Status
   ============================================================

   PROVEN:
     ✓ ImmuneGraph extending Hypergraph.Graph with an immune Finset
     ✓ EvictionOp + applyEviction with simp lemmas
     ✓ immune_not_in_effective_removal — local Finset lemma
     ✓ eviction_immune_blind — key supporting lemma
     ✓ tier_immune_node_preserves_eviction — PR #39 §4 main claim
     ✓ tier_immune_preserved_under_eviction_sequence (List corollary)
     ✓ invariant_under_eviction_on_immune (abstract corollary)

   No sorry, no user-defined axiom — Phase 2 CI gate compliant.

   FOLLOW-ON SCOPE:
     ◦ W-Toddle-2 eviction-execution primitive (Wyrd-side Go impl
       that walks the graph and applies an EvictionOp per Contextus
       Spec v1.3 §5.4 cap-per-RetentionTier policy)
     ◦ Salience modulation theorems (when the salience-ordered
       eviction rule lands; soft retention-priority signal, not a
       structural invariant)
     ◦ Edge pruning under eviction (this file deliberately leaves
       edges unmodified — downstream W-Toddle-2 concern)

   PR #39 §4.4 estimate was ~35 LOC; the actual file is ~140 LOC
   including comments and corollaries (the theorem body itself is
   under 30 LOC).
-/

end TierImmunity
end Wyrd
