// Package tools wires the joplin REST client to MCP tool handlers.
package tools

type NoteOut struct {
	ID                string `json:"id"`
	ParentID          string `json:"parent_id,omitempty"`
	Title             string `json:"title"`
	Body              string `json:"body,omitempty" jsonschema:"empty when item is encrypted"`
	IsTodo            bool   `json:"is_todo,omitempty"`
	TodoDue           int64  `json:"todo_due,omitempty"`
	TodoCompleted     int64  `json:"todo_completed,omitempty"`
	CreatedTime       int64  `json:"created_time,omitempty"`
	UpdatedTime       int64  `json:"updated_time,omitempty"`
	EncryptionApplied bool   `json:"encryption_applied" jsonschema:"true when item is still encrypted on the local device"`
	MasterKeyID       string `json:"master_key_id,omitempty" jsonschema:"master key id required to decrypt"`
	IsShared          bool   `json:"is_shared,omitempty"`
	MarkupLanguage    int    `json:"markup_language,omitempty"`
}

type FolderOut struct {
	ID                string `json:"id"`
	ParentID          string `json:"parent_id,omitempty"`
	Title             string `json:"title"`
	CreatedTime       int64  `json:"created_time,omitempty"`
	UpdatedTime       int64  `json:"updated_time,omitempty"`
	EncryptionApplied bool   `json:"encryption_applied"`
	MasterKeyID       string `json:"master_key_id,omitempty"`
	IsShared          bool   `json:"is_shared,omitempty"`
	Icon              string `json:"icon,omitempty"`
}

type TagOut struct {
	ID                string `json:"id"`
	ParentID          string `json:"parent_id,omitempty"`
	Title             string `json:"title"`
	EncryptionApplied bool   `json:"encryption_applied"`
	IsShared          bool   `json:"is_shared,omitempty"`
}

type ResourceOut struct {
	ID                      string `json:"id"`
	Title                   string `json:"title,omitempty"`
	Mime                    string `json:"mime,omitempty"`
	Filename                string `json:"filename,omitempty"`
	FileExtension           string `json:"file_extension,omitempty"`
	Size                    int64  `json:"size,omitempty"`
	CreatedTime             int64  `json:"created_time,omitempty"`
	UpdatedTime             int64  `json:"updated_time,omitempty"`
	EncryptionApplied       bool   `json:"encryption_applied"`
	EncryptionBlobEncrypted bool   `json:"encryption_blob_encrypted,omitempty" jsonschema:"true when the resource binary is still encrypted; download_resource will refuse"`
	MasterKeyID             string `json:"master_key_id,omitempty"`
	IsShared                bool   `json:"is_shared,omitempty"`
}

type RevisionOut struct {
	ID                string `json:"id"`
	ItemID            string `json:"item_id,omitempty"`
	ItemType          int    `json:"item_type,omitempty"`
	TitleDiff         string `json:"title_diff,omitempty"`
	BodyDiff          string `json:"body_diff,omitempty"`
	MetadataDiff      string `json:"metadata_diff,omitempty"`
	CreatedTime       int64  `json:"created_time,omitempty"`
	UpdatedTime       int64  `json:"updated_time,omitempty"`
	EncryptionApplied bool   `json:"encryption_applied"`
}

type EventOut struct {
	ID          int64  `json:"id"`
	ItemType    int    `json:"item_type,omitempty"`
	ItemID      string `json:"item_id,omitempty"`
	Type        int    `json:"type,omitempty" jsonschema:"event type — 1 create, 2 update, 3 delete"`
	CreatedTime int64  `json:"created_time,omitempty"`
}

type PageOut[T any] struct {
	Items                 []T  `json:"items"`
	HasMore               bool `json:"has_more"`
	EncryptedItemsSkipped int  `json:"encrypted_items_skipped" jsonschema:"how many items in this response were returned in encrypted form"`
}

type EventsOut struct {
	Items   []EventOut `json:"items"`
	Cursor  string     `json:"cursor" jsonschema:"opaque cursor; pass back as 'since' on the next call to resume"`
	HasMore bool       `json:"has_more"`
}

type NoteContextOut struct {
	Note      NoteOut       `json:"note"`
	Tags      []TagOut      `json:"tags"`
	Resources []ResourceOut `json:"resources"`
}

// NoArgs is the input type for tools that take no parameters. The MCP SDK
// can't synthesise a JSON schema from a literal struct{}, so we use this
// empty-but-named struct instead.
type NoArgs struct{}

// Bytes are base64-encoded so they can travel through MCP's JSON transport.
type DownloadResourceOut struct {
	Base64Data  string `json:"base64_data"`
	ContentType string `json:"content_type,omitempty"`
	Size        int    `json:"size_bytes"`
}
