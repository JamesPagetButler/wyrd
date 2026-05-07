package model

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// missingID is a sentinel HyperedgeID/NodeID that is never added to test
// graphs, used to drive missing-edge / unknown-node error paths.
const missingID = "ghost"

// --- PromoteBatch happy-path + error-cases ---

func setupBatchGraphs(t *testing.T) (*Graph, *Graph) {
	t.Helper()
	src := NewGraph()
	dst := NewGraph()
	for _, id := range []NodeID{"a", "b", "c", "d"} {
		if err := src.AddNode(mkNode(id, TierComplex)); err != nil {
			t.Fatalf("setup src node %s: %v", id, err)
		}
		if err := dst.AddNode(mkNode(id, TierComplex)); err != nil {
			t.Fatalf("setup dst node %s: %v", id, err)
		}
	}
	for i, ns := range [][]NodeID{
		{"a", "b"},
		{"b", "c"},
		{"c", "d"},
	} {
		eid := HyperedgeID(fmt.Sprintf("e%d", i))
		if err := src.AddHyperedge(mkEdge(eid, ns, TierComplex)); err != nil {
			t.Fatalf("setup edge %s: %v", eid, err)
		}
	}
	return src, dst
}

func TestPromoteBatch_HappyPath(t *testing.T) {
	src, dst := setupBatchGraphs(t)
	ids := []HyperedgeID{"e0", "e1", "e2"}
	preTotal := src.EdgeCount() + dst.EdgeCount()

	if err := src.PromoteBatch(dst, ids); err != nil {
		t.Fatalf("PromoteBatch: %v", err)
	}

	if src.EdgeCount() != 0 {
		t.Errorf("source edge count = %d, want 0", src.EdgeCount())
	}
	if dst.EdgeCount() != 3 {
		t.Errorf("dst edge count = %d, want 3", dst.EdgeCount())
	}
	if src.EdgeCount()+dst.EdgeCount() != preTotal {
		t.Errorf("count not conserved: pre=%d post=%d", preTotal, src.EdgeCount()+dst.EdgeCount())
	}
}

func TestPromoteBatch_AllOrNothing_MissingEdge(t *testing.T) {
	src, dst := setupBatchGraphs(t)
	srcCountBefore := src.EdgeCount()
	dstCountBefore := dst.EdgeCount()

	// Include a non-existent edge mid-batch.
	err := src.PromoteBatch(dst, []HyperedgeID{"e0", missingID, "e2"})
	if !errors.Is(err, ErrBatchEdgeNotFound) {
		t.Fatalf("expected ErrBatchEdgeNotFound, got %v", err)
	}

	// All-or-nothing: nothing should have moved.
	if src.EdgeCount() != srcCountBefore {
		t.Errorf("source mutated: count %d → %d", srcCountBefore, src.EdgeCount())
	}
	if dst.EdgeCount() != dstCountBefore {
		t.Errorf("dst mutated: count %d → %d", dstCountBefore, dst.EdgeCount())
	}
}

func TestPromoteBatch_AlreadyExistsInDest(t *testing.T) {
	src, dst := setupBatchGraphs(t)
	// Make e0 exist in dst already.
	_ = dst.AddHyperedge(mkEdge("e0", []NodeID{"a", "b"}, TierComplex))

	err := src.PromoteBatch(dst, []HyperedgeID{"e0"})
	if !errors.Is(err, ErrBatchEdgeAlreadyExists) {
		t.Fatalf("expected ErrBatchEdgeAlreadyExists, got %v", err)
	}
	if _, ok := src.Hyperedge("e0"); !ok {
		t.Error("source edge was removed despite preflight failure")
	}
}

func TestPromoteBatch_MissingNodeInDest(t *testing.T) {
	src, dst := setupBatchGraphs(t)
	// Add a fresh source-only node; create an edge referencing it.
	_ = src.AddNode(mkNode("source-only", TierComplex))
	_ = src.AddHyperedge(mkEdge("e-orphan", []NodeID{"a", "source-only"}, TierComplex))

	err := src.PromoteBatch(dst, []HyperedgeID{"e-orphan"})
	if !errors.Is(err, ErrBatchMissingNode) {
		t.Fatalf("expected ErrBatchMissingNode, got %v", err)
	}
	// State unchanged.
	if _, ok := src.Hyperedge("e-orphan"); !ok {
		t.Error("source edge removed despite missing-node error")
	}
}

