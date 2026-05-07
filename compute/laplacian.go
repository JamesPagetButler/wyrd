// Package compute extension: Hypergraph Laplacian primitives.
//
// This file is the v0.1 STUB landing the gonum/mat dependency per
// issue #24 (Q4 from the 2026-05-06 closeout). The mathematical body
// — spectral Laplacian construction, eigendecomposition, the
// smoothness-residual signal Gemini's Hypergraph-Inference-BMA paper
// proposes for the BMA Meta-Watchdog — is implemented in a follow-on
// PR once the query/ subpackage (issue #23) lands. That follow-on
// PR consumes query.API for the traversal step.
//
// Soundness anchor (forthcoming, not in v0.1): Tang 2023 + Jost &
// Mulas 2019 establish that the symmetric normalised Hypergraph
// Laplacian's spectral gap is a stable signal of structural
// connectivity changes; the smoothness residual r(x) = xᵀLx / xᵀx
// over a node attribute vector x is the per-node contribution.
// References to be cited inline when the body lands.

package compute

import (
	"errors"

	"gonum.org/v1/gonum/mat"

	"github.com/JamesPagetButler/wyrd/model"
)

// ErrLaplacianNotImplemented marks every method on this stub as not
// yet wired up. Callers that hit it have reached the v0.1 stub
// boundary; the follow-on PR removes this sentinel as each method
// gains its real implementation.
var ErrLaplacianNotImplemented = errors.New(
	"compute: Laplacian primitive not implemented at v0.1 — see issue #24",
)

// LaplacianSmoothnessResidual returns the per-node smoothness
// residual of the symmetric normalised Hypergraph Laplacian over the
// graph g, evaluated against the attribute vector x indexed by the
// nodes in nodeOrder.
//
// The returned slice is parallel to nodeOrder: result[i] is the
// residual at nodeOrder[i].
//
// v0.1 STATUS: stub returns ErrLaplacianNotImplemented. The dense
// matrix construction, normalisation, and eigenvector iteration land
// in the follow-on PR that depends on query.API (issue #23).
//
// Soundness anchor (planned): Tang 2023 §3.1 — the symmetric
// normalised Hypergraph Laplacian L = I - D⁻¹⁄² H W H D⁻¹⁄² has
// real, non-negative spectrum; the smoothness residual is a
// well-defined real-valued per-node signal.
func LaplacianSmoothnessResidual(
	g *model.Graph,
	x *mat.VecDense,
	nodeOrder []model.NodeID,
) ([]float64, error) {
	_ = g
	_ = x
	_ = nodeOrder
	return nil, ErrLaplacianNotImplemented
}
