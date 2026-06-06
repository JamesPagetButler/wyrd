package wyrd

// Soundness-citation sweep (Sprint 3 W3-4; wyrd D2 spec-compliance
// gate per inter/crawl-completion-framework.md).
//
// The repo's soundness pattern: every Go API carrying a formal claim
// cites its Lean anchor as `Wyrd.<Module>.<name>` in a doc comment,
// and "diverging from the spec without updating the theorem (or vice
// versa) is an audit failure" (README §Soundness pattern). This test
// makes the citation half of that audit mechanical: every citation in
// non-test Go source must resolve to a name that actually appears in
// the cited Lean module.
//
// NT_SEAM_RECORD_001 (Notary Cycle 1, 2026-05-20) was exactly a
// phantom citation — compute/quaternion.go citing a theorem that
// `lean/Wyrd/Foundations.lean` did not contain. This sweep would have
// caught it mechanically.
//
// Forward-pin exemption: a citation whose surrounding comment block
// contains the word "forthcoming" is an explicit forward-pin to an
// unauthored theorem (per the federation phantom-artifact rule:
// forward-pins must be MARKED, not silent). Those are skipped here —
// but if the cited module file exists AND already contains the name,
// the "forthcoming" marker is stale and the test fails so the
// citation gets promoted to a real one.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// citationRe matches `Wyrd.<Module>.<name>` Lean citations in Go
// source. Module is UpperCamelCase; name is a Lean identifier (may be
// a trailing-underscore prefix family like `no_surjection_`).
var citationRe = regexp.MustCompile(`Wyrd\.([A-Z][A-Za-z0-9]*)\.([A-Za-z_][A-Za-z0-9_]*)`)

// forthcomingWindow is how many lines around a citation are scanned
// for the "forthcoming" forward-pin marker. Kept tight (±1) so a
// marker only exempts the citation it is actually attached to —
// e.g. "per (forthcoming) Wyrd.X.y" (marker one line above) or
// "Wyrd.X.y\n// (forthcoming; …)" (marker one line below) — and a
// neighboring bullet's marker cannot leak onto a real citation.
const forthcomingWindow = 1

func TestSoundnessCitationsResolve(t *testing.T) {
	leanSrc := loadLeanModules(t)

	var goFiles []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == "lean" || name == "testdata" || name == ".git" || name == ".lake" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if len(goFiles) == 0 {
		t.Fatal("no Go source files found — sweep is miswired")
	}

	citations := 0
	for _, path := range goFiles {
		// #nosec G304 -- path comes from filepath.Walk over the repo's
		// own source tree (test-only sweep; no external input).
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		lines := strings.Split(string(raw), "\n")
		for i, line := range lines {
			for _, m := range citationRe.FindAllStringSubmatch(line, -1) {
				module, name := m[1], m[2]
				citations++
				leanBody, moduleExists := leanSrc[module]
				resolves := moduleExists && strings.Contains(leanBody, name)
				forthcoming := markedForthcoming(lines, i)

				switch {
				case forthcoming && resolves:
					t.Errorf("%s:%d: citation Wyrd.%s.%s is marked \"forthcoming\" but the name ALREADY EXISTS in lean/Wyrd/%s.lean — promote the forward-pin to a real citation",
						path, i+1, module, name, module)
				case forthcoming:
					// Honest forward-pin; allowed.
				case !moduleExists:
					t.Errorf("%s:%d: phantom citation Wyrd.%s.%s — lean/Wyrd/%s.lean does not exist (NT_SEAM_RECORD_001 class)",
						path, i+1, module, name, module)
				case !resolves:
					t.Errorf("%s:%d: phantom citation Wyrd.%s.%s — name %q not found in lean/Wyrd/%s.lean (NT_SEAM_RECORD_001 class)",
						path, i+1, module, name, name, module)
				}
			}
		}
	}

	if citations == 0 {
		t.Fatal("no Wyrd.<Module>.<name> citations found — regex or layout drifted; sweep is vacuous")
	}
	t.Logf("swept %d Go files, %d citations", len(goFiles), citations)
}

// loadLeanModules reads every lean/Wyrd/*.lean file into a
// module-name → source map.
func loadLeanModules(t *testing.T) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join("lean", "Wyrd"))
	if err != nil {
		t.Fatalf("read lean/Wyrd: %v (run from repo root)", err)
	}
	out := make(map[string]string, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".lean") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join("lean", "Wyrd", e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		out[strings.TrimSuffix(e.Name(), ".lean")] = string(raw)
	}
	if len(out) == 0 {
		t.Fatal("no Lean modules found under lean/Wyrd/")
	}
	return out
}

// markedForthcoming reports whether any line within
// forthcomingWindow of index i contains the forward-pin marker.
func markedForthcoming(lines []string, i int) bool {
	lo := max(i-forthcomingWindow, 0)
	hi := min(i+forthcomingWindow, len(lines)-1)
	for j := lo; j <= hi; j++ {
		if strings.Contains(strings.ToLower(lines[j]), "forthcoming") {
			return true
		}
	}
	return false
}
