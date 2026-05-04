package model

import "fmt"

// Weight is a tier-tagged numeric value attached to a hyperedge.
// The number of components varies with the tier (2, 4, 8, 16); the
// type carries enough capacity for any tier and zero-fills unused slots.
//
// At ℂ (TierComplex):    Components[0..1] = (real, imag)
// At ℍ (TierQuaternion): Components[0..3] = (re, imI, imJ, imK)        — Hamilton order
// At 𝕆 (TierOctonion):   Components[0..7] = (re, e1..e7) per Cayley-Dickson
// At 𝕊 (TierSedenion):   Components[0..15] per Cayley-Dickson
//
// Soundness: a Weight at tier T can be projected down (T → T') for T' < T
// per `Wyrd.Projection.kernel_supervisor_safe`, but a process operating
// at T' cannot synthesize an authentic tier-T weight without an explicit
// capability. See [github.com/JamesPagetButler/wyrd/compute].CanSynthesize.
type Weight struct {
	Tier       Tier
	Components [16]float64
}

// NewComplexWeight constructs a Weight at TierComplex from (re, im).
func NewComplexWeight(re, im float64) Weight {
	w := Weight{Tier: TierComplex}
	w.Components[0] = re
	w.Components[1] = im
	return w
}

// NewQuaternionWeight constructs a Weight at TierQuaternion from
// Hamilton components (re, imI, imJ, imK).
func NewQuaternionWeight(re, imI, imJ, imK float64) Weight {
	w := Weight{Tier: TierQuaternion}
	w.Components[0] = re
	w.Components[1] = imI
	w.Components[2] = imJ
	w.Components[3] = imK
	return w
}

// Validate returns nil if the weight's components beyond its tier's
// dimensionality are zero (any non-zero value there would indicate a
// privilege-tier mismatch — a quaternion claiming to be only complex,
// for example).
func (w Weight) Validate() error {
	if !w.Tier.IsValid() {
		return fmt.Errorf("model: weight: invalid tier %v", w.Tier)
	}
	dim := w.Tier.Components()
	for i := dim; i < len(w.Components); i++ {
		if w.Components[i] != 0 {
			return fmt.Errorf("model: weight: component %d non-zero at tier %s (dim=%d)",
				i, w.Tier, dim)
		}
	}
	return nil
}

// Re returns the scalar (real) part of the weight regardless of tier.
func (w Weight) Re() float64 { return w.Components[0] }

// Active returns the slice of components actually used at this tier.
// The returned slice references the Weight's internal array; do not
// mutate it.
func (w Weight) Active() []float64 {
	return w.Components[:w.Tier.Components()]
}

// IsZero reports whether all active components are zero.
func (w Weight) IsZero() bool {
	for _, c := range w.Active() {
		if c != 0 {
			return false
		}
	}
	return true
}
