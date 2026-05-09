package joplin

import (
	"context"
	"net/http"
	"net/url"
)

// defaultNoteFields are requested explicitly so list endpoints return the body
// (Joplin omits body from list responses by default).
var defaultNoteFields = []string{
	"id", "parent_id", "title", "body", "created_time", "updated_time",
	"is_todo", "todo_due", "todo_completed", "encryption_applied",
	"master_key_id", "markup_language", "is_shared", "user_data",
}

// ListNotes returns one page of notes.
func (c *Client) ListNotes(ctx context.Context, opts ListOptions) (Page[Note], error) {
	if len(opts.Fields) == 0 {
		opts.Fields = defaultNoteFields
	}
	var p Page[Note]
	if err := c.do(ctx, http.MethodGet, "/notes", listQuery(opts), nil, &p); err != nil {
		return Page[Note]{}, err
	}
	return p, nil
}

// GetNote returns a single note by ID.
func (c *Client) GetNote(ctx context.Context, id string) (Note, error) {
	q := url.Values{}
	q.Set("fields", joinFields(defaultNoteFields))
	var n Note
	if err := c.do(ctx, http.MethodGet, "/notes/"+url.PathEscape(id), q, nil, &n); err != nil {
		return Note{}, err
	}
	return n, nil
}

// CreateNote creates a note and returns the created object.
func (c *Client) CreateNote(ctx context.Context, in CreateNoteInput) (Note, error) {
	var n Note
	if err := c.do(ctx, http.MethodPost, "/notes", nil, in, &n); err != nil {
		return Note{}, err
	}
	return n, nil
}

// UpdateNote applies a partial update and returns the updated object.
func (c *Client) UpdateNote(ctx context.Context, id string, in UpdateNoteInput) (Note, error) {
	var n Note
	if err := c.do(ctx, http.MethodPut, "/notes/"+url.PathEscape(id), nil, in, &n); err != nil {
		return Note{}, err
	}
	return n, nil
}

// DeleteNote deletes a note. If permanent is true, the note is purged rather
// than moved to trash.
func (c *Client) DeleteNote(ctx context.Context, id string, permanent bool) error {
	q := url.Values{}
	if permanent {
		q.Set("permanent", "1")
	}
	return c.do(ctx, http.MethodDelete, "/notes/"+url.PathEscape(id), q, nil, nil)
}

// ListNoteTags returns all tags attached to the given note (paginated).
func (c *Client) ListNoteTags(ctx context.Context, noteID string, opts ListOptions) (Page[Tag], error) {
	var p Page[Tag]
	if err := c.do(ctx, http.MethodGet, "/notes/"+url.PathEscape(noteID)+"/tags", listQuery(opts), nil, &p); err != nil {
		return Page[Tag]{}, err
	}
	return p, nil
}

// ListNoteResources returns all resources attached to the given note (paginated).
func (c *Client) ListNoteResources(ctx context.Context, noteID string, opts ListOptions) (Page[Resource], error) {
	var p Page[Resource]
	if err := c.do(ctx, http.MethodGet, "/notes/"+url.PathEscape(noteID)+"/resources", listQuery(opts), nil, &p); err != nil {
		return Page[Resource]{}, err
	}
	return p, nil
}

func joinFields(fs []string) string {
	if len(fs) == 0 {
		return ""
	}
	out := fs[0]
	for _, f := range fs[1:] {
		out += "," + f
	}
	return out
}
