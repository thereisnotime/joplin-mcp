package joplin

import (
	"context"
	"net/http"
	"net/url"
)

// MasterKey is the read-only metadata Joplin exposes for an encryption master
// key. The encrypted-content blob and any private material are intentionally
// not surfaced — joplin-mcp never decrypts.
type MasterKey struct {
	ID                string `json:"id"`
	SourceApplication string `json:"source_application,omitempty"`
	EncryptionMethod  int    `json:"encryption_method,omitempty"`
	Checksum          string `json:"checksum,omitempty"`
	Hint              string `json:"hint,omitempty"`
	Enabled           int    `json:"enabled,omitempty"`
	CreatedTime       int64  `json:"created_time,omitempty"`
	UpdatedTime       int64  `json:"updated_time,omitempty"`
}

var defaultMasterKeyFields = []string{
	"id", "source_application", "encryption_method", "checksum",
	"hint", "enabled", "created_time", "updated_time",
}

// ListMasterKeys returns one page of master-key metadata.
func (c *Client) ListMasterKeys(ctx context.Context, opts ListOptions) (Page[MasterKey], error) {
	if len(opts.Fields) == 0 {
		opts.Fields = defaultMasterKeyFields
	}
	var p Page[MasterKey]
	if err := c.do(ctx, http.MethodGet, "/master_keys", listQuery(opts), nil, &p); err != nil {
		return Page[MasterKey]{}, err
	}
	return p, nil
}

// GetMasterKey returns metadata for one master key by ID.
func (c *Client) GetMasterKey(ctx context.Context, id string) (MasterKey, error) {
	q := url.Values{}
	q.Set("fields", joinFields(defaultMasterKeyFields))
	var k MasterKey
	if err := c.do(ctx, http.MethodGet, "/master_keys/"+url.PathEscape(id), q, nil, &k); err != nil {
		return MasterKey{}, err
	}
	return k, nil
}
