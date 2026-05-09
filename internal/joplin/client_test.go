package joplin

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const testToken = "abc123abc123abc123abc123abc123ab"

func newTestClient(t *testing.T, h http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := New(Options{Token: testToken, BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c, srv
}

func TestNew_TokenRequired(t *testing.T) {
	if _, err := New(Options{}); err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestNew_DefaultsApplied(t *testing.T) {
	c, err := New(Options{Token: testToken})
	if err != nil {
		t.Fatal(err)
	}
	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
	if c.http.Timeout != DefaultTimeout {
		t.Errorf("timeout = %v, want %v", c.http.Timeout, DefaultTimeout)
	}
}

// Encryption-transparency-spec scenario: token attached to every request.
func TestRequest_TokenInQueryNoAuthHeader(t *testing.T) {
	var sawHeader bool
	var sawToken string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			sawHeader = true
		}
		sawToken = r.URL.Query().Get("token")
		_, _ = io.WriteString(w, `{"items":[],"has_more":false}`)
	})
	if _, err := c.ListNotes(context.Background(), ListOptions{}); err != nil {
		t.Fatal(err)
	}
	if sawHeader {
		t.Error("Authorization header set; should not be")
	}
	if sawToken != testToken {
		t.Errorf("token query = %q, want %q", sawToken, testToken)
	}
}

// joplin-rest-client-spec scenario: typed APIError with status code.
func TestRequest_NonOKReturnsAPIError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, "note not found")
	})
	_, err := c.GetNote(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err is %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if !strings.Contains(apiErr.Error(), "note not found") {
		t.Errorf("Error() = %q, missing body message", apiErr.Error())
	}
	if !IsNotFound(err) {
		t.Error("IsNotFound = false, want true")
	}
}

// joplin-rest-client-spec scenario: context cancellation aborts a request.
func TestRequest_ContextCancellation(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
		w.WriteHeader(http.StatusOK)
	})
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(20 * time.Millisecond); cancel() }()
	_, err := c.ListNotes(ctx, ListOptions{})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

// joplin-rest-client-spec scenario: full type coverage.
// Verifies that parent_id, encryption_applied, and master_key_id are populated.
func TestGetNote_PopulatesFullFieldSet(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{
			"id":"abc","title":"hello","body":"world",
			"parent_id":"folder1","encryption_applied":true,
			"master_key_id":"mk1","is_todo":true,"created_time":1700000000000
		}`)
	})
	n, err := c.GetNote(context.Background(), "abc")
	if err != nil {
		t.Fatal(err)
	}
	if n.ParentID != "folder1" {
		t.Errorf("ParentID = %q, want folder1", n.ParentID)
	}
	if !n.EncryptionApplied {
		t.Error("EncryptionApplied = false, want true")
	}
	if n.MasterKeyID != "mk1" {
		t.Errorf("MasterKeyID = %q, want mk1", n.MasterKeyID)
	}
	if !n.IsTodo {
		t.Error("IsTodo = false, want true")
	}
}

// joplin-rest-client-spec scenario: unknown fields do not break parsing.
func TestGetNote_UnknownFieldsIgnored(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"id":"a","title":"t","totally_new_field_2030":42}`)
	})
	n, err := c.GetNote(context.Background(), "a")
	if err != nil {
		t.Fatal(err)
	}
	if n.ID != "a" {
		t.Errorf("ID = %q, want a", n.ID)
	}
}

// joplin-rest-client-spec scenario: pagination helper walks all pages.
func TestCollectAll_WalksAllPages(t *testing.T) {
	var hits int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		atomic.AddInt32(&hits, 1)
		switch page {
		case "1", "":
			_, _ = io.WriteString(w, `{"items":[{"id":"a","title":"A"},{"id":"b","title":"B"}],"has_more":true}`)
		case "2":
			_, _ = io.WriteString(w, `{"items":[{"id":"c","title":"C"}],"has_more":false}`)
		default:
			t.Fatalf("unexpected page %q", page)
		}
	})
	all, err := CollectAll(context.Background(), func(ctx context.Context, page int) (Page[Note], error) {
		return c.ListNotes(ctx, ListOptions{Page: page, Limit: 2})
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := len(all); got != 3 {
		t.Errorf("len(all) = %d, want 3", got)
	}
	if hits != 2 {
		t.Errorf("hit count = %d, want 2", hits)
	}
}

func TestCreateNote_PostsBodyAndDecodes(t *testing.T) {
	var bodySeen string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		b, _ := io.ReadAll(r.Body)
		bodySeen = string(b)
		_, _ = io.WriteString(w, `{"id":"new1","title":"t","body":"b"}`)
	})
	isTodo := true
	n, err := c.CreateNote(context.Background(), CreateNoteInput{Title: "t", Body: "b", IsTodo: &isTodo})
	if err != nil {
		t.Fatal(err)
	}
	if n.ID != "new1" {
		t.Errorf("ID = %q, want new1", n.ID)
	}
	if !strings.Contains(bodySeen, `"title":"t"`) || !strings.Contains(bodySeen, `"is_todo":true`) {
		t.Errorf("body = %q, missing fields", bodySeen)
	}
}

func TestUpdateNote_PartialUpdate(t *testing.T) {
	var bodySeen string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		b, _ := io.ReadAll(r.Body)
		bodySeen = string(b)
		_, _ = io.WriteString(w, `{"id":"abc","title":"new"}`)
	})
	newTitle := "new"
	if _, err := c.UpdateNote(context.Background(), "abc", UpdateNoteInput{Title: &newTitle}); err != nil {
		t.Fatal(err)
	}
	if bodySeen != `{"title":"new"}` {
		t.Errorf("body = %q, want partial update body", bodySeen)
	}
}

