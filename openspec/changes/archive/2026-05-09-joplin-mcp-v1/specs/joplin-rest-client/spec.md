## ADDED Requirements

### Requirement: Authenticated localhost client
The client SHALL authenticate to Joplin's Web Clipper API by passing the configured token as a `?token=` query parameter on every request, and SHALL default to the local-only base URL `http://localhost:41184`.

#### Scenario: token attached to every request
- **WHEN** any client method issues an HTTP request
- **THEN** the request URL includes the configured token as a `token` query parameter
  AND no `Authorization` header is set

#### Scenario: configurable base URL
- **WHEN** the client is constructed with a non-default `BaseURL`
- **THEN** all subsequent requests target that base URL

### Requirement: Bounded request lifetime
Every client method SHALL accept a `context.Context`, propagate it to the underlying HTTP request, and respect a default 10-second timeout when the caller does not set one.

#### Scenario: context cancellation aborts a request
- **WHEN** the caller cancels the supplied context
- **THEN** the in-flight HTTP request is aborted and the method returns
  `context.Canceled`

#### Scenario: default timeout applied
- **WHEN** Joplin Desktop does not respond within the default timeout window
- **THEN** the method returns a timeout error rather than blocking indefinitely

### Requirement: Full type coverage
Response types SHALL include every documented field returned by the Joplin REST API for the corresponding resource, and unknown fields SHALL be ignored without failing deserialisation.

#### Scenario: declared fields are populated
- **WHEN** the API returns a note with `parent_id`, `encryption_applied`, and
  `master_key_id`
- **THEN** the resulting `Note` value has each of those fields populated

#### Scenario: unknown fields do not break parsing
- **WHEN** a future Joplin release adds a previously-unknown JSON field to a response
- **THEN** the client deserialises the response successfully and ignores the field

### Requirement: Pagination helper
The client SHALL provide a generic helper that walks every page of a paginated endpoint and yields all items, while still allowing single-page access for callers that opt out.

#### Scenario: walking all pages
- **WHEN** the caller invokes the pagination helper for a list endpoint with more
  than one page of results
- **THEN** the helper issues subsequent requests with incremented `page` values until
  `has_more=false` and returns the concatenated items

### Requirement: Typed errors
Non-2xx HTTP responses SHALL be returned as a typed `*APIError` carrying the HTTP status code and the message body returned by Joplin.

#### Scenario: 404 from get_note
- **WHEN** the client calls `GetNote` with an unknown ID and Joplin returns 404
- **THEN** the method returns a non-nil `*APIError` whose `StatusCode` is 404
