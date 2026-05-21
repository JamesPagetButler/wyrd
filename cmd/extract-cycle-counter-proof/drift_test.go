// Package main — drift-detection test for the Lean ↔ Go pragmatic
// extraction harness pairing.
//
// Per Spec 9.2 §3.1 amendment "Mode (b) extraction pragmatism"
// (`repo-inter` PR #6, merged e774069) + design surface PR #63 §5
// (Phase C-PR-10 design):
//
//	The Lean predicate definitions and the Go validator code carry
//	paired doc-comments referencing each other (so manual review can
//	confirm they encode the same property). A CI drift-detection test
//	compares the source text of both representations against a
//	snapshot; CI fails if either drifts without the paired side
//	updating.
//
// This file implements that drift-detection check. On every CI run,
// it computes SHA-256 of:
//   - lean/Wyrd/SubstrateTrace.lean       (the Lean predicate source)
//   - cmd/extract-cycle-counter-proof/main.go (the Go validator)
//
// and compares against the committed snapshot at
// testdata/lean-go-parity.snap. If either hash drifts, the test
// fails, forcing the developer to:
//
//  1. Confirm both sides have been updated in lockstep (or roll back
//     the unintended change to one side).
//  2. Regenerate the snapshot via `go test -update` (the standard
//     snapshot-regen idiom).
//
// This is a forcing function, not a deep semantic check: any
// modification to either file trips it, including comment edits. The
// forcing function is the load-bearing discipline — every change to
// the predicate-pairing region requires an explicit "I checked the
// other side" gesture from the developer.

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// updateSnap regenerates the snapshot file when set true via
// `go test -run TestLeanGoParityDrift -args -update`.
var updateSnap = flag.Bool("update", false, "regenerate testdata/lean-go-parity.snap from current source")

// pairedFiles enumerates the source files whose paired-predicate
// contents are tracked by the drift snapshot. Adding a new pair
// (e.g., for a future substrate-tier theorem) is a per-line addition
// here + a corresponding snapshot regeneration via `-update`.
var pairedFiles = []struct {
	label string // appears in the snapshot file
	path  string // relative to the Wyrd repo root
}{
	{label: "lean.SubstrateTrace", path: "lean/Wyrd/SubstrateTrace.lean"},
	{label: "go.extractCycleCounterProof", path: "cmd/extract-cycle-counter-proof/main.go"},
}

// findRepoRoot walks up from the test's working directory to find
// the directory containing manifest/CURRENT (the Wyrd repo root).
// The test's CWD is the package directory (cmd/extract-cycle-counter-proof/)
// so we go up two levels.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed; cannot locate repo root")
	}
	// drift_test.go is at cmd/extract-cycle-counter-proof/drift_test.go
	// → repo root is two dirs up.
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

// hashFile returns the SHA-256 hex digest of the file at path.
func hashFile(t *testing.T, path string) string {
	t.Helper()
	// #nosec G304 -- path is constructed from a known-set of paired
	// files relative to the Wyrd repo root; no external taint.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// computeCurrentSnapshot returns the snapshot string built from
// current paired-file hashes. Snapshot format is two lines:
//
//	<label1> <sha256-hex>
//	<label2> <sha256-hex>
//
// Stable + grep-friendly + diff-friendly.
func computeCurrentSnapshot(t *testing.T, root string) string {
	t.Helper()
	var b strings.Builder
	for _, p := range pairedFiles {
		fullPath := filepath.Join(root, p.path)
		hash := hashFile(t, fullPath)
		b.WriteString(p.label)
		b.WriteString(" ")
		b.WriteString(hash)
		b.WriteString("\n")
	}
	return b.String()
}

// TestLeanGoParityDrift is the drift-detection forcing function. CI
// runs this on every PR touching either paired side (lean/** or
// cmd/extract-cycle-counter-proof/**); the test fails if either side
// changed without the snapshot being regenerated.
func TestLeanGoParityDrift(t *testing.T) {
	root := findRepoRoot(t)
	current := computeCurrentSnapshot(t, root)

	snapPath := filepath.Join(root, "cmd", "extract-cycle-counter-proof", "testdata", "lean-go-parity.snap")

	if *updateSnap {
		// #nosec G306 -- snapshot file is committed-source; 0600 is
		// fine since CI runs as a single user.
		if err := os.WriteFile(snapPath, []byte(current), 0o600); err != nil {
			t.Fatalf("update snapshot at %s: %v", snapPath, err)
		}
		t.Logf("snapshot regenerated at %s", snapPath)
		return
	}

	// #nosec G304 -- snapPath is constructed from runtime.Caller +
	// the known testdata layout.
	want, err := os.ReadFile(snapPath)
	if err != nil {
		t.Fatalf("read snapshot at %s: %v\n\n"+
			"first-run regeneration: go test -run TestLeanGoParityDrift -args -update",
			snapPath, err)
	}

	if string(want) != current {
		t.Errorf("lean ↔ go paired-source drift detected.\n\n"+
			"WANT (committed snapshot):\n%s\n"+
			"GOT  (current source hashes):\n%s\n"+
			"resolution: confirm both paired sides updated in lockstep,\n"+
			"then regenerate via:\n"+
			"  go test -run TestLeanGoParityDrift ./cmd/extract-cycle-counter-proof -args -update",
			string(want), current)
	}
}
