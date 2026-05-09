package tools

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

type SearchArgs struct {
	Query        string `json:"query" jsonschema:"Joplin search query — supports filters like tag:work notebook:Inbox created:day-7 body:foo"`
	Page         int    `json:"page,omitempty"`
	Limit        int    `json:"limit,omitempty"`
	WaitForIndex bool   `json:"wait_for_index,omitempty" jsonschema:"set true when searching for a note created in the last few seconds; the search will retry briefly to wait for Joplin's full-text index to catch up"`
}

// Backoff schedule for wait_for_index. Joplin's FT index typically settles
// within 2-3 seconds; six tries with this schedule cover ~7s total.
var searchRetryBackoff = []time.Duration{500 * time.Millisecond, 750 * time.Millisecond, 1 * time.Second, 1500 * time.Millisecond, 1500 * time.Millisecond, 2 * time.Second}

func registerSearchTools(srv *mcp.Server, c *joplin.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "search",
		Description: "Full-text search using Joplin's query syntax. Note: Joplin's FT index updates a few seconds after a note is created or modified — set wait_for_index=true if searching for something you just wrote.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, PageOut[NoteOut], error) {
		p, err := c.SearchNotes(ctx, args.Query, joplin.ListOptions{Page: args.Page, Limit: args.Limit})
		if err != nil {
			return nil, PageOut[NoteOut]{}, err
		}
		if args.WaitForIndex && len(p.Items) == 0 {
			for _, wait := range searchRetryBackoff {
				select {
				case <-ctx.Done():
					return nil, PageOut[NoteOut]{}, ctx.Err()
				case <-time.After(wait):
				}
				p, err = c.SearchNotes(ctx, args.Query, joplin.ListOptions{Page: args.Page, Limit: args.Limit})
				if err != nil {
					return nil, PageOut[NoteOut]{}, err
				}
				if len(p.Items) > 0 {
					break
				}
			}
		}
		return nil, notesPage(p), nil
	})
}
