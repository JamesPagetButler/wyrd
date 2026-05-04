package compute

import (
	"errors"
	"fmt"

	"github.com/JamesPagetButler/wyrd/model"
)

// ErrBridgeUnknownEdge is returned by Bridge.Promote when the named
// hyperedge does not exist in the source graph.
var ErrBridgeUnknownEdge = errors.New("wyrd: bridge: source edge not found")

// ErrBridgeAlreadyPromoted is returned when an edge with the given
// ID already exists in the destination graph.
var ErrBridgeAlreadyPromoted = errors.New("wyrd: bridge: edge already in destination")

// Bridge performs the Contextus → CTH atomic promotion of hyperedges.
//
// Soundness — `Wyrd.Bridge` (Phase 2 v1.1):
//
//   - bridge_promote_preserves_count (C-20c): total edge count is
//     conserved across the promotion (one removed from source, one
//     added to destination).
//   - bridge_promote_signal_in_cth: post-promotion the signal is in
//     the destination.
//   - bridge_promote_signal_not_in_contextus: post-promotion the signal
//     is NOT in the source.
//   - bridge_promote_exactly_one_side: combined: the signal is in
//     exactly one of the two queues post-promotion.
//
// Promote is atomic at the Go level: if any step fails the source and
// destination are left unchanged. (Atomicity at process level requires
// a transactional store; see [github.com/JamesPagetButler/wyrd/store].)
type Bridge struct {
	Source      *model.Graph
	Destination *model.Graph
}

// Promote moves the hyperedge with the given ID from Source to
// Destination. The edge's nodes must already exist in Destination
// (typically by prior promotion or by mirroring); the bridge does not
// auto-create nodes.
func (b *Bridge) Promote(id model.HyperedgeID) error {
	if b == nil || b.Source == nil || b.Destination == nil {
		return fmt.Errorf("wyrd: bridge: source or destination nil")
	}

	edge, ok := b.Source.Hyperedge(id)
	if !ok {
		return fmt.Errorf("%w: %s", ErrBridgeUnknownEdge, id)
	}
	if _, exists := b.Destination.Hyperedge(id); exists {
		return fmt.Errorf("%w: %s", ErrBridgeAlreadyPromoted, id)
	}

	// Phase 1: stage on destination first so that a destination failure
	// (e.g. missing referenced node) leaves the source unchanged.
	if err := b.Destination.AddHyperedge(edge); err != nil {
		return fmt.Errorf("wyrd: bridge: stage destination: %w", err)
	}

	// Phase 2: remove from source. If this fails (it shouldn't — we
	// just looked the edge up), roll back the destination insertion to
	// preserve the count invariant.
	if err := b.Source.RemoveHyperedge(id); err != nil {
		_ = b.Destination.RemoveHyperedge(id)
		return fmt.Errorf("wyrd: bridge: remove from source: %w", err)
	}
	return nil
}
