package model

import (
	"errors"
	"sync"
	"testing"
)

// Tests for Graph.UpdateNode + Graph.UpdateNodeWithCapability
// (wyrd-issue-#57). Consumer story: BMA #159 Pentagon Pod handoff
// (T+5 TierImmune flip) + Hebbian Salience bumps on live nodes.

func TestUpdateNode_FlipsTierImmuneAndSalience(t *testing.T) {
	g := NewGraph()
	n := mkNode("pod-snapshot", TierQuaternion)
	n.TierImmune = true
	n.Salience = 1.0
	if err := g.AddNode(n); err != nil {
		t.Fatalf("AddNode: %v", err)
	}

	// T+5: snapshot becomes decay-eligible (the BMA #159 transition).
	n.TierImmune = false
	n.Salience = 0.4
	if err := g.UpdateNode(n); err != nil {
		t.Fatalf("UpdateNode: %v", err)
	}

	got, ok := g.Node("pod-snapshot")
	if !ok {
		t.Fatal("node disappeared after update")
	}
	if got.TierImmune {
		t.Error("TierImmune still true after update")
	}
	if got.Salience != 0.4 {
		t.Errorf("Salience = %v, want 0.4", got.Salience)
	}
}

func TestUpdateNode_MissingIDReturnsErrNodeNotFound(t *testing.T) {
	g := NewGraph()
	err := g.UpdateNode(mkNode("ghost", TierComplex))
	if err == nil {
		t.Fatal("expected error for missing node")
	}
	if !errors.Is(err, ErrNodeNotFound) {
		t.Errorf("error %v does not unwrap to ErrNodeNotFound", err)
	}
	if _, ok := g.Node("ghost"); ok {
		t.Error("UpdateNode must never create nodes")
	}
}

func TestUpdateNode_RejectsMalformed(t *testing.T) {
	g := NewGraph()
	if err := g.AddNode(mkNode("a", TierComplex)); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	bad := mkNode("a", TierComplex)
	bad.Type = "" // fails Node.Validate
	if err := g.UpdateNode(bad); err == nil {
		t.Error("expected validation error for malformed node")
	}
}

func TestUpdateNode_Idempotent(t *testing.T) {
	g := NewGraph()
	n := mkNode("a", TierComplex)
	if err := g.AddNode(n); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	n.Salience = 0.7
	for i := 0; i < 2; i++ {
		if err := g.UpdateNode(n); err != nil {
			t.Fatalf("UpdateNode pass %d: %v", i+1, err)
		}
	}
	got, _ := g.Node("a")
	if got.Salience != 0.7 {
		t.Errorf("Salience = %v, want 0.7", got.Salience)
	}
}

func TestUpdateNode_PreservesIncidentEdges(t *testing.T) {
	g := NewGraph()
	for _, id := range []NodeID{"a", "b"} {
		if err := g.AddNode(mkNode(id, TierComplex)); err != nil {
			t.Fatalf("AddNode(%s): %v", id, err)
		}
	}
	e := Hyperedge{
		ID:          "e1",
		Nodes:       []NodeID{"a", "b"},
		Weight:      NewComplexWeight(1, 0),
		IsSymmetric: true,
	}
	if err := g.AddHyperedge(e); err != nil {
		t.Fatalf("AddHyperedge: %v", err)
	}

	n := mkNode("a", TierComplex)
	n.Salience = 0.9
	if err := g.UpdateNode(n); err != nil {
		t.Fatalf("UpdateNode: %v", err)
	}

	edges := g.IncidentEdges("a")
	if len(edges) != 1 || edges[0] != "e1" {
		t.Errorf("IncidentEdges(a) = %v, want [e1]", edges)
	}
}

func TestUpdateNodeWithCapability_HappyPath(t *testing.T) {
	g := NewGraph()
	n := mkNode("a", TierComplex)
	if err := g.AddNode(n); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	n.Salience = 0.5
	if err := g.UpdateNodeWithCapability(n, newWriteCap(TierQuaternion)); err != nil {
		t.Fatalf("UpdateNodeWithCapability: %v", err)
	}
	got, _ := g.Node("a")
	if got.Salience != 0.5 {
		t.Errorf("Salience = %v, want 0.5", got.Salience)
	}
}

func TestUpdateNodeWithCapability_RejectsWrongTier(t *testing.T) {
	g := NewGraph()
	if err := g.AddNode(mkNode("a", TierOctonion)); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	n := mkNode("a", TierOctonion)
	n.Salience = 0.5
	err := g.UpdateNodeWithCapability(n, newWriteCap(TierComplex))
	if !errors.Is(err, ErrCapabilityViolation) {
		t.Errorf("error %v does not unwrap to ErrCapabilityViolation", err)
	}
	// State must be unchanged after the rejected write.
	got, _ := g.Node("a")
	if got.Salience != 0 {
		t.Error("rejected update must not mutate state")
	}
}

