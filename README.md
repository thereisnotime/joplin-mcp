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

To upgrade later, re-run the same command — `@latest` always pulls the most
recent published tag.

## Configuration

joplin-mcp reads its configuration from environment variables.

| Variable | Required | Default | Description |
|---|---|---|---|
| `JOPLIN_TOKEN` | yes | | Joplin Web Clipper API token |
| `JOPLIN_BASE_URL` | no | `http://localhost:41184` | Joplin Web Clipper base URL |
| `JOPLIN_TIMEOUT` | no | `10s` | HTTP request timeout (Go duration syntax) |
| `JOPLIN_LOG_LEVEL` | no | `info` | `debug`, `info`, `warn`, or `error` |
| `JOPLIN_MAX_RESOURCE_BYTES` | no | `52428800` (50 MiB) | Cap on `download_resource` and `upload_resource` payload size |

To get a token: open Joplin Desktop → **Tools → Options → Web Clipper**, enable the
Web Clipper Service, copy the API token shown.

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or
`%APPDATA%\Claude\claude_desktop_config.json` (Windows). If you installed with
`go install`, the binary lives at `$(go env GOPATH)/bin/joplin-mcp` — typically
`~/go/bin/joplin-mcp` on Linux/macOS and `%USERPROFILE%\go\bin\joplin-mcp.exe`
on Windows. Claude Desktop does not expand `~` or `$HOME`, so use the absolute
path.

```json
{
  "mcpServers": {
    "joplin": {
      "command": "/home/YOU/go/bin/joplin-mcp",
      "env": {
        "JOPLIN_TOKEN": "your-token-here"
      }
    }
  }
}
```

macOS:

```json
{
  "mcpServers": {
    "joplin": {
      "command": "/Users/YOU/go/bin/joplin-mcp",
      "env": {
        "JOPLIN_TOKEN": "your-token-here"
      }
    }
  }
}
```

Windows:

```json
{
  "mcpServers": {
    "joplin": {
      "command": "C:\\Users\\YOU\\go\\bin\\joplin-mcp.exe",
      "env": {
        "JOPLIN_TOKEN": "your-token-here"
      }
    }
  }
}
```

For multiple Joplin profiles, register each as its own MCP server entry with a
different `JOPLIN_BASE_URL` port.

### Cursor

