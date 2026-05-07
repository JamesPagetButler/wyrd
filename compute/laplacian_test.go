package compute

import (
	"errors"
	"testing"

	"gonum.org/v1/gonum/mat"

	"github.com/JamesPagetButler/wyrd/model"
)

// TestLaplacianSmoothnessResidual_StubBoundary confirms the v0.1 stub
// returns the documented sentinel. When the follow-on PR replaces the
// body with a real implementation, this test is replaced by happy-path
// + edge-case tests; it stays here as a "stub still in place" canary
// for as long as the function returns ErrLaplacianNotImplemented.
func TestLaplacianSmoothnessResidual_StubBoundary(t *testing.T) {
	g := model.NewGraph()
	x := mat.NewVecDense(1, []float64{0})
	got, err := LaplacianSmoothnessResidual(g, x, nil)
	if !errors.Is(err, ErrLaplacianNotImplemented) {
		t.Fatalf("want ErrLaplacianNotImplemented, got %v", err)
	}
	if got != nil {
		t.Errorf("want nil result alongside stub error, got %v", got)
	}
}

// TestGonumMatLinkable confirms the gonum/mat dep is wired correctly
// at the build level — landing it ahead of the consumer Laplacian
// body lets us catch any toolchain incompatibility (Q4 was the
// posture decision, this is the build-time validation).
func TestGonumMatLinkable(t *testing.T) {
	m := mat.NewDense(2, 2, []float64{1, 0, 0, 1})
	r, c := m.Dims()
	if r != 2 || c != 2 {
		t.Fatalf("identity Dims = (%d,%d), want (2,2)", r, c)
	}
	if got := m.At(0, 0); got != 1 {
		t.Errorf("identity At(0,0) = %v, want 1", got)
	}
}
