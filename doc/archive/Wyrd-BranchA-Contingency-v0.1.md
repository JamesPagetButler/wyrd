# Branch A Contingency Analysis

## How Wyrd / Skuld Survive (or Redesign) Under the ℂ⊕ℍ⊕M₃(ℂ) Substrate

**Helpful Engineering — Quaternion-Based Physics Programme**
Principal Investigator: James Paget Butler
Editor: Claude Opus 4.7 (red-team / architecture role)
April 2026 — Rev 0.1 — DRAFT

> **Concern.** The Wyrd privilege model assumes the QBP substrate is the Cayley-Dickson tower ℂ ⊂ ℍ ⊂ 𝕆 ⊂ 𝕊. If Branch A prevails — the spectral-action substrate ℂ ⊕ ℍ ⊕ M₃(ℂ) (Chamseddine/Connes) — the substrate is a *direct sum*, not a tower, and the "privilege rings nest" assumption fails. This document analyzes the redesign space.

> **Structure.** Section 1 frames the difference. Section 2 enumerates the redesign options and gives a recommendation. Section 3 walks through what survives unchanged from v0.2. Section 4 walks through what changes. Section 5 reframes the formal verification corpus. Section 6 lays out a hedge strategy for Crawl-phase work.

---

## 1. The structural difference

### Cayley-Dickson tower (v0.2 assumption)

```
ℂ ⊂ ℍ ⊂ 𝕆 ⊂ 𝕊
```

Each inclusion is a faithful subalgebra embedding. The privilege model exploits this:

- **Synthesis is impossible.** A process operating in ℂ cannot produce ℍ-valued elements because ℍ contains generators (j, k) absent from ℂ. (T2.1)
- **Projection is well-defined.** A computation in 𝕆 on inputs all in ℍ produces an output in ℍ. (T2.2)
- **Boundary detection is algebraic.** The associator vanishes inside ℍ but not in 𝕆. (T1.2)

The model works because each ring strictly contains the previous one and adds detectable structure.

### Branch A direct sum

The Standard Model substrate per Chamseddine-Connes is

$$\mathcal{A}_F = \mathbb{C} \oplus \mathbb{H} \oplus M_3(\mathbb{C})$$

A direct sum, not a tower. An element is a triple `(c, h, m)` with `c ∈ ℂ`, `h ∈ ℍ`, `m ∈ M₃(ℂ)`. Multiplication is componentwise:

$$(c, h, m)(c', h', m') = (cc', hh', mm')$$

Critically:
- **No nesting.** ℍ is not a subalgebra of M₃(ℂ); they are independent components.
- **No synthesis impossibility argument.** A ℂ-valued process in the first slot cannot produce an ℍ-valued element in the second slot — but this is trivially true (different slots, different components), not a structural impossibility argument that survives in interesting form.
- **No associator-based boundary detection.** Each component is separately associative (ℂ, ℍ, and M₃(ℂ) all associate). The associator is uniformly zero across the sum.

The Cayley-Dickson privilege model has nothing to enforce in Branch A. Direct port fails.

### Branch B (extended Cayley-Dickson, hypothetical)

If Branch B prevails — the algebra extends past 𝕊 in some yet-to-be-determined way to accommodate dark matter, the Cayley-Dickson framing extends naturally. The privilege model survives with possibly one more ring.

This document doesn't dwell on Branch B because the privilege model is robust to it. The problematic case is Branch A.

---

## 2. Redesign options under Branch A

There are four genuine architectural responses. I'll lay them out before recommending.

### Option A1: Component-as-ring

Treat each component of the direct sum as a separate privilege "ring":

| Wyrd ring | Branch A component | Privilege role |
|---|---|---|
| Ring 3 (user) | ℂ slot | Application code |
| Ring 2 (supervisor) | ℍ slot | Skuld itself |
| Ring 1 (kernel) | M₃(ℂ) slot | QBP-CU operations |

Privilege escalation = activating a previously-zero component.

