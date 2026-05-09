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

## Commits

Commit subjects MUST follow [Conventional Commits](https://www.conventionalcommits.org/).

- **Format:** `type(optional-scope)!?: subject`
- **Allowed types:** `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`,
  `build`, `ci`, `chore`, `revert`
- **Subject:** lowercase, no trailing punctuation, max 100 characters
- **Breaking change:** add `!` after the type/scope (e.g. `feat(api)!: drop legacy field`)
  and explain in the body

Examples:

```
feat(joplin): add list_changes_since tool
fix(tools): respect context cancellation in download_resource
docs: clarify encryption transparency in README
```

Enforcement points (all active):

- Local: `.githooks/commit-msg` blocks malformed messages on commit.
  Enable per-clone with `git config core.hooksPath .githooks`.
- CI: `.github/workflows/commitlint.yaml` runs on every PR via
  `wagoid/commitlint-github-action` against `.commitlintrc.yaml` —
  blocks merge if any commit fails.

## Versioning

The project follows [SemVer](https://semver.org/). Tags look like `vMAJOR.MINOR.PATCH`.

- `feat:` commits since the last tag → next release bumps **MINOR**.
- `fix:` commits → next release bumps **PATCH**.
- Any `!` breaking change → next release bumps **MAJOR**.

This is a manual, human-driven process — no `release-please` etc. wired up. When
cutting a release, audit the commits since the last tag and pick the bump that
honours these rules.

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
