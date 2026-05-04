package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JamesPagetButler/wyrd/model"
)

func mkNode(id model.NodeID, tier model.Tier) model.Node {
	return model.Node{ID: id, Type: "test", Tier: tier, Created: time.Unix(0, 0)}
}

func TestJSONFile_RoundTrip(t *testing.T) {
	g := model.NewGraph()
	for _, id := range []model.NodeID{"a", "b", "c", "d"} {
		_ = g.AddNode(mkNode(id, model.TierQuaternion))
	}
	_ = g.AddHyperedge(model.Hyperedge{
		ID:      "e_abc",
		Nodes:   []model.NodeID{"a", "b", "c"},
		Weight:  model.NewQuaternionWeight(0, 1, 0, 0),
		Created: time.Unix(0, 0),
	})
	_ = g.AddHyperedge(model.Hyperedge{
		ID:      "e_bd",
		Nodes:   []model.NodeID{"b", "d"},
		Weight:  model.NewQuaternionWeight(0, 0, 1, 0),
		Created: time.Unix(0, 0),
	})

	dir := t.TempDir()
	path := filepath.Join(dir, "graph.json")
	store := JSONFile{Path: path}
	if err := store.Save(g); err != nil {
		t.Fatalf("Save: %v", err)
	}

	g2, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if g2.NodeCount() != g.NodeCount() {
		t.Errorf("node count: got %d, want %d", g2.NodeCount(), g.NodeCount())
	}
	if g2.EdgeCount() != g.EdgeCount() {
		t.Errorf("edge count: got %d, want %d", g2.EdgeCount(), g.EdgeCount())
	}
	// Spot-check incidence index was rebuilt:
	if got := len(g2.IncidentEdges("b")); got != 2 {
		t.Errorf("IncidentEdges(b) after load = %d, want 2", got)
	}
	// Spot-check a quaternion-weight component survived round-trip.
	e, ok := g2.Hyperedge("e_abc")
	if !ok {
		t.Fatal("e_abc not loaded")
	}
	if e.Weight.Components[1] != 1 {
		t.Errorf("e_abc weight imI = %v, want 1", e.Weight.Components[1])
	}
}

func TestJSONFile_VersionMismatchRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wrong-version.json")
	// Hand-craft an unsupported version.
	if err := writeFile(path, []byte(`{"version":99,"nodes":[],"hyperedges":[]}`)); err != nil {
		t.Fatalf("setup: %v", err)
	}
	store := JSONFile{Path: path}
	if _, err := store.Load(); err == nil {
		t.Error("expected error loading unsupported version")
	}
}

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}
