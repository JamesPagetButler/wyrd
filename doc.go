// Package wyrd is the top-level metadata package for the Wyrd hypergraph
// database. The actual data model lives in subpackages:
//
//   - [github.com/JamesPagetButler/wyrd/model]   — typed hypergraph data structures
//   - [github.com/JamesPagetButler/wyrd/compute] — algebraic-privilege checks, bridge promotion, consistency
//   - [github.com/JamesPagetButler/wyrd/store]   — persistence backends (JSON for Crawl; future: MuninnDB, SurrealDB)
//
// # Soundness substrate
//
// Wyrd's runtime contracts are formally verified in the Lean 4 corpus at
// the repo's `lean/` directory. Each Go API that has a load-bearing
// invariant cites the corresponding Lean theorem in its doc comment.
// The corpus index is `Wyrd-Proofs-Reference-v1.4.md` (see `doc/archive/`).
//
// # Phases
//
//   - Crawl (v0.1.x): in-memory + JSON persistence; pure-Go model and compute.
//   - Walk  (v0.2.x): MuninnDB-flavoured persistence, NATS event integration,
//     QBP-CU emulator hardware-accelerated quaternion arithmetic.
//   - Run   (v0.3.x): SurrealDB structural ground truth, Skuld supervisor
//     enforcing algebraic privilege at the hardware boundary.
//
// Wyrd is consumed by:
//
//   - confluent-trust (CTH): hyperedge storage for trust-anchor inventories
//   - bma-systema (BMA): cognitive-layer hypergraph (autonomic / subconscious / conscious)
//   - Contextus: cross-domain pattern matching with InsightSignals as hyperedges
//
// Each consumer's integration guide is in `doc/integration/`.
package wyrd
