package compute

import (
	"fmt"

	"github.com/JamesPagetButler/qbp-compute-unit/emulator"
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
// `Wyrd.HamiltonProduct.hamilton_product_formula` in
// `lean/Wyrd/HamiltonProduct.lean` — the Wyrd-local named theorem
// proving the 16-mul/12-add formula equals mathlib4 Quaternion.mul.
// Phase 5 ISA semantics (issue #5) will land an ε-tolerance theorem
// specifying the runtime behaviour at each Width, including the
// QW128 DD path's renormalisation guarantees.
//
// [wyrd#2]: https://github.com/JamesPagetButler/wyrd/issues/2
func HamiltonProduct(a, b model.Weight) (model.Weight, error) {
	if a.Tier != b.Tier {
		return model.Weight{}, errMixedTiers(a.Tier, b.Tier)
	}
	switch a.Tier {
	case model.TierQuaternion:
		// Backend: qbp-compute-unit/emulator/v0.1.0-rc1 Gearbox.QMul64
		// (per doc/wyrd-integration.md §6a.1 + qbp-cu-walk seq=41/42).
		// The Gearbox is stateless from caller view; algebraic-contract
		// invariance preserved (Wyrd.Foundations / Wyrd.Capability).
		g := emulator.NewGearbox()
		ain := [4]float64{a.Components[0], a.Components[1], a.Components[2], a.Components[3]}
		bin := [4]float64{b.Components[0], b.Components[1], b.Components[2], b.Components[3]}
		out := g.QMul64(ain, bin)
		return model.NewQuaternionWeight(out[0], out[1], out[2], out[3]), nil
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
	// Backend: qbp-compute-unit/emulator/v0.1.0-rc1 (per
	// doc/wyrd-integration.md §6a.1 + qbp-cu-walk seq=42).
	//
	// The emulator's QMulHighPrec accepts only the slow-path widths
	// (W256/W512/W1024); for prec at or below QW64 precision (53-bit
	// mantissa), it directs callers to QMul64. We honour that by
	// dispatching to QMul64 directly for fast-path requests.
	g := emulator.NewGearbox()
	ain := [4]float64{a.Components[0], a.Components[1], a.Components[2], a.Components[3]}
	bin := [4]float64{b.Components[0], b.Components[1], b.Components[2], b.Components[3]}

	if prec <= 53 {
		// Fast path; equivalent to HamiltonProduct's QW64.
		out := g.QMul64(ain, bin)
		return model.NewQuaternionWeight(out[0], out[1], out[2], out[3]), nil
	}

	// Slow path: map mantissa-bit prec to emulator.Width per the
	// existing convention in this function's godoc.
	var w emulator.Width
	switch {
	case prec <= 237:
		w = emulator.W256
	case prec <= 489:
		w = emulator.W512
	default:
		w = emulator.W1024
	}
	out, err := g.QMulHighPrec(w, ain, bin)
	if err != nil {
		return model.Weight{}, fmt.Errorf("compute: HamiltonProductHighPrec: %w", err)
	}
	return model.NewQuaternionWeight(out[0], out[1], out[2], out[3]), nil
}

// complexProduct returns the ℂ algebra product (a + bi)(c + di) =
// (ac − bd) + (ad + bc)i.
//
// Note: complex multiplication stays in-tree at v0.1; the §6a.1
// contract per doc/wyrd-integration.md scopes the Gearbox migration
// to Hamilton-product call sites only. ℂ is the inner-projection
// case and runs at native float64 with no precision-tier dispatch
// needed.
func complexProduct(a, b model.Weight) model.Weight {
	ar, ai := a.Components[0], a.Components[1]
	br, bi := b.Components[0], b.Components[1]
	return model.NewComplexWeight(ar*br-ai*bi, ar*bi+ai*br)
}
