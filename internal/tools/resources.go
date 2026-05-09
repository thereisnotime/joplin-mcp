package tools

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

var ErrResourceEncrypted = errors.New("resource is encrypted on this device; unlock Joplin and retry")

type ListResourcesArgs struct {
	Page  int `json:"page,omitempty"`
	Limit int `json:"limit,omitempty"`
}

type GetResourceArgs struct {
	ResourceID string `json:"resource_id"`
}

type DownloadResourceArgs struct {
	ResourceID string `json:"resource_id"`
}

type UploadResourceArgs struct {
	Filename   string `json:"filename" jsonschema:"file name including extension — Joplin uses this to infer MIME type"`
	Title      string `json:"title,omitempty"`
	Base64Data string `json:"base64_data" jsonschema:"the file's bytes, base64-encoded"`
}

type DeleteResourceArgs struct {
	ResourceID string `json:"resource_id"`
}

func registerResourceTools(srv *mcp.Server, c *joplin.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_resources",
		Description: "List resources (attachments), paginated.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ListResourcesArgs) (*mcp.CallToolResult, PageOut[ResourceOut], error) {
		p, err := c.ListResources(ctx, joplin.ListOptions{Page: args.Page, Limit: args.Limit})
		if err != nil {
			return nil, PageOut[ResourceOut]{}, err
		}
		return nil, resourcesPage(p), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_resource_metadata",
		Description: "Get metadata for a single resource (does not include the file bytes).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GetResourceArgs) (*mcp.CallToolResult, ResourceOut, error) {
		r, err := c.GetResource(ctx, args.ResourceID)
		if err != nil {
			return nil, ResourceOut{}, err
		}
		return nil, resourceOut(r), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "download_resource",
		Description: "Download a resource's file bytes (base64-encoded). Refuses when the resource is still encrypted on the local device.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args DownloadResourceArgs) (*mcp.CallToolResult, DownloadResourceOut, error) {
		meta, err := c.GetResource(ctx, args.ResourceID)
		if err != nil {
			return nil, DownloadResourceOut{}, err
		}
		if meta.EncryptionApplied || meta.EncryptionBlobEncrypted {
			return nil, DownloadResourceOut{}, ErrResourceEncrypted
		}
		data, ct, err := c.DownloadResource(ctx, args.ResourceID)
		if err != nil {
			return nil, DownloadResourceOut{}, err
		}
		return nil, DownloadResourceOut{
			Base64Data:  base64.StdEncoding.EncodeToString(data),
			ContentType: ct,
			Size:        len(data),
		}, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "upload_resource",
		Description: "Upload a new resource. Provide the bytes as base64.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args UploadResourceArgs) (*mcp.CallToolResult, ResourceOut, error) {
		data, err := base64.StdEncoding.DecodeString(args.Base64Data)
		if err != nil {
			return nil, ResourceOut{}, err
		}
		r, err := c.UploadResource(ctx, data, args.Filename, args.Title)
		if err != nil {
			return nil, ResourceOut{}, err
		}
		return nil, resourceOut(r), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_resource",
		Description: "Delete a resource.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args DeleteResourceArgs) (*mcp.CallToolResult, DeleteOut, error) {
		if err := c.DeleteResource(ctx, args.ResourceID); err != nil {
			return nil, DeleteOut{}, err
		}
		return nil, DeleteOut{OK: true}, nil
	})
}
