package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/suntao12138/pikpakdriver/pkg/pikpak"
)

// ── Helpers ────────────────────────────────────────────────────────────────

func errorResult(format string, args ...interface{}) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}}, IsError: true}
}
func successResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}
func jsonResult(v interface{}) *mcp.CallToolResult {
	b, _ := json.MarshalIndent(v, "", "  ")
	return successResult(string(b))
}

// ── File Tools ─────────────────────────────────────────────────────────────

type FileTools struct{ client *pikpak.Client }

func NewFileTools(client *pikpak.Client) *FileTools { return &FileTools{client: client} }

type listFilesArgs struct {
	ParentID string `json:"parent_id" jsonschema:"folder ID, empty string for root"`
	Limit    int    `json:"limit,omitempty" jsonschema:"max files, default 50"`
}
type mkdirArgs struct {
	ParentID string `json:"parent_id" jsonschema:"parent folder ID"`
	Name     string `json:"name" jsonschema:"new folder name"`
}
type renameArgs struct {
	FileID  string `json:"file_id" jsonschema:"file or folder ID"`
	NewName string `json:"new_name" jsonschema:"new name"`
}
type batchFilesArgs struct {
	IDs        []string `json:"ids" jsonschema:"list of file/folder IDs"`
	ToParentID string   `json:"to_parent_id,omitempty" jsonschema:"target parent folder ID"`
}
type fileIDArgs struct {
	FileID string `json:"file_id" jsonschema:"file ID"`
}
type limitArgs struct {
	Limit int `json:"limit,omitempty" jsonschema:"max results, default 50"`
}

type starArgs struct {
	IDs []string `json:"ids" jsonschema:"list of file IDs to star/unstar"`
}

func (ft *FileTools) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "listFiles", Description: "List files and folders in a PikPak directory",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args listFilesArgs) (*mcp.CallToolResult, any, error) {
		pid := args.ParentID
		limit := args.Limit
		if limit <= 0 {
			limit = 50
		}
		list, err := ft.client.ListFiles(pid, limit)
		if err != nil {
			return errorResult("listFiles failed: %v", err), nil, nil
		}
		return jsonResult(list), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "mkdir", Description: "Create a new folder in PikPak",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args mkdirArgs) (*mcp.CallToolResult, any, error) {
		f, err := ft.client.Mkdir(args.ParentID, args.Name)
		if err != nil {
			return errorResult("mkdir failed: %v", err), nil, nil
		}
		return jsonResult(f), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "rename", Description: "Rename a file or folder in PikPak",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args renameArgs) (*mcp.CallToolResult, any, error) {
		if err := ft.client.Rename(args.FileID, args.NewName); err != nil {
			return errorResult("rename failed: %v", err), nil, nil
		}
		return successResult(`{"renamed": true}`), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "moveFiles", Description: "Move files or folders to another directory",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args batchFilesArgs) (*mcp.CallToolResult, any, error) {
		if err := ft.client.MoveFiles(args.IDs, args.ToParentID); err != nil {
			return errorResult("move failed: %v", err), nil, nil
		}
		return successResult(`{"moved": true}`), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "copyFiles", Description: "Copy files or folders to another directory",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args batchFilesArgs) (*mcp.CallToolResult, any, error) {
		if err := ft.client.CopyFiles(args.IDs, args.ToParentID); err != nil {
			return errorResult("copy failed: %v", err), nil, nil
		}
		return successResult(`{"copied": true}`), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "trashFiles", Description: "Move files or folders to trash",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args batchFilesArgs) (*mcp.CallToolResult, any, error) {
		if err := ft.client.TrashFiles(args.IDs); err != nil {
			return errorResult("trash failed: %v", err), nil, nil
		}
		return successResult(`{"trashed": true}`), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "deleteFiles", Description: "Permanently delete files or folders (cannot be undone)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args batchFilesArgs) (*mcp.CallToolResult, any, error) {
		if err := ft.client.DeleteFiles(args.IDs); err != nil {
			return errorResult("delete failed: %v", err), nil, nil
		}
		return successResult(`{"deleted": true}`), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "untrashFiles", Description: "Restore files from trash",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args batchFilesArgs) (*mcp.CallToolResult, any, error) {
		if err := ft.client.UntrashFiles(args.IDs); err != nil {
			return errorResult("untrash failed: %v", err), nil, nil
		}
		return successResult(`{"untrashed": true}`), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "emptyTrash", Description: "Permanently empty the entire trash",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args emptyArgs) (*mcp.CallToolResult, any, error) {
		if err := ft.client.EmptyTrash(); err != nil {
			return errorResult("emptyTrash failed: %v", err), nil, nil
		}
		return successResult(`{"trash_emptied": true}`), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "listTrash", Description: "List files in trash",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args limitArgs) (*mcp.CallToolResult, any, error) {
		limit := args.Limit
		if limit <= 0 {
			limit = 50
		}
		list, err := ft.client.ListTrash(limit)
		if err != nil {
			return errorResult("listTrash failed: %v", err), nil, nil
		}
		return jsonResult(list), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "getFileInfo", Description: "Get detailed info about a specific file or folder",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fileIDArgs) (*mcp.CallToolResult, any, error) {
		info, err := ft.client.GetFileInfo(args.FileID)
		if err != nil {
			return errorResult("getFileInfo failed: %v", err), nil, nil
		}
		return jsonResult(info), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "getDownloadLink", Description: "Get a direct download URL for a file",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args fileIDArgs) (*mcp.CallToolResult, any, error) {
		link, err := ft.client.GetDownloadLink(args.FileID)
		if err != nil {
			return errorResult("getDownloadLink failed: %v", err), nil, nil
		}
		return jsonResult(link), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "starFiles", Description: "Star (bookmark) files or folders",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args starArgs) (*mcp.CallToolResult, any, error) {
		if err := ft.client.StarFiles(args.IDs); err != nil {
			return errorResult("star failed: %v", err), nil, nil
		}
		return successResult(`{"starred": true}`), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "unstarFiles", Description: "Unstar (remove bookmark from) files or folders",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args starArgs) (*mcp.CallToolResult, any, error) {
		if err := ft.client.UnstarFiles(args.IDs); err != nil {
			return errorResult("unstar failed: %v", err), nil, nil
		}
		return successResult(`{"unstarred": true}`), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "listStarred", Description: "List all starred (bookmarked) files",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args limitArgs) (*mcp.CallToolResult, any, error) {
		limit := args.Limit
		if limit <= 0 {
			limit = 50
		}
		list, err := ft.client.ListStarred(limit)
		if err != nil {
			return errorResult("listStarred failed: %v", err), nil, nil
		}
		return jsonResult(list), nil, nil
	})
}
