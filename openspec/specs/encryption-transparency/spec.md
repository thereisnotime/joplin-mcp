# encryption-transparency

## Purpose

Cross-cutting behaviour that ensures every tool response surfaces the
encryption state of every item it touches. The server never decrypts; it
only reports what Joplin Desktop has already decrypted, and is honest about
items it could not.

## Requirements

### Requirement: Per-item encryption state
The server SHALL include the `encryption_applied` boolean on every item it returns (notes, folders, tags, resources, revisions), and when that flag is true the response SHALL also include the `master_key_id` of the key required to decrypt the item.

#### Scenario: returning a decrypted note
- **WHEN** the server returns a note whose Joplin row has `encryption_applied=0`
- **THEN** the tool response includes `encryption_applied: false` for that note

#### Scenario: returning an encrypted note
- **WHEN** the server returns a note whose Joplin row has `encryption_applied=1`
- **THEN** the tool response includes `encryption_applied: true` and the
  `master_key_id` field, and does NOT include the empty `body` field as if it were
  the real note content

### Requirement: Skipped-item count on list and search responses
List and search responses SHALL include an `encrypted_items_skipped` integer that reports how many items in the response were returned in encrypted form.

#### Scenario: list with mixed encryption state
- **WHEN** a list of 12 notes contains 3 still-encrypted notes
- **THEN** the response carries `encrypted_items_skipped: 3` so the LLM can inform
  the user

### Requirement: No silent decryption attempts
The server SHALL NOT attempt to decrypt items, hold master keys, or accept passphrases, and SHALL only surface what the Joplin Desktop client has already decrypted.

#### Scenario: no decryption code path
- **WHEN** any tool encounters an item with `encryption_applied=true`
- **THEN** the server passes the encryption metadata through unchanged and never
  invokes any decryption routine

### Requirement: Resource download safety
`download_resource` SHALL refuse to return ciphertext bytes silently, and when the target resource is still encrypted the tool SHALL return an explicit error indicating the encryption state instead of the raw blob.

#### Scenario: refusing to download an encrypted resource
- **WHEN** an MCP client calls `download_resource` for a resource whose metadata
  reports `encryption_applied=true`
- **THEN** the tool returns an error of the form
  "resource is encrypted on this device; unlock Joplin and retry" and does NOT
  return the encrypted bytes