func TestPromoteBatch_NilGuards(t *testing.T) {
	g := NewGraph()
	if err := g.PromoteBatch(nil, nil); err == nil {
		t.Error("PromoteBatch(nil): expected error")
	}
	if err := g.PromoteBatch(g, nil); err == nil {
		t.Error("PromoteBatch(self): expected error")
	}
	var nilG *Graph
	if err := nilG.PromoteBatch(g, nil); err == nil {
		t.Error("(nil).PromoteBatch: expected error")
	}
}

// --- RemoveBatch ---

func TestRemoveBatch_HappyPath(t *testing.T) {
	g := NewGraph()
	for _, id := range []NodeID{"a", "b", "c"} {
		_ = g.AddNode(mkNode(id, TierComplex))
	}
	_ = g.AddHyperedge(mkEdge("e0", []NodeID{"a", "b"}, TierComplex))
	_ = g.AddHyperedge(mkEdge("e1", []NodeID{"b", "c"}, TierComplex))
	_ = g.AddHyperedge(mkEdge("e2", []NodeID{"a", "c"}, TierComplex))

	if err := g.RemoveBatch([]HyperedgeID{"e0", "e1", "e2"}); err != nil {
		t.Fatalf("RemoveBatch: %v", err)
	}
	if g.EdgeCount() != 0 {
		t.Errorf("edge count = %d, want 0", g.EdgeCount())
	}
}

func TestRemoveBatch_AllOrNothing_MissingEdge(t *testing.T) {
	g := NewGraph()
	for _, id := range []NodeID{"a", "b"} {
		_ = g.AddNode(mkNode(id, TierComplex))
	}
	_ = g.AddHyperedge(mkEdge("e0", []NodeID{"a", "b"}, TierComplex))
	pre := g.EdgeCount()

	err := g.RemoveBatch([]HyperedgeID{"e0", missingID})
	if !errors.Is(err, ErrBatchEdgeNotFound) {
		t.Fatalf("expected ErrBatchEdgeNotFound, got %v", err)
	}
	if g.EdgeCount() != pre {
		t.Errorf("edge count changed: %d → %d (should be unchanged)", pre, g.EdgeCount())
	}
}

// (Capability-gated batch tests land with the *WithCapability methods
// in the follow-up PR. The capability tier check is exercised in
// model/capability_test.go on the single-edge methods; the batch
// extension is mechanical.)

// --- Concurrent two-direction stress test (deadlock smoke) ---

// TestPromoteBatch_ConcurrentReverseDirection issues PromoteBatch calls
// in opposing directions concurrently. With pointer-address-ordered
// locking, no deadlock should occur. Run under -race.
func TestPromoteBatch_ConcurrentReverseDirection(t *testing.T) {
	t.Parallel()
	a := NewGraph()
	b := NewGraph()
	for _, id := range []NodeID{"x", "y"} {
		_ = a.AddNode(mkNode(id, TierComplex))
		_ = b.AddNode(mkNode(id, TierComplex))
	}

	const dur = 100 * time.Millisecond
	deadline := time.Now().Add(dur)
	var wg sync.WaitGroup

	// Goroutine 1: a → b promotions.
	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		for time.Now().Before(deadline) {
			eid := HyperedgeID(fmt.Sprintf("ab-%d", i))
			if err := a.AddHyperedge(mkEdge(eid, []NodeID{"x", "y"}, TierComplex)); err == nil {
				_ = a.PromoteBatch(b, []HyperedgeID{eid})
				_ = b.RemoveBatch([]HyperedgeID{eid})
			}
			i++
		}
	}()

	// Goroutine 2: b → a promotions (opposite direction).
	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		for time.Now().Before(deadline) {
			eid := HyperedgeID(fmt.Sprintf("ba-%d", i))
			if err := b.AddHyperedge(mkEdge(eid, []NodeID{"x", "y"}, TierComplex)); err == nil {
				_ = b.PromoteBatch(a, []HyperedgeID{eid})
				_ = a.RemoveBatch([]HyperedgeID{eid})
			}
			i++
		}
	}()

	// Use a deadline-bounded wait; if locking is wrong the test hangs
	// past `dur` and the test framework will eventually time it out.
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
		// good
	case <-time.After(dur + 5*time.Second):
		t.Fatal("two-direction PromoteBatch deadlocked")
	}
}
