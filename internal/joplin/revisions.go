package joplin

import (
	"context"
	"net/http"
	"net/url"
)

var defaultRevisionFields = []string{
	"id", "parent_id", "item_type", "item_id", "item_updated_time",
	"title_diff", "body_diff", "metadata_diff",
	"encryption_applied", "created_time", "updated_time",
}

// Joplin's /revisions endpoint does not filter by note ID directly; callers
// that want a single note's history should use ListNoteRevisions.
func (c *Client) ListRevisions(ctx context.Context, opts ListOptions) (Page[Revision], error) {
	if len(opts.Fields) == 0 {
		opts.Fields = defaultRevisionFields
	}
	var p Page[Revision]
	if err := c.do(ctx, http.MethodGet, "/revisions", listQuery(opts), nil, &p); err != nil {
		return Page[Revision]{}, err
	}
	return p, nil
}

func (c *Client) GetRevision(ctx context.Context, id string) (Revision, error) {
	q := url.Values{}
	q.Set("fields", joinFields(defaultRevisionFields))
	var r Revision
	if err := c.do(ctx, http.MethodGet, "/revisions/"+url.PathEscape(id), q, nil, &r); err != nil {
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
