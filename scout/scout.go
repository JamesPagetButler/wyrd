package scout

import (
	"errors"
	"fmt"

	"github.com/JamesPagetButler/wyrd/model"
	"github.com/JamesPagetButler/wyrd/query"
)

// Volume is a 4D spatiotemporal bounding box. v0.1 mirrors the shape;
// v0.2 switches to emulator.Volume once that ships on qbp-compute-unit
// main (per @qbp-cu-implementor #addendum-18-walk seq=13, ETA ~1 week).
//
// Per Gemini #addendum-18-walk seq=37: per-component (lat, lon, time,
// height) semantics inherited from NT_SCOPE_PHYSICAL.
type Volume struct {
	Min [4]float64 // lat, lon, time, height (lower bound)
	Max [4]float64 // lat, lon, time, height (upper bound)
}

// Width is the precision tier for the query computation. v0.1 stub
// until emulator.Width ships; v0.2 type-alias to emulator.Width.
type Width uint8

const (
	// W8 is the peripheral register precision (QW8, A18 §3.1).
	W8 Width = 8
	// W128 is the foveal register precision (QW128, A18 §3.2).
	W128 Width = 128
)

// Intersection records one Active Agent whose world-line crosses the
// source→sink path within the queried Locale Volume.
type Intersection struct {
	AgentID            model.NodeID
	AgentType          model.NodeType
	IntersectionLocale [4]float64 // v0.2: emulator.Locale
	AbsorptionGain     float64    // QW8 estimate; uniform at v0.1
	Provenance         []model.NodeID
}

// ErrScoutQueryInvalid is returned when ScoutQuery inputs are malformed
// (nil graph, empty source/sink, unrecognised precision).
var ErrScoutQueryInvalid = errors.New("scout: invalid query parameters")

// uniformPlaceholderGain is the v0.1 placeholder AbsorptionGain.
// Replaced by the spectral computation when the body PR lands.
const uniformPlaceholderGain = 0.5

// ScoutQuery dispatches a Stance × Locale × source × sink × agent-type
// query and returns all qualifying intersections in the focal cone.
//
// v0.1 PLACEHOLDER BEHAVIOUR — DO NOT WRITE STANCE-DEPENDENT CONSUMER
// CODE YET. v0.1 always returns the trivial intersection set with
// uniform AbsorptionGain = 0.5 regardless of the Stance carried in
// Provenance. Stance-Algorithm dispatch is deferred to v0.2; the
// precision Width arg does NOT yet drive routing either. Consumers
// writing "if Stance.includes(X) then expect higher AbsorptionGain"
// logic against v0.1 will fail silently when v0.2 lands real spectral
// routing — there's no behaviour to observe at v0.1.
//
// Per @bma (Marcy Gen 61) #addendum-18-walk seq=48 §I4 read.
//
// Algorithm at v0.1: for each known node in the graph matching
// agentTypes, emit one Intersection with the supplied source/sink as
// Provenance and the uniform placeholder gain. Order is unspecified.
// Locale Volume is recorded but not yet used to filter spatially.
//
// Soundness: per Wyrd.Hypergraph.hyperedge_preserves_incident_edges
// (Phase 2 C-20a). The placeholder uses model.Graph reads that respect
// the existing concurrency contract via query.API.
//
//nolint:revive // ScoutQuery is the federation-vocabulary name from A18 §6 + D9 lock; renaming to scout.Query would silently drop the vocabulary binding consumers (BMA, Contextus) rely on.
func ScoutQuery(
	g *model.Graph,
	locale Volume,
	source model.NodeID,
	sink model.NodeID,
	agentTypes []model.NodeType,
	precision Width,
) ([]Intersection, error) {
	if g == nil {
		return nil, fmt.Errorf("%w: nil graph", ErrScoutQueryInvalid)
	}
	if source == "" {
		return nil, fmt.Errorf("%w: empty source", ErrScoutQueryInvalid)
	}
	if sink == "" {
		return nil, fmt.Errorf("%w: empty sink", ErrScoutQueryInvalid)
	}
	if precision != W8 && precision != W128 {
		return nil, fmt.Errorf("%w: precision %d not in {W8, W128}", ErrScoutQueryInvalid, precision)
	}

	// v0.1 placeholder: walk all graph nodes; emit any node whose Type
	// matches an entry in agentTypes. Real implementation will narrow
	// to a locale-bounded oriented hypergraph and compute spectral
	// absorption — currently blocked on the query/ + oriented-edge +
	// laplacian-body chain.
	_ = locale    // recorded; not yet used spatially at v0.1
	_ = precision // recorded; not yet used at v0.1

	q := query.New(g)
	_ = q // reserved for the v0.2 body's traversal step

	if len(agentTypes) == 0 {
		return []Intersection{}, nil
	}

	wanted := make(map[model.NodeType]struct{}, len(agentTypes))
	for _, t := range agentTypes {
		wanted[t] = struct{}{}
	}

	provenance := []model.NodeID{source, sink}

	var out []Intersection
	for _, n := range g.Nodes() {
		if _, ok := wanted[n.Type]; !ok {
			continue
		}
		out = append(out, Intersection{
			AgentID:        n.ID,
			AgentType:      n.Type,
			AbsorptionGain: uniformPlaceholderGain,
			Provenance:     provenance,
		})
	}
	if out == nil {
		return []Intersection{}, nil
	}
	return out, nil
}
