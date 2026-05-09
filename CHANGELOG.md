# Changelog

All notable changes to joplin-mcp are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the
project uses [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

### Added

- Initial repository scaffolding: AGPL-3.0 license, Go module, project layout,
  justfile, OpenSpec change `joplin-mcp-v1`.
- CI/CD scaffolding: GitHub Actions for build/test/coverage, gosec SAST,
  govulncheck and Trivy SCA, OpenSSF Scorecard, GoReleaser-driven release pipeline
  with cosign keyless signing and syft SBOMs.
- Contributor docs: README with badges, CONTRIBUTING, SECURITY, CODE_OF_CONDUCT,
  CHANGELOG.
- Issue and pull request templates.
- Conventional Commits enforcement via commitlint.

[Unreleased]: https://github.com/thereisnotime/joplin-mcp/compare/main...HEAD
