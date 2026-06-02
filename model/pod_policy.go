package model

// Pentagon Pod NodeType strings — the five cognitive architecture cells
// per BMA Spec v9.1 §14 + Theory v3.0 §2.1 (A20.0 §0.2 ring assignments).
//
// Ring-tier mapping (A20.0 §0.2):
//
//	Conscious  = ℍ Supervisor ring  → authority to write cross-cell
//	Subconscious = ℂ User ring      → authority within bilateral cell
//
// Dev pod (Walk-phase) is deliberately absent — deferred per the
// Crawl/Toddle phase table (BMA Spec v9.1 §14.9 gate criterion).
// Add NodeTypePodDev here when Walk-phase begins.
//
// All four Crawl-phase pod nodes are TierImmune: they represent
// persistent cognitive identity structure that must survive any
// eviction cycle. Decay of a pod node would invalidate the cognitive
// architecture's bilateral symmetry invariant.
//
// Soundness: TierImmune=true here composes with
// Wyrd.TierImmunity.tier_immune_node_preserves_eviction (PR #46) —
// pod nodes survive eviction structurally, not by policy exception.
const (
	// NodeTypePodConsciousA is the Conscious-A cell (ℍ Supervisor ring).
	// Primary conscious cognitive layer; bilateral peer of Conscious-B.
	NodeTypePodConsciousA NodeType = "bma.pod.conscious-a"

	// NodeTypePodConsciousB is the Conscious-B cell (ℍ Supervisor ring).
	// Secondary conscious cognitive layer; bilateral peer of Conscious-A.
	NodeTypePodConsciousB NodeType = "bma.pod.conscious-b"

	// NodeTypePodSubconsciousL is the Subconscious-Left cell (ℂ User ring).
	// Left-hemisphere subconscious layer; bilateral peer of Subconscious-R.
	NodeTypePodSubconsciousL NodeType = "bma.pod.subconscious-l"

	// NodeTypePodSubconsciousR is the Subconscious-Right cell (ℂ User ring).
	// Right-hemisphere subconscious layer; bilateral peer of Subconscious-L.
	NodeTypePodSubconsciousR NodeType = "bma.pod.subconscious-r"
)

type podPolicy struct {
	TierImmune bool
	Salience   float64
}

// podNodeTypePolicy maps pod NodeTypes to their canonical
// (TierImmune, Salience) defaults.
//
// Conscious pods: Salience=1.0 — highest retention priority; ℍ ring.
// Subconscious pods: Salience=0.9 — persistent bilateral structure; ℂ ring.
// Both are TierImmune=true — pod nodes are persistent identity structure.
var podNodeTypePolicy = map[NodeType]podPolicy{
	NodeTypePodConsciousA:    {TierImmune: true, Salience: 1.0},
	NodeTypePodConsciousB:    {TierImmune: true, Salience: 1.0},
	NodeTypePodSubconsciousL: {TierImmune: true, Salience: 0.9},
	NodeTypePodSubconsciousR: {TierImmune: true, Salience: 0.9},
}

// PodPolicy returns the canonical TierImmune + Salience defaults for a
// pod NodeType. The third return value reports whether the type is known.
// Unknown types return (false, 0.0, false).
func PodPolicy(t NodeType) (immune bool, salience float64, known bool) {
	p, ok := podNodeTypePolicy[t]
	return p.TierImmune, p.Salience, ok
}

// ApplyPodPolicy mutates n.TierImmune and n.Salience to the canonical
// defaults for n.Type if a pod policy is registered. No-op for unknown
// types. Idempotent — safe to call multiple times.
func ApplyPodPolicy(n *Node) {
	if n == nil {
		return
	}
	p, ok := podNodeTypePolicy[n.Type]
	if !ok {
		return
	}
	n.TierImmune = p.TierImmune
	n.Salience = p.Salience
}

// PodPolicyNodeTypes returns the set of NodeTypes for which a pod policy
// is registered, in unspecified order.
func PodPolicyNodeTypes() []NodeType {
	out := make([]NodeType, 0, len(podNodeTypePolicy))
	for t := range podNodeTypePolicy {
		out = append(out, t)
	}
	return out
}