**Pros:** Maps cleanly onto Branch A's structure. The three components are *named in the physics* — that's not arbitrary, it's the actual fundamental decomposition.

**Cons:** No T2.1-style impossibility theorem. A ℂ-process *can* construct elements of any other component just by setting that component's value. Privilege would have to be enforced by ACL ("you may not write to the ℍ slot") rather than algebraically. This is back to DBOS-level enforcement, not Wyrd-level. Architectural regression.

### Option A2: Module-over-component

Use M₃(ℂ) as the kernel, ℍ as the supervisor, and define **left modules** of each component over the others as the privilege boundary:

- ℂ-module: a vector space over ℂ. User ring.
- ℍ-module over ℂ-module: requires ℍ-linear structure. Supervisor ring.
- M₃(ℂ)-module over ℍ-module: requires matrix structure. Kernel ring.

A user-ring element holds a ℂ-vector; to act on it as ℍ requires the ℍ-module structure, which the user can't synthesize.

**Pros:** Recovers a structural impossibility argument. ℍ-module structure is algebraic data the user lacks; can't be forged.

**Cons:** Forces every Wyrd node into a module-typed wrapper. Significant refactor. Also, M₃(ℂ) is non-commutative (matrix multiplication), so the boundary detector is the commutator — but every M₃(ℂ) operation has nonzero commutator generically, so the detector fires constantly. ε threshold isn't useful because the noise floor is the signal.

### Option A3: Trace-based privilege (hybrid)

Use the trace of the M₃(ℂ) component as the privilege "phase":

- Element with trace = 0: user (lives in the ℂ ⊕ ℍ subspace effectively)
- Element with trace ≠ 0 in M₃(ℂ): supervisor or kernel
- Detection: trace computation as boundary check

**Pros:** Respects Branch A's structure. Trace is a single-cycle hardware operation. The "privilege as phase" framing from earlier brainstorming maps onto trace value.

**Cons:** The mathematical content is thin — a user can fabricate a nonzero trace just by writing a value into the M₃(ℂ) slot. Same enforcement-by-policy problem as A1.

### Option A4: Use the spectral action as the privilege metric

The Chamseddine-Connes spectral action assigns a *cost* to each algebraic configuration. The action is high for "unphysical" configurations and low for physical ones. Use the spectral action gradient as the privilege boundary:

- User-ring computations: low spectral action throughout
- Supervisor / kernel computations: high spectral action regions where the user is not "supposed" to operate
- Detection: spectral action exceeding threshold

**Pros:** Most physically meaningful. The privilege rings correspond to genuinely different physics regimes. This is the most QBP-native reading of Branch A.

**Cons:** Computationally expensive — spectral action involves a regularized trace of an operator, not a pointwise check. Too slow for the watchdog to compute on every cycle. Possibly tractable as a sampled / amortized check, but the "every-op detection" property of the Cayley-Dickson model is lost.

### Recommendation: hybrid model — A2 for type-level enforcement, A3 for runtime detection

If Branch A prevails, the right design is a **two-layer** privilege model:

**Type-level (compile-time) enforcement** via Option A2's module structure. User code is statically typed as operating on ℂ-modules; supervisor code on ℍ-modules. Skuld's API enforces the type discipline; the type system catches privilege violations at the point of code generation.

**Runtime detection** via Option A3's trace computation. The watchdog computes the trace of the M₃(ℂ) component on every operation. Anomalous traces (ε_priv exceeded) flag potential violations that escaped the type system.

This recovers most of what made the v0.2 model work, with two layers of enforcement instead of one. The mathematics is genuinely different (no Cayley-Dickson tower theorem) but the *security architecture* is recognizable.

**Single hard-decision criterion:** if Branch A prevails, **Wyrd retains its name and database structure, but the privilege model becomes "Branch-A-flavored," not Cayley-Dickson-flavored.** This is a genuine redesign for sections 2-3 of the v0.2 supervisor doc, but the rest survives.

