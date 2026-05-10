package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/thereisnotime/joplin-mcp/internal/joplin"
)

type HealthOut struct {
	JoplinReachable bool   `json:"joplin_reachable"`
	Reachability    string `json:"reachability_detail,omitempty" jsonschema:"empty when reachable; otherwise the underlying error"`
	JoplinBaseURL   string `json:"joplin_base_url"`
	ServerVersion   string `json:"server_version"`
	MasterKeyCount  int    `json:"master_key_count" jsonschema:"how many encryption master keys this profile has registered"`
}

type MasterKeyOut struct {
	ID                string `json:"id"`
	SourceApplication string `json:"source_application,omitempty"`
	EncryptionMethod  int    `json:"encryption_method,omitempty"`
	Checksum          string `json:"checksum,omitempty"`
	Hint              string `json:"hint,omitempty" jsonschema:"the user's password hint, if they set one when creating the key"`
	Enabled           bool   `json:"enabled"`
	CreatedTime       int64  `json:"created_time,omitempty"`
	UpdatedTime       int64  `json:"updated_time,omitempty"`
}

type GetMasterKeyArgs struct {
	MasterKeyID string `json:"master_key_id"`
}

type MasterKeysOut struct {
	Items []MasterKeyOut `json:"items"`
}

func registerDiagnosticTools(srv *mcp.Server, c *joplin.Client, baseURL, version string) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "health",
		Description: "Quick connectivity check: pings Joplin, reports the configured base URL, joplin-mcp version, and how many master keys this profile has. Useful as a first call when a tool fails — confirms whether the problem is reachability or something else.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ NoArgs) (*mcp.CallToolResult, HealthOut, error) {
		out := HealthOut{JoplinBaseURL: baseURL, ServerVersion: version}
		if err := c.Ping(ctx); err != nil {
			out.JoplinReachable = false
			out.Reachability = err.Error()
			return nil, out, nil
		}
		out.JoplinReachable = true
		// Best-effort master-key count; don't fail the whole health call if
		// the metadata fetch errors.
		if p, err := c.ListMasterKeys(ctx, joplin.ListOptions{Limit: 100}); err == nil {
			out.MasterKeyCount = len(p.Items)
		}
		return nil, out, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_master_keys",
		Description: "List the encryption master keys registered in this Joplin profile (read-only metadata: id, hint, encryption method, etc). joplin-mcp never decrypts; this is purely diagnostic.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ NoArgs) (*mcp.CallToolResult, MasterKeysOut, error) {
		all, err := joplin.CollectAll(ctx, func(ctx context.Context, page int) (joplin.Page[joplin.MasterKey], error) {
			return c.ListMasterKeys(ctx, joplin.ListOptions{Page: page, Limit: 100})
		})
		if err != nil {
			return nil, MasterKeysOut{}, err
		}
		out := MasterKeysOut{Items: make([]MasterKeyOut, 0, len(all))}
		for _, k := range all {
			out.Items = append(out.Items, masterKeyOut(k))
		}
		return nil, out, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_master_key",
		Description: "Get metadata for one encryption master key by ID. No private material is exposed.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GetMasterKeyArgs) (*mcp.CallToolResult, MasterKeyOut, error) {
		k, err := c.GetMasterKey(ctx, args.MasterKeyID)
		if err != nil {
			return nil, MasterKeyOut{}, err
		}
		return nil, masterKeyOut(k), nil
	})
}

func masterKeyOut(k joplin.MasterKey) MasterKeyOut {
	return MasterKeyOut{
		ID:                k.ID,
		SourceApplication: k.SourceApplication,
		EncryptionMethod:  k.EncryptionMethod,
		Checksum:          k.Checksum,
		Hint:              k.Hint,
		Enabled:           k.Enabled != 0,
		CreatedTime:       k.CreatedTime,
		UpdatedTime:       k.UpdatedTime,
	}
}
