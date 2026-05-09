package joplin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

var defaultResourceFields = []string{
	"id", "title", "mime", "filename", "file_extension", "size",
	"created_time", "updated_time", "encryption_applied",
	"encryption_blob_encrypted", "master_key_id", "is_shared",
}

func (c *Client) ListResources(ctx context.Context, opts ListOptions) (Page[Resource], error) {
	if len(opts.Fields) == 0 {
		opts.Fields = defaultResourceFields
	}
	var p Page[Resource]
	if err := c.do(ctx, http.MethodGet, "/resources", listQuery(opts), nil, &p); err != nil {
		return Page[Resource]{}, err
	}
	return p, nil
}

func (c *Client) GetResource(ctx context.Context, id string) (Resource, error) {
	q := url.Values{}
	q.Set("fields", joinFields(defaultResourceFields))
	var r Resource
	if err := c.do(ctx, http.MethodGet, "/resources/"+url.PathEscape(id), q, nil, &r); err != nil {
		return Resource{}, err
	}
	return r, nil
}

// DownloadResource fetches the raw file bytes of a resource together with its
// content type.
//
// NOTE: callers MUST check the resource's EncryptionApplied flag before calling
// this method. Joplin will happily return ciphertext bytes for an encrypted
// resource. The MCP tool layer enforces the safety invariant; this client
// method is intentionally low-level.
func (c *Client) DownloadResource(ctx context.Context, id string) (data []byte, contentType string, err error) {
	return c.rawGET(ctx, "/resources/"+url.PathEscape(id)+"/file", nil)
}

// UploadResource uploads a new resource. The data slice is the file bytes.
// The optional title is used in Joplin's resource panel; filename should
// include the extension so Joplin infers the MIME type correctly.
func (c *Client) UploadResource(ctx context.Context, data []byte, filename, title string) (Resource, error) {
	if filename == "" {
		return Resource{}, fmt.Errorf("joplin: filename is required for resource upload")
	}

	// Joplin's POST /resources expects multipart/form-data with a "data" file
	// part and a "props" JSON part.
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	props := struct {
		Title string `json:"title,omitempty"`
	}{Title: title}
	propsJSON, err := json.Marshal(props)
	if err != nil {
		return Resource{}, fmt.Errorf("joplin: marshal props: %w", err)
	}
	if err := w.WriteField("props", string(propsJSON)); err != nil {
		return Resource{}, fmt.Errorf("joplin: write props: %w", err)
	}
	fw, err := w.CreateFormFile("data", filename)
	if err != nil {
		return Resource{}, fmt.Errorf("joplin: create form file: %w", err)
	}
	if _, err := io.Copy(fw, bytes.NewReader(data)); err != nil {
		return Resource{}, fmt.Errorf("joplin: write file: %w", err)
	}
	if err := w.Close(); err != nil {
		return Resource{}, fmt.Errorf("joplin: close multipart writer: %w", err)
	}

	q := url.Values{}
	q.Set("token", c.token)
	u := c.baseURL + "/resources?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, body)
	if err != nil {
		return Resource{}, fmt.Errorf("joplin: build upload request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return Resource{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return Resource{}, &APIError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(msg))}
	}

	var r Resource
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return Resource{}, fmt.Errorf("joplin: decode upload response: %w", err)
	}
	return r, nil
}

func (c *Client) UpdateResource(ctx context.Context, id string, in UpdateResourceInput) (Resource, error) {
	var r Resource
	if err := c.do(ctx, http.MethodPut, "/resources/"+url.PathEscape(id), nil, in, &r); err != nil {
		return Resource{}, err
	}
	return r, nil
}

// ListResourceNotes returns the notes that reference the given resource.
// The reverse direction of /notes/:id/resources.
func (c *Client) ListResourceNotes(ctx context.Context, resourceID string, opts ListOptions) (Page[Note], error) {
	if len(opts.Fields) == 0 {
		opts.Fields = defaultNoteFields
	}
	var p Page[Note]
	if err := c.do(ctx, http.MethodGet, "/resources/"+url.PathEscape(resourceID)+"/notes", listQuery(opts), nil, &p); err != nil {
		return Page[Note]{}, err
	}
	return p, nil
}

// DeleteResource removes a resource.
func (c *Client) DeleteResource(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/resources/"+url.PathEscape(id), nil, nil, nil)
}
