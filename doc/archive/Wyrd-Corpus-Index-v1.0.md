# Wyrd / Skuld / QBP-CU Corpus

## Master Index — April 2026

**Helpful Engineering — Quaternion-Based Physics Programme**
Principal Investigator: James Paget Butler
April 2026 — Rev 1.0

> **Purpose.** This document is the canonical index of the Wyrd / Skuld / QBP-CU corpus as of April 2026. It records what each document covers, its status, its authoritative version, and its relationship to the others. Use this as the entry point when navigating the corpus.

> **Naming convention.** Documents follow the pattern `{System}-{Topic}-v{Major.Minor}.md` or `Wyrd-{Theorem-or-Topic}-v{x}.lean` for proofs. The `Wyrd-` prefix identifies project membership; the version number tracks substantive revisions.

---

## Quick reference: where to start

| You want to... | Read this first |
|---|---|
| Understand the project's overall architecture | *Wyrd-Supervisor-Architecture-v0.2.md* |
| Implement Skuld | *Skuld-Spec-v1.0.md* (then *Skuld-Spec-Hammer-Review.md* for v1.1 amendments) |
| Implement the QBP-CU hardware interface | *QBP-CU-SiFive-Interface-Spec-v0.2.md* |
| Decide what instructions go in the ISA | *Wyrd-Workload-ISA-v0.1.md* |
| See the binding ISA encoding decisions | *QBP-CU-ISA-Freeze-C01-C02.md* |
| Understand the privilege model formally | *Wyrd-Algebraic-Privilege-Proofs-v0.4.lean* |
| Understand the substrate-uncertainty hedge | *Wyrd-BranchA-Contingency-v0.1.md* |
| See what tickets to start at Crawl | *Wyrd-Supervisor-Architecture-v0.2.md* §6 |

---

## 1. Architecture documents (read first)

### `Wyrd-Supervisor-Architecture-v0.2.md` — **CURRENT**

The canonical architectural document. Defines:
- The thesis: Skuld = Wyrd executing privileged queries against itself
- The privilege ring tower (ℂ user / ℍ supervisor / 𝕆 kernel / 𝕊 firmware)
- Naming (Wyrd, Skuld, QBP-CU)
- The phased migration Crawl → Walk → Run → Sprint
- The C-01..C-12 ticket plan
- Open architectural questions

Supersedes: `Wyrd-Supervisor-Architecture-v0.1.md`.

### `Wyrd-Spec-Front-Matter-v1.0.md`

The renamed front matter for the database spec. Records the MuninnDB → Wyrd transition and gives the etymology footnote. Drops on top of the existing `QBP-Native-Database-Spec.md`.

---

## 2. Specifications (the contracts)

### `Skuld-Spec-v1.0.md` — **CURRENT for Crawl baseline**

The Crawl-phase API specification for the supervisor. Defines:
- The 9-call user-visible API surface
- The 2 supervisor-internal calls
- Process management semantics
- Capability mechanism
- Failure semantics
- Concurrency model

Pending update: see *Skuld-Spec-Hammer-Review.md* for v1.1 amendments.

### `QBP-CU-SiFive-Interface-Spec-v0.2.md` — **CURRENT**

The QBP-CU hardware interface specification, including SiFive integration. Defines:
- The 10-instruction ISA (with FROZEN encodings via *QBP-CU-ISA-Freeze-C01-C02*)
- VCIX and SSCI port allocations
- Pipeline timing
- WDEvent struct, Custom CSRs
- Validation tiers and cosim harness
- Forward path through Phases 0-4

Supersedes: implicit v0.1 (the original document, not in this archive).

### `QBP-CU-ISA-Freeze-C01-C02.md` — **DECIDED Rev 1.0**

The binding ISA and CSR encoding decisions. Output of tickets C-01 (ISA freeze) and C-02 (CSR additions). Constituents:
- 10 instructions with funct7/funct6 final allocations
- AlgebraID renumbering (0=ℂ, 1=ℍ, 2=𝕆, 3=𝕊, 4=Branch A, 5=Branch B)
- `qbp_invariant` CSR layout
- WDEvent struct binary layout
- Compiler intrinsic names

---

## 3. Analyses (the why)

### `Wyrd-Workload-ISA-v0.1.md` — **CURRENT**

