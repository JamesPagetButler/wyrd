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
	// Load the actual manifest/compute-manifest-v0_2.yaml shipped on
	// main (per manifest/CURRENT). As of repo-wyrd-pr-#69 (federation's
	// first substrate-tier promotion), the bootstrap sentinel was
	// retired and a real 40-char hex SHA pinned for the
	// qbp-compute-unit/emulator@v0.1.0-rc1 commit. Strict-default load
	// MUST succeed.
	root := ".."
	m, err := LoadComputeManifest(root)
	if err != nil {
		t.Fatalf("loading the actual shipped manifest must pass with strict-default; got %v", err)
	}
	if m.Phase != PhaseCrawl {
		t.Errorf("actual manifest phase: got %s, want %s", m.Phase, PhaseCrawl)
	}
	if m.Substrate.Kind != SubstrateEmulator {
		t.Errorf("actual manifest substrate.kind: got %s, want %s", m.Substrate.Kind, SubstrateEmulator)
	}
	if m.Substrate.CommitSHA == BootstrapSentinel {
		t.Errorf("actual manifest substrate.commit_sha is still bootstrap sentinel; expected a real 40-char hex SHA (sentinel retired at PR #69)")
	}
	if !SubstrateCommitSHARegex.MatchString(m.Substrate.CommitSHA) {
		t.Errorf("actual manifest substrate.commit_sha %q is not a valid 40-char hex SHA", m.Substrate.CommitSHA)
	}
}

func TestLoadComputeManifest_ActualManifestFile_AllowBootstrapNoLongerNeeded(t *testing.T) {
	// Companion to TestLoadComputeManifest_ActualManifestFile —
	// documents that AllowBootstrapSentinel is no longer required for
	// the shipped manifest (sentinel retired at PR #69). Both strict-
	// default and opt-in loads succeed.
	root := ".."
	if _, err := LoadComputeManifest(root); err != nil {
		t.Errorf("strict-default load should succeed (sentinel retired): %v", err)
	}
	if _, err := LoadComputeManifestWithOptions(root, happyOpts()); err != nil {
		t.Errorf("opt-in load should also succeed: %v", err)
	}
}

// ============================================================
// v0.2 credibility-window tests (per Spec 9.2 §3.1 amendment;
// repo-bma-systema-issue-#171 Phase B-PR-7)
// ============================================================

// realSHA is a valid 40-char hex string for tests that need a real
// substrate.commit_sha (not the bootstrap sentinel).
const realSHA = "abc123def456abc123def456abc123def456abcd"

// happyManifestWithCredibility returns a manifest with the bootstrap
// sentinel replaced by a real SHA and populated credibility fields.
// Tests mutate copies of this base to exercise the IsModeBEligible
// truth-table.
func happyManifestWithCredibility(now time.Time) ComputeManifest {
	m := happyManifest()
	m.Version = "v0.2"
	m.Substrate.CommitSHA = realSHA
	m.Credibility = &Credibility{
		LastPassingTierA: &TierVerification{
			Timestamp:          now.Add(-1 * time.Hour),
			SubstrateCommitSHA: realSHA,
		},
		LastPassingTierB: &TierVerification{
			Timestamp:          now.Add(-2 * time.Hour),
			SubstrateCommitSHA: realSHA,
		},
	}
	return m
}

func TestIsBestEffortPhase(t *testing.T) {
	cases := []struct {
		phase ComputeManifestPhase
		want  bool
	}{
		{PhaseCrawl, true},
		{PhaseToddle, true},
		{PhaseWalk, false},
		{PhaseRunInitial, false},
		{PhaseRunMature, false},
	}
	for _, c := range cases {
		t.Run(string(c.phase), func(t *testing.T) {
			if got := c.phase.IsBestEffortPhase(); got != c.want {
				t.Errorf("phase %q: got %v, want %v", c.phase, got, c.want)
			}
		})
	}
}

