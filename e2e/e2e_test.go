// Package e2e contains end-to-end tests that exercise joplin-mcp against a
// real Joplin Desktop instance.
//
// Skipped unless JOPLIN_E2E=1. Requires JOPLIN_TOKEN; honours JOPLIN_BASE_URL.
// Every test cleans up the items it creates, even on failure.
package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

const e2eTimeout = 30 * time.Second

func setup(t *testing.T) (*joplin.Client, context.Context) {
	t.Helper()
	if os.Getenv("JOPLIN_E2E") != "1" {
		t.Skip("set JOPLIN_E2E=1 to run against a real Joplin Desktop")
	}
	token := strings.TrimSpace(os.Getenv("JOPLIN_TOKEN"))
	if token == "" {
		t.Fatal("JOPLIN_E2E=1 requires JOPLIN_TOKEN")
	}
	c, err := joplin.New(joplin.Options{
		Token:   token,
		BaseURL: strings.TrimSpace(os.Getenv("JOPLIN_BASE_URL")),
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), e2eTimeout)
	t.Cleanup(cancel)
	return c, ctx
}

// All e2e-created items get this prefix so we can recognise (and a human can
// purge) anything left behind by a crashed run.
func name(suffix string) string {
	return fmt.Sprintf("joplin-mcp-e2e-%d-%s", time.Now().UnixNano(), suffix)
}

// 1×1 transparent PNG. Used to exercise resource upload/download.
var tinyPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
	0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
	0x89, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
	0x42, 0x60, 0x82,
}

func TestE2E_Ping(t *testing.T) {
	c, ctx := setup(t)
	// Use a single cheap call to confirm reachability before we start
	// creating side effects.
	if _, err := c.ListFolders(ctx, joplin.ListOptions{Limit: 1}); err != nil {
		t.Fatalf("Joplin not reachable on the configured base URL: %v", err)
	}
}

func TestE2E_FolderCRUD(t *testing.T) {
	c, ctx := setup(t)
	title := name("folder")

	f, err := c.CreateFolder(ctx, joplin.CreateFolderInput{Title: title})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteFolder(context.Background(), f.ID, true) })

	if f.Title != title {
		t.Errorf("create returned title %q, want %q", f.Title, title)
	}

	got, err := c.GetFolder(ctx, f.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != f.ID || got.Title != title {
		t.Errorf("get mismatch: %+v", got)
	}

	renamed := title + "-renamed"
	upd, err := c.UpdateFolder(ctx, f.ID, joplin.UpdateFolderInput{Title: &renamed})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if upd.Title != renamed {
		t.Errorf("update returned title %q, want %q", upd.Title, renamed)
	}

	if err := c.DeleteFolder(ctx, f.ID, true); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := c.GetFolder(ctx, f.ID); !joplin.IsNotFound(err) {
		t.Errorf("expected 404 after delete, got %v", err)
	}
}

