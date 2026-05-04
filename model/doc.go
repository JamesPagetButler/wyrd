// Package model defines the typed hypergraph data structures that make up
// a Wyrd graph: nodes, hyperedges, weights, and the algebraic-privilege
// tier system.
//
// The model deliberately mirrors the structures formalised in the Lean
// corpus (`lean/Wyrd/`):
//
//   - [Tier] corresponds to the four-tier Cayley-Dickson tower
//     (ℂ ⊂ ℍ ⊂ 𝕆 ⊂ 𝕊) proven closed in `lean/Wyrd/Foundations.lean`.
//   - [Hyperedge] of arity k ≥ 3 is irreducible to pair decompositions
//     per `lean/Wyrd/HolographicHypergraph.lean` and the higher-arity
//     generalisation in `HolographicHypergraphHigherArity.lean`.
//   - [Graph]'s incident-edge invariants follow `lean/Wyrd/Hypergraph.lean`
//     (C-20a: non-incident edge addition preserves a node's incident set).
//
// The model is pure-Go with stdlib-only dependencies. Storage and
// numeric backends live in sibling packages.
package model
