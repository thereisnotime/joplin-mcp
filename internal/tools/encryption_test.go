package tools

import (
	"testing"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

func TestNoteOut_DropsBodyWhenEncrypted(t *testing.T) {
	n := joplin.Note{
		ID:                "x",
		Title:             "hidden",
		Body:              "secret content",
		EncryptionApplied: joplin.Boolish(true),
		MasterKeyID:       "mk1",
	}
	out := noteOut(n)
	if out.Body != "" {
		t.Errorf("encrypted note body leaked: %q", out.Body)
	}
	if !out.EncryptionApplied {
		t.Error("EncryptionApplied not surfaced")
	}
	if out.MasterKeyID != "mk1" {
		t.Errorf("MasterKeyID = %q", out.MasterKeyID)
	}
}

func TestNoteOut_PreservesBodyWhenDecrypted(t *testing.T) {
	n := joplin.Note{ID: "x", Title: "ok", Body: "plain", EncryptionApplied: joplin.Boolish(false)}
	out := noteOut(n)
	if out.Body != "plain" {
		t.Errorf("Body = %q, want plain", out.Body)
	}
}

func TestPagesOf_SkipCount(t *testing.T) {
	p := joplin.Page[joplin.Note]{
		Items: []joplin.Note{
			{ID: "a", EncryptionApplied: joplin.Boolish(false)},
			{ID: "b", EncryptionApplied: joplin.Boolish(true)},
			{ID: "c", EncryptionApplied: joplin.Boolish(true)},
			{ID: "d", EncryptionApplied: joplin.Boolish(false)},
		},
		HasMore: true,
	}
	out := notesPage(p)
	if out.EncryptedItemsSkipped != 2 {
		t.Errorf("skipped = %d, want 2", out.EncryptedItemsSkipped)
	}
	if !out.HasMore {
		t.Error("HasMore not propagated")
	}
	if len(out.Items) != 4 {
		t.Errorf("items = %d, want 4", len(out.Items))
	}
}

func TestFolderTagResource_Outs(t *testing.T) {
	// One assertion per projection so each branch executes.
	if folderOut(joplin.Folder{ID: "f", Title: "F", IsShared: joplin.Boolish(true)}).IsShared != true {
		t.Error("folderOut IsShared lost")
	}
	if tagOut(joplin.Tag{ID: "t", Title: "T", EncryptionApplied: joplin.Boolish(true)}).EncryptionApplied != true {
		t.Error("tagOut EncryptionApplied lost")
	}
	if resourceOut(joplin.Resource{ID: "r", Mime: "image/png", EncryptionBlobEncrypted: joplin.Boolish(true)}).EncryptionBlobEncrypted != true {
		t.Error("resourceOut EncryptionBlobEncrypted lost")
	}
	if revisionOut(joplin.Revision{ID: "v", ItemID: "n"}).ItemID != "n" {
		t.Error("revisionOut ItemID lost")
	}
	if eventOut(joplin.Event{ID: 7, ItemID: "n"}).ID != 7 {
		t.Error("eventOut ID lost")
	}
}
