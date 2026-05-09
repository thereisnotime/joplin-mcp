# Contributing to joplin-mcp

Thanks for taking the time to contribute. Here is everything you need to get started.

## Getting started

```sh
git clone git@github.com:thereisnotime/joplin-mcp.git
cd joplin-mcp
go mod download
just build   # outputs bin/joplin-mcp
just test    # run tests with race detector
```

Requirements: [Go 1.25+](https://go.dev/dl/), [just](https://just.systems/).

## Making changes

1. Fork the repo and create a branch from `main`.
2. Make your change. If it adds behaviour, add or update tests.
3. Run `just test` and confirm everything passes.
4. Run `go vet ./...` — no new warnings.
5. Open a pull request against `main`. Keep the title short and descriptive.

PRs require at least one approving review and all CI checks to pass before merging.

## Conventional commits

Commit subjects follow [Conventional Commits](https://www.conventionalcommits.org/) and
are enforced by commitlint:

```
feat: add list_changes_since tool
fix: respect context cancellation in download_resource
docs: clarify encryption-transparency in README
```

Allowed types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`,
`ci`, `chore`, `revert`.

## Running the full CI suite locally

```sh
just test                                                    # unit tests + race detector
go install github.com/securego/gosec/v2/cmd/gosec@v2.25.0
gosec ./...                                                  # SAST
go install golang.org/x/vuln/cmd/govulncheck@v1.1.4
govulncheck ./...                                            # SCA
```

## Project layout

```
cmd/joplin-mcp/      Entry point — env parsing, MCP server wiring, signal handling
internal/
  joplin/            REST client for Joplin's Web Clipper API
  tools/             MCP tool handlers using the official Go MCP SDK
  version/           Build-time version string (ldflags-injected)
openspec/            Spec-driven development artefacts (proposals, designs, tasks)
```

## Tests

Tests live next to the code they cover (`*_test.go`). Joplin client tests run against
`net/http/httptest.Server` and do not need a real Joplin Desktop. End-to-end tests
gated behind `JOPLIN_E2E=1` hit a real Joplin Desktop with a real token.

Aim to keep overall coverage above 80%. Check with:

```sh
just cover
```

## Reporting issues

Bug reports and feature requests are welcome via [GitHub Issues](https://github.com/thereisnotime/joplin-mcp/issues).
For security vulnerabilities, please follow the process in [SECURITY.md](SECURITY.md).

## Coding standards

All contributions must follow the official Go coding standards:

- **[Effective Go](https://go.dev/doc/effective_go)** — the primary style reference.
- **[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)** —
  common mistakes and idioms reviewed in Go PRs.
- **[Go Test Comments](https://github.com/golang/go/wiki/TestComments)** — conventions
  for writing good Go tests.

In addition:

- Format all code with `gofmt` (or `goimports`) before committing.
- Keep `cmd/` thin — business logic belongs in `internal/`.
- Error messages are lowercase and do not end with punctuation
  (e.g. `"note not found"`, not `"Note not found."`).
- No `Co-Authored-By` trailers in commits.
- stdout is reserved for the MCP transport. Logs go to stderr via `log/slog`.

## License

By contributing you agree that your work will be released under the
[AGPL-3.0 License](LICENSE).
