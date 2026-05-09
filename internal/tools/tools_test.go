package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

func newTestServerPair(t *testing.T, h http.HandlerFunc) (*mcp.ClientSession, func()) {
	return newTestServerPairWithMax(t, 0, h)
}

func newTestServerPairWithMax(t *testing.T, maxBytes int64, h http.HandlerFunc) (*mcp.ClientSession, func()) {
	t.Helper()
	joplinSrv := httptest.NewServer(h)
	t.Cleanup(joplinSrv.Close)

	jc, err := joplin.New(joplin.Options{Token: "tok-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", BaseURL: joplinSrv.URL})
	if err != nil {
		t.Fatal(err)
	}

	srv := New(jc, Options{Version: "test", MaxResourceBytes: maxBytes})
	ctx := context.Background()

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := srv.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	return clientSession, func() {
		_ = clientSession.Close()
		serverSession.Wait()
	}
}

func decodeStructured[T any](t *testing.T, res *mcp.CallToolResult) T {
	t.Helper()
	var out T
	switch sc := res.StructuredContent.(type) {
	case nil:
		// fall through to text content path below
	case json.RawMessage:
		if err := json.Unmarshal(sc, &out); err != nil {
			t.Fatalf("unmarshal structured content: %v", err)
		}
		return out
	default:
		b, err := json.Marshal(sc)
		if err != nil {
			t.Fatalf("marshal structured content: %v", err)
		}
		if err := json.Unmarshal(b, &out); err != nil {
			t.Fatalf("unmarshal structured content: %v", err)
		}
		return out
	}
	if len(res.Content) == 0 {
		t.Fatal("no structured content and no text content in result")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content[0] = %T, want *TextContent", res.Content[0])
	}
	if err := json.Unmarshal([]byte(tc.Text), &out); err != nil {
		t.Fatalf("unmarshal text content: %v", err)
	}
	return out
}

func TestListNotes_AnnotatesEncryptionAndCountsSkipped(t *testing.T) {
	cs, cleanup := newTestServerPair(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"items":[
			{"id":"a","title":"A","body":"plain","encryption_applied":false},
			{"id":"b","title":"B","encryption_applied":true,"master_key_id":"mk1"},
			{"id":"c","title":"C","body":"plain","encryption_applied":false}
		],"has_more":false}`)
	})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_notes"})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %v", res.Content)
	}
	out := decodeStructured[PageOut[NoteOut]](t, res)
	if len(out.Items) != 3 {
		t.Errorf("got %d items, want 3", len(out.Items))
	}
	if out.EncryptedItemsSkipped != 1 {
		t.Errorf("EncryptedItemsSkipped = %d, want 1", out.EncryptedItemsSkipped)
	}
	for _, n := range out.Items {
		if n.ID == "b" {
			if !n.EncryptionApplied {
				t.Error("note b: EncryptionApplied = false, want true")
			}
			if n.Body != "" {
				t.Errorf("note b: Body = %q, want empty (encrypted)", n.Body)
			}
			if n.MasterKeyID != "mk1" {
				t.Errorf("note b: MasterKeyID = %q, want mk1", n.MasterKeyID)
			}
		}
	}
}

func TestGetNote_DropsBodyWhenEncrypted(t *testing.T) {
	cs, cleanup := newTestServerPair(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"id":"x","title":"T","body":"this should be hidden","encryption_applied":true,"master_key_id":"mk1"}`)
	})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_note",
		Arguments: map[string]any{"note_id": "x"},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := decodeStructured[NoteOut](t, res)
	if !out.EncryptionApplied {
		t.Error("EncryptionApplied = false")
	}
	if out.Body != "" {
		t.Errorf("Body = %q, want empty when encrypted", out.Body)
	}
	if out.MasterKeyID != "mk1" {
		t.Errorf("MasterKeyID = %q", out.MasterKeyID)
	}
}

func TestCreateNote_PassesArguments(t *testing.T) {
	var bodySeen string
	cs, cleanup := newTestServerPair(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		b, _ := io.ReadAll(r.Body)
		bodySeen = string(b)
		_, _ = io.WriteString(w, `{"id":"n1","title":"T","body":"B"}`)
	})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "create_note",
		Arguments: map[string]any{"title": "T", "body": "B", "is_todo": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := decodeStructured[NoteOut](t, res)
	if out.ID != "n1" {
		t.Errorf("ID = %q", out.ID)
	}
	if !strings.Contains(bodySeen, `"title":"T"`) {
		t.Errorf("upstream body missing title: %s", bodySeen)
	}
	if !strings.Contains(bodySeen, `"is_todo":true`) {
		t.Errorf("upstream body missing is_todo: %s", bodySeen)
	}
}

