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

	// Register implant generation tools
	registerImplantTools(mcpServer, sliverClient)

	// Register process management tools
	registerProcessTools(mcpServer, sliverClient)

	// Add more tool registrations here as they are implemented

	return mcpServer
}

// registerImplantTools registers all tools related to implant generation
func registerImplantTools(mcpServer *server.MCPServer, sliverClient *client.SliverClient) {
	// Generate Implant
	mcpServer.AddTool(mcp.NewTool("generateImplant",
		mcp.WithDescription("Generate a new Sliver implant"),
		mcp.WithObject("config",
			mcp.Description("The implant configuration"),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("The name for the implant"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleGenerateImplant(ctx, request, sliverClient)
	})

	// List Implant Profiles
	mcpServer.AddTool(mcp.NewTool("listImplantProfiles",
		mcp.WithDescription("List implant profiles"),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleListImplantProfiles(ctx, request, sliverClient)
	})

	// Save Implant Profile
	mcpServer.AddTool(mcp.NewTool("saveImplantProfile",
		mcp.WithDescription("Save an implant profile"),
		mcp.WithString("name",
			mcp.Description("The name for the profile"),
			mcp.Required(),
		),
		mcp.WithObject("config",
			mcp.Description("The implant configuration"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleSaveImplantProfile(ctx, request, sliverClient)
	})

	// Delete Implant Profile
	mcpServer.AddTool(mcp.NewTool("deleteImplantProfile",
		mcp.WithDescription("Delete an implant profile"),
		mcp.WithString("profileID",
			mcp.Description("The ID of the profile to delete"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleDeleteImplantProfile(ctx, request, sliverClient)
	})

	// List Implant Builds
	mcpServer.AddTool(mcp.NewTool("listImplantBuilds",
		mcp.WithDescription("List available implant builds"),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleListImplantBuilds(ctx, request, sliverClient)
	})

	// Regenerate Implant
	mcpServer.AddTool(mcp.NewTool("regenerateImplant",
		mcp.WithDescription("Regenerate an existing implant"),
		mcp.WithString("implantName",
			mcp.Description("The name of the implant to regenerate"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleRegenerateImplant(ctx, request, sliverClient)
	})

	// Generate Stager
	mcpServer.AddTool(mcp.NewTool("generateStage",
		mcp.WithDescription("Generate a stager implant"),
		mcp.WithString("profile",
			mcp.Description("The profile to use for the stager"),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("The name for the stager"),
		),
		mcp.WithString("aesEncryptKey",
			mcp.Description("AES encryption key for the stager"),
		),
		mcp.WithString("aesEncryptIv",
			mcp.Description("AES encryption IV for the stager"),
		),
		mcp.WithString("rc4EncryptKey",
			mcp.Description("RC4 encryption key for the stager"),
		),
		mcp.WithString("compress",
			mcp.Description("Compression mode"),
		),
		mcp.WithString("compressF",
			mcp.Description("Compression format"),
		),
		mcp.WithBoolean("prependSize",
			mcp.Description("Prepend size to the stager"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleGenerateStager(ctx, request, sliverClient)
	})
}

// registerProcessTools registers all process management tools
func registerProcessTools(mcpServer *server.MCPServer, sliverClient *client.SliverClient) {
	// List Processes
	mcpServer.AddTool(mcp.NewTool("ps",
		mcp.WithDescription("List processes on the remote system"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to use"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandlePs(ctx, request, sliverClient)
	})

	// Terminate Process
	mcpServer.AddTool(mcp.NewTool("terminate",
		mcp.WithDescription("Terminate a remote process"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to use"),
			mcp.Required(),
		),
		mcp.WithNumber("pid",
			mcp.Description("The process ID to terminate"),
			mcp.Required(),
		),
		mcp.WithBoolean("force",
			mcp.Description("Force terminate the process"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleTerminateProcess(ctx, request, sliverClient)
	})

	// Execute Command
	mcpServer.AddTool(mcp.NewTool("execute",
		mcp.WithDescription("Execute a command on the remote system"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to use"),
			mcp.Required(),
		),
		mcp.WithString("command",
			mcp.Description("The command to execute"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleExecute(ctx, request, sliverClient)
	})
}

// registerSessionTools registers all session-related tools
func registerSessionTools(mcpServer *server.MCPServer, sliverClient *client.SliverClient) {
	// List Sessions
	mcpServer.AddTool(mcp.NewTool("listSessions",
		mcp.WithDescription("List all active Sliver sessions"),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleListSessions(ctx, request, sliverClient)
	})

	// Get Session
	mcpServer.AddTool(mcp.NewTool("getSession",
		mcp.WithDescription("Get information about a specific session"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to retrieve"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleGetSession(ctx, request, sliverClient)
	})

	// Rename Session
	mcpServer.AddTool(mcp.NewTool("renameSession",
		mcp.WithDescription("Rename a session"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to rename"),
			mcp.Required(),
		),
		mcp.WithString("name",
			mcp.Description("The new name for the session"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleRenameSession(ctx, request, sliverClient)
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

	// Remove Beacon
	mcpServer.AddTool(mcp.NewTool("removeBeacon",
		mcp.WithDescription("Remove a beacon"),
		mcp.WithString("beaconID",
			mcp.Description("The ID of the beacon to remove"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleRemoveBeacon(ctx, request, sliverClient)
	})

	// Get Beacon Tasks
	mcpServer.AddTool(mcp.NewTool("getBeaconTasks",
		mcp.WithDescription("Get tasks for a specific beacon"),
		mcp.WithString("beaconID",
			mcp.Description("The ID of the beacon"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleGetBeaconTasks(ctx, request, sliverClient)
	})

	// Cancel Beacon Task
	mcpServer.AddTool(mcp.NewTool("cancelBeaconTask",
		mcp.WithDescription("Cancel a specific beacon task"),
		mcp.WithString("beaconID",
			mcp.Description("The ID of the beacon"),
			mcp.Required(),
		),
		mcp.WithString("taskID",
			mcp.Description("The ID of the task to cancel"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleCancelBeaconTask(ctx, request, sliverClient)
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

	// Move Files
	mcpServer.AddTool(mcp.NewTool("mv",
		mcp.WithDescription("Move a file or directory on the remote system"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to use"),
			mcp.Required(),
		),
		mcp.WithString("srcPath",
			mcp.Description("The source path"),
			mcp.Required(),
		),
		mcp.WithString("dstPath",
			mcp.Description("The destination path"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleMv(ctx, request, sliverClient)
	})

	// Copy Files
	mcpServer.AddTool(mcp.NewTool("cp",
		mcp.WithDescription("Copy a file or directory on the remote system"),
		mcp.WithString("sessionID",
			mcp.Description("The ID of the session to use"),
			mcp.Required(),
		),
		mcp.WithString("srcPath",
			mcp.Description("The source path"),
			mcp.Required(),
		),
		mcp.WithString("dstPath",
			mcp.Description("The destination path"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleCp(ctx, request, sliverClient)
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
