# joplin-mcp justfile
# Run `just` to see all available recipes.

set shell := ["bash", "-uc"]
set dotenv-load := true        # auto-load .env into every recipe
set dotenv-required := false   # but don't fail if it's missing

module := "github.com/thereisnotime/joplin-mcp"
binary := "joplin-mcp"
version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
commit  := `git rev-parse --short HEAD 2>/dev/null || echo "none"`
date    := `date -u +%Y-%m-%dT%H:%M:%SZ`
ldflags := "-s -w -X " + module + "/internal/version.Version=" + version + " -X " + module + "/internal/version.Commit=" + commit + " -X " + module + "/internal/version.Date=" + date

# ANSI colours for nicer output
bold   := '\033[1m'
red    := '\033[0;31m'
green  := '\033[0;32m'
yellow := '\033[0;33m'
blue   := '\033[0;34m'
reset  := '\033[0m'

# Show available recipes (the default).
default:
    @just --list --unsorted

# ── build ────────────────────────────────────────────────────────────

# Build the binary into bin/joplin-mcp.
build:
    @printf "{{blue}}> build{{reset}}\n"
    @mkdir -p bin
    go build -ldflags '{{ldflags}}' -o bin/{{binary}} ./cmd/{{binary}}
    @printf "{{green}}✓ bin/{{binary}}{{reset}} ({{version}})\n"

# Cross-compile for linux/darwin/windows × amd64/arm64.
build-all:
    @printf "{{blue}}> build-all{{reset}}\n"
    @mkdir -p bin
    GOOS=linux   GOARCH=amd64 go build -ldflags '{{ldflags}}' -o bin/{{binary}}-linux-amd64       ./cmd/{{binary}}
    GOOS=linux   GOARCH=arm64 go build -ldflags '{{ldflags}}' -o bin/{{binary}}-linux-arm64       ./cmd/{{binary}}
    GOOS=darwin  GOARCH=amd64 go build -ldflags '{{ldflags}}' -o bin/{{binary}}-darwin-amd64      ./cmd/{{binary}}
    GOOS=darwin  GOARCH=arm64 go build -ldflags '{{ldflags}}' -o bin/{{binary}}-darwin-arm64      ./cmd/{{binary}}
    GOOS=windows GOARCH=amd64 go build -ldflags '{{ldflags}}' -o bin/{{binary}}-windows-amd64.exe ./cmd/{{binary}}
    GOOS=windows GOARCH=arm64 go build -ldflags '{{ldflags}}' -o bin/{{binary}}-windows-arm64.exe ./cmd/{{binary}}
    @printf "{{green}}✓ cross-compiled to bin/{{reset}}\n"

# `go install` the binary into $GOPATH/bin with version ldflags.
install:
    @printf "{{blue}}> install{{reset}}\n"
    go install -ldflags '{{ldflags}}' ./cmd/{{binary}}
    @printf "{{green}}✓ installed to $(go env GOPATH)/bin/{{binary}}{{reset}}\n"

# Run the binary directly (requires JOPLIN_TOKEN in env).
run *ARGS:
    go run ./cmd/{{binary}} {{ARGS}}

# Print the version that `build` would stamp.
version:
    @echo "{{version}} ({{commit}}, {{date}})"

# ── test ─────────────────────────────────────────────────────────────

# Unit tests with race detector.
test:
    @printf "{{blue}}> test{{reset}}\n"
    go test ./... -race -shuffle=on
    @printf "{{green}}✓ tests passed{{reset}}\n"

# Coverage report (text).
cover:
    @printf "{{blue}}> cover{{reset}}\n"
    go test -race -shuffle=on -coverprofile=coverage.out -covermode=atomic -coverpkg=./internal/... ./...
    @go tool cover -func=coverage.out | tail -1

# Coverage report (HTML, opens coverage.html).
cover-html:
    @printf "{{blue}}> cover-html{{reset}}\n"
    go test -race -shuffle=on -coverprofile=coverage.out -covermode=atomic -coverpkg=./internal/... ./...
    go tool cover -html=coverage.out -o coverage.html
    @printf "{{green}}✓ wrote coverage.html{{reset}}\n"

# Run end-to-end tests against a real Joplin Desktop. Requires JOPLIN_TOKEN.
e2e:
    @printf "{{blue}}> e2e (real Joplin){{reset}}\n"
    JOPLIN_E2E=1 go test -count=1 -v ./e2e/...

# Run benchmarks (none yet, placeholder).
bench:
    @printf "{{blue}}> bench{{reset}}\n"
    go test -bench=. -benchmem -run=^$ ./...

# Drop test cache, useful when chasing flakes.
clean-cache:
    @printf "{{yellow}}> clean test cache{{reset}}\n"
    go clean -testcache

# ── lint ─────────────────────────────────────────────────────────────

# Run gofmt -l; fail if anything would change.
fmt-check:
    @printf "{{blue}}> fmt-check{{reset}}\n"
    @bad=$(gofmt -s -l .); \
    if [ -n "$bad" ]; then \
      printf "{{red}}gofmt would change:{{reset}}\n%s\n" "$bad"; exit 1; \
    fi
    @printf "{{green}}✓ gofmt clean{{reset}}\n"

# Format all Go files in place.
fmt:
    @printf "{{blue}}> fmt{{reset}}\n"
    gofmt -s -w .
    @printf "{{green}}✓ formatted{{reset}}\n"

# Run go vet.
vet:
    @printf "{{blue}}> vet{{reset}}\n"
    go vet ./...
    @printf "{{green}}✓ vet clean{{reset}}\n"

# Run golangci-lint (must be installed).
lint:
    @printf "{{blue}}> lint{{reset}}\n"
    golangci-lint run ./...
    @printf "{{green}}✓ lint clean{{reset}}\n"

# ── security ─────────────────────────────────────────────────────────

# gosec SAST scan.
sast:
    @printf "{{blue}}> gosec{{reset}}\n"
    @command -v gosec >/dev/null || go install github.com/securego/gosec/v2/cmd/gosec@v2.25.0
    gosec ./...

# govulncheck — Go module / stdlib vulnerability scan.
vulncheck:
    @printf "{{blue}}> govulncheck{{reset}}\n"
    @command -v govulncheck >/dev/null || go install golang.org/x/vuln/cmd/govulncheck@v1.1.4
    govulncheck ./...

# Run all security scanners.
audit: sast vulncheck

# ── modules ──────────────────────────────────────────────────────────

# go mod tidy.
tidy:
    @printf "{{blue}}> tidy{{reset}}\n"
    go mod tidy
    @printf "{{green}}✓ go.mod tidy{{reset}}\n"

# List direct dependencies.
deps:
    @go list -m -f '{{{{.Path}}}} {{{{.Version}}}}' all | grep -v "^{{module}}" | head -30

# Bump every direct dependency to its latest minor / patch.
update-deps:
    @printf "{{blue}}> update-deps{{reset}}\n"
    go get -u ./...
    go mod tidy

# ── release ──────────────────────────────────────────────────────────

# Dry-run goreleaser locally (no upload, no signing).
release-check:
    @printf "{{blue}}> goreleaser check{{reset}}\n"
    goreleaser release --snapshot --clean --skip=publish,sign

# ── meta ─────────────────────────────────────────────────────────────

# Run everything CI runs, locally. Use before pushing.
check: fmt-check vet test
    @printf "{{green}}✓ check passed{{reset}}\n"

# Wipe build artefacts and coverage.
clean:
    @printf "{{yellow}}> clean{{reset}}\n"
    rm -rf bin/ dist/ coverage.out coverage.html
    @printf "{{green}}✓ cleaned{{reset}}\n"
