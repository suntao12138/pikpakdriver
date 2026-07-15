package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/suntao12138/pikpakdriver/pkg/pikpak"
)

// ── Offline Tools ──────────────────────────────────────────────────────────

type OfflineTools struct{ client *pikpak.Client }

func NewOfflineTools(client *pikpak.Client) *OfflineTools { return &OfflineTools{client: client} }

type addOfflineTaskArgs struct {
	URL      string `json:"url" jsonschema:"magnet link or HTTP URL to download"`
	ParentID string `json:"parent_id,omitempty" jsonschema:"optional target folder ID, empty = default DOWNLOAD folder"`
	Name     string `json:"name,omitempty" jsonschema:"optional custom filename"`
}

type listOfflineTasksArgs struct {
	Limit  int      `json:"limit,omitempty" jsonschema:"max tasks, default 20"`
	Phases []string `json:"phases,omitempty" jsonschema:"filter by phase, e.g. PHASE_TYPE_RUNNING,PHASE_TYPE_COMPLETE"`
}

type deleteOfflineTaskArgs struct {
	TaskID     string `json:"task_id" jsonschema:"task ID to delete"`
	DeleteFile bool   `json:"delete_file,omitempty" jsonschema:"also delete downloaded file"`
}

type getOfflineTaskArgs struct {
	TaskID string `json:"task_id" jsonschema:"task ID to inspect"`
}

type retryOfflineTaskArgs struct {
	TaskID string `json:"task_id" jsonschema:"task ID to retry"`
}

func (ot *OfflineTools) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "addOfflineTask",
		Description: "Add a magnet link or URL to PikPak offline download queue",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args addOfflineTaskArgs) (*mcp.CallToolResult, any, error) {
		if args.URL == "" {
			return errorResult("URL is required"), nil, nil
		}
		resp, err := ot.client.AddOfflineTask(args.URL, args.ParentID, args.Name)
		if err != nil {
			return errorResult("addOfflineTask failed: %v", err), nil, nil
		}
		return jsonResult(resp), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "listOfflineTasks",
		Description: "List PikPak offline download tasks",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args listOfflineTasksArgs) (*mcp.CallToolResult, any, error) {
		limit := args.Limit
		if limit <= 0 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}
		phases := args.Phases
		if len(phases) == 0 {
			phases = []string{"PHASE_TYPE_RUNNING", "PHASE_TYPE_COMPLETE", "PHASE_TYPE_ERROR"}
		}
		list, err := ot.client.ListOfflineTasks(limit, phases)
		if err != nil {
			return errorResult("listOfflineTasks failed: %v", err), nil, nil
		}
		return jsonResult(list), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "getOfflineTask",
		Description: "Get detailed info about a specific offline download task",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args getOfflineTaskArgs) (*mcp.CallToolResult, any, error) {
		task, err := ot.client.GetOfflineTask(args.TaskID)
		if err != nil {
			return errorResult("getOfflineTask failed: %v", err), nil, nil
		}
		return jsonResult(task), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "deleteOfflineTask",
		Description: "Delete a PikPak offline download task",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args deleteOfflineTaskArgs) (*mcp.CallToolResult, any, error) {
		if args.TaskID == "" {
			return errorResult("task_id is required"), nil, nil
		}
		if err := ot.client.DeleteOfflineTask(args.TaskID, args.DeleteFile); err != nil {
			return errorResult("deleteOfflineTask failed: %v", err), nil, nil
		}
		return successResult(`{"deleted":true,"task_id":"` + args.TaskID + `"}`), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "retryOfflineTask",
		Description: "Retry a failed offline download task",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args retryOfflineTaskArgs) (*mcp.CallToolResult, any, error) {
		if err := ot.client.RetryOfflineTask(args.TaskID); err != nil {
			return errorResult("retryOfflineTask failed: %v", err), nil, nil
		}
		return successResult(`{"retried":true,"task_id":"` + args.TaskID + `"}`), nil, nil
	})
}
