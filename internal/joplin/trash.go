package joplin

import (
	"context"
	"net/http"
	"net/url"
)

// trashNoteFields requests the columns we need to identify a trashed item
// and decide what to do with it; deleted_time must be present for filtering.
var trashNoteFields = []string{
	"id", "parent_id", "title", "deleted_time", "encryption_applied",
	"created_time", "updated_time",
}

// ListTrashedNotes returns one page of notes whose deleted_time is non-zero
// (i.e. in the trash). Joplin's REST API has no dedicated trash endpoint;
// we list with include_deleted=1 and filter client-side.
func (c *Client) ListTrashedNotes(ctx context.Context, opts ListOptions) (Page[Note], error) {
	if len(opts.Fields) == 0 {
		opts.Fields = trashNoteFields
	}
	q := listQuery(opts)
	q.Set("include_deleted", "1")
	var p Page[Note]
	if err := c.do(ctx, http.MethodGet, "/notes", q, nil, &p); err != nil {
		return Page[Note]{}, err
	}
	// Drop live notes that the include_deleted=1 listing also returns.
	out := Page[Note]{HasMore: p.HasMore}
	for _, n := range p.Items {
		if n.DeletedTime > 0 {
			out.Items = append(out.Items, n)
		}
	}
	return out, nil
}

// RestoreNote moves a note out of the trash by clearing deleted_time.
func (c *Client) RestoreNote(ctx context.Context, id string) (Note, error) {
	zero := int64(0)
	body := struct {
		DeletedTime *int64 `json:"deleted_time"`
	}{DeletedTime: &zero}
	var n Note
	if err := c.do(ctx, http.MethodPut, "/notes/"+url.PathEscape(id), nil, body, &n); err != nil {
		return Note{}, err
	}
	return n, nil
}
