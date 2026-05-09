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
}

type UpdateFolderArgs struct {
	FolderID string  `json:"folder_id"`
	Title    *string `json:"title,omitempty"`
	ParentID *string `json:"parent_id,omitempty"`
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
		Description: "Create a folder. Set parent_id for nested folders.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreateFolderArgs) (*mcp.CallToolResult, FolderOut, error) {
		f, err := c.CreateFolder(ctx, joplin.CreateFolderInput{Title: args.Title, ParentID: args.ParentID})
		if err != nil {
			return nil, FolderOut{}, err
		}
		return nil, folderOut(f), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_folder",
		Description: "Partially update a folder.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args UpdateFolderArgs) (*mcp.CallToolResult, FolderOut, error) {
		f, err := c.UpdateFolder(ctx, args.FolderID, joplin.UpdateFolderInput{Title: args.Title, ParentID: args.ParentID})
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
