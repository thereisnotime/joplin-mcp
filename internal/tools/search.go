package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

type SearchArgs struct {
	Query string `json:"query" jsonschema:"Joplin search query — supports filters like tag:work notebook:Inbox created:day-7 body:foo"`
	Page  int    `json:"page,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

func registerSearchTools(srv *mcp.Server, c *joplin.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "search",
		Description: "Full-text search using Joplin's query syntax. Returns matching notes, paginated.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, PageOut[NoteOut], error) {
		p, err := c.SearchNotes(ctx, args.Query, joplin.ListOptions{Page: args.Page, Limit: args.Limit})
		if err != nil {
			return nil, PageOut[NoteOut]{}, err
		}
		return nil, notesPage(p), nil
	})
}
