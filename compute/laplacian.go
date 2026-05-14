// Package compute extension: Hypergraph Laplacian primitives.
//
// The body implements the symmetric normalised Hypergraph Laplacian
// smoothness-residual computation per BMA Theory Addendum 18 §5
// (Locale-Bounded Absorption Estimation) and the Hypergraph-Inference-
// BMA paper (Gemini, 2026-05-05). The substrate signal Gemini's
// Meta-Watchdog framework consumes for detecting structural
// connectivity changes.
//
// Soundness anchors:
//   - Tang 2023 §3.1 — the symmetric normalised Hypergraph Laplacian
//     L = I - D⁻¹⁄² H W H D⁻¹⁄² has real, non-negative spectrum; the
//     smoothness residual is a well-defined real-valued per-node signal
//   - Jost & Mulas 2019 — oriented Hypergraph Laplacian properties
//     (basis for v0.2's oriented variant when issue #30's schema lands)
//
// At v0.1 we ship the SYMMETRIC (clique-expansion) form, which is the
// case Wyrd's current model.Hyperedge supports (no orientation
// metadata; see PR #26 §2.1). The oriented variant lands alongside
// issue #30's oriented-hyperedge model extension.
//
// References to gonum.org/v1/gonum/mat: dense matrices for v0.1
// readability; sparse variants are a v0.2 candidate once Walk-α
// profiling on real BMA graphs identifies allocation pressure.

package compute

import (
	"errors"
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"

	"github.com/JamesPagetButler/wyrd/model"
)

// Sentinel errors. Consumers (BMA Meta-Watchdog, ScoutQuery v0.2
// body) use errors.Is to dispatch on failure shape.
var (
	// ErrLaplacianInvalidInput is returned when the graph + nodeOrder
	// + attribute vector are inconsistent (mismatched lengths,
	// unknown nodes, etc.).
	ErrLaplacianInvalidInput = errors.New("compute: invalid input to Laplacian computation")

	// ErrLaplacianZeroVector is returned when the attribute vector x
	// has zero L2 norm — the residual ratio is undefined.
	ErrLaplacianZeroVector = errors.New("compute: attribute vector has zero L2 norm")
)

