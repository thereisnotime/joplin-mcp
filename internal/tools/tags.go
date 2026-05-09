package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

type ListTagsArgs struct {
	Page  int `json:"page,omitempty"`
	Limit int `json:"limit,omitempty"`
}

type GetTagArgs struct {
	TagID string `json:"tag_id"`
}

type CreateTagArgs struct {
	Title string `json:"title"`
}

type DeleteTagArgs struct {
	TagID string `json:"tag_id"`
}

type TagNoteArgs struct {
	TagID  string `json:"tag_id"`
	NoteID string `json:"note_id"`
}

type ListTagNotesArgs struct {
	TagID string `json:"tag_id"`
	Page  int    `json:"page,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

func registerTagTools(srv *mcp.Server, c *joplin.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_tags",
		Description: "List tags, paginated.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ListTagsArgs) (*mcp.CallToolResult, PageOut[TagOut], error) {
		p, err := c.ListTags(ctx, joplin.ListOptions{Page: args.Page, Limit: args.Limit})
		if err != nil {
			return nil, PageOut[TagOut]{}, err
		}
		return nil, tagsPage(p), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_tag",
		Description: "Get a single tag by ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GetTagArgs) (*mcp.CallToolResult, TagOut, error) {
		t, err := c.GetTag(ctx, args.TagID)
		if err != nil {
			return nil, TagOut{}, err
		}
		return nil, tagOut(t), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_tag",
		Description: "Create a tag.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreateTagArgs) (*mcp.CallToolResult, TagOut, error) {
		t, err := c.CreateTag(ctx, joplin.CreateTagInput{Title: args.Title})
		if err != nil {
			return nil, TagOut{}, err
		}
		return nil, tagOut(t), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_tag",
		Description: "Delete a tag.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args DeleteTagArgs) (*mcp.CallToolResult, DeleteOut, error) {
		if err := c.DeleteTag(ctx, args.TagID); err != nil {
			return nil, DeleteOut{}, err
		}
		return nil, DeleteOut{OK: true}, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "tag_note",
		Description: "Attach a tag to a note.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args TagNoteArgs) (*mcp.CallToolResult, DeleteOut, error) {
		if err := c.TagNote(ctx, args.TagID, args.NoteID); err != nil {
			return nil, DeleteOut{}, err
		}
		return nil, DeleteOut{OK: true}, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "untag_note",
		Description: "Detach a tag from a note.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args TagNoteArgs) (*mcp.CallToolResult, DeleteOut, error) {
		if err := c.UntagNote(ctx, args.TagID, args.NoteID); err != nil {
			return nil, DeleteOut{}, err
		}
		return nil, DeleteOut{OK: true}, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_notes_with_tag",
		Description: "List notes that have the given tag.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ListTagNotesArgs) (*mcp.CallToolResult, PageOut[NoteOut], error) {
		p, err := c.ListTagNotes(ctx, args.TagID, joplin.ListOptions{Page: args.Page, Limit: args.Limit})
		if err != nil {
			return nil, PageOut[NoteOut]{}, err
		}
		return nil, notesPage(p), nil
	})
}