func TestIsModeBEligible_BootstrapSentinel_Always_BestEffort(t *testing.T) {
	// When substrate.commit_sha is the bootstrap sentinel, mode (b)
	// eligibility is best-effort regardless of phase (no real
	// substrate to verify yet).
	for _, phase := range []ComputeManifestPhase{PhaseCrawl, PhaseToddle, PhaseWalk, PhaseRunInitial, PhaseRunMature} {
		t.Run(string(phase), func(t *testing.T) {
			m := happyManifest() // uses BootstrapSentinel
			m.Phase = phase
			if phase != PhaseCrawl && phase != PhaseToddle {
				// Rule 8 requires kind change; switch kind to legal for phase
				for k := range LegalPhaseKindPairs[phase] {
					m.Substrate.Kind = k
					if k != SubstrateEmulator {
						m.Substrate.Module = ""
					}
					break
				}
			}
			eligible, reason := m.IsModeBEligible(time.Now(), 24*time.Hour)
			if !eligible {
				t.Errorf("bootstrap-sentinel substrate should be best-effort eligible; got (false, %q)", reason)
			}
			if !strings.Contains(reason, "bootstrap sentinel") {
				t.Errorf("expected reason to cite bootstrap sentinel; got %q", reason)
			}
		})
	}
}

func TestIsModeBEligible_Crawl_BestEffort_AbsentCredibility(t *testing.T) {
	m := happyManifest()
	m.Substrate.CommitSHA = realSHA // real SHA but no credibility block
	eligible, reason := m.IsModeBEligible(time.Now(), 24*time.Hour)
	if !eligible {
		t.Fatalf("Crawl + real SHA + nil credibility should be best-effort eligible; got (false, %q)", reason)
	}
	if !strings.Contains(reason, "best-effort") {
		t.Errorf("expected best-effort reason; got %q", reason)
	}
}

func TestIsModeBEligible_Crawl_BestEffort_StaleTierB(t *testing.T) {
	now := time.Now()
	m := happyManifestWithCredibility(now)
	// Push Tier B way out of window
	m.Credibility.LastPassingTierB.Timestamp = now.Add(-48 * time.Hour)
	eligible, reason := m.IsModeBEligible(now, 24*time.Hour)
	if !eligible {
		t.Fatalf("Crawl phase should tolerate stale Tier B as best-effort; got (false, %q)", reason)
	}
	if !strings.Contains(reason, "best-effort") {
		t.Errorf("expected best-effort reason; got %q", reason)
	}
}

func TestIsModeBEligible_Walk_Strict_FreshTierB_Pass(t *testing.T) {
	now := time.Now()
	m := happyManifestWithCredibility(now)
	m.Phase = PhaseWalk
	m.Substrate.Kind = SubstrateGearboxCPU
	m.Substrate.Module = "" // gearbox-cpu doesn't require module
	eligible, reason := m.IsModeBEligible(now, 24*time.Hour)
	if !eligible {
		t.Fatalf("Walk + fresh Tier B (SHA match, within window) should pass strict; got (false, %q)", reason)
	}
	if reason != "ok" {
		t.Errorf("expected reason 'ok' for strict pass; got %q", reason)
	}
}

func TestIsModeBEligible_Walk_Strict_StaleTierB_Block(t *testing.T) {
	now := time.Now()
	m := happyManifestWithCredibility(now)
	m.Phase = PhaseWalk
	m.Substrate.Kind = SubstrateGearboxCPU
	m.Substrate.Module = ""
	// Push Tier B way out of 24h window
	m.Credibility.LastPassingTierB.Timestamp = now.Add(-48 * time.Hour)
	eligible, reason := m.IsModeBEligible(now, 24*time.Hour)
	if eligible {
		t.Fatalf("Walk + stale Tier B (>24h) should BLOCK; got (true, %q)", reason)
	}
	if !strings.Contains(reason, "exceeds window") {
		t.Errorf("expected reason to cite window exceedance; got %q", reason)
	}
}

func TestIsModeBEligible_Walk_Strict_TierASHAMismatch_Block(t *testing.T) {
	now := time.Now()
	m := happyManifestWithCredibility(now)
	m.Phase = PhaseWalk
	m.Substrate.Kind = SubstrateGearboxCPU
	m.Substrate.Module = ""
	// Substrate code changed since last Tier A
	m.Substrate.CommitSHA = "0000000000000000000000000000000000000000"
	// (LastPassingTierA still references realSHA)
	eligible, reason := m.IsModeBEligible(now, 24*time.Hour)
	if eligible {
		t.Fatalf("Walk + Tier A SHA mismatch should BLOCK; got (true, %q)", reason)
	}
	if !strings.Contains(reason, "Tier A SHA mismatch") {
		t.Errorf("expected reason to cite Tier A SHA mismatch; got %q", reason)
	}
}

