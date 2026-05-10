package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

type LinkedNote struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
}

type LinksOut struct {
	NoteID string       `json:"note_id"`
	Items  []LinkedNote `json:"items"`
}

type ListOutboundLinksArgs struct {
	NoteID  string `json:"note_id"`
	Resolve bool   `json:"resolve_titles,omitempty" jsonschema:"if true, also fetch each linked item's title (extra round trips)"`
}

type ListBacklinksArgs struct {
	NoteID string `json:"note_id"`
	Page   int    `json:"page,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

func registerLinkTools(srv *mcp.Server, c *joplin.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_outbound_links",
		Description: "List the Joplin item IDs referenced from a note's body via :/<id> markdown links and image embeds. Set resolve_titles=true to also fetch each target's title.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ListOutboundLinksArgs) (*mcp.CallToolResult, LinksOut, error) {
		n, err := c.GetNote(ctx, args.NoteID)
		if err != nil {
			return nil, LinksOut{}, err
		}
		ids := joplin.ExtractLinkedIDs(n.Body)
		out := LinksOut{NoteID: args.NoteID, Items: make([]LinkedNote, 0, len(ids))}
		for _, id := range ids {
			item := LinkedNote{ID: id}
			if args.Resolve {
				// A linked ID could be a note OR a resource. Try note first
				// (cheap) and fall back to resource. Either way we just want
				// a human-readable title for the LLM's context.
				if note, err := c.GetNote(ctx, id); err == nil {
					item.Title = note.Title
				} else if res, err := c.GetResource(ctx, id); err == nil {
					item.Title = res.Title
				}
			}
			out.Items = append(out.Items, item)
		}
		return nil, out, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_backlinks",
		Description: "List notes whose body references the given note via :/<id>. Useful for Zettelkasten-style navigation. Joplin doesn't expose backlinks natively, so this is a search across all notes — paginated.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ListBacklinksArgs) (*mcp.CallToolResult, PageOut[NoteOut], error) {
		// Joplin's full-text search treats ":/<id>" as a token, so a quoted
		// substring search finds every note that embeds or links to the
		// target. The 32-char hex ID is unique enough that false positives
		// are not a real concern.
		query := fmt.Sprintf(`":/%s"`, args.NoteID)
		p, err := c.SearchNotes(ctx, query, joplin.ListOptions{Page: args.Page, Limit: args.Limit})
		if err != nil {
			return nil, PageOut[NoteOut]{}, err
		}
		// Don't return the source note as its own backlink (a self-reference
		// could happen but is noise for the LLM).
		filtered := joplin.Page[joplin.Note]{HasMore: p.HasMore}
		for _, n := range p.Items {
			if n.ID != args.NoteID {
				filtered.Items = append(filtered.Items, n)
			}
		}
		return nil, notesPage(filtered), nil
	})
}
