package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/suntao12138/pikpakdriver/pkg/pikpak"
)

// ── Account Tools ──────────────────────────────────────────────────────────

type emptyArgs struct{}

type AccountTools struct{ client *pikpak.Client }

func NewAccountTools(client *pikpak.Client) *AccountTools { return &AccountTools{client: client} }

func (at *AccountTools) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "getAccountInfo",
		Description: "Get PikPak account info including user profile, storage quota, transfer quota, and VIP status",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args emptyArgs) (*mcp.CallToolResult, any, error) {
		user, _ := at.client.GetUserInfo()
		quota, _ := at.client.GetQuota()
		transfer, _ := at.client.GetTransferQuota()
		vip, _ := at.client.GetVipInfo()
		return jsonResult(map[string]interface{}{
			"user": user, "quota": quota, "transfer": transfer, "vip": vip,
		}), nil, nil
	})
}
