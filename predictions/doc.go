// Package predictions defines the Wyrd-NT_SIGNAL-side prediction
// record per BMA Theory Addendum 18 §2.4 ("no signal without a
// referent"). One of three layers in the federation prediction
// infrastructure:
//
//   - BMA owns param-predictions (params.ProposalStore — bma-systema
//     PR #93 Phase A-D merged)
//   - Wyrd owns NT_SIGNAL predictions (this package)
//   - CTH owns scoring algorithms (compute.NetCompressionDetail +
//     ChainFidelity — confluent-trust v0.1.0)
//
// Schema-level coordination across the three layers is the BMA-pair's
// §I4 concern; this package owns the Wyrd-side slot.
//
// Storage model: predictions are persisted as `model.Node` of
// `Type = "bma.prediction"` with the Prediction struct as the
// Node.Payload (JSON-encoded). Reuses Wyrd's tier / capability /
// lifecycle plumbing; no parallel store.
//
// See doc/design/scoutquery.md §4 for the design surface.
package predictions
