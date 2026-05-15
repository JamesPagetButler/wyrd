package model

import (
	"slices"
	"testing"
	"time"
)

// TestBMAPolicy_AllTD4Entries verifies every TD-4 inventory entry
// (@bma-implementor live-test seq=99) plus the W-Toddle-2-extension
// entries (Marcy #toddle-design seq=24 constitutional prerequisite)
// has the documented policy.
func TestBMAPolicy_AllTD4Entries(t *testing.T) {
	cases := []struct {
		typ      NodeType
		immune   bool
		salience float64
	}{
		// Original TD-4 inventory.
		{NodeTypeBMASeed, true, 1.0},
		{NodeTypeBMALifeCertificate, true, 1.0},
		{NodeTypeBMADeathCertificate, true, 1.0},
		{NodeTypeBMAObservation, false, 0.0},
		{NodeTypeBMAParamProposal, true, 1.0},
		{NodeTypeBMAParamTrustState, true, 1.0},
		{NodeTypeBMALastWords, true, 1.0},
		{NodeTypeBMAEulogy, true, 1.0},
		// W-Toddle-2-extension (Marcy constitutional prerequisite).
		{NodeTypeBMAIdentity, true, 1.0},
		{NodeTypeBMAMemorial, true, 1.0},
		// W-Toddle-2-extension (semantic-memory, decay-eligible).
		{NodeTypeBMAEntity, false, 0.0},
		{NodeTypeBMAConcept, false, 0.0},
		{NodeTypeBMAPattern, false, 0.0},
	}
	for _, tc := range cases {
		t.Run(string(tc.typ), func(t *testing.T) {
			immune, salience, known := BMAPolicy(tc.typ)
			if !known {
				t.Fatalf("BMAPolicy(%q) returned known=false; expected registered", tc.typ)
			}
			if immune != tc.immune {
				t.Errorf("BMAPolicy(%q) immune = %v, want %v", tc.typ, immune, tc.immune)
			}
			if salience != tc.salience {
				t.Errorf("BMAPolicy(%q) salience = %g, want %g", tc.typ, salience, tc.salience)
			}
		})
	}
}

func TestBMAPolicy_UnknownType(t *testing.T) {
	immune, salience, known := BMAPolicy("not-a-bma-type")
	if known {
		t.Error("expected known=false for unregistered type")
	}
	if immune {
		t.Error("expected immune=false for unregistered type")
	}
	if salience != 0.0 {
		t.Errorf("expected salience=0.0 for unregistered type, got %g", salience)
	}
}

func TestApplyBMAPolicy_NilNoop(t *testing.T) {
	// Must not panic.
	ApplyBMAPolicy(nil)
}

func TestApplyBMAPolicy_UnknownNodeTypeNoop(t *testing.T) {
	n := Node{
		ID:      "n1",
		Type:    "not-a-bma-type",
		Tier:    TierComplex,
		Created: time.Unix(0, 0),
		// Pre-existing values that must NOT be modified for unknown types.
		TierImmune: true,
		Salience:   0.42,
	}
	ApplyBMAPolicy(&n)
	if !n.TierImmune {
		t.Error("ApplyBMAPolicy mutated TierImmune for unknown type")
	}
	if n.Salience != 0.42 {
		t.Errorf("ApplyBMAPolicy mutated Salience for unknown type: got %g, want 0.42", n.Salience)
	}
}

func TestApplyBMAPolicy_SeedSetsImmuneAndMaxSalience(t *testing.T) {
	n := Node{
		ID:      "seed-1",
		Type:    NodeTypeBMASeed,
		Tier:    TierComplex,
		Created: time.Unix(0, 0),
	}
	ApplyBMAPolicy(&n)
	if !n.TierImmune {
		t.Error("seed node should be TierImmune=true")
	}
	if n.Salience != 1.0 {
		t.Errorf("seed node should have Salience=1.0, got %g", n.Salience)
	}
}

func TestApplyBMAPolicy_ObservationStaysDecayEligible(t *testing.T) {
	n := Node{
		ID:      "obs-1",
		Type:    NodeTypeBMAObservation,
		Tier:    TierComplex,
		Created: time.Unix(0, 0),
		// Pre-existing values that ApplyBMAPolicy SHOULD overwrite.
		TierImmune: true,
		Salience:   0.99,
	}
	ApplyBMAPolicy(&n)
	if n.TierImmune {
		t.Error("observation node should NOT be TierImmune; got true")
	}
	if n.Salience != 0.0 {
		t.Errorf("observation node should reset Salience to 0.0, got %g", n.Salience)
	}
}

// TestApplyBMAPolicy_Idempotent — applying twice yields the same result
// as applying once. Important for the BMA hg/ shim which may call this
// at multiple write sites.
func TestApplyBMAPolicy_Idempotent(t *testing.T) {
	for _, typ := range BMAPolicyNodeTypes() {
		t.Run(string(typ), func(t *testing.T) {
			n := Node{ID: "x", Type: typ, Tier: TierComplex, Created: time.Unix(0, 0)}
			ApplyBMAPolicy(&n)
			first := n
			ApplyBMAPolicy(&n)
			if n.TierImmune != first.TierImmune || n.Salience != first.Salience {
				t.Errorf("not idempotent: first=(%v, %g), second=(%v, %g)",
					first.TierImmune, first.Salience, n.TierImmune, n.Salience)
			}
		})
	}
}

func TestBMAPolicyNodeTypes_CountMatchesTD4(t *testing.T) {
	got := BMAPolicyNodeTypes()
	// 8 original TD-4 entries + 5 W-Toddle-2-extension entries
	// (Identity + Memorial immune; Entity + Concept + Pattern decay).
	if len(got) != 13 {
		t.Errorf("expected 13 BMA policy entries (8 TD-4 + 5 extension), got %d: %v", len(got), got)
	}
	// Verify all expected types are present (order-independent).
	want := []NodeType{
		NodeTypeBMASeed,
		NodeTypeBMALifeCertificate,
		NodeTypeBMADeathCertificate,
		NodeTypeBMAObservation,
		NodeTypeBMAParamProposal,
		NodeTypeBMAParamTrustState,
		NodeTypeBMALastWords,
		NodeTypeBMAEulogy,
		NodeTypeBMAIdentity,
		NodeTypeBMAMemorial,
		NodeTypeBMAEntity,
		NodeTypeBMAConcept,
		NodeTypeBMAPattern,
	}
	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Errorf("missing BMA policy entry: %s", w)
		}
	}
}

// TestApplyBMAPolicy_IntegrationWithGraph confirms the policy works
// end-to-end: ApplyBMAPolicy then AddNode lands the immunity flag in
// the graph as expected.
func TestApplyBMAPolicy_IntegrationWithGraph(t *testing.T) {
	g := NewGraph()
	n := Node{
		ID:      "seed-protocol-step-9",
		Type:    NodeTypeBMASeed,
		Tier:    TierComplex,
		Created: time.Unix(0, 0),
	}
	ApplyBMAPolicy(&n)
	if err := g.AddNode(n); err != nil {
		t.Fatalf("AddNode after ApplyBMAPolicy: %v", err)
	}
	got, ok := g.Node("seed-protocol-step-9")
	if !ok {
		t.Fatal("node not found in graph")
	}
	if !got.TierImmune {
		t.Error("seed node should be TierImmune after policy + graph round-trip")
	}
	if got.Salience != 1.0 {
		t.Errorf("seed Salience should be 1.0 after policy + graph round-trip, got %g", got.Salience)
	}
}
