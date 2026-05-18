package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const happyV02YAML = `version: "v0.2"
authored_at: "2026-05-18T00:00:00Z"
phase: "crawl"
substrate:
  name: "QBP-CU emulator"
  kind: "emulator"
  repo: "github.com/JamesPagetButler/qbp-compute-unit"
  module: "emulator"
  commit_sha: "TBD-pinned-at-PR-time"
  pinned_tag: "v0.1.0-rc1"
`

// fixtureRoot writes a minimal manifest tree under t.TempDir() and
// returns the root.
func fixtureRoot(t *testing.T, yamlBody string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "manifest"), 0o750); err != nil {
		t.Fatalf("mkdir manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "manifest", "CURRENT"), []byte("compute-manifest-v0_2.yaml\n"), 0o600); err != nil {
		t.Fatalf("write CURRENT: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "manifest", "compute-manifest-v0_2.yaml"), []byte(yamlBody), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return root
}

func TestRun_CrawlBootstrapSentinel_BestEffortEligible(t *testing.T) {
	root := fixtureRoot(t, happyV02YAML)
	out, err := os.CreateTemp("", "mode-b-out-*")
	if err != nil {
		t.Fatalf("temp out: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(out.Name()) })

	if err := run(root, 72*time.Hour, true, out); err != nil {
		t.Fatalf("Crawl + bootstrap sentinel + allowBootstrap=true should be eligible; got %v", err)
	}

	_ = out.Close()
	body, _ := os.ReadFile(out.Name())
	s := string(body)
	if !strings.Contains(s, "mode_b_eligible=true") {
		t.Errorf("expected mode_b_eligible=true; got %q", s)
	}
	if !strings.Contains(s, "mode_b_warning=mode-b-best-effort") {
		t.Errorf("expected mode-b-best-effort warning; got %q", s)
	}
	if !strings.Contains(s, "mode_b_phase=crawl") {
		t.Errorf("expected mode_b_phase=crawl; got %q", s)
	}
	if !strings.Contains(s, "best-effort: substrate.commit_sha is bootstrap sentinel") {
		t.Errorf("expected bootstrap-sentinel reason; got %q", s)
	}
}

func TestRun_CrawlBootstrapSentinel_StrictLoadFails(t *testing.T) {
	// allowBootstrap=false → loader rejects the sentinel at rule 7.
	// Returns exit-code-2-class error (manifest load failure).
	root := fixtureRoot(t, happyV02YAML)
	out, err := os.CreateTemp("", "mode-b-out-*")
	if err != nil {
		t.Fatalf("temp out: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(out.Name()) })

	err = run(root, 72*time.Hour, false, out)
	if err == nil {
		t.Fatal("strict load of bootstrap-sentinel manifest should fail; got nil error")
	}
	if errors.Is(err, errIneligible) {
		t.Errorf("expected load error, not ineligible-sentinel; got %v", err)
	}
}

func TestRun_MissingCurrent_Errors(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "manifest"), 0o750); err != nil {
		t.Fatalf("mkdir manifest: %v", err)
	}
	// No CURRENT pointer; loader returns ErrComputeManifestMissing.
	out, err := os.CreateTemp("", "mode-b-out-*")
	if err != nil {
		t.Fatalf("temp out: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(out.Name()) })

	err = run(root, 72*time.Hour, true, out)
	if err == nil {
		t.Fatal("missing CURRENT should error")
	}
	if errors.Is(err, errIneligible) {
		t.Errorf("expected load error, not ineligible-sentinel; got %v", err)
	}
}

// TestIsBestEffortReason exercises the reason-prefix detector that
// drives the mode-b-best-effort warning annotation.
func TestIsBestEffortReason(t *testing.T) {
	cases := []struct {
		reason string
		want   bool
	}{
		{"best-effort: substrate.commit_sha is bootstrap sentinel", true},
		{"best-effort: credibility block absent (Crawl/Toddle phase)", true},
		{"ok", false},
		{"Tier B is 48h0m0s old; exceeds window 24h0m0s (strict phase)", false},
		{"", false},
	}
	for _, c := range cases {
		t.Run(c.reason, func(t *testing.T) {
			if got := isBestEffortReason(c.reason); got != c.want {
				t.Errorf("isBestEffortReason(%q) = %v, want %v", c.reason, got, c.want)
			}
		})
	}
}
