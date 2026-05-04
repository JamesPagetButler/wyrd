package compute

import (
	"errors"
	"fmt"

	"github.com/JamesPagetButler/wyrd/model"
)

// ErrPrivilegeViolation is returned when a process attempts to
// synthesise a value at an outer tier from an inner-tier capability.
var ErrPrivilegeViolation = errors.New("wyrd: privilege violation: cannot synthesize across ring boundary")

// PrivilegeError carries the caller and target tiers along with the
// underlying ErrPrivilegeViolation so callers can distinguish boundary
// pairs in errors.Is checks.
type PrivilegeError struct {
	Caller model.Tier
	Target model.Tier
}

// Error implements the error interface.
func (e PrivilegeError) Error() string {
	return fmt.Sprintf("wyrd: privilege violation: caller at %s cannot synthesize at %s", e.Caller, e.Target)
}

// Unwrap returns ErrPrivilegeViolation so errors.Is(err, ErrPrivilegeViolation)
// matches.
func (e PrivilegeError) Unwrap() error {
	return ErrPrivilegeViolation
}

// CanSynthesize reports whether a process operating at caller tier can
// safely produce a value at target tier. It returns nil if so, and a
// PrivilegeError otherwise.
//
// Soundness — the four-tier ring-tower closure (`Wyrd.Foundations`):
//
//	no_surjection_complex_to_quaternion (T2.1.a): ℂ ↛ ℍ
//	no_surjection_quaternion_to_octonion (T2.1.b): ℍ ↛ 𝕆
//	no_surjection_octonion_to_sedenion (T2.1.c): 𝕆 ↛ 𝕊
//
// No process operating in an inner ring can synthesize an outer-ring
// value via any sequence of ring operations. Inner-ring projection of
// an outer-ring value (downward direction) is always safe per
// `Wyrd.Projection.kernel_supervisor_safe`.
func CanSynthesize(caller, target model.Tier) error {
	if !caller.IsValid() {
		return fmt.Errorf("wyrd: invalid caller tier %v", caller)
	}
	if !target.IsValid() {
		return fmt.Errorf("wyrd: invalid target tier %v", target)
	}
	if target <= caller {
		return nil
	}
	return PrivilegeError{Caller: caller, Target: target}
}

// CheckEdgeAccess returns nil if the caller may author or modify the
// hyperedge, given the caller's tier. The check is symmetric to
// CanSynthesize: a caller at tier T can author edges at any tier T' ≤ T.
func CheckEdgeAccess(callerTier model.Tier, e model.Hyperedge) error {
	return CanSynthesize(callerTier, e.Tier())
}