// LaplacianSmoothnessResidual returns the per-node smoothness
// residual of the symmetric Hypergraph Laplacian L over the graph g,
// evaluated against the attribute vector x indexed by nodeOrder.
//
// The returned slice is parallel to nodeOrder: result[i] is the
// contribution at nodeOrder[i] to the total smoothness ratio
// xᵀLx / xᵀx.
//
// Mathematical form (v0.1; symmetric / clique-expansion only):
//   - Build the |V|×|V| symmetric Laplacian L = D - A on the
//     clique-expanded weighted graph, where A[u,v] = sum over edges
//     e containing both u and v of (weight(e) / arity(e)) and D is
//     the diagonal of row-sums of A.
//   - Compute the per-node contribution r[v] = x[v] · (Lx)[v] / xᵀx.
//   - Sum of r[v] across all nodes equals xᵀLx / xᵀx (the global
//     smoothness ratio).
//
// Edge weight: scalar = L2 norm of e.Weight.Components for the active
// tier. Hyperedges with zero-norm weight contribute zero to L; this
// is consistent with the "unweighted" case being weight=1.0 by
// construction (NewQuaternionWeight, NewComplexWeight, etc.).
//
// Concurrency: takes [model.Graph.RLock] via the underlying
// Graph.Hyperedges + Graph.Node reads; observer-compatible per
// ADR-003 §I1.
//
// Soundness: per Tang 2023 §3.1, L is symmetric positive-semidefinite
// for non-negative edge weights; thus xᵀLx ≥ 0 and the per-node
// contributions admit a sign interpretation (positive = node value
// agrees with neighbour-average; large = structurally surprising).
func LaplacianSmoothnessResidual(
	g *model.Graph,
	x *mat.VecDense,
	nodeOrder []model.NodeID,
) ([]float64, error) {
	if g == nil {
		return nil, fmt.Errorf("%w: nil graph", ErrLaplacianInvalidInput)
	}
	n := len(nodeOrder)
	if n == 0 {
		return nil, fmt.Errorf("%w: empty nodeOrder", ErrLaplacianInvalidInput)
	}
	if x == nil {
		return nil, fmt.Errorf("%w: nil attribute vector", ErrLaplacianInvalidInput)
	}
	if x.Len() != n {
		return nil, fmt.Errorf("%w: x.Len() = %d, want %d", ErrLaplacianInvalidInput, x.Len(), n)
	}

	// Index map for nodeOrder → matrix row. Detect duplicates.
	idx := make(map[model.NodeID]int, n)
	for i, id := range nodeOrder {
		if _, dup := idx[id]; dup {
			return nil, fmt.Errorf("%w: duplicate node %q in nodeOrder", ErrLaplacianInvalidInput, id)
		}
		idx[id] = i
		if _, ok := g.Node(id); !ok {
			return nil, fmt.Errorf("%w: node %q not in graph", ErrLaplacianInvalidInput, id)
		}
	}

	// Build the symmetric weighted-adjacency matrix A on the clique
	// expansion. A is symmetric by construction (we add both (u, v)
	// and (v, u) for each unordered pair).
	A := mat.NewDense(n, n, nil)
	for _, e := range g.Hyperedges() {
		k := len(e.Nodes)
		if k < 2 {
			// Self-loops (k=1) contribute no off-diagonal mass; skip.
			continue
		}
		w := edgeScalarWeight(e)
		if w == 0 {
			continue
		}
		share := w / float64(k)
		// Add share to every unordered pair (u, v) of distinct nodes in e.
		for i := range e.Nodes {
			iIdx, ok := idx[e.Nodes[i]]
			if !ok {
				// Edge references a node outside nodeOrder; skip.
				continue
			}
			for j := i + 1; j < len(e.Nodes); j++ {
				jIdx, ok := idx[e.Nodes[j]]
				if !ok {
					continue
				}
				A.Set(iIdx, jIdx, A.At(iIdx, jIdx)+share)
				A.Set(jIdx, iIdx, A.At(jIdx, iIdx)+share)
			}
		}
	}

	// Degree matrix D = diag(row-sums of A). Construct L = D - A.
	// Since D is diagonal, we can fold into a single Dense
	// subtraction: L = D - A where D[i,i] = sum_j A[i,j].
	L := mat.NewDense(n, n, nil)
	for i := range n {
		var rowSum float64
		for j := range n {
			rowSum += A.At(i, j)
		}
		// L[i,j] = -A[i,j] for i≠j; L[i,i] = D[i,i] - A[i,i] = rowSum - A[i,i].
		// Since A is symmetric with A[i,i] = 0 by construction, L[i,i] = rowSum.
		for j := range n {
			if i == j {
				L.Set(i, j, rowSum-A.At(i, j))
			} else {
				L.Set(i, j, -A.At(i, j))
			}
		}
	}

	// Compute Lx.
	Lx := mat.NewVecDense(n, nil)
	Lx.MulVec(L, x)

	// xᵀx denominator.
	xTx := mat.Dot(x, x)
	if xTx == 0 {
		return nil, ErrLaplacianZeroVector
	}

	// Per-node contribution: r[v] = x[v] · (Lx)[v] / (xᵀx).
	// Sum of r[v] across all v equals xᵀLx / xᵀx.
	out := make([]float64, n)
	for i := range n {
		out[i] = x.AtVec(i) * Lx.AtVec(i) / xTx
	}
	return out, nil
}

// edgeScalarWeight extracts a scalar edge weight from a model.Weight.
// Convention (v0.1): the L2 norm of the active-tier components.
//
// For TierQuaternion: sqrt(w² + x² + y² + z²) over Components[0..3].
// For TierComplex: sqrt(re² + im²) over Components[0..1].
// For higher tiers: norm over Components[0..2^tier - 1].
//
// NewQuaternionWeight(1, 0, 0, 0) (the identity) yields weight 1.0,
// matching the "unweighted-edge default" intuition.
func edgeScalarWeight(e model.Hyperedge) float64 {
	n := tierComponentCount(e.Weight.Tier)
	if n == 0 {
		return 0
	}
	var sumSq float64
	for i := 0; i < n && i < len(e.Weight.Components); i++ {
		c := e.Weight.Components[i]
		sumSq += c * c
	}
	return math.Sqrt(sumSq)
}

// tierComponentCount returns the number of components carried by a
// Weight at the given tier. ℂ=2, ℍ=4, 𝕆=8, 𝕊=16.
func tierComponentCount(t model.Tier) int {
	switch t {
	case model.TierComplex:
		return 2
	case model.TierQuaternion:
		return 4
	case model.TierOctonion:
		return 8
	case model.TierSedenion:
		return 16
	default:
		return 0
	}
}
