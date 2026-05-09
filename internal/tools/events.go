package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

type ListChangesSinceArgs struct {
	Since string `json:"since,omitempty" jsonschema:"opaque cursor returned from a previous call; empty returns events from the beginning of the available window"`
}

func registerEventTools(srv *mcp.Server, c *joplin.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_changes_since",
		Description: "List Joplin change events with an ID greater than the supplied cursor. The response carries a new cursor to pass on the next call.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ListChangesSinceArgs) (*mcp.CallToolResult, EventsOut, error) {
		p, err := c.ListEvents(ctx, args.Since)
		if err != nil {
			return nil, EventsOut{}, err
		}
		out := EventsOut{Cursor: p.Cursor, HasMore: p.HasMore, Items: make([]EventOut, 0, len(p.Items))}
		for _, e := range p.Items {
			out.Items = append(out.Items, eventOut(e))
		}
		return nil, out, nil
	})
}
