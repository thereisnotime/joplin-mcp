# Proposal: joplin-mcp v1

## Why

Joplin users who want to expose their notes to LLM assistants (Claude Desktop, Cursor,
IDE agents) currently have one realistic option: a third-party Python MCP server that
ships with several real bugs (silently dropped fields, no HTTP timeout, blocking I/O
inside async handlers, broken wheel packaging) and exposes only a small slice of
Joplin's REST API (notes only — no folders, tags, resources, search syntax, or change
events). There is no first-class, statically-linked, easily-distributed MCP server for
Joplin, and no existing server is honest about Joplin's end-to-end encryption — items
that are still ciphertext on the local device are silently returned as empty bodies.

## What Changes

- Add a Go MCP server that wraps Joplin's local Web Clipper REST API end-to-end.
- Cover ~25 tools: notes, folders, tags, search, resources, events, revisions.
- Surface encryption state (`encryption_applied`, `master_key_id`) on every response.
- Distribute as a single static binary via GitHub Releases and `go install`.
- Use the official Go MCP SDK v1.6 (Anthropic + Google co-maintained) over stdio.
- Configure exclusively via environment variables; multi-profile = multiple MCP entries.
- Explicitly omit arbitrary local file reads (no `import_markdown`-style tool) — that is
  a prompt-injection footgun.

## Capabilities

### New Capabilities

- `joplin-rest-client`: Internal Go client for Joplin's Web Clipper REST API. Owns
  HTTP, auth, pagination, error mapping, and full type coverage.
- `mcp-tools`: MCP tool handlers exposed to clients over stdio. One handler per
  user-visible operation; validation, encryption annotation, output shaping.
- `encryption-transparency`: Cross-cutting behaviour that ensures every tool response
  surfaces the encryption state of every item it touches.

### Modified Capabilities

(none — greenfield project)

## Impact

- New repository: `github.com/thereisnotime/joplin-mcp`.
- Public Go module: anyone can `go install ...@latest`.
- GitHub Releases will publish signed binaries (cosign keyless) and SBOMs (syft).
- License: AGPL-3.0 — networked-deployment forks must publish source.
- Requires the user to enable Joplin Desktop's Web Clipper service and obtain a token.
- No changes to Joplin itself, no changes to the user's notebook data structure.
