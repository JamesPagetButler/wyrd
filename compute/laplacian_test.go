package compute

import (
	"errors"
	"math"
	"testing"
	"time"

	"gonum.org/v1/gonum/mat"

	"github.com/JamesPagetButler/wyrd/model"
)

// mkNode and mkEdge are shared with bridge_test.go in this package.

// TestGonumMatLinkable confirms the gonum/mat dep is wired correctly.
func TestGonumMatLinkable(t *testing.T) {
	m := mat.NewDense(2, 2, []float64{1, 0, 0, 1})
	r, c := m.Dims()
	if r != 2 || c != 2 {
		t.Fatalf("identity Dims = (%d,%d), want (2,2)", r, c)
	}
}

// TestLaplacianSmoothnessResidual_InvalidInputs covers the validation
// error paths.
func TestLaplacianSmoothnessResidual_NilGraph(t *testing.T) {
	x := mat.NewVecDense(1, []float64{1})
	_, err := LaplacianSmoothnessResidual(nil, x, []model.NodeID{"a"})
	if !errors.Is(err, ErrLaplacianInvalidInput) {
		t.Errorf("want ErrLaplacianInvalidInput, got %v", err)
	}
}

func TestLaplacianSmoothnessResidual_EmptyNodeOrder(t *testing.T) {
	g := model.NewGraph()
	x := mat.NewVecDense(1, []float64{1})
	_, err := LaplacianSmoothnessResidual(g, x, nil)
	if !errors.Is(err, ErrLaplacianInvalidInput) {
		t.Errorf("want ErrLaplacianInvalidInput, got %v", err)
	}
}

func TestLaplacianSmoothnessResidual_NilVector(t *testing.T) {
	g := model.NewGraph()
	_, err := LaplacianSmoothnessResidual(g, nil, []model.NodeID{"a"})
	if !errors.Is(err, ErrLaplacianInvalidInput) {
		t.Errorf("want ErrLaplacianInvalidInput, got %v", err)
	}
}

func TestLaplacianSmoothnessResidual_MismatchedLength(t *testing.T) {
	g := model.NewGraph()
	_ = g.AddNode(mkNode("a"))
	_ = g.AddNode(mkNode("b"))
	x := mat.NewVecDense(1, []float64{1})
	_, err := LaplacianSmoothnessResidual(g, x, []model.NodeID{"a", "b"})
	if !errors.Is(err, ErrLaplacianInvalidInput) {
		t.Errorf("want ErrLaplacianInvalidInput, got %v", err)
	}
}

func TestLaplacianSmoothnessResidual_UnknownNode(t *testing.T) {
	g := model.NewGraph()
	x := mat.NewVecDense(1, []float64{1})
	_, err := LaplacianSmoothnessResidual(g, x, []model.NodeID{"ghost"})
	if !errors.Is(err, ErrLaplacianInvalidInput) {
		t.Errorf("want ErrLaplacianInvalidInput, got %v", err)
	}
}

func TestLaplacianSmoothnessResidual_DuplicateInNodeOrder(t *testing.T) {
	g := model.NewGraph()
	_ = g.AddNode(mkNode("a"))
	x := mat.NewVecDense(2, []float64{1, 1})
	_, err := LaplacianSmoothnessResidual(g, x, []model.NodeID{"a", "a"})
	if !errors.Is(err, ErrLaplacianInvalidInput) {
		t.Errorf("want ErrLaplacianInvalidInput, got %v", err)
	}
}

func TestLaplacianSmoothnessResidual_ZeroVector(t *testing.T) {
	g := model.NewGraph()
	_ = g.AddNode(mkNode("a"))
	x := mat.NewVecDense(1, []float64{0})
	_, err := LaplacianSmoothnessResidual(g, x, []model.NodeID{"a"})
	if !errors.Is(err, ErrLaplacianZeroVector) {
		t.Errorf("want ErrLaplacianZeroVector, got %v", err)
	}
}

