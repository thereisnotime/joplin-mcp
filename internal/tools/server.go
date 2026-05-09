package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

func New(c *joplin.Client, version string) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{Name: "joplin-mcp", Version: version}, nil)
	registerNoteTools(srv, c)
	registerFolderTools(srv, c)
	registerTagTools(srv, c)
	registerSearchTools(srv, c)
	registerResourceTools(srv, c)
	registerEventTools(srv, c)
	registerRevisionTools(srv, c)
	return srv
}
