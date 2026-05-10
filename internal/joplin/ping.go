package joplin

import (
	"context"
	"net/http"
)

// Ping returns nil iff Joplin's Web Clipper service responds with 2xx on /ping.
// Used for connectivity / token-validity checks.
func (c *Client) Ping(ctx context.Context) error {
	return c.do(ctx, http.MethodGet, "/ping", nil, nil, nil)
}
