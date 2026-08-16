## Summary
<!-- Reference the Linear issue, if any -->
Refs: GHP-

<!-- What changed and why, in 1-3 bullets -->

## How to test

<!-- Steps for the reviewer to reproduce/validate the change -->

- [ ]

## Checklist

- [ ] `gofmt -l ./src` no output and `go vet ./src/...` clean; `go test -short ./src/... -race` and `go test ./src/test/integration/... -race` pass locally
- [ ] Tests covering the new case (CI requires minimum 90% coverage)
- [ ] PR title follows [Conventional Commits](https://www.conventionalcommits.org/en/) (`type(scope): description`)
- [ ] Documentation updated if necessary (`docs/`, `README.md`)

## Impact

- Breaking change (`.ghp` syntax or `runtime` API)? y/n
- Manual action needed after merge (migration, new config, secret)? y/n
