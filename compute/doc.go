// Package compute provides the operations on Wyrd graphs whose soundness
// is anchored in the Lean corpus:
//
//   - [CanSynthesize] enforces the four-tier ring-tower closure
//     (Phase 1: `lean/Wyrd/Foundations.lean`).
//   - [Bridge] performs atomic Contextus → CTH promotion preserving
//     signal count (Phase 2: `lean/Wyrd/Bridge.lean`).
//   - [TriangleConsistent] / [TriangleConsistentH] verify the additive /
//     multiplicative triangle constraints from Phase 4
//     (`lean/Wyrd/HolographicHypergraph*.lean`).
//
// Each function's doc comment cites the specific Lean theorem it
// implements. Diverging from the spec without updating the theorem (or
// vice-versa) is an audit failure.
package compute
