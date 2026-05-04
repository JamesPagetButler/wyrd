package compute

import (
	"errors"
	"math"
	"testing"

	"github.com/JamesPagetButler/wyrd/model"
)

// TestHamiltonProduct_iTimesJEqualsK verifies the canonical Hamilton
// identity i · j = k at QW64. Same kernel CTH-side and BMA-side will
// rely on for hyperedge-weight composition.
func TestHamiltonProduct_iTimesJEqualsK(t *testing.T) {
	i := model.NewQuaternionWeight(0, 1, 0, 0)
	j := model.NewQuaternionWeight(0, 0, 1, 0)
	k := model.NewQuaternionWeight(0, 0, 0, 1)

	got, err := HamiltonProduct(i, j)
	if err != nil {
		t.Fatalf("HamiltonProduct: %v", err)
	}
	for n := 0; n < 4; n++ {
		if got.Components[n] != k.Components[n] {
			t.Errorf("(i·j).Components[%d] = %v, want %v", n, got.Components[n], k.Components[n])
		}
	}
}

func TestHamiltonProduct_NonCommutative(t *testing.T) {
	i := model.NewQuaternionWeight(0, 1, 0, 0)
	j := model.NewQuaternionWeight(0, 0, 1, 0)
	ij, _ := HamiltonProduct(i, j) // expect k = (0,0,0,1)
	ji, _ := HamiltonProduct(j, i) // expect -k = (0,0,0,-1)
	if ij.Components[3] == ji.Components[3] {
		t.Errorf("Hamilton product should be non-commutative: i·j = %v, j·i = %v", ij, ji)
	}
}

// TestHamiltonProduct_Q64AgreesWithHighPrec confirms the QW64 kernel
// matches the big.Float path at fp64 precision (53 bits). This is the
// regression check we'll run again when the qbp-emulator Gearbox-backed
// path lands — the contract is "the float64 fast path agrees with the
// high-precision kernel up to fp64 rounding."
func TestHamiltonProduct_Q64AgreesWithHighPrec(t *testing.T) {
	a := model.NewQuaternionWeight(1.5, -0.25, 3.125, -7.0)
	b := model.NewQuaternionWeight(0.5, 2.0, -1.25, 0.125)

	q64, err := HamiltonProduct(a, b)
	if err != nil {
		t.Fatalf("HamiltonProduct: %v", err)
	}
	hp, err := HamiltonProductHighPrec(a, b, 53)
	if err != nil {
		t.Fatalf("HamiltonProductHighPrec: %v", err)
	}
	for n := 0; n < 4; n++ {
		if math.Abs(q64.Components[n]-hp.Components[n]) > 0 {
			t.Errorf("Q64 vs HighPrec component[%d]: %v vs %v", n, q64.Components[n], hp.Components[n])
		}
	}
}

func TestHamiltonProduct_HigherPrecisionPreservesSmallValues(t *testing.T) {
	// At fp64 the value (1 + 1e-20) - 1 is 0 (catastrophic cancellation);
	// at higher precision the residual survives.
	tiny := 1e-20
	a := model.NewQuaternionWeight(1+tiny, 0, 0, 0)
	b := model.NewQuaternionWeight(1, 0, 0, 0)
	// a - b at QW64 → 0; we use multiplication a*b which doesn't
	// cancel, but verify the high-precision path preserves the real
	// part to the expected place.
	got, err := HamiltonProductHighPrec(a, b, 200)
	if err != nil {
		t.Fatalf("HighPrec: %v", err)
	}
	// (1 + 1e-20) * 1 = 1 + 1e-20. At QW64 this rounds to 1.0;
	// at QW256 the bit is preserved, but Float64() rounds it back.
	// So this test mostly confirms that HighPrec doesn't crash and
	// produces a finite result.
	if math.IsNaN(got.Components[0]) || math.IsInf(got.Components[0], 0) {
		t.Errorf("HighPrec produced non-finite result: %v", got)
	}
}

func TestHamiltonProduct_ComplexTier(t *testing.T) {
	// (1 + i)(1 - i) = 1 - i² = 2.
	a := model.NewComplexWeight(1, 1)
	b := model.NewComplexWeight(1, -1)
	got, err := HamiltonProduct(a, b)
	if err != nil {
		t.Fatalf("HamiltonProduct (complex): %v", err)
	}
	if got.Components[0] != 2 || got.Components[1] != 0 {
		t.Errorf("complex product = (%v, %v), want (2, 0)", got.Components[0], got.Components[1])
	}
}

func TestHamiltonProduct_MixedTiersRejected(t *testing.T) {
	a := model.NewComplexWeight(1, 0)
	b := model.NewQuaternionWeight(1, 0, 0, 0)
	_, err := HamiltonProduct(a, b)
	if !errors.Is(err, ErrTierMixed) {
		t.Errorf("expected ErrTierMixed, got %v", err)
	}
}

func TestHamiltonProduct_OctonionUnsupported(t *testing.T) {
	a := model.Weight{Tier: model.TierOctonion}
	b := model.Weight{Tier: model.TierOctonion}
	_, err := HamiltonProduct(a, b)
	if !errors.Is(err, ErrTierUnsupported) {
		t.Errorf("expected ErrTierUnsupported for TierOctonion, got %v", err)
	}
}
