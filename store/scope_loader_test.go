package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JamesPagetButler/wyrd/model"
)

const validYAML = `
physical_scopes:
  - id: "contextus:scope:physical:cascadia"
    description: "Cascadia subduction zone"
    bounds:
      lat: [40.0, 50.0]
      lon: [-130.0, -120.0]
      time: ["2010-01-01T00:00:00Z", "2030-01-01T00:00:00Z"]
      height: [-50000.0, 0.0]
    type_nodes: ["bma.runtime.geophysical"]

conceptual_scopes:
  - id: "contextus:scope:conceptual:slow-slip"
    description: "Episodic Tremor and Slip phenomena"
    type_nodes: ["bma.runtime.slow-slip"]

  - id: "contextus:scope:conceptual:hamilton-product"
    description: "Hamilton-product algebraic structure"
    type_nodes: ["qbp.algebra.hamilton"]
    tier_immune: true
    salience: 1.0

scope_memberships:
  - scope: "contextus:scope:physical:cascadia"
    member: "contextus:scope:conceptual:slow-slip"
    weight_tier: "complex"
    confidence: 0.95
`

const validJSON = `{
  "physical_scopes": [
    {"id": "p1", "description": "x", "type_nodes": ["t"]}
  ],
  "conceptual_scopes": [
    {"id": "c1", "description": "y", "type_nodes": ["t"]}
  ],
  "scope_memberships": [
    {"scope": "p1", "member": "c1", "weight_tier": "complex"}
  ]
}`

