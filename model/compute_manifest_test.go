package model

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// happyManifest returns a manifest that passes every validation rule
// when LoadOptions.AllowBootstrapSentinel is true. Tests mutate copies
// of this base to exercise individual rule failures.
func happyManifest() ComputeManifest {
	return ComputeManifest{
		Version:    "v0.1",
		AuthoredAt: time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC),
		Phase:      PhaseCrawl,
		Substrate: Substrate{
			Name:      "QBP-CU emulator",
			Kind:      SubstrateEmulator,
			Repo:      "github.com/JamesPagetButler/qbp-compute-unit",
			Module:    "emulator",
			CommitSHA: BootstrapSentinel,
			PinnedTag: "v0.1.0",
		},
	}
}

func happyOpts() LoadOptions {
	return LoadOptions{AllowBootstrapSentinel: true}
}

func TestComputeManifest_RoundTrip(t *testing.T) {
	m := happyManifest()
	raw, err := yaml.Marshal(&m)
	if err != nil {
		t.Fatalf("yaml.Marshal: %v", err)
	}
	got, err := LoadComputeManifestReader(strings.NewReader(string(raw)), happyOpts())
	if err != nil {
		t.Fatalf("LoadComputeManifestReader on marshalled happyManifest: %v", err)
	}
	if got.Version != m.Version || got.Phase != m.Phase ||
		got.Substrate.Name != m.Substrate.Name ||
		got.Substrate.Kind != m.Substrate.Kind ||
		got.Substrate.Repo != m.Substrate.Repo ||
		got.Substrate.Module != m.Substrate.Module ||
		got.Substrate.CommitSHA != m.Substrate.CommitSHA ||
		got.Substrate.PinnedTag != m.Substrate.PinnedTag {
		t.Errorf("round-trip lost fields: got %+v, want %+v", got, m)
	}
	if !got.AuthoredAt.Equal(m.AuthoredAt) {
		t.Errorf("round-trip AuthoredAt: got %v, want %v", got.AuthoredAt, m.AuthoredAt)
	}
}

func TestValidate_AllRulesGreen(t *testing.T) {
	m := happyManifest()
	if err := m.Validate(happyOpts()); err != nil {
		t.Fatalf("happy manifest failed validation: %v", err)
	}
}

func TestValidate_Rule1_BadVersion(t *testing.T) {
	cases := []string{"", "1.0", "v0", "v0.1.0", "v0_1", "v0.x"}
	for _, v := range cases {
		t.Run(v, func(t *testing.T) {
			m := happyManifest()
			m.Version = v
			err := m.Validate(happyOpts())
			if err == nil {
				t.Errorf("expected rule-1 failure for version %q, got nil", v)
			}
			if !errors.Is(err, ErrComputeManifestInvalid) {
				t.Errorf("expected ErrComputeManifestInvalid, got %v", err)
			}
			if !strings.Contains(err.Error(), "rule 1") {
				t.Errorf("expected rule-1 attribution in error %q", err.Error())
			}
		})
	}
}

func TestValidate_Rule2_ZeroAuthoredAt(t *testing.T) {
	m := happyManifest()
	m.AuthoredAt = time.Time{}
	err := m.Validate(happyOpts())
	if !errors.Is(err, ErrComputeManifestInvalid) {
		t.Fatalf("expected ErrComputeManifestInvalid, got %v", err)
	}
	if !strings.Contains(err.Error(), "rule 2") {
		t.Errorf("expected rule-2 attribution, got %q", err.Error())
	}
}

func TestValidate_Rule2_FutureAuthoredAt(t *testing.T) {
	m := happyManifest()
	m.AuthoredAt = time.Now().Add(24 * time.Hour)
	err := m.Validate(happyOpts())
	if !errors.Is(err, ErrComputeManifestInvalid) {
		t.Fatalf("expected ErrComputeManifestInvalid, got %v", err)
	}
	if !strings.Contains(err.Error(), "rule 2") {
		t.Errorf("expected rule-2 attribution, got %q", err.Error())
	}
}

func TestValidate_Rule3_UnknownPhase(t *testing.T) {
	m := happyManifest()
	m.Phase = ComputeManifestPhase("post-runs")
	err := m.Validate(happyOpts())
	if !errors.Is(err, ErrComputeManifestInvalid) {
		t.Fatalf("expected ErrComputeManifestInvalid, got %v", err)
	}
	if !strings.Contains(err.Error(), "rule 3") {
		t.Errorf("expected rule-3 attribution, got %q", err.Error())
	}
}

func TestValidate_Rule4_EmptyName(t *testing.T) {
	m := happyManifest()
	m.Substrate.Name = ""
	err := m.Validate(happyOpts())
	if !errors.Is(err, ErrComputeManifestInvalid) {
		t.Fatalf("expected ErrComputeManifestInvalid, got %v", err)
	}
	if !strings.Contains(err.Error(), "rule 4") {
		t.Errorf("expected rule-4 attribution, got %q", err.Error())
	}
}

