# Integrating Wyrd's Compute Manifest

Status: Toddle-phase deliverable per Spec 9.2 §11 (currently shipped). Walk-α adds the credibility-window extension per `repo-bma-systema-issue-#171`; Run-α adds substrate transitions per Spec 9.2 §5.

## What the Compute Manifest is

The **Federation Compute Manifest** is the Wyrd-owned operational document where the federation names "the blessed compute substrate" referenced in Spec 9.2 §3 (the Compute-Substrate Gate). One file, one truth:

- Canonical data: `manifest/compute-manifest-v0_1.yaml`
- Single-line pointer: `manifest/CURRENT`
- Go schema + loader: `model/compute_manifest.go`
- Lean soundness anchor: `lean/Wyrd/ComputeManifest.lean`

Consumers reach the manifest through three loader entry points:

```go
import "github.com/JamesPagetButler/wyrd/model"

// Strict default: rejects the bootstrap sentinel; consumers should
// expect this to fail with ErrComputeManifestInvalid on the very
// first manifest PR (Phase A impl-1 commit) until a real
// substrate_commit_sha is pinned.
m, err := model.LoadComputeManifest(root)

// Bootstrap-aware: federation CI sets AllowBootstrapSentinel = true
// only when running on the manifest-bootstrap PR's branch (per design
// doc §2.5).
m, err := model.LoadComputeManifestWithOptions(root, model.LoadOptions{
    AllowBootstrapSentinel: true,
})

// io.Reader form: useful for test fixtures, HTTP/S3 distribution.
m, err := model.LoadComputeManifestReader(reader, opts)
```

Failure dispatch is `errors.Is`-friendly:

```go
switch {
case errors.Is(err, model.ErrComputeManifestMissing):
    // manifest/CURRENT pointer absent or empty
case errors.Is(err, model.ErrComputeManifestParse):
    // YAML parse / read error
case errors.Is(err, model.ErrComputeManifestInvalid):
    // Validation rule 1-9 failure (per design doc §2.4)
}
```

## Who consumes the manifest

| Consumer | Where it consumes | Sequencing |
|---|---|---|
| **BMA reins (`bma compute-manifest current`)** | Operator-side CLI inspection; reins-wrapper hits `model.LoadComputeManifest` and surfaces phase + substrate identity to the beekeeper | Walk-α; stub-now-ready (`repo-bma-systema-issue-#177` per @bma-implementor seq=167) |
| **Federation CI mode-(b) gate** | `.github/workflows/ci-compute-manifest.yml` (Phase B-PR-8) loads the manifest on every `manifest/**` PR; gates mode-(b) promotion PRs against `IsModeBEligible` per Spec 9.2 §3.1 amendment | Walk-α (Phase B of `repo-bma-systema-issue-#171`) |
| **Translation Functor extraction harness** | `cmd/extract-cycle-counter-proof/` (Phase C-PR-13) Go-imports `qbp-compute-unit/emulator` at the SHA the manifest names; runs the emulator to validate the substrate-tier theorem at mode (b) | Phase C of `repo-bma-systema-issue-#170` |
| **Federation tenant runtime build-time pinning** | QBP-tenant scout daemon + Sharp Butler + Möbius + Materia each Go-import the substrate at the manifest's named SHA; tenant CI re-runs against the SHA on each manifest version-bump | Per `repo-bma-systema-issue-#176` (filed post-PR-#59) clarifies the `pinned_tag` semantics for cross-tenant build pinning |
| **Substrate-transition PR template** | Crawl → Walk M1 Gearbox transition PR (Walk-α opening) is a single-file amendment to `manifest/compute-manifest-v0_1.yaml` → `compute-manifest-v0_2.yaml` + `manifest/CURRENT` pointer update | Walk-α phase boundary |

## BMA-side reins-wrapper sketch

