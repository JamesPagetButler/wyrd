package compute

import (
	"testing"
	"time"

	"github.com/JamesPagetButler/wyrd/model"
)

func mkNode(id model.NodeID) model.Node {
	return model.Node{ID: id, Type: "test", Tier: model.TierQuaternion, Created: time.Unix(0, 0)}
}

func mkEdge(id model.HyperedgeID, nodes ...model.NodeID) model.Hyperedge {
	return model.Hyperedge{
		ID:      id,
		Nodes:   nodes,
		Weight:  model.NewQuaternionWeight(1, 0, 0, 0),
		Created: time.Unix(0, 0),
	}
}

// TestBridge_Promote_PreservesCount is the runtime check for
// `Wyrd.Bridge.bridge_promote_preserves_count` (C-20c, Phase 2).
func TestBridge_Promote_PreservesCount(t *testing.T) {
	src := model.NewGraph()
	dst := model.NewGraph()
	for _, id := range []model.NodeID{"a", "b", "c"} {
		_ = src.AddNode(mkNode(id))
		_ = dst.AddNode(mkNode(id))
	}
	_ = src.AddHyperedge(mkEdge("e", "a", "b", "c"))

	totalBefore := src.EdgeCount() + dst.EdgeCount()
	br := &Bridge{Source: src, Destination: dst}
	if err := br.Promote("e"); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	totalAfter := src.EdgeCount() + dst.EdgeCount()

	if totalBefore != totalAfter {
		t.Errorf("count not preserved: before=%d after=%d", totalBefore, totalAfter)
	}
}

// TestBridge_Promote_ExactlyOneSide is the runtime check for
// `Wyrd.Bridge.bridge_promote_exactly_one_side` (Phase 2).
func TestBridge_Promote_ExactlyOneSide(t *testing.T) {
	src := model.NewGraph()
	dst := model.NewGraph()
	for _, id := range []model.NodeID{"a", "b"} {
		_ = src.AddNode(mkNode(id))
		_ = dst.AddNode(mkNode(id))
	}
	_ = src.AddHyperedge(mkEdge("e", "a", "b"))

	br := &Bridge{Source: src, Destination: dst}
	if err := br.Promote("e"); err != nil {
		t.Fatalf("Promote: %v", err)
	}

	_, inSrc := src.Hyperedge("e")
	_, inDst := dst.Hyperedge("e")
	if inSrc {
		t.Error("edge still in source after promote")
	}
	if !inDst {
		t.Error("edge not in destination after promote")
	}
}

// TestBridge_Promote_RollbackOnDestFailure verifies atomicity: if the
// destination rejects the edge (e.g., missing referenced node), the
// source is left unchanged.
func TestBridge_Promote_RollbackOnDestFailure(t *testing.T) {
	src := model.NewGraph()
	dst := model.NewGraph()
	_ = src.AddNode(mkNode("a"))
	_ = src.AddNode(mkNode("b"))
	// Destination intentionally missing "b".
	_ = dst.AddNode(mkNode("a"))
	_ = src.AddHyperedge(mkEdge("e", "a", "b"))

	br := &Bridge{Source: src, Destination: dst}
	if err := br.Promote("e"); err == nil {
		t.Error("expected promote to fail (missing node in dst)")
	}

	if _, ok := src.Hyperedge("e"); !ok {
		t.Error("source edge missing after failed promote (rollback failed)")
	}
}

func TestBridge_Promote_UnknownEdge(t *testing.T) {
	br := &Bridge{Source: model.NewGraph(), Destination: model.NewGraph()}
	if err := br.Promote("nonexistent"); err == nil {
		t.Error("expected error promoting unknown edge")
	}
}
