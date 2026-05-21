// Command extract-cycle-counter-proof is the pragmatic mode-(b)
// extraction-and-execute harness for the substrate-tier theorem
// `Wyrd.CycleCounterCrossPhase.cycle_counter_monotonic_per_phase`
// (Phase C-PR-12 / PR #66, merged c81b7a7).
//
// Per Spec 9.2 §3.1 amendment "Mode (b) extraction pragmatism"
// (`repo-inter` PR #6, merged e774069):
//
//	Lean 4 lacks a stable extract-to-executable pipeline analogous to
//	Lean 3's. The federation has explicitly accepted a pragmatic-
//	extraction discipline: hand-write the substrate-runtime harness in
//	Go, with paired doc-comments referencing the Lean theorem, plus
//	CI drift-detection.
//
// This binary:
//
//  1. Loads the Compute Manifest to learn which substrate to run
//     against (Crawl phase → qbp-compute-unit/emulator).
//  2. Constructs a small QBP-CU emulator program of ≥1000
//     retired single-cycle instructions (QADD; cycle += 1 per Step).
//  3. Steps the emulator instruction-by-instruction, capturing the
//     CPU.Cycles counter after each Step into an InstructionEvent
//     slice.
//  4. Validates the captured trace against the SAME predicates
//     `Wyrd.SubstrateTrace.Monotonic` and `Wyrd.SubstrateTrace.AdvanceByOne`
//     define in the Lean source (this file's Monotonic + AdvanceByOne
//     functions are the Go-side paired definitions).
//  5. Emits a structured run log to stdout + writes a sample copy
//     to testdata/crawl-emulator-run.log.
//  6. Exit 0 if both predicates hold; non-zero if either fails.
//
// SCOPE — single-cycle opcode subset (load-bearing limitation):
//
// The QBP-CU emulator's QROT opcode increments CPU.Cycles by 2 per
// retired instruction ("Two QMULs" — composite operation per
// repo-qbp-compute-unit-pr-#33 §5.4). This violates the strict-
// equality AdvanceByOne predicate. The harness's program uses ONLY
// the single-cycle opcode subset (QADD; QMUL/QCONJ/QNORM/FANO are
// also single-cycle but QADD is the simplest to encode without
// side-state). v0.2 theorem refinement may relax AdvanceByOne to
// admit composite-op cycle accounting (per PR #63 §10 NOT-DECIDED:
// "Cycle-counter monotonicity vs concurrent dispatch ... out of scope
// for v0.1; Walk-α may require an A22 §4.2 amendment").
//
// Paired-doc-comment discipline (Lean ↔ Go drift detection):
//
// The Monotonic + AdvanceByOne Go functions below carry doc comments
// that quote the Lean predicates from `lean/Wyrd/SubstrateTrace.lean`
// verbatim. drift_test.go computes SHA-256 of both representations
// against a snapshot at testdata/lean-go-parity.snap; CI fails if
// either drifts without the paired side updating.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/JamesPagetButler/qbp-compute-unit/emulator"
	"github.com/JamesPagetButler/wyrd/model"
)

// InstructionEvent is the Go-side paired definition of the Lean
// `Wyrd.SubstrateTrace.InstructionEvent` structure:
//
//	structure InstructionEvent where
//	  cycle : Nat
//	  deriving DecidableEq, Repr
//
// Minimal abstraction per PR #63 §3.3 deliberate-minimal-abstraction —
// downstream consumers refine richer event data at their own anchor
// site without invalidating this layer.
type InstructionEvent struct {
	Cycle uint64
}

// Monotonic is the Go-side paired definition of the Lean
// `Wyrd.SubstrateTrace.Monotonic` predicate:
//
//	def Monotonic (t : SubstrateTrace m) : Prop :=
//	  ∀ (i j : Nat) (hij : i < j) (hj : j < t.events.length),
//	    (t.events.get ⟨i, Nat.lt_trans hij hj⟩).cycle ≤
//	    (t.events.get ⟨j, hj⟩).cycle
//
// The cycle counter is non-decreasing across the trace.
//
// Returns (true, 0) when satisfied; (false, i) where i is the first
// index where monotonicity breaks (i.e., trace[i].Cycle > trace[i+1].Cycle).
func Monotonic(events []InstructionEvent) (bool, int) {
	for i := 0; i+1 < len(events); i++ {
		if events[i].Cycle > events[i+1].Cycle {
			return false, i
		}
	}
	return true, 0
}

