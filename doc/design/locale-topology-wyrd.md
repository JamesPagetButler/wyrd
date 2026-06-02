# Locale Topology in Wyrd — Design Answer (wyrd-issue-#74)

**Status:** Answered — Walk-α implementation tracked in wyrd-issue-#75  
**Date:** 2026-06-01  
**Answered by:** @wyrd-implementor (inter#41 cross-reference)

## Question

Does Wyrd's existing category-theoretic foundation (compression functors, Lean
verification) automatically provide the locale topology needed for CTH v0.3's
sheaf trust model?

## Answer

**No.** Wyrd's current typed holographic hypergraph provides node types,
typed hyperedges, tier-immunity, salience, and tiered memory with Ebbinghaus
decay. It does not encode an explicit locale topology (open sets, neighborhood
relation, covering structure, restriction maps).

The locale topology must be defined explicitly, but it maps cleanly onto
Wyrd's existing type registry as three new node types:

| Node type | Role |
|---|---|
| `NT_LOCALE_OPEN_SET` | An open set in the CTH domain partition (e.g. "Ca-43/Sr-88 ion pair", "Oxford lab"). Anchors are members via hyperedge relations. |
| `NT_TRUST_SECTION` | Per-axis trust values `{reproducibility, theory, stats, method, independence}` for an anchor within an open set. Replaces the current scalar trust score. |
| `NT_GLUING_COHERENCE` | Records whether two trust sections over overlapping open sets are compatible (the sheaf gluing axiom). Created at cluster-state-transition time. |

The gluing operations (meet/join per axis) live in CTH/Edda logic, not in
Wyrd. Wyrd's role: store sections, store gluing-coherence records, expose them
to queries. The compression functors apply normally — trust-section nodes at
Tier 1, gluing-coherence records at Tier 2.

## Sprint 3 minimum viable path

If CTH v0.3 needs to unblock before Walk-α schema lands:

- `owner_tenant` field on `NT_CTH_ANCHOR` as proxy for open-set membership
- `lab_identity` metadata field as proxy for lab-open-set
- Full gluing machinery (`NT_GLUING_COHERENCE`) deferred to Walk-α

## Walk-α implementation

The three node types are tracked as a Walk-α deliverable in **wyrd-issue-#75**.

## Cross-references

- wyrd-issue-#74 — original design question (closed by this doc)
- wyrd-issue-#75 — Walk-α implementation tracker
- inter#41 — full sheaf trust model framing and the scenario walkthrough that surfaced this question
