// Package joplin is a typed client for Joplin Desktop's local Web Clipper REST API.
//
// The client wraps every endpoint joplin-mcp surfaces to LLM clients: notes,
// folders, tags, resources, search, change events, and note revisions. Every
// declared field is populated from the API response (the prior-art Python
// MCP server silently drops half of them); unknown fields are ignored so the
// client survives Joplin schema additions.
package joplin

// Page is a single page of a paginated Joplin response.
type Page[T any] struct {
	Items   []T  `json:"items"`
	HasMore bool `json:"has_more"`
}

// Note is a Joplin note. Field set follows the public Web Clipper API.
//
// Reference: https://joplinapp.org/help/api/references/rest_api/#notes
type Note struct {
	ID                   string  `json:"id"`
	ParentID             string  `json:"parent_id,omitempty"`
	Title                string  `json:"title"`
	Body                 string  `json:"body,omitempty"`
	CreatedTime          int64   `json:"created_time,omitempty"`
	UpdatedTime          int64   `json:"updated_time,omitempty"`
	IsConflict           Boolish `json:"is_conflict,omitempty"`
	Latitude             string  `json:"latitude,omitempty"`
	Longitude            string  `json:"longitude,omitempty"`
	Altitude             string  `json:"altitude,omitempty"`
	Author               string  `json:"author,omitempty"`
	SourceURL            string  `json:"source_url,omitempty"`
	IsTodo               Boolish `json:"is_todo,omitempty"`
	TodoDue              int64   `json:"todo_due,omitempty"`
	TodoCompleted        int64   `json:"todo_completed,omitempty"`
	Source               string  `json:"source,omitempty"`
	SourceApplication    string  `json:"source_application,omitempty"`
	ApplicationData      string  `json:"application_data,omitempty"`
	Order                int64   `json:"order,omitempty"`
	UserCreatedTime      int64   `json:"user_created_time,omitempty"`
	UserUpdatedTime      int64   `json:"user_updated_time,omitempty"`
	EncryptionCipherText string  `json:"encryption_cipher_text,omitempty"`
	EncryptionApplied    Boolish `json:"encryption_applied"`
	MarkupLanguage       int     `json:"markup_language,omitempty"`
	IsShared             Boolish `json:"is_shared,omitempty"`
	ShareID              string  `json:"share_id,omitempty"`
	ConflictOriginalID   string  `json:"conflict_original_id,omitempty"`
	MasterKeyID          string  `json:"master_key_id,omitempty"`
	UserData             string  `json:"user_data,omitempty"`
	Deleted              Boolish `json:"deleted,omitempty"`
	DeletedTime          int64   `json:"deleted_time,omitempty"`
	BodyHTML             string  `json:"body_html,omitempty"`
	BaseURL              string  `json:"base_url,omitempty"`
	ImageDataURL         string  `json:"image_data_url,omitempty"`
	CropRect             string  `json:"crop_rect,omitempty"`
}

// Folder is a Joplin notebook.
//
// Reference: https://joplinapp.org/help/api/references/rest_api/#folders
type Folder struct {
	ID                   string  `json:"id"`
	ParentID             string  `json:"parent_id,omitempty"`
	Title                string  `json:"title"`
	CreatedTime          int64   `json:"created_time,omitempty"`
	UpdatedTime          int64   `json:"updated_time,omitempty"`
	UserCreatedTime      int64   `json:"user_created_time,omitempty"`
	UserUpdatedTime      int64   `json:"user_updated_time,omitempty"`
	EncryptionCipherText string  `json:"encryption_cipher_text,omitempty"`
	EncryptionApplied    Boolish `json:"encryption_applied"`
	IsShared             Boolish `json:"is_shared,omitempty"`
	ShareID              string  `json:"share_id,omitempty"`
	MasterKeyID          string  `json:"master_key_id,omitempty"`
	Icon                 string  `json:"icon,omitempty"`
	UserData             string  `json:"user_data,omitempty"`
	Deleted              Boolish `json:"deleted,omitempty"`
	DeletedTime          int64   `json:"deleted_time,omitempty"`
}

// Tag is a Joplin tag.
//
// Reference: https://joplinapp.org/help/api/references/rest_api/#tags
type Tag struct {
	ID                   string  `json:"id"`
	ParentID             string  `json:"parent_id,omitempty"`
	Title                string  `json:"title"`
	CreatedTime          int64   `json:"created_time,omitempty"`
	UpdatedTime          int64   `json:"updated_time,omitempty"`
	UserCreatedTime      int64   `json:"user_created_time,omitempty"`
	UserUpdatedTime      int64   `json:"user_updated_time,omitempty"`
	EncryptionCipherText string  `json:"encryption_cipher_text,omitempty"`
	EncryptionApplied    Boolish `json:"encryption_applied"`
	IsShared             Boolish `json:"is_shared,omitempty"`
}

