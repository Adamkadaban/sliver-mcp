package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adamkadaban/sliver-mcp/internal/client"
	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/mark3labs/mcp-go/mcp"
)

// ImplantConfig holds global configuration for implant generation
var ImplantConfig struct {
	OutputDir string // Directory where generated implants are saved
}

func init() {
	// Default output directory
	ImplantConfig.OutputDir = "implants"
}

func HandleGenerateImplant(ctx context.Context, request mcp.CallToolRequest, client *client.SliverClient) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Check if the client explicitly requests to include binary data
	includeBinary := false
	if includeParam, ok := arguments["includeBinary"].(bool); ok {
		includeBinary = includeParam
	}

	configMap, ok := arguments["config"].(map[string]interface{})
	if !ok {
		return nil, NewInvalidArgsError("config must be an object with OS, architecture, and C2 configuration")
	}

	implantConfig := &clientpb.ImplantConfig{}

	// Handle OS - convert to lowercase for consistent handling
	if goos, ok := configMap["OS"].(string); ok {
		implantConfig.GOOS = strings.ToLower(goos)
	} else if goos, ok := configMap["os"].(string); ok {
		implantConfig.GOOS = strings.ToLower(goos)
	}

	// Standardize OS name
	switch implantConfig.GOOS {
	case "mac", "macos", "osx":
		implantConfig.GOOS = "darwin"
	case "win":
		implantConfig.GOOS = "windows"
	case "lin":
		implantConfig.GOOS = "linux"
	}

	// Handle architecture - convert to standard Go arch format
	if goarch, ok := configMap["arch"].(string); ok {
		goarch = strings.ToLower(goarch)
		// Standardize architecture names
		switch goarch {
		case "x64", "x86_64", "amd64", "64", "64bit":
			implantConfig.GOARCH = "amd64"
		case "x86", "i386", "386", "32", "32bit":
			implantConfig.GOARCH = "386"
		case "arm64", "aarch64":
			implantConfig.GOARCH = "arm64"
		default:
			implantConfig.GOARCH = goarch
		}
	}

	// Set defaults if not provided
	if implantConfig.GOOS == "" {
		implantConfig.GOOS = "windows" // Default to windows if not specified
	}
	if implantConfig.GOARCH == "" {
		implantConfig.GOARCH = "amd64" // Default to amd64 if not specified
	}

	// Validate supported platforms
	platform := fmt.Sprintf("%s/%s", implantConfig.GOOS, implantConfig.GOARCH)
	supportedPlatforms := map[string]bool{
		"windows/amd64": true,
		"windows/386":   true,
		"linux/amd64":   true,
		"linux/386":     true,
		"darwin/amd64":  true,
		"darwin/arm64":  true,
	}
	if _, ok := supportedPlatforms[platform]; !ok {
		return nil, NewInvalidArgsError(fmt.Sprintf("unsupported platform: %s - supported platforms are: windows/amd64, windows/386, linux/amd64, linux/386, darwin/amd64, darwin/arm64", platform))
	}

	// Always use EXECUTABLE format by default to avoid compatibility issues
	implantConfig.Format = clientpb.OutputFormat_EXECUTABLE

	// Only set other format flags if explicitly specified
	if format, ok := configMap["format"].(string); ok {
		format = strings.ToLower(format)

		// Only use executable format for now
		switch {
		case format == "shared-lib" || format == "sharedlib" || format == "dll" || format == "so" || format == "dylib":
			fmt.Printf("WARNING: Shared library format requested, but defaulting to executable for compatibility\n")
			// Disabled formats
			// implantConfig.Format = clientpb.OutputFormat_SHARED_LIB
			// implantConfig.IsSharedLib = true
		case format == "shellcode":
			fmt.Printf("WARNING: Shellcode format requested, but defaulting to executable for compatibility\n")
			// Disabled formats
			// implantConfig.Format = clientpb.OutputFormat_SHELLCODE
			// implantConfig.IsShellcode = true
		case format == "service":
			fmt.Printf("WARNING: Service format requested, but defaulting to executable for compatibility\n")
			// Disabled formats
			// implantConfig.Format = clientpb.OutputFormat_SERVICE
			// implantConfig.IsService = true
		}
	}

	// Process C2 configuration
	foundC2Config := false
	if c2Configs, ok := configMap["c2"].([]interface{}); ok && len(c2Configs) > 0 {
		foundC2Config = true
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
	} else if mtlsC2, ok := configMap["mtlsc2"].([]interface{}); ok {
		foundC2Config = true
		for i, server := range mtlsC2 {
			if serverStr, ok := server.(string); ok {
				if !strings.HasPrefix(serverStr, "mtls://") {
					serverStr = fmt.Sprintf("mtls://%s", serverStr)
				}
				implantConfig.C2 = append(implantConfig.C2, &clientpb.ImplantC2{
					Priority: uint32(i), // #nosec G115
					URL:      serverStr,
				})
			} else {
				fmt.Printf("WARNING: Invalid mtlsc2 server format at index %d: %v\n", i, server)
			}
		}

		if len(implantConfig.C2) == 0 && len(mtlsC2) > 0 {
			// We had mtlsC2 entries but none were valid
			return nil, NewInvalidArgsError(fmt.Sprintf("invalid mtlsc2 configuration: %+v - mtlsc2 must be an array of strings", mtlsC2))
		}
	} else if httpC2, ok := configMap["httpc2"].([]interface{}); ok && len(httpC2) > 0 {
		foundC2Config = true
		for i, server := range httpC2 {
			if serverStr, ok := server.(string); ok {
				if !strings.HasPrefix(serverStr, "http://") {
					serverStr = fmt.Sprintf("http://%s", serverStr)
				}
				implantConfig.C2 = append(implantConfig.C2, &clientpb.ImplantC2{
					Priority: uint32(i), // #nosec G115
					URL:      serverStr,
				})
			}
		}
	} else if httpsC2, ok := configMap["httpsc2"].([]interface{}); ok && len(httpsC2) > 0 {
		foundC2Config = true
		for i, server := range httpsC2 {
			if serverStr, ok := server.(string); ok {
				if !strings.HasPrefix(serverStr, "https://") {
					serverStr = fmt.Sprintf("https://%s", serverStr)
				}
				implantConfig.C2 = append(implantConfig.C2, &clientpb.ImplantC2{
					Priority: uint32(i), // #nosec G115
					URL:      serverStr,
				})
			}
		}
	} else if c2URL, ok := configMap["c2"].(string); ok && c2URL != "" {
		foundC2Config = true
		implantConfig.C2 = append(implantConfig.C2, &clientpb.ImplantC2{
			Priority: 0,
			URL:      c2URL,
		})
	}

	if !foundC2Config || len(implantConfig.C2) == 0 {
		return nil, NewInvalidArgsError("no C2 configuration found - please specify at least one C2 server using mtlsc2, httpc2, httpsc2, or c2 fields")
	}

	// Validate C2 URLs
	for i, c2 := range implantConfig.C2 {
		if c2.URL == "" {
			return nil, NewInvalidArgsError("empty C2 URL detected - please provide valid C2 URLs")
		}

		// Validate URL format based on protocol
		switch {
		case strings.HasPrefix(c2.URL, "mtls://"):
			parts := strings.Split(strings.TrimPrefix(c2.URL, "mtls://"), ":")
			if len(parts) != 2 && (len(parts) != 1 || parts[0] == "") {
				return nil, NewInvalidArgsError(fmt.Sprintf("invalid MTLS URL format at index %d: %s - format should be mtls://host:port or mtls://host", i, c2.URL))
			}
		case strings.HasPrefix(c2.URL, "http://"), strings.HasPrefix(c2.URL, "https://"):
			if strings.Count(c2.URL, "/") < 3 {
				return nil, NewInvalidArgsError(fmt.Sprintf("invalid HTTP(S) URL format at index %d: %s - format should include host, e.g., http://domain.com", i, c2.URL))
			}
		default:
			return nil, NewInvalidArgsError(fmt.Sprintf("invalid URL protocol at index %d: %s - supported protocols are mtls://, http://, and https://", i, c2.URL))
		}
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

	// Log detailed configuration information
	fmt.Printf("Generating implant with platform: %s, format: %s\n", platform, implantConfig.Format.String())
	fmt.Printf("C2 Endpoints: %d configured\n", len(implantConfig.C2))

	// Check for required toolchain components based on target platform
	if implantConfig.GOOS == "windows" && implantConfig.GOARCH == "amd64" {
		compilerPath := "/usr/bin/x86_64-w64-mingw32-gcc"
		if _, err := os.Stat(compilerPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("missing required compiler for Windows/amd64: %s - please install mingw-w64 package", compilerPath)
		}
	} else if implantConfig.GOOS == "windows" && implantConfig.GOARCH == "386" {
		compilerPath := "/usr/bin/i686-w64-mingw32-gcc"
		if _, err := os.Stat(compilerPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("missing required compiler for Windows/386: %s - please install mingw-w64 package", compilerPath)
		}
	}

	generate, err := client.Generate(ctx, implantConfig, implantName)
	if err != nil {
		errorMsg := err.Error()

		// Provide more specific error messages based on known error patterns
		if strings.Contains(errorMsg, "exit status 1") {
			// Try to identify common build failures
			if strings.Contains(errorMsg, "CC") || strings.Contains(errorMsg, "gcc") {
				return nil, fmt.Errorf("cross-compiler error: required cross-compiler not found or failed. For Windows targets, ensure mingw-w64 is installed. For Linux targets, ensure gcc is installed")
			}
			if strings.Contains(errorMsg, "unknown architecture") || strings.Contains(errorMsg, "unknown OS") {
				return nil, fmt.Errorf("unsupported platform: %s - please choose a supported OS/architecture combination", platform)
			}
			if strings.Contains(errorMsg, "connection") || strings.Contains(errorMsg, "network") {
				return nil, fmt.Errorf("network error while building implant: server may be unavailable or connection disrupted")
			}

			// Generic build error
			return nil, fmt.Errorf("build failed: compilation error during implant generation - check server logs for details")
		}

		// If the error is not one of the recognized patterns, return the original with additional context
		return nil, fmt.Errorf("failed to generate implant: %v - check server logs for more details", err)
	}

	if generate.File == nil || generate.File.Data == nil {
		return nil, fmt.Errorf("failed to generate implant: no file data returned from server - build may have failed")
	}

	// Determine output directory - use per-request override if provided
	outputDir := ImplantConfig.OutputDir
	if customOutputDir, ok := arguments["outputDir"].(string); ok && customOutputDir != "" {
		outputDir = customOutputDir
	}

	// Create output directory if it doesn't exist
	if mkdirErr := os.MkdirAll(outputDir, 0700); mkdirErr != nil {
		return nil, fmt.Errorf("failed to create output directory %s: %v", outputDir, mkdirErr)
	}

	// Save the implant to disk
	savePath := filepath.Join(outputDir, generate.File.Name)
	if writeErr := os.WriteFile(savePath, generate.File.Data, 0600); writeErr != nil {
		return nil, fmt.Errorf("failed to save implant to disk: %v", writeErr)
	}

	// Get absolute path to show to user
	absPath := savePath
	if absPathResult, pathErr := filepath.Abs(savePath); pathErr == nil {
		absPath = absPathResult
	}

	// Prepare response data
	responseData := map[string]interface{}{
		"name":      implantName,
		"os":        implantConfig.GOOS,
		"arch":      implantConfig.GOARCH,
		"format":    implantConfig.Format.String(),
		"isBeacon":  implantConfig.IsBeacon,
		"fileSize":  len(generate.File.Data),
		"fileName":  generate.File.Name,
		"filePath":  absPath,
		"generated": true,
	}

	// Only include binary data if explicitly requested
	if includeBinary {
		responseData["fileData"] = base64.StdEncoding.EncodeToString(generate.File.Data)
		responseData["message"] = fmt.Sprintf("Implant generated successfully and saved to %s. Binary data also included in response.", absPath)
	} else {
		responseData["message"] = fmt.Sprintf("Implant generated successfully and saved to %s. Binary data not included in response to prevent LLM context overflow.", absPath)
	}

	result, err := json.Marshal(responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize result: %v", err)
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

	// Handle OS - convert to lowercase for consistent handling
	if goos, ok := configMap["OS"].(string); ok {
		implantConfig.GOOS = strings.ToLower(goos)
	} else if goos, ok := configMap["os"].(string); ok {
		implantConfig.GOOS = strings.ToLower(goos)
	}

	// Standardize OS name
	switch implantConfig.GOOS {
	case "mac", "macos", "osx":
		implantConfig.GOOS = "darwin"
	case "win":
		implantConfig.GOOS = "windows"
	case "lin":
		implantConfig.GOOS = "linux"
	}

	// Handle architecture - convert to standard Go arch format
	if goarch, ok := configMap["arch"].(string); ok {
		goarch = strings.ToLower(goarch)
		// Standardize architecture names
		switch goarch {
		case "x64", "x86_64", "amd64", "64", "64bit":
			implantConfig.GOARCH = "amd64"
		case "x86", "i386", "386", "32", "32bit":
			implantConfig.GOARCH = "386"
		case "arm64", "aarch64":
			implantConfig.GOARCH = "arm64"
		default:
			implantConfig.GOARCH = goarch
		}
	}

	// Set defaults if not provided
	if implantConfig.GOOS == "" {
		implantConfig.GOOS = "windows" // Default to windows if not specified
	}
	if implantConfig.GOARCH == "" {
		implantConfig.GOARCH = "amd64" // Default to amd64 if not specified
	}

	// Always use EXECUTABLE format by default
	implantConfig.Format = clientpb.OutputFormat_EXECUTABLE

	// Show warnings for unsupported formats
	if format, ok := configMap["format"].(string); ok {
		format = strings.ToLower(format)

		// We'll only use executable format to avoid compatibility issues
		switch {
		case format == "shared-lib" || format == "sharedlib" || format == "dll" || format == "so" || format == "dylib":
			fmt.Printf("WARNING: Shared library format requested in profile, but defaulting to executable\n")
		case format == "shellcode":
			fmt.Printf("WARNING: Shellcode format requested in profile, but defaulting to executable\n")
		case format == "service":
			fmt.Printf("WARNING: Service format requested in profile, but defaulting to executable\n")
		}
	}

	// Log format configuration
	fmt.Printf("Profile format: Format=%d, IsSharedLib=%v, IsShellcode=%v, IsService=%v\n",
		implantConfig.Format, implantConfig.IsSharedLib, implantConfig.IsShellcode, implantConfig.IsService)

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

	// Check if the client explicitly requests to include binary data
	includeBinary := false
	if includeParam, ok := arguments["includeBinary"].(bool); ok {
		includeBinary = includeParam
	}

	implantName, ok := arguments["implantName"].(string)
	if !ok || implantName == "" {
		return nil, NewInvalidArgsError("implantName must be a non-empty string")
	}

	generate, err := client.Regenerate(ctx, implantName)
	if err != nil {
		return nil, err
	}

	// Determine output directory - use per-request override if provided
	outputDir := ImplantConfig.OutputDir
	if customOutputDir, ok := arguments["outputDir"].(string); ok && customOutputDir != "" {
		outputDir = customOutputDir
	}

	// Create output directory if it doesn't exist
	if mkdirErr := os.MkdirAll(outputDir, 0700); mkdirErr != nil {
		return nil, fmt.Errorf("failed to create output directory %s: %v", outputDir, mkdirErr)
	}

	// Save the implant to disk
	savePath := filepath.Join(outputDir, generate.File.Name)
	if writeErr := os.WriteFile(savePath, generate.File.Data, 0600); writeErr != nil {
		return nil, fmt.Errorf("failed to save implant to disk: %v", writeErr)
	}

	// Get absolute path to show to user
	absPath := savePath
	if absPathResult, pathErr := filepath.Abs(savePath); pathErr == nil {
		absPath = absPathResult
	}

	// Prepare response data
	responseData := map[string]interface{}{
		"name":      implantName,
		"fileSize":  len(generate.File.Data),
		"fileName":  generate.File.Name,
		"filePath":  absPath,
		"generated": true,
	}

	// Only include binary data if explicitly requested
	if includeBinary {
		responseData["fileData"] = base64.StdEncoding.EncodeToString(generate.File.Data)
		responseData["message"] = fmt.Sprintf("Implant regenerated successfully and saved to %s. Binary data also included in response.", absPath)
	} else {
		responseData["message"] = fmt.Sprintf("Implant regenerated successfully and saved to %s. Binary data not included in response to prevent LLM context overflow.", absPath)
	}

	result, err := json.Marshal(responseData)
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

// Handle the generateStage tool, but with compatibility note
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

	// NOTE: GenerateStage is not implemented in the client due to protobuf compatibility issues
	// with sliver version v1.5.x. Will need to update sliver version or adapt to available API.

	paramJSON, _ := json.Marshal(map[string]interface{}{
		"profile":       profile,
		"name":          name,
		"aesEncryptKey": arguments["aesEncryptKey"],
		"aesEncryptIv":  arguments["aesEncryptIv"],
		"rc4EncryptKey": arguments["rc4EncryptKey"],
		"compress":      arguments["compress"],
		"compressF":     arguments["compressF"],
		"prependSize":   arguments["prependSize"],
	})

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Stager generation not implemented due to compatibility issues with sliver v1.5.x. Would generate a stager with parameters: %s", string(paramJSON)),
			},
		},
	}, nil
}
