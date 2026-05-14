/-
  Wyrd.lean

  Top-level module for the Wyrd / Skuld algebraic privilege model formal proofs.
  Imports all submodules in dependency order.

  Helpful Engineering — Quaternion-Based Physics Programme
  Principal Investigator: James Paget Butler
  April 2026
-/

-- Phase 1 — algebraic privilege boundaries
import Wyrd.CayleyDickson
import Wyrd.Foundations
import Wyrd.Projection
import Wyrd.Capability
import Wyrd.Noise
import Wyrd.SedenionWitness
import Wyrd.OctonionAlternative

-- Phase 2 — Class B hypergraph reasoning (CTH / Bridge)
import Wyrd.Hypergraph
import Wyrd.CTH
import Wyrd.Bridge

-- Phase 3 — Class C operational semantics (cart, transactions, judges, constitutional pin)
import Wyrd.Cart
import Wyrd.Transaction
import Wyrd.JudgeCollective
import Wyrd.Constitutional

-- Phase 4 — physical instantiation (PROT-HH-001 holographic hypergraph)
import Wyrd.HolographicHypergraph
import Wyrd.HolographicHypergraphQuaternion
import Wyrd.HolographicHypergraphHigherArity

-- Phase 4 (CTH lift) — NaryMI synergy positivity
import Wyrd.NaryMI

-- Phase 2 (extension) — W-Toddle-1 tier-immunity soundness anchor
import Wyrd.TierImmunity

-- Phase 2 (extension) — scope-loader atomicity soundness anchor
import Wyrd.ScopeLoader