The workload analysis driving the 10-instruction ISA. Establishes:
- Five workload categories (particle dynamics, field evolution, correlation, algebraic prediction, quantum)
- Quaternion-vs-octonion fractions (95-99% quaternion-dominated for physics)
- Why QFMA, QSAND, QNORM are essential
- Why QASSOC and QALT are NOT in the ISA (subsumed by watchdog + qbp_invariant)
- Per-lane throughput targets

### `Wyrd-SiFive-Spec-Review-v0.1.md` — **HISTORICAL**

The review document that motivated the v0.2 spec. Identified the watchdog-as-privilege-detector finding. Superseded by integration into v0.2 spec, but preserved as the reasoning trail.

### `Wyrd-BranchA-Contingency-v0.1.md` — **CURRENT**

What happens to the privilege model if the QBP substrate is ℂ⊕ℍ⊕M₃(ℂ) direct sum (Branch A) instead of the Cayley-Dickson tower. Establishes:
- Architecture survives, mathematics gets replaced
- The hedge strategy: H3 for hardware (configurable), H1 for privilege model (commit Cayley-Dickson, accept redesign cost if Branch A wins)
- Trigger conditions for revisiting

### `Skuld-Spec-Hammer-Review.md` — **CURRENT**

Critical reading of *Skuld-Spec-v1.0.md* against Hammer's actual workload. Identifies five gaps and proposes v1.1 amendments:
- QBPSubmitBatch, QBPSubmitStream
- WyrdPrepare for prepared queries
- Stream.NextBatch for batched event consumption
- DeterministicConfig for replay/debug

---

## 4. Lean proofs (the formal foundation)

### `Wyrd-Algebraic-Privilege-Proofs-v0.4.lean` — **CURRENT**

The master proof file. Source-verified against mathlib4 master. Contains:
- Abstract structural lemmas (no_surjection_*)
- T1.2: boundary detector vanishing
- T1.2.a witness: commutator [i, j] ≠ 0 in ℍ (source-verified)
- T2.1.a: no surjection ℂ → ℍ (the user → supervisor boundary, fully proven)

Supersedes: v0.1, v0.2, v0.3. The mathematics is locked; the API names are source-verified against the cloned mathlib4 repo at HEAD.

### `Wyrd-CayleyDickson-Types-v0.1.lean`

The Cayley-Dickson construction for octonions and sedenions, since mathlib4 doesn't include these. Provides:
- Generic `CayleyDickson A` type
- `Octonion R := CayleyDickson (Quaternion R)`
- `Sedenion R := CayleyDickson (Octonion R)`
- Octonion basis elements e₀..e₇
- T1.2.b witness: associator(e₁, e₂, e₄) ≠ 0 in 𝕆

Pending: API name updates to current mathlib4 conventions (3-parameter `QuaternionAlgebra`).

### `Wyrd-T2.2-Projection-v0.1.lean`

T2.2: projection well-definedness. The headline theorem `kernel_supervisor_safe`: kernel-ring computations on supervisor-ring values, projected back, equal supervisor-ring computations.

### `Wyrd-T2.3-Capability-Soundness-v0.1.lean`

T2.3: capability mechanism. The `Capability` structure, sandwich preservation (associative case), positive part (capability grants safe access), negative part (no capability ⇒ no synthesis), capability projection (wider subsumes narrower), and the Hammer simulation as a theorem.

### `Wyrd-T3.1-Noise-Bound-v0.2.lean` — **CURRENT**

T3.1: parametric framework over an abstract `RoundingModel`, with the `abs_error_two_muls` proof body fully written out. fp32 specialization gives noise floor ~3e-6 for unit-magnitude components, ~3e-3 for components of magnitude 10. Sets ε_priv defensibly.

Supersedes: v0.1 (the v0.1 had one `sorry`; v0.2 has zero).

### `Wyrd-Sedenion-Alternator-Witness-v0.1.lean`

The explicit sedenion alternator witness derived by hand. α = (e₁ᴼ, e₄ᴼ), β = (e₂ᴼ, 0). Alternator = (0, −2 e₇ᴼ). Mathematics fully verified; one `sorry` for the destructuring tactic.

### `Wyrd-Octonion-Alternativity-v0.1.lean`

The structured proof that 𝕆 is alternative. Mathematics fully expanded by hand. Two `sorry`s in the closing `ring_nf` invocations and two `axiom`s standing in for known mathlib4 results — these need replacement and refinement in a live Lean session.

### `Wyrd-Lean-Airtightening-Session-Report.md`

Records the verification work performed in the airtightening session. Documents what was source-verified against mathlib4 master, what remains, and the estimated 1-2 hours of remaining live-Lean work to drive sorries to zero.