func TestDeleteNote_PermanentFlag(t *testing.T) {
	var permSeen string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		permSeen = r.URL.Query().Get("permanent")
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.DeleteNote(context.Background(), "x", true); err != nil {
		t.Fatal(err)
	}
	if permSeen != "1" {
		t.Errorf("permanent param = %q, want 1", permSeen)
	}
}

func TestSearchNotes_AddsQueryAndType(t *testing.T) {
	var qSeen, typeSeen string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		qSeen = r.URL.Query().Get("query")
		typeSeen = r.URL.Query().Get("type")
		_, _ = io.WriteString(w, `{"items":[],"has_more":false}`)
	})
	if _, err := c.SearchNotes(context.Background(), "tag:work notebook:Inbox", ListOptions{Limit: 5}); err != nil {
		t.Fatal(err)
	}
	if qSeen != "tag:work notebook:Inbox" {
		t.Errorf("query = %q", qSeen)
	}
	if typeSeen != "note" {
		t.Errorf("type = %q, want note", typeSeen)
	}
}

func TestListEvents_CursorParam(t *testing.T) {
	var cursorSeen string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		cursorSeen = r.URL.Query().Get("cursor")
		_, _ = io.WriteString(w, `{"items":[],"cursor":42,"has_more":false}`)
	})
	p, err := c.ListEvents(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if cursorSeen != "10" {
		t.Errorf("cursor = %q, want 10", cursorSeen)
	}
	if p.Cursor != 42 {
		t.Errorf("response cursor = %d, want 42", p.Cursor)
	}
}

func TestDownloadResource_RawBytesAndContentType(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte{0x89, 'P', 'N', 'G'})
	})
	data, ct, err := c.DownloadResource(context.Background(), "rid")
	if err != nil {
		t.Fatal(err)
	}
	if ct != "image/png" {
		t.Errorf("content-type = %q", ct)
	}
	if len(data) != 4 || data[0] != 0x89 {
		t.Errorf("data = %v", data)
	}
}

func TestUploadResource_Multipart(t *testing.T) {
	var ctSeen string
	var bodyContains bool
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		ctSeen = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		bodyContains = strings.Contains(string(b), `"title":"hello"`) && strings.Contains(string(b), "PNGDATA")
		_, _ = io.WriteString(w, `{"id":"r1","title":"hello","mime":"image/png"}`)
	})
	r, err := c.UploadResource(context.Background(), []byte("PNGDATA"), "x.png", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if r.ID != "r1" {
		t.Errorf("ID = %q", r.ID)
	}
	if !strings.HasPrefix(ctSeen, "multipart/form-data") {
		t.Errorf("content-type = %q, want multipart/form-data", ctSeen)
	}
	if !bodyContains {
		t.Error("multipart body missing expected props/data parts")
	}
}

func TestUploadResource_FilenameRequired(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("server should not have been called")
	})
	if _, err := c.UploadResource(context.Background(), []byte("x"), "", "title"); err == nil {
		t.Fatal("expected error for empty filename")
	}
}

func TestTagNote_PostsNoteID(t *testing.T) {
	var bodySeen string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		b, _ := io.ReadAll(r.Body)
		bodySeen = string(b)
		w.WriteHeader(http.StatusOK)
	})
	if err := c.TagNote(context.Background(), "tag1", "note1"); err != nil {
		t.Fatal(err)
	}
	if bodySeen != `{"id":"note1"}` {
		t.Errorf("body = %q", bodySeen)
	}
}

func TestListFolderNotes_Endpoint(t *testing.T) {
	var pathSeen string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		pathSeen = r.URL.Path
		_, _ = io.WriteString(w, `{"items":[],"has_more":false}`)
	})
	if _, err := c.ListFolderNotes(context.Background(), "f1", ListOptions{}); err != nil {
		t.Fatal(err)
	}
	if pathSeen != "/folders/f1/notes" {
		t.Errorf("path = %q", pathSeen)
	}
}

func TestListNoteRevisions_FiltersByItemID(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"items":[
			{"id":"r1","item_id":"note1","created_time":1},
			{"id":"r2","item_id":"note2","created_time":2},
			{"id":"r3","item_id":"note1","created_time":3}
		],"has_more":false}`)
	})
	rs, err := c.ListNoteRevisions(context.Background(), "note1")
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) != 2 {
		t.Errorf("len = %d, want 2", len(rs))
	}
	for _, r := range rs {
		if r.ItemID != "note1" {
			t.Errorf("got revision for %q, want note1", r.ItemID)
		}
	}
}

func TestListQuery_SetsAllOptions(t *testing.T) {
	q := listQuery(ListOptions{Page: 3, Limit: 50, OrderBy: "updated_time", OrderDir: "ASC", Fields: []string{"id", "title"}})
	if q.Get("page") != "3" || q.Get("limit") != "50" || q.Get("order_by") != "updated_time" ||
		q.Get("order_dir") != "ASC" || q.Get("fields") != "id,title" {
		t.Errorf("listQuery = %v", q)
	}
}

func TestAPIError_StringFormatting(t *testing.T) {
	if got := (&APIError{StatusCode: 500}).Error(); got != "joplin: http 500" {
		t.Errorf("Error() = %q", got)
	}
	if got := (&APIError{StatusCode: 404, Message: "x"}).Error(); got != "joplin: http 404: x" {
		t.Errorf("Error() = %q", got)
	}
}

// Ensure JSON serialisation of UpdateNoteInput omits unset pointer fields.
// Defends against a regression that would resend zero-valued fields.
func TestUpdateNoteInput_OmitsUnsetFields(t *testing.T) {
	in := UpdateNoteInput{}
	b, _ := json.Marshal(in)
	if string(b) != "{}" {
		t.Errorf("marshalled = %q, want {}", string(b))
	}
}