The BMA reins layer wraps `model.LoadComputeManifest` for operator-side inspection. The wrapper lives BMA-side at `cmd/bma/compute_manifest_cmd.go` (per `repo-bma-systema-issue-#177` scope; matches BMA's `cmd/bma/<feature>_cmd.go` reins convention per existing `cmd/bma/cart_cmd.go` + `cmd/bma/scope_cmd.go` + `cmd/bma/graph_cmd.go` precedents) and shapes the user-facing command:

```bash
$ bma compute-manifest current
phase:       crawl
substrate:
  name:      QBP-CU emulator
  kind:      emulator
  repo:      github.com/JamesPagetButler/qbp-compute-unit
  module:    emulator
  commit:    abc123def456abc123def456abc123def456abcd
  tag:       v0.1.0-rc1
authored_at: 2026-05-17T00:00:00Z
version:     v0.1
```

Behind the scenes:

```go
// cmd/bma/compute_manifest_cmd.go (BMA-side; not Wyrd-side)
import (
    "github.com/JamesPagetButler/wyrd/model"
)

func currentCmd(workspaceRoot string) error {
    m, err := model.LoadComputeManifest(workspaceRoot)
    if err != nil {
        switch {
        case errors.Is(err, model.ErrComputeManifestMissing):
            return fmt.Errorf("compute-manifest: not configured at %s (run `bma compute-manifest init` to bootstrap)", workspaceRoot)
        case errors.Is(err, model.ErrComputeManifestInvalid):
            return fmt.Errorf("compute-manifest: %w (operator action: review the manifest YAML)", err)
        }
        return fmt.Errorf("compute-manifest: %w", err)
    }
    return printManifest(os.Stdout, m)
}
```

The wrapper preserves Wyrd's typed-error contract through to operator output so error categories surface as actionable hints rather than opaque strings.

## Federation CI mode-(b) gate sketch

`.github/workflows/ci-compute-manifest.yml` (Phase B-PR-8 of `repo-bma-systema-issue-#171`) consumes the manifest at federation-PR-CI time:

```yaml
# .github/workflows/ci-compute-manifest.yml (Phase B-PR-8 — lands later)
name: ci-compute-manifest
on:
  pull_request:
    paths: ['manifest/**', 'model/compute_manifest.go']

jobs:
  validate-manifest:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version-file: 'go.mod' }
      - name: Validate Compute Manifest
        run: go test -run TestLoadComputeManifest_ActualManifestFile -v ./model/...

  mode-b-eligibility:
    if: contains(github.event.pull_request.labels.*.name, 'mode-b-promotion')
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version-file: 'go.mod' }
      - name: Check substrate credibility window
        run: |
          go run ./cmd/mode-b-eligibility-check --root . --window 24h
```

The Wyrd `model.LoadComputeManifest` + `IsModeBEligible(now, window)` predicate is the load-bearing contract. v0.1 ships without `IsModeBEligible` (credibility-window fields land in v0.2 per `repo-bma-systema-issue-#171` Phase B-PR-7). The CI workflow lands in Phase B-PR-8.

## Manifest version-bump pattern

When the federation transitions to a new substrate phase (Crawl → Walk, Walk → Run-initial, etc.), the manifest amendment is a single-PR change:

1. Add the new YAML: `manifest/compute-manifest-v0_2.yaml` (with the new phase + substrate.kind per Spec 9.2 §4)
2. Update the pointer: `manifest/CURRENT` from `compute-manifest-v0_1.yaml` to `compute-manifest-v0_2.yaml`
3. Don't delete the old YAML — keep as historical reference

Both files travel in one commit; consumers re-read `manifest/CURRENT` on demand, so version bumps don't break readers in flight (per §2.5b atomicity-via-Git-commit-boundary contract in PR #58 design doc).

## Soundness citations

Consumers of the Compute Manifest inherit:

- **`Wyrd.ComputeManifest.manifest_load_atomic`** — the loader returns either `validated` or `rejected` outcomes; no third state visible to consumers. Type-level contract that the Go runtime cannot return a non-nil `*ComputeManifest` alongside a non-nil error.
- **`Wyrd.ComputeManifest.load_deterministic`** — same input always produces same output. Useful for reproducibility audits (Notary discipline per `repo-bma-systema-issue-#175`+ co-verification targets).
- **`Wyrd.ComputeManifest.load_validated_iff_valid`** + **`load_rejected_iff_invalid`** — the validation Bool is decisive; consumers can dispatch on outcome variant without rechecking validity.

## Cross-references

- Design surface: `doc/design/compute-manifest.md` (PR #58, merged 2026-05-15 `953ccf2`)
- Schema + loader: `model/compute_manifest.go` (PR #59, merged 2026-05-18 `8c73c65`)
- Lean anchor: `lean/Wyrd/ComputeManifest.lean` (PR #60, Phase A impl-2)
- Spec 9.2 §3 + §4 + §11: `inter/spec/BMA-Spec-Addendum-9_2-Federation-Lean-Promotion-Protocol.md`
- Two-phase load pattern precedent: `doc/integration/contextus.md` + `store/scope_loader.go`
- BMA-side reins-wrapper tracking: `repo-bma-systema-issue-#177`
- Phase B (credibility-window extension): `repo-bma-systema-issue-#171`
- Phase C (Translation Functor cycle-counter cross-phase invariant): `repo-bma-systema-issue-#170`
- v0.2 housekeeping items: `repo-bma-systema-issue-#175` (Spec 9.2 §4 legacy-peer pattern), `repo-bma-systema-issue-#176` (pinned_tag git-tag-vs-module-version disambiguation)