func TestValidate_Rule5_UnknownKind(t *testing.T) {
	m := happyManifest()
	m.Substrate.Kind = SubstrateKind("quantum")
	err := m.Validate(happyOpts())
	if !errors.Is(err, ErrComputeManifestInvalid) {
		t.Fatalf("expected ErrComputeManifestInvalid, got %v", err)
	}
	if !strings.Contains(err.Error(), "rule 5") {
		t.Errorf("expected rule-5 attribution, got %q", err.Error())
	}
}

func TestValidate_Rule6_NonGitHubRepo(t *testing.T) {
	cases := []string{
		"gitlab.com/foo/bar",
		"gitea.example.com/foo/bar",
		"foo/bar",
		"",
	}
	for _, r := range cases {
		t.Run(r, func(t *testing.T) {
			m := happyManifest()
			m.Substrate.Repo = r
			err := m.Validate(happyOpts())
			if !errors.Is(err, ErrComputeManifestInvalid) {
				t.Fatalf("expected ErrComputeManifestInvalid, got %v", err)
			}
			if !strings.Contains(err.Error(), "rule 6") {
				t.Errorf("expected rule-6 attribution, got %q", err.Error())
			}
		})
	}
}

func TestValidate_Rule7_BadSHA(t *testing.T) {
	cases := []string{
		"ABC123",                                     // too short
		"XYZNOT-A-HEX-AT-ALL",                        // non-hex
		"abc123def456abc123def456abc123def456",       // 36 chars (too short)
		"abc123def456abc123def456abc123def456abc12X", // 41 chars + non-hex
		"abc123def456abc123def456abc123def456ABCD",   // uppercase A-F (regex requires [0-9a-f])
	}
	for _, sha := range cases {
		t.Run(sha, func(t *testing.T) {
			m := happyManifest()
			m.Substrate.CommitSHA = sha
			err := m.Validate(happyOpts())
			if !errors.Is(err, ErrComputeManifestInvalid) {
				t.Fatalf("expected ErrComputeManifestInvalid, got %v", err)
			}
			if !strings.Contains(err.Error(), "rule 7") {
				t.Errorf("expected rule-7 attribution, got %q", err.Error())
			}
		})
	}
}

func TestValidate_Rule7_BootstrapSentinel_Strict(t *testing.T) {
	m := happyManifest()
	// happyManifest already uses the sentinel; strict opts must reject.
	err := m.Validate(LoadOptions{AllowBootstrapSentinel: false})
	if !errors.Is(err, ErrComputeManifestInvalid) {
		t.Fatalf("strict default must reject bootstrap sentinel; got %v", err)
	}
	if !strings.Contains(err.Error(), "rule 7") {
		t.Errorf("expected rule-7 attribution, got %q", err.Error())
	}
}

func TestValidate_Rule7_BootstrapSentinel_Allowed(t *testing.T) {
	m := happyManifest()
	if err := m.Validate(LoadOptions{AllowBootstrapSentinel: true}); err != nil {
		t.Fatalf("AllowBootstrapSentinel=true must accept sentinel; got %v", err)
	}
}

func TestValidate_Rule7_RealSHA(t *testing.T) {
	m := happyManifest()
	m.Substrate.CommitSHA = "abc123def456abc123def456abc123def456abcd"
	// 40-char hex. Should pass even with strict opts (sentinel-irrelevant).
	if err := m.Validate(LoadOptions{AllowBootstrapSentinel: false}); err != nil {
		t.Fatalf("real 40-hex SHA must pass strict validation; got %v", err)
	}
}

func TestValidate_Rule8_PhaseKindMatrix(t *testing.T) {
	// Every legal pair per LegalPhaseKindPairs must pass.
	for phase, kinds := range LegalPhaseKindPairs {
		for kind := range kinds {
			t.Run(string(phase)+"+"+string(kind), func(t *testing.T) {
				m := happyManifest()
				m.Phase = phase
				m.Substrate.Kind = kind
				// Module conditional (rule 9): only required for emulator.
				if kind != SubstrateEmulator {
					m.Substrate.Module = ""
				}
				if err := m.Validate(happyOpts()); err != nil {
					t.Errorf("legal pair (%s, %s) failed: %v", phase, kind, err)
				}
			})
		}
	}

	// Sample illegal pairs must fail at rule 8.
	illegal := []struct {
		phase ComputeManifestPhase
		kind  SubstrateKind
	}{
		{PhaseCrawl, SubstrateSilicon},
		{PhaseCrawl, SubstrateGPUAccelerator},
		{PhaseToddle, SubstrateGearboxCPU},
		{PhaseWalk, SubstrateEmulator},
		{PhaseRunInitial, SubstrateEmulator},
		{PhaseRunMature, SubstrateEmulator},
	}
	for _, c := range illegal {
		t.Run("illegal/"+string(c.phase)+"+"+string(c.kind), func(t *testing.T) {
			m := happyManifest()
			m.Phase = c.phase
			m.Substrate.Kind = c.kind
			err := m.Validate(happyOpts())
			if !errors.Is(err, ErrComputeManifestInvalid) {
				t.Fatalf("expected ErrComputeManifestInvalid, got %v", err)
			}
			if !strings.Contains(err.Error(), "rule 8") {
				t.Errorf("expected rule-8 attribution, got %q", err.Error())
			}
		})
	}
}

