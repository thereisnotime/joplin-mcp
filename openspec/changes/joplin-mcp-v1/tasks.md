# Tasks: joplin-mcp v1

## Phase 0 — Bootstrap
- [x] Create GitHub repo `thereisnotime/joplin-mcp` (public, AGPL-3.0)
- [x] Clone via SSH into `~/Private/Projects/P/joplin-mcp`
- [x] `openspec init --tools claude`
- [x] Draft proposal/design/tasks
- [ ] `go.mod` (module `github.com/thereisnotime/joplin-mcp`, Go 1.23+)
- [ ] Directory skeleton: `cmd/joplin-mcp/`, `internal/joplin/`, `internal/tools/`,
      `internal/version/`
- [ ] `.gitignore`, `.editorconfig`, `.golangci.yml`
- [ ] `LICENSE` already present (AGPL-3.0 from `gh repo create`)
- [ ] `README.md` skeleton with badges
- [ ] `CONTRIBUTING.md`, `SECURITY.md`, `CHANGELOG.md`, `CLAUDE.md`
- [ ] `justfile`
- [ ] Initial commit + push

## Phase 1 — Joplin REST client (`internal/joplin`)
- [ ] `client.go` — `Client`, `ClientOptions`, request helpers, `APIError`
- [ ] `paginate.go` — generic pagination helper
- [ ] `types.go` — full structs for Note, Folder, Tag, Resource, Revision, Event,
      MasterKey, with every documented field
- [ ] `notes.go` — list/get/create/update/delete
- [ ] `folders.go` — list/get/create/update/delete + list-notes-in-folder
- [ ] `tags.go` — list/get/create/delete + tag/untag/list-notes-with-tag
- [ ] `search.go` — search with query syntax + pagination
- [ ] `resources.go` — list/get-metadata/download/upload/delete (binary handling)
- [ ] `events.go` — `/events?cursor=`
- [ ] `revisions.go` — list/get note revisions
- [ ] `*_test.go` — `httptest.Server` for every method, ≥80% coverage

## Phase 2 — MCP wiring (`internal/tools` + `cmd/joplin-mcp`)
- [ ] `internal/tools/tools.go` — `JoplinClient` interface, registration helper
- [ ] `internal/tools/notes.go` — 6 note tools
- [ ] `internal/tools/folders.go` — 6 folder tools
- [ ] `internal/tools/tags.go` — 7 tag tools
- [ ] `internal/tools/search.go` — search tool
- [ ] `internal/tools/resources.go` — 5 resource tools
- [ ] `internal/tools/events.go` — events tool
- [ ] `internal/tools/revisions.go` — 2 revision tools
- [ ] `internal/tools/encryption.go` — shared helper that annotates outputs with
      encryption state and counts skipped encrypted items
- [ ] `cmd/joplin-mcp/main.go` — env parsing, server wiring, signal handling
- [ ] `internal/version/version.go` — ldflags-injected version
- [ ] `internal/tools/*_test.go` — fake-client unit tests

## Phase 3 — Infra
- [ ] `.goreleaser.yaml` — linux/darwin/windows × amd64/arm64, cosign keyless,
      SBOM via syft
- [ ] `.github/workflows/ci.yaml` — build, vet, gofmt, test+coverage, gosec, govulncheck,
      Trivy SCA, Codecov upload, 80% coverage gate
- [ ] `.github/workflows/release.yaml` — GoReleaser on `v*` tags
- [ ] `.github/workflows/scorecard.yaml` — OpenSSF Scorecard
- [ ] `.github/ISSUE_TEMPLATE/bug_report.md`, `feature_request.md`
- [ ] `.github/pull_request_template.md`
- [ ] `codecov.yml`

## Phase 4 — Polish + release
- [ ] README — quickstart, Claude Desktop config, full tool table, env vars,
      encryption notes, badges, links to Releases / pkg.go.dev / Codecov / Scorecard
- [ ] CHANGELOG entry for `v0.1.0`
- [ ] `just build` smoke test
- [ ] Manual smoke test against a real Joplin Desktop (if available)
- [ ] Tag `v0.1.0` and push
- [ ] Verify GoReleaser workflow succeeds, artifacts published
- [ ] Apply for OpenSSF Best Practices badge (manual — fills out a survey at
      bestpractices.dev; can be done after first release)
- [ ] `openspec archive joplin-mcp-v1 --skip-specs -y`
