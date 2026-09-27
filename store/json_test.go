package store

import (
	"encoding/json"
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

// TestJSONFile_HyperedgeTypeRoundTrip asserts a tenant-set edge Type
// survives the real Save -> Load path (wyrd#92).
func TestJSONFile_HyperedgeTypeRoundTrip(t *testing.T) {
	g := model.NewGraph()
	for _, id := range []model.NodeID{"a", "b"} {
		_ = g.AddNode(mkNode(id, model.TierQuaternion))
	}
	_ = g.AddHyperedge(model.Hyperedge{
		ID:      "e_typed",
		Nodes:   []model.NodeID{"a", "b"},
		Weight:  model.NewQuaternionWeight(0, 1, 0, 0),
		Type:    "cth.opcode.mul",
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
	e, ok := g2.Hyperedge("e_typed")
	if !ok {
		t.Fatal("e_typed not loaded")
	}
	if e.Type != "cth.opcode.mul" {
		t.Errorf("edge Type after round-trip = %q, want %q", e.Type, "cth.opcode.mul")
	}
}

// TestJSONFile_HyperedgeTypeBackwardCompat asserts an untyped edge
// (Type == "") behaves exactly as before: it round-trips unchanged and
// omitempty keeps a spurious "type" field out of the serialised form
// (wyrd#92).
func TestJSONFile_HyperedgeTypeBackwardCompat(t *testing.T) {
	g := model.NewGraph()
	for _, id := range []model.NodeID{"a", "b"} {
		_ = g.AddNode(mkNode(id, model.TierQuaternion))
	}
	// Constructed the old way — no Type field set.
	_ = g.AddHyperedge(model.Hyperedge{
		ID:      "e_untyped",
		Nodes:   []model.NodeID{"a", "b"},
		Weight:  model.NewQuaternionWeight(0, 0, 1, 0),
		Created: time.Unix(0, 0),
	})

	dir := t.TempDir()
	path := filepath.Join(dir, "graph.json")
	store := JSONFile{Path: path}
	if err := store.Save(g); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// omitempty: an untyped edge must not emit a "type" key. Inspect the
	// serialised hyperedge object specifically (nodes carry their own
	// required "type" field, so a raw substring scan would be a false
	// positive).
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var envelope struct {
		Hyperedges []map[string]json.RawMessage `json:"hyperedges"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("Unmarshal envelope: %v", err)
	}
	if len(envelope.Hyperedges) != 1 {
		t.Fatalf("got %d hyperedges, want 1", len(envelope.Hyperedges))
	}
	if _, present := envelope.Hyperedges[0]["type"]; present {
		t.Errorf("untyped edge serialised a spurious \"type\" field:\n%s", raw)
	}

	g2, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	e, ok := g2.Hyperedge("e_untyped")
	if !ok {
		t.Fatal("e_untyped not loaded")
	}
	if e.Type != "" {
		t.Errorf("untyped edge Type after round-trip = %q, want empty", e.Type)
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
