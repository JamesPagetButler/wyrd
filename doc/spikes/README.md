# CTH-into-Wyrd migration spike — record

Durable, doc-only record of the **CTH-into-Wyrd migration spike** (two sub-spikes, A and B). Per spike rule 1, the spike code itself is not here — it lives only on branch `spike/cth-wyrd-migration` (not merged to `main`). This directory holds the pre-registration and, once run, the results and raw evidence.

## Origin

`inter#324` — beekeeper proposal (2026-11-27, in session with cth-implementor), following from `confluent-trust#139`'s empirical composability findings. The beekeeper wants real stress-testing of Wyrd-as-CTH-substrate started now, deliberately ahead of any single consumer's current need (Gleipnir's tactic (d) doesn't require this — `live-test` seq 5972).

## Roles

- **Director:** `wyrd-implementor` (both spikes — the open questions are symmetric Wyrd-side ones; agreed on `inter#324`/`live-test` seq 5977–5980).
- **Owner / feasibility, Spike A (Gleipnir):** `cartan`, with `harlequin` as data owner (`live-test` seq 5978).
- **Owner / feasibility, Spike B (QBP):** `qbp-oppenheimer`, with `qbp-implementor` (`live-test` seq 5979).
- **Named runner:** not yet accepted — pending.
- **Box-safety clearance:** not yet requested — pending (trivial resource footprint expected; see pre-reg §Resource estimate).
- **§I4 reviewer:** `qbp-architecture`.

## Status

**DRAFT.** Not frozen, not run. Sequencing is **after Sprint 4** (hard ordering constraint, not negotiable) and the data freeze points named in the pre-reg (Spike A: after Gleipnir tactic (c) closes; Spike B: at the QBP master commit that closes Sprint 4) have not yet occurred. See `prereg/cth-wyrd-migration-prereg-DRAFT-v1.md`.

## Cross-links

- `inter#324` (origin, open questions, role acceptances)
- `confluent-trust#139` (composability design surface this spike stress-tests against real data)
- `doc/integration/cth.md` (the Anchor/Chain/Confluence→Node/Hyperedge mapping sketch this spike implements for the first time)
- `lean/Wyrd/HolographicHypergraph.lean` (`theorem2_irreducibility` — the claim Spike A/B's confluence data exercises)
- Gleipnir#181, QBP#464 Phase A (the inverse anchor audit qbp-oppenheimer named as Spike B's comparison workload)
