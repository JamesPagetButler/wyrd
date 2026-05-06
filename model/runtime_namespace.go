package model

import "strings"

// Reserved Node.Type prefix for runtime-generated anchors authored
// by the BMA WDEvent observer. Values under this prefix are written
// at runtime in response to algebraic events from QBP-CU; they are
// distinct from authored-state hyperedges (engrams, derivation
// chains, etc.) authored during sleep cycles or scout activity.
//
// The closed list of M1+M2 anchor types was specified by
// @bma-implementor on qbp-cu-walk seq=11 (2026-05-06):
//
//   - flag-* — asserted invariant violation (a constraint the
//     observer believes was crossed; downstream ρ_net consumers
//     should treat this as evidence of algebraic stress)
//   - obs-*  — runtime observation (a fact the observer recorded;
//     not necessarily a violation)
//
// Each anchor's instance ID receives the WAL seqno appended at
// write time per `BMA/doc/handoff/2026-05-05-bma-reply-architecture.md`
// §5.2; the typed constants here are the type-namespace prefix only.
//
// Soundness: layered with CTH's `cth_id` audit anchor (per ADR-003
// §I2 — unified RUNTIME-* schema) — `cth_id` is the audit-trail
// identifier on the CTH side; `Node.Type` is Wyrd's typed dispatch.
// Both exist on a single anchor; they are not duplicates.
const (
	// RuntimeAnchorPrefix is the reserved namespace; consumers MUST
	// use the typed constants below and MUST NOT author new
	// `bma.runtime.*` types without coordinating an update here +
	// in the BMA reply doc.
	RuntimeAnchorPrefix NodeType = "bma.runtime."

	// NodeTypeBMARuntimeFlagNormDrift is asserted when a WDEvent's
	// `NormDelta` exceeds the configured ε. Per-node anchor —
	// instance ID resolves to a single Wyrd Node whose norm is the
	// event subject.
	NodeTypeBMARuntimeFlagNormDrift NodeType = "bma.runtime.flag-norm-drift"

	// NodeTypeBMARuntimeObsZDDetected is recorded when a WDEvent's
	// `ZDClass` is not `NotZD`. Per-(i, j, k, l) anchor — instance
	// ID resolves to the basis quad whose product produced the
	// zero-divisor signal.
	NodeTypeBMARuntimeObsZDDetected NodeType = "bma.runtime.obs-zd-detected"

	// NodeTypeBMARuntimeObsCounter is the ambient per-algebra
	// op counter, aggregated by window. Not a per-event anchor —
	// one instance per (algebra-id, window) tuple, summarising
	// activity across that window.
	NodeTypeBMARuntimeObsCounter NodeType = "bma.runtime.obs-runtime-counter"

	// NodeTypeBMARuntimeObsFault carries WDEvent fault codes
	// 0x10–0x14 (ILLEGAL_DECRYSTALLISATION, PSEL_TIMEOUT,
	// BSEL_TIMEOUT, BUS_STATE_NONZERO, MALFORMED_BASIS_SUM) per
	// `qbp-compute-unit/architecture/peer-review-005-stream-migration.md`
	// §"What changes for BMA at M1". One instance per fault event.
	NodeTypeBMARuntimeObsFault NodeType = "bma.runtime.obs-fault"
)

// IsRuntimeAnchor reports whether t is a runtime-anchor Node.Type
// (i.e., authored by the WDEvent observer in response to an algebraic
// event, not by sleep-cycle / scout / authored-state code paths).
//
// True for any value matching the [RuntimeAnchorPrefix]; not just
// the four canonical constants — future additions under the
// reserved prefix are correctly classified.
func IsRuntimeAnchor(t NodeType) bool {
	return strings.HasPrefix(string(t), string(RuntimeAnchorPrefix))
}