Add to `~/.cursor/mcp.json` (or your project's `.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "joplin": {
      "command": "/home/YOU/go/bin/joplin-mcp",
      "env": { "JOPLIN_TOKEN": "your-token-here" }
    }
  }
}
```

### Continue

Add to your `~/.continue/config.json` under `experimental.modelContextProtocolServers`:

```json
{
  "experimental": {
    "modelContextProtocolServers": [
      {
        "transport": {
          "type": "stdio",
          "command": "/home/YOU/go/bin/joplin-mcp",
          "env": { "JOPLIN_TOKEN": "your-token-here" }
        }
      }
    ]
  }
}
```

### Cline (VS Code extension)

Open Cline's MCP settings (Command Palette → "Cline: MCP Servers") and add:

```json
{
  "mcpServers": {
    "joplin": {
      "command": "/home/YOU/go/bin/joplin-mcp",
      "env": { "JOPLIN_TOKEN": "your-token-here" },
      "disabled": false,
      "autoApprove": []
    }
  }
}
```

## Tools

42 tools across ten groups.

### Notes

| Tool | Description |
|---|---|
| `list_notes` | List notes, paginated. Returns `encryption_applied` per item and an `encrypted_items_skipped` count. |
| `get_note` | Get a single note by ID. `encryption_applied` indicates whether the body could be returned. |
| `get_note_with_context` | Get a note plus its tags and attached resources, fetched in parallel — saves the LLM 2 round trips. |
| `create_note` | Create a note. |
| `update_note` | Partially update a note. Only fields that are set are sent to Joplin. |
| `delete_note` | Move to trash, or set `permanent=true` to bypass trash. |

### Folders

| Tool | Description |
|---|---|
| `list_folders` | List notebooks (folders), paginated. |
| `get_folder` | Get a single folder by ID. |
| `create_folder` | Create a folder. Set `parent_id` for nested folders. |
| `update_folder` | Partially update a folder. |
| `delete_folder` | Move to trash, or set `permanent=true` to bypass trash. |
| `list_notes_in_folder` | List notes whose `parent_id` is the given folder. |
| `set_folder_icon` | Set or clear a folder's sidebar icon emoji. (Also settable inline via the `emoji` arg on `create_folder` / `update_folder`.) |

### Tags

| Tool | Description |
|---|---|
| `list_tags` | List tags, paginated. |
| `get_tag` | Get a single tag by ID. |
| `create_tag` | Create a tag. |
| `update_tag` | Rename a tag in place. Existing attachments are preserved. |
| `delete_tag` | Delete a tag. |
| `tag_note` | Attach a tag to a note. |
| `untag_note` | Detach a tag from a note. |
| `list_notes_with_tag` | List notes that have the given tag. |

### Search

| Tool | Description |
|---|---|
| `search` | Full-text search using Joplin's query syntax (e.g. `tag:work notebook:Inbox created:day-7 body:foo`). Paginated. |

### Resources (attachments)

| Tool | Description |
|---|---|
| `list_resources` | List resources, paginated. |
| `get_resource_metadata` | Get metadata for a single resource (no bytes). |
| `download_resource` | Download a resource's bytes (base64-encoded). Refuses when the resource is still encrypted on the local device. |
| `upload_resource` | Upload a new resource. Provide bytes as base64. |
| `update_resource` | Rename a resource (title / filename). Bytes are immutable; replace via delete + upload. |
| `list_notes_using_resource` | List notes that reference a given resource. Useful for "where is this attachment used?". |
| `delete_resource` | Delete a resource. |

### Change events

| Tool | Description |
|---|---|
| `list_changes_since` | List Joplin change events with an ID greater than the supplied cursor. The response carries a new cursor for the next call. |

### Revisions

| Tool | Description |
|---|---|
| `list_note_revisions` | List the revision history for a specific note. |
| `get_revision` | Get a single revision by ID. |

### Links

| Tool | Description |
|---|---|
| `list_outbound_links` | List Joplin item IDs referenced from a note's body via `:/<id>` markdown links and image embeds. Set `resolve_titles=true` to also fetch each target's title. |
| `list_backlinks` | List notes whose body references the given note. Joplin doesn't expose backlinks natively; this is a search across all notes. |

### Attach

| Tool | Description |
|---|---|
| `attach_resource_to_note` | Upload a file as a Joplin resource AND insert a properly-formatted markdown reference into the note body in one call. Image MIME types use `![]()`; others use `[]()`. |

### Bulk

| Tool | Description |
|---|---|
| `bulk_tag_notes` | Attach the same tag to many notes in parallel. |
| `bulk_untag_notes` | Detach the same tag from many notes in parallel. |
| `bulk_move_notes` | Move many notes into the same folder in parallel. |
| `bulk_delete_notes` | Delete many notes (default trash, set `permanent=true` to bypass). |

### Trash

| Tool | Description |
|---|---|
| `list_trash` | List notes currently in the trash (`deleted_time` is set). |
| `restore_note_from_trash` | Move a trashed note back to its folder by clearing `deleted_time`. |
| `empty_trash` | Permanently delete every note currently in the trash. Irreversible. |

## CLI mode

The same binary doubles as a one-shot CLI for scripting and ad-hoc inspection.
The CLI runs the same tool handlers in-process via the SDK's in-memory
transport — same code path as the stdio server, no drift.

```sh
# List every tool the server exposes.
joplin-mcp tools

# Invoke a tool with no arguments.
joplin-mcp call list_folders

# Pass arguments as a JSON literal.
joplin-mcp call list_notes --json '{"limit":5}'
joplin-mcp call search    --json '{"query":"tag:work","limit":10}'
joplin-mcp call get_note  --json '{"note_id":"abc..."}'

# Read JSON from a file (curl-style, '@' prefix).
joplin-mcp call create_note --json @new-note.json

# Read JSON from stdin — best for multi-line markdown bodies that
# would otherwise be a nightmare to escape on the shell.
joplin-mcp call create_note --json - <<'EOF'
{
  "parent_id": "abc...",
  "title": "Today's notes",
  "body": "## Heading\n\n- item one\n- item two\n\n```go\nfunc main() {}\n```"
}
EOF
```

Output is structured JSON on stdout (so you can pipe into `jq`); errors and
logs go to stderr. CLI mode auto-suppresses INFO logs unless you set
`JOPLIN_LOG_LEVEL=debug`. The same env vars (`JOPLIN_TOKEN` etc.) apply.

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
