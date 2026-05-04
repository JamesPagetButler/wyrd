## Summary

<!-- 1-3 bullets explaining what this PR does and why. -->

## Closes / Related

<!-- "Closes #N" links automatically. List related but-not-closed issues separately. -->

Closes #

## Test plan

- [ ] `go build ./...`
- [ ] `go test -race ./...`
- [ ] `golangci-lint run`
- [ ] If `lean/` touched: `cd lean && lake build` (zero sorries, zero axioms)
- [ ] If a Lean theorem gained a Go consumer: doc comment on the Go API cites the theorem by file:name

## Soundness check

<!-- If this PR adds or changes a Go API that has a load-bearing invariant, name the Lean theorem
     it relies on (or "no Lean anchor — runtime-only, no formal claim made"). -->
