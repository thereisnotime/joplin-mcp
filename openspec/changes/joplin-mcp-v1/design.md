# Design: joplin-mcp v1

## Context

Joplin Desktop ships with a local "Web Clipper Service" — an HTTP REST API on
`127.0.0.1:41184` that any local process can call with a user-issued token. The API
exposes notes, folders, tags, resources, full-text search, change events, and
revisions. End-to-end encryption is handled by the desktop client itself: items
already decrypted in the local SQLite database are served as plaintext over the API,
items still encrypted are served with `encryption_applied=true` and an empty body.

The existing Python MCP server (`dweigend/joplin-mcp-server`) is the only published
prior art and has multiple defects: a `from_api_response` that silently drops half of
the fields it declares, no HTTP timeouts, blocking `requests` inside `async def`
tools, a broken wheel that does not include the server module, and no folder/tag/
resource/events/revisions surface at all. It also gives the LLM no signal when an
item is encrypted — it just returns an empty body.

Go's official MCP SDK reached v1.0 in 2025, is co-maintained with Google, and is now
at v1.6. Distribution as a single static binary removes an entire class of "Python
interpreter / venv / abs-path" installation issues that plague the current ecosystem.

## Goals / Non-Goals

**Goals:**

- Cover ~80% of Joplin's REST surface in the first release: all CRUD on notes/
  folders/tags, full search syntax, resources (incl. binary upload/download), the
  events change feed, and note revisions.
- Be honest about encryption: every tool response includes the encryption state of
  the items it touched; list/search responses report a count of items skipped because
  they were still ciphertext.
- Single static binary distribution, signed (cosign keyless) with SBOM (syft).
- Clean, idiomatic Go: stdlib HTTP, stdlib `log/slog`, no DI framework, no logger
  library, no third-party HTTP client.
- 80% test coverage gate enforced in CI.

**Non-Goals:**

- No master-key handling. The server never decrypts. Decryption is exclusively the
  desktop client's responsibility.
- No WebDAV / sync-target access. Joplin Desktop must be running.
- No multi-profile bridging in a single server instance. Each Joplin profile uses a
  different Web Clipper port; expose each as a separate MCP entry in the client
  config.
- No arbitrary local file reads. The Python server's `import_markdown` is intentionally
  not reimplemented.

## Decisions

- **Layered architecture, sshroute-style.** `cmd/joplin-mcp/` for the entry point,
  `internal/joplin/` for the REST client, `internal/tools/` for MCP handlers,
  `internal/version/` for ldflags-injected version info. `internal/` only — this is
  an application, not a library.
- **stdlib `net/http`.** Joplin's API is plain JSON, requests are small, no streaming
  needed. No external HTTP client. Every request takes a `context.Context` so MCP
  call cancellation propagates.
- **Token via `?token=` query param.** Joplin's canonical auth method. The Python
  server set both an `Authorization` header and the query param, redundantly; we
  only set what is needed.
- **Default 10-second HTTP timeout.** The Python server has no timeout; if Joplin
  hangs, the MCP process hangs forever. We set one.
- **Field coverage over ergonomics.** Every JSON field Joplin returns is a Go field
  on our types. Better to expose and ignore than to silently drop.
- **`get_note_with_context` convenience tool.** Three parallel API calls
  (`/notes/:id`, `/notes/:id/tags`, `/notes/:id/resources`) returning a merged
  response. Saves the LLM round trips on the most common workflow.
- **Encryption annotation is a wrapper, not per-handler logic.** A single helper in
  `internal/tools/encryption.go` augments every response with encryption state and
  counts skipped encrypted items, so handlers stay simple.
- **Env-only configuration.** `JOPLIN_TOKEN` (required), `JOPLIN_BASE_URL`,
  `JOPLIN_TIMEOUT`, `JOPLIN_LOG_LEVEL`. No config file. Multi-profile is solved by
  registering multiple MCP server entries with different `JOPLIN_BASE_URL` ports.
- **stderr for logs.** stdout belongs to the MCP transport and any pollution breaks
  the protocol. Structured logging via `log/slog`.
- **AGPL-3.0 license.** Matches the spirit of Joplin core (also AGPL) and ensures
  that any networked-deployment fork must publish source.

## Risks / Trade-offs

- **Breaking changes in the Joplin REST API** would require a release. Mitigation:
  pin no field names that aren't in Joplin's documented schema; keep the type layer
  permissive (extra fields ignored, missing fields zero-valued).
- **Breaking changes in the Go MCP SDK before our v1.0** are possible despite SDK
  v1 stability — the surface we use is small, so a bump should be a localized fix.
- **`get_note_with_context` makes 3 calls in parallel** which means one slow folder
  can serialise behind it. Acceptable trade-off vs. forcing the LLM to issue 3
  sequential round trips.
- **AGPL-3.0 may deter some adopters** who can't comply with networked-source
  obligations. This is intentional — the project is aimed at hobbyists and
  individuals, not vendor-sold SaaS.
- **No `import_markdown` tool** removes a feature the existing Python server had.
  Users who want to bulk-import a directory of Markdown can do so via Joplin's own
  UI; exposing arbitrary local file reads to an LLM is not worth the convenience.
- **Coverage gate of 80%** can become a drag on small fixes. Acceptable cost for the
  signal it provides early in the project's life.
