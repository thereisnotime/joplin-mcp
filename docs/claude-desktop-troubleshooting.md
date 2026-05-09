# Troubleshooting Claude Desktop ↔ joplin-mcp

This is the page to land on when something doesn't work. Most issues are
configuration, not the binary.

## "The server failed to start" / nothing happens

1. **Confirm the binary exists at the path in your config.**
   ```sh
   /path/in/your/mcp.json --version
   ```
   If you used `go install`, the path is `$(go env GOPATH)/bin/joplin-mcp`.
   Claude Desktop does **not** expand `~` or `$HOME` — use absolute paths.

2. **Check stderr in Claude Desktop's MCP logs.** macOS:
   `~/Library/Logs/Claude/mcp*.log`. Look for lines prefixed `joplin-mcp:`.

## "cannot reach Joplin at http://localhost:41184"

The binary itself is running fine. Joplin Desktop's Web Clipper service is
not reachable. Fix:

1. **Open Joplin Desktop.** It must be running for the API to work.
2. **Tools → Options → Web Clipper → Enable Web Clipper Service.**
3. **Copy the API token** shown in that screen and put it in
   `JOPLIN_TOKEN` (env var or `.env` file).

If Joplin runs on a different port (rare), set `JOPLIN_BASE_URL`.

## "JOPLIN_TOKEN is required"

The token isn't reaching the binary. Three places it can come from, in
priority order:

1. Shell-set env var (`export JOPLIN_TOKEN=...`)
2. `env:` block in your MCP client's config
3. `.env` file in the binary's working directory

Claude Desktop's spawned subprocess doesn't inherit your shell env, so use
option 2 or 3.

## "tool error: http 401" or "http 403"

Token is wrong. Regenerate via Joplin Desktop → Tools → Options → Web
Clipper → "Renew token", and update your config.

## "search returns no results for a note I just created"

Joplin's full-text index updates a few seconds after a write. Set
`wait_for_index: true` in the search call:

```json
{"query": "my new note title", "wait_for_index": true}
```

The server will retry briefly until the index catches up (~7s ceiling).

## "encryption_applied is true and body is empty"

The note is still encrypted on the local Joplin device. Either:

1. The master key isn't unlocked — open Joplin and re-enter your password
   (Tools → Encryption Settings).
2. The note was encrypted with a master key this profile doesn't have.

joplin-mcp **does not decrypt anything**. It only surfaces what Joplin
Desktop has already decrypted into its local DB.

## "download_resource: resource size N bytes exceeds configured limit"

Default cap is 50 MiB. Bump it via the env var:

```sh
JOPLIN_MAX_RESOURCE_BYTES=104857600  # 100 MiB
```

Beware: very large blobs may exhaust your MCP client's per-message budget.

## "Multiple Joplin profiles" / Joplin Cloud / Joplin Server

joplin-mcp talks only to Joplin Desktop's local Web Clipper service. To
connect to multiple profiles, register multiple MCP server entries with
different `JOPLIN_BASE_URL` ports.

Joplin Cloud and Joplin Server are not supported — they don't expose the
Web Clipper REST API.

## How to verify the binary works without Claude Desktop

```sh
JOPLIN_TOKEN=... joplin-mcp tools          # list registered tools
JOPLIN_TOKEN=... joplin-mcp call list_folders
```

If those work, the binary is fine and the problem is in the MCP client
configuration.

## Filing a bug

If you've ruled out the above, please open an issue with:

- `joplin-mcp --version` output
- Joplin Desktop version (Help → About)
- MCP client name and version (Claude Desktop x.y.z, Cursor, etc.)
- Operating system
- The exact tool call that failed and the exact error
- Logs from your MCP client's MCP-server log directory

[Open an issue](https://github.com/thereisnotime/joplin-mcp/issues/new/choose)