---

## 3. What survives unchanged from v0.2

Even under full Branch A redesign:

| v0.2 component | Status under Branch A |
|---|---|
| Wyrd as a hypergraph database | ✓ Unchanged |
| MuninnDB → Wyrd naming | ✓ Unchanged |
| Skuld as the supervisor | ✓ Unchanged |
| The 9-call user API surface | ✓ Mostly unchanged; types update |
| QBP-CU integration patterns (VCIX/SSCI ports, watchdog event stream) | ✓ Unchanged |
| The 10-instruction ISA | ⚠ Mostly unchanged; QRING semantics shift |
| QFMA, QSAND, QNORM (quaternion ops) | ✓ Unchanged — ℍ component is still there |
| QPERM, QPERMR, QNEAR, QDEC, QREC (Fano/encoding) | ⚠ Replace Fano LUT with M₃(ℂ) operation table |
| QCOMM | ✓ Unchanged (still useful for ℂ-vs-ℍ component distinction) |
| WyrdQuery, WyrdSubscribe semantics | ✓ Unchanged |
| Phased migration Crawl/Walk/Run/Sprint | ✓ Unchanged |
| QW1024 word format | ⚠ Component allocation changes; total width same |
| The watchdog as privilege detector | ✓ Unchanged role; different invariants |
| The C-01..C-12 ticket plan | ⚠ C-01 (ISA freeze) and C-02 (CSR additions) acquire Branch-A-conditional content |

The pattern: **architecture survives, mathematics is replaced.** Skuld doesn't disappear; the proofs change.

---

## 4. What changes under Branch A

### 4.1 The privilege ring assignments

```
v0.2 (Cayley-Dickson):           Branch A:
  Ring 3 (user)       ℂ            Ring 3 (user)        ℂ-module
  Ring 2 (supervisor) ℍ            Ring 2 (supervisor)  ℍ-module
  Ring 1 (kernel)     𝕆            Ring 1 (kernel)      M₃(ℂ)-module
  Ring 0 (firmware)   𝕊            Ring 0 (firmware)    full ℂ⊕ℍ⊕M₃(ℂ)
```

The user ring stays ℂ-flavored. The supervisor stays ℍ-flavored. The kernel changes from octonions to matrix-algebra modules. The firmware now means "access to the full direct sum."

### 4.2 The boundary detectors

```
v0.2:                                  Branch A:
  ℂ ⊂ ℍ:   commutator [a, b]            ℂ-mod ⊂ ℍ-mod:   ℍ-action presence
  ℍ ⊂ 𝕆:   associator (a, b, c)         ℍ-mod ⊂ M₃-mod:  matrix-action presence + trace
  𝕆 ⊂ 𝕊:   alternator [a, a, b]         M₃-mod ⊂ full:   full direct-sum reach
```

The commutator survives at the ℂ → ℍ boundary; quaternions are still noncommutative. The associator-based detection at ℍ → 𝕆 is replaced by "is the matrix-module action being invoked," which is structural (in the type system) plus trace-magnitude (at runtime).

### 4.3 The Lean proof corpus

Major restructuring required:

| v0.2 proof | Branch A status |
|---|---|
| T1.2.a (commutator vanishes in commutative ring) | ✓ Unchanged |
| T1.2.b (associator vanishes in associative ring) | ✓ Unchanged but unused |
| T1.2.c (alternator vanishes in alternative ring) | ✓ Unchanged but unused |
| T2.1.a (no surjection ℂ → ℍ) | ✓ Unchanged — still the user/supervisor boundary |
| T2.1.b (no surjection ℍ → 𝕆) | ✗ Replaced by module-structure analog |
| T2.1.c (no surjection 𝕆 → 𝕊) | ✗ Replaced |
| T2.2 (projection) | ⚠ Replaced with module-projection |
| T2.3 (capability) | ✓ Mostly unchanged; capability is module-structure-token |
| T2.4 (sandwich preservation) | ✗ Sandwich semantics change for matrix algebras |
| T3.x (noise bounds) | ✓ Unchanged |
| T4.x (word integrity) | ⚠ Field allocations change |
| T5.x (meta) | ✓ Unchanged |

