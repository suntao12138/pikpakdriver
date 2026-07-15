package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/suntao12138/pikpakdriver/pkg/pikpak"
)

// ── Share Tools ────────────────────────────────────────────────────────────

type ShareTools struct{ client *pikpak.Client }

func NewShareTools(client *pikpak.Client) *ShareTools { return &ShareTools{client: client} }

type createShareArgs struct {
	FileIDs    []string `json:"file_ids" jsonschema:"IDs of files to share"`
	ExpireDays int      `json:"expire_days,omitempty" jsonschema:"expiration in days, 0 = never"`
	PassCode   string   `json:"pass_code,omitempty" jsonschema:"optional password protection"`
}

type getShareInfoArgs struct {
	ShareID  string `json:"share_id" jsonschema:"share link ID"`
	PassCode string `json:"pass_code,omitempty" jsonschema:"share password if required"`
}

type saveShareArgs struct {
	ShareID       string   `json:"share_id" jsonschema:"share ID to save from"`
	PassCodeToken string   `json:"pass_code_token" jsonschema:"pass code token from getShareInfo"`
	FileIDs       []string `json:"file_ids" jsonschema:"file IDs to save"`
	ToParentID    string   `json:"to_parent_id" jsonschema:"target folder ID"`
}

type shareDetailArgs struct {
	ShareID       string `json:"share_id" jsonschema:"share ID"`
	PassCodeToken string `json:"pass_code_token" jsonschema:"pass code token from getShareInfo"`
	DirID         string `json:"dir_id,omitempty" jsonschema:"folder ID inside share (for subdirectories)"`
}

type deleteShareArgs struct {
	ShareIDs []string `json:"share_ids" jsonschema:"list of share IDs to delete"`
}

func (st *ShareTools) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createShare",
		Description: "Create a share link for files or folders",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args createShareArgs) (*mcp.CallToolResult, any, error) {
		expire := args.ExpireDays
		if expire <= 0 {
			expire = 0
		}
		resp, err := st.client.CreateShare(args.FileIDs, expire, args.PassCode)
		if err != nil {
			return errorResult("createShare failed: %v", err), nil, nil
		}
		return jsonResult(resp), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "getShareInfo",
		Description: "Get information about a shared folder, including pass_code_token needed for saveShare",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args getShareInfoArgs) (*mcp.CallToolResult, any, error) {
		resp, err := st.client.GetShareInfo(args.ShareID, args.PassCode)
		if err != nil {
			return errorResult("getShareInfo failed: %v", err), nil, nil
		}
		return jsonResult(resp), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "saveShare",
		Description: "Save shared files to your own PikPak",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args saveShareArgs) (*mcp.CallToolResult, any, error) {
		if err := st.client.SaveShare(args.ShareID, args.PassCodeToken, args.FileIDs, args.ToParentID); err != nil {
			return errorResult("saveShare failed: %v", err), nil, nil
		}
		return successResult(`{"saved": true}`), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "shareDetail",
		Description: "List files inside a shared folder",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args shareDetailArgs) (*mcp.CallToolResult, any, error) {
		resp, err := st.client.ShareDetail(args.ShareID, args.PassCodeToken, args.DirID)
		if err != nil {
			return errorResult("shareDetail failed: %v", err), nil, nil
		}
		return jsonResult(resp), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "listShares",
		Description: "List all your created share links",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args emptyArgs) (*mcp.CallToolResult, any, error) {
		resp, err := st.client.ListShares()
		if err != nil {
			return errorResult("listShares failed: %v", err), nil, nil
		}
		return jsonResult(resp), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteShares",
		Description: "Delete share links",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args deleteShareArgs) (*mcp.CallToolResult, any, error) {
		if err := st.client.DeleteShares(args.ShareIDs); err != nil {
			return errorResult("deleteShares failed: %v", err), nil, nil
		}
		return successResult(`{"deleted": true}`), nil, nil
	})
}
