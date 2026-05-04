/-
  Wyrd/Hypergraph.lean

  Class B Phase 2 — Lean foundation for Wyrd / CTH hypergraph reasoning.

  Helpful Engineering — Quaternion-Based Physics Programme
  April 2026 — Rev 0.1

  ============================================================
  PURPOSE
  ============================================================

  Wyrd / CTH hypergraph reasoning (Class B in the workload taxonomy)
  rests on properties of hypergraphs that are not yet covered by
  the Phase 1 algebraic-privilege corpus. This file establishes the
  minimal hypergraph model needed and proves:

    (C-20a) hyperedge_preserves_node_invariants:
      adding an edge e to a hypergraph G does not change any
      property of node v that depends only on edges incident
      to v, provided v is not a premise nor the target of e.

  Companion files:
    - Wyrd/CTH.lean      — entropy + measurement monotonicity (C-20b)
    - Wyrd/Bridge.lean   — atomic promotion (C-20c)

  ============================================================
  MODEL
  ============================================================

  A hypergraph has vertices (anchors in CTH parlance) and hyperedges
  (derivations in CTH parlance). Each hyperedge has a finite premise
  set S ⊆ V and a single target t ∈ V, written e = (S, t).

  Edges are stored as a Finset of HyperEdge values. Vertices are
  stored as a Finset. The model matches the real Wyrd / MuninnDB:
  finite, decidably-equal nodes and edges.
-/

import Mathlib.Data.Finset.Basic
import Mathlib.Data.Finset.Card
import Mathlib.Data.Finset.Insert

namespace Wyrd
namespace Hypergraph

/-- A hyperedge with a finite premise set and a single target.
    Matches CTH's notion of a directed hyperedge `e = (S_e, t_e)`. -/
structure HyperEdge (V : Type*) [DecidableEq V] where
  premises : Finset V
  target : V

instance {V : Type*} [DecidableEq V] : DecidableEq (HyperEdge V) := fun a b =>
  if h_p : a.premises = b.premises then
    if h_t : a.target = b.target then
      isTrue (by cases a; cases b; simp_all)
    else
      isFalse (fun h => h_t (by rw [h]))
  else
    isFalse (fun h => h_p (by rw [h]))

/-- A hypergraph: finite vertex set, finite hyperedge set. -/
structure Graph (V : Type*) [DecidableEq V] where
  vertices : Finset V
  edges : Finset (HyperEdge V)

variable {V : Type*} [DecidableEq V]

/- ============================================================
   PART 1 — Incidence and incident-edge sets
   ============================================================ -/

/-- A node v is incident to an edge e if v is a premise of e or its target. -/
def HyperEdge.incident (e : HyperEdge V) (v : V) : Prop :=
  v ∈ e.premises ∨ v = e.target

instance HyperEdge.decidableIncident (e : HyperEdge V) (v : V) : Decidable (e.incident v) :=
  inferInstanceAs (Decidable (v ∈ e.premises ∨ v = e.target))

/-- The set of edges in G that are incident to v. -/
def Graph.incidentEdges (G : Graph V) (v : V) : Finset (HyperEdge V) :=
  G.edges.filter (·.incident v)

/-- The set of edges in G whose target is v (incoming edges to v). -/
def Graph.incomingEdges (G : Graph V) (v : V) : Finset (HyperEdge V) :=
  G.edges.filter (·.target = v)

/-- The set of edges in G that have v as a premise (outgoing edges from v). -/
def Graph.outgoingEdges (G : Graph V) (v : V) : Finset (HyperEdge V) :=
  G.edges.filter (v ∈ ·.premises)

/- ============================================================
   PART 2 — Edge addition
   ============================================================ -/

/-- Add an edge to a hypergraph. The vertex set is left unchanged
    (the edge's endpoints must already exist for a well-formed graph,
    but we don't enforce this here — out of scope for the invariance
    theorem). -/
def Graph.addEdge (G : Graph V) (e : HyperEdge V) : Graph V :=
  { vertices := G.vertices, edges := insert e G.edges }

@[simp] theorem Graph.addEdge_edges (G : Graph V) (e : HyperEdge V) :
    (G.addEdge e).edges = insert e G.edges := rfl

@[simp] theorem Graph.addEdge_vertices (G : Graph V) (e : HyperEdge V) :
    (G.addEdge e).vertices = G.vertices := rfl