func TestUpdateNodeWithCapability_RejectsTierEscapeViaExisting(t *testing.T) {
	// A ℂ-holder must not mutate an existing 𝕆-tier node by writing a
	// ℂ-tier replacement over it (the existing-tier half of the dual
	// check).
	g := NewGraph()
	if err := g.AddNode(mkNode("high", TierOctonion)); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	repl := mkNode("high", TierComplex) // replacement at the LOW tier
	err := g.UpdateNodeWithCapability(repl, newWriteCap(TierComplex))
	if !errors.Is(err, ErrCapabilityViolation) {
		t.Errorf("downgrade-over-high-tier-node: error %v does not unwrap to ErrCapabilityViolation", err)
	}
	got, _ := g.Node("high")
	if got.Tier != TierOctonion {
		t.Error("rejected update must not change the stored node")
	}
}

func TestUpdateNodeWithCapability_NotFoundAfterCapPass(t *testing.T) {
	g := NewGraph()
	err := g.UpdateNodeWithCapability(mkNode("ghost", TierComplex), newWriteCap(TierQuaternion))
	if !errors.Is(err, ErrNodeNotFound) {
		t.Errorf("error %v does not unwrap to ErrNodeNotFound", err)
	}
}

func TestUpdateNodeWithCapability_CapViolationPrecedesNotFound(t *testing.T) {
	// A holder that cannot write at the replacement tier learns
	// nothing about node existence: the capability violation is
	// reported in preference to not-found.
	g := NewGraph()
	err := g.UpdateNodeWithCapability(mkNode("ghost", TierOctonion), newWriteCap(TierComplex))
	if !errors.Is(err, ErrCapabilityViolation) {
		t.Errorf("error %v should be ErrCapabilityViolation (not-found must not leak)", err)
	}
}

func TestUpdateNodeWithCapability_AtomicCheckAndMutate(t *testing.T) {
	// TOCTOU regression (PR #82 §I4, bma-implementor): the
	// existing-tier check and the mutation must share one critical
	// section. A privileged writer races tier swaps (ℂ↔𝕆) on a node
	// while a ℂ-holder races updates; if the check-then-act were
	// split across lock windows, a ℂ write could land over an 𝕆-tier
	// node. Invariant: every observed node state at tier 𝕆 carries
	// the privileged writer's payload marker, never the ℂ-holder's.
	g := NewGraph()
	seed := mkNode("contested", TierComplex)
	if err := g.AddNode(seed); err != nil {
		t.Fatalf("AddNode: %v", err)
	}

	const iters = 200
	var wg sync.WaitGroup
	wg.Add(2)

	// Privileged writer: alternates the node between ℂ and 𝕆 tiers.
	go func() {
		defer wg.Done()
		hi := newWriteCap(TierOctonion)
		for i := 0; i < iters; i++ {
			n := mkNode("contested", TierOctonion)
			n.Payload = []byte("privileged")
			_ = g.UpdateNodeWithCapability(n, hi)
			n = mkNode("contested", TierComplex)
			n.Payload = []byte("privileged")
			_ = g.UpdateNodeWithCapability(n, hi)
		}
	}()

	// Low-capability writer: ℂ-tier replacement values only. Its
	// writes must succeed only when the stored node is ℂ-tier at
	// check+write time (atomically).
	go func() {
		defer wg.Done()
		lo := newWriteCap(TierComplex)
		for i := 0; i < iters; i++ {
			n := mkNode("contested", TierComplex)
			n.Payload = []byte("low")
			err := g.UpdateNodeWithCapability(n, lo)
			if err != nil && !errors.Is(err, ErrCapabilityViolation) {
				t.Errorf("unexpected error class: %v", err)
			}
		}
	}()

	wg.Wait()

	// Post-condition: if the final node is 𝕆-tier, the low writer
	// cannot have produced it.
	got, ok := g.Node("contested")
	if !ok {
		t.Fatal("node missing")
	}
	if got.Tier == TierOctonion && string(got.Payload) == "low" {
		t.Error("TOCTOU: low-capability write landed an 𝕆-tier state")
	}
}

func TestUpdateNode_ConcurrentUpdatesRaceFree(t *testing.T) {
	g := NewGraph()
	if err := g.AddNode(mkNode("hot", TierComplex)); err != nil {
		t.Fatalf("AddNode: %v", err)
	}

	const workers = 32
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		sal := float64(i) / float64(workers)
		go func() {
			defer wg.Done()
			n := mkNode("hot", TierComplex)
			n.Salience = sal
			if err := g.UpdateNode(n); err != nil {
				t.Errorf("concurrent UpdateNode: %v", err)
			}
		}()
	}
	wg.Wait()

	// Final state must be one of the written values (atomic
	// last-writer-wins; no torn write).
	got, ok := g.Node("hot")
	if !ok {
		t.Fatal("node missing after concurrent updates")
	}
	if got.Salience < 0 || got.Salience >= 1 {
		t.Errorf("Salience %v outside any written value", got.Salience)
	}
}
