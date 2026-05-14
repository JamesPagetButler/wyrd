package predictions

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JamesPagetButler/wyrd/model"
)

func mkValidPrediction() Prediction {
	return Prediction{
		SignalID: "signal:test",
		Referent: Referent{
			Kind:        KindScalar,
			Identifier:  "cascadia.ets.event.tremor_onset",
			Description: "ETS tremor onset time prediction",
		},
		PredictedValue: 1.5,
		Stance:         []model.NodeID{"contextus:scope:conceptual:slow-slip"},
		Locale:         []model.NodeID{"contextus:scope:physical:cascadia"},
		PredictedAt:    time.Unix(1700000000, 0).UTC(),
	}
}

func TestPrediction_Validate_HappyPath(t *testing.T) {
	if err := mkValidPrediction().Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrediction_Validate_EmptySignalID(t *testing.T) {
	p := mkValidPrediction()
	p.SignalID = ""
	err := p.Validate()
	if !errors.Is(err, ErrPredictionInvalid) {
		t.Errorf("want ErrPredictionInvalid, got %v", err)
	}
}

func TestPrediction_Validate_EmptyReferentIdentifier(t *testing.T) {
	p := mkValidPrediction()
	p.Referent.Identifier = ""
	err := p.Validate()
	if !errors.Is(err, ErrPredictionInvalid) {
		t.Errorf("want ErrPredictionInvalid, got %v", err)
	}
}

func TestPrediction_Validate_ScalarRequiresFloat(t *testing.T) {
	p := mkValidPrediction()
	p.Referent.Kind = KindScalar
	p.PredictedValue = "not-a-float"
	err := p.Validate()
	if !errors.Is(err, ErrPredictionInvalid) {
		t.Errorf("want ErrPredictionInvalid for KindScalar with string value, got %v", err)
	}
}

func TestPrediction_Validate_CategoricalRequiresString(t *testing.T) {
	p := mkValidPrediction()
	p.Referent.Kind = KindCategorical
	p.PredictedValue = 1.5
	err := p.Validate()
	if !errors.Is(err, ErrPredictionInvalid) {
		t.Errorf("want ErrPredictionInvalid for KindCategorical with float value, got %v", err)
	}
}

func TestPrediction_Validate_CategoricalHappyPath(t *testing.T) {
	p := mkValidPrediction()
	p.Referent.Kind = KindCategorical
	p.PredictedValue = "tremor-onset-detected"
	if err := p.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestPrediction_Validate_ProcessKindRejected confirms P7 deferral.
func TestPrediction_Validate_ProcessKindRejected(t *testing.T) {
	p := mkValidPrediction()
	p.Referent.Kind = KindProcess
	err := p.Validate()
	if !errors.Is(err, ErrPredictionInvalid) {
		t.Errorf("want ErrPredictionInvalid for KindProcess, got %v", err)
	}
	if !strings.Contains(err.Error(), "v0.2") {
		t.Errorf("error should reference v0.2 deferral; got %v", err)
	}
}

func TestPrediction_Validate_UnknownKind(t *testing.T) {
	p := mkValidPrediction()
	p.Referent.Kind = "made-up-kind"
	err := p.Validate()
	if !errors.Is(err, ErrPredictionInvalid) {
		t.Errorf("want ErrPredictionInvalid for unknown kind, got %v", err)
	}
}

func TestPrediction_Validate_EmptyStance(t *testing.T) {
	p := mkValidPrediction()
	p.Stance = nil
	err := p.Validate()
	if !errors.Is(err, ErrPredictionInvalid) {
		t.Errorf("want ErrPredictionInvalid, got %v", err)
	}
}

func TestPrediction_Validate_EmptyLocale(t *testing.T) {
	p := mkValidPrediction()
	p.Locale = nil
	err := p.Validate()
	if !errors.Is(err, ErrPredictionInvalid) {
		t.Errorf("want ErrPredictionInvalid, got %v", err)
	}
}

func TestPrediction_Validate_ZeroPredictedAt(t *testing.T) {
	p := mkValidPrediction()
	p.PredictedAt = time.Time{}
	err := p.Validate()
	if !errors.Is(err, ErrPredictionInvalid) {
		t.Errorf("want ErrPredictionInvalid, got %v", err)
	}
}

// TestReferent_KindFieldJSONTag confirms the contextus-impl seq=44 ask
// (`referent_kind` JSON field surfaces explicitly).
func TestReferent_KindFieldJSONTag(t *testing.T) {
	r := Referent{Kind: KindScalar, Identifier: "x"}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"referent_kind":"scalar"`) {
		t.Errorf("want `referent_kind` JSON tag; got %s", b)
	}
}

func TestPrediction_JSONRoundTrip(t *testing.T) {
	in := mkValidPrediction()
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Prediction
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.SignalID != in.SignalID {
		t.Errorf("SignalID lost: %s vs %s", out.SignalID, in.SignalID)
	}
	if out.Referent.Kind != in.Referent.Kind {
		t.Errorf("Kind lost: %s vs %s", out.Referent.Kind, in.Referent.Kind)
	}
	// JSON decodes any → float64 for numeric values.
	gotVal, ok := out.PredictedValue.(float64)
	if !ok || gotVal != 1.5 {
		t.Errorf("PredictedValue lost: %v (%T)", out.PredictedValue, out.PredictedValue)
	}
}

// TestPrediction_CTHAnchor_OmittedWhenNil confirms the optional
// cth_anchor field is absent in JSON when nil (per cth-implementor
// seq=11 ask: pointer for nil = BMA-internal-only).
func TestPrediction_CTHAnchor_OmittedWhenNil(t *testing.T) {
	p := mkValidPrediction()
	p.CTHAnchor = nil
	b, _ := json.Marshal(p)
	if strings.Contains(string(b), "cth_anchor") {
		t.Errorf("nil CTHAnchor should be omitted via omitempty; got %s", b)
	}
}

func TestPrediction_CTHAnchor_PresentWhenSet(t *testing.T) {
	p := mkValidPrediction()
	p.CTHAnchor = &CTHAnchor{AnchorID: "PRED-cascadia-2026-tremor"}
	b, _ := json.Marshal(p)
	if !strings.Contains(string(b), `"anchor_id":"PRED-cascadia-2026-tremor"`) {
		t.Errorf("CTHAnchor not serialized; got %s", b)
	}
}

// TestPrediction_AsNodePayload confirms the §4.3 design — stored as
// Node.Payload bytes on a NodeTypePrediction node.
func TestPrediction_AsNodePayload(t *testing.T) {
	p := mkValidPrediction()
	payload, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	g := model.NewGraph()
	n := model.Node{
		ID:      model.NodeID(p.SignalID),
		Type:    NodeTypePrediction,
		Tier:    model.TierComplex,
		Created: p.PredictedAt,
		Payload: payload,
	}
	if err := g.AddNode(n); err != nil {
		t.Fatalf("AddNode: %v", err)
	}

	got, ok := g.Node(model.NodeID(p.SignalID))
	if !ok {
		t.Fatal("node not found in graph")
	}
	if got.Type != NodeTypePrediction {
		t.Errorf("Node.Type = %s, want %s", got.Type, NodeTypePrediction)
	}
	var roundTrip Prediction
	if err := json.Unmarshal(got.Payload, &roundTrip); err != nil {
		t.Fatalf("payload unmarshal: %v", err)
	}
	if roundTrip.SignalID != p.SignalID {
		t.Errorf("round-trip SignalID broken: %s vs %s", roundTrip.SignalID, p.SignalID)
	}
}
