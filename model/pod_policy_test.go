package model

import (
	"testing"
)

func TestPodPolicy_KnownTypes(t *testing.T) {
	cases := []struct {
		name    string
		typ     NodeType
		wantImm bool
		wantSal float64
	}{
		{"conscious-a", NodeTypePodConsciousA, true, 1.0},
		{"conscious-b", NodeTypePodConsciousB, true, 1.0},
		{"subconscious-l", NodeTypePodSubconsciousL, true, 0.9},
		{"subconscious-r", NodeTypePodSubconsciousR, true, 0.9},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			imm, sal, known := PodPolicy(tc.typ)
			if !known {
				t.Fatalf("PodPolicy(%q) returned known=false", tc.typ)
			}
			if imm != tc.wantImm {
				t.Errorf("TierImmune: got %v, want %v", imm, tc.wantImm)
			}
			if sal != tc.wantSal {
				t.Errorf("Salience: got %v, want %v", sal, tc.wantSal)
			}
		})
	}
}

func TestPodPolicy_UnknownType(t *testing.T) {
	imm, sal, known := PodPolicy("bma.pod.dev")
	if known {
		t.Fatal("expected known=false for Walk-phase dev pod")
	}
	if imm || sal != 0.0 {
		t.Errorf("expected (false, 0.0) for unknown type, got (%v, %v)", imm, sal)
	}
}

func TestApplyPodPolicy_SetsFields(t *testing.T) {
	n := &Node{Type: NodeTypePodConsciousA, TierImmune: false, Salience: 0.0}
	ApplyPodPolicy(n)
	if !n.TierImmune {
		t.Error("expected TierImmune=true after ApplyPodPolicy")
	}
	if n.Salience != 1.0 {
		t.Errorf("expected Salience=1.0, got %v", n.Salience)
	}
}

func TestApplyPodPolicy_NoopForUnknown(t *testing.T) {
	n := &Node{Type: "bma.observation", TierImmune: false, Salience: 0.5}
	ApplyPodPolicy(n)
	if n.TierImmune || n.Salience != 0.5 {
		t.Error("ApplyPodPolicy should not modify node with non-pod type")
	}
}

func TestApplyPodPolicy_NilSafe(t *testing.T) {
	ApplyPodPolicy(nil) // must not panic
}

func TestPodPolicyNodeTypes_Count(t *testing.T) {
	types := PodPolicyNodeTypes()
	if len(types) != 4 {
		t.Errorf("expected 4 pod types, got %d", len(types))
	}
}
