package compute

import (
	"math/big"

	"github.com/JamesPagetButler/wyrd/model"
)

// HamiltonProduct returns a · b under the Hamilton product, dispatching
// to the appropriate precision tier based on a.Tier.
//
// At [model.TierQuaternion] (the standard ℍ algebra, four float64
// components) this is a 16-multiply / 12-add inline kernel — the same
// shape as `qbp-emulator`'s `qmul64Scalar` / `qmul64AVX`. Performance
// at this tier matches a direct float64 implementation.
//
// At higher tiers (TierOctonion, TierSedenion) a placeholder error is
// returned via a zero Weight; full tier support requires either:
//
//   - the qbp-emulator dependency landing once its module path
//     aligns with the qbp-compute-unit repo URL (see [wyrd#2]), or
//   - an in-tree Cayley-Dickson construction (deferred to v0.2+).
//
// Soundness: this is the runtime implementation of
// `Quaternion.mul` in `lean/Wyrd/Foundations.lean`'s ℍ algebra.
// `Wyrd.Capability.sandwich_mul` (T2.4) is the load-bearing soundness
// theorem for compositions of this product. Phase 5 ISA semantics
// (issue #5) will land an ε-tolerance theorem specifying the runtime
// behaviour at each Width, including the QW128 DD path's
// renormalisation guarantees.
//
// [wyrd#2]: https://github.com/JamesPagetButler/wyrd/issues/2
func HamiltonProduct(a, b model.Weight) (model.Weight, error) {
	if a.Tier != b.Tier {
		return model.Weight{}, errMixedTiers(a.Tier, b.Tier)
	}
	switch a.Tier {
	case model.TierQuaternion:
		return hamiltonProductQ64(a, b), nil
	case model.TierComplex:
		return complexProduct(a, b), nil
	case model.TierOctonion, model.TierSedenion:
		return model.Weight{}, errTierUnsupported(a.Tier)
	default:
		return model.Weight{}, errInvalidTier(a.Tier)
	}
}

// HamiltonProductHighPrec computes a · b at arbitrary precision using
// math/big.Float. Used for callers whose Width selection puts them
// above the QW64 fast path. At Crawl phase this is the canonical
// "quaternion at QW128 / QW256 / QW512" path; at Walk phase the
// implementation will dispatch to qbp-emulator's `Gearbox.Mul` once
// the upstream module-path issue is resolved (see [wyrd#2]).
//
// `prec` is the mantissa-bit precision applied to all intermediate
// big.Float operations. Convention: 53 → QW64, 113 → QW128 (binary128
// equivalent), 237 → QW256, 489 → QW512.
//
// Returns a TierQuaternion Weight with the four float64 fields rounded
// from the high-precision result. Higher tiers (octonion, sedenion)
// are not yet supported here either.
//
// [wyrd#2]: https://github.com/JamesPagetButler/wyrd/issues/2
func HamiltonProductHighPrec(a, b model.Weight, prec uint) (model.Weight, error) {
	if a.Tier != model.TierQuaternion || b.Tier != model.TierQuaternion {
		return model.Weight{}, errTierUnsupported(a.Tier)
	}
	aw := bigFromF64(a.Components[0], prec)
	ax := bigFromF64(a.Components[1], prec)
	ay := bigFromF64(a.Components[2], prec)
	az := bigFromF64(a.Components[3], prec)
	bw := bigFromF64(b.Components[0], prec)
	bx := bigFromF64(b.Components[1], prec)
	by := bigFromF64(b.Components[2], prec)
	bz := bigFromF64(b.Components[3], prec)

	// (a · b).w = aw·bw − ax·bx − ay·by − az·bz
	rw := mulBig(aw, bw, prec)
	rw.Sub(rw, mulBig(ax, bx, prec))
	rw.Sub(rw, mulBig(ay, by, prec))
	rw.Sub(rw, mulBig(az, bz, prec))

	// (a · b).x = aw·bx + ax·bw + ay·bz − az·by
	rx := mulBig(aw, bx, prec)
	rx.Add(rx, mulBig(ax, bw, prec))
	rx.Add(rx, mulBig(ay, bz, prec))
	rx.Sub(rx, mulBig(az, by, prec))

	// (a · b).y = aw·by − ax·bz + ay·bw + az·bx
	ry := mulBig(aw, by, prec)
	ry.Sub(ry, mulBig(ax, bz, prec))
	ry.Add(ry, mulBig(ay, bw, prec))
	ry.Add(ry, mulBig(az, bx, prec))

	// (a · b).z = aw·bz + ax·by − ay·bx + az·bw
	rz := mulBig(aw, bz, prec)
	rz.Add(rz, mulBig(ax, by, prec))
	rz.Sub(rz, mulBig(ay, bx, prec))
	rz.Add(rz, mulBig(az, bw, prec))

	rwF, _ := rw.Float64()
	rxF, _ := rx.Float64()
	ryF, _ := ry.Float64()
	rzF, _ := rz.Float64()
	return model.NewQuaternionWeight(rwF, rxF, ryF, rzF), nil
}

// hamiltonProductQ64 is the QW64-precision Hamilton product. Same
// formula as `qbp-emulator/qmath_scalar.go::qmul64Scalar`.
func hamiltonProductQ64(a, b model.Weight) model.Weight {
	aw, ax, ay, az := a.Components[0], a.Components[1], a.Components[2], a.Components[3]
	bw, bx, by, bz := b.Components[0], b.Components[1], b.Components[2], b.Components[3]
	return model.NewQuaternionWeight(
		aw*bw-ax*bx-ay*by-az*bz,
		aw*bx+ax*bw+ay*bz-az*by,
		aw*by-ax*bz+ay*bw+az*bx,
		aw*bz+ax*by-ay*bx+az*bw,
	)
}

// complexProduct returns the ℂ algebra product (a + bi)(c + di) =
// (ac − bd) + (ad + bc)i.
func complexProduct(a, b model.Weight) model.Weight {
	ar, ai := a.Components[0], a.Components[1]
	br, bi := b.Components[0], b.Components[1]
	return model.NewComplexWeight(ar*br-ai*bi, ar*bi+ai*br)
}

func bigFromF64(x float64, prec uint) *big.Float {
	return new(big.Float).SetPrec(prec).SetFloat64(x)
}

func mulBig(a, b *big.Float, prec uint) *big.Float {
	return new(big.Float).SetPrec(prec).Mul(a, b)
}
