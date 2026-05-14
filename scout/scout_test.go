package scout

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/JamesPagetButler/wyrd/model"
)

const testType model.NodeType = "test.agent"
const otherType model.NodeType = "test.other"

func mkNode(id model.NodeID, typ model.NodeType) model.Node {
	return model.Node{
		ID:      id,
		Type:    typ,
		Tier:    model.TierComplex,
		Created: time.Unix(0, 0),
	}
}

func mkVolume() Volume {
	return Volume{
		Min: [4]float64{40.0, -130.0, 0.0, -50000.0},
		Max: [4]float64{50.0, -120.0, 1.0, 0.0},
	}
}

func TestScoutQuery_NilGraph(t *testing.T) {
	_, err := ScoutQuery(nil, mkVolume(), "a", "b", []model.NodeType{testType}, W8)
	if !errors.Is(err, ErrScoutQueryInvalid) {
		t.Errorf("want ErrScoutQueryInvalid, got %v", err)
	}
}

func TestScoutQuery_EmptySource(t *testing.T) {
	g := model.NewGraph()
	_, err := ScoutQuery(g, mkVolume(), "", "sink", []model.NodeType{testType}, W8)
	if !errors.Is(err, ErrScoutQueryInvalid) {
		t.Errorf("want ErrScoutQueryInvalid, got %v", err)
	}
}

func TestScoutQuery_EmptySink(t *testing.T) {
	g := model.NewGraph()
	_, err := ScoutQuery(g, mkVolume(), "src", "", []model.NodeType{testType}, W8)
	if !errors.Is(err, ErrScoutQueryInvalid) {
		t.Errorf("want ErrScoutQueryInvalid, got %v", err)
	}
}

func TestScoutQuery_BadPrecision(t *testing.T) {
	g := model.NewGraph()
	_, err := ScoutQuery(g, mkVolume(), "src", "sink", []model.NodeType{testType}, Width(42))
	if !errors.Is(err, ErrScoutQueryInvalid) {
		t.Errorf("want ErrScoutQueryInvalid, got %v", err)
	}
}

func TestScoutQuery_EmptyAgentTypes(t *testing.T) {
	g := model.NewGraph()
	got, err := ScoutQuery(g, mkVolume(), "src", "sink", nil, W8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Error("want empty slice, got nil")
	}
	if len(got) != 0 {
		t.Errorf("want empty, got %v", got)
	}
}

func TestScoutQuery_NoMatchingNodes(t *testing.T) {
	g := model.NewGraph()
	_ = g.AddNode(mkNode("n1", otherType))
	got, err := ScoutQuery(g, mkVolume(), "src", "sink", []model.NodeType{testType}, W8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("want empty (no matching), got %v", got)
	}
}

func TestScoutQuery_PlaceholderHappyPath(t *testing.T) {
	g := model.NewGraph()
	_ = g.AddNode(mkNode("a1", testType))
	_ = g.AddNode(mkNode("a2", testType))
	_ = g.AddNode(mkNode("other", otherType))

	got, err := ScoutQuery(g, mkVolume(), "src", "sink", []model.NodeType{testType}, W8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 intersections, got %d", len(got))
	}

	// Sort by AgentID for deterministic check.
	slices.SortFunc(got, func(a, b Intersection) int {
		if a.AgentID < b.AgentID {
			return -1
		} else if a.AgentID > b.AgentID {
			return 1
		}
		return 0
	})

	for i, want := range []model.NodeID{"a1", "a2"} {
		if got[i].AgentID != want {
			t.Errorf("Intersection[%d].AgentID = %s, want %s", i, got[i].AgentID, want)
		}
		if got[i].AgentType != testType {
			t.Errorf("Intersection[%d].AgentType = %s, want %s", i, got[i].AgentType, testType)
		}
		if got[i].AbsorptionGain != uniformPlaceholderGain {
			t.Errorf("Intersection[%d].AbsorptionGain = %g, want %g (placeholder)", i, got[i].AbsorptionGain, uniformPlaceholderGain)
		}
		// Provenance contains [source, sink].
		if len(got[i].Provenance) != 2 || got[i].Provenance[0] != "src" || got[i].Provenance[1] != "sink" {
			t.Errorf("Intersection[%d].Provenance = %v, want [src, sink]", i, got[i].Provenance)
		}
	}
}

func TestScoutQuery_PlaceholderIsStanceBlind(t *testing.T) {
	// Per PR #35 §3.1 godoc: v0.1 must NOT vary AbsorptionGain by Stance.
	// Same graph + same agents but different Provenance should produce
	// identical AbsorptionGain values for matched intersections.
	g := model.NewGraph()
	_ = g.AddNode(mkNode("agent1", testType))

	resultA, _ := ScoutQuery(g, mkVolume(), "stance-a", "sink", []model.NodeType{testType}, W8)
	resultB, _ := ScoutQuery(g, mkVolume(), "stance-b", "sink", []model.NodeType{testType}, W128)

	if len(resultA) != 1 || len(resultB) != 1 {
		t.Fatalf("setup error: want 1 intersection each; got %d / %d", len(resultA), len(resultB))
	}
	if resultA[0].AbsorptionGain != resultB[0].AbsorptionGain {
		t.Errorf("v0.1 must be Stance-blind; got A=%g vs B=%g", resultA[0].AbsorptionGain, resultB[0].AbsorptionGain)
	}
}
