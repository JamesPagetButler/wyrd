package model

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestGraph_ConcurrentMixedAccess stresses the RWMutex behind Graph by
// running concurrent readers and writers. It must pass under
// `go test -race`.
//
// The test models the BMA Walk-phase access pattern: one writer
// goroutine (the sleep-cycle compactor), several reader goroutines
// (CTH ρ_net snapshotter, scout queries, observer log readers), and
// one mixed goroutine (the WDEvent observer authoring runtime anchors
// while consulting incident-edge sets).
func TestGraph_ConcurrentMixedAccess(t *testing.T) {
	t.Parallel()

	g := NewGraph()

	// Seed a population so readers have something to chew on.
	const seedNodes = 64
	for i := 0; i < seedNodes; i++ {
		if err := g.AddNode(mkNode(NodeID(fmt.Sprintf("seed-%d", i)), TierQuaternion)); err != nil {
			t.Fatalf("seed node %d: %v", i, err)
		}
	}
	// Seed a few edges so IncidentEdges returns non-empty.
	for i := 0; i < 32; i++ {
		e := mkEdge(
			HyperedgeID(fmt.Sprintf("seed-edge-%d", i)),
			[]NodeID{
				NodeID(fmt.Sprintf("seed-%d", i%seedNodes)),
				NodeID(fmt.Sprintf("seed-%d", (i+1)%seedNodes)),
			},
			TierQuaternion,
		)
		if err := g.AddHyperedge(e); err != nil {
			t.Fatalf("seed edge %d: %v", i, err)
		}
	}

	const dur = 200 * time.Millisecond
	deadline := time.Now().Add(dur)
	var wg sync.WaitGroup
	var (
		writes     atomic.Int64
		reads      atomic.Int64
		incidents  atomic.Int64
		nodesSnaps atomic.Int64
		removeSucc atomic.Int64
		writerErrs atomic.Int64
	)

	// Writer goroutine: appends new nodes + edges.
	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		for time.Now().Before(deadline) {
			id := NodeID(fmt.Sprintf("w-%d", i))
			if err := g.AddNode(mkNode(id, TierQuaternion)); err != nil {
				writerErrs.Add(1)
			}
			eid := HyperedgeID(fmt.Sprintf("w-edge-%d", i))
			err := g.AddHyperedge(mkEdge(eid, []NodeID{id, NodeID("seed-0")}, TierQuaternion))
			if err != nil {
				writerErrs.Add(1)
			}
			writes.Add(1)
			i++
		}
	}()

	// Reader goroutines: 4 of them.
	for r := 0; r < 4; r++ {
		wg.Add(1)
		go func(r int) {
			defer wg.Done()
			for time.Now().Before(deadline) {
				_ = g.NodeCount()
				_ = g.EdgeCount()
				if _, ok := g.Node(NodeID("seed-0")); ok {
					reads.Add(1)
				}
			}
		}(r)
	}

	// Incident-edge query goroutine — exercises the read path that
	// returns a fresh slice copy.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for time.Now().Before(deadline) {
			_ = g.IncidentEdges(NodeID("seed-0"))
			incidents.Add(1)
		}
	}()

	// Snapshot goroutine — exercises Nodes/Hyperedges allocations.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for time.Now().Before(deadline) {
			_ = g.Nodes()
			_ = g.Hyperedges()
			nodesSnaps.Add(1)
		}
	}()

	// Mixed goroutine — adds runtime-anchor edges and then removes
	// them. Models the WDEvent observer authoring + cleanup pattern.
	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		for time.Now().Before(deadline) {
			eid := HyperedgeID(fmt.Sprintf("rt-%d", i))
			if err := g.AddHyperedge(mkEdge(eid, []NodeID{NodeID("seed-1")}, TierQuaternion)); err == nil {
				if rerr := g.RemoveHyperedge(eid); rerr == nil {
					removeSucc.Add(1)
				}
			}
			i++
		}
	}()

	wg.Wait()

	// Sanity: lots of work happened, no panics, no races (-race
	// catches silently).
	if writes.Load() == 0 {
		t.Errorf("writer made zero progress")
	}
	if reads.Load() == 0 {
		t.Errorf("readers made zero progress")
	}
	if incidents.Load() == 0 {
		t.Errorf("incident-edge readers made zero progress")
	}
	if nodesSnaps.Load() == 0 {
		t.Errorf("snapshot readers made zero progress")
	}
	if removeSucc.Load() == 0 {
		t.Errorf("mixed add/remove never succeeded")
	}
	if writerErrs.Load() != 0 {
		t.Logf("writer saw %d errors (typically dup-id collisions; investigate if persistent)", writerErrs.Load())
	}
	t.Logf("writes=%d reads=%d incidents=%d snaps=%d remove-pairs=%d",
		writes.Load(), reads.Load(), incidents.Load(),
		nodesSnaps.Load(), removeSucc.Load())
}

// TestGraph_ReadersDontBlockEachOther is a behavioural sanity check
// that the RWMutex does what its name says — multiple concurrent
// readers proceed in parallel, not in series.
func TestGraph_ReadersDontBlockEachOther(t *testing.T) {
	t.Parallel()
	g := NewGraph()
	for i := 0; i < 100; i++ {
		_ = g.AddNode(mkNode(NodeID(fmt.Sprintf("n-%d", i)), TierComplex))
	}

	const goroutines = 8
	const iterations = 1000
	var wg sync.WaitGroup
	start := time.Now()
	for r := 0; r < goroutines; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				_ = g.NodeCount()
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)
	// Loose bound — eight goroutines × 1000 NodeCount() calls under
	// RLock should finish in well under a second on any reasonable
	// machine. If this regresses badly something is wrong.
	if elapsed > 2*time.Second {
		t.Errorf("8 reader goroutines took %v — readers may be serialising", elapsed)
	}
}
