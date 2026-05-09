module := "github.com/thereisnotime/joplin-mcp"
binary := "joplin-mcp"
version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
commit  := `git rev-parse --short HEAD 2>/dev/null || echo "none"`
date    := `date -u +%Y-%m-%dT%H:%M:%SZ`
ldflags := "-s -w -X " + module + "/internal/version.Version=" + version + " -X " + module + "/internal/version.Commit=" + commit + " -X " + module + "/internal/version.Date=" + date

default:
    @just --list

build:
    mkdir -p bin
    go build -ldflags '{{ldflags}}' -o bin/{{binary}} ./cmd/{{binary}}

build-all:
    mkdir -p bin
    GOOS=linux   GOARCH=amd64 go build -ldflags '{{ldflags}}' -o bin/{{binary}}-linux-amd64   ./cmd/{{binary}}
    GOOS=linux   GOARCH=arm64 go build -ldflags '{{ldflags}}' -o bin/{{binary}}-linux-arm64   ./cmd/{{binary}}
    GOOS=darwin  GOARCH=amd64 go build -ldflags '{{ldflags}}' -o bin/{{binary}}-darwin-amd64  ./cmd/{{binary}}
    GOOS=darwin  GOARCH=arm64 go build -ldflags '{{ldflags}}' -o bin/{{binary}}-darwin-arm64  ./cmd/{{binary}}
    GOOS=windows GOARCH=amd64 go build -ldflags '{{ldflags}}' -o bin/{{binary}}-windows-amd64.exe ./cmd/{{binary}}

test:
    go test ./... -race -shuffle=on

cover:
    go test -race -shuffle=on -coverprofile=coverage.out -covermode=atomic ./...
    go tool cover -func=coverage.out | tail -1

cover-html:
    go test -race -shuffle=on -coverprofile=coverage.out -covermode=atomic ./...
    go tool cover -html=coverage.out -o coverage.html

lint:
    golangci-lint run ./...

vet:
    go vet ./...

tidy:
    go mod tidy

clean:
    rm -rf bin/ dist/ coverage.out coverage.html

install:
    go install -ldflags '{{ldflags}}' ./cmd/{{binary}}
