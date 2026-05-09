package joplin

import (
	"context"
	"net/http"
)

// SearchType narrows what /search returns. Joplin defaults to notes when unset.
type SearchType string

const (
	SearchTypeNote     SearchType = "note"
	SearchTypeFolder   SearchType = "folder"
	SearchTypeResource SearchType = "resource"
	SearchTypeTag      SearchType = "tag"
)

// SearchNotes runs a full-text/Joplin-syntax search and returns matching notes.
func (c *Client) SearchNotes(ctx context.Context, query string, opts ListOptions) (Page[Note], error) {
	q := listQuery(opts)
	q.Set("query", query)
	q.Set("type", string(SearchTypeNote))
	if len(opts.Fields) == 0 {
		q.Set("fields", joinFields(defaultNoteFields))
	}
	var p Page[Note]
	if err := c.do(ctx, http.MethodGet, "/search", q, nil, &p); err != nil {
		return Page[Note]{}, err
	}
	return p, nil
}

// SearchFolders runs a search restricted to folders.
func (c *Client) SearchFolders(ctx context.Context, query string, opts ListOptions) (Page[Folder], error) {
	q := listQuery(opts)
	q.Set("query", query)
	q.Set("type", string(SearchTypeFolder))
	var p Page[Folder]
	if err := c.do(ctx, http.MethodGet, "/search", q, nil, &p); err != nil {
		return Page[Folder]{}, err
	}
	return p, nil
}

// SearchTags runs a search restricted to tags.
func (c *Client) SearchTags(ctx context.Context, query string, opts ListOptions) (Page[Tag], error) {
	q := listQuery(opts)
	q.Set("query", query)
	q.Set("type", string(SearchTypeTag))
	var p Page[Tag]
	if err := c.do(ctx, http.MethodGet, "/search", q, nil, &p); err != nil {
		return Page[Tag]{}, err
	}
	return p, nil
}

// SearchResources runs a search restricted to resources.
func (c *Client) SearchResources(ctx context.Context, query string, opts ListOptions) (Page[Resource], error) {
	q := listQuery(opts)
	q.Set("query", query)
	q.Set("type", string(SearchTypeResource))
	var p Page[Resource]
	if err := c.do(ctx, http.MethodGet, "/search", q, nil, &p); err != nil {
		return Page[Resource]{}, err
	}
	return p, nil
}
