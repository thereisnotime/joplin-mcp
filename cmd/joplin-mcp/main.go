// Command joplin-mcp is a Model Context Protocol server for Joplin notes.
//
// It wraps Joplin Desktop's local Web Clipper REST API and exposes notes,
// folders, tags, search, resources, events, and revisions as MCP tools over
// stdio.
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

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "joplin-mcp:", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`joplin-mcp — Model Context Protocol server for Joplin

Usage:
  joplin-mcp           Run the MCP server on stdio
  joplin-mcp --version Print version and exit
  joplin-mcp --help    Print this help and exit

Environment:
  JOPLIN_TOKEN                (required) Joplin Web Clipper API token
  JOPLIN_BASE_URL             (default http://localhost:41184) Web Clipper base URL
  JOPLIN_TIMEOUT              (default 10s) per-request HTTP timeout (Go duration syntax)
  JOPLIN_LOG_LEVEL            (default info) debug | info | warn | error
  JOPLIN_MAX_RESOURCE_BYTES   (default 10485760) max bytes for download/upload_resource

The server speaks Model Context Protocol over stdio. Wire it up to an MCP
client (e.g. Claude Desktop) per its documentation.`)
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
	// stdout is reserved for the MCP transport; logs MUST go to stderr.
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

	srv := tools.New(client, tools.Options{Version: version.Version, MaxResourceBytes: maxResource})

	logger.Info("starting joplin-mcp",
		"version", version.Version,
		"base_url", coalesce(baseURL, joplin.DefaultBaseURL),
		"timeout", timeout,
		"max_resource_bytes", maxResource)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil {
		return err
	}
	return nil
}

func coalesce(a, b string) string {
	if a == "" {
		return b
	}
	return a
}
