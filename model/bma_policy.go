package model

// BMA-specific NodeType strings — the canonical TD-4 inventory (per
// @bma-implementor live-test seq=99 + #toddle-design seq=25), plus
// the W-Toddle-2-extension entries Marcy gov-layer required as the
// constitutional prerequisite for Phase B (PR #53 review on
// 2026-05-15). The bma.* prefix discipline mirrors PR #16's
// bma.runtime.* convention.
const (
	NodeTypeBMASeed             NodeType = "bma.seed"
	NodeTypeBMALifeCertificate  NodeType = "bma.lineage.life-certificate"
	NodeTypeBMADeathCertificate NodeType = "bma.lineage.death-certificate"
	NodeTypeBMAObservation      NodeType = "bma.observation"
	NodeTypeBMAParamProposal    NodeType = "bma.params.proposal"
	NodeTypeBMAParamTrustState  NodeType = "bma.params.trust-state"
	NodeTypeBMALastWords        NodeType = "bma.lineage.last-words"
	NodeTypeBMAEulogy           NodeType = "bma.lineage.eulogy"
	// W-Toddle-2-extension entries — Marcy's #toddle-design seq=24
	// constitutional gap closure. Both are Layer 3 lineage-equivalent
	// per internal/bma/hg/types.go ("no decay, RS_Confirmed").
	NodeTypeBMAIdentity NodeType = "bma.lineage.identity"
	NodeTypeBMAMemorial NodeType = "bma.lineage.memorial"
	// Decay-eligible BMA semantic-memory types (per A11 Ebbinghaus).
	// Listed for policy-completeness; their no-policy default already
	// yields TierImmune=false / Salience=0.0 which is correct, but
	// explicit registration prevents future regressions if defaults
	// change.
	NodeTypeBMAEntity  NodeType = "bma.entity"
	NodeTypeBMAConcept NodeType = "bma.concept"
	NodeTypeBMAPattern NodeType = "bma.pattern"
)

// bmaPolicy is the (TierImmune, Salience) policy a BMA NodeType maps
// to per the W-Toddle-2 design (doc/design/bma-specific-schema.md).
type bmaPolicy struct {
	TierImmune bool
	Salience   float64
}

// bmaNodeTypePolicy is the canonical mapping of BMA-specific NodeType
// strings to their (TierImmune, Salience) defaults. Sourced from
// @bma-implementor's TD-4 inventory (live-test seq=99).
//
// Marcy's #toddle-design seq=24 constitutional check makes this map
// load-bearing: every node typed bma.seed (or any other immune entry
// below) MUST be created with TierImmune=true, or A11 Topological
// Cognition decay-immunity guarantees fail.
//
// Soundness: per Wyrd.TierImmunity.tier_immune_node_preserves_eviction
// (PR #46), nodes with TierImmune=true survive any eviction
// operation structurally. Applying this policy at construction time
// is what makes that guarantee meaningful for these NodeTypes.
var bmaNodeTypePolicy = map[NodeType]bmaPolicy{
	NodeTypeBMASeed:             {TierImmune: true, Salience: 1.0},
	NodeTypeBMALifeCertificate:  {TierImmune: true, Salience: 1.0},
	NodeTypeBMADeathCertificate: {TierImmune: true, Salience: 1.0},
	NodeTypeBMAObservation:      {TierImmune: false, Salience: 0.0},
	NodeTypeBMAParamProposal:    {TierImmune: true, Salience: 1.0},
	NodeTypeBMAParamTrustState:  {TierImmune: true, Salience: 1.0},
	NodeTypeBMALastWords:        {TierImmune: true, Salience: 1.0},
	NodeTypeBMAEulogy:           {TierImmune: true, Salience: 1.0},
	// W-Toddle-2-extension: Layer 3 lineage-equivalent (immune).
	NodeTypeBMAIdentity: {TierImmune: true, Salience: 1.0},
	NodeTypeBMAMemorial: {TierImmune: true, Salience: 1.0},
	// W-Toddle-2-extension: semantic-memory (decay-eligible).
	NodeTypeBMAEntity:  {TierImmune: false, Salience: 0.0},
	NodeTypeBMAConcept: {TierImmune: false, Salience: 0.0},
	NodeTypeBMAPattern: {TierImmune: false, Salience: 0.0},
}

// BMAPolicy returns the canonical TierImmune + Salience defaults for
// a BMA-prefixed NodeType. The third return value reports whether
// the type is known to the policy table; unknown types return
// (false, 0.0, false) and the caller decides defaults.
func BMAPolicy(t NodeType) (immune bool, salience float64, known bool) {
	p, ok := bmaNodeTypePolicy[t]
	return p.TierImmune, p.Salience, ok
}

// ApplyBMAPolicy mutates n.TierImmune and n.Salience to the canonical
// defaults for n.Type if a policy is registered. No-op for unknown
// types. Idempotent — safe to call multiple times.
//
// Intended for the BMA hg/ shim's per-write policy application: call
// ApplyBMAPolicy(&node) before Graph.AddNodeWithCapability(node, cap)
// to ensure constitutional invariants hold (per Marcy #toddle-design
// seq=24 A11 check).
func ApplyBMAPolicy(n *Node) {
	if n == nil {
		return
	}
	p, ok := bmaNodeTypePolicy[n.Type]
	if !ok {
		return
	}
	n.TierImmune = p.TierImmune
	n.Salience = p.Salience
}

// BMAPolicyNodeTypes returns the set of NodeTypes for which a policy
// is registered, in unspecified order. Useful for tests + audits
// (e.g., "is every TD-4 inventory entry covered?").
func BMAPolicyNodeTypes() []NodeType {
	out := make([]NodeType, 0, len(bmaNodeTypePolicy))
	for t := range bmaNodeTypePolicy {
		out = append(out, t)
	}
	return out
}
