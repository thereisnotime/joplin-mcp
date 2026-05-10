package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

// 50 MiB. Most LLM clients can handle this much in a single tool response,
// and real-world Joplin attachments routinely exceed 10 MiB (PDFs, scans).
// Override via JOPLIN_MAX_RESOURCE_BYTES if your client has a tighter budget.
const DefaultMaxResourceBytes int64 = 50 * 1024 * 1024

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
	registerAttachTools(srv, c, opts.MaxResourceBytes)
	registerEventTools(srv, c)
	registerRevisionTools(srv, c)
	registerLinkTools(srv, c)
	registerBulkTools(srv, c)
	registerTrashTools(srv, c)
	return srv
}