func TestValidate_Rule9_EmulatorRequiresModule(t *testing.T) {
	m := happyManifest()
	m.Substrate.Module = ""
	err := m.Validate(happyOpts())
	if !errors.Is(err, ErrComputeManifestInvalid) {
		t.Fatalf("expected ErrComputeManifestInvalid, got %v", err)
	}
	if !strings.Contains(err.Error(), "rule 9") {
		t.Errorf("expected rule-9 attribution, got %q", err.Error())
	}
}

func TestValidate_Rule9_NonEmulatorModuleOptional(t *testing.T) {
	// gearbox-cpu, gpu-accelerator, silicon all permit empty module
	// — but only when paired with a legal phase per rule 8.
	cases := []struct {
		phase ComputeManifestPhase
		kind  SubstrateKind
	}{
		{PhaseWalk, SubstrateGearboxCPU},
		{PhaseRunInitial, SubstrateGPUAccelerator},
		{PhaseRunMature, SubstrateSilicon},
	}
	for _, c := range cases {
		t.Run(string(c.kind), func(t *testing.T) {
			m := happyManifest()
			m.Phase = c.phase
			m.Substrate.Kind = c.kind
			m.Substrate.Module = ""
			if err := m.Validate(happyOpts()); err != nil {
				t.Errorf("non-emulator with empty module should pass; got %v", err)
			}
		})
	}
}

// fixture creates a temporary manifest/ tree under t.TempDir() and
// returns the root.
func fixtureRoot(t *testing.T, current string, yamlBody string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "manifest"), 0o750); err != nil {
		t.Fatalf("mkdir manifest: %v", err)
	}
	if current != "" {
		if err := os.WriteFile(filepath.Join(root, "manifest", "CURRENT"), []byte(current), 0o600); err != nil {
			t.Fatalf("write CURRENT: %v", err)
		}
	}
	if yamlBody != "" {
		// Use whatever basename the CURRENT pointer points at; default
		// to "compute-manifest-v0_1.yaml" if current is non-trivial.
		basename := strings.TrimSpace(current)
		if basename == "" {
			basename = "compute-manifest-v0_1.yaml"
		}
		if err := os.WriteFile(filepath.Join(root, "manifest", basename), []byte(yamlBody), 0o600); err != nil {
			t.Fatalf("write %s: %v", basename, err)
		}
	}
	return root
}

const happyYAML = `version: "v0.1"
authored_at: "2026-05-17T00:00:00Z"
phase: "crawl"
substrate:
  name: "QBP-CU emulator"
  kind: "emulator"
  repo: "github.com/JamesPagetButler/qbp-compute-unit"
  module: "emulator"
  commit_sha: "TBD-pinned-at-PR-time"
  pinned_tag: "v0.1.0"
`

func TestLoadComputeManifest_ReadsCurrentPointer(t *testing.T) {
	root := fixtureRoot(t, "compute-manifest-v0_1.yaml\n", happyYAML)
	m, err := LoadComputeManifestWithOptions(root, happyOpts())
	if err != nil {
		t.Fatalf("LoadComputeManifest: %v", err)
	}
	if m.Phase != PhaseCrawl {
		t.Errorf("phase: got %s, want %s", m.Phase, PhaseCrawl)
	}
	if m.Substrate.Name != "QBP-CU emulator" {
		t.Errorf("substrate.name: got %q, want %q", m.Substrate.Name, "QBP-CU emulator")
	}
}

func TestLoadComputeManifest_StrictDefault_RejectsSentinel(t *testing.T) {
	root := fixtureRoot(t, "compute-manifest-v0_1.yaml\n", happyYAML)
	_, err := LoadComputeManifest(root)
	if !errors.Is(err, ErrComputeManifestInvalid) {
		t.Fatalf("strict LoadComputeManifest must reject sentinel; got %v", err)
	}
}

func TestLoadComputeManifest_MissingCurrent(t *testing.T) {
	root := fixtureRoot(t, "", "")
	_, err := LoadComputeManifestWithOptions(root, happyOpts())
	if !errors.Is(err, ErrComputeManifestMissing) {
		t.Fatalf("expected ErrComputeManifestMissing, got %v", err)
	}
}

