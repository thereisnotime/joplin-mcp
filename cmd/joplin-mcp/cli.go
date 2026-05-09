package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

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

// readJSONSource resolves a --json value into a JSON string:
//
//   - "-"        → read from stdin
//   - "@PATH"    → read from the file at PATH
//   - anything   → treated as the literal JSON
//
// Same convention as curl's --data flag, so users with multi-line markdown
// bodies don't have to fight shell escaping.
func readJSONSource(spec string) (string, error) {
	switch {
	case spec == "-":
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return string(b), nil
	case strings.HasPrefix(spec, "@"):
		path := spec[1:]
		b, err := os.ReadFile(path) // #nosec G304 -- path is user-supplied by design
		if err != nil {
			return "", fmt.Errorf("read %s: %w", path, err)
		}
		return string(b), nil
	default:
		return spec, nil
	}
}

const callHelp = `usage: joplin-mcp call <tool_name> [--json <SOURCE>]

--json accepts:
  '{...}'        a JSON literal
  -              read JSON from stdin
  @path/to/file  read JSON from a file (curl-style)

Examples:
  joplin-mcp call list_notes --json '{"limit":5}'
  joplin-mcp call create_note --json @note.json
  echo '{"note_id":"abc"}' | joplin-mcp call get_note --json -

Heredoc with multi-line markdown body:
  joplin-mcp call create_note --json - <<'EOF'
  {
    "title": "Today's notes",
    "parent_id": "...",
    "body": "## Heading\n\n- item one\n- item two"
  }
  EOF
`

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
		return errors.New("usage: joplin-mcp call <tool_name> [--json <SOURCE>]")
	}
	tool := args[0]

	payload := "{}"
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			if i+1 >= len(args) {
				return errors.New("--json requires a value")
			}
			raw, err := readJSONSource(args[i+1])
			if err != nil {
				return err
			}
			payload = raw
			i++
		case "--help", "-h":
			fmt.Print(callHelp)
			return nil
		default:
			return fmt.Errorf("unknown flag %q (try --json <SOURCE>)", args[i])
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
