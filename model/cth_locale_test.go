package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCTHLocaleNodeTypes_Values(t *testing.T) {
	want := map[NodeType]string{
		NodeTypeCTHLocaleOpenSet:         "cth.locale.open-set",
		NodeTypeCTHLocaleTrustSection:    "cth.locale.trust-section",
		NodeTypeCTHLocaleGluingCoherence: "cth.locale.gluing-coherence",
	}
	for nt, s := range want {
		if string(nt) != s {
			t.Errorf("NodeType value = %q, want %q", string(nt), s)
		}
	}
}

func TestCTHLocaleNodeTypes_Enumeration(t *testing.T) {
	got := CTHLocaleNodeTypes()
	want := []NodeType{
		NodeTypeCTHLocaleOpenSet,      // opens
		NodeTypeCTHLocaleTrustSection, // sections
		NodeTypeCTHLocaleGluingCoherence,
	}
	if len(got) != len(want) {
		t.Fatalf("CTHLocaleNodeTypes() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("CTHLocaleNodeTypes()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// cth.locale.* are consumer-owned Types; they must not use the reserved
// `wyrd.` prefix (node.go NodeType convention).
func TestCTHLocaleNodeTypes_ConsumerOwnedNotReserved(t *testing.T) {
	for _, nt := range CTHLocaleNodeTypes() {
		if strings.HasPrefix(string(nt), "wyrd.") {
			t.Errorf("%q uses the reserved wyrd. prefix", nt)
		}
		if !strings.HasPrefix(string(nt), "cth.locale.") {
			t.Errorf("%q is not under the cth.locale. namespace", nt)
		}
	}
}

// A node carrying a locale Type validates, is not TierImmune by default
// (locale sections/records follow default eviction), and JSON round-trips
// (v0.1 wire-format: additive, no new fields on Node).
func TestCTHLocaleNode_ValidatesAndRoundTrips(t *testing.T) {
	for _, nt := range CTHLocaleNodeTypes() {
		n := Node{ID: NodeID("cth:locale:" + string(nt)), Type: nt}
		if err := n.Validate(); err != nil {
			t.Fatalf("Validate() for %q = %v, want nil", nt, err)
		}
		if n.TierImmune {
			t.Errorf("%q node should not be TierImmune by default", nt)
		}

		b, err := json.Marshal(n)
		if err != nil {
			t.Fatalf("Marshal(%q) = %v", nt, err)
		}
		var got Node
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatalf("Unmarshal(%q) = %v", nt, err)
		}
		if got.Type != nt {
			t.Errorf("round-trip Type = %q, want %q", got.Type, nt)
		}
	}
}
