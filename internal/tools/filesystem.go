package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/adamkadaban/sliver-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// HandleLs : ls tool request
func HandleLs(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	sessionID, ok := arguments["sessionID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("sessionID must be a string")
	}

	path := "./"
	if pathArg, ok := arguments["path"].(string); ok && pathArg != "" {
		path = pathArg
	}

	ls, err := client.Ls(ctx, sessionID, path)
	if err != nil {
		return nil, err
	}

	var formattedFiles []map[string]interface{}
	for _, file := range ls.Files {
		modTime := time.Unix(0, file.ModTime).Format(time.RFC3339)

		formattedFiles = append(formattedFiles, map[string]interface{}{
			"name":    file.Name,
			"isDir":   file.IsDir,
			"size":    file.Size,
			"modTime": modTime,
			"mode":    file.Mode,
		})
	}

	result, err := json.Marshal(map[string]interface{}{
		"path":   ls.Path,
		"exists": ls.Exists,
		"files":  formattedFiles,
	})
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(result),
			},
		},
	}, nil
}

// HandlePwd : pwd tool request
func HandlePwd(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	sessionID, ok := arguments["sessionID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("sessionID must be a string")
	}

	pwd, err := client.Pwd(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	result, err := json.Marshal(map[string]interface{}{
		"path": pwd.Path,
	})
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(result),
			},
		},
	}, nil
}

// HandleCd : cd tool request
func HandleCd(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	sessionID, ok := arguments["sessionID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("sessionID must be a string")
	}

	path, ok := arguments["path"].(string)
	if !ok {
		return nil, NewInvalidArgsError("path must be a string")
	}

	pwd, err := client.Cd(ctx, sessionID, path)
	if err != nil {
		return nil, err
	}

	result, err := json.Marshal(map[string]interface{}{
		"path": pwd.Path,
	})
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(result),
			},
		},
	}, nil
}

// HandleDownload : download
func HandleDownload(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	sessionID, ok := arguments["sessionID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("sessionID must be a string")
	}

	remotePath, ok := arguments["remotePath"].(string)
	if !ok {
		return nil, NewInvalidArgsError("remotePath must be a string")
	}

	download, err := client.Download(ctx, sessionID, remotePath)
	if err != nil {
		return nil, err
	}

	var fileData string
	if download.Data != nil {
		fileData = base64.StdEncoding.EncodeToString(download.Data)
	}

	result, err := json.Marshal(map[string]interface{}{
		"path":   download.Path,
		"exists": download.Exists,
		"isDir":  download.IsDir,
		"data":   fileData,
		"size":   len(download.Data),
	})
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(result),
			},
		},
	}, nil
}

// HandleUpload : upload tool request
func HandleUpload(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	sessionID, ok := arguments["sessionID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("sessionID must be a string")
	}

	remotePath, ok := arguments["remotePath"].(string)
	if !ok {
		return nil, NewInvalidArgsError("remotePath must be a string")
	}

	data, ok := arguments["data"].(string)
	if !ok {
		return nil, NewInvalidArgsError("data must be a base64-encoded string")
	}

	fileData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode file data: %v", err)
	}

	upload, err := client.Upload(ctx, sessionID, remotePath, fileData)
	if err != nil {
		return nil, err
	}

	result, err := json.Marshal(map[string]interface{}{
		"path": upload.Path,
	})
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(result),
			},
		},
	}, nil
}

// HandleMkdir : mkdir tool request
func HandleMkdir(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract and validate arguments
	sessionID, ok := arguments["sessionID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("sessionID must be a string")
	}

	path, ok := arguments["path"].(string)
	if !ok {
		return nil, NewInvalidArgsError("path must be a string")
	}

	// Call the client's Mkdir method
	mkdir, err := client.Mkdir(ctx, sessionID, path)
	if err != nil {
		return nil, err
	}

	// Return the result as JSON
	result, err := json.Marshal(map[string]interface{}{
		"path": mkdir.Path,
	})
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(result),
			},
		},
	}, nil
}

// HandleRm handles the rm tool request
func HandleRm(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract and validate arguments
	sessionID, ok := arguments["sessionID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("sessionID must be a string")
	}

	path, ok := arguments["path"].(string)
	if !ok {
		return nil, NewInvalidArgsError("path must be a string")
	}

	recursive := false
	if recursiveArg, ok := arguments["recursive"].(bool); ok {
		recursive = recursiveArg
	}

	force := false
	if forceArg, ok := arguments["force"].(bool); ok {
		force = forceArg
	}

	// Call the client's Rm method
	rm, err := client.Rm(ctx, sessionID, path, recursive, force)
	if err != nil {
		return nil, err
	}

	// Return the result as JSON
	result, err := json.Marshal(map[string]interface{}{
		"path": rm.Path,
	})
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(result),
			},
		},
	}, nil
}