// TestLaplacianSmoothnessResidual_IsolatedNode — a graph with one node
// and no edges has L = 0, so the residual is 0 for every node.
func TestLaplacianSmoothnessResidual_IsolatedNode(t *testing.T) {
	g := model.NewGraph()
	_ = g.AddNode(mkNode("a"))
	x := mat.NewVecDense(1, []float64{1.0})
	r, err := LaplacianSmoothnessResidual(g, x, []model.NodeID{"a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r) != 1 {
		t.Fatalf("result len = %d, want 1", len(r))
	}
	if r[0] != 0 {
		t.Errorf("isolated node residual should be 0, got %g", r[0])
	}
}

// TestLaplacianSmoothnessResidual_ConstantVectorIsSmooth — a constant
// attribute vector on a connected graph yields zero residual at every
// node (constant is in the null space of L). Load-bearing invariant.
func TestLaplacianSmoothnessResidual_ConstantVectorIsSmooth(t *testing.T) {
	g := model.NewGraph()
	for _, id := range []model.NodeID{"a", "b", "c"} {
		_ = g.AddNode(mkNode(id))
	}
	_ = g.AddHyperedge(mkEdge("e1", "a", "b"))
	_ = g.AddHyperedge(mkEdge("e2", "b", "c"))

	x := mat.NewVecDense(3, []float64{2.5, 2.5, 2.5})
	r, err := LaplacianSmoothnessResidual(g, x, []model.NodeID{"a", "b", "c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, v := range r {
		if math.Abs(v) > 1e-12 {
			t.Errorf("constant-vector residual at node %d should be 0, got %g", i, v)
		}
	}
}

// TestLaplacianSmoothnessResidual_TwoNodeEdge — two nodes connected by
// one edge; if x = [1, -1] the residual is high (sign disagreement);
// if x = [1, 1] it's zero. Validates the sign interpretation.
func TestLaplacianSmoothnessResidual_TwoNodeEdge_HighDisagreement(t *testing.T) {
	g := model.NewGraph()
	_ = g.AddNode(mkNode("a"))
	_ = g.AddNode(mkNode("b"))
	_ = g.AddHyperedge(mkEdge("e", "a", "b"))

	xDisagree := mat.NewVecDense(2, []float64{1, -1})
	r, err := LaplacianSmoothnessResidual(g, xDisagree, []model.NodeID{"a", "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Sum of residuals = xᵀLx / xᵀx. For x = [1,-1] with unit edge
	// weight on a 2-node edge: share = 1/2; A = [[0, 0.5], [0.5, 0]];
	// D = diag(0.5, 0.5); L = [[0.5, -0.5], [-0.5, 0.5]]
	// Lx = [0.5·1 + (-0.5)·(-1), -0.5·1 + 0.5·(-1)] = [1, -1]
	// xᵀLx = 1·1 + (-1)·(-1) = 2; xᵀx = 2; ratio = 1.
	// Per-node: r[0] = 1·1/2 = 0.5; r[1] = (-1)·(-1)/2 = 0.5.
	total := r[0] + r[1]
	if math.Abs(total-1.0) > 1e-10 {
		t.Errorf("sum of residuals = %g, want 1.0 (xᵀLx/xᵀx for x=[1,-1])", total)
	}
}

// TestLaplacianSmoothnessResidual_HigherArityEdge confirms the
// clique-expansion weighting: weight is shared across all unordered
// pairs in the edge.
func TestLaplacianSmoothnessResidual_HigherArityEdge(t *testing.T) {
	g := model.NewGraph()
	for _, id := range []model.NodeID{"a", "b", "c"} {
		_ = g.AddNode(mkNode(id))
	}
	// One arity-3 edge with unit weight. share = 1/3 between each pair.
	_ = g.AddHyperedge(mkEdge("e", "a", "b", "c"))

	// x = [1, -1, 1] — b disagrees with both neighbours.
	x := mat.NewVecDense(3, []float64{1, -1, 1})
	r, err := LaplacianSmoothnessResidual(g, x, []model.NodeID{"a", "b", "c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r) != 3 {
		t.Fatalf("result len = %d, want 3", len(r))
	}
	// All three residuals should be positive (each node disagrees with
	// at least one neighbour).
	for i, v := range r {
		if v <= 0 {
			t.Errorf("node %d (x = [1,-1,1]) should have positive residual, got %g", i, v)
		}
	}
}

// TestLaplacianSmoothnessResidual_NegativeWeightEdgeContributesZero
// — an edge with all-zero Weight.Components should contribute nothing
// to the Laplacian (weight = 0 → no connectivity at that edge).
func TestLaplacianSmoothnessResidual_ZeroWeightEdgeContributesZero(t *testing.T) {
	g := model.NewGraph()
	_ = g.AddNode(mkNode("a"))
	_ = g.AddNode(mkNode("b"))
	_ = g.AddHyperedge(model.Hyperedge{
		ID:      "e_zero",
		Nodes:   []model.NodeID{"a", "b"},
		Weight:  model.Weight{Tier: model.TierQuaternion}, // all components 0
		Created: time.Unix(0, 0),
	})

	x := mat.NewVecDense(2, []float64{1, -1})
	r, err := LaplacianSmoothnessResidual(g, x, []model.NodeID{"a", "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, v := range r {
		if math.Abs(v) > 1e-12 {
			t.Errorf("zero-weight edge should produce zero residual at node %d, got %g", i, v)
		}
	}
}

func TestEdgeScalarWeight(t *testing.T) {
	cases := []struct {
		name string
		w    model.Weight
		want float64
	}{
		{"quaternion identity", model.NewQuaternionWeight(1, 0, 0, 0), 1.0},
		{"complex identity", model.NewComplexWeight(1, 0), 1.0},
		{"quaternion magnitude 2", model.NewQuaternionWeight(0, 2, 0, 0), 2.0},
		{"zero weight", model.Weight{Tier: model.TierQuaternion}, 0.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := model.Hyperedge{Weight: tc.w}
			got := edgeScalarWeight(e)
			if math.Abs(got-tc.want) > 1e-12 {
				t.Errorf("edgeScalarWeight(%v) = %g, want %g", tc.w, got, tc.want)
			}
		})
	}
}