// AdvanceByOne is the Go-side paired definition of the Lean
// `Wyrd.SubstrateTrace.AdvanceByOne` predicate:
//
//	def AdvanceByOne (t : SubstrateTrace m) : Prop :=
//	  ∀ (i : Nat) (hsucc : i + 1 < t.events.length),
//	    (t.events.get ⟨i+1, hsucc⟩).cycle =
//	    (t.events.get ⟨i, Nat.lt_of_succ_lt hsucc⟩).cycle + 1
//
// The cycle counter advances by exactly 1 per retired instruction.
//
// Returns (true, 0) when satisfied; (false, i) where i is the first
// index where the strict-equality property fails (i.e., trace[i+1].Cycle
// != trace[i].Cycle + 1).
func AdvanceByOne(events []InstructionEvent) (bool, int) {
	for i := 0; i+1 < len(events); i++ {
		if events[i+1].Cycle != events[i].Cycle+1 {
			return false, i
		}
	}
	return true, 0
}

// substrateCommitSHARequired is the federation-canonical phase tag
// this harness validates against. Per the merged Compute Manifest
// at manifest/CURRENT, Crawl is the current phase; substrate is
// qbp-compute-unit/emulator. If the manifest names a different
// substrate, the harness refuses to run (the substrate would not be
// the one this paired-doc-comment harness was written against).
const expectedSubstrateRepo = "github.com/JamesPagetButler/qbp-compute-unit"

func main() {
	root := flag.String("root", ".", "Wyrd repo root containing manifest/CURRENT")
	steps := flag.Int("steps", 1024, "number of single-cycle instructions to retire (must be ≥ 1000 per design doc §5)")
	logPath := flag.String("log", "", "optional: write structured run log to this path in addition to stdout")
	flag.Parse()

	if *steps < 1000 {
		_, _ = fmt.Fprintf(os.Stderr, "extract-cycle-counter-proof: --steps must be ≥ 1000 per PR #63 §5; got %d\n", *steps)
		os.Exit(2)
	}

	if err := run(*root, *steps, *logPath, os.Stdout); err != nil {
		if errors.Is(err, errPredicateViolation) {
			os.Exit(1)
		}
		_, _ = fmt.Fprintf(os.Stderr, "extract-cycle-counter-proof: %v\n", err)
		os.Exit(2)
	}
}

var errPredicateViolation = errors.New("substrate-trace predicate violation (Monotonic or AdvanceByOne)")

func run(root string, steps int, logPath string, stdout io.Writer) error {
	// Step 1: confirm the manifest names the expected substrate.
	m, err := model.LoadComputeManifestWithOptions(root, model.LoadOptions{
		AllowBootstrapSentinel: true,
	})
	if err != nil {
		return fmt.Errorf("load Compute Manifest at %s: %w", root, err)
	}
	if m.Substrate.Repo != expectedSubstrateRepo {
		return fmt.Errorf("manifest names substrate.repo=%q; harness paired-doc-comments target %q only", m.Substrate.Repo, expectedSubstrateRepo)
	}

	// Step 2: build a single-cycle QADD program. Each QADD instruction
	// retires in 1 cycle on the QBP-CU emulator (per
	// repo-qbp-compute-unit/emulator/isa.go Funct7QADD case).
	//
	// Instruction encoding (per emulator/isa.go Decode):
	//   bits  [6:0]   Opcode    = OpcodeCustom0 (0x0B)
	//   bits  [11:7]  Rd        = destination register
	//   bits  [14:12] Funct3    = gearbox width (3 = W64 = single-cycle fast path)
	//   bits  [19:15] Rs1       = source register 1
	//   bits  [24:20] Rs2       = source register 2
	//   bits  [31:25] Funct7    = Funct7QADD (1) = single-cycle qadd64
	program := buildQADDProgram(steps)

	// Step 3: run the emulator step-by-step, capturing CPU.Cycles
	// after each retired instruction.
	cpu := emulator.NewCPU()
	cpu.SetWidth(emulator.W64)

	events := make([]InstructionEvent, 0, steps)
	for _, word := range program {
		if err := cpu.Step(word); err != nil {
			return fmt.Errorf("emulator Step failed at instruction %d (word 0x%08X): %w", len(events), word, err)
		}
		events = append(events, InstructionEvent{Cycle: cpu.Cycles})
	}

	// Step 4: validate the predicates.
	monoOK, monoFailIdx := Monotonic(events)
	advOK, advFailIdx := AdvanceByOne(events)

	// Step 5: emit structured run log.
	var logBuf strings.Builder
	emitRunLog(&logBuf, m, events, monoOK, monoFailIdx, advOK, advFailIdx)
	logStr := logBuf.String()
	if _, err := io.WriteString(stdout, logStr); err != nil {
		return fmt.Errorf("write stdout: %w", err)
	}
	if logPath != "" {
		// #nosec G304 G306 -- logPath is a CLI flag from the operator
		// or CI runner; intentionally write to wherever they specify.
		// Permissions: 0600 (only the runner can read; CI re-reads
		// after the binary exits).
		if err := os.WriteFile(logPath, []byte(logStr), 0o600); err != nil {
			return fmt.Errorf("write log file %s: %w", logPath, err)
		}
	}

	if !monoOK || !advOK {
		return errPredicateViolation
	}
	return nil
}

