package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

type ListTrashArgs struct {
	Page  int `json:"page,omitempty"`
	Limit int `json:"limit,omitempty"`
}

type RestoreArgs struct {
	NoteID string `json:"note_id"`
}

type EmptyTrashOut struct {
	PurgedCount int      `json:"purged_count"`
	Failed      []string `json:"failed,omitempty" jsonschema:"IDs that could not be permanently deleted"`
}

func registerTrashTools(srv *mcp.Server, c *joplin.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_trash",
		Description: "List notes currently in the trash (deleted_time is set but the row is still recoverable). Paginated.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ListTrashArgs) (*mcp.CallToolResult, PageOut[NoteOut], error) {
		p, err := c.ListTrashedNotes(ctx, joplin.ListOptions{Page: args.Page, Limit: args.Limit})
		if err != nil {
			return nil, PageOut[NoteOut]{}, err
		}
		return nil, notesPage(p), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "restore_note_from_trash",
		Description: "Move a trashed note back to its original folder by clearing deleted_time. The note's parent_id is preserved.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args RestoreArgs) (*mcp.CallToolResult, NoteOut, error) {
		n, err := c.RestoreNote(ctx, args.NoteID)
		if err != nil {
			return nil, NoteOut{}, err
		}
		return nil, noteOut(n), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "empty_trash",
		Description: "Permanently delete every note currently in the trash. Irreversible. Returns the number purged and any IDs that could not be deleted.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, EmptyTrashOut, error) {
		// Walk every page; trash can be larger than one page.
		all, err := joplin.CollectAll(ctx, func(ctx context.Context, page int) (joplin.Page[joplin.Note], error) {
			return c.ListTrashedNotes(ctx, joplin.ListOptions{Page: page, Limit: 100})
		})
		if err != nil {
			return nil, EmptyTrashOut{}, err
		}
		out := EmptyTrashOut{}
		for _, n := range all {
			if err := c.DeleteNote(ctx, n.ID, true); err != nil {
				out.Failed = append(out.Failed, n.ID)
				continue
			}
			out.PurgedCount++
		}
		return nil, out, nil
	})
}