About 40-50% of the corpus would need rework.

### 4.4 The QBP-CU hardware

The Fano-plane LUT is octonion-specific. Under Branch A it becomes a 9-element multiplication table for the standard basis of M₃(ℂ) (the 3×3 = 9 matrix units). The hardware shape is the same: a small SRAM with concurrent read ports, populated with structure constants of the relevant algebra. This is the place the SiFive spec's `LUT_capacity_provision = 16` foresight pays off: 16 ≥ max(8 octonion basis, 9 matrix unit basis), so either substrate fits.

The watchdog needs to compute different invariants:
- v0.2: commutator, associator, alternator
- Branch A: commutator (still useful for ℂ vs ℍ), trace of M₃(ℂ) component, module-action invocation flag

These are not strictly more expensive in silicon — the trace is cheaper than the alternator. **The hardware shape from the SiFive spec survives Branch A with an LUT-population swap and a watchdog-invariant swap. No silicon redesign.**

### 4.5 The QW1024 word format

The State and Spin components of QW1024 currently allocate 8 fp32 each (octonion). Under Branch A:

| Component | Branch A allocation |
|---|---|
| Loc.Pos | 4 × fp16 = 64 bits |
| Loc.Vel | 4 × fp16 = 64 bits |
| State.ℂ | 2 × fp32 = 64 bits |
| State.ℍ | 4 × fp32 = 128 bits |
| State.M₃(ℂ) | 18 × fp32 = 576 bits (9 complex entries) |
| Spin (operation queue) | 64 bits (smaller; ops are typed, not algebraic-rich) |
| Metadata + ChainRef + privilege | 64 bits |
| **Total** | 1024 bits |

The M₃(ℂ) component is the largest (576 bits for the 3×3 complex matrix). State and Spin are reorganized but total width is preserved. This means the QW1024 commitment from the workload analysis is robust to the substrate question.

---

## 5. Reframing the verification corpus

Under Branch A, the foundational corpus would be approximately:

```
Wyrd-BranchA-Module-Privilege-Proofs-v0.1.lean
  T1.A — commutator vanishing (user/supervisor)        REUSED from v0.2
  T2.A.1 — module-structure non-synthesis              NEW
  T2.A.2 — module-projection                            NEW
  T2.A.3 — capability as module-action token           ADAPTED from v0.2
  T2.A.4 — module-action mediation (replaces sandwich) NEW
  T3.A — noise bound for trace computation             ADAPTED
  T4.A — word integrity for new component layout       ADAPTED
```

The depth and difficulty are comparable — module theory is well-developed in mathlib. The quaternion-and-octonion-specific machinery from v0.2 doesn't reuse cleanly, but module-theoretic primitives are mature. Honest estimate: 2-3x the time investment compared to v0.2's corpus to reach equivalent rigor.

---

## 6. Crawl-phase hedge strategy

**The decision-theoretic question:** what to do at Crawl, given uncertainty about which substrate wins.

Three options:

### Option H1: Commit to Cayley-Dickson, redesign if Branch A wins later

Build Skuld v0.2 as currently specified. If Branch A wins, redesign substantially.

**Cost:** Skuld v0.2 ships faster. Branch A redesign costs 2-3 months and re-issues several documents.
**Risk:** If Branch A wins, the privilege model rework is expensive.

### Option H2: Build substrate-agnostic infrastructure now, defer privilege model

Build the QBP-CU integration, the Wyrd database, the Skuld API surface — but treat the privilege model as a pluggable component. At Crawl, ship a stub privilege model that uses simple type-tags. Defer the algebraic privilege enforcement to Walk, by which time the substrate question may be settled.