### `Wyrd-Mathlib-API-Verification-Checklist.md`

The original API verification checklist (now mostly resolved by *Wyrd-Lean-Airtightening-Session-Report.md*). Retains value as a reference for "where to look in mathlib4" if API drift recurs.

---

## 5. Status summary

### What's tight

- **Architecture (v0.2)** — committed. The ring tower, the watchdog as privilege detector, the phased migration are all settled.
- **ISA (Rev 1.0 frozen)** — committed. Encodings, CSR layout, WDEvent struct, intrinsic names all final.
- **Workload analysis** — closes the "why these instructions" question definitively.
- **T1.2.a / T2.1.a (ℂ → ℍ boundary)** — fully source-verified Lean proof. The user → supervisor boundary is formally established.
- **T3.1 (noise bound)** — zero sorries.
- **T2.2, T2.3** — closed (modulo mathlib API touch-ups).
- **Branch A contingency** — analyzed, hedge strategy committed.

### What needs live-Lean time

- Updating CayleyDickson types file to mathlib4's 3-parameter QuaternionAlgebra
- Final destructuring tactic in sedenion witness
- `ring_nf`-with-`star` interaction in octonion alternativity (replaces two `axiom`s)
- One full `lake build` cycle on the entire corpus

Estimated: 1-2 hours focused work in a live Lean environment.

### What needs implementation

- C-03 onward (Go simulator, skuld package, integration tests). All unblocked by the freeze decision.
- Skuld-Spec v1.1 amendments per the Hammer review.

### What's still genuinely open

- Watchdog event rate as supervisor bottleneck (needs benchmark)
- "Who supervises the supervisor" at Walk and beyond
- QSANDWICH ISA promotion timing (deferred)
- fp16 QFMA precision interaction with ε_priv (per-workload empirical)

---

## 6. Cross-reference table

Which document references which:

| Source document | References |
|---|---|
| `Wyrd-Supervisor-Architecture-v0.2.md` | Skuld-Spec, ISA-Freeze, Workload-ISA, BranchA-Contingency, all proofs |
| `Skuld-Spec-v1.0.md` | Wyrd-Supervisor-Architecture-v0.2, qbpcu, wyrd packages |
| `QBP-CU-SiFive-Interface-Spec-v0.2.md` | Workload-ISA, Wyrd-Supervisor-Architecture-v0.2, ISA-Freeze, BranchA-Contingency |
| `QBP-CU-ISA-Freeze-C01-C02.md` | SiFive-Spec-v0.2, Workload-ISA, Skuld-Spec, Wyrd-Supervisor-Architecture, BranchA-Contingency |
| `Wyrd-Workload-ISA-v0.1.md` | Wyrd-Supervisor-Architecture-v0.2 |
| `Wyrd-BranchA-Contingency-v0.1.md` | Wyrd-Supervisor-Architecture-v0.2, Workload-ISA, QBP-Dark-Matter-Fork-Analysis (external) |
| `Wyrd-Algebraic-Privilege-Proofs-v0.4.lean` | imports mathlib4; cited by Wyrd-Supervisor-Architecture and Skuld-Spec |
| `Skuld-Spec-Hammer-Review.md` | Skuld-Spec-v1.0, Workload-ISA |

The references form a directed graph. Updates to root documents (architecture v0.2, the freeze) propagate to leaves (proofs, specs).

---

## 7. Reading order recommendations

### For the implementer (BMA, post-Crawl.Heartbeat)

1. *Wyrd-Supervisor-Architecture-v0.2.md* — the architectural picture
2. *QBP-CU-ISA-Freeze-C01-C02.md* — the binding hardware decisions
3. *Skuld-Spec-v1.0.md* — the supervisor API to implement
4. *Skuld-Spec-Hammer-Review.md* — the v1.1 amendments to fold in
5. *Wyrd-Workload-ISA-v0.1.md* — context for the throughput targets
6. *QBP-CU-SiFive-Interface-Spec-v0.2.md* — the hardware interface for `qbpcu` package

### For the theorist (Gemini collaboration, QBP physics)

1. *Wyrd-Supervisor-Architecture-v0.2.md* §3 — the privilege ring tower
2. *Wyrd-Algebraic-Privilege-Proofs-v0.4.lean* — the formal underpinning
3. *Wyrd-BranchA-Contingency-v0.1.md* — what changes if substrate is direct sum
4. *Wyrd-Workload-ISA-v0.1.md* §2 — the physics workload categories
5. *QBP-CU-SiFive-Interface-Spec-v0.2.md* §6 — what the watchdog computes

