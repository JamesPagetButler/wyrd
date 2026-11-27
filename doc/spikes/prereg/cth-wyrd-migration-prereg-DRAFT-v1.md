# CTH-into-Wyrd migration spike — pre-registration (DRAFT v1, NOT FROZEN)

**Director:** wyrd-implementor
**Origin:** `inter#324`; `confluent-trust#139`
**Status:** DRAFT — not yet eligible for a run go. See §Confirmation gates.
**Numbering note:** `inter#324` was restructured by the beekeeper (2026-11-27) into three sequenced rounds — Round 1 = Spike A (Gleipnir) below, Round 2 = Spike B (QBP) below, Round 3 = Edda access (runs on Nidavellir). This pre-reg covers Rounds 1–2 only; Round 3 is out of scope here per the Edda-decoupling already agreed, and gets its own pre-reg once Rounds 1–2 land.

---

## 0. Why this is still a draft, not a freeze

A FROZEN pre-reg needs concrete, pinned facts this draft cannot yet state:

- **Spike A's data freeze point** — "after Gleipnir tactic (c) closes" (cartan, `live-test` seq 5978) has not occurred. No commit sha can be pinned yet.
- **Spike B's data freeze point** — "the QBP master commit that closes Sprint 4" (qbp-oppenheimer, seq 5979) has not occurred.
- **Sequencing** — both spikes are gated "after Sprint 4" (beekeeper, `inter#324`), a hard ordering constraint per `housekeeping-before-sprint`, not a target to negotiate around.
- **No named runner has accepted yet.**
- **No box-safety clearance from deming yet.**

Everything else below (question, method, success criteria, resource estimate) is settled and will not change at freeze — only the pinned shas, runner name, and deming's clearance get added. When all five above are in hand, this document is replaced in place by a `-FROZEN-v1` version, committed, sha256-posted, and run go follows per the three-gate rule.

---

## 1. Question

**One shared question, two independent instances:** does Wyrd's existing (but never-exercised-on-real-data) typed-hypergraph mapping for CTH data — `model.Anchor`→`Node`, `model.Chain`→`Hyperedge`, `model.Confluence`→arity-≥3 `Hyperedge` (`doc/integration/cth.md`) — actually hold when fed a real inventory, with no silent data loss and no mapping gap the design sketch didn't anticipate?

This is an **E-move, not an F-move** (per the spike-run grading model): it confronts an existing design claim with real data. It does not attempt to prove anything new, and `F` stays `napkin` regardless of outcome — the sketch in `doc/integration/cth.md` is today unconfirmed against any real inventory.

- **Spike A:** run the question against a frozen test copy of Gleipnir's CTH inventory.
- **Spike B:** run the question against a frozen test copy of QBP's CTH inventory (`archive/cth-inventory/confluent-trust-inventory-v5_3.v0.3.json` in the QBP repo — confirmed present, 953,603 bytes at the time of this draft).

Both spikes use the same migration code (same spike branch, same adapter), run twice against two independently-schemaed real inventories. A result that holds for one but not the other is itself informative (schema-specific, not Wyrd-generic).

**Explicitly out of scope:** Edda access (deferred as a smaller follow-on per `inter#324` agreement, seq 5977–5980); any change to either repo's live canonical ledger; any claim about Gleipnir tactic (d)'s actual needs (already settled as not requiring this, seq 5972).

---

## 2. Method

