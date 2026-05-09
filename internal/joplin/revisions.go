package joplin

import (
	"context"
	"net/http"
	"net/url"
)

// ListRevisions returns one page of all revisions across the database.
//
// Joplin's /revisions endpoint does not filter by note ID directly; callers
// that want a single note's history should use ListNoteRevisions.
func (c *Client) ListRevisions(ctx context.Context, opts ListOptions) (Page[Revision], error) {
	var p Page[Revision]
	if err := c.do(ctx, http.MethodGet, "/revisions", listQuery(opts), nil, &p); err != nil {
		return Page[Revision]{}, err
	}
	return p, nil
}

// GetRevision returns a single revision by ID.
func (c *Client) GetRevision(ctx context.Context, id string) (Revision, error) {
	var r Revision
	if err := c.do(ctx, http.MethodGet, "/revisions/"+url.PathEscape(id), nil, nil, &r); err != nil {
		return Revision{}, err
	}
	return r, nil
}

// ListNoteRevisions returns the revision history for a specific note. This is
// implemented client-side by walking /revisions and filtering on item_id;
// Joplin does not expose a direct per-note revisions endpoint.
func (c *Client) ListNoteRevisions(ctx context.Context, noteID string) ([]Revision, error) {
	all, err := CollectAll(ctx, func(ctx context.Context, page int) (Page[Revision], error) {
		return c.ListRevisions(ctx, ListOptions{Page: page, Limit: 100, OrderBy: "created_time", OrderDir: "DESC"})
	})
	if err != nil {
		return nil, err
	}
	out := all[:0]
	for _, r := range all {
		if r.ItemID == noteID {
			out = append(out, r)
		}
	}
	return out, nil
}