/- ============================================================
   PART 3 — Theorem C-20a: hyperedge preserves node invariants
   ============================================================ -/

/-- KEY LEMMA: if v is not a premise of e and not the target of e,
    then v is not incident to e. -/
theorem not_incident_of_not_in_edge (e : HyperEdge V) (v : V)
    (hv_p : v ∉ e.premises) (hv_t : v ≠ e.target) : ¬ e.incident v := by
  intro h
  cases h with
  | inl h_p => exact hv_p h_p
  | inr h_t => exact hv_t h_t

/-- THEOREM C-20a: adding an edge e to G preserves the incident-edge set
    of every node v that is neither a premise nor the target of e.

    SECURITY / CORRECTNESS INTERPRETATION: in CTH, when a confluence-point
    edge is added connecting two existing chains, the local properties
    of unrelated nodes (their η entropy, their incoming-edge set, etc.)
    are unchanged. This is the formal foundation for "incremental updates
    to the hypergraph have local effect." Without this theorem, every
    insertion would require a full graph audit. -/
theorem hyperedge_preserves_incident_edges
    (G : Graph V) (e : HyperEdge V) (v : V)
    (hv_p : v ∉ e.premises) (hv_t : v ≠ e.target) :
    (G.addEdge e).incidentEdges v = G.incidentEdges v := by
  unfold Graph.incidentEdges
  rw [Graph.addEdge_edges]
  rw [Finset.filter_insert]
  -- The conditional `if e.incident v then insert e else id` reduces
  -- to `id` because v is not incident to e.
  have h_not_incident : ¬ e.incident v := not_incident_of_not_in_edge e v hv_p hv_t
  simp [h_not_incident]

/-- COROLLARY (the form most useful for CTH): any property of node v
    that is computed from G.incidentEdges v alone is preserved when
    a non-incident edge is added.

    This is the "abstract invariance" theorem. CTH's η, μ-incoming,
    and any other locally-defined metric all instantiate this pattern. -/
theorem invariant_under_nonincident_addition
    {α : Type*} (G : Graph V) (e : HyperEdge V) (v : V)
    (hv_p : v ∉ e.premises) (hv_t : v ≠ e.target)
    (f : Graph V → V → α)
    (h_local : ∀ G₁ G₂ : Graph V, G₁.incidentEdges v = G₂.incidentEdges v → f G₁ v = f G₂ v) :
    f (G.addEdge e) v = f G v := by
  apply h_local
  exact hyperedge_preserves_incident_edges G e v hv_p hv_t

/-- COROLLARY: incoming-edge set is preserved (since incoming ⊆ incident). -/
theorem hyperedge_preserves_incoming_edges
    (G : Graph V) (e : HyperEdge V) (v : V) (hv_t : v ≠ e.target) :
    (G.addEdge e).incomingEdges v = G.incomingEdges v := by
  unfold Graph.incomingEdges
  rw [Graph.addEdge_edges]
  rw [Finset.filter_insert]
  have h_target_ne : e.target ≠ v := fun h => hv_t h.symm
  simp [h_target_ne]

/-- COROLLARY: outgoing-edge set is preserved. -/
theorem hyperedge_preserves_outgoing_edges
    (G : Graph V) (e : HyperEdge V) (v : V) (hv_p : v ∉ e.premises) :
    (G.addEdge e).outgoingEdges v = G.outgoingEdges v := by
  unfold Graph.outgoingEdges
  rw [Graph.addEdge_edges]
  rw [Finset.filter_insert]
  simp [hv_p]

/- ============================================================
   PART 4 — Status
   ============================================================

   PROVEN:
     ✓ Basic Hypergraph types (HyperEdge, Graph) with incident, incoming,
       outgoing relations
     ✓ Graph.addEdge with simp lemmas for vertices and edges
     ✓ not_incident_of_not_in_edge — local lemma
     ✓ C-20a: hyperedge_preserves_incident_edges
     ✓ invariant_under_nonincident_addition (abstract corollary)
     ✓ hyperedge_preserves_incoming_edges
     ✓ hyperedge_preserves_outgoing_edges

   PHASE 2 NEXT:
     ◦ Define η entropy in Wyrd/CTH.lean using incident-edge sets
     ◦ Apply invariant_under_nonincident_addition to η
       → cth_eta_invariant_under_nonincident_addition

   This file is COMPLETE for the C-20a deliverable.
-/

end Hypergraph
end Wyrd