func TestIsModeBEligible_Walk_Strict_NilCredibility_Block(t *testing.T) {
	m := happyManifest()
	m.Phase = PhaseWalk
	m.Substrate.Kind = SubstrateGearboxCPU
	m.Substrate.Module = ""
	m.Substrate.CommitSHA = realSHA
	m.Credibility = nil
	eligible, reason := m.IsModeBEligible(time.Now(), 24*time.Hour)
	if eligible {
		t.Fatalf("Walk + nil credibility should BLOCK strict; got (true, %q)", reason)
	}
	if !strings.Contains(reason, "strict") {
		t.Errorf("expected reason to cite strict; got %q", reason)
	}
}

func TestIsModeBEligible_Walk_Strict_NilTierA_Block(t *testing.T) {
	now := time.Now()
	m := happyManifestWithCredibility(now)
	m.Phase = PhaseWalk
	m.Substrate.Kind = SubstrateGearboxCPU
	m.Substrate.Module = ""
	m.Credibility.LastPassingTierA = nil
	eligible, reason := m.IsModeBEligible(now, 24*time.Hour)
	if eligible {
		t.Fatalf("Walk + nil Tier A should BLOCK strict; got (true, %q)", reason)
	}
	if !strings.Contains(reason, "last_passing_tier_a absent") {
		t.Errorf("expected reason to cite Tier A absence; got %q", reason)
	}
}

func TestIsModeBEligible_Walk_Strict_FutureTimestamp_Block(t *testing.T) {
	now := time.Now()
	m := happyManifestWithCredibility(now)
	m.Phase = PhaseWalk
	m.Substrate.Kind = SubstrateGearboxCPU
	m.Substrate.Module = ""
	// Tier B timestamp in the future (clock skew or paperwork bug)
	m.Credibility.LastPassingTierB.Timestamp = now.Add(1 * time.Hour)
	eligible, reason := m.IsModeBEligible(now, 24*time.Hour)
	if eligible {
		t.Fatalf("Walk + future Tier B should BLOCK strict; got (true, %q)", reason)
	}
	if !strings.Contains(reason, "future") {
		t.Errorf("expected reason to cite future timestamp; got %q", reason)
	}
}

func TestCredibility_RoundTrip_v02(t *testing.T) {
	// v0.2-shaped manifest with credibility round-trips through YAML.
	now := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)
	m := happyManifestWithCredibility(now)
	raw, err := yaml.Marshal(&m)
	if err != nil {
		t.Fatalf("yaml.Marshal v0.2 manifest: %v", err)
	}
	got, err := LoadComputeManifestReader(strings.NewReader(string(raw)), LoadOptions{})
	if err != nil {
		t.Fatalf("LoadComputeManifestReader on v0.2 round-trip: %v", err)
	}
	if got.Credibility == nil {
		t.Fatal("Credibility block lost in round-trip")
	}
	if got.Credibility.LastPassingTierA == nil || got.Credibility.LastPassingTierB == nil {
		t.Fatal("Tier A/B fields lost in round-trip")
	}
	if got.Credibility.LastPassingTierA.SubstrateCommitSHA != realSHA {
		t.Errorf("Tier A SHA: got %q, want %q", got.Credibility.LastPassingTierA.SubstrateCommitSHA, realSHA)
	}
	if !got.Credibility.LastPassingTierA.Timestamp.Equal(m.Credibility.LastPassingTierA.Timestamp) {
		t.Errorf("Tier A timestamp lost precision: got %v, want %v", got.Credibility.LastPassingTierA.Timestamp, m.Credibility.LastPassingTierA.Timestamp)
	}
}

func TestCredibility_v01_ManifestStill_Loads(t *testing.T) {
	// v0.1-shaped YAML (no credibility block) still loads — Credibility
	// pointer is nil. Backward-compat preserved.
	v01YAML := `version: "v0.1"
authored_at: "2026-05-17T00:00:00Z"
phase: "crawl"
substrate:
  name: "QBP-CU emulator"
  kind: "emulator"
  repo: "github.com/JamesPagetButler/qbp-compute-unit"
  module: "emulator"
  commit_sha: "TBD-pinned-at-PR-time"
  pinned_tag: "v0.1.0-rc1"
`
	m, err := LoadComputeManifestReader(strings.NewReader(v01YAML), happyOpts())
	if err != nil {
		t.Fatalf("v0.1 YAML should still load: %v", err)
	}
	if m.Credibility != nil {
		t.Errorf("v0.1 YAML should produce nil Credibility; got %+v", m.Credibility)
	}
}
