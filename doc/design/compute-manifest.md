# Compute Manifest v0.1 — Wyrd-owned operational document naming the federation's blessed compute substrate

**Status:** Design **v0.1** — open for review per ADR-003 §I4
**Tracks:** Spec 9.2 §11 Toddle deliverable ("Compute Manifest v0.1 authored in `repo-wyrd`"). Precursor to `repo-bma-systema-issue-#170` (Translation Functor cycle-counter cross-phase invariant) and `repo-bma-systema-issue-#171` (Spec 9.2 §3 credibility-window amendment).
**Governance anchor:** ADR-003 §I4; Spec 9.2 §4 (Compute Manifest as Wyrd-owned operational document); Spec 9.2 §11 (Toddle deliverable list); parent `repo-bma-systema-issue-#164` (Federation Lean Promotion Protocol)
**Authors:** wyrd-implementor

---

## 0. §I4 invariant — design-doc-as-S-01-review-surface

This document is the §I4 review surface for the Wyrd-owned Compute Manifest. Implementation PR blocked on explicit sign-off from named reviewers (§9).

The Compute Manifest is the single operational document where the federation names "the blessed compute substrate" referenced in Spec 9.2 §3 (the Compute-Substrate Gate). Two downstream sub-issues of `repo-bma-systema-issue-#164` directly depend on this artefact existing:

- `repo-bma-systema-issue-#170` references the manifest in its `Translation Functor invariant statement (substrate-tier per A22 §4.2)` (load-bearing commitment 1) to enumerate the phases the invariant must hold across.
- `repo-bma-systema-issue-#171` adds a schema extension to this manifest (`last_passing_tier_b`) per its closes-when criterion 2, and a CI gate per criterion 3.

Shipping the manifest base shape first means #170 and #171 land on stable ground, not a moving target.

## 1. Motivation

Spec 9.2 §4 makes the manifest a normative deliverable but leaves authoring to the Toddle-phase substrate owner (`repo-wyrd`). It is currently a paper concept with a table in the spec:

| Phase | Compute Manifest names |
|---|---|
| Crawl / Toddle | QBP-CU emulator (Go library) |
| Walk | QBP-CU M1 Gearbox (CSR-bound stateful + QW8 + QW128) |
| Run-initial | QBP-CU M2 ternary matmul + ROCm acceleration |
| Run-mature+ | Possibly QBP-CU silicon (per `workspace-phase-architecture.md` §0.13.2) |

What does NOT exist:

