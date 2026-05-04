package model

import "testing"

func TestWeight_Validate(t *testing.T) {
	w := NewQuaternionWeight(1, 2, 3, 4)
	if err := w.Validate(); err != nil {
		t.Errorf("valid quaternion weight rejected: %v", err)
	}

	// Tampered: a quaternion weight with non-zero octonion-component.
	w.Components[5] = 0.1
	if err := w.Validate(); err == nil {
		t.Error("expected error: components beyond tier dimension non-zero")
	}
}

func TestWeight_Active(t *testing.T) {
	w := NewQuaternionWeight(1, 2, 3, 4)
	if got := len(w.Active()); got != 4 {
		t.Errorf("Active() len = %d, want 4", got)
	}
	if w.Active()[0] != 1 || w.Active()[3] != 4 {
		t.Errorf("Active() values wrong: %v", w.Active())
	}
}

func TestWeight_IsZero(t *testing.T) {
	if !(Weight{Tier: TierComplex}).IsZero() {
		t.Error("zero complex weight reported non-zero")
	}
	if NewComplexWeight(0, 1).IsZero() {
		t.Error("non-zero complex weight reported zero")
	}
}
