package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

type AttachResourceArgs struct {
	NoteID     string `json:"note_id"`
	Filename   string `json:"filename" jsonschema:"file name including extension — Joplin uses this to infer MIME type"`
	Base64Data string `json:"base64_data" jsonschema:"the file's bytes, base64-encoded"`
	Title      string `json:"title,omitempty"`
	AltText    string `json:"alt_text,omitempty" jsonschema:"alt text for the markdown link; defaults to the filename"`
	Position   string `json:"position,omitempty" jsonschema:"where to insert the markdown reference: 'top' or 'bottom' (default)"`
}

type AttachResourceOut struct {
	Resource ResourceOut `json:"resource"`
	Note     NoteOut     `json:"note"`
	// The exact markdown line that was inserted. The LLM can re-use this
	// when it needs to reference the same attachment elsewhere.
	InsertedMarkdown string `json:"inserted_markdown"`
}

func registerAttachTools(srv *mcp.Server, c *joplin.Client, maxBytes int64) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "attach_resource_to_note",
		Description: "Upload a file as a Joplin resource AND insert a properly-formatted markdown reference into the note body in one call. Image MIME types use ![]() syntax (rendered inline); other types use []() (rendered as a download link).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args AttachResourceArgs) (*mcp.CallToolResult, AttachResourceOut, error) {
		if args.Filename == "" {
			return nil, AttachResourceOut{}, fmt.Errorf("filename is required")
		}
		if int64(base64.StdEncoding.DecodedLen(len(args.Base64Data))) > maxBytes {
			return nil, AttachResourceOut{}, errResourceTooLarge(int64(base64.StdEncoding.DecodedLen(len(args.Base64Data))), maxBytes)
		}
		data, err := base64.StdEncoding.DecodeString(args.Base64Data)
		if err != nil {
			return nil, AttachResourceOut{}, err
		}
		if int64(len(data)) > maxBytes {
			return nil, AttachResourceOut{}, errResourceTooLarge(int64(len(data)), maxBytes)
		}

		res, err := c.UploadResource(ctx, data, args.Filename, args.Title)
		if err != nil {
			return nil, AttachResourceOut{}, err
		}

		// Pull the source note so we can append/prepend without losing body.
		note, err := c.GetNote(ctx, args.NoteID)
		if err != nil {
			return nil, AttachResourceOut{}, err
		}

		alt := args.AltText
		if alt == "" {
			alt = args.Filename
		}
		// Image MIME types render inline; everything else as a clickable
		// download link.
		prefix := ""
		if strings.HasPrefix(res.Mime, "image/") {
			prefix = "!"
		}
		ref := fmt.Sprintf("%s[%s](:/%s)", prefix, alt, res.ID)

		var newBody string
		switch strings.ToLower(args.Position) {
		case "top":
			if note.Body == "" {
				newBody = ref
			} else {
				newBody = ref + "\n\n" + note.Body
			}
		default: // "bottom" or unspecified
			if note.Body == "" {
				newBody = ref
			} else {
				newBody = note.Body + "\n\n" + ref
			}
		}

		updated, err := c.UpdateNote(ctx, args.NoteID, joplin.UpdateNoteInput{Body: &newBody})
		if err != nil {
			return nil, AttachResourceOut{}, err
		}

		return nil, AttachResourceOut{
			Resource:         resourceOut(res),
			Note:             noteOut(updated),
			InsertedMarkdown: ref,
		}, nil
	})
}
