package model

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ComputeManifestPhase is the federation phase tag per Spec 9.2 §4.
type ComputeManifestPhase string

// Federation phase tags per the Spec 9.2 §4 table.
const (
	PhaseCrawl      ComputeManifestPhase = "crawl"
	PhaseToddle     ComputeManifestPhase = "toddle"
	PhaseWalk       ComputeManifestPhase = "walk"
	PhaseRunInitial ComputeManifestPhase = "run-initial"
	PhaseRunMature  ComputeManifestPhase = "run-mature"
)

// SubstrateKind classifies the blessed substrate's implementation form.
type SubstrateKind string

// SubstrateKind values per the Spec 9.2 §4 table.
const (
	SubstrateEmulator       SubstrateKind = "emulator"
	SubstrateGearboxCPU     SubstrateKind = "gearbox-cpu"
	SubstrateGPUAccelerator SubstrateKind = "gpu-accelerator"
	SubstrateSilicon        SubstrateKind = "silicon"
)

// Exported regex constants per design doc §2.4. SubstrateRepoRegex is
// exported as a var (not a const) so v0.2 host-broadening can swap the
// pattern without a schema bump (per @qbp-architecture concern 2 +
// @qbp-implementor F3 on PR #58).
var (
	VersionRegex            = regexp.MustCompile(`^v[0-9]+\.[0-9]+$`)
	SubstrateRepoRegex      = regexp.MustCompile(`^github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
	SubstrateCommitSHARegex = regexp.MustCompile(`^[0-9a-f]{40}$`)
)

// BootstrapSentinel is the literal commit_sha accepted only on the
// manifest-bootstrap PR's CI run (per design doc §2.5). v0.1 strict
// default REJECTS this value; callers opting in via LoadOptions.
// AllowBootstrapSentinel = true accept it. The federation CI workflow
// landed by repo-bma-systema-issue-#171 Phase B-PR-8 sets the opt-in
// only when running on the bootstrap branch.
const BootstrapSentinel = "TBD-pinned-at-PR-time"

// LegalPhaseKindPairs encodes the Spec 9.2 §4 phase table as the
// authoritative set of legal (phase, substrate.kind) pairings.
// Validation rule 8 rejects any pairing not present here.
//
// Per @qbp-architecture PR #58 pre-impl recommendation: lives as a
// package-level const-shaped var alongside the enums so the schema
// rule and the doc reference are grep-able from a single point.
var LegalPhaseKindPairs = map[ComputeManifestPhase]map[SubstrateKind]bool{
	PhaseCrawl:      {SubstrateEmulator: true},
	PhaseToddle:     {SubstrateEmulator: true},
	PhaseWalk:       {SubstrateGearboxCPU: true},
	PhaseRunInitial: {SubstrateGPUAccelerator: true, SubstrateGearboxCPU: true},
	PhaseRunMature:  {SubstrateSilicon: true, SubstrateGPUAccelerator: true, SubstrateGearboxCPU: true},
}

// Substrate names the federation's blessed compute substrate per
// Spec 9.2 §3 (the Compute-Substrate Gate). The Module field is
// REQUIRED when Kind == SubstrateEmulator (validation rule 9); for
// other kinds it is optional and typically empty (silicon / GPU
// substrates don't expose a Go module).
type Substrate struct {
	Name      string        `yaml:"name" json:"name"`
	Kind      SubstrateKind `yaml:"kind" json:"kind"`
	Repo      string        `yaml:"repo" json:"repo"`
	Module    string        `yaml:"module,omitempty" json:"module,omitempty"`
	CommitSHA string        `yaml:"commit_sha" json:"commit_sha"`
	PinnedTag string        `yaml:"pinned_tag,omitempty" json:"pinned_tag,omitempty"`
}

// ComputeManifest is the v0.1 typed snapshot of the Wyrd-owned
// Compute Manifest. Credibility-window fields (last_passing_tier_a,
// last_passing_tier_b) are deferred to v0.2 per
// repo-bma-systema-issue-#171 amendment.
//
// Forward-compatibility: unknown top-level YAML keys are silently
// ignored by the loader (gopkg.in/yaml.v3 default); a v0.2-shaped
// manifest is consumable by a v0.1 reader with v0.1-only semantics.
// Per @bma-implementor PR #58 non-blocking observation.
type ComputeManifest struct {
	Version    string               `yaml:"version" json:"version"`
	AuthoredAt time.Time            `yaml:"authored_at" json:"authored_at"`
	Phase      ComputeManifestPhase `yaml:"phase" json:"phase"`
	Substrate  Substrate            `yaml:"substrate" json:"substrate"`
}

// Sentinel errors per design doc §6. Consumers (BMA reins wrapper,
// federation CI mode-(b) gate) use errors.Is to dispatch on failure
// shape.
var (
	// ErrComputeManifestInvalid is returned when validation rules 1-9
	// (per design doc §2.4) fail.
	ErrComputeManifestInvalid = errors.New("model: compute-manifest validation failed")

	// ErrComputeManifestParse is returned for YAML parse errors or
	// I/O read errors on the underlying reader.
	ErrComputeManifestParse = errors.New("model: compute-manifest parse error")

	// ErrComputeManifestMissing is returned when manifest/CURRENT does
	// not exist or is empty.
	ErrComputeManifestMissing = errors.New("model: compute-manifest CURRENT pointer missing")
)

// LoadOptions controls strict-vs-bootstrap behavior of the loader.
type LoadOptions struct {
	// AllowBootstrapSentinel permits substrate.commit_sha to equal
	// BootstrapSentinel. Set true only by federation CI running on the
	// manifest-bootstrap PR (per design doc §2.5).
	AllowBootstrapSentinel bool
}

// LoadComputeManifest reads the manifest pointed to by
// manifest/CURRENT under root, validates strictly, and returns the
// typed snapshot. Returns:
//
//   - ErrComputeManifestMissing if manifest/CURRENT is absent or empty.
//   - ErrComputeManifestParse for YAML / I/O errors.
//   - ErrComputeManifestInvalid for validation failures.
func LoadComputeManifest(root string) (*ComputeManifest, error) {
	return LoadComputeManifestWithOptions(root, LoadOptions{})
}

// LoadComputeManifestWithOptions is the options-aware load form.
func LoadComputeManifestWithOptions(root string, opts LoadOptions) (*ComputeManifest, error) {
	currentPath := filepath.Join(root, "manifest", "CURRENT")
	// #nosec G304 G703 -- currentPath is constructed from a trusted root + the hard-coded "manifest/CURRENT" pair. File-inclusion-via-variable is the loader's documented contract.
	pointerBytes, err := os.ReadFile(currentPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", ErrComputeManifestMissing, currentPath)
		}
		return nil, fmt.Errorf("model: read %s: %w", currentPath, err)
	}
	pointed := strings.TrimSpace(string(pointerBytes))
	if pointed == "" {
		return nil, fmt.Errorf("%w: %s is empty", ErrComputeManifestMissing, currentPath)
	}
	if strings.ContainsAny(pointed, "/\\") {
		// The pointer is a basename within manifest/, not a path.
		// Reject path-shaped pointers to prevent escape from the
		// manifest directory.
		return nil, fmt.Errorf("%w: pointer %q must be a basename within manifest/, not a path", ErrComputeManifestInvalid, pointed)
	}
	manifestPath := filepath.Join(root, "manifest", pointed)
	// #nosec G304 G703 -- manifestPath is constructed from the trusted root + the validated pointer basename above (path-shape escape rejected by the basename check).
	f, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("model: open %s: %w", manifestPath, err)
	}
	defer func() { _ = f.Close() }()
	return LoadComputeManifestReader(f, opts)
}

// LoadComputeManifestReader is the io.Reader form of the loader.
// Useful for test fixtures, HTTP / S3 distribution, and any consumer
// that has the manifest bytes in hand without a filesystem path.
//
// Forward-compat: unknown top-level YAML keys are silently ignored
// (gopkg.in/yaml.v3 default behavior; design doc §2.4).
func LoadComputeManifestReader(r io.Reader, opts LoadOptions) (*ComputeManifest, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("%w: read: %w", ErrComputeManifestParse, err)
	}
	var m ComputeManifest
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("%w: yaml: %w", ErrComputeManifestParse, err)
	}
	if err := m.Validate(opts); err != nil {
		return nil, err
	}
	return &m, nil
}

// Validate enforces design doc §2.4 rules 1-9. Returns
// ErrComputeManifestInvalid (wrapped) on any failure.
func (m *ComputeManifest) Validate(opts LoadOptions) error {
	// Rule 1: version regex
	if !VersionRegex.MatchString(m.Version) {
		return fmt.Errorf("%w: rule 1: version %q does not match %s", ErrComputeManifestInvalid, m.Version, VersionRegex.String())
	}
	// Rule 2: authored_at parseable + not in the future
	if m.AuthoredAt.IsZero() {
		return fmt.Errorf("%w: rule 2: authored_at missing or unparseable", ErrComputeManifestInvalid)
	}
	if m.AuthoredAt.After(time.Now()) {
		return fmt.Errorf("%w: rule 2: authored_at %s is in the future", ErrComputeManifestInvalid, m.AuthoredAt.Format(time.RFC3339))
	}
	// Rule 3: phase enum membership (implicit via LegalPhaseKindPairs
	// key set; Rule 8 cross-check would also catch unknown phases, but
	// checking here gives a clearer error message attributing the
	// failure to Rule 3). Per @bma-implementor PR #59 non-blocking
	// observation.
	if _, ok := LegalPhaseKindPairs[m.Phase]; !ok {
		return fmt.Errorf("%w: rule 3: phase %q is not a known ComputeManifestPhase", ErrComputeManifestInvalid, m.Phase)
	}
	// Rule 4: substrate.name non-empty
	if m.Substrate.Name == "" {
		return fmt.Errorf("%w: rule 4: substrate.name is empty", ErrComputeManifestInvalid)
	}
	// Rule 5: substrate.kind enum
	if !isKnownSubstrateKind(m.Substrate.Kind) {
		return fmt.Errorf("%w: rule 5: substrate.kind %q is not a known SubstrateKind", ErrComputeManifestInvalid, m.Substrate.Kind)
	}
	// Rule 6: substrate.repo regex
	if !SubstrateRepoRegex.MatchString(m.Substrate.Repo) {
		return fmt.Errorf("%w: rule 6: substrate.repo %q does not match %s", ErrComputeManifestInvalid, m.Substrate.Repo, SubstrateRepoRegex.String())
	}
	// Rule 7: substrate.commit_sha (40 hex OR bootstrap sentinel when allowed)
	if m.Substrate.CommitSHA == BootstrapSentinel {
		if !opts.AllowBootstrapSentinel {
			return fmt.Errorf("%w: rule 7: substrate.commit_sha is bootstrap sentinel %q but LoadOptions.AllowBootstrapSentinel is false", ErrComputeManifestInvalid, BootstrapSentinel)
		}
	} else if !SubstrateCommitSHARegex.MatchString(m.Substrate.CommitSHA) {
		return fmt.Errorf("%w: rule 7: substrate.commit_sha %q is neither a 40-char hex nor the bootstrap sentinel", ErrComputeManifestInvalid, m.Substrate.CommitSHA)
	}
	// Rule 8: phase × kind cross-check
	legal := LegalPhaseKindPairs[m.Phase]
	if !legal[m.Substrate.Kind] {
		return fmt.Errorf("%w: rule 8: phase %q with substrate.kind %q is not a legal pairing per Spec 9.2 §4 table", ErrComputeManifestInvalid, m.Phase, m.Substrate.Kind)
	}
	// Rule 9: Substrate.Module conditional requirement
	if m.Substrate.Kind == SubstrateEmulator && m.Substrate.Module == "" {
		return fmt.Errorf("%w: rule 9: substrate.kind == %q requires non-empty substrate.module", ErrComputeManifestInvalid, SubstrateEmulator)
	}
	return nil
}

func isKnownSubstrateKind(k SubstrateKind) bool {
	switch k {
	case SubstrateEmulator, SubstrateGearboxCPU, SubstrateGPUAccelerator, SubstrateSilicon:
		return true
	}
	return false
}
