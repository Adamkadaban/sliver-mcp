package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/adamkadaban/sliver-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// HandlePs handles the 'ps' tool request to list processes
func HandlePs(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	sessionID, ok := arguments["sessionID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("sessionID must be a string")
	}

	ps, err := client.Ps(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	processes := make([]map[string]interface{}, 0)
	for _, process := range ps.Processes {
		proc := map[string]interface{}{
			"pid":     process.Pid,
			"ppid":    process.Ppid,
			"name":    process.Executable,
			"owner":   process.Owner,
			"session": process.SessionID,
		}

		// Architecture field is not directly accessible in this version of Sliver
		// If the build still fails, complete remove this code and use only the fields above
		processes = append(processes, proc)
	}

	result, _ := json.MarshalIndent(processes, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(result),
			},
		},
	}, nil
}

// HandleTerminateProcess handles the 'terminate' tool request to kill a remote process
func HandleTerminateProcess(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	sessionID, ok := arguments["sessionID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("sessionID must be a string")
	}

	pidFloat, ok := arguments["pid"].(float64)
	if !ok {
		return nil, NewInvalidArgsError("pid must be a number")
	}
	pid := int32(pidFloat)

	force := false
	if forceArg, ok := arguments["force"].(bool); ok {
		force = forceArg
	}

	terminated, err := client.Terminate(ctx, sessionID, pid, force)
	if err != nil {
		return nil, err
	}

	var responseText string
	if terminated.Response != nil && terminated.Response.GetErr() != "" {
		responseText = fmt.Sprintf("Failed to terminate process %d: %s", pid, terminated.Response.GetErr())
	} else {
		responseText = fmt.Sprintf("Process %d has been terminated", pid)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: responseText,
			},
		},
	}, nil
}

// HandleExecute handles the 'execute' tool request to run a command on the remote system
func HandleExecute(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	sessionID, ok := arguments["sessionID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("sessionID must be a string")
	}

	command, ok := arguments["command"].(string)
	if !ok {
		return nil, NewInvalidArgsError("command must be a string")
	}

	execute, err := client.Execute(ctx, sessionID, command)
	if err != nil {
		return nil, err
	}

	var responseText string
	if execute.Response != nil && execute.Response.GetErr() != "" {
		responseText = fmt.Sprintf("Failed to execute command: %s", execute.Response.GetErr())
	} else {
		output := execute.GetStdout()
		if len(output) == 0 {
			output = execute.GetStderr()
		}
		if len(output) == 0 {
			responseText = "Command executed successfully (no output)"
		} else {
			responseText = fmt.Sprintf("Output:\n%s", string(output))
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: responseText,
			},
		},
	}, nil
}