**Cost:** Skuld at Crawl has weaker privilege properties (just type-tags, not algebraic enforcement). The "structurally enforced" headline claim of the architecture isn't yet true.
**Risk:** The stub privilege model becomes the permanent one. Algebraic enforcement is the genuinely novel part of this work; deferring it risks losing the distinguishing property.

### Option H3: Build for both, choose at Walk

Build the substrate-aware components (LUT, watchdog invariants, QW1024 layout) as configurable rather than hardcoded. At Crawl, default to Cayley-Dickson configuration. At Walk, choose based on substrate evidence.

**Cost:** Modest engineering overhead — a few extra abstractions, configuration knobs. Most of the code is shared.
**Risk:** The configurability could become a permanent burden if the substrate question is never decisively settled.

### Recommendation: Option H3 for the QBP-CU; Option H1 for the privilege model

The hardware (LUT, watchdog, ISA) is genuinely substrate-agnostic with modest design care: provision 16-element LUT, design watchdog to compute commutator + (associator OR trace) gated by AlgebraID, design QW1024 with reconfigurable allocation. This is good engineering practice anyway.

The privilege model is harder to build configurably without sacrificing rigor. **Commit to Cayley-Dickson at Crawl** — it's the most likely substrate (Furey's program is the strongest theoretical candidate; Branch A is the spectral-action fallback). If Branch A wins, accept the redesign cost.

This gives the clean Skuld v0.2 architecture for Crawl. The hedge is in the hardware, not in Skuld itself. If the substrate question swings, the hardware adapts and Skuld redesigns. The redesign is bounded — sections 2-3 of the v0.2 supervisor doc, plus 40-50% of the Lean corpus. Recoverable in 2-3 months.

The bet is reasonable: ~80% probability Cayley-Dickson tower (saves 2-3 months at Crawl), ~20% probability Branch A redesign (costs 2-3 months at the redesign point). Expected cost is far less than building substrate-agnostic privilege model from the start, which would slow Crawl by ~6 months for an architecture that may not need the flexibility.

---

## 7. What this analysis settles

**Settled:**
- Branch A is a genuine architectural risk to the v0.2 privilege model.
- The QBP-CU hardware design is substrate-agnostic with modest care.
- Wyrd as a database survives unchanged under either substrate.
- Skuld as a name and a supervisor concept survives.
- Crawl-phase work proceeds as planned in v0.2.

**Open (and worth periodic review):**
- The substrate question itself. Update this document if QBP physics work substantially advances either branch.
- The "stub privilege model" question for Walk if Branch A is ascendant. Don't ship the stub; redesign properly.
- Whether the trace-based detection scheme (Section 2, Option A3) deserves prototyping in Crawl as a hedge — leaning no, but the question is open.

**Triggers for revisiting this document:**
- New theoretical paper on the QBP substrate.
- Empirical measurement that decisively favors one branch (e.g., dark matter detection at predicted QBP energy).
- BMA or Materia work that needs substrate-specific assumptions.
- Quarterly re-read regardless.

---

## 8. References

**Internal:**
- *Wyrd-Supervisor-Architecture-v0.2.md* — the architecture this hedges against
- *QBP-Dark-Matter-Fork-Analysis.md* — the canonical Branch A vs Branch B reference
- *Wyrd-Workload-ISA-v0.1.md* — workload analysis driving substrate-agnostic ISA decisions

**External:**
- Chamseddine, Connes (2007). The Spectral Action Principle. — Branch A foundational
- Furey, C. (2018). Standard Model from Octonions. — Cayley-Dickson tower foundational
- Boyle, Farnsworth (2014). New Algebraic Approach to Standard Model — bridges branches
- Dixon, G. (1994). Division Algebras: Octonions Quaternions Complex Numbers — extended Cayley-Dickson

**Attribution per QBP standing rule:** Furey, Dixon, Günaydin/Gürsey, Boyle/Farnsworth, Singh, Chamseddine/Connes, Koide, Baez.

---

*End of Branch A Contingency Analysis v0.1 — DRAFT*
