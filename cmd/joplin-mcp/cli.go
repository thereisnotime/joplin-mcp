package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
	"github.com/thereisnotime/joplin-mcp/internal/tools"
	"github.com/thereisnotime/joplin-mcp/internal/version"
)

// CLI shares the tool implementations with the stdio server by spinning up the
// same mcp.Server in-process and calling it through an in-memory transport.
// Same code path, no drift.
func runCLI(ctx context.Context, c *joplin.Client, maxResource int64, sub string, args []string) error {
	srv := tools.New(c, tools.Options{Version: version.Version, MaxResourceBytes: maxResource})
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := srv.Connect(ctx, serverTransport, nil)
	if err != nil {
		return fmt.Errorf("server connect: %w", err)
	}
	defer serverSession.Wait()

	client := mcp.NewClient(&mcp.Implementation{Name: "joplin-mcp-cli", Version: version.Version}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		return fmt.Errorf("client connect: %w", err)
	}
	defer clientSession.Close()

	switch sub {
	case "tools":
		return cliListTools(ctx, clientSession)
	case "call":
		return cliCall(ctx, clientSession, args)
	default:
		return fmt.Errorf("unknown subcommand %q (try 'tools' or 'call')", sub)
	}
}

func cliListTools(ctx context.Context, cs *mcp.ClientSession) error {
	res, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return err
	}
	for _, t := range res.Tools {
		fmt.Printf("%-26s  %s\n", t.Name, t.Description)
	}
	return nil
}

func cliCall(ctx context.Context, cs *mcp.ClientSession, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: joplin-mcp call <tool_name> [--json '{...}']")
	}
	tool := args[0]

	payload := "{}"
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			if i+1 >= len(args) {
				return errors.New("--json requires a value")
			}
			payload = args[i+1]
			i++
		case "--help", "-h":
			fmt.Println("usage: joplin-mcp call <tool_name> [--json '{...}']")
			fmt.Println("example: joplin-mcp call list_notes --json '{\"limit\":5}'")
			return nil
		default:
			return fmt.Errorf("unknown flag %q (try --json '{...}')", args[i])
		}
	}

	var arguments map[string]any
	if err := json.Unmarshal([]byte(payload), &arguments); err != nil {
		return fmt.Errorf("invalid --json payload: %w", err)
	}

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: tool, Arguments: arguments})
	if err != nil {
		return err
	}

	if res.IsError {
		for _, c := range res.Content {
			if tc, ok := c.(*mcp.TextContent); ok {
				fmt.Fprintln(os.Stderr, tc.Text)
			}
		}
		return errors.New("tool error")
	}

	if res.StructuredContent != nil {
		b, err := json.MarshalIndent(res.StructuredContent, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			fmt.Println(tc.Text)
		}
	}
	return nil
}
