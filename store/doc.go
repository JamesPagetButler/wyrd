// Package store provides persistence backends for Wyrd graphs.
//
//   - [JSONFile] — Crawl-phase: marshal/unmarshal a Graph to a JSON file.
//     Pure stdlib; round-trip stable; no external deps.
//
// Future backends (deferred to Walk/Run phases):
//
//   - MuninnDB — co-activation-aware engram store with Hebbian / Ebbinghaus dynamics
//   - SurrealDB — structural ground truth for the federation
//   - HAMA     — holographic-medium store for Tier-N consolidated memory
package store
