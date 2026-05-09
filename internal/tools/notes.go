package tools

import (
	"context"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

type ListNotesArgs struct {
	Page  int `json:"page,omitempty"`
	Limit int `json:"limit,omitempty"`
}

type GetNoteArgs struct {
	NoteID string `json:"note_id"`
}

type CreateNoteArgs struct {
	Title    string `json:"title"`
	Body     string `json:"body,omitempty"`
	ParentID string `json:"parent_id,omitempty"`
	IsTodo   bool   `json:"is_todo,omitempty"`
}

type UpdateNoteArgs struct {
	NoteID   string  `json:"note_id"`
	Title    *string `json:"title,omitempty"`
	Body     *string `json:"body,omitempty"`
	ParentID *string `json:"parent_id,omitempty"`
	IsTodo   *bool   `json:"is_todo,omitempty"`
}

type DeleteNoteArgs struct {
	NoteID    string `json:"note_id"`
	Permanent bool   `json:"permanent,omitempty" jsonschema:"if true bypass trash"`
}

type DeleteOut struct {
	OK bool `json:"ok"`
}

func registerNoteTools(srv *mcp.Server, c *joplin.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_notes",
		Description: "List notes, paginated. Returns encryption_applied per item and an encrypted_items_skipped count.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ListNotesArgs) (*mcp.CallToolResult, PageOut[NoteOut], error) {
		p, err := c.ListNotes(ctx, joplin.ListOptions{Page: args.Page, Limit: args.Limit})
		if err != nil {
			return nil, PageOut[NoteOut]{}, err
		}
		return nil, notesPage(p), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_note",
		Description: "Get a single note by ID. encryption_applied indicates whether the body could be returned.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GetNoteArgs) (*mcp.CallToolResult, NoteOut, error) {
		n, err := c.GetNote(ctx, args.NoteID)
		if err != nil {
			return nil, NoteOut{}, err
		}
		return nil, noteOut(n), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_note_with_context",
		Description: "Get a note plus its tags and attached resources, fetched in parallel.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GetNoteArgs) (*mcp.CallToolResult, NoteContextOut, error) {
		var (
			out     NoteContextOut
			noteErr error
			tagsErr error
			resErr  error
			noteRaw joplin.Note
			tagsRaw joplin.Page[joplin.Tag]
			resRaw  joplin.Page[joplin.Resource]
			wg      sync.WaitGroup
		)
		wg.Add(3)
		go func() { defer wg.Done(); noteRaw, noteErr = c.GetNote(ctx, args.NoteID) }()
		go func() { defer wg.Done(); tagsRaw, tagsErr = c.ListNoteTags(ctx, args.NoteID, joplin.ListOptions{}) }()
		go func() { defer wg.Done(); resRaw, resErr = c.ListNoteResources(ctx, args.NoteID, joplin.ListOptions{}) }()
		wg.Wait()
		if noteErr != nil {
			return nil, out, noteErr
		}
		if tagsErr != nil {
			return nil, out, tagsErr
		}
		if resErr != nil {
			return nil, out, resErr
		}
		out.Note = noteOut(noteRaw)
		// Initialise as empty slices so JSON serialises [] not null —
		// don't make the LLM second-guess whether the field was unset.
		out.Tags = make([]TagOut, 0, len(tagsRaw.Items))
		out.Resources = make([]ResourceOut, 0, len(resRaw.Items))
		for _, t := range tagsRaw.Items {
			out.Tags = append(out.Tags, tagOut(t))
		}
		for _, r := range resRaw.Items {
			out.Resources = append(out.Resources, resourceOut(r))
		}
		return nil, out, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_note",
		Description: "Create a note. Returns the created note.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreateNoteArgs) (*mcp.CallToolResult, NoteOut, error) {
		isTodo := args.IsTodo
		n, err := c.CreateNote(ctx, joplin.CreateNoteInput{
			Title:    args.Title,
			Body:     args.Body,
			ParentID: args.ParentID,
			IsTodo:   &isTodo,
		})
		if err != nil {
			return nil, NoteOut{}, err
		}
		return nil, noteOut(n), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_note",
		Description: "Partially update a note. Only fields that are set will be sent to Joplin.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args UpdateNoteArgs) (*mcp.CallToolResult, NoteOut, error) {
		n, err := c.UpdateNote(ctx, args.NoteID, joplin.UpdateNoteInput{
			Title: args.Title, Body: args.Body, ParentID: args.ParentID, IsTodo: args.IsTodo,
		})
		if err != nil {
			return nil, NoteOut{}, err
		}
		return nil, noteOut(n), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_note",
		Description: "Delete a note (default: move to trash). Set permanent=true to bypass trash.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args DeleteNoteArgs) (*mcp.CallToolResult, DeleteOut, error) {
		if err := c.DeleteNote(ctx, args.NoteID, args.Permanent); err != nil {
			return nil, DeleteOut{}, err
		}
		return nil, DeleteOut{OK: true}, nil
	})
}
