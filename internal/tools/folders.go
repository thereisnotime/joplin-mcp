package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

type ListFoldersArgs struct {
	Page  int `json:"page,omitempty"`
	Limit int `json:"limit,omitempty"`
}

type GetFolderArgs struct {
	FolderID string `json:"folder_id"`
}

type CreateFolderArgs struct {
	Title    string `json:"title"`
	ParentID string `json:"parent_id,omitempty" jsonschema:"optional parent folder ID for nested folders"`
	Emoji    string `json:"emoji,omitempty" jsonschema:"optional folder icon emoji (e.g. 🚀); shown in the Joplin sidebar"`
}

type UpdateFolderArgs struct {
	FolderID string  `json:"folder_id"`
	Title    *string `json:"title,omitempty"`
	ParentID *string `json:"parent_id,omitempty"`
	Emoji    *string `json:"emoji,omitempty" jsonschema:"set or change the folder's icon emoji; pass empty string to clear it"`
}

type SetFolderIconArgs struct {
	FolderID string `json:"folder_id"`
	Emoji    string `json:"emoji" jsonschema:"the emoji to use as the folder icon; pass empty string to clear"`
}

type DeleteFolderArgs struct {
	FolderID  string `json:"folder_id"`
	Permanent bool   `json:"permanent,omitempty"`
}

type ListFolderNotesArgs struct {
	FolderID string `json:"folder_id"`
	Page     int    `json:"page,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

func registerFolderTools(srv *mcp.Server, c *joplin.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_folders",
		Description: "List notebooks (folders), paginated.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ListFoldersArgs) (*mcp.CallToolResult, PageOut[FolderOut], error) {
		p, err := c.ListFolders(ctx, joplin.ListOptions{Page: args.Page, Limit: args.Limit})
		if err != nil {
			return nil, PageOut[FolderOut]{}, err
		}
		return nil, foldersPage(p), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_folder",
		Description: "Get a single folder by ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GetFolderArgs) (*mcp.CallToolResult, FolderOut, error) {
		f, err := c.GetFolder(ctx, args.FolderID)
		if err != nil {
			return nil, FolderOut{}, err
		}
		return nil, folderOut(f), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_folder",
		Description: "Create a folder. Set parent_id for nested folders. Optional emoji becomes the folder's sidebar icon.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreateFolderArgs) (*mcp.CallToolResult, FolderOut, error) {
		in := joplin.CreateFolderInput{Title: args.Title, ParentID: args.ParentID}
		if args.Emoji != "" {
			in.Icon = joplin.EmojiIcon(args.Emoji)
		}
		f, err := c.CreateFolder(ctx, in)
		if err != nil {
			return nil, FolderOut{}, err
		}
		return nil, folderOut(f), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_folder",
		Description: "Partially update a folder. Pass emoji='' to clear the icon.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args UpdateFolderArgs) (*mcp.CallToolResult, FolderOut, error) {
		in := joplin.UpdateFolderInput{Title: args.Title, ParentID: args.ParentID}
		if args.Emoji != nil {
			icon := joplin.EmojiIcon(*args.Emoji)
			in.Icon = &icon
		}
		f, err := c.UpdateFolder(ctx, args.FolderID, in)
		if err != nil {
			return nil, FolderOut{}, err
		}
		return nil, folderOut(f), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "set_folder_icon",
		Description: "Convenience tool: set or clear a folder's sidebar icon emoji. Pass empty string to clear.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SetFolderIconArgs) (*mcp.CallToolResult, FolderOut, error) {
		icon := joplin.EmojiIcon(args.Emoji)
		f, err := c.UpdateFolder(ctx, args.FolderID, joplin.UpdateFolderInput{Icon: &icon})
		if err != nil {
			return nil, FolderOut{}, err
		}
		return nil, folderOut(f), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_folder",
		Description: "Delete a folder. Set permanent=true to bypass trash.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args DeleteFolderArgs) (*mcp.CallToolResult, DeleteOut, error) {
		if err := c.DeleteFolder(ctx, args.FolderID, args.Permanent); err != nil {
			return nil, DeleteOut{}, err
		}
		return nil, DeleteOut{OK: true}, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_notes_in_folder",
		Description: "List notes whose parent_id is the given folder.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ListFolderNotesArgs) (*mcp.CallToolResult, PageOut[NoteOut], error) {
		p, err := c.ListFolderNotes(ctx, args.FolderID, joplin.ListOptions{Page: args.Page, Limit: args.Limit})
		if err != nil {
			return nil, PageOut[NoteOut]{}, err
		}
		return nil, notesPage(p), nil
	})
}
