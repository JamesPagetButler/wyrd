package compute

import (
	"math"

	"github.com/JamesPagetButler/wyrd/model"
)

// TriangleAdditive checks the additive triangle constraint
// φ_ik = φ_ij + φ_jk for three scalar relative phases. Used for ℂ-tier
// (TierComplex) and ℝ-tier consistency. Returns the residual; it is
// zero (within tolerance) iff the three phases are consistent.
//
// Soundness: the image of `Wyrd.HolographicHypergraph.tripleToPairs`
// satisfies this constraint by `tripleToPairs_consistent`. The higher-
// arity generalisation `Wyrd.HolographicHypergraphHigherArity.IsConsistent`
// requires this to hold for ALL triples on N beams.
func TriangleAdditive(phiIJ, phiJK, phiIK float64) float64 {
	return phiIK - (phiIJ + phiJK)
}

// IsAdditiveTriangleConsistent reports whether the residual is within
// tolerance of zero.
func IsAdditiveTriangleConsistent(phiIJ, phiJK, phiIK, tolerance float64) bool {
	return math.Abs(TriangleAdditive(phiIJ, phiJK, phiIK)) <= tolerance
}

// TriangleMultiplicative checks the multiplicative triangle constraint
// q_ik = q_ij · q_jk for three quaternion polarisation states (Hamilton
// product). Used for ℍ-tier (TierQuaternion) consistency.
//
// Soundness: the image of `Wyrd.HolographicHypergraphQuaternion.tripleToPairsH`
// satisfies this constraint by `tripleToPairsH_consistent` (Phase 4 v1.5).
//
// Returns the squared norm of the residual q_ik - q_ij·q_jk. Zero (within
// tolerance) means the three rotations are consistent.
func TriangleMultiplicative(qIJ, qJK, qIK model.Weight) float64 {
	if qIJ.Tier != model.TierQuaternion ||
		qJK.Tier != model.TierQuaternion ||
		qIK.Tier != model.TierQuaternion {
		return math.NaN()
	}
	prod := hamiltonProduct(qIJ, qJK)
	dW := qIK.Components[0] - prod.Components[0]
	dX := qIK.Components[1] - prod.Components[1]
	dY := qIK.Components[2] - prod.Components[2]
	dZ := qIK.Components[3] - prod.Components[3]
	return dW*dW + dX*dX + dY*dY + dZ*dZ
}

// IsMultiplicativeTriangleConsistent reports whether the squared
// residual is within tolerance of zero.
func IsMultiplicativeTriangleConsistent(qIJ, qJK, qIK model.Weight, tolerance float64) bool {
	r := TriangleMultiplicative(qIJ, qJK, qIK)
	if math.IsNaN(r) {
		return false
	}
	return r <= tolerance
}

// hamiltonProduct returns the Hamilton product of two quaternion-tier
// weights. Result is at TierQuaternion regardless of input.
//
// (a · b).w =  a.w·b.w − a.x·b.x − a.y·b.y − a.z·b.z
// (a · b).x =  a.w·b.x + a.x·b.w + a.y·b.z − a.z·b.y
// (a · b).y =  a.w·b.y − a.x·b.z + a.y·b.w + a.z·b.x
// (a · b).z =  a.w·b.z + a.x·b.y − a.y·b.x + a.z·b.w
func hamiltonProduct(a, b model.Weight) model.Weight {
	aw, ax, ay, az := a.Components[0], a.Components[1], a.Components[2], a.Components[3]
	bw, bx, by, bz := b.Components[0], b.Components[1], b.Components[2], b.Components[3]
	return model.NewQuaternionWeight(
		aw*bw-ax*bx-ay*by-az*bz,
		aw*bx+ax*bw+ay*bz-az*by,
		aw*by-ax*bz+ay*bw+az*bx,
		aw*bz+ax*by-ay*bx+az*bw,
	)
}
