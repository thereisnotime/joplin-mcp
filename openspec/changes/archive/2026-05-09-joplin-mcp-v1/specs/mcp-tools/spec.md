## ADDED Requirements

### Requirement: Stdio MCP server
The binary SHALL expose its tools via the official Go MCP SDK's stdio transport without writing any non-protocol bytes to stdout.

#### Scenario: stdout is reserved for the protocol
- **WHEN** the server starts and processes any tool call
- **THEN** all log output is written to stderr and only MCP protocol frames appear
  on stdout

### Requirement: Note tools
The server SHALL expose tools that allow listing, getting, creating, updating, and deleting Joplin notes, plus a convenience tool that returns a note together with its tags and resources in a single response.

#### Scenario: creating a note
- **WHEN** an MCP client calls the `create_note` tool with a title and body
- **THEN** the server creates the note in Joplin and returns the created note's ID

#### Scenario: get_note_with_context returns merged data
- **WHEN** an MCP client calls `get_note_with_context` for an existing note
- **THEN** the response includes the note body, the list of tag names, and the list
  of resource IDs attached to the note, fetched in parallel

### Requirement: Folder tools
The server SHALL expose tools to list, get, create, update, and delete folders (notebooks), and to list the notes within a folder.

#### Scenario: listing notes in a folder
- **WHEN** an MCP client calls `list_notes_in_folder` with a folder ID
- **THEN** the server returns all notes whose `parent_id` matches that folder

### Requirement: Tag tools
The server SHALL expose tools to list, get, create, and delete tags, attach and detach tags from notes, and list all notes with a given tag.

#### Scenario: tagging a note
- **WHEN** an MCP client calls `tag_note` with a note ID and a tag ID
- **THEN** the server attaches the tag to the note via Joplin's API

### Requirement: Search tool
The server SHALL expose a single `search` tool that accepts Joplin's full search query syntax and returns a paginated list of matching items.

#### Scenario: searching with Joplin query syntax
- **WHEN** an MCP client calls `search` with the query `tag:work notebook:Inbox`
- **THEN** the server forwards the query to `/search` and returns matching notes

### Requirement: Resource tools
The server SHALL expose tools to list resources, fetch a resource's metadata, download its binary content, upload a new resource, and delete a resource.

#### Scenario: downloading a resource
- **WHEN** an MCP client calls `download_resource` with a resource ID for a
  decrypted resource
- **THEN** the server returns the resource's bytes together with its MIME type

### Requirement: Events tool
The server SHALL expose a tool that returns Joplin change events since a supplied cursor, allowing the client to discover what changed between sessions.

#### Scenario: listing changes since a cursor
- **WHEN** an MCP client calls `list_changes_since` with a previously returned
  cursor
- **THEN** the server returns all events with a higher cursor value and a new cursor
  the caller can pass on the next call

### Requirement: Revision tools
The server SHALL expose tools to list a note's revision history and fetch the content of a specific revision.

#### Scenario: viewing a revision
- **WHEN** an MCP client calls `get_revision` with a revision ID
- **THEN** the server returns the revision's body and timestamp
