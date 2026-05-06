package model

import (
	"errors"
	"fmt"
	"time"
)

// ErrCapabilityViolation is the package-level sentinel signalling that
// a Read/Write capability did not authorise the requested operation.
// It is a sibling sentinel to [github.com/JamesPagetButler/wyrd/compute.ErrPrivilegeViolation];
// they are semantically equivalent but report through different
// packages so callers can distinguish "capability check failed at the
// graph mutation boundary" (this sentinel) from "privilege check
// failed at the free-function CanSynthesize boundary" (compute's).
var ErrCapabilityViolation = errors.New("model: capability violation: holder tier cannot authorise operation at target tier")

// CapabilityError is the typed error returned by [ReadCapability.AllowsRead]
// and [WriteCapability.AllowsWrite] when the holder tier is below the
// target tier. It Unwraps to [ErrCapabilityViolation] so callers can
// use [errors.Is].
type CapabilityError struct {
	// Holder is the tier the capability authorises at.
	Holder Tier
	// Target is the tier of the operation being attempted.
	Target Tier
	// Mode distinguishes "read" vs "write" in the error message; the
	// algebraic check itself is identical for both per the design doc
	// (Phase 1 T2.2 says reads project safely; T2.1.* says writes do
	// not synthesise upward).
	Mode CapabilityMode
}

// CapabilityMode distinguishes read vs write capabilities for the
// purposes of error reporting and logging. The algebraic check is the
// same in v0.2 — the type-level split exists to reflect intent, not
// because the underlying ring-tower closure differs by mode.
type CapabilityMode uint8

// Capability mode tags.
const (
	CapabilityModeRead CapabilityMode = iota
	CapabilityModeWrite
)

// String returns a human-readable label for the mode.
func (m CapabilityMode) String() string {
	switch m {
	case CapabilityModeRead:
		return "read"
	case CapabilityModeWrite:
		return "write"
	default:
		return fmt.Sprintf("mode(%d)", uint8(m))
	}
}

// Error implements the error interface.
func (e CapabilityError) Error() string {
	return fmt.Sprintf("model: capability violation: %s capability at tier %s cannot operate at tier %s",
		e.Mode, e.Holder, e.Target)
}

// Unwrap allows [errors.Is](err, [ErrCapabilityViolation]) to match.
func (e CapabilityError) Unwrap() error {
	return ErrCapabilityViolation
}

// ReadCapability authorises reads of state at HolderTier or below.
// Inner-tier reads are always safe per
// `Wyrd.Projection.kernel_supervisor_safe` (Phase 1 T2.2): outer-ring
// values can always be projected down without privilege violation.
//
// In v0.2, ReadCapability is OPTIONAL on the read path — Wyrd's read
// methods do not require a capability argument. ReadCapability exists
// as a typed value that downstream callers (CTH ρ_net audit
// reconstruction, future read-audit code) can pass for logging /
// classification, but the read itself is unrestricted.
//
// See `doc/design/capability-enforcement.md` v0.2 §4 for the
// resolution to Option A (unrestricted reads, capability-gated
// writes) and the reasoning.
type ReadCapability struct {
	HolderTier Tier
	GrantedAt  time.Time
	Issuer     string // identity of the granting authority (typically "skuld")
}

// AllowsRead reports whether the capability authorises a read at
// target tier. Returns nil if so, or a [CapabilityError] otherwise.
func (c ReadCapability) AllowsRead(target Tier) error {
	return capabilityCheck(c.HolderTier, target, CapabilityModeRead)
}

// WriteCapability authorises mutations to state at HolderTier or
// below. Outer-tier writes from an inner-tier holder are forbidden
// by `Wyrd.Foundations.no_surjection_*` (Phase 1 T2.1.a/b/c).
//
// WriteCapability is REQUIRED on every Graph mutation that goes
// through the *WithCapability methods. The bare AddNode /
// AddHyperedge / RemoveHyperedge methods accept default-tier-Complex
// callers (additive migration; no break for existing consumers).
type WriteCapability struct {
	HolderTier Tier
	GrantedAt  time.Time
	Issuer     string
}

// AllowsWrite reports whether the capability authorises a write at
// target tier. Returns nil if so, or a [CapabilityError] otherwise.
func (c WriteCapability) AllowsWrite(target Tier) error {
	return capabilityCheck(c.HolderTier, target, CapabilityModeWrite)
}

// capabilityCheck is the shared algebraic check: a holder at tier T
// can operate at tier T' ≤ T, never at T' > T. Mirrors the semantics
// of [github.com/JamesPagetButler/wyrd/compute.CanSynthesize] but
// reports through this package's [CapabilityError] /
// [ErrCapabilityViolation] sentinel so callers can scope their
// errors.Is checks to the mutation-boundary path.
func capabilityCheck(holder, target Tier, mode CapabilityMode) error {
	if !holder.IsValid() {
		return fmt.Errorf("model: capability: invalid holder tier %v", holder)
	}
	if !target.IsValid() {
		return fmt.Errorf("model: capability: invalid target tier %v", target)
	}
	if target <= holder {
		return nil
	}
	return CapabilityError{Holder: holder, Target: target, Mode: mode}
}
