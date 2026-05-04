package model

import (
	"encoding/json"
	"fmt"
)

// Tier represents a position in the Cayley-Dickson algebraic privilege
// tower: ℂ ⊂ ℍ ⊂ 𝕆 ⊂ 𝕊. Each tier corresponds to a doubling of the
// previous algebra and a strict loss of an algebraic property
// (commutativity at ℂ→ℍ, associativity at ℍ→𝕆, alternativity at 𝕆→𝕊).
//
// Soundness: per `lean/Wyrd/Foundations.lean`, no ring homomorphism
// from an inner tier to an outer tier is surjective:
//
//	no_surjection_complex_to_quaternion (T2.1.a)   — ℂ ↛ ℍ
//	no_surjection_quaternion_to_octonion (T2.1.b)  — ℍ ↛ 𝕆
//	no_surjection_octonion_to_sedenion (T2.1.c)    — 𝕆 ↛ 𝕊
//
// The Skuld supervisor enforces the resulting privilege model at the
// runtime boundary (see [github.com/JamesPagetButler/wyrd/compute].CanSynthesize).
type Tier int

const (
	// TierComplex is ℂ — the user-facing algebra. Two real components.
	TierComplex Tier = iota
	// TierQuaternion is ℍ — the supervisor algebra (Hamilton).
	// Four real components, non-commutative.
	TierQuaternion
	// TierOctonion is 𝕆 — the kernel algebra. Eight real components,
	// non-commutative and non-associative; alternative.
	TierOctonion
	// TierSedenion is 𝕊 — the firmware algebra. Sixteen real
	// components, non-alternative. The outermost tier in the standard
	// four-tier privilege model.
	TierSedenion
)

// Canonical lowercase tier names. Used by String / MarshalJSON /
// UnmarshalJSON. Exported so consumers (CTH, BMA, Contextus) can
// use the same identifiers when authoring node/edge metadata.
const (
	TierNameComplex    = "complex"
	TierNameQuaternion = "quaternion"
	TierNameOctonion   = "octonion"
	TierNameSedenion   = "sedenion"
)

// String returns the canonical lowercase name of the tier.
func (t Tier) String() string {
	switch t {
	case TierComplex:
		return TierNameComplex
	case TierQuaternion:
		return TierNameQuaternion
	case TierOctonion:
		return TierNameOctonion
	case TierSedenion:
		return TierNameSedenion
	default:
		return fmt.Sprintf("tier(%d)", int(t))
	}
}

// IsValid reports whether the tier is one of the four defined values.
func (t Tier) IsValid() bool {
	return t >= TierComplex && t <= TierSedenion
}

// Components returns the number of real components at this tier
// (Cayley-Dickson dimension): 2 at ℂ, 4 at ℍ, 8 at 𝕆, 16 at 𝕊.
func (t Tier) Components() int {
	switch t {
	case TierComplex:
		return 2
	case TierQuaternion:
		return 4
	case TierOctonion:
		return 8
	case TierSedenion:
		return 16
	default:
		return 0
	}
}

// MarshalJSON encodes the tier as a JSON string. iota integers are NEVER
// serialised directly — see CTH's enum convention.
func (t Tier) MarshalJSON() ([]byte, error) {
	if !t.IsValid() {
		return nil, fmt.Errorf("model: invalid tier value %d", int(t))
	}
	return json.Marshal(t.String())
}

// UnmarshalJSON decodes a tier from its canonical lowercase name.
func (t *Tier) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("model: tier: %w", err)
	}
	switch s {
	case TierNameComplex:
		*t = TierComplex
	case TierNameQuaternion:
		*t = TierQuaternion
	case TierNameOctonion:
		*t = TierOctonion
	case TierNameSedenion:
		*t = TierSedenion
	default:
		return fmt.Errorf("model: unknown tier %q", s)
	}
	return nil
}
