# joplin-mcp

## GitHub Actions

Always pin actions to a full commit SHA, never use a tag or branch reference alone.
Include the version as a comment for readability:

```yaml
uses: actions/checkout@<full-sha>  # v6.0.2
```

This applies to all actions added or updated — including new ones introduced during fixes
or features.

## Coding standards

- Format with `gofmt` / `goimports` before committing.
- Keep `cmd/` thin — wiring only, no business logic.
- Error messages are lowercase and do not end with punctuation
  (e.g. `"note not found"`, not `"Note not found."`).
- No `Co-Authored-By` trailers in commits.
- Conventional commits: `feat:`, `fix:`, `docs:`, `chore:`, etc. Enforced by commitlint.

## Comments

Default to writing no comments. Add one only when the **why** is non-obvious:
a decision someone would otherwise question, a workaround for upstream
behaviour, or an invariant the code does not express. Keep them short and
natural — one line is usually enough. Don't write docstrings on every
exported symbol just because Go convention suggests it; well-named
identifiers carry their own meaning.

## MCP transport hygiene

stdout is reserved for MCP protocol frames. Never write to stdout from anywhere except
the MCP SDK transport. All logs go to stderr via `log/slog`.

## Encryption transparency

Never silently return an empty body from a tool just because Joplin returned ciphertext.
Every response includes `encryption_applied`; list/search responses include
`encrypted_items_skipped`.
