package predictions

import (
	"errors"
	"fmt"
	"time"

	"github.com/JamesPagetButler/wyrd/model"
)

// NodeTypePrediction is the model.NodeType value Wyrd stores
// predictions under. Follows the bma.* prefix discipline established
// in PR #16 (bma.runtime.*).
const NodeTypePrediction model.NodeType = "bma.prediction"

// ReferentKind discriminates the predicted-value shape. v0.1 admits
// scalar and categorical; KindProcess (MI on probability distributions)
// is deferred to v0.2 per @qbp-architecture #addendum-18-walk seq=6 P7.
type ReferentKind string

const (
	// KindScalar — PredictedValue is float64.
	KindScalar ReferentKind = "scalar"
	// KindCategorical — PredictedValue is string.
	KindCategorical ReferentKind = "categorical"
	// KindProcess — deferred to v0.2; explicit error at v0.1.
	KindProcess ReferentKind = "process"
)

// Referent is the designated real-world quantity the prediction is
// about. Required at NT_SIGNAL mint time per A18 §2.4 invariant
// "no signal without a referent."
//
// The `referent_kind` JSON field surfaces explicitly per
// @contextus-impl #addendum-18-walk seq=44.
type Referent struct {
	Kind        ReferentKind `json:"referent_kind"`
	Identifier  string       `json:"identifier"`
	Description string       `json:"description,omitempty"`
}

// CTHAnchor is the optional cth_id stamp for predictions that are
// also CTH PRED-* anchors. Per @cth-implementor #addendum-18-walk
// seq=11: NT_SIGNAL carries cth_id when the prediction is
// federation-scored; nil when the prediction is BMA-internal only.
type CTHAnchor struct {
	AnchorID string `json:"anchor_id"` // "PRED-*" prefix per CTH convention
}

// Prediction is the persistent record. Stored as model.Node.Payload
// (JSON) on a node of Type = NodeTypePrediction. The Node.ID is the
// SignalID; the Node.Created is the PredictedAt time.
type Prediction struct {
	SignalID       model.NodeID   `json:"signal_id"`
	Referent       Referent       `json:"referent"`
	PredictedValue any            `json:"predicted_value"`
	Stance         []model.NodeID `json:"stance"`
	Locale         []model.NodeID `json:"locale"`
	PredictedAt    time.Time      `json:"predicted_at"`
	CTHAnchor      *CTHAnchor     `json:"cth_anchor,omitempty"`
	ObservedValue  any            `json:"observed_value,omitempty"`
	ObservedAt     *time.Time     `json:"observed_at,omitempty"`
	Score          *float64       `json:"score,omitempty"`
}

// ErrPredictionInvalid is the sentinel returned by Validate when a
// Prediction fails A18 §2.4 invariants. Wrap with fmt.Errorf at use
// sites; consumers unwrap with errors.Is.
var ErrPredictionInvalid = errors.New("predictions: invalid prediction")

// Validate enforces the A18 §2.4 invariant: every Prediction carries
// a non-empty Referent.Identifier, a v0.1-admitted Kind, a non-zero
// PredictedAt, at least one Stance ref, at least one Locale ref, and
// a PredictedValue type matching Kind. KindProcess is explicitly
// rejected at v0.1 per P7 deferral.
func (p Prediction) Validate() error {
	if p.SignalID == "" {
		return fmt.Errorf("%w: empty SignalID", ErrPredictionInvalid)
	}
	if p.Referent.Identifier == "" {
		return fmt.Errorf("%w: empty Referent.Identifier", ErrPredictionInvalid)
	}
	switch p.Referent.Kind {
	case KindScalar:
		if _, ok := p.PredictedValue.(float64); !ok {
			return fmt.Errorf("%w: KindScalar requires PredictedValue float64, got %T",
				ErrPredictionInvalid, p.PredictedValue)
		}
	case KindCategorical:
		if _, ok := p.PredictedValue.(string); !ok {
			return fmt.Errorf("%w: KindCategorical requires PredictedValue string, got %T",
				ErrPredictionInvalid, p.PredictedValue)
		}
	case KindProcess:
		return fmt.Errorf("%w: KindProcess deferred to v0.2 per A18 P7", ErrPredictionInvalid)
	default:
		return fmt.Errorf("%w: unknown Referent.Kind %q", ErrPredictionInvalid, p.Referent.Kind)
	}
	if len(p.Stance) == 0 {
		return fmt.Errorf("%w: Stance is empty (need ≥1 NT_SCOPE_CONCEPTUAL ref)", ErrPredictionInvalid)
	}
	if len(p.Locale) == 0 {
		return fmt.Errorf("%w: Locale is empty (need ≥1 NT_SCOPE_PHYSICAL ref)", ErrPredictionInvalid)
	}
	if p.PredictedAt.IsZero() {
		return fmt.Errorf("%w: PredictedAt is zero", ErrPredictionInvalid)
	}
	return nil
}
