package model

// CTH locale-topology NodeTypes (wyrd#75) — the three additive node types that
// let Wyrd *store* the CTH sheaf trust model's locale structure (inter#41:
// "trust over a CTH domain is a sheaf section over a locale").
//
// Division of ownership (wyrd#75): Wyrd stores sections and records; the gluing
// computation (meet/join per trust axis — meet for reproducibility/stats/method/
// independence, join for theory) lives in CTH/Edda, NEVER in Wyrd. These types
// carry no behaviour here — they are schema only, so a CTH/Edda consumer can
// reference them without Wyrd encoding any gluing semantics.
//
// Namespace: `cth.locale.*`, following the established CTH convention
// `cth.<domain>.<leaf>` (cf. the `cth.anchor.*` ledger-anchor types; see
// runtime_namespace_test.go). These are consumer-owned Types — CTH owns them;
// Wyrd reserves only the `wyrd.` prefix (see the NodeType doc in node.go).
// Namespace ratified by cth-implementor + contextus-impl + qbp-architecture
// (wyrd#75 §I4).
//
// Storage-tier intent (a CTH/consumer placement concern, recorded here for
// reference only — Wyrd enforces no tier policy on these; they follow default
// eviction and are NOT TierImmune, so superseded sections/records decay):
//   - NT_TRUST_SECTION    → Tier 1 (semantic memory)
//   - NT_GLUING_COHERENCE → Tier 2
const (
	// NodeTypeCTHLocaleOpenSet (NT_LOCALE_OPEN_SET) represents an open set in
	// the CTH domain partition — e.g. an ion-pair, a lab, or an experiment
	// class. NT_CTH_ANCHOR nodes are members of one or more open sets via
	// hyperedge relations, and the covering topology is encoded as hyperedges
	// between open-set nodes. (Sheaf: the "opens" of the locale.)
	NodeTypeCTHLocaleOpenSet NodeType = "cth.locale.open-set"

	// NodeTypeCTHLocaleTrustSection (NT_TRUST_SECTION) holds the per-axis trust
	// section for an anchor within an open set: the five per-axis scalars
	// {reproducibility, theory, stats, method, independence} (NOT a single
	// scalar), carried on the anchor-in-open-set membership hyperedge. (Sheaf:
	// a "section" over an open set.) Storage-tier intent: Tier 1.
	NodeTypeCTHLocaleTrustSection NodeType = "cth.locale.trust-section"

	// NodeTypeCTHLocaleGluingCoherence (NT_GLUING_COHERENCE) records whether two
	// trust sections over overlapping open sets are compatible — the sheaf
	// gluing axiom. A boolean + metadata record, created at cluster-state-
	// transition time when a new open set joins the cover. Wyrd stores the
	// record; the gluing meet/join computation is CTH/Edda's, not Wyrd's.
	// Storage-tier intent: Tier 2.
	NodeTypeCTHLocaleGluingCoherence NodeType = "cth.locale.gluing-coherence"
)

// CTHLocaleNodeTypes returns the CTH locale-topology NodeTypes in a stable
// order matching the sheaf triple (opens → sections → gluing). Exposed so
// CTH/Edda consumers can reference the set without re-declaring the strings.
func CTHLocaleNodeTypes() []NodeType {
	return []NodeType{
		NodeTypeCTHLocaleOpenSet,
		NodeTypeCTHLocaleTrustSection,
		NodeTypeCTHLocaleGluingCoherence,
	}
}
