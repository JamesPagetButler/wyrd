/-
  Wyrd/Substrate.lean

  Substrate-tier import aggregator — the canonical Lean module that
  imports all federation substrate-tier promoted theorems.

  Helpful Engineering — Quaternion-Based Physics Programme
  May 2026 — Rev 0.1

  ============================================================
  PURPOSE
  ============================================================

  Per Spec 9.2 §2 (Promotion Gate) + §5 (Substrate Immutability),
  theorems promoted to the substrate tier are constitutionally
  frozen — their statements MUST NOT be edited post-promotion;
  amendments are deprecate-and-replace only.

  This module is the canonical aggregator import. Adding a theorem
  here is the **promotion** action — the act of placing the
  `import Wyrd.<TheoremModule>` line in Substrate.lean is itself
  the constitutional commitment that the theorem is substrate-tier
  per Spec 9.2 §2.

  Pattern mirrors mathlib's `Mathlib.lean` import-aggregator: a
  single, grep-discoverable index of every theorem the federation
  has constitutionally pinned. Downstream consumers can:

    import Wyrd.Substrate

  to bring the entire substrate-tier corpus into scope without
  needing to know each individual theorem's module path.

  ============================================================
  PROMOTION DISCIPLINE
  ============================================================

  Adding a line below is a substrate-tier promotion event. It
  requires:

    - The theorem must satisfy Spec 9.2 §2 four-criteria (compiles
      end-to-end on pinned toolchain; no `sorry`; no user-defined
      `axiom`; passes the Compute-Substrate Gate per §3 mode (a)
      or mode (a) + (b))
    - Per Spec 9.2 §9, first-10 substrate promotions require
      explicit @beekeeper HVR
    - Per Spec 9.2 §5 substrate immutability, the theorem's
      statement is frozen post-promotion (constitutionally
      immutable; amendments are deprecate-and-replace)

  Removing a line below is a deprecation event. Per §5, the
  deprecated theorem remains proved; downstream consumers
  migrate at their own pace.

  ============================================================
  SUBSTRATE-TIER REGISTRY
  ============================================================

  Each substrate-tier theorem promoted to this registry is
  documented at `doc/promotion/<YYYY-MM>-<theorem-name>.md`
  with mode declaration + verification evidence.
-/

-- Promotion #1 (2026-05-21): cycle_counter_monotonic_per_phase
-- Mode: (a) + (b)
-- Verification: lean elaborator (5 phase instances)
--               + pragmatic extraction harness against Crawl-phase
--                 QBP-CU emulator (cmd/extract-cycle-counter-proof
--                 testdata/crawl-emulator-run.log: 1024 instructions,
--                 verdict=mode_b_eligible)
-- Promotion PR: repo-wyrd-pr-#68
-- Promotion doc: doc/promotion/2026-05-cycle-counter-cross-phase.md
-- First-10 HVR (Spec 9.2 §9): #1 of 10 first-10 promotions
import Wyrd.CycleCounterCrossPhase