// Resource is a Joplin attachment (image, file, etc.).
//
// Reference: https://joplinapp.org/help/api/references/rest_api/#resources
type Resource struct {
	ID                      string  `json:"id"`
	Title                   string  `json:"title,omitempty"`
	Mime                    string  `json:"mime,omitempty"`
	Filename                string  `json:"filename,omitempty"`
	CreatedTime             int64   `json:"created_time,omitempty"`
	UpdatedTime             int64   `json:"updated_time,omitempty"`
	UserCreatedTime         int64   `json:"user_created_time,omitempty"`
	UserUpdatedTime         int64   `json:"user_updated_time,omitempty"`
	FileExtension           string  `json:"file_extension,omitempty"`
	EncryptionCipherText    string  `json:"encryption_cipher_text,omitempty"`
	EncryptionApplied       Boolish `json:"encryption_applied"`
	EncryptionBlobEncrypted Boolish `json:"encryption_blob_encrypted,omitempty"`
	Size                    int64   `json:"size,omitempty"`
	IsShared                Boolish `json:"is_shared,omitempty"`
	ShareID                 string  `json:"share_id,omitempty"`
	MasterKeyID             string  `json:"master_key_id,omitempty"`
	UserData                string  `json:"user_data,omitempty"`
}

// Revision is a Joplin note revision.
//
// Reference: https://joplinapp.org/help/api/references/rest_api/#revisions
type Revision struct {
	ID                   string  `json:"id"`
	ParentID             string  `json:"parent_id,omitempty"`
	ItemType             int     `json:"item_type,omitempty"`
	ItemID               string  `json:"item_id,omitempty"`
	ItemUpdatedTime      int64   `json:"item_updated_time,omitempty"`
	TitleDiff            string  `json:"title_diff,omitempty"`
	BodyDiff             string  `json:"body_diff,omitempty"`
	MetadataDiff         string  `json:"metadata_diff,omitempty"`
	EncryptionCipherText string  `json:"encryption_cipher_text,omitempty"`
	EncryptionApplied    Boolish `json:"encryption_applied"`
	UpdatedTime          int64   `json:"updated_time,omitempty"`
	CreatedTime          int64   `json:"created_time,omitempty"`
}

// Event is a Joplin change event.
//
// Reference: https://joplinapp.org/help/api/references/rest_api/#events
type Event struct {
	ID               int64  `json:"id"`
	ItemType         int    `json:"item_type,omitempty"`
	ItemID           string `json:"item_id,omitempty"`
	Type             int    `json:"type,omitempty"` // 1=create, 2=update, 3=delete
	CreatedTime      int64  `json:"created_time,omitempty"`
	Source           int    `json:"source,omitempty"`
	BeforeChangeItem string `json:"before_change_item,omitempty"`
}

// Cursor is a quoted string in Joplin's wire format, so we keep it as opaque
// text — callers should round-trip it, not do arithmetic.
type EventsPage struct {
	Items   []Event `json:"items"`
	Cursor  string  `json:"cursor"`
	HasMore bool    `json:"has_more"`
}

// CreateNoteInput is the writable subset of fields for creating a note.
type CreateNoteInput struct {
	Title     string `json:"title,omitempty"`
	Body      string `json:"body,omitempty"`
	ParentID  string `json:"parent_id,omitempty"`
	IsTodo    *bool  `json:"is_todo,omitempty"`
	SourceURL string `json:"source_url,omitempty"`
	Author    string `json:"author,omitempty"`
}

// UpdateNoteInput is the writable subset of fields for updating a note.
// Pointer fields make it possible to distinguish "not set" from "set to zero".
type UpdateNoteInput struct {
	Title         *string `json:"title,omitempty"`
	Body          *string `json:"body,omitempty"`
	ParentID      *string `json:"parent_id,omitempty"`
	IsTodo        *bool   `json:"is_todo,omitempty"`
	TodoCompleted *int64  `json:"todo_completed,omitempty"`
	TodoDue       *int64  `json:"todo_due,omitempty"`
}

// CreateFolderInput is the writable subset for creating a folder.
type CreateFolderInput struct {
	Title    string `json:"title,omitempty"`
	ParentID string `json:"parent_id,omitempty"`
}

// UpdateFolderInput is the writable subset for updating a folder.
type UpdateFolderInput struct {
	Title    *string `json:"title,omitempty"`
	ParentID *string `json:"parent_id,omitempty"`
}

// CreateTagInput is the writable subset for creating a tag.
type CreateTagInput struct {
	Title string `json:"title"`
}

// ListOptions controls pagination, sorting, and field selection on list endpoints.
type ListOptions struct {
	Page     int      // 1-based; 0 means default (1)
	Limit    int      // 0 means Joplin default (100)
	OrderBy  string   // e.g. "updated_time"
	OrderDir string   // "ASC" or "DESC"
	Fields   []string // optional field selector
}
