package tools

import "github.com/thereisnotime/joplin-mcp/internal/joplin"

// Drop the (empty) body when the source is encrypted — never silently
// return an empty body that the LLM might mistake for a real empty note.
func noteOut(n joplin.Note) NoteOut {
	out := NoteOut{
		ID:                n.ID,
		ParentID:          n.ParentID,
		Title:             n.Title,
		IsTodo:            n.IsTodo,
		TodoDue:           n.TodoDue,
		TodoCompleted:     n.TodoCompleted,
		CreatedTime:       n.CreatedTime,
		UpdatedTime:       n.UpdatedTime,
		EncryptionApplied: n.EncryptionApplied,
		MasterKeyID:       n.MasterKeyID,
		IsShared:          n.IsShared,
		MarkupLanguage:    n.MarkupLanguage,
	}
	if !n.EncryptionApplied {
		out.Body = n.Body
	}
	return out
}

func folderOut(f joplin.Folder) FolderOut {
	return FolderOut{
		ID:                f.ID,
		ParentID:          f.ParentID,
		Title:             f.Title,
		CreatedTime:       f.CreatedTime,
		UpdatedTime:       f.UpdatedTime,
		EncryptionApplied: f.EncryptionApplied,
		MasterKeyID:       f.MasterKeyID,
		IsShared:          f.IsShared,
		Icon:              f.Icon,
	}
}

func tagOut(t joplin.Tag) TagOut {
	return TagOut{
		ID:                t.ID,
		ParentID:          t.ParentID,
		Title:             t.Title,
		EncryptionApplied: t.EncryptionApplied,
		IsShared:          t.IsShared,
	}
}

func resourceOut(r joplin.Resource) ResourceOut {
	return ResourceOut{
		ID:                      r.ID,
		Title:                   r.Title,
		Mime:                    r.Mime,
		Filename:                r.Filename,
		FileExtension:           r.FileExtension,
		Size:                    r.Size,
		CreatedTime:             r.CreatedTime,
		UpdatedTime:             r.UpdatedTime,
		EncryptionApplied:       r.EncryptionApplied,
		EncryptionBlobEncrypted: r.EncryptionBlobEncrypted,
		MasterKeyID:             r.MasterKeyID,
		IsShared:                r.IsShared,
	}
}

func revisionOut(r joplin.Revision) RevisionOut {
	return RevisionOut{
		ID:                r.ID,
		ItemID:            r.ItemID,
		ItemType:          r.ItemType,
		TitleDiff:         r.TitleDiff,
		BodyDiff:          r.BodyDiff,
		MetadataDiff:      r.MetadataDiff,
		CreatedTime:       r.CreatedTime,
		UpdatedTime:       r.UpdatedTime,
		EncryptionApplied: r.EncryptionApplied,
	}
}

func eventOut(e joplin.Event) EventOut {
	return EventOut{
		ID:          e.ID,
		ItemType:    e.ItemType,
		ItemID:      e.ItemID,
		Type:        e.Type,
		CreatedTime: e.CreatedTime,
	}
}

func pageOf[Src any, Dst any](p joplin.Page[Src], conv func(Src) Dst, isEncrypted func(Src) bool) PageOut[Dst] {
	items := make([]Dst, 0, len(p.Items))
	skipped := 0
	for _, it := range p.Items {
		items = append(items, conv(it))
		if isEncrypted(it) {
			skipped++
		}
	}
	return PageOut[Dst]{Items: items, HasMore: p.HasMore, EncryptedItemsSkipped: skipped}
}

func notesPage(p joplin.Page[joplin.Note]) PageOut[NoteOut] {
	return pageOf(p, noteOut, func(n joplin.Note) bool { return n.EncryptionApplied })
}

func foldersPage(p joplin.Page[joplin.Folder]) PageOut[FolderOut] {
	return pageOf(p, folderOut, func(f joplin.Folder) bool { return f.EncryptionApplied })
}

func tagsPage(p joplin.Page[joplin.Tag]) PageOut[TagOut] {
	return pageOf(p, tagOut, func(t joplin.Tag) bool { return t.EncryptionApplied })
}

func resourcesPage(p joplin.Page[joplin.Resource]) PageOut[ResourceOut] {
	return pageOf(p, resourceOut, func(r joplin.Resource) bool { return r.EncryptionApplied })
}
