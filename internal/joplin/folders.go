package joplin

import (
	"context"
	"net/http"
	"net/url"
)

// ListFolders returns one page of folders.
func (c *Client) ListFolders(ctx context.Context, opts ListOptions) (Page[Folder], error) {
	var p Page[Folder]
	if err := c.do(ctx, http.MethodGet, "/folders", listQuery(opts), nil, &p); err != nil {
		return Page[Folder]{}, err
	}
	return p, nil
}

// GetFolder returns a single folder by ID.
func (c *Client) GetFolder(ctx context.Context, id string) (Folder, error) {
	var f Folder
	if err := c.do(ctx, http.MethodGet, "/folders/"+url.PathEscape(id), nil, nil, &f); err != nil {
		return Folder{}, err
	}
	return f, nil
}

// CreateFolder creates a folder.
func (c *Client) CreateFolder(ctx context.Context, in CreateFolderInput) (Folder, error) {
	var f Folder
	if err := c.do(ctx, http.MethodPost, "/folders", nil, in, &f); err != nil {
		return Folder{}, err
	}
	return f, nil
}

// UpdateFolder applies a partial update.
func (c *Client) UpdateFolder(ctx context.Context, id string, in UpdateFolderInput) (Folder, error) {
	var f Folder
	if err := c.do(ctx, http.MethodPut, "/folders/"+url.PathEscape(id), nil, in, &f); err != nil {
		return Folder{}, err
	}
	return f, nil
}

// DeleteFolder deletes a folder. If permanent is true, the folder is purged
// rather than moved to trash.
func (c *Client) DeleteFolder(ctx context.Context, id string, permanent bool) error {
	q := url.Values{}
	if permanent {
		q.Set("permanent", "1")
	}
	return c.do(ctx, http.MethodDelete, "/folders/"+url.PathEscape(id), q, nil, nil)
}

// ListFolderNotes returns notes whose parent_id is the given folder.
func (c *Client) ListFolderNotes(ctx context.Context, folderID string, opts ListOptions) (Page[Note], error) {
	if len(opts.Fields) == 0 {
		opts.Fields = defaultNoteFields
	}
	var p Page[Note]
	if err := c.do(ctx, http.MethodGet, "/folders/"+url.PathEscape(folderID)+"/notes", listQuery(opts), nil, &p); err != nil {
		return Page[Note]{}, err
	}
	return p, nil
}
