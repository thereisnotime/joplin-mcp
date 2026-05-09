# Tasks: joplin-mcp v1

## Phase 0 — Bootstrap
- [x] Create GitHub repo `thereisnotime/joplin-mcp` (public, AGPL-3.0)
- [x] Clone via SSH into `~/Private/Projects/P/joplin-mcp`
- [x] `openspec init --tools claude`
- [x] Draft proposal/design/tasks
- [x] `go.mod` (module `github.com/thereisnotime/joplin-mcp`, Go 1.26+)
- [x] Directory skeleton: `cmd/joplin-mcp/`, `internal/joplin/`, `internal/tools/`,
      `internal/version/`
- [x] `.gitignore`, `.editorconfig`, `.golangci.yml`
- [x] `LICENSE` (AGPL-3.0)
- [x] `README.md` with badges
- [x] `CONTRIBUTING.md`, `SECURITY.md`, `CHANGELOG.md`, `CLAUDE.md`,
      `CODE_OF_CONDUCT.md`
- [x] `justfile`
- [x] `.commitlintrc.yaml` (Conventional Commits)
- [x] `.github/CODEOWNERS`, `dependabot.yml`, `PULL_REQUEST_TEMPLATE.md`,
      `ISSUE_TEMPLATE/*.yml`, `codecov.yml`
- [x] Initial commit + push, default branch set to `main`

## Phase 1 — Joplin REST client (`internal/joplin`)
- [x] `client.go` — `Client`, `Options`, request helpers, `APIError`,
      `IsNotFound`, raw GET for binary downloads
- [x] `paginate.go` — generic `CollectAll` helper
- [x] `types.go` — full structs for `Note`, `Folder`, `Tag`, `Resource`,
      `Revision`, `Event`, with every documented field including
      `encryption_applied` and `master_key_id`
- [x] `notes.go` — list / get / create / update / delete + note tags +
      note resources
- [x] `folders.go` — list / get / create / update / delete + folder notes
- [x] `tags.go` — list / get / create / delete + tag/untag note + tag notes
- [x] `search.go` — search across notes, folders, tags, resources with
      Joplin query syntax + pagination
- [x] `resources.go` — list / get-metadata / download / upload / delete
      with multipart binary upload
- [x] `events.go` — `/events?cursor=`
- [x] `revisions.go` — list / get / per-note revision filter
- [x] `client_test.go` + `coverage_test.go` — `httptest.Server`-driven
      tests for every method (auth, timeouts, context cancel, full-field
      decoding, unknown-field tolerance, pagination, multipart, errors)

## Phase 2 — MCP wiring (`internal/tools` + `cmd/joplin-mcp`)
- [x] `internal/tools/types.go` — output projections with encryption
      annotations
- [x] `internal/tools/encryption.go` — `noteOut` etc., `pageOf` skip count
- [x] `internal/tools/server.go` — `New()` constructor that registers
      every tool
- [x] `internal/tools/notes.go` — 6 note tools incl. `get_note_with_context`
- [x] `internal/tools/folders.go` — 6 folder tools
- [x] `internal/tools/tags.go` — 7 tag tools
- [x] `internal/tools/search.go` — search tool
- [x] `internal/tools/resources.go` — 5 resource tools (including
      `download_resource` refusal on encrypted resources)
- [x] `internal/tools/events.go` — `list_changes_since`
- [x] `internal/tools/revisions.go` — 2 revision tools
- [x] `cmd/joplin-mcp/main.go` — env parsing
      (`JOPLIN_TOKEN`/`BASE_URL`/`TIMEOUT`/`LOG_LEVEL`), `log/slog` to
      stderr, signal handling
- [x] `internal/version/version.go` — ldflags-injected version string
- [x] `internal/tools/tools_test.go` + `coverage_test.go` — end-to-end
      tests via in-memory MCP transport against an httptest fake Joplin

## Phase 3 — Infra
- [x] `.goreleaser.yaml` — linux/darwin/windows × amd64/arm64, cosign
      keyless, syft SBOM
- [x] `.github/workflows/ci.yaml` — build, vet, gofmt, test+coverage
      (80% gate, `coverpkg=./internal/...`), gosec, govulncheck, Trivy SCA,
      Codecov upload
- [x] `.github/workflows/release.yaml` — GoReleaser on `v*` tags +
      build-provenance attestation
- [x] `.github/workflows/scorecard.yaml` — OpenSSF Scorecard
- [x] `.github/ISSUE_TEMPLATE/bug_report.yml`, `feature_request.yml`,
      `config.yml`
- [x] `.github/PULL_REQUEST_TEMPLATE.md`
- [x] `codecov.yml`

## Phase 4 — Polish + release
- [ ] README — full tool table, env vars, encryption notes, badges
- [ ] CHANGELOG entry for `v0.1.0`
- [ ] Tag `v0.1.0` and push
- [ ] Verify GoReleaser workflow succeeds, artifacts published
- [ ] Apply for OpenSSF Best Practices badge (manual; bestpractices.dev)
- [ ] `openspec archive joplin-mcp-v1 --skip-specs -y`
