package model

import (
	"errors"
	"testing"
	"time"
)

func newWriteCap(tier Tier) WriteCapability {
	return WriteCapability{HolderTier: tier, GrantedAt: time.Unix(0, 0), Issuer: "test"}
}

func newReadCap(tier Tier) ReadCapability {
	return ReadCapability{HolderTier: tier, GrantedAt: time.Unix(0, 0), Issuer: "test"}
}

func TestWriteCapability_AllowsWrite_DownwardOK(t *testing.T) {
	cases := []struct {
		holder, target Tier
	}{
		{TierQuaternion, TierComplex},
		{TierOctonion, TierQuaternion},
		{TierSedenion, TierOctonion},
		{TierSedenion, TierComplex},
		{TierComplex, TierComplex},
		{TierSedenion, TierSedenion},
	}
	for _, c := range cases {
		cap := newWriteCap(c.holder)
		if err := cap.AllowsWrite(c.target); err != nil {
			t.Errorf("AllowsWrite(holder=%s, target=%s): unexpected error %v", c.holder, c.target, err)
		}
	}
}

func TestWriteCapability_AllowsWrite_UpwardBlocked(t *testing.T) {
	cases := []struct {
		holder, target Tier
	}{
		{TierComplex, TierQuaternion},  // T2.1.a — ℂ ↛ ℍ
		{TierQuaternion, TierOctonion}, // T2.1.b — ℍ ↛ 𝕆
		{TierOctonion, TierSedenion},   // T2.1.c — 𝕆 ↛ 𝕊
		{TierComplex, TierSedenion},    // transitive
	}
	for _, c := range cases {
		cap := newWriteCap(c.holder)
		err := cap.AllowsWrite(c.target)
		if err == nil {
			t.Errorf("AllowsWrite(%s → %s): expected capability violation", c.holder, c.target)
			continue
		}
		if !errors.Is(err, ErrCapabilityViolation) {
			t.Errorf("AllowsWrite(%s → %s): error %v not ErrCapabilityViolation", c.holder, c.target, err)
		}
		var capErr CapabilityError
		if !errors.As(err, &capErr) {
			t.Errorf("AllowsWrite(%s → %s): error %v not CapabilityError", c.holder, c.target, err)
		} else if capErr.Mode != CapabilityModeWrite {
			t.Errorf("AllowsWrite: error mode = %s, want write", capErr.Mode)
		}
	}
}

func TestReadCapability_AllowsRead_SameAlgebraAsWrite(t *testing.T) {
	for _, c := range []struct {
		holder, target Tier
		wantErr        bool
	}{
		{TierComplex, TierQuaternion, true},
		{TierQuaternion, TierComplex, false},
		{TierSedenion, TierSedenion, false},
	} {
		got := newReadCap(c.holder).AllowsRead(c.target)
		if (got != nil) != c.wantErr {
			t.Errorf("AllowsRead(%s → %s): err=%v, wantErr=%v", c.holder, c.target, got, c.wantErr)
		}
	}
}

func TestCapabilityError_ReadMode(t *testing.T) {
	rerr := newReadCap(TierComplex).AllowsRead(TierQuaternion)
	var capErr CapabilityError
	if !errors.As(rerr, &capErr) {
		t.Fatalf("expected CapabilityError, got %T", rerr)
	}
	if capErr.Mode != CapabilityModeRead {
		t.Errorf("read error mode = %s, want read", capErr.Mode)
	}
}

func TestCapabilityCheck_RejectsInvalidTier(t *testing.T) {
	bad := Tier(99)
	if err := newWriteCap(bad).AllowsWrite(TierComplex); err == nil {
		t.Error("AllowsWrite with invalid holder tier: expected error")
	}
	if err := newWriteCap(TierComplex).AllowsWrite(bad); err == nil {
		t.Error("AllowsWrite with invalid target tier: expected error")
	}
}

// --- Graph integration tests for *WithCapability methods ---

func TestGraph_AddNodeWithCapability_PrivilegeRespected(t *testing.T) {
	g := NewGraph()
	cap := newWriteCap(TierQuaternion)

	if err := g.AddNodeWithCapability(mkNode("ok", TierComplex), cap); err != nil {
		t.Errorf("AddNodeWithCapability (complex with quaternion holder): %v", err)
	}

	err := g.AddNodeWithCapability(mkNode("blocked", TierOctonion), cap)
	if err == nil {
		t.Fatal("AddNodeWithCapability (octonion with quaternion holder): expected violation")
	}
	if !errors.Is(err, ErrCapabilityViolation) {
		t.Errorf("expected ErrCapabilityViolation, got %v", err)
	}
	if _, ok := g.Node("blocked"); ok {
		t.Error("blocked node was inserted despite capability violation")
	}
}

func TestGraph_AddHyperedgeWithCapability_PrivilegeRespected(t *testing.T) {
	g := NewGraph()
	cap := newWriteCap(TierQuaternion)
	for _, id := range []NodeID{"a", "b"} {
		_ = g.AddNodeWithCapability(mkNode(id, TierComplex), cap)
	}
	e := Hyperedge{
		ID:    "e1",
		Nodes: []NodeID{"a", "b"},
		Weight: Weight{
			Tier:       TierOctonion,
			Components: [16]float64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		Created: time.Unix(0, 0),
	}
	err := g.AddHyperedgeWithCapability(e, cap)
	if err == nil {
		t.Fatal("AddHyperedgeWithCapability (octonion edge, quaternion holder): expected violation")
	}
	if !errors.Is(err, ErrCapabilityViolation) {
		t.Errorf("expected ErrCapabilityViolation, got %v", err)
	}
	if _, ok := g.Hyperedge("e1"); ok {
		t.Error("blocked edge was inserted despite capability violation")
	}
}

func TestGraph_RemoveHyperedgeWithCapability_PrivilegeRespected(t *testing.T) {
	g := NewGraph()
	for _, id := range []NodeID{"a", "b"} {
		_ = g.AddNode(mkNode(id, TierQuaternion))
	}
	e := Hyperedge{
		ID:     "octonion-edge",
		Nodes:  []NodeID{"a", "b"},
		Weight: Weight{Tier: TierOctonion},
	}
	if err := g.AddHyperedge(e); err != nil {
		t.Fatalf("setup: %v", err)
	}
	err := g.RemoveHyperedgeWithCapability("octonion-edge", newWriteCap(TierQuaternion))
	if err == nil {
		t.Fatal("RemoveHyperedgeWithCapability (octonion edge, quaternion holder): expected violation")
	}
	if !errors.Is(err, ErrCapabilityViolation) {
		t.Errorf("expected ErrCapabilityViolation, got %v", err)
	}
	if _, ok := g.Hyperedge("octonion-edge"); !ok {
		t.Error("edge was removed despite capability violation")
	}
	if err := g.RemoveHyperedgeWithCapability("octonion-edge", newWriteCap(TierOctonion)); err != nil {
		t.Fatalf("RemoveHyperedgeWithCapability with octonion cap: %v", err)
	}
}

func TestGraph_RemoveHyperedgeWithCapability_NotFound(t *testing.T) {
	g := NewGraph()
	err := g.RemoveHyperedgeWithCapability("ghost", newWriteCap(TierComplex))
	if err == nil {
		t.Error("expected error for non-existent edge")
	}
	if errors.Is(err, ErrCapabilityViolation) {
		t.Errorf("not-found should not be a CapabilityError; got %v", err)
	}
}
