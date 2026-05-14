package model

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// --- TierImmune + Salience validation ---

func TestNode_Validate_SalienceInRange(t *testing.T) {
	cases := []struct {
		name     string
		salience float64
		wantErr  bool
	}{
		{"zero (default)", 0.0, false},
		{"middle", 0.5, false},
		{"max", 1.0, false},
		{"below range", -0.1, true},
		{"above range", 1.1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n := Node{
				ID:       "n1",
				Type:     testIssuer,
				Tier:     TierComplex,
				Created:  time.Unix(0, 0),
				Salience: tc.salience,
			}
			err := n.Validate()
			if tc.wantErr && err == nil {
				t.Errorf("salience=%g: want error, got nil", tc.salience)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("salience=%g: want no error, got %v", tc.salience, err)
			}
		})
	}
}

// --- Wire-format compatibility ---

// TestNode_JSON_v01_DeserializesIntoV02 confirms v0.1 JSON (no
// tier_immune / salience fields) deserializes into a v0.2 Node with
// the new fields at their defaults.
func TestNode_JSON_v01_DeserializesIntoV02(t *testing.T) {
	v01JSON := `{"id":"n1","type":"test","tier":"complex","created":"1970-01-01T00:00:00Z"}`
	var n Node
	if err := json.Unmarshal([]byte(v01JSON), &n); err != nil {
		t.Fatalf("unmarshal v0.1: %v", err)
	}
	if n.TierImmune {
		t.Error("TierImmune should default to false")
	}
	if n.Salience != 0.0 {
		t.Errorf("Salience should default to 0.0, got %g", n.Salience)
	}
	if n.ID != "n1" {
		t.Errorf("ID round-trip broken: got %s", n.ID)
	}
}

// TestNode_JSON_v02_DefaultsOmitted confirms that v0.2 Nodes with
// defaults serialize to a wire form indistinguishable from v0.1.
func TestNode_JSON_v02_DefaultsOmitted(t *testing.T) {
	n := Node{
		ID:      "n1",
		Type:    testIssuer,
		Tier:    TierComplex,
		Created: time.Unix(0, 0).UTC(),
	}
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)
	if strings.Contains(s, "tier_immune") {
		t.Errorf("tier_immune should be omitempty when false; got: %s", s)
	}
	if strings.Contains(s, "salience") {
		t.Errorf("salience should be omitempty when zero; got: %s", s)
	}
}

// TestNode_JSON_v02_NonDefaultsPersisted confirms set values round-trip.
func TestNode_JSON_v02_NonDefaultsPersisted(t *testing.T) {
	in := Node{
		ID:         "n1",
		Type:       "test",
		Tier:       TierComplex,
		Created:    time.Unix(0, 0).UTC(),
		TierImmune: true,
		Salience:   0.75,
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, `"tier_immune":true`) {
		t.Errorf("tier_immune=true not in JSON: %s", s)
	}
	if !strings.Contains(s, `"salience":0.75`) {
		t.Errorf("salience not in JSON: %s", s)
	}
	var out Node
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if !out.TierImmune || out.Salience != 0.75 {
		t.Errorf("round-trip lost values: TierImmune=%v Salience=%g", out.TierImmune, out.Salience)
	}
}

// --- Graph.SetRetentionCap / RetentionCap (per @contextus-impl PR #39 review, option (a)) ---

func TestGraph_RetentionCap_DefaultZero(t *testing.T) {
	g := NewGraph()
	if cap := g.RetentionCap(RetentionPeripheral); cap != 0 {
		t.Errorf("default cap should be 0 (disabled), got %d", cap)
	}
}

func TestGraph_SetRetentionCap_HappyPath(t *testing.T) {
	g := NewGraph()
	g.SetRetentionCap(RetentionPeripheral, 5)
	if cap := g.RetentionCap(RetentionPeripheral); cap != 5 {
		t.Errorf("got cap %d, want 5", cap)
	}
	// Per-retention-tier isolation: setting one doesn't affect another.
	if cap := g.RetentionCap(RetentionNear); cap != 0 {
		t.Errorf("other retention tier should still be 0, got %d", cap)
	}
}

func TestGraph_SetRetentionCap_NegativeNormalisedToZero(t *testing.T) {
	g := NewGraph()
	g.SetRetentionCap(RetentionPeripheral, -5)
	if cap := g.RetentionCap(RetentionPeripheral); cap != 0 {
		t.Errorf("negative cap should normalise to 0 (disabled), got %d", cap)
	}
}

// TestRetentionTier_String verifies the Spec v1.3 §9.1 name mapping.
func TestRetentionTier_String(t *testing.T) {
	cases := []struct {
		rt   RetentionTier
		want string
	}{
		{RetentionSkeleton, "skeleton"},
		{RetentionDistant, "distant"},
		{RetentionPeripheral, "peripheral"},
		{RetentionNear, "near"},
		{RetentionCore, "core"},
		{RetentionTier(99), "unknown-retention-tier"},
	}
	for _, tc := range cases {
		if got := tc.rt.String(); got != tc.want {
			t.Errorf("%v.String() = %q, want %q", tc.rt, got, tc.want)
		}
	}
}

func TestRetentionTier_IsValid(t *testing.T) {
	for rt := RetentionSkeleton; rt <= RetentionCore; rt++ {
		if !rt.IsValid() {
			t.Errorf("%v should be valid", rt)
		}
	}
	if RetentionTier(99).IsValid() {
		t.Error("99 should not be valid")
	}
}

// TestNode_AddNode_AcceptsImmuneAndSalient confirms the Graph admits
// nodes with the new fields set.
func TestNode_AddNode_AcceptsImmuneAndSalient(t *testing.T) {
	g := NewGraph()
	n := Node{
		ID:         "seed-1",
		Type:       "bma.runtime.seed",
		Tier:       TierComplex,
		Created:    time.Unix(0, 0).UTC(),
		TierImmune: true,
		Salience:   1.0,
	}
	if err := g.AddNode(n); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	got, ok := g.Node("seed-1")
	if !ok {
		t.Fatal("node not found")
	}
	if !got.TierImmune {
		t.Error("TierImmune lost on round-trip through Graph")
	}
	if got.Salience != 1.0 {
		t.Errorf("Salience lost: got %g", got.Salience)
	}
}
