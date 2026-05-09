# joplin-mcp

[![CI](https://github.com/thereisnotime/joplin-mcp/actions/workflows/ci.yaml/badge.svg)](https://github.com/thereisnotime/joplin-mcp/actions/workflows/ci.yaml)
[![Release](https://github.com/thereisnotime/joplin-mcp/actions/workflows/release.yaml/badge.svg)](https://github.com/thereisnotime/joplin-mcp/actions/workflows/release.yaml)
[![Latest Release](https://img.shields.io/github/v/release/thereisnotime/joplin-mcp)](https://github.com/thereisnotime/joplin-mcp/releases/latest)
[![codecov](https://codecov.io/gh/thereisnotime/joplin-mcp/branch/main/graph/badge.svg)](https://codecov.io/gh/thereisnotime/joplin-mcp)
[![Go Report Card](https://goreportcard.com/badge/github.com/thereisnotime/joplin-mcp)](https://goreportcard.com/report/github.com/thereisnotime/joplin-mcp)
[![Go Reference](https://pkg.go.dev/badge/github.com/thereisnotime/joplin-mcp.svg)](https://pkg.go.dev/github.com/thereisnotime/joplin-mcp)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/thereisnotime/joplin-mcp/badge)](https://scorecard.dev/viewer/?uri=github.com/thereisnotime/joplin-mcp)
[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL%203.0-blue.svg)](LICENSE)

Model Context Protocol server for [Joplin](https://joplinapp.org/). Exposes notes,
folders, tags, search, resources, change events, and revisions to MCP clients
(Claude Desktop, Cursor, IDE agents) over stdio.

Built in Go on the official [Model Context Protocol Go SDK](https://github.com/modelcontextprotocol/go-sdk).
Single static binary. Honest about end-to-end encryption: items still encrypted on
the local Joplin device are surfaced as such, never silently returned as empty bodies.

> **Status:** v0.1 in active development. The first tagged release will land once
> the Joplin REST client (`internal/joplin`) and MCP tool surface (`internal/tools`)
> are complete. See [openspec/changes/joplin-mcp-v1/](openspec/changes/joplin-mcp-v1/)
> for the spec and implementation plan.

## How it works

```
┌─────────────────┐    stdio MCP    ┌──────────────┐    HTTP    ┌──────────────────┐
│  MCP client     │ ───────────────▶│  joplin-mcp  │ ──────────▶│  Joplin Desktop  │
│ (Claude, etc.)  │                 │              │  :41184    │  Web Clipper     │
└─────────────────┘                 └──────────────┘            └──────────────────┘
```

joplin-mcp talks only to the local Joplin Desktop API. It does not touch your sync
target, does not hold master keys, and does not send data anywhere else.

## Why?

The existing third-party Python MCP server has real defects (silently dropped fields,
no HTTP timeouts, blocking I/O in async handlers, broken wheel) and exposes only a
small slice of Joplin's REST API. This project rebuilds the same idea in Go, with:

- **Full API coverage** — notes, folders, tags, search, resources, events, revisions.
- **Encryption transparency** — every response includes `encryption_applied`; list
  responses include `encrypted_items_skipped`.
- **Single static binary** — `go install` or download from Releases. No `uv`/`venv`,
  no Python interpreter.
- **Real timeouts and context propagation** — the LLM cancelling a tool call
  actually cancels the HTTP request.

## Installation

### Binary download

Download the latest release from [GitHub Releases](https://github.com/thereisnotime/joplin-mcp/releases).
Binaries are available for Linux, macOS, and Windows on amd64 and arm64.

### Go install

```sh
go install github.com/thereisnotime/joplin-mcp/cmd/joplin-mcp@latest
```

## Configuration

joplin-mcp reads its configuration from environment variables.

| Variable | Required | Default | Description |
|---|---|---|---|
| `JOPLIN_TOKEN` | yes | | Joplin Web Clipper API token |
| `JOPLIN_BASE_URL` | no | `http://localhost:41184` | Joplin Web Clipper base URL |
| `JOPLIN_TIMEOUT` | no | `10s` | HTTP request timeout |
| `JOPLIN_LOG_LEVEL` | no | `info` | `debug`, `info`, `warn`, or `error` |

To get a token: open Joplin Desktop → **Tools → Options → Web Clipper**, enable the
Web Clipper Service, copy the API token shown.

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or
`%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "joplin": {
      "command": "/path/to/joplin-mcp",
      "env": {
        "JOPLIN_TOKEN": "your-token-here"
      }
    }
  }
}
```

For multiple Joplin profiles, register each as its own MCP server entry with a
different `JOPLIN_BASE_URL` port.

## Tools

Documented in [openspec/changes/joplin-mcp-v1/specs/mcp-tools/spec.md](openspec/changes/joplin-mcp-v1/specs/mcp-tools/spec.md).
A full table will appear here once v0.1 ships.

## Encryption

joplin-mcp does **not** decrypt anything — Joplin Desktop owns decryption. Items
already decrypted on the local device are returned as plaintext; items still
encrypted are returned with `encryption_applied: true` and a `master_key_id`.

In list and search responses, `encrypted_items_skipped` reports how many items
were returned in encrypted form, so the LLM can tell the user "5 of 12 notes are
still encrypted on your device — try unlocking Joplin and retrying."

The `download_resource` tool refuses to return ciphertext bytes silently; it
returns an explicit error if the resource is encrypted.

## Building from source

Requires [Go 1.26+](https://go.dev/dl/) and [just](https://just.systems/).

```sh
git clone git@github.com:thereisnotime/joplin-mcp.git
cd joplin-mcp

just build        # outputs bin/joplin-mcp
just build-all    # cross-compile linux/darwin/windows × amd64/arm64
just test         # run tests with race detector
just install      # go install with version ldflags injected
```

## Community

**Get the software** — download a pre-built binary from
[Releases](https://github.com/thereisnotime/joplin-mcp/releases), install with
`go install github.com/thereisnotime/joplin-mcp/cmd/joplin-mcp@latest`, or
[build from source](#building-from-source).

**Feedback and bug reports** — open an issue on
[GitHub Issues](https://github.com/thereisnotime/joplin-mcp/issues).

**Contributing** — see [CONTRIBUTING.md](CONTRIBUTING.md). Security vulnerabilities
should be reported privately via
[GitHub Security Advisories](https://github.com/thereisnotime/joplin-mcp/security/advisories/new).

## License

[AGPL-3.0](LICENSE)
