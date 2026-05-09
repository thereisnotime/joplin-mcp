package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

type ListNoteRevisionsArgs struct {
	NoteID string `json:"note_id"`
}

type GetRevisionArgs struct {
	RevisionID string `json:"revision_id"`
}

type RevisionsOut struct {
	Items                 []RevisionOut `json:"items"`
	EncryptedItemsSkipped int           `json:"encrypted_items_skipped"`
}

func registerRevisionTools(srv *mcp.Server, c *joplin.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_note_revisions",
		Description: "List the revision history for a specific note.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ListNoteRevisionsArgs) (*mcp.CallToolResult, RevisionsOut, error) {
		rs, err := c.ListNoteRevisions(ctx, args.NoteID)
		if err != nil {
			return nil, RevisionsOut{}, err
		}
		out := RevisionsOut{Items: make([]RevisionOut, 0, len(rs))}
		for _, r := range rs {
			out.Items = append(out.Items, revisionOut(r))
			if r.EncryptionApplied {
				out.EncryptedItemsSkipped++
			}
		}
		return nil, out, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_revision",
		Description: "Get a single revision by ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GetRevisionArgs) (*mcp.CallToolResult, RevisionOut, error) {
		r, err := c.GetRevision(ctx, args.RevisionID)
		if err != nil {
			return nil, RevisionOut{}, err
		}
		return nil, revisionOut(r), nil
	})

}

var _ = joplin.IsNotFound
