# Changelog

All notable changes to joplin-mcp are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the
project uses [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

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
