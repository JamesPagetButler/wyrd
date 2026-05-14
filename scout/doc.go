// Package scout implements the BMA Theory Addendum 18 §6 reasoning
// primitive. ScoutQuery dispatches a focal-cone query that returns
// Active Agent intersections of the predicted source→sink path within
// a Locale Volume.
//
// v0.1 ships the API + types + a Crawl-shippable PLACEHOLDER BODY
// returning uniform AbsorptionGain. The real Absorption Gain
// computation (A18 §5 Locale-Bounded Absorption Estimation) lands
// when the dependency chain converges: query/ traversal +
// oriented-hyperedge schema + compute/laplacian.go body.
//
// Federation vocabulary lock (D9, governance-binding per addendum-18-
// walk seq=6):
//   - NT_SCOPE_PHYSICAL ≈ Locale Volume
//   - NT_SCOPE_CONCEPTUAL ≈ Stance
//   - {Conceptual × Physical} = focal cone
//
// Soundness anchors:
//   - Wyrd.Hypergraph.hyperedge_preserves_incident_edges (Phase 2
//     C-20a) — incidence semantics for focal-cone membership
//   - A18 §5 Locale-Bounded Absorption Estimation (forthcoming body
//     PR)
//
// See doc/design/scoutquery.md for the design surface this package
// implements.
package scout
