package server

import (
	"context"
	"log"

	"github.com/adamkadaban/sliver-mcp/internal/client"
	"github.com/adamkadaban/sliver-mcp/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewSliverMCPServer(configPath string) *server.MCPServer {
	hooks := &server.Hooks{}

	// Setup hooks for logging and debugging
	hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
		log.Printf("Tool call: %s", message.Params.Name)
	})

	hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		log.Printf("Tool call completed: %s", message.Params.Name)
	})

	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		log.Printf("Error in method %s: %v", method, err)
	})

	mcpServer := server.NewMCPServer(
		"sliver-mcp",
		"1.0.0",
		server.WithLogging(),
		server.WithHooks(hooks),
	)

	// Initialize the Sliver client
	sliverClient, err := client.NewSliverClient(configPath)
	if err != nil {
		log.Fatalf("Failed to initialize Sliver client: %v", err)
	}

	// Register session management tools
	registerSessionTools(mcpServer, sliverClient)

	// Register file system tools
	registerFileSystemTools(mcpServer, sliverClient)

	// Add more tool registrations here as they are implemented

	return mcpServer
}

// registerSessionTools registers all session-related tools
func registerSessionTools(mcpServer *server.MCPServer, sliverClient *client.SliverClient) {
	// List Sessions
	mcpServer.AddTool(mcp.NewTool("listSessions",
		mcp.WithDescription("List all active Sliver sessions"),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleListSessions(ctx, request, sliverClient)
	})

	// Kill Session
	mcpServer.AddTool(mcp.NewTool("killSession",
		mcp.WithDescription("Terminate a specific session"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to terminate"),
			mcp.Required(),
		),
		mcp.WithBoolean("force",
			mcp.Description("Force kill the session"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleKillSession(ctx, request, sliverClient)
	})

	// List Beacons
	mcpServer.AddTool(mcp.NewTool("listBeacons",
		mcp.WithDescription("List all active Sliver beacons"),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleListBeacons(ctx, request, sliverClient)
	})

	// Get Beacon
	mcpServer.AddTool(mcp.NewTool("getBeacon",
		mcp.WithDescription("Get information about a specific beacon"),
		mcp.WithString("beaconID",
			mcp.Description("The ID of the beacon to retrieve"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleGetBeacon(ctx, request, sliverClient)
	})

	// List Jobs
	mcpServer.AddTool(mcp.NewTool("listJobs",
		mcp.WithDescription("List all active Sliver jobs (listeners)"),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleListJobs(ctx, request, sliverClient)
	})

	// Kill Job
	mcpServer.AddTool(mcp.NewTool("killJob",
		mcp.WithDescription("Kill a specific job (listener)"),
		mcp.WithNumber("jobID",
			mcp.Description("The ID of the job to kill"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleKillJob(ctx, request, sliverClient)
	})
}

// registerFileSystemTools registers all file system tools
func registerFileSystemTools(mcpServer *server.MCPServer, sliverClient *client.SliverClient) {
	// List Files
	mcpServer.AddTool(mcp.NewTool("ls",
		mcp.WithDescription("List files on the remote system"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to use"),
			mcp.Required(),
		),
		mcp.WithString("path",
			mcp.Description("The path to list files from (default: current directory)"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleLs(ctx, request, sliverClient)
	})

	// Present Working Directory
	mcpServer.AddTool(mcp.NewTool("pwd",
		mcp.WithDescription("Get the current working directory on the remote system"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to use"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandlePwd(ctx, request, sliverClient)
	})

	// Change Directory
	mcpServer.AddTool(mcp.NewTool("cd",
		mcp.WithDescription("Change the current working directory on the remote system"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to use"),
			mcp.Required(),
		),
		mcp.WithString("path",
			mcp.Description("The path to change to"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleCd(ctx, request, sliverClient)
	})

	// Download File
	mcpServer.AddTool(mcp.NewTool("download",
		mcp.WithDescription("Download a file from the remote system"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to use"),
			mcp.Required(),
		),
		mcp.WithString("remotePath",
			mcp.Description("The path on the remote system to download"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleDownload(ctx, request, sliverClient)
	})

	// Upload File
	mcpServer.AddTool(mcp.NewTool("upload",
		mcp.WithDescription("Upload a file to the remote system"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to use"),
			mcp.Required(),
		),
		mcp.WithString("remotePath",
			mcp.Description("The path on the remote system to upload to"),
			mcp.Required(),
		),
		mcp.WithString("data",
			mcp.Description("The base64-encoded file data to upload"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleUpload(ctx, request, sliverClient)
	})

	// Make Directory
	mcpServer.AddTool(mcp.NewTool("mkdir",
		mcp.WithDescription("Create a directory on the remote system"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to use"),
			mcp.Required(),
		),
		mcp.WithString("path",
			mcp.Description("The path to create"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleMkdir(ctx, request, sliverClient)
	})

	// Remove File/Directory
	mcpServer.AddTool(mcp.NewTool("rm",
		mcp.WithDescription("Remove a file or directory on the remote system"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to use"),
			mcp.Required(),
		),
		mcp.WithString("path",
			mcp.Description("The path to remove"),
			mcp.Required(),
		),
		mcp.WithBoolean("recursive",
			mcp.Description("Recursively remove directories"),
		),
		mcp.WithBoolean("force",
			mcp.Description("Force removal"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleRm(ctx, request, sliverClient)
	})
}