func TestE2E_NoteCRUD(t *testing.T) {
	c, ctx := setup(t)

	// Notes need a parent folder.
	folder, err := c.CreateFolder(ctx, joplin.CreateFolderInput{Title: name("notes-folder")})
	if err != nil {
		t.Fatalf("create folder: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteFolder(context.Background(), folder.ID, true) })

	title := name("note")
	body := "## hello\n\nfrom e2e"
	isTodo := true

	n, err := c.CreateNote(ctx, joplin.CreateNoteInput{
		Title:    title,
		Body:     body,
		ParentID: folder.ID,
		IsTodo:   &isTodo,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteNote(context.Background(), n.ID, true) })

	if n.Title != title {
		t.Errorf("create returned title %q", n.Title)
	}

	got, err := c.GetNote(ctx, n.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Body != body {
		t.Errorf("get body = %q, want %q", got.Body, body)
	}
	if !got.IsTodo {
		t.Error("get IsTodo = false, want true")
	}
	if got.ParentID != folder.ID {
		t.Errorf("get ParentID = %q, want %q", got.ParentID, folder.ID)
	}

	newTitle := title + "-updated"
	newBody := body + "\n\nupdated"
	if _, err := c.UpdateNote(ctx, n.ID, joplin.UpdateNoteInput{Title: &newTitle, Body: &newBody}); err != nil {
		t.Fatalf("update: %v", err)
	}
	got2, err := c.GetNote(ctx, n.ID)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if got2.Title != newTitle || got2.Body != newBody {
		t.Errorf("update did not persist: title=%q body=%q", got2.Title, got2.Body)
	}

	// list_notes_in_folder should find it.
	page, err := c.ListFolderNotes(ctx, folder.ID, joplin.ListOptions{Limit: 50})
	if err != nil {
		t.Fatalf("list folder notes: %v", err)
	}
	found := false
	for _, x := range page.Items {
		if x.ID == n.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("created note not found in list_notes_in_folder")
	}

	// Delete (trash) then delete again permanent.
	if err := c.DeleteNote(ctx, n.ID, false); err != nil {
		t.Fatalf("delete (trash): %v", err)
	}
	if err := c.DeleteNote(ctx, n.ID, true); err != nil && !joplin.IsNotFound(err) {
		t.Fatalf("delete (permanent): %v", err)
	}
}

func TestE2E_TagCRUD(t *testing.T) {
	c, ctx := setup(t)

	folder, err := c.CreateFolder(ctx, joplin.CreateFolderInput{Title: name("tags-folder")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.DeleteFolder(context.Background(), folder.ID, true) })

	note, err := c.CreateNote(ctx, joplin.CreateNoteInput{
		Title:    name("tagged-note"),
		Body:     "tag me",
		ParentID: folder.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.DeleteNote(context.Background(), note.ID, true) })

	tagTitle := strings.ToLower(name("tag"))
	tag, err := c.CreateTag(ctx, joplin.CreateTagInput{Title: tagTitle})
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteTag(context.Background(), tag.ID) })

	if err := c.TagNote(ctx, tag.ID, note.ID); err != nil {
		t.Fatalf("tag_note: %v", err)
	}

	tags, err := c.ListNoteTags(ctx, note.ID, joplin.ListOptions{Limit: 50})
	if err != nil {
		t.Fatalf("list note tags: %v", err)
	}
	found := false
	for _, x := range tags.Items {
		if x.ID == tag.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("attached tag not present on note")
	}

	notes, err := c.ListTagNotes(ctx, tag.ID, joplin.ListOptions{Limit: 50})
	if err != nil {
		t.Fatalf("list tag notes: %v", err)
	}
	found = false
	for _, x := range notes.Items {
		if x.ID == note.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("note not present in list_notes_with_tag")
	}

	if err := c.UntagNote(ctx, tag.ID, note.ID); err != nil {
		t.Fatalf("untag_note: %v", err)
	}

	if err := c.DeleteTag(ctx, tag.ID); err != nil {
		t.Fatalf("delete tag: %v", err)
	}
	if _, err := c.GetTag(ctx, tag.ID); !joplin.IsNotFound(err) {
		t.Errorf("expected 404 after delete, got %v", err)
	}
}

func TestE2E_ResourceCRUD(t *testing.T) {
	c, ctx := setup(t)

	uploadTitle := name("resource")
	res, err := c.UploadResource(ctx, tinyPNG, "tiny.png", uploadTitle)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteResource(context.Background(), res.ID) })

	if res.ID == "" {
		t.Fatal("upload returned empty ID")
	}

	meta, err := c.GetResource(ctx, res.ID)
	if err != nil {
		t.Fatalf("get metadata: %v", err)
	}
	if meta.Mime != "image/png" {
		t.Errorf("mime = %q, want image/png", meta.Mime)
	}
	if meta.EncryptionApplied {
		t.Error("freshly uploaded resource should not be encrypted")
	}

	bytes, ct, err := c.DownloadResource(ctx, res.ID)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	if !strings.HasPrefix(ct, "image/") {
		t.Errorf("content-type = %q, want image/*", ct)
	}
	if len(bytes) != len(tinyPNG) {
		t.Errorf("downloaded %d bytes, uploaded %d", len(bytes), len(tinyPNG))
	}
	for i := range bytes {
		if bytes[i] != tinyPNG[i] {
			t.Errorf("byte %d mismatch: got %#x want %#x", i, bytes[i], tinyPNG[i])
			break
		}
	}

	if err := c.DeleteResource(ctx, res.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := c.GetResource(ctx, res.ID); !joplin.IsNotFound(err) {
		t.Errorf("expected 404 after delete, got %v", err)
	}
}