1. **A file at a canonical path** the federation CI can load to know which substrate is current.
2. **A schema** other tools (federation CI, mode-(b) gate per #171, substrate-transition PRs per Spec 9.2 §5) can validate against.
3. **A Go type** consumers in `repo-wyrd` (and downstream `repo-bma-systema` via the BMA-Wyrd integration) can reference symbolically.

Without these, every consumer hand-codes "the substrate is QBP-CU emulator" as a string literal. The silicon exit ramp (§4 of the spec) — the property that *"the gate does not change; the Manifest does"* — relies on this single artefact being substitutable.

**Crawl-shippable framing:** the manifest is small (a YAML file + a Go struct + a loader + tests). It is the cheapest unblocker for the rest of the §164 chain, and lands cleanly inside Toddle phase per Spec 9.2 §11.

## 2. Decision — `manifest/compute-manifest-v0_1.yaml` + `model/compute_manifest.go`

### 2.1 Canonical file location

```
repo-wyrd/
├── manifest/
│   └── compute-manifest-v0_1.yaml      ← canonical data, version-pinned in filename
└── model/
    ├── compute_manifest.go              ← schema + loader
    └── compute_manifest_test.go         ← round-trip + validation
```

Versioning in the filename (`v0_1`) follows the Lean / spec versioning convention already used in `inter/spec/BMA-Spec-Addendum-9_2-...md`. The v0.2 form (post-#171 credibility-window extension) ships as `compute-manifest-v0_2.yaml` with the v0.1 retained as historical reference. **Wyrd's federation CI always loads the latest version** based on a single-line `manifest/CURRENT` pointer file (next paragraph).

A pointer file `manifest/CURRENT` contains the basename of the current manifest:

```
$ cat manifest/CURRENT
compute-manifest-v0_1.yaml
```

This decouples consumers from filename pinning. The pointer is updated atomically as part of any version-bump PR; consumers tail the pointer not the file.

### 2.2 Canonical YAML schema (v0.1)

```yaml
# manifest/compute-manifest-v0_1.yaml
#
# The Federation Compute Manifest names the blessed compute substrate
# per Spec 9.2 §3 (the Compute-Substrate Gate). One file, one truth.

version: "v0.1"
authored_at: "2026-05-15T00:00:00Z"

phase: "crawl"
# One of: crawl | toddle | walk | run-initial | run-mature
# Per Spec 9.2 §4 phase table. The phase value MUST match the
# beekeeper's declared workspace phase; CI cross-checks against
# the BMA Spec workspace phase declaration at federation review time.

substrate:
  name: "QBP-CU emulator"
  kind: "emulator"
  # One of: emulator | gearbox-cpu | gpu-accelerator | silicon
  repo: "github.com/JamesPagetButler/qbp-compute-unit"
  module: "emulator"
  # Go module path inside repo (for substrates that expose a Go API)
  commit_sha: "TBD-pinned-at-PR-time"
  # Full 40-char Git SHA. Each PR that bumps the manifest pins this.
  pinned_tag: "v0.1.0"
  # Optional human-friendly version label

# v0.2 extension placeholder (per repo-bma-systema-issue-#171):
# credibility:
#   last_passing_tier_a: { timestamp: "...", substrate_commit_sha: "..." }
#   last_passing_tier_b: { timestamp: "...", substrate_commit_sha: "..." }
#
# These fields are defined in #171's amendment + schema extension and
# land on this manifest in v0.2. v0.1 omits them; mode-(b) promotion
# is best-effort during Crawl/Toddle (Spec 9.2 §3.1 amendment text).
```

### 2.3 Go schema and loader

```go
// Package model: Federation Compute Manifest.
//
// LoadComputeManifest reads the Wyrd-canonical Compute Manifest YAML
// at the path indicated by `manifest/CURRENT` and returns a typed
// snapshot of the federation's blessed compute substrate. Pattern
// reuses store.LoadScopeConfig's two-phase load (validate then commit
// per ADR-003 §I3).
//
// Soundness anchor: forthcoming Wyrd.ComputeManifest.manifest_load_atomic
// (proven; CI Phase 2 gate). Mirrors scope_loader_atomic structure.
package model

import (
    "time"
)

// ComputeManifestPhase is the federation phase tag per Spec 9.2 §4.
type ComputeManifestPhase string

const (
    PhaseCrawl       ComputeManifestPhase = "crawl"
    PhaseToddle      ComputeManifestPhase = "toddle"
    PhaseWalk        ComputeManifestPhase = "walk"
    PhaseRunInitial  ComputeManifestPhase = "run-initial"
    PhaseRunMature   ComputeManifestPhase = "run-mature"
)

// SubstrateKind classifies the blessed substrate's implementation form.
type SubstrateKind string

const (
    SubstrateEmulator       SubstrateKind = "emulator"
    SubstrateGearboxCPU     SubstrateKind = "gearbox-cpu"
    SubstrateGPUAccelerator SubstrateKind = "gpu-accelerator"
    SubstrateSilicon        SubstrateKind = "silicon"
)

// Substrate names the blessed compute substrate.
type Substrate struct {
    Name       string        `yaml:"name"       json:"name"`
    Kind       SubstrateKind `yaml:"kind"       json:"kind"`
    Repo       string        `yaml:"repo"       json:"repo"`
    Module     string        `yaml:"module"     json:"module,omitempty"`
    CommitSHA  string        `yaml:"commit_sha" json:"commit_sha"`
    PinnedTag  string        `yaml:"pinned_tag" json:"pinned_tag,omitempty"`
}

// ComputeManifest is the v0.1 typed snapshot.
type ComputeManifest struct {
    Version    string               `yaml:"version"     json:"version"`
    AuthoredAt time.Time            `yaml:"authored_at" json:"authored_at"`
    Phase      ComputeManifestPhase `yaml:"phase"       json:"phase"`
    Substrate  Substrate            `yaml:"substrate"   json:"substrate"`
    // Credibility is reserved for v0.2 (per repo-bma-systema-issue-#171).
    // Zero-value during v0.1; absence means "best-effort mode-(b) per
    // Spec 9.2 §3.1 amendment text — Crawl/Toddle only".
}

// LoadComputeManifest reads the manifest file pointed to by manifest/CURRENT
// under root, validates it, and returns the typed snapshot. Returns
// ErrComputeManifestInvalid on validation failure. Returns wrapped
// fs / parse errors for I/O issues.
func LoadComputeManifest(root string) (*ComputeManifest, error)

// LoadComputeManifestReader reads a manifest from any io.Reader (for
// HTTP / S3 / test fixtures). Both forms share the same validation.
func LoadComputeManifestReader(r io.Reader) (*ComputeManifest, error)
```

### 2.4 Validation rules (v0.1)

1. `version` non-empty and matches `^v[0-9]+\.[0-9]+$`.
2. `authored_at` is parseable RFC-3339 timestamp; not in the future.
3. `phase` is one of the five declared `ComputeManifestPhase` values.
4. `substrate.name` non-empty.
5. `substrate.kind` is one of the four declared `SubstrateKind` values.
6. `substrate.repo` matches `^github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$` (Crawl-phase tightening; v0.2 may relax).
7. `substrate.commit_sha` matches `^[0-9a-f]{40}$` OR equals `"TBD-pinned-at-PR-time"` (the bootstrap sentinel — the federation CI rejects this in any non-bootstrap PR).

Validation failure returns `ErrComputeManifestInvalid` wrapping a specific cause via `fmt.Errorf("...: %w", ErrComputeManifestInvalid)`. Consumers unwrap with `errors.Is` per stdlib convention.

### 2.5 The bootstrap sentinel

The very first PR landing this manifest cannot yet name a real `commit_sha` for the QBP-CU emulator that has been federation-blessed (no prior blessing exists). To avoid a chicken-and-egg deadlock, v0.1 accepts the literal string `"TBD-pinned-at-PR-time"` as the commit_sha *only when CI is running on the manifest-bootstrap PR itself* (detected by branch name match against `manifest/compute-manifest-v0_1`). All subsequent PRs must pin a real SHA. The bootstrap-PR concession is documented in §11 below for traceability.

## 3. Not in v0.1

- **Credibility-window fields** (`last_passing_tier_a`, `last_passing_tier_b`) — these are `repo-bma-systema-issue-#171`'s deliverable. v0.1 reserves the schema slot via a top-level placeholder comment but does not require the fields. v0.2 lands them.
- **Multi-substrate phases.** Walk-α may want multi-tenant substrates (one for QBP-CU M1 Gearbox, one for ROCm GPU acceleration on the same workload); v0.1 names a single substrate. Walk concern; not Toddle.
- **Substrate-transition automation.** Spec 9.2 §5 says deprecations remain proved; downstream tenant proofs continue to verify. The mechanics of writing a substrate-transition PR (Crawl → Walk) are a Walk-phase concern; v0.1 just gives that PR a single file to amend.
- **In-tree substrate vs. external substrate.** v0.1 substrates can only be external (referenced by `repo` + `commit_sha`). An in-tree substrate (e.g., a Wyrd-internal mock for testing) is excluded; tests use `LoadComputeManifestReader` with a fixture instead.
- **Cryptographic attestation of substrate identity.** v0.1 trusts the Git SHA. Signing the manifest with a beekeeper YubiKey is a Walk concern (per `repo-bma-systema-issue-#34` YubiKey-Walk-phase tracking).

## 4. Soundness anchors

- **`Wyrd.ComputeManifest.manifest_load_atomic`** (forthcoming, lands with impl PR). Proof structure: validation is a pure predicate on parsed YAML; load is either-validate-and-return or return-error. Trivially atomic since there's no graph mutation; ~15 LOC Lean estimate.
- **ADR-003 §I3** — N/A for this file (no graph mutation). The atomicity here is the file-load itself.
- **Spec 9.2 §4** — the manifest table is normative; this design encodes it as schema. Any change to the spec table requires a manifest schema version bump.

## 5. Cross-repo coordination

The Compute Manifest is Wyrd-owned per Spec 9.2 §4, but several federation peers consume it:

| Concern | Owner | Deliverable |
|---|---|---|
| Manifest YAML + Go schema + loader (this design) | wyrd-implementor | `repo-wyrd/manifest/`, `repo-wyrd/model/compute_manifest.go` |
| Credibility-window fields (v0.2) | wyrd-implementor (#171) | `last_passing_tier_{a,b}` + CI gate |
| Substrate publishing Tier A / Tier B verification timestamps | qbp-cu-implementor | `repo-qbp-compute-unit` exposes signed timestamps consumed by #171's amendment |
| BMA runtime consumption (`bma compute-manifest current`) | bma-implementor | reins wrapper reads this manifest |
| Substrate-transition PR pattern | qbp-architecture | guidance doc when Crawl → Walk transition happens |

`@bma-implementor` can stub a reins wrapper against the API signature defined here **starting now**; no need to wait for impl PR. Same pattern as scope-loader (§5 of `doc/design/scope-loader.md`).

## 6. What this design PR ships

Only this design doc. The impl PR (`manifest/compute-manifest-v0_1` branch → main) ships:

```
manifest/compute-manifest-v0_1.yaml          — canonical data
manifest/CURRENT                             — single-line pointer
model/compute_manifest.go                    — schema + loader
model/compute_manifest_test.go               — round-trip + validation tests
lean/Wyrd/ComputeManifest.lean               — manifest_load_atomic (proven; no sorry)
doc/integration/compute-manifest.md          — usage sketch for BMA + CI consumers
```

**Typed errors exported from `model/compute_manifest.go`:**

```go
var (
    ErrComputeManifestInvalid = errors.New("model: compute-manifest validation failed")
    ErrComputeManifestParse   = errors.New("model: compute-manifest parse error")
    ErrComputeManifestMissing = errors.New("model: compute-manifest CURRENT pointer missing")
)
```

## 7. Open questions for §I4 reviewers

1. **Schema location: `manifest/` top-level vs. `internal/manifest/` vs. `model/manifest/`.** My lean: top-level `manifest/`, because the YAML is a federation artefact (consumed by external tools — `gh`, federation CI, possibly BMA reins from another repo) and a top-level location signals that. Pushback OK.
2. **Pointer file (`manifest/CURRENT`) vs. fixed filename.** My lean: pointer file, because v0.2 lands a different filename (`compute-manifest-v0_2.yaml`) and we want consumers to follow the pointer rather than chase the filename. Cost: one extra file. Pushback OK.
3. **YAML vs. JSON as canonical.** My lean: YAML, matching `manifest/compute-manifest-v0_1.yaml` being a hand-authored beekeeper-edited file. Auto-generated tooling reads/writes either via the same Go loader. Same call as scope-loader §7 Q1.
4. **`commit_sha` bootstrap sentinel.** My lean: accept `"TBD-pinned-at-PR-time"` only on the bootstrap PR's CI run; reject elsewhere. Cleanest unblocker. Pushback if `@qbp-cu-implementor` wants to pin a real SHA from the start (in which case the bootstrap PR explicitly co-lands a tagged release of `repo-qbp-compute-unit/emulator` for SHA-pinning).
5. **Phase-vs-substrate independence.** Should the loader cross-check that `(phase, substrate.kind)` is one of the legal Spec 9.2 §4 table combinations (e.g., reject `phase: crawl` with `substrate.kind: silicon`)? My lean: **yes, validate at load time** — the spec table is normative and a wrong pairing means a paperwork bug. Pushback OK.

## 8. Migration path

1. Land this design doc — §I4 sign-off from named reviewers (§9).
2. Open impl PR (`manifest/compute-manifest-v0_1` branch); CI green; round-trip tests pass; bootstrap SHA pinned (or sentinel accepted per §2.5).
3. Land `repo-bma-systema-issue-#171` v0.2 schema extension: credibility fields + CI mode-(b) gate (own PR chain per the parent plan).
4. Land `repo-bma-systema-issue-#170` Translation Functor invariant Lean encoding (the first substrate-tier theorem that names this manifest in its statement).
5. (Walk-α) substrate-transition PR template lands — uses this manifest as the single point of change for Crawl → Walk Gearbox.
6. (Run-initial+) silicon exit-ramp PR pattern, same template, different substrate kind.

## 9. §I4 named reviewers

Per the parent issues' reader-list (#170 + #171 both name the same five):

- `@wyrd-implementor` — author (substrate ownership; Compute Manifest authoring)
- `@qbp-cu-implementor` — substrate publisher (the `qbp-compute-unit/emulator` named in v0.1)
- `@bma-implementor` — runtime consumer (BMA reins wraps the loader; future federation gate consumer)
- `@qbp-architecture` — federation-coherence + Spec 9.2 §4 owner
- `@beekeeper` — HVR on Toddle-deliverable landing

## 10. Items NOT decided here

- **The bootstrap PR's pinned `commit_sha` value.** Determined at impl-PR time; coordinated with `@qbp-cu-implementor` to identify the right emulator commit (latest passing PR-gating tests in `repo-qbp-compute-unit` at impl-PR open time).
- **Substrate-transition cadence.** Crawl → Walk timing is a beekeeper-directed phase decision, not a manifest concern. The manifest just names the current substrate; when it's amended, what triggered the amendment is a phase-transition concern.
- **What happens if the manifest disagrees with reality.** If `manifest/CURRENT` names QBP-CU emulator commit X but the workspace's installed Go module is at commit Y, that's a runtime drift the manifest cannot detect. Detection is a future v0.x concern (probably part of `repo-qbp-compute-unit`'s self-reporting, not this manifest).

## 11. Cross-references

- `repo-bma-systema-issue-#164` — A21.0 Federation Lean Promotion Protocol (parent of #170 + #171)
- `repo-bma-systema-issue-#170` — Translation Functor §4.2 substrate-tier invariant (downstream consumer of this manifest)
- `repo-bma-systema-issue-#171` — Spec 9.2 §3 credibility-window amendment (extends this manifest's schema)
- `inter/spec/BMA-Spec-Addendum-9_2-Federation-Lean-Promotion-Protocol.md` §4 (Compute Manifest definition), §11 (Toddle deliverable list)
- `inter/theory/BMA-Theory-Addendum-22_0-Cross-Tenant-Autonomic-Translation-Layer.md` §4.2 (substrate-tier promotion criteria — A22 §4.2)
- `inter/workspace-phase-architecture.md` §0.13.2 (silicon exit ramp; informs Walk → Run-mature transition)
- `repo-wyrd/doc/design/scope-loader.md` — template for the two-phase load pattern reused here
- `repo-wyrd/store/scope_loader.go` (PR #49) — implementation of the two-phase load pattern
- ADR-003 §I3 (atomicity), §I4 (design-doc-as-S-01-review-surface)

---

*Status: DRAFT v0.1 — open for §I4 review. Implementation PR blocked on explicit sign-off from `@qbp-cu-implementor`, `@bma-implementor`, `@qbp-architecture`, and the beekeeper.*
