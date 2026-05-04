package compute

import (
	"math"
	"testing"

	"github.com/JamesPagetButler/wyrd/model"
)

// TestTriangleAdditive_ConsistentReturnsZero is the runtime check for
// `Wyrd.HolographicHypergraph.tripleToPairs_consistent` (Phase 4 v1.4):
// φ_ik = φ_ij + φ_jk produces residual zero.
func TestTriangleAdditive_ConsistentReturnsZero(t *testing.T) {
	// Coherent triple: φ_12 = 1.0, φ_23 = 2.0, φ_13 = 3.0.
	if got := TriangleAdditive(1.0, 2.0, 3.0); got != 0 {
		t.Errorf("consistent triple residual = %v, want 0", got)
	}
}

// TestTriangleAdditive_TheoremTwoWitness reproduces the witness from
// `theorem2_irreducibility` (Phase 4 v1.4) — pair config ⟨0, 0, π⟩ is
// inconsistent.
func TestTriangleAdditive_TheoremTwoWitness(t *testing.T) {
	residual := TriangleAdditive(0.0, math.Pi, 0.0)
	if residual == 0 {
		t.Error("witness ⟨0, π, 0⟩ should have non-zero residual")
	}
}

// TestTriangleMultiplicative_TheoremTwoQuaternionWitness reproduces the
// witness from `theorem2_irreducibility_quaternion` (Phase 4 v1.5):
// q_AB = i, q_BC = j, q_AC = i has residual ‖i − k‖² = 2 (since i·j = k).
func TestTriangleMultiplicative_TheoremTwoQuaternionWitness(t *testing.T) {
	qAB := model.NewQuaternionWeight(0, 1, 0, 0) // i
	qBC := model.NewQuaternionWeight(0, 0, 1, 0) // j
	qAC := model.NewQuaternionWeight(0, 1, 0, 0) // i (witness — should be k = i·j)

	residual := TriangleMultiplicative(qAB, qBC, qAC)
	if math.Abs(residual-2.0) > 1e-12 {
		t.Errorf("residual = %v, want 2.0 (‖i − k‖²)", residual)
	}
}

func TestTriangleMultiplicative_Consistent(t *testing.T) {
	// Coherent triple: q_AB = i, q_BC = j, q_AC = i·j = k = (0,0,0,1).
	qAB := model.NewQuaternionWeight(0, 1, 0, 0)
	qBC := model.NewQuaternionWeight(0, 0, 1, 0)
	qAC := model.NewQuaternionWeight(0, 0, 0, 1) // k

	residual := TriangleMultiplicative(qAB, qBC, qAC)
	if residual > 1e-12 {
		t.Errorf("consistent triple residual = %v, want ~0", residual)
	}
}

func TestTriangleMultiplicative_RejectsNonQuaternion(t *testing.T) {
	cw := model.NewComplexWeight(1, 0)
	qw := model.NewQuaternionWeight(0, 1, 0, 0)
	if !math.IsNaN(TriangleMultiplicative(cw, qw, qw)) {
		t.Error("expected NaN for mixed-tier inputs")
	}
}
