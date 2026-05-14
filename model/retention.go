package model

// RetentionTier is the Contextus Spec v1.3 §9.1 retention-tier axis.
//
// Distinct from [Tier] (the Cayley-Dickson algebraic privilege tower).
// Eviction caps are set per-retention-tier; algebraic-tier is
// irrelevant to retention policy.
//
// Per @contextus-impl PR #39 review (#toddle-design seq=19): naming
// the cap-policy axis explicitly disambiguates from algebraic Tier
// and prevents API-level confusion between two orthogonal concerns
// that share the word "tier".
//
// The five values match Spec v1.3 §9.1 (Skeleton / Distant /
// Peripheral / Near / Core) in outer-to-inner order. Default caps
// per Spec §5.4 are 1 / 5 / 5 / 20 / 50.
type RetentionTier uint8

const (
	// RetentionSkeleton is the outermost retention tier — smallest cap.
	RetentionSkeleton RetentionTier = iota
	// RetentionDistant is the second-outermost retention tier.
	RetentionDistant
	// RetentionPeripheral is the middle retention tier.
	RetentionPeripheral
	// RetentionNear is the second-innermost retention tier.
	RetentionNear
	// RetentionCore is the innermost retention tier — largest cap.
	RetentionCore
)

// String returns the Spec v1.3 §9.1 name.
func (rt RetentionTier) String() string {
	switch rt {
	case RetentionSkeleton:
		return "skeleton"
	case RetentionDistant:
		return "distant"
	case RetentionPeripheral:
		return "peripheral"
	case RetentionNear:
		return "near"
	case RetentionCore:
		return "core"
	default:
		return "unknown-retention-tier"
	}
}

// IsValid reports whether rt is one of the five recognised retention
// tiers (Spec v1.3 §9.1).
func (rt RetentionTier) IsValid() bool {
	return rt <= RetentionCore
}
