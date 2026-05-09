# Changelog

All notable changes to joplin-mcp are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the
project uses [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

### Added

- `.env` support: the binary auto-loads a `.env` file from the current
  working directory at startup, populating any environment variables not
  already set by the shell. Stdlib-only loader, no third-party
  dependency. `.env.example` ships as a template.
- `just` automatically loads `.env` for every recipe (`set dotenv-load`),
  so `just e2e` etc. work without an explicit `JOPLIN_TOKEN=…` prefix.

- One-shot CLI mode on the same binary: `joplin-mcp tools` lists every
  registered tool; `joplin-mcp call <tool> [--json '{...}']` invokes any
  tool through the same in-process tool handlers used by the stdio server
  and prints the structured response as JSON. CLI mode auto-suppresses
  INFO logs unless `JOPLIN_LOG_LEVEL=debug` is set.
- `JOPLIN_MAX_RESOURCE_BYTES` env var (default 10 MiB) capping
  `download_resource` and `upload_resource` payload size; refusal returns
  an explicit error instead of OOM-ing the client.
- README configuration snippets for Cursor, Continue, and Cline.
- Pre-commit hook (`.githooks/pre-commit`) that runs `gofmt` and
  `go vet` on staged `.go` files. Enable with
  `git config core.hooksPath .githooks`.
- Commit-msg hook (`.githooks/commit-msg`) that enforces Conventional
  Commits locally. Pure bash, zero deps.
- Commitlint CI workflow (`.github/workflows/commitlint.yaml`) that runs
  on every PR and blocks merge on malformed commit subjects.
- OpenSpec base specs in `openspec/specs/` for `joplin-rest-client`,
  `mcp-tools`, and `encryption-transparency` so future changes can
  diff against a stated baseline.
- README "upgrade" note explaining that `go install ...@latest` re-fetches
  the latest tag.
- End-to-end test suite under `e2e/` exercising full CRUD against a real
  Joplin Desktop. Self-cleaning (everything created uses a
  `joplin-mcp-e2e-` prefix and is registered with `t.Cleanup`). Run with
  `just e2e`. Requires `JOPLIN_E2E=1` and `JOPLIN_TOKEN`.

### Changed

- Branch protection enabled on `main`: required status checks
  (Build & Test, SAST, SCA), no force-pushes, no deletions, linear
  history required, conversation resolution required.
- `joplin.Client` now decodes both JSON booleans and SQLite-style integer
  booleans (0/1) for every documented bool field via a new `Boolish`
  helper type. Joplin's REST API returns integer booleans for fields like
  `is_todo`, `encryption_applied`, `is_shared` — the prior bool typing
  failed to deserialise any non-empty list.
- Justfile rewritten with grouped recipes (build/test/lint/security/
  modules/release/meta), ANSI-coloured output, and a `just check`
  pre-push pipeline.
- CLAUDE.md now spells out Conventional Commits as a hard requirement
  with examples and lists both enforcement points; adds a manual SemVer
  bump rule.
- CONTRIBUTING.md mentions both git hooks and removes a false claim that
  a `JOPLIN_E2E=1` test exists (now noted as roadmap).

### Fixed

- `release.yaml`: pin syft's `install.sh` to the v1.42.3 commit SHA
  instead of fetching from `main` on every release. Closes the only
  unpinned external dependency in the release pipeline.
- `EventsPage.Cursor` and `ListChangesSinceArgs.Since` are now `string`,
  not `int64`. Joplin returns the events cursor as a quoted JSON string,
  which would crash `list_changes_since` with a JSON unmarshal error.
- Bool fields on Note, Folder, Tag, Resource, Revision now decode from
  both JSON booleans and SQLite-style integer booleans (0/1) via a new
  `joplin.Boolish` type. The prior strict-bool typing crashed
  `list_notes` against any non-empty notebook.
- `get_note_with_context` now returns `tags: []` and `resources: []`
  instead of `null` when a note has neither — avoids ambiguity in the
  LLM's view of "missing" vs "empty".
- `GetFolder`, `ListFolders`, `GetTag`, `ListTags`, `GetResource`,
  `ListResources`, `GetRevision`, `ListRevisions` now request explicit
  `?fields=` selectors. Joplin's defaults strip key fields like
  `mime`, `size`, and `encryption_applied`, which silently violated the
  encryption-transparency spec ("every item carries `encryption_applied`").

[Unreleased]: https://github.com/thereisnotime/joplin-mcp/compare/v0.1.0...HEAD

---

## [0.1.0] - 2026-05-09

Initial public release.

### Added

- MCP server over stdio built on the official Go SDK v1.6, exposing 28 tools
  across notes, folders, tags, search, resources, change events, and note
  revisions.
- Joplin REST client (`internal/joplin`) covering every documented Web Clipper
  endpoint with full type coverage, including the `encryption_applied` and
  `master_key_id` fields the prior-art Python MCP server silently dropped.
- Encryption transparency: every tool response surfaces the encryption state of
  every item it touches; list and search responses include
  `encrypted_items_skipped` so the LLM can tell the user when items came back
  as ciphertext; `download_resource` refuses to return ciphertext bytes
  silently, returning an explicit error instead.
- `get_note_with_context` convenience tool: returns a note together with its
  tags and resources via three parallel API calls.
- Bounded request lifetime: every client method takes a `context.Context`,
  default 10 s HTTP timeout, context cancellation propagates to in-flight
  requests.
- Env-only configuration: `JOPLIN_TOKEN` (required), `JOPLIN_BASE_URL`,
  `JOPLIN_TIMEOUT`, `JOPLIN_LOG_LEVEL`. Multi-profile is supported by
  registering multiple MCP server entries with different base URLs.
- Logging via `log/slog` to stderr — stdout is reserved for the MCP transport.
- `--version`, `-v`, `version`, `--help`, `-h`, `help` flags.
- GoReleaser pipeline: binaries for Linux/macOS/Windows on amd64 and arm64,
  cosign keyless signing, SBOM generation via syft, build-provenance
  attestation.
- CI: build, vet, gofmt check, test with race detector and shuffle, 80%
  coverage gate (measured on `internal/...`), Codecov upload, gosec SAST,
  govulncheck + Trivy SCA, OpenSSF Scorecard.
- Contributor docs: README with badges, CONTRIBUTING, SECURITY,
  CODE_OF_CONDUCT, CHANGELOG, CLAUDE.md, GitHub issue and PR templates,
  Conventional Commits enforced via commitlint, dependabot for `gomod` and
  `github-actions`, `CODEOWNERS`.
- OpenSpec change `joplin-mcp-v1` documenting the joplin-rest-client,
  mcp-tools, and encryption-transparency capabilities.

[0.1.0]: https://github.com/thereisnotime/joplin-mcp/releases/tag/v0.1.0