### For the reviewer (red-team, audit)

1. *Wyrd-Supervisor-Architecture-v0.2.md* §11 — what's committed and what's open
2. *Wyrd-SiFive-Spec-Review-v0.1.md* — the review pattern to follow
3. *Skuld-Spec-Hammer-Review.md* — example of workload-driven critique
4. *Wyrd-BranchA-Contingency-v0.1.md* — the substrate-risk analysis
5. The Lean proofs (any of them) — to verify mathematical claims

### For James (PI, future re-read)

1. This document — start here
2. *Wyrd-Supervisor-Architecture-v0.2.md* §11 — the still-open questions
3. *Wyrd-BranchA-Contingency-v0.1.md* §7 — the trigger conditions for revisiting
4. *Skuld-Spec-Hammer-Review.md* — the v1.1 amendments awaiting decision

---

## 8. Versioning policy

**Semantic versioning** for substantive revisions:
- v0.x — exploratory, expected to change
- v1.0 — first committed version, changes need migration
- v1.x — backward-compatible extensions
- v2.0 — breaking changes

**Document supersession.** When v(N).y supersedes v(N-1).y, the older version is preserved in this archive but marked HISTORICAL. New work cites the latest.

**Multi-version files in the archive.** When the same content has multiple versions (e.g., `Wyrd-Algebraic-Privilege-Proofs-v0.1/v0.2/v0.3/v0.4`), only the latest is CURRENT. Older versions are kept for the audit trail.

---

## 9. Document outputs by category

### Architecture (CURRENT)
- `Wyrd-Supervisor-Architecture-v0.2.md`
- `Wyrd-Spec-Front-Matter-v1.0.md`

### Specifications (CURRENT)
- `Skuld-Spec-v1.0.md`
- `QBP-CU-SiFive-Interface-Spec-v0.2.md`
- `QBP-CU-ISA-Freeze-C01-C02.md`

### Analyses (CURRENT)
- `Wyrd-Workload-ISA-v0.1.md`
- `Wyrd-BranchA-Contingency-v0.1.md`
- `Skuld-Spec-Hammer-Review.md`

### Reviews (HISTORICAL — preserved as reasoning trail)
- `Wyrd-SiFive-Spec-Review-v0.1.md`

### Lean proofs (CURRENT)
- `Wyrd-Algebraic-Privilege-Proofs-v0.4.lean`
- `Wyrd-CayleyDickson-Types-v0.1.lean`
- `Wyrd-T2.2-Projection-v0.1.lean`
- `Wyrd-T2.3-Capability-Soundness-v0.1.lean`
- `Wyrd-T3.1-Noise-Bound-v0.2.lean`
- `Wyrd-Sedenion-Alternator-Witness-v0.1.lean`
- `Wyrd-Octonion-Alternativity-v0.1.lean`

### Lean proofs (HISTORICAL)
- `Wyrd-Algebraic-Privilege-Proofs-v0.1.lean` (early version, preserved)
- `Wyrd-Algebraic-Privilege-Proofs-v0.2.lean` (mathlib3-style, preserved)
- `Wyrd-Algebraic-Privilege-Proofs-v0.3.lean` (web-search-verified, preserved)
- `Wyrd-T3.1-Noise-Bound-v0.1.lean` (had one sorry, now resolved in v0.2)

### Process documents
- `Wyrd-Lean-Airtightening-Session-Report.md` — records the verification work
- `Wyrd-Mathlib-API-Verification-Checklist.md` — original checklist (mostly resolved)
- `Wyrd-Corpus-Index-v1.0.md` — this document

---

## 10. Next steps

In rough priority order:

1. **Live-Lean session** — drive proof corpus to zero sorries / zero axioms. ~1-2 hours.
2. **Skuld-Spec v1.1 draft** — fold in Hammer review amendments. ~1 day.
3. **Squam Lake use case validation** — replicate Hammer review with field-evolution workload. ~½ day.
4. **C-03 begin** — qbpcu Go package with Mock and Golden accelerators. ~2-3 weeks BMA work.
5. **Quarterly re-read of BranchA-Contingency** — trigger condition check.
6. **Update BMA archive references** — Start-Here.md, Seed-Manifest.md cite the new corpus.

---

*End of Wyrd / Skuld / QBP-CU Corpus Master Index v1.0*
