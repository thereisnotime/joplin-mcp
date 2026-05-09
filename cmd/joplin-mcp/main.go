// Command joplin-mcp is a Model Context Protocol server for Joplin notes.
//
// It wraps Joplin Desktop's local Web Clipper REST API and exposes notes,
// folders, tags, search, resources, events, and revisions as MCP tools over
// stdio.
package main

import (
	"fmt"
	"os"

	"github.com/thereisnotime/joplin-mcp/internal/version"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Println(version.String())
		return
	}

	// MCP server wiring lands in M2.
	fmt.Fprintln(os.Stderr, "joplin-mcp: server bootstrap pending (M2)")
	os.Exit(1)
}
