package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/adamkadaban/sliver-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func HandleListSessions(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	sessions, err := client.GetSessions(ctx)
	if err != nil {
		return nil, err
	}

	var formattedSessions []map[string]interface{}
	for _, session := range sessions.Sessions {
		lastCheckin := time.Unix(0, session.LastCheckin).Format(time.RFC3339)

		formattedSessions = append(formattedSessions, map[string]interface{}{
			"id":            session.ID,
			"name":          session.Name,
			"hostname":      session.Hostname,
			"os":            session.OS,
			"arch":          session.Arch,
			"username":      session.Username,
			"pid":           session.PID,
			"transport":     session.Transport,
			"remoteAddress": session.RemoteAddress,
			"lastCheckin":   lastCheckin,
			"isDead":        session.IsDead,
		})
	}

	result, err := json.Marshal(map[string]interface{}{
		"sessions": formattedSessions,
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

func HandleKillSession(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	sessionID, ok := arguments["sessionID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("sessionID must be a string")
	}

	force := false
	if forceArg, ok := arguments["force"].(bool); ok {
		force = forceArg
	}

	err := client.Kill(ctx, sessionID, force)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Session %s successfully terminated", sessionID),
			},
		},
	}, nil
}

func HandleListBeacons(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	beacons, err := client.GetBeacons(ctx)
	if err != nil {
		return nil, err
	}

	var formattedBeacons []map[string]interface{}
	for _, beacon := range beacons.Beacons {
		lastCheckin := time.Unix(0, beacon.LastCheckin).Format(time.RFC3339)
		nextCheckin := time.Unix(0, beacon.NextCheckin).Format(time.RFC3339)

		formattedBeacons = append(formattedBeacons, map[string]interface{}{
			"id":            beacon.ID,
			"name":          beacon.Name,
			"hostname":      beacon.Hostname,
			"os":            beacon.OS,
			"arch":          beacon.Arch,
			"username":      beacon.Username,
			"pid":           beacon.PID,
			"transport":     beacon.Transport,
			"remoteAddress": beacon.RemoteAddress,
			"lastCheckin":   lastCheckin,
			"nextCheckin":   nextCheckin,
			"interval":      beacon.Interval,
			"jitter":        beacon.Jitter,
			"isDead":        beacon.IsDead,
		})
	}

	result, err := json.Marshal(map[string]interface{}{
		"beacons": formattedBeacons,
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

func HandleGetBeacon(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	beaconID, ok := arguments["beaconID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("beaconID must be a string")
	}

	beacon, err := client.GetBeacon(ctx, beaconID)
	if err != nil {
		return nil, err
	}

	lastCheckin := time.Unix(0, beacon.LastCheckin).Format(time.RFC3339)
	nextCheckin := time.Unix(0, beacon.NextCheckin).Format(time.RFC3339)

	formattedBeacon := map[string]interface{}{
		"id":                  beacon.ID,
		"name":                beacon.Name,
		"hostname":            beacon.Hostname,
		"os":                  beacon.OS,
		"arch":                beacon.Arch,
		"username":            beacon.Username,
		"pid":                 beacon.PID,
		"transport":           beacon.Transport,
		"remoteAddress":       beacon.RemoteAddress,
		"lastCheckin":         lastCheckin,
		"nextCheckin":         nextCheckin,
		"interval":            beacon.Interval,
		"jitter":              beacon.Jitter,
		"tasksCount":          beacon.TasksCount,
		"tasksCountCompleted": beacon.TasksCountCompleted,
		"isDead":              beacon.IsDead,
	}

	result, err := json.Marshal(formattedBeacon)
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

func HandleRemoveBeacon(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	beaconID, ok := arguments["beaconID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("beaconID must be a string")
	}

	_, err := client.RmBeacon(ctx, beaconID)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully removed beacon with ID %s", beaconID),
			},
		},
	}, nil
}

func HandleGetBeaconTasks(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	beaconID, ok := arguments["beaconID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("beaconID must be a string")
	}

	tasks, err := client.GetBeaconTasks(ctx, beaconID)
	if err != nil {
		return nil, err
	}

	var formattedTasks []map[string]interface{}
	for _, task := range tasks.Tasks {
				state := task.State
		if state == "" {
			state = "unknown"
		}

		formattedTasks = append(formattedTasks, map[string]interface{}{
			"id":          task.ID,
			"description": task.Description,
			"state":       state,
			"sentAt":      task.SentAt,
			"completedAt": task.CompletedAt,
			"createdAt":   task.CreatedAt,
		})
	}

	result, err := json.Marshal(map[string]interface{}{
		"beaconID": beaconID,
		"tasks":    formattedTasks,
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

// Not working in client yet
func HandleCancelBeaconTask(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	beaconID, ok := arguments["beaconID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("beaconID must be a string")
	}

	taskID, ok := arguments["taskID"].(string)
	if !ok {
		return nil, NewInvalidArgsError("taskID must be a string")
	}

	// Not calling client.CancelBeaconTask due to implementation issues
	// cancelledTask, err := client.CancelBeaconTask(ctx, beaconID, taskID)
	// if err != nil {
	//	return nil, err
	// }

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Cancel beacon task not yet fully implemented. Would cancel task %s for beacon %s", taskID, beaconID),
			},
		},
	}, nil
}

func HandleListJobs(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	jobs, err := client.GetJobs(ctx)
	if err != nil {
		return nil, err
	}

	var formattedJobs []map[string]interface{}
	for _, job := range jobs.Active {
		formattedJobs = append(formattedJobs, map[string]interface{}{
			"id":          job.ID,
			"name":        job.Name,
			"description": job.Description,
			"protocol":    job.Protocol,
			"port":        job.Port,
			"domains":     job.Domains,
		})
	}

	result, err := json.Marshal(map[string]interface{}{
		"jobs": formattedJobs,
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

func HandleKillJob(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	jobIDFloat, ok := arguments["jobID"].(float64)
	if !ok {
		return nil, NewInvalidArgsError("jobID must be a number")
	}
	jobID := uint32(jobIDFloat)

	killJob, err := client.KillJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	success := "failed"
	if killJob.Success {
		success = "succeeded"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Kill job %d %s", jobID, success),
			},
		},
	}, nil
}
