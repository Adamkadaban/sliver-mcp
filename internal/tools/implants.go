package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/adamkadaban/sliver-mcp/internal/client"
	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/mark3labs/mcp-go/mcp"
)

func HandleGenerateImplant(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	configMap, ok := arguments["config"].(map[string]interface{})
	if !ok {
		return nil, NewInvalidArgsError("config must be an object")
	}

	implantConfig := &clientpb.ImplantConfig{}

	if goos, ok := configMap["os"].(string); ok {
		implantConfig.GOOS = goos
	}
	if goarch, ok := configMap["arch"].(string); ok {
		implantConfig.GOARCH = goarch
	}

	if format, ok := configMap["format"].(string); ok {
		switch format {
		case "executable":
			implantConfig.Format = clientpb.OutputFormat_EXECUTABLE
		case "shared-lib", "sharedlib":
			implantConfig.Format = clientpb.OutputFormat_SHARED_LIB
			implantConfig.IsSharedLib = true
		case "shellcode":
			implantConfig.Format = clientpb.OutputFormat_SHELLCODE
			implantConfig.IsShellcode = true
		case "service":
			implantConfig.Format = clientpb.OutputFormat_SERVICE
			implantConfig.IsService = true
		}
	}

	if c2Configs, ok := configMap["c2"].([]interface{}); ok {
		for i, c2Config := range c2Configs {
			if c2Map, ok := c2Config.(map[string]interface{}); ok {
				c2 := &clientpb.ImplantC2{
					Priority: uint32(i), // #nosec G115 - priority index will always be small
				}

				if url, ok := c2Map["url"].(string); ok {
					c2.URL = url
				}

				implantConfig.C2 = append(implantConfig.C2, c2)
			}
		}
	} else if c2URL, ok := configMap["c2"].(string); ok {
		implantConfig.C2 = append(implantConfig.C2, &clientpb.ImplantC2{
			Priority: 0,
			URL:      c2URL,
		})
	}

	if isBeacon, ok := configMap["isBeacon"].(bool); ok {
		implantConfig.IsBeacon = isBeacon
	}
	if beaconInterval, ok := configMap["beaconInterval"].(float64); ok {
		implantConfig.BeaconInterval = int64(beaconInterval)
	}
	if beaconJitter, ok := configMap["beaconJitter"].(float64); ok {
		implantConfig.BeaconJitter = int64(beaconJitter)
	}

	if debug, ok := configMap["debug"].(bool); ok {
		implantConfig.Debug = debug
	}
	if evasion, ok := configMap["evasion"].(bool); ok {
		implantConfig.Evasion = evasion
	}
	if obfuscateSymbols, ok := configMap["obfuscateSymbols"].(bool); ok {
		implantConfig.ObfuscateSymbols = obfuscateSymbols
	}

	implantName := "generated-implant"
	if name, ok := arguments["name"].(string); ok && name != "" {
		implantName = name
	}

	generate, err := client.Generate(ctx, implantConfig, implantName)
	if err != nil {
		return nil, err
	}

	fileData := ""
	if generate.File != nil && generate.File.Data != nil {
		fileData = base64.StdEncoding.EncodeToString(generate.File.Data)
	}

	result, err := json.Marshal(map[string]interface{}{
		"name":     implantName,
		"os":       implantConfig.GOOS,
		"arch":     implantConfig.GOARCH,
		"format":   implantConfig.Format.String(),
		"isBeacon": implantConfig.IsBeacon,
		"fileSize": len(generate.File.Data),
		"fileName": generate.File.Name,
		"fileData": fileData,
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

func HandleListImplantProfiles(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	profiles, err := client.ImplantProfiles(ctx)
	if err != nil {
		return nil, err
	}

	var formattedProfiles []map[string]interface{}
	for _, profile := range profiles.Profiles {
		formattedProfiles = append(formattedProfiles, map[string]interface{}{
			"name":     profile.Name,
			"os":       profile.Config.GOOS,
			"arch":     profile.Config.GOARCH,
			"format":   profile.Config.Format.String(),
			"isBeacon": profile.Config.IsBeacon,
			"c2Count":  len(profile.Config.C2),
		})
	}

	result, err := json.Marshal(map[string]interface{}{
		"profiles": formattedProfiles,
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

func HandleSaveImplantProfile(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	name, ok := arguments["name"].(string)
	if !ok || name == "" {
		return nil, NewInvalidArgsError("name must be a non-empty string")
	}

	configMap, ok := arguments["config"].(map[string]interface{})
	if !ok {
		return nil, NewInvalidArgsError("config must be an object")
	}

	implantConfig := &clientpb.ImplantConfig{}

	if goos, ok := configMap["os"].(string); ok {
		implantConfig.GOOS = goos
	}
	if goarch, ok := configMap["arch"].(string); ok {
		implantConfig.GOARCH = goarch
	}

	if format, ok := configMap["format"].(string); ok {
		switch format {
		case "executable":
			implantConfig.Format = clientpb.OutputFormat_EXECUTABLE
		case "shared-lib", "sharedlib":
			implantConfig.Format = clientpb.OutputFormat_SHARED_LIB
			implantConfig.IsSharedLib = true
		case "shellcode":
			implantConfig.Format = clientpb.OutputFormat_SHELLCODE
			implantConfig.IsShellcode = true
		case "service":
			implantConfig.Format = clientpb.OutputFormat_SERVICE
			implantConfig.IsService = true
		}
	}

	if c2Configs, ok := configMap["c2"].([]interface{}); ok {
		for i, c2Config := range c2Configs {
			if c2Map, ok := c2Config.(map[string]interface{}); ok {
				c2 := &clientpb.ImplantC2{
					Priority: uint32(i), // #nosec G115 - priority index will always be small
				}

				if url, ok := c2Map["url"].(string); ok {
					c2.URL = url
				}

				implantConfig.C2 = append(implantConfig.C2, c2)
			}
		}
	} else if c2URL, ok := configMap["c2"].(string); ok {
		implantConfig.C2 = append(implantConfig.C2, &clientpb.ImplantC2{
			Priority: 0,
			URL:      c2URL,
		})
	}

	if isBeacon, ok := configMap["isBeacon"].(bool); ok {
		implantConfig.IsBeacon = isBeacon
	}
	if beaconInterval, ok := configMap["beaconInterval"].(float64); ok {
		implantConfig.BeaconInterval = int64(beaconInterval)
	}
	if beaconJitter, ok := configMap["beaconJitter"].(float64); ok {
		implantConfig.BeaconJitter = int64(beaconJitter)
	}

	if debug, ok := configMap["debug"].(bool); ok {
		implantConfig.Debug = debug
	}
	if evasion, ok := configMap["evasion"].(bool); ok {
		implantConfig.Evasion = evasion
	}
	if obfuscateSymbols, ok := configMap["obfuscateSymbols"].(bool); ok {
		implantConfig.ObfuscateSymbols = obfuscateSymbols
	}

	profile := &clientpb.ImplantProfile{
		Name:   name,
		Config: implantConfig,
	}

	savedProfile, err := client.SaveImplantProfile(ctx, profile)
	if err != nil {
		return nil, err
	}

	result, err := json.Marshal(map[string]interface{}{
		"name":     savedProfile.Name,
		"os":       savedProfile.Config.GOOS,
		"arch":     savedProfile.Config.GOARCH,
		"format":   savedProfile.Config.Format.String(),
		"isBeacon": savedProfile.Config.IsBeacon,
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

func HandleDeleteImplantProfile(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	profileID, ok := arguments["profileID"].(string)
	if !ok || profileID == "" {
		return nil, NewInvalidArgsError("profileID must be a non-empty string")
	}

	_, err := client.DeleteImplantProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully deleted implant profile with ID %s", profileID),
			},
		},
	}, nil
}

func HandleListImplantBuilds(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	builds, err := client.ImplantBuilds(ctx)
	if err != nil {
		return nil, err
	}

	var formattedBuilds []map[string]interface{}
	for _, build := range builds.Configs {
		formattedBuilds = append(formattedBuilds, map[string]interface{}{
			"id":       build.ID,
			"name":     build.Name,
			"os":       build.GOOS,
			"arch":     build.GOARCH,
			"format":   build.Format.String(),
			"isBeacon": build.IsBeacon,
			"c2Count":  len(build.C2),
		})
	}

	result, err := json.Marshal(map[string]interface{}{
		"builds": formattedBuilds,
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

func HandleRegenerateImplant(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	implantName, ok := arguments["implantName"].(string)
	if !ok || implantName == "" {
		return nil, NewInvalidArgsError("implantName must be a non-empty string")
	}

	generate, err := client.Regenerate(ctx, implantName)
	if err != nil {
		return nil, err
	}

	fileData := ""
	if generate.File != nil && generate.File.Data != nil {
		fileData = base64.StdEncoding.EncodeToString(generate.File.Data)
	}

	result, err := json.Marshal(map[string]interface{}{
		"name":     implantName,
		"fileSize": len(generate.File.Data),
		"fileName": generate.File.Name,
		"fileData": fileData,
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

// Not working on client
func HandleGenerateStager(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	profile, ok := arguments["profile"].(string)
	if !ok || profile == "" {
		return nil, NewInvalidArgsError("profile must be a non-empty string")
	}

	name := "generated-stager"
	if nameArg, ok := arguments["name"].(string); ok && nameArg != "" {
		name = nameArg
	}

	params := map[string]interface{}{
		"profile": profile,
		"name":    name,
	}

	if aesKey, ok := arguments["aesEncryptKey"].(string); ok {
		params["aesEncryptKey"] = aesKey
	}
	if aesIv, ok := arguments["aesEncryptIv"].(string); ok {
		params["aesEncryptIv"] = aesIv
	}
	if rc4Key, ok := arguments["rc4EncryptKey"].(string); ok {
		params["rc4EncryptKey"] = rc4Key
	}
	if compress, ok := arguments["compress"].(string); ok {
		params["compress"] = compress
	}
	if compressF, ok := arguments["compressF"].(string); ok {
		params["compressF"] = compressF
	}
	if prependSize, ok := arguments["prependSize"].(bool); ok {
		params["prependSize"] = prependSize
	}

	paramJSON, _ := json.Marshal(params)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Stager generation not yet implemented. Would generate a stager with parameters: %s", string(paramJSON)),
			},
		},
	}, nil
}