func TestSearch_PassesQueryToJoplin(t *testing.T) {
	var qSeen string
	cs, cleanup := newTestServerPair(t, func(w http.ResponseWriter, r *http.Request) {
		qSeen = r.URL.Query().Get("query")
		_, _ = io.WriteString(w, `{"items":[{"id":"a","title":"A"}],"has_more":false}`)
	})
	defer cleanup()

	_, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "search",
		Arguments: map[string]any{"query": "tag:work"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if qSeen != "tag:work" {
		t.Errorf("upstream query = %q", qSeen)
	}
}

// Encryption-transparency-spec scenario: download_resource refuses to return
// ciphertext bytes silently.
func TestDownloadResource_RefusesEncrypted(t *testing.T) {
	cs, cleanup := newTestServerPair(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/file") {
			t.Error("file endpoint hit; expected refusal before download")
			return
		}
		_, _ = io.WriteString(w, `{"id":"r1","encryption_applied":true,"master_key_id":"mk1"}`)
	})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "download_resource",
		Arguments: map[string]any{"resource_id": "r1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected IsError=true for encrypted resource")
	}
	if len(res.Content) == 0 {
		t.Fatal("no error content")
	}
	if tc, ok := res.Content[0].(*mcp.TextContent); ok {
		if !strings.Contains(tc.Text, "encrypted") {
			t.Errorf("error text = %q, missing 'encrypted'", tc.Text)
		}
	}
}

func TestDownloadResource_RefusesOversizeFromMetadata(t *testing.T) {
	cs, cleanup := newTestServerPairWithMax(t, 100, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/file") {
			t.Error("file endpoint hit; expected refusal before download")
			return
		}
		_, _ = io.WriteString(w, `{"id":"r1","size":999999,"encryption_applied":false}`)
	})
	defer cleanup()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "download_resource",
		Arguments: map[string]any{"resource_id": "r1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected IsError for oversize resource")
	}
	if tc, ok := res.Content[0].(*mcp.TextContent); ok && !strings.Contains(tc.Text, "exceeds") {
		t.Errorf("error text = %q, missing 'exceeds'", tc.Text)
	}
}

func TestUploadResource_RefusesOversize(t *testing.T) {
	cs, cleanup := newTestServerPairWithMax(t, 4, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("server should not be hit for oversize upload")
		_ = w
	})
	defer cleanup()
	// "ABCDEFGH" = 8 bytes raw, well above the 4-byte cap.
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "upload_resource",
		Arguments: map[string]any{"filename": "x.bin", "base64_data": base64.StdEncoding.EncodeToString([]byte("ABCDEFGH"))},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected IsError for oversize upload")
	}
}

func TestDownloadResource_ReturnsBase64WhenDecrypted(t *testing.T) {
	cs, cleanup := newTestServerPair(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/file") {
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("PNGBYTES"))
			return
		}
		_, _ = io.WriteString(w, `{"id":"r1","mime":"image/png","encryption_applied":false}`)
	})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "download_resource",
		Arguments: map[string]any{"resource_id": "r1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %v", res.Content)
	}
	out := decodeStructured[DownloadResourceOut](t, res)
	if out.ContentType != "image/png" {
		t.Errorf("content_type = %q", out.ContentType)
	}
	dec, err := base64.StdEncoding.DecodeString(out.Base64Data)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	if string(dec) != "PNGBYTES" {
		t.Errorf("decoded = %q, want PNGBYTES", string(dec))
	}
}

func TestListChangesSince_PassesCursor(t *testing.T) {
	var cSeen string
	cs, cleanup := newTestServerPair(t, func(w http.ResponseWriter, r *http.Request) {
		cSeen = r.URL.Query().Get("cursor")
		_, _ = io.WriteString(w, `{"items":[{"id":42,"item_id":"n1","type":2}],"cursor":42,"has_more":false}`)
	})
	defer cleanup()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "list_changes_since",
		Arguments: map[string]any{"since": 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := decodeStructured[EventsOut](t, res)
	if cSeen != "10" {
		t.Errorf("upstream cursor = %q", cSeen)
	}
	if out.Cursor != 42 {
		t.Errorf("returned cursor = %d", out.Cursor)
	}
}

func TestListTools_AllRegistered(t *testing.T) {
	cs, cleanup := newTestServerPair(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{}`)
	})
	defer cleanup()

	res, err := cs.ListTools(context.Background(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"list_notes", "get_note", "get_note_with_context", "create_note", "update_note", "delete_note",
		"list_folders", "get_folder", "create_folder", "update_folder", "delete_folder", "list_notes_in_folder",
		"list_tags", "get_tag", "create_tag", "delete_tag", "tag_note", "untag_note", "list_notes_with_tag",
		"search",
		"list_resources", "get_resource_metadata", "download_resource", "upload_resource", "delete_resource",
		"list_changes_since",
		"list_note_revisions", "get_revision",
	}
	got := map[string]bool{}
	for _, tl := range res.Tools {
		got[tl.Name] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("missing tool: %s", w)
		}
	}
}
