# Contributing to Wyrd

Wyrd follows the same Helpful Engineering conventions as
[BMA][bma] and [confluent-trust][cth]. The two binding documents are the
go-coding-guide and github-practices in `bma-systema/doc/` (private repo).
This file summarises the parts that matter most for first-time contributors,
plus what is specific to Wyrd.

## Dev setup

Requirements:

- Go 1.24+
- Lean toolchain matching `lean/lean-toolchain` (currently `v4.30.0-rc1`)
  via [elan](https://github.com/leanprover/elan)
- `golangci-lint` (install: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
- `gh` CLI authenticated (`gh auth status`)

```bash
git clone git@github.com:JamesPagetButler/wyrd.git
cd wyrd
go build ./...
go test -race ./...
golangci-lint run
```

To build the Lean corpus:

```bash
cd lean
lake exe cache get
lake build
```

## Workflow

For each issue:

1. **Plan** — read the issue, design approach, post the plan as an issue comment
2. **Branch** — `feat/<short-slug>` (e.g., `feat/04-hyperedge-symmetry`)
3. **Build** — implement, write tests, keep PRs under ~400 lines when possible
4. **PR** — `gh pr create` with the template; link the issue with `Closes #N`
5. **Review** — wait for CI green and self-review the diff in GitHub UI
6. **Squash-merge** — single clean commit on `main`

## Commits — Conventional Commits

```
type(scope): subject

Body explains *why*, not *what*. Wrap at 72.

Co-authored-by: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
```

Types: `feat`, `fix`, `refactor`, `test`, `doc`, `chore`, `perf`, `lean`.
Scopes: `model`, `compute`, `store`, `lean`, `cmd`, `schema`, `ci`,
`infra`, `doc`.

Subject: imperative, lowercase after the colon, no period, ≤50 chars.

The `lean` type/scope is Wyrd-specific: changes to the Lean corpus
should use `lean(<file>): <subject>`, e.g.
`lean(holographic): add quaternion-arity v1.5 theorem`.

## Code style highlights (Go)

- `model/` is **stdlib-only**. No external deps.
- `compute/` is stdlib-by-default but **may take well-vetted numerical
  dependencies** (e.g. `gonum.org/v1/gonum`) when the work in compute/
  needs them and the Go-stdlib alternative is materially worse.
  This is a posture shift from the original "stdlib-only everywhere"
  rule, settled at the 2026-05-06 closeout (Q4): the Hypergraph
  Laplacian primitive needs eigenvector computation, and a hand-rolled
  power-method in `compute/` carries more correctness risk than a
  vetted gonum import. The bar is "well-vetted by the broader Go
  numerical-computing community" — gonum, scientific/Go subpackages,
  similar. Casual deps are still discouraged. **`store/` keeps the
  stdlib-only convention** until a backend explicitly demands
  otherwise (MuninnDB / SurrealDB will, but those land at Walk-phase
  with their own §I4 review).
- Package names lowercase, single word, no underscores. No `utils`,
  `common`, `helpers`.
- Acronyms uppercase: `ID`, `MI`, `JSON`. So `NodeID`, `HyperedgeID`,
  `ChainID`.
- Errors lowercase, no period: `fmt.Errorf("model: hyperedge %s: %w", id, err)`.
  Always prefix with the package name so messages are greppable.
- Don't create interfaces with one implementation. Wait for the second
  backend.
- `t.Helper()` in test helpers; `t.Context()` for cancellation.
- `iota` enums need explicit `MarshalJSON`/`UnmarshalJSON` — never
  serialise the integer.

Run `gofmt -w .`, `goimports -w .`, `go vet ./...`, `golangci-lint run`
before pushing.

## Style highlights (Lean)

- Keep each theorem cluster in its own file under `lean/Wyrd/`.
- Top-of-file docstring includes: PHASE, COMPANION FILES, WITNESS (if
  applicable), CONNECTION TO THE WYRD CORPUS.
- Never land a `sorry` or a user-defined `axiom`. CI fails the build
  if either is detected.
- Match mathlib API at the pinned rev (`a090f46d`); when API drift
  bites, prefer fixing imports over downgrading.
- Quaternion / Cayley-Dickson constructions follow `Foundations.lean`
  conventions — `⟨0, 1, 0, 0⟩` for `i`, etc.

## Soundness anchor convention

Every Go API that makes a load-bearing claim about behaviour must cite
the Lean theorem(s) it relies on, with the cite in the doc comment:

```go
// Soundness: per `Wyrd.Hypergraph.hyperedge_preserves_incident_edges`
// (Phase 2 C-20a), …
```

PR review will reject Go APIs that claim invariants without a Lean cite,
unless the PR explicitly states "no Lean anchor — runtime-only, no
formal claim made."

## Adding a new `bma.runtime.*` runtime-anchor type

Wyrd reserves the `bma.runtime.*` `model.NodeType` prefix for runtime
anchors authored by BMA's WDEvent observer (per ADR-003 §I2 and
`@bma-implementor` `qbp-cu-walk` seq=11). The four current types
(`flag-norm-drift`, `obs-zd-detected`, `obs-runtime-counter`,
`obs-fault`) are the closed set as of v0.1; **adding a new type is an
additive change requiring §I4 review** because it widens a
cross-project interface contract (BMA writes them, CTH reads them, the
`IsRuntimeAnchor` helper classifies them).

The process — settled at the 2026-05-06 closeout (Q5):

1. **File a Wyrd issue** describing the new type: name, semantics,
   instance-ID schema, which WDEvent class triggers it, what CTH
   `cth_id` form (per ADR-003 §I2) it pairs with.
2. **§I4 review** by `@bma` and `@bma-implementor` on the issue.
   The new type changes the I-4 contract; both halves of the BMA pair
   approve before code lands.
3. **PR** adding:
   - The new typed constant in `model/runtime_namespace.go`
   - A test in `model/runtime_namespace_test.go` confirming the
     constant value and `IsRuntimeAnchor` classification
   - A line in `CHANGELOG.md` under the next release
   - A row added to the namespace table in `doc/integration/bma.md`
4. **PR review** by `@bma-implementor` (consumer) before merge.

Existing types (the four named above) are stable; renaming or
removing one is a breaking change for the BMA classifier and CTH
audit trail and would require a separate, named, migration design
doc.

## Adding a new theorem

1. Decide the Lean file (existing cluster or new file under `lean/Wyrd/`).
2. Add to `lean/Wyrd.lean` import list if a new file.
3. Build: `cd lean && lake build`. Must remain green; zero sorries,
   zero user-defined axioms.
4. Update `doc/archive/Wyrd-Proofs-Reference-vN.md` with the new theorem
   listing.
5. If a Go API is the consumer, add the cite to its doc comment in the
   same PR.

## Signed commits (SSH)

Branch protection on `main` will eventually require signed commits. Set
up locally:

```bash
git config --local commit.gpgsign true
git config --local gpg.format ssh
git config --local user.signingkey ~/.ssh/id_ed25519.pub
```

Upload `id_ed25519.pub` to GitHub as both an Authentication key **and**
a Signing key.

[bma]: https://github.com/JamesPagetButler/bma-systema
[cth]: https://github.com/JamesPagetButler/confluent-trust