// buildQADDProgram constructs a program of `steps` QADD instructions
// at W64 width (Funct3 = 3) operating on alternating register pairs.
// Each instruction retires in exactly 1 cycle per the QBP-CU emulator
// at the Crawl phase (per `cpu.Cycles += 1` for Funct7QADD).
func buildQADDProgram(steps int) []uint32 {
	program := make([]uint32, steps)
	for i := range program {
		rd := uint32(uint8(i%30) + 1) // skip r0 (zero reg conventionally)
		rs1 := uint32(uint8((i+1)%30) + 1)
		rs2 := uint32(uint8((i+2)%30) + 1)
		program[i] = uint32(emulator.OpcodeCustom0) |
			(rd << 7) |
			(uint32(3) << 12) | // Funct3 = 3 → W64 fast path
			(rs1 << 15) |
			(rs2 << 20) |
			(uint32(emulator.Funct7QADD) << 25)
	}
	return program
}

func emitRunLog(w io.Writer, m *model.ComputeManifest, events []InstructionEvent, monoOK bool, monoFailIdx int, advOK bool, advFailIdx int) {
	_, _ = fmt.Fprintf(w, "# extract-cycle-counter-proof — mode-(b) extraction harness run log\n")
	_, _ = fmt.Fprintf(w, "# Per repo-bma-systema-issue-#170 Phase C-PR-13; substrate-tier theorem\n")
	_, _ = fmt.Fprintf(w, "# Wyrd.CycleCounterCrossPhase.cycle_counter_monotonic_per_phase (PR #66, c81b7a7).\n")
	_, _ = fmt.Fprintf(w, "captured_at=%s\n", time.Now().UTC().Format(time.RFC3339))
	_, _ = fmt.Fprintf(w, "manifest_phase=%s\n", m.Phase)
	_, _ = fmt.Fprintf(w, "substrate_repo=%s\n", m.Substrate.Repo)
	_, _ = fmt.Fprintf(w, "substrate_module=%s\n", m.Substrate.Module)
	_, _ = fmt.Fprintf(w, "substrate_pinned_tag=%s\n", m.Substrate.PinnedTag)
	_, _ = fmt.Fprintf(w, "instructions_retired=%d\n", len(events))

	if len(events) > 0 {
		_, _ = fmt.Fprintf(w, "first_cycle=%d\n", events[0].Cycle)
		_, _ = fmt.Fprintf(w, "last_cycle=%d\n", events[len(events)-1].Cycle)
		_, _ = fmt.Fprintf(w, "cycle_delta=%d\n", events[len(events)-1].Cycle-events[0].Cycle)
	}

	_, _ = fmt.Fprintf(w, "predicate_monotonic=%t\n", monoOK)
	if !monoOK {
		_, _ = fmt.Fprintf(w, "predicate_monotonic_fail_index=%d\n", monoFailIdx)
		_, _ = fmt.Fprintf(w, "predicate_monotonic_fail_neighborhood=cycle[%d]=%d cycle[%d]=%d\n",
			monoFailIdx, events[monoFailIdx].Cycle,
			monoFailIdx+1, events[monoFailIdx+1].Cycle)
	}

	_, _ = fmt.Fprintf(w, "predicate_advance_by_one=%t\n", advOK)
	if !advOK {
		_, _ = fmt.Fprintf(w, "predicate_advance_by_one_fail_index=%d\n", advFailIdx)
		_, _ = fmt.Fprintf(w, "predicate_advance_by_one_fail_neighborhood=cycle[%d]=%d cycle[%d]=%d (expected %d, got %d)\n",
			advFailIdx, events[advFailIdx].Cycle,
			advFailIdx+1, events[advFailIdx+1].Cycle,
			events[advFailIdx].Cycle+1, events[advFailIdx+1].Cycle)
	}

	verdict := "mode_b_eligible"
	if !monoOK || !advOK {
		verdict = "mode_b_predicate_violation"
	}
	_, _ = fmt.Fprintf(w, "verdict=%s\n", verdict)
}
