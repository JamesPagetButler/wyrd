// Command mode-b-eligibility-check is the Wyrd-side CLI consumed by
// the federation CI mode-(b) gate workflow per Spec 9.2 §3.1
// amendment (substrate-credibility-window for mode (b) extraction-
// and-execute promotion).
//
// Usage:
//
//	mode-b-eligibility-check --root <wyrd-root> [--window <duration>]
//	mode-b-eligibility-check --root . --window 72h
//
// Behavior:
//
//   - Loads the Compute Manifest from <wyrd-root>/manifest/CURRENT
//     via model.LoadComputeManifestWithOptions (bootstrap sentinel
//     opt-in honored only when --allow-bootstrap-sentinel is passed).
//
//   - Calls (*ComputeManifest).IsModeBEligible(time.Now(), window).
//
//   - Emits the verdict to stdout in a structured form the
//     federation CI workflow can consume:
//
//     mode_b_eligible=<true|false>
//     mode_b_reason=<reason string>
//     mode_b_phase=<phase>
//     mode_b_warning=<empty|"mode-b-best-effort">
//
//   - Exit codes:
//
//     0 — eligible (either "ok" at Walk-α+ or "best-effort" at
//     Crawl/Toddle); CI step passes
//     1 — NOT eligible at the given phase; CI step fails the PR
//     2 — error loading or validating the manifest itself (not a
//     policy decision; investigate manifest health)
//
// Per repo-bma-systema-issue-#171 Phase B-PR-8 deliverable; consumes
// model.IsModeBEligible per Phase B-PR-7 (PR #62, merged 36d2231).
// Cited verbatim by .github/workflows/ci-compute-manifest.yml.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/JamesPagetButler/wyrd/model"
)

// defaultWindow is the federation-canonical Walk-α substrate-
// credibility window per Spec 9.2 §3.1 amendment (recovery v0.2,
// `repo-inter` PR #6 merged `e774069`). Matches BMA's 72h Step 8
// continuous-operation gate (federation-pattern reuse).
const defaultWindow = 72 * time.Hour

func main() {
	root := flag.String("root", ".", "Wyrd repo root containing manifest/CURRENT")
	window := flag.Duration("window", defaultWindow, "substrate-credibility-window (Spec 9.2 §3.1; default 72h Walk-α target)")
	allowBootstrap := flag.Bool("allow-bootstrap-sentinel", false, "permit substrate.commit_sha == BootstrapSentinel (federation CI sets true only on the manifest-bootstrap PR's branch)")
	flag.Parse()

	if err := run(*root, *window, *allowBootstrap, os.Stdout); err != nil {
		if errors.Is(err, errIneligible) {
			os.Exit(1)
		}
		_, _ = fmt.Fprintf(os.Stderr, "mode-b-eligibility-check: %v\n", err)
		os.Exit(2)
	}
}

// errIneligible is the sentinel returned by run() when the predicate
// returns (false, reason) — distinguishes policy decisions from
// manifest-load failures.
var errIneligible = errors.New("mode (b) not eligible per Spec 9.2 §3.1")

func run(root string, window time.Duration, allowBootstrap bool, out *os.File) error {
	m, err := model.LoadComputeManifestWithOptions(root, model.LoadOptions{
		AllowBootstrapSentinel: allowBootstrap,
	})
	if err != nil {
		return fmt.Errorf("load Compute Manifest at %s: %w", root, err)
	}

	eligible, reason := m.IsModeBEligible(time.Now(), window)

	// Federation-CI-consumable output. Each line is `key=value` so the
	// workflow can shell-eval or grep without parsing free-form text.
	// Write errors on stdout are CI-runner-level failures, not policy
	// decisions; ignored intentionally — the caller's exit code is
	// the load-bearing signal.
	_, _ = fmt.Fprintf(out, "mode_b_eligible=%t\n", eligible)
	_, _ = fmt.Fprintf(out, "mode_b_reason=%s\n", reason)
	_, _ = fmt.Fprintf(out, "mode_b_phase=%s\n", m.Phase)
	if eligible && isBestEffortReason(reason) {
		_, _ = fmt.Fprintf(out, "mode_b_warning=mode-b-best-effort\n")
	} else {
		_, _ = fmt.Fprintf(out, "mode_b_warning=\n")
	}

	if !eligible {
		return errIneligible
	}
	return nil
}

// isBestEffortReason detects the best-effort reason prefix that
// model.IsModeBEligible emits during Crawl/Toddle. Used to decide
// whether to emit the `mode-b-best-effort` warning annotation.
func isBestEffortReason(reason string) bool {
	// model.IsModeBEligible best-effort branches all return reasons
	// starting with "best-effort:" — see model/compute_manifest.go.
	return len(reason) >= len("best-effort") && reason[:len("best-effort")] == "best-effort"
}
