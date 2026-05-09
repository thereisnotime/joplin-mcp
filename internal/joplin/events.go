package joplin

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ListEvents returns Joplin change events with an ID greater than the supplied
// cursor. A cursor of zero returns events from the beginning of the available
// history.
func (c *Client) ListEvents(ctx context.Context, cursor int64) (EventsPage, error) {
	q := url.Values{}
	if cursor > 0 {
		q.Set("cursor", strconv.FormatInt(cursor, 10))
	}
	var p EventsPage
	if err := c.do(ctx, http.MethodGet, "/events", q, nil, &p); err != nil {
		return EventsPage{}, err
	}
	return p, nil
}