func TestE2E_Search(t *testing.T) {
	c, ctx := setup(t)

	folder, err := c.CreateFolder(ctx, joplin.CreateFolderInput{Title: name("search-folder")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.DeleteFolder(context.Background(), folder.ID, true) })

	// Use a unique token in the body so the search query can find this exact note.
	marker := name("search-marker")
	note, err := c.CreateNote(ctx, joplin.CreateNoteInput{
		Title:    name("search-note"),
		Body:     "needle " + marker + " haystack",
		ParentID: folder.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.DeleteNote(context.Background(), note.ID, true) })

	// Joplin's full-text index is async — give it a moment.
	deadline := time.Now().Add(15 * time.Second)
	var page joplin.Page[joplin.Note]
	for time.Now().Before(deadline) {
		page, err = c.SearchNotes(ctx, marker, joplin.ListOptions{Limit: 10})
		if err == nil && len(page.Items) > 0 {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	found := false
	for _, x := range page.Items {
		if x.ID == note.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("search for %q did not return the created note (got %d hits)", marker, len(page.Items))
	}
}

func TestE2E_Events(t *testing.T) {
	c, ctx := setup(t)

	// Capture the cursor before our mutations.
	before, err := c.ListEvents(ctx, "")
	if err != nil {
		t.Fatalf("list events: %v", err)
	}

	folder, err := c.CreateFolder(ctx, joplin.CreateFolderInput{Title: name("events-folder")})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.DeleteFolder(context.Background(), folder.ID, true) }()

	note, err := c.CreateNote(ctx, joplin.CreateNoteInput{
		Title:    name("events-note"),
		Body:     "trigger an event",
		ParentID: folder.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.DeleteNote(context.Background(), note.ID, true) }()

	// Joplin only emits events for *notes* (per the docs). It may take a moment.
	deadline := time.Now().Add(10 * time.Second)
	var saw bool
	for time.Now().Before(deadline) {
		after, err := c.ListEvents(ctx, before.Cursor)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range after.Items {
			if e.ItemID == note.ID {
				saw = true
				break
			}
		}
		if saw {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !saw {
		t.Logf("note create event not surfaced via /events within 10s (Joplin sometimes batches; this is informational, not a hard failure)")
	}
}

func TestE2E_Revisions(t *testing.T) {
	c, ctx := setup(t)
	// Just verify the endpoint works; new notes typically have no revisions
	// because Joplin generates them on a schedule. We don't depend on count.
	if _, err := c.ListRevisions(ctx, joplin.ListOptions{Limit: 1}); err != nil {
		t.Fatalf("list revisions: %v", err)
	}
}

// Sanity: make sure the e2e helper rejects a missing token.
func TestE2E_BadTokenRejected(t *testing.T) {
	if os.Getenv("JOPLIN_E2E") != "1" {
		t.Skip("set JOPLIN_E2E=1 to run end-to-end tests")
	}
	c, err := joplin.New(joplin.Options{Token: "definitely-not-a-real-token-padding-padding-padding"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = c.ListFolders(ctx, joplin.ListOptions{Limit: 1})
	if err == nil {
		t.Fatal("expected an error from a bogus token")
	}
	var apiErr *joplin.APIError
	if !errors.As(err, &apiErr) {
		t.Logf("non-APIError returned (likely network-level): %v", err)
		return
	}
	if apiErr.StatusCode != 403 && apiErr.StatusCode != 401 {
		t.Errorf("expected 401/403, got %d", apiErr.StatusCode)
	}
}
