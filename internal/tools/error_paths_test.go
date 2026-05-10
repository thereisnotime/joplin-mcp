package tools

import (
	"context"
	"net/http"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestErrorPaths walks every tool group and confirms that when Joplin returns
// 500, the tool surfaces an MCP error rather than panicking or returning a
// silent empty success. Pure coverage exercise — exhaustive happy-path
// testing lives in tools_test.go and the e2e suite.
func TestErrorPaths(t *testing.T) {
	cs, cleanup := newTestServerPair(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer cleanup()

	calls := []struct {
		name string
		args map[string]any
	}{
		{"list_notes", nil},
		{"get_note", map[string]any{"note_id": "x"}},
		{"create_note", map[string]any{"title": "x"}},
		{"update_note", map[string]any{"note_id": "x", "title": ptrString("new")}},
		{"delete_note", map[string]any{"note_id": "x"}},
		{"list_folders", nil},
		{"get_folder", map[string]any{"folder_id": "x"}},
		{"create_folder", map[string]any{"title": "x"}},
		{"update_folder", map[string]any{"folder_id": "x", "title": ptrString("new")}},
		{"set_folder_icon", map[string]any{"folder_id": "x", "emoji": "🎯"}},
		{"delete_folder", map[string]any{"folder_id": "x"}},
		{"list_notes_in_folder", map[string]any{"folder_id": "x"}},
		{"list_tags", nil},
		{"get_tag", map[string]any{"tag_id": "x"}},
		{"create_tag", map[string]any{"title": "x"}},
		{"delete_tag", map[string]any{"tag_id": "x"}},
		{"tag_note", map[string]any{"tag_id": "x", "note_id": "y"}},
		{"untag_note", map[string]any{"tag_id": "x", "note_id": "y"}},
		{"list_notes_with_tag", map[string]any{"tag_id": "x"}},
		{"search", map[string]any{"query": "x"}},
		{"list_resources", nil},
		{"get_resource_metadata", map[string]any{"resource_id": "x"}},
		{"delete_resource", map[string]any{"resource_id": "x"}},
		{"list_changes_since", nil},
		{"list_note_revisions", map[string]any{"note_id": "x"}},
		{"get_revision", map[string]any{"revision_id": "x"}},
		{"list_outbound_links", map[string]any{"note_id": "x"}},
		{"list_backlinks", map[string]any{"note_id": "x"}},
		{"list_trash", nil},
		{"restore_note_from_trash", map[string]any{"note_id": "x"}},
		{"empty_trash", nil},
		{"health", nil}, // health swallows the error and reports unreachable
		{"list_master_keys", nil},
		{"get_master_key", map[string]any{"master_key_id": "x"}},
	}

	ctx := context.Background()
	for _, c := range calls {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: c.name, Arguments: c.args})
		if err != nil {
			// Protocol-level errors are also acceptable — what we don't
			// want is a successful response containing wrong data.
			continue
		}
		if c.name == "health" {
			// health treats Joplin unreachable as a normal response with
			// joplin_reachable=false; it should NOT be IsError.
			if res.IsError {
				t.Errorf("%s should swallow upstream error into structured response, but IsError=true", c.name)
			}
			continue
		}
		if !res.IsError {
			t.Errorf("%s with upstream 500: IsError=false, expected error surface", c.name)
		}
	}
}
