package tools

import (
	"context"
	"sort"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

// bulkConcurrency caps how many in-flight requests we send to Joplin's local
// API. Higher than this risks tripping the desktop's request handler queue
// without saving real time on a localhost loop.
const bulkConcurrency = 4

type BulkResult struct {
	Succeeded []string          `json:"succeeded"`
	Failed    map[string]string `json:"failed,omitempty" jsonschema:"item_id → error message for items that failed"`
}

type BulkTagArgs struct {
	NoteIDs []string `json:"note_ids"`
	TagID   string   `json:"tag_id"`
}

type BulkUntagArgs struct {
	NoteIDs []string `json:"note_ids"`
	TagID   string   `json:"tag_id"`
}

type BulkMoveArgs struct {
	NoteIDs  []string `json:"note_ids"`
	ParentID string   `json:"parent_id" jsonschema:"target folder ID for all the notes"`
}

type BulkDeleteArgs struct {
	NoteIDs   []string `json:"note_ids"`
	Permanent bool     `json:"permanent,omitempty" jsonschema:"if true, bypass trash"`
}

func registerBulkTools(srv *mcp.Server, c *joplin.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "bulk_tag_notes",
		Description: "Attach the same tag to many notes in parallel. Returns per-id success/failure.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args BulkTagArgs) (*mcp.CallToolResult, BulkResult, error) {
		return nil, runBulk(ctx, args.NoteIDs, func(ctx context.Context, id string) error {
			return c.TagNote(ctx, args.TagID, id)
		}), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "bulk_untag_notes",
		Description: "Detach the same tag from many notes in parallel. Returns per-id success/failure.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args BulkUntagArgs) (*mcp.CallToolResult, BulkResult, error) {
		return nil, runBulk(ctx, args.NoteIDs, func(ctx context.Context, id string) error {
			return c.UntagNote(ctx, args.TagID, id)
		}), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "bulk_move_notes",
		Description: "Move many notes into the same folder in parallel. Returns per-id success/failure.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args BulkMoveArgs) (*mcp.CallToolResult, BulkResult, error) {
		parent := args.ParentID
		return nil, runBulk(ctx, args.NoteIDs, func(ctx context.Context, id string) error {
			_, err := c.UpdateNote(ctx, id, joplin.UpdateNoteInput{ParentID: &parent})
			return err
		}), nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "bulk_delete_notes",
		Description: "Delete many notes (default: trash; set permanent=true to bypass trash). Returns per-id success/failure.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args BulkDeleteArgs) (*mcp.CallToolResult, BulkResult, error) {
		return nil, runBulk(ctx, args.NoteIDs, func(ctx context.Context, id string) error {
			return c.DeleteNote(ctx, id, args.Permanent)
		}), nil
	})
}

func runBulk(ctx context.Context, ids []string, op func(context.Context, string) error) BulkResult {
	res := BulkResult{Failed: map[string]string{}}
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, bulkConcurrency)
	for _, id := range ids {
		select {
		case <-ctx.Done():
			mu.Lock()
			res.Failed[id] = ctx.Err().Error()
			mu.Unlock()
			continue
		default:
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(id string) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := op(ctx, id); err != nil {
				mu.Lock()
				res.Failed[id] = err.Error()
				mu.Unlock()
				return
			}
			mu.Lock()
			res.Succeeded = append(res.Succeeded, id)
			mu.Unlock()
		}(id)
	}
	wg.Wait()
	// Stable, predictable ordering for the LLM.
	sort.Strings(res.Succeeded)
	if len(res.Failed) == 0 {
		res.Failed = nil
	}
	return res
}
