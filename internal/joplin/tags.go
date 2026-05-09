package joplin

import (
	"context"
	"net/http"
	"net/url"
)

var defaultTagFields = []string{
	"id", "parent_id", "title", "created_time", "updated_time",
	"encryption_applied", "is_shared",
}

func (c *Client) ListTags(ctx context.Context, opts ListOptions) (Page[Tag], error) {
	if len(opts.Fields) == 0 {
		opts.Fields = defaultTagFields
	}
	var p Page[Tag]
	if err := c.do(ctx, http.MethodGet, "/tags", listQuery(opts), nil, &p); err != nil {
		return Page[Tag]{}, err
	}
	return p, nil
}

func (c *Client) GetTag(ctx context.Context, id string) (Tag, error) {
	q := url.Values{}
	q.Set("fields", joinFields(defaultTagFields))
	var t Tag
	if err := c.do(ctx, http.MethodGet, "/tags/"+url.PathEscape(id), q, nil, &t); err != nil {
		return Tag{}, err
	}
	return t, nil
}

// CreateTag creates a tag.
func (c *Client) CreateTag(ctx context.Context, in CreateTagInput) (Tag, error) {
	var t Tag
	if err := c.do(ctx, http.MethodPost, "/tags", nil, in, &t); err != nil {
		return Tag{}, err
	}
	return t, nil
}

// DeleteTag deletes a tag.
func (c *Client) DeleteTag(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/tags/"+url.PathEscape(id), nil, nil, nil)
}

// TagNote attaches the tag to the note.
func (c *Client) TagNote(ctx context.Context, tagID, noteID string) error {
	body := struct {
		ID string `json:"id"`
	}{ID: noteID}
	return c.do(ctx, http.MethodPost, "/tags/"+url.PathEscape(tagID)+"/notes", nil, body, nil)
}

// UntagNote detaches the tag from the note.
func (c *Client) UntagNote(ctx context.Context, tagID, noteID string) error {
	return c.do(ctx, http.MethodDelete, "/tags/"+url.PathEscape(tagID)+"/notes/"+url.PathEscape(noteID), nil, nil, nil)
}

// ListTagNotes returns notes that have the given tag.
func (c *Client) ListTagNotes(ctx context.Context, tagID string, opts ListOptions) (Page[Note], error) {
	if len(opts.Fields) == 0 {
		opts.Fields = defaultNoteFields
	}
	var p Page[Note]
	if err := c.do(ctx, http.MethodGet, "/tags/"+url.PathEscape(tagID)+"/notes", listQuery(opts), nil, &p); err != nil {
		return Page[Note]{}, err
	}
	return p, nil
}