1. **Freeze.** Take a frozen, read-only snapshot of each inventory at its named freeze point (commit sha pinned at freeze time, recorded in the FROZEN pre-reg — never the live ledger).
2. **Enumerate every top-level node-bearing collection** in each inventory first, explicitly — not just `anchors`/`chains`. (This is the specific trap `cth-implementor` hit on `confluent-trust#139`: both Gleipnir's and QBP's real inventories have node-bearing collections beyond the obvious two — axioms, inputs, derived_principles, confluence_points, etc. A migration that only walks the obvious two reproduces the exact silent-undercount failure the round-trip-count discipline exists to catch.) The enumerated list is recorded in the run output before any mapping code runs.
3. **Migrate.** Build the adapter described in `doc/integration/cth.md` (`model.Anchor`→`model.Node`, `model.Chain`→`model.Hyperedge`, `model.Confluence`→arity-≥3 `model.Hyperedge`) against `model/graph.go`'s `Graph` via direct Go API calls (`g.AddNode`, `g.AddHyperedge`) — **no Edda involvement**, per the agreed decoupling.
4. **Round-trip.** Walk the resulting `Graph` back out and reconstruct a comparable structure; diff counts per collection against the input enumeration from step 2.
5. **Comparison workload (Spike B only, qbp-oppenheimer's ask, seq 5979, precision notes seq 5986):** re-run QBP's existing inverse anchor audit (`scripts/anchor_inverse_audit.py`, QBP#464 Phase A — confirmed: it's the `DEFAULT_LEDGER` consumer, real path, 953,603 bytes at `origin/master`) once over the raw JSON inventory and once over the Wyrd-backed copy; expect identical results.
   - The audit takes `--ledger <path>` (JSON) and `--proofs-dir proofs` (the Lean corpus). "Over the Wyrd copy" means: **export** the Wyrd copy back to JSON, then run the audit with `--ledger <export>` against the **same pinned** `proofs/` tree used for the JSON-baseline run. One QBP sha covers both the inventory and `proofs/`.
   - **This audit's `load_anchors` reads only anchor-bearing keys.** A match is evidence of **anchor round-trip fidelity only** — it does not cover the other node-bearing collections from step 2. Success criterion 1 (full per-collection enumeration) still has to carry the rest; this comparison workload is a *supplement* to criterion 1, not a substitute for it.
   - Run as `python3 -I` under `boxrun`; write output to the run directory only, never into `analysis/`; do **not** pass `--update-baseline`.
6. **Comparison workload (Spike A, cth-implementor's suggestion, seq 5976):** run Gleipnir tactic (d) round 1's three passes (proven → research → napkin) once over the JSON slices and once over the Wyrd/Edda-shaped copy at the next post-tactic-(c) window; expect identical answers. (This step depends on tactic (d) actually running in that window — recorded as a stretch success criterion, not a blocking one, since it is downstream of Gleipnir's own schedule, not this spike's.)
7. **Exercise the irreducibility claim.** For each inventory's real `confluence_points`/multi-source chains (not a synthetic example), check arity and confirm whether `theorem2_irreducibility`'s "not losslessly pairwise-decomposable" case is actually reached, or whether real confluences are all arity-2 in practice.
8. **Tier check.** For every migrated Anchor, confirm its CTH tier (axiom/proof/measurement/prediction) maps onto one of the Wyrd algebraic tiers the sketch anticipates (`TierComplex`/`TierQuaternion`) with no case the sketch doesn't cover.
9. **Schema-direction check (cth-implementor's fourth criterion, seq 5980).** Record whether `chain_source_to_target`/`prediction_chain`-style direction mismatches (already found on both inventories by `confluent-trust#139`) manifest as a structural problem once the data is in Wyrd's typed form, informing whether flow-direction needs to be a typed property at the Wyrd level rather than a convention.

---

## 3. Success criteria

1. **Round-trip fidelity per node-bearing collection** (not just anchors+chains) — zero silent drops; any intentional omission is stated, never silent.
2. **The n-ary-hyperedge claim is actually exercised** by real confluence data in at least one of the two inventories (or: recorded as not exercised, which is itself a valid, informative result).
3. **The CTH/Wyrd tier mapping holds** on real data with no uncovered case.
4. **The schema-direction question is answered empirically** (does it need to be a typed Wyrd-level property, informing `confluent-trust#139`'s open design question).
5. *(Stretch, non-blocking)* Spike A's tactic-(d)-round-1 comparison and Spike B's inverse-anchor-audit comparison both give identical answers against the JSON baseline and the Wyrd copy.

### 3.1 Known data-quality caveats on the source data (cth-implementor's pre-Sprint-4 napkin spike, `inter#324` issuecomment-7225663192)

Three gaps surfaced by an independent pre-Sprint-4 discovery spike against the real inventories, disposed here rather than left implicit:

- **`confluent-trust#141` (node-universe-undercount, third repro).** Directly relevant to method step 2 and success criterion 1 — this is the exact trap this pre-reg already names as a risk. If a canonical "walk every node-bearing collection" helper lands before this spike runs, use it in step 2 instead of an ad hoc enumeration; if it hasn't landed, step 2's explicit enumeration (already planned) stands on its own and is not blocked by #141. Either way, record which path was used.
- **`Gleipnir#186`** (12/42 proof-grade anchors, 29%, missing `tier_F`) **and `QBP#844`** (0/349 anchors have any formality-tier field). **Neither blocks this spike's gates.** These concern `tier_F` — the **formality grade** (napkin/research/formal) tactic-(d)'s own proven-slice filter reads — which is a **different field** from this spike's success criterion 3, the **CTH `Anchor.Tier`** (axiom/proof/measurement/prediction) that maps onto Wyrd's algebraic tier per `doc/integration/cth.md`. Worth stating explicitly so the two "tier" concepts aren't conflated by a future reader: this spike's migration and round-trip do not depend on `tier_F` being populated at all. They matter to tactic-(d)'s interim and to Sprint-4 planning on `inter#194`, not to this pre-reg's confirmation gates.

A **FAIL on criterion 1** (silent, undisclosed data loss) is the one outcome that blocks this from being recorded as a usable result at all — it would mean the migration itself is wrong, not that the question was answered unfavorably. Every other criterion can land either way and still close the spike with a grounded answer.

---

## 4. Resource estimate (`pre-run-resource-estimate` hard gate)

- **Input size:** Gleipnir's inventory ~325 KB JSON (seq 5955, growing ~40 KB/tactic); QBP's inventory 953,603 bytes JSON (confirmed, this draft). Combined well under 1.5 MB at freeze time.
- **In-memory footprint:** parsing both JSON files plus building the equivalent `Graph` (nodes + hyperedges + incidence index, each a small multiple of the source data) is generously bounded at 10× input size — under 15 MB. No GPU, no large matrices, no training loop.
- **Disk:** the frozen snapshots themselves (≤1.5 MB) plus any raw-output JSON the run commits (same order of magnitude). No disk growth beyond low single-digit MB.
- **CPU/threads:** single-threaded Go program; no parallelism needed for data this size.
- **Runtime:** parsing + migration + round-trip + audits, all pure in-memory/CPU work on sub-MB inputs — expected well under one minute total, likely single-digit seconds.
- **Headroom:** this is far below 50% of any reasonable box constraint; no explicit kill condition (ulimit, depth bound) is proposed beyond the standard `boxrun` default-core wrapper, since there is no realistic runaway case in this workload shape (fixed, small, pre-enumerated input; no recursion over unbounded external data; no network calls). Flagging this reasoning explicitly for deming's box-safety review rather than asserting it unchecked.

---

## 5. Time-box

Knowledge-limited: done when the question (§1) is answered for both inventories, or when a structural blocker is found that the migration can't resolve without a design change (in which case the spike closes with that finding, not a workaround). Capped at one working session of implementor time (~4 hours) from run go; an overrun requires a signed amendment, not a silent extension.

---

## 6. Confirmation gates (must all tick before this becomes FROZEN + eligible for run go)

- [ ] Spike A's data freeze point reached (Gleipnir tactic (c) closed) and its snapshot sha pinned.
- [ ] Spike B's data freeze point reached (QBP's Sprint-4-close commit) and its snapshot sha pinned.
- [ ] Sprint 4 closed federation-wide (sequencing constraint).
- [ ] Named runner accepted.
- [ ] deming's box-safety clearance posted against this resource estimate.
- [ ] §I4 co-sign from qbp-architecture on the frozen text.

Only once every box above is ticked does this document get replaced by a `-FROZEN-v1` version, committed + pushed, its sha256 posted, and a run go issued citing all three run-go gates (runner accepted, FROZEN pre-reg sha posted, deming's box-safety clearance).
