package tools

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Drives every tool that we don't otherwise hit in tools_test.go so the
// register* functions and the joplin → MCP plumbing are exercised end to end.
func TestAllRemainingTools_Smoke(t *testing.T) {
	cs, cleanup := newTestServerPair(t, func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		// folders
		case path == "/folders" && r.Method == http.MethodGet:
			_, _ = io.WriteString(w, `{"items":[{"id":"f1","title":"F"}],"has_more":false}`)
		case path == "/folders" && r.Method == http.MethodPost:
			_, _ = io.WriteString(w, `{"id":"f2","title":"new"}`)
		case strings.HasPrefix(path, "/folders/") && strings.HasSuffix(path, "/notes"):
			_, _ = io.WriteString(w, `{"items":[{"id":"n1","title":"N","encryption_applied":false}],"has_more":false}`)
		case strings.HasPrefix(path, "/folders/") && r.Method == http.MethodGet:
			_, _ = io.WriteString(w, `{"id":"f1","title":"F"}`)
		case strings.HasPrefix(path, "/folders/") && r.Method == http.MethodPut:
			_, _ = io.WriteString(w, `{"id":"f1","title":"renamed"}`)
		case strings.HasPrefix(path, "/folders/") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		// tags
		case path == "/tags" && r.Method == http.MethodGet:
			_, _ = io.WriteString(w, `{"items":[{"id":"t1","title":"T"}],"has_more":false}`)
		case path == "/tags" && r.Method == http.MethodPost:
			_, _ = io.WriteString(w, `{"id":"t2","title":"new"}`)
		case strings.HasPrefix(path, "/tags/") && strings.HasSuffix(path, "/notes") && r.Method == http.MethodGet:
			_, _ = io.WriteString(w, `{"items":[{"id":"n1","title":"N","encryption_applied":false}],"has_more":false}`)
		case strings.HasPrefix(path, "/tags/") && strings.HasSuffix(path, "/notes") && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)
		case strings.HasPrefix(path, "/tags/") && strings.Contains(path, "/notes/") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case strings.HasPrefix(path, "/tags/") && r.Method == http.MethodGet:
			_, _ = io.WriteString(w, `{"id":"t1","title":"T"}`)
		case strings.HasPrefix(path, "/tags/") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		// resources
		case path == "/resources" && r.Method == http.MethodGet:
			_, _ = io.WriteString(w, `{"items":[{"id":"r1"}],"has_more":false}`)
		case path == "/resources" && r.Method == http.MethodPost:
			_, _ = io.WriteString(w, `{"id":"r2","title":"u"}`)
		case strings.HasPrefix(path, "/resources/") && r.Method == http.MethodGet:
			_, _ = io.WriteString(w, `{"id":"r1","mime":"image/png","encryption_applied":false}`)
		case strings.HasPrefix(path, "/resources/") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		// notes (used by update + get_note_with_context)
		case path == "/notes/n1/tags":
			_, _ = io.WriteString(w, `{"items":[],"has_more":false}`)
		case path == "/notes/n1/resources":
			_, _ = io.WriteString(w, `{"items":[],"has_more":false}`)
		case strings.HasPrefix(path, "/notes/") && r.Method == http.MethodGet:
			_, _ = io.WriteString(w, `{"id":"n1","title":"N","body":"B","encryption_applied":false}`)
		case strings.HasPrefix(path, "/notes/") && r.Method == http.MethodPut:
			_, _ = io.WriteString(w, `{"id":"n1","title":"new","encryption_applied":false}`)
		case strings.HasPrefix(path, "/notes/") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		// revisions
		case path == "/revisions":
			_, _ = io.WriteString(w, `{"items":[{"id":"rev1","item_id":"n1","encryption_applied":false}],"has_more":false}`)
		case strings.HasPrefix(path, "/revisions/"):
			_, _ = io.WriteString(w, `{"id":"rev1","item_id":"n1"}`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, path)
		}
	})
	defer cleanup()

	ctx := context.Background()
	calls := []struct {
		name string
		args map[string]any
	}{
		{"list_folders", nil},
		{"get_folder", map[string]any{"folder_id": "f1"}},
		{"create_folder", map[string]any{"title": "new"}},
		{"update_folder", map[string]any{"folder_id": "f1", "title": "renamed"}},
		{"delete_folder", map[string]any{"folder_id": "f1"}},
		{"list_notes_in_folder", map[string]any{"folder_id": "f1"}},
		{"list_tags", nil},
		{"get_tag", map[string]any{"tag_id": "t1"}},
		{"create_tag", map[string]any{"title": "new"}},
		{"delete_tag", map[string]any{"tag_id": "t1"}},
		{"tag_note", map[string]any{"tag_id": "t1", "note_id": "n1"}},
		{"untag_note", map[string]any{"tag_id": "t1", "note_id": "n1"}},
		{"list_notes_with_tag", map[string]any{"tag_id": "t1"}},
		{"list_resources", nil},
		{"get_resource_metadata", map[string]any{"resource_id": "r1"}},
		{"upload_resource", map[string]any{"filename": "x.png", "title": "u", "base64_data": "UE5H"}},
		{"delete_resource", map[string]any{"resource_id": "r1"}},
		{"get_note_with_context", map[string]any{"note_id": "n1"}},
		{"update_note", map[string]any{"note_id": "n1", "title": ptrString("new")}},
		{"delete_note", map[string]any{"note_id": "n1"}},
		{"list_note_revisions", map[string]any{"note_id": "n1"}},
		{"get_revision", map[string]any{"revision_id": "rev1"}},
	}
	for _, c := range calls {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: c.name, Arguments: c.args})
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if res.IsError {
			t.Errorf("%s: tool error: %v", c.name, res.Content)
		}
	}
}

func ptrString(s string) *string { return &s }

func TestUploadResource_RejectsBadBase64(t *testing.T) {
	cs, cleanup := newTestServerPair(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("server should not be called for bad base64")
		_ = w
	})
	defer cleanup()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "upload_resource",
		Arguments: map[string]any{"filename": "x.png", "base64_data": "not-base64!!!"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected IsError for bad base64")
	}
}
