package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

// 10 MiB. Larger blobs blow up an LLM's context budget and rarely make sense
// in a tool response. Override via JOPLIN_MAX_RESOURCE_BYTES.
const DefaultMaxResourceBytes int64 = 10 * 1024 * 1024

type Options struct {
	Version          string
	MaxResourceBytes int64
}

func New(c *joplin.Client, opts Options) *mcp.Server {
	if opts.MaxResourceBytes <= 0 {
		opts.MaxResourceBytes = DefaultMaxResourceBytes
	}
	srv := mcp.NewServer(&mcp.Implementation{Name: "joplin-mcp", Version: opts.Version}, nil)
	registerNoteTools(srv, c)
	registerFolderTools(srv, c)
	registerTagTools(srv, c)
	registerSearchTools(srv, c)
	registerResourceTools(srv, c, opts.MaxResourceBytes)
	registerEventTools(srv, c)
	registerRevisionTools(srv, c)
	return srv
}
