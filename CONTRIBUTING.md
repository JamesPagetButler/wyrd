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

- Pure Go in `model/` and `compute/` — stdlib only, no external deps.
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
