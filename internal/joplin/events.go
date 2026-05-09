package joplin

import (
	"context"
	"net/http"
	"net/url"
)

// ListEvents returns Joplin change events newer than the supplied cursor. An
// empty cursor returns events from the beginning of the available history.
// Cursors are opaque strings; pass back what the previous response returned.
func (c *Client) ListEvents(ctx context.Context, cursor string) (EventsPage, error) {
	q := url.Values{}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var p EventsPage
	if err := c.do(ctx, http.MethodGet, "/events", q, nil, &p); err != nil {
		return EventsPage{}, err
	}
	return p, nil
}