func TestLoadComputeManifest_EmptyCurrent(t *testing.T) {
	root := fixtureRoot(t, "   \n", "")
	_, err := LoadComputeManifestWithOptions(root, happyOpts())
	if !errors.Is(err, ErrComputeManifestMissing) {
		t.Fatalf("expected ErrComputeManifestMissing on whitespace-only CURRENT, got %v", err)
	}
}

func TestLoadComputeManifest_PathShapedPointer(t *testing.T) {
	root := fixtureRoot(t, "../escape.yaml\n", "")
	_, err := LoadComputeManifestWithOptions(root, happyOpts())
	if !errors.Is(err, ErrComputeManifestInvalid) {
		t.Fatalf("path-shaped pointer must be rejected; got %v", err)
	}
}

func TestLoadComputeManifest_PointerToMissingFile(t *testing.T) {
	root := fixtureRoot(t, "compute-manifest-v0_99.yaml\n", "")
	_, err := LoadComputeManifestWithOptions(root, happyOpts())
	if err == nil {
		t.Fatal("expected error on missing pointed-to file, got nil")
	}
	if errors.Is(err, ErrComputeManifestInvalid) || errors.Is(err, ErrComputeManifestParse) || errors.Is(err, ErrComputeManifestMissing) {
		t.Errorf("expected raw fs error (not a sentinel), got %v", err)
	}
}

func TestLoadComputeManifest_UnknownTopLevelKeys(t *testing.T) {
	// Per @bma-implementor PR #58 observation: extra top-level keys
	// silently ignored (additive forward-compat).
	yamlWithExtras := happyYAML + `
credibility:
  last_passing_tier_a:
    timestamp: "2026-05-17T00:00:00Z"
    substrate_commit_sha: "abc123def456abc123def456abc123def456abcd"
verified_invariants:
  - theorem: "Wyrd.CycleCounterCrossPhase.cycle_counter_monotonic_per_phase"
unknown_v0_99_field: "ignored"
`
	root := fixtureRoot(t, "compute-manifest-v0_1.yaml\n", yamlWithExtras)
	m, err := LoadComputeManifestWithOptions(root, happyOpts())
	if err != nil {
		t.Fatalf("unknown top-level keys should be silently ignored; got %v", err)
	}
	if m.Phase != PhaseCrawl {
		t.Errorf("phase from forward-compat fixture: got %s, want %s", m.Phase, PhaseCrawl)
	}
}

func TestLoadComputeManifest_ParseError(t *testing.T) {
	root := fixtureRoot(t, "compute-manifest-v0_1.yaml\n", "not-yaml:\n  - [unclosed")
	_, err := LoadComputeManifestWithOptions(root, happyOpts())
	if !errors.Is(err, ErrComputeManifestParse) {
		t.Fatalf("expected ErrComputeManifestParse, got %v", err)
	}
}

func TestLoadComputeManifest_ActualManifestFile(t *testing.T) {
	// Load the actual manifest/compute-manifest-v0_1.yaml shipped in
	// this PR (relative to the package root). This is the bootstrap
	// PR; AllowBootstrapSentinel must be true.
	//
	// The test package runs from model/; the manifest is two levels
	// up... well, just one level up: ../manifest/.
	root := ".."
	m, err := LoadComputeManifestWithOptions(root, happyOpts())
	if err != nil {
		t.Fatalf("loading the actual shipped manifest must pass with AllowBootstrapSentinel=true; got %v", err)
	}
	if m.Phase != PhaseCrawl {
		t.Errorf("actual manifest phase: got %s, want %s", m.Phase, PhaseCrawl)
	}
	if m.Substrate.Kind != SubstrateEmulator {
		t.Errorf("actual manifest substrate.kind: got %s, want %s", m.Substrate.Kind, SubstrateEmulator)
	}
	if m.Substrate.CommitSHA != BootstrapSentinel {
		t.Errorf("actual manifest substrate.commit_sha: got %q, want %q (bootstrap PR)", m.Substrate.CommitSHA, BootstrapSentinel)
	}
}

func TestLoadComputeManifest_ActualManifestFile_StrictRejectsSentinel(t *testing.T) {
	// Same actual file, but strict opts. Documents that the federation
	// CI workflow (Phase B-PR-8) must set AllowBootstrapSentinel=true
	// on the bootstrap branch, OR the impl-1 PR must coordinate with
	// qbp-cu-implementor to pin a real SHA.
	root := ".."
	_, err := LoadComputeManifest(root)
	if !errors.Is(err, ErrComputeManifestInvalid) {
		t.Fatalf("strict load of bootstrap manifest must fail with rule-7; got %v", err)
	}
}
