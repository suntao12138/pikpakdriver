package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/suntao12138/pikpakdriver/pkg/pikpak"
)

// ── Events Tools ───────────────────────────────────────────────────────────

type EventsTools struct{ client *pikpak.Client }

func NewEventsTools(client *pikpak.Client) *EventsTools { return &EventsTools{client: client} }

type listEventsArgs struct {
	Limit int `json:"limit,omitempty" jsonschema:"max events to return, default 50"`
}

func (et *EventsTools) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listEvents",
		Description: "List recent file events (uploads, shares, etc.)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args listEventsArgs) (*mcp.CallToolResult, any, error) {
		limit := args.Limit
		if limit <= 0 {
			limit = 50
		}
		resp, err := et.client.ListEvents(limit)
		if err != nil {
			return errorResult("listEvents failed: %v", err), nil, nil
		}
		return jsonResult(resp), nil, nil
	})
}
