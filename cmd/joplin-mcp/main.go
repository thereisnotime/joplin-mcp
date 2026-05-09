// Command joplin-mcp is a Model Context Protocol server for Joplin notes.
//
// Two modes:
//   - default (no subcommand) — runs as an MCP server on stdio
//   - tools / call <tool> [--json '{...}'] — one-shot CLI dispatch through the
//     same tool handlers, useful for scripting and ad-hoc inspection
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
	"github.com/thereisnotime/joplin-mcp/internal/tools"
	"github.com/thereisnotime/joplin-mcp/internal/version"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Println(version.String())
			return
		case "--help", "-h", "help":
			printHelp()
			return
		}
	}

	// Best-effort .env load from cwd. Shell-set env vars always win.
	if err := loadDotEnv(".env"); err != nil {
		fmt.Fprintln(os.Stderr, "joplin-mcp: warning: could not read .env:", err)
	}

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "joplin-mcp:", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`joplin-mcp — Model Context Protocol server for Joplin

Usage:
  joplin-mcp                          Run as an MCP server on stdio (default)
  joplin-mcp tools                    List every tool the server exposes
  joplin-mcp call <tool> [--json X]   One-shot CLI: invoke a tool and print
                                      its structured response as JSON.
                                      X may be a JSON literal, '-' (stdin),
                                      or '@path' (read from file).
  joplin-mcp --version                Print version and exit
  joplin-mcp --help                   Print this help and exit

Environment:
  JOPLIN_TOKEN                (required) Joplin Web Clipper API token
  JOPLIN_BASE_URL             (default http://localhost:41184) Web Clipper base URL
  JOPLIN_TIMEOUT              (default 10s) per-request HTTP timeout (Go duration)
  JOPLIN_LOG_LEVEL            (default info) debug | info | warn | error
  JOPLIN_MAX_RESOURCE_BYTES   (default 10485760) max bytes for download/upload_resource

Examples:
  joplin-mcp tools
  joplin-mcp call list_folders
  joplin-mcp call list_notes --json '{"limit":5}'
  joplin-mcp call search --json '{"query":"tag:work","limit":10}'
  joplin-mcp call create_note --json @new-note.json
  echo '{"note_id":"abc"}' | joplin-mcp call get_note --json -

The default mode speaks Model Context Protocol over stdio. Wire it up to an
MCP client (e.g. Claude Desktop) per its documentation.`)
}

func run() error {
	logLevel := slog.LevelInfo
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("JOPLIN_LOG_LEVEL"))); v != "" {
		switch v {
		case "debug":
			logLevel = slog.LevelDebug
		case "info":
			logLevel = slog.LevelInfo
		case "warn", "warning":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		default:
			return fmt.Errorf("invalid JOPLIN_LOG_LEVEL %q", v)
		}
	}

	// Detect CLI mode (tools / call) up front so we can stay quiet — no
	// "starting" log line polluting CLI output.
	var subcommand string
	var subargs []string
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "tools", "call":
			subcommand = os.Args[1]
			subargs = os.Args[2:]
			// CLI mode: only log warnings/errors unless user opted into debug.
			if logLevel < slog.LevelWarn {
				logLevel = slog.LevelWarn
			}
		}
	}

	// stdout is reserved for the MCP transport in server mode and for the
	// command's own JSON output in CLI mode; logs always go to stderr.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	token := strings.TrimSpace(os.Getenv("JOPLIN_TOKEN"))
	if token == "" {
		return fmt.Errorf("JOPLIN_TOKEN is required (Tools → Options → Web Clipper in Joplin Desktop)")
	}

	baseURL := strings.TrimSpace(os.Getenv("JOPLIN_BASE_URL"))

	timeout := joplin.DefaultTimeout
	if v := strings.TrimSpace(os.Getenv("JOPLIN_TIMEOUT")); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("invalid JOPLIN_TIMEOUT %q: %w", v, err)
		}
		timeout = d
	}

	client, err := joplin.New(joplin.Options{
		Token:      token,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: timeout},
	})
	if err != nil {
		return err
	}

	maxResource := tools.DefaultMaxResourceBytes
	if v := strings.TrimSpace(os.Getenv("JOPLIN_MAX_RESOURCE_BYTES")); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n <= 0 {
			return fmt.Errorf("invalid JOPLIN_MAX_RESOURCE_BYTES %q: must be a positive integer", v)
		}
		maxResource = n
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if subcommand != "" {
		return runCLI(ctx, client, maxResource, subcommand, subargs)
	}

	srv := tools.New(client, tools.Options{Version: version.Version, MaxResourceBytes: maxResource})

	logger.Info("starting joplin-mcp",
		"version", version.Version,
		"base_url", coalesce(baseURL, joplin.DefaultBaseURL),
		"timeout", timeout,
		"max_resource_bytes", maxResource)

	return srv.Run(ctx, &mcp.StdioTransport{})
}

func coalesce(a, b string) string {
	if a == "" {
		return b
	}
	return a
}