func writeTempConfig(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

func TestLoadScopeConfig_YAML_HappyPath(t *testing.T) {
	path := writeTempConfig(t, "scope.yaml", validYAML)
	g := model.NewGraph()
	if err := LoadScopeConfig(g, path); err != nil {
		t.Fatalf("LoadScopeConfig: %v", err)
	}

	// All three scope nodes present.
	for _, id := range []model.NodeID{
		"contextus:scope:physical:cascadia",
		"contextus:scope:conceptual:slow-slip",
		"contextus:scope:conceptual:hamilton-product",
	} {
		n, ok := g.Node(id)
		if !ok {
			t.Errorf("scope node %q not found", id)
			continue
		}
		if id == "contextus:scope:physical:cascadia" && n.Type != NodeTypeScopePhysical {
			t.Errorf("physical scope wrong Type: %s", n.Type)
		}
	}

	// Hamilton-product node has TierImmune+Salience set per YAML.
	ham, ok := g.Node("contextus:scope:conceptual:hamilton-product")
	if !ok {
		t.Fatal("hamilton-product not found")
	}
	if !ham.TierImmune {
		t.Error("hamilton-product should have TierImmune=true per YAML")
	}
	if ham.Salience != 1.0 {
		t.Errorf("hamilton-product salience = %g, want 1.0", ham.Salience)
	}

	// Slow-slip node has DEFAULTS (no YAML override).
	slow, ok := g.Node("contextus:scope:conceptual:slow-slip")
	if !ok {
		t.Fatal("slow-slip not found")
	}
	if slow.TierImmune {
		t.Error("slow-slip should default to TierImmune=false")
	}
	if slow.Salience != 0.0 {
		t.Errorf("slow-slip salience = %g, want 0.0 (default)", slow.Salience)
	}

	// Membership edge present.
	if g.EdgeCount() != 1 {
		t.Errorf("EdgeCount = %d, want 1", g.EdgeCount())
	}
}

func TestLoadScopeConfig_JSON_HappyPath(t *testing.T) {
	path := writeTempConfig(t, "scope.json", validJSON)
	g := model.NewGraph()
	if err := LoadScopeConfig(g, path); err != nil {
		t.Fatalf("LoadScopeConfig (JSON): %v", err)
	}
	if g.NodeCount() != 2 {
		t.Errorf("NodeCount = %d, want 2", g.NodeCount())
	}
	if g.EdgeCount() != 1 {
		t.Errorf("EdgeCount = %d, want 1", g.EdgeCount())
	}
}

func TestLoadScopeConfig_NilGraph(t *testing.T) {
	path := writeTempConfig(t, "scope.yaml", validYAML)
	err := LoadScopeConfig(nil, path)
	if !errors.Is(err, ErrScopeConfigInvalid) {
		t.Errorf("want ErrScopeConfigInvalid, got %v", err)
	}
}

func TestLoadScopeConfig_FileNotFound(t *testing.T) {
	g := model.NewGraph()
	err := LoadScopeConfig(g, "/nonexistent/path/scope.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadScopeConfig_BadYAML(t *testing.T) {
	path := writeTempConfig(t, "scope.yaml", "not: valid: yaml: {{{")
	g := model.NewGraph()
	err := LoadScopeConfig(g, path)
	if !errors.Is(err, ErrScopeConfigParse) {
		t.Errorf("want ErrScopeConfigParse, got %v", err)
	}
}

// TestLoadScopeConfig_Atomicity_MissingMembershipReference confirms
// PR #40 §2.3: phase-1 failure leaves graph untouched.
func TestLoadScopeConfig_Atomicity_MissingMembershipReference(t *testing.T) {
	cfg := `
physical_scopes:
  - id: "p1"
    description: "x"
    type_nodes: ["t"]
scope_memberships:
  - scope: "p1"
    member: "missing-c1"
    weight_tier: "complex"
`
	path := writeTempConfig(t, "scope.yaml", cfg)
	g := model.NewGraph()
	if err := LoadScopeConfig(g, path); !errors.Is(err, ErrScopeConfigInvalid) {
		t.Errorf("want ErrScopeConfigInvalid, got %v", err)
	}
	// Graph MUST be empty — phase-1 caught the issue before any commit.
	if g.NodeCount() != 0 {
		t.Errorf("atomicity violation: graph has %d nodes after failed load", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Errorf("atomicity violation: graph has %d edges after failed load", g.EdgeCount())
	}
}

func TestLoadScopeConfig_DuplicateIDInConfig(t *testing.T) {
	cfg := `
physical_scopes:
  - id: "p1"
    description: "first"
    type_nodes: ["t"]
  - id: "p1"
    description: "duplicate"
    type_nodes: ["t"]
`
	path := writeTempConfig(t, "scope.yaml", cfg)
	g := model.NewGraph()
	err := LoadScopeConfig(g, path)
	if !errors.Is(err, ErrScopeConfigInvalid) {
		t.Errorf("want ErrScopeConfigInvalid for duplicate, got %v", err)
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error should mention duplicate; got %v", err)
	}
}

func TestLoadScopeConfig_ConflictWithExistingNode(t *testing.T) {
	g := model.NewGraph()
	// Pre-populate with the ID the YAML will try to insert.
	if err := g.AddNode(model.Node{
		ID:      "p1",
		Type:    "test",
		Tier:    model.TierComplex,
		Created: time.Unix(0, 0),
	}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfg := `
physical_scopes:
  - id: "p1"
    description: "conflict"
    type_nodes: ["t"]
`
	path := writeTempConfig(t, "scope.yaml", cfg)
	err := LoadScopeConfig(g, path)
	if !errors.Is(err, ErrScopeLoadConflict) {
		t.Errorf("want ErrScopeLoadConflict, got %v", err)
	}
}

func TestLoadScopeConfig_BadWeightTier(t *testing.T) {
	cfg := `
physical_scopes:
  - id: "p1"
    description: "x"
    type_nodes: ["t"]
conceptual_scopes:
  - id: "c1"
    description: "y"
    type_nodes: ["t"]
scope_memberships:
  - scope: "p1"
    member: "c1"
    weight_tier: "bogus"
`
	path := writeTempConfig(t, "scope.yaml", cfg)
	g := model.NewGraph()
	err := LoadScopeConfig(g, path)
	if !errors.Is(err, ErrScopeConfigInvalid) {
		t.Errorf("want ErrScopeConfigInvalid, got %v", err)
	}
}

func TestLoadScopeConfig_OutOfRangeSalience(t *testing.T) {
	cfg := `
physical_scopes:
  - id: "p1"
    description: "x"
    type_nodes: ["t"]
    salience: 1.5
`
	path := writeTempConfig(t, "scope.yaml", cfg)
	g := model.NewGraph()
	err := LoadScopeConfig(g, path)
	if !errors.Is(err, ErrScopeConfigInvalid) {
		t.Errorf("want ErrScopeConfigInvalid for bad salience, got %v", err)
	}
}

func TestLoadScopeConfigReader_JSON(t *testing.T) {
	g := model.NewGraph()
	if err := LoadScopeConfigReader(g, strings.NewReader(validJSON), ".json"); err != nil {
		t.Fatalf("LoadScopeConfigReader: %v", err)
	}
	if g.NodeCount() != 2 || g.EdgeCount() != 1 {
		t.Errorf("unexpected counts: nodes=%d edges=%d", g.NodeCount(), g.EdgeCount())
	}
}

