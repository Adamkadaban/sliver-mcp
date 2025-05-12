package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bishopfox/sliver/client/assets"
	"github.com/bishopfox/sliver/client/transport"
	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/bishopfox/sliver/protobuf/rpcpb"
	"github.com/bishopfox/sliver/protobuf/sliverpb"
	"google.golang.org/grpc"
)

type SliverClient struct {
	RPCClient    rpcpb.SliverRPCClient
	GRPCConn     *grpc.ClientConn
	ConfigPath   string
	ClientConfig *assets.ClientConfig
}

func NewSliverClient(configPath string) (*SliverClient, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	rpcClient, conn, err := transport.MTLSConnect(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Sliver server: %v", err)
	}

	return &SliverClient{
		RPCClient:    rpcClient,
		GRPCConn:     conn,
		ConfigPath:   configPath,
		ClientConfig: config,
	}, nil
}

func loadConfig(configPath string) (*assets.ClientConfig, error) {
	if configPath == "" {
		// Use first config found if not specified
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %v", err)
		}

		configsDir := filepath.Join(homeDir, ".sliver-client/configs")
		entries, err := os.ReadDir(configsDir)
		if err != nil {
			return nil, fmt.Errorf("unable to find configurations automatically in %s: %v", configsDir, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".cfg") {
				configPath = filepath.Join(configsDir, entry.Name())
				break
			}
		}

		if configPath == "" {
			return nil, fmt.Errorf("no configuration files found in %s", configsDir)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config assets.ClientConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *SliverClient) Close() error {
	if c.GRPCConn != nil {
		return c.GRPCConn.Close()
	}
	return nil
}

func (c *SliverClient) GetVersion(ctx context.Context) (*clientpb.Version, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	version, err := c.RPCClient.GetVersion(ctx, &commonpb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %v", err)
	}

	return version, nil
}

func (c *SliverClient) GetSessions(ctx context.Context) (*clientpb.Sessions, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	sessions, err := c.RPCClient.GetSessions(ctx, &commonpb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %v", err)
	}

	return sessions, nil
}

func (c *SliverClient) GetBeacons(ctx context.Context) (*clientpb.Beacons, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	beacons, err := c.RPCClient.GetBeacons(ctx, &commonpb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get beacons: %v", err)
	}

	return beacons, nil
}

func (c *SliverClient) GetBeacon(ctx context.Context, beaconID string) (*clientpb.Beacon, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	beacon, err := c.RPCClient.GetBeacon(ctx, &clientpb.Beacon{ID: beaconID})
	if err != nil {
		return nil, fmt.Errorf("failed to get beacon: %v", err)
	}

	return beacon, nil
}

func (c *SliverClient) GetJobs(ctx context.Context) (*clientpb.Jobs, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	jobs, err := c.RPCClient.GetJobs(ctx, &commonpb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %v", err)
	}

	return jobs, nil
}

func (c *SliverClient) Generate(ctx context.Context, config *clientpb.ImplantConfig, name string) (*clientpb.Generate, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
	}

	// Validate and normalize config
	if config.GOOS == "" {
		config.GOOS = "windows" // Default to windows if not specified
	}

	// Standardize OS name
	config.GOOS = strings.ToLower(config.GOOS)
	switch config.GOOS {
	case "mac", "macos", "osx":
		config.GOOS = "darwin"
	case "win":
		config.GOOS = "windows"
	case "lin":
		config.GOOS = "linux"
	}

	// Standardize architecture name
	if config.GOARCH == "" {
		config.GOARCH = "amd64" // Default to amd64 if not specified
	}
	config.GOARCH = strings.ToLower(config.GOARCH)
	if config.GOARCH == "x64" || config.GOARCH == "x86_64" || strings.HasPrefix(config.GOARCH, "64") {
		config.GOARCH = "amd64"
	} else if config.GOARCH == "x86" || config.GOARCH == "i386" || strings.HasPrefix(config.GOARCH, "32") {
		config.GOARCH = "386"
	}

	// Verify the platform is supported
	platform := fmt.Sprintf("%s/%s", config.GOOS, config.GOARCH)
	supportedPlatforms := map[string]bool{
		"darwin/amd64":  true,
		"darwin/arm64":  true,
		"linux/386":     true,
		"linux/amd64":   true,
		"windows/386":   true,
		"windows/amd64": true,
	}
	if _, ok := supportedPlatforms[platform]; !ok {
		fmt.Printf("Warning: Potentially unsupported platform %s\n", platform)
	}

	// Set the name in the config if provided
	if name != "" {
		config.Name = name
	}

	// Debug log
	fmt.Printf("Generating implant with OS=%s, ARCH=%s, FORMAT=%s, Platform=%s\n",
		config.GOOS, config.GOARCH, config.Format.String(), platform)

	// Print detailed debug info
	fmt.Printf("==== DEBUG INFO ====\n")
	fmt.Printf("GOOS: %s, GOARCH: %s, Format: %s\n", config.GOOS, config.GOARCH, config.Format.String())
	fmt.Printf("IsBeacon: %v, IsSharedLib: %v, IsService: %v, IsShellcode: %v\n",
		config.IsBeacon, config.IsSharedLib, config.IsService, config.IsShellcode)

	// Check for cross-compiler availability
	cc32Path := os.Getenv("SLIVER_CC_32")
	cc64Path := os.Getenv("SLIVER_CC_64")
	fmt.Printf("SLIVER_CC_32: %s\n", cc32Path)
	fmt.Printf("SLIVER_CC_64: %s\n", cc64Path)

	// Check for mingw compilers
	if config.GOOS == "windows" {
		// For 32-bit Windows target
		if config.GOARCH == "386" {
			compilerPath := "/usr/bin/i686-w64-mingw32-gcc"
			if _, err := os.Stat(compilerPath); os.IsNotExist(err) {
				fmt.Printf("Warning: 32-bit Windows cross-compiler not found at %s\n", compilerPath)
			} else {
				fmt.Printf("Found 32-bit Windows cross-compiler at %s\n", compilerPath)
			}
		}

		// For 64-bit Windows target
		if config.GOARCH == "amd64" {
			compilerPath := "/usr/bin/x86_64-w64-mingw32-gcc"
			if _, err := os.Stat(compilerPath); os.IsNotExist(err) {
				fmt.Printf("Warning: 64-bit Windows cross-compiler not found at %s\n", compilerPath)
			} else {
				fmt.Printf("Found 64-bit Windows cross-compiler at %s\n", compilerPath)
			}
		}
	}

	// Make the RPC call
	generateReq := &clientpb.GenerateReq{
		Config: config,
	}

	generate, err := c.RPCClient.Generate(ctx, generateReq)
	if err != nil {
		// Try to provide more context on the error
		errorMsg := err.Error()

		if strings.Contains(errorMsg, "CC") || strings.Contains(errorMsg, "gcc") || strings.Contains(errorMsg, "mingw") {
			fmt.Printf("ERROR: This appears to be a cross-compiler issue.\n")
			fmt.Printf("For Windows targets, ensure mingw-w64 is installed.\n")
			fmt.Printf("For Linux targets, ensure gcc and appropriate dev packages are installed.\n")
			fmt.Printf("You can set SLIVER_CC_32 and SLIVER_CC_64 environment variables to point to the compilers.\n")
			return nil, fmt.Errorf("failed to generate implant: cross-compiler error - %v", err)
		}

		if strings.Contains(errorMsg, "go build") || strings.Contains(errorMsg, "exit status") {
			fmt.Printf("ERROR: Go build process failed.\n")
			fmt.Printf("This could be due to missing dependencies or incompatible Go version.\n")
			fmt.Printf("Check the server logs for detailed build output.\n")
			return nil, fmt.Errorf("failed to generate implant: build process error - %v", err)
		}

		if strings.Contains(errorMsg, "connect") || strings.Contains(errorMsg, "network") || strings.Contains(errorMsg, "connection") {
			fmt.Printf("ERROR: Network connection issue detected.\n")
			fmt.Printf("Ensure the Sliver server is running and accessible.\n")
			return nil, fmt.Errorf("failed to generate implant: network error - %v", err)
		}

		// Generic error with detailed context
		return nil, fmt.Errorf("failed to generate implant: %v (platform: %s, format: %s)",
			err, platform, config.Format.String())
	}

	return generate, nil
}

func (c *SliverClient) Regenerate(ctx context.Context, implantName string) (*clientpb.Generate, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
	}

	generate, err := c.RPCClient.Regenerate(ctx, &clientpb.RegenerateReq{
		ImplantName: implantName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to regenerate implant: %v", err)
	}

	return generate, nil
}

func (c *SliverClient) ImplantProfiles(ctx context.Context) (*clientpb.ImplantProfiles, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	profiles, err := c.RPCClient.ImplantProfiles(ctx, &commonpb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get implant profiles: %v", err)
	}

	return profiles, nil
}

func (c *SliverClient) SaveImplantProfile(ctx context.Context, profile *clientpb.ImplantProfile) (*clientpb.ImplantProfile, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	savedProfile, err := c.RPCClient.SaveImplantProfile(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to save implant profile: %v", err)
	}

	return savedProfile, nil
}

func (c *SliverClient) DeleteImplantProfile(ctx context.Context, profileID string) (*commonpb.Empty, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	empty, err := c.RPCClient.DeleteImplantProfile(ctx, &clientpb.DeleteReq{
		Name: profileID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete implant profile: %v", err)
	}

	return empty, nil
}

func (c *SliverClient) ImplantBuilds(ctx context.Context) (*clientpb.ImplantBuilds, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	builds, err := c.RPCClient.ImplantBuilds(ctx, &commonpb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get implant builds: %v", err)
	}

	return builds, nil
}

func (c *SliverClient) DeleteImplantBuild(ctx context.Context, buildID string) (*commonpb.Empty, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	empty, err := c.RPCClient.DeleteImplantBuild(ctx, &clientpb.DeleteReq{
		Name: buildID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete implant build: %v", err)
	}

	return empty, nil
}

// TODO: GenerateStage needs to be implemented
// Protobuf definitions/implementation not found in sliver version v1.5.x
// Will need to update sliver version or adapt to available API

func (c *SliverClient) RmBeacon(ctx context.Context, beaconID string) (*commonpb.Empty, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	empty, err := c.RPCClient.RmBeacon(ctx, &clientpb.Beacon{
		ID: beaconID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to remove beacon: %v", err)
	}

	return empty, nil
}

func (c *SliverClient) GetBeaconTasks(ctx context.Context, beaconID string) (*clientpb.BeaconTasks, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	tasks, err := c.RPCClient.GetBeaconTasks(ctx, &clientpb.Beacon{
		ID: beaconID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get beacon tasks: %v", err)
	}

	return tasks, nil
}

// TODO: CancelBeaconTask needs to be implemented
// Protobuf definitions/implementation not found in sliver version v1.5.x
// Will need to update sliver version or adapt to available API

func (c *SliverClient) StartMTLSListener(ctx context.Context, host string, port uint32) (interface{}, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	job, err := c.RPCClient.StartMTLSListener(ctx, &clientpb.MTLSListenerReq{
		Host: host,
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start MTLS listener: %v", err)
	}

	return job, nil
}

func (c *SliverClient) StartHTTPListener(ctx context.Context, domain, host string, port uint32) (interface{}, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	job, err := c.RPCClient.StartHTTPListener(ctx, &clientpb.HTTPListenerReq{
		Domain: domain,
		Host:   host,
		Port:   port,
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start HTTP listener: %v", err)
	}

	return job, nil
}

func (c *SliverClient) StartHTTPSListener(ctx context.Context, domain, host string, port uint32, cert, key []byte) (interface{}, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	job, err := c.RPCClient.StartHTTPSListener(ctx, &clientpb.HTTPListenerReq{
		Domain: domain,
		Host:   host,
		Port:   port,
		Secure: true,
		Cert:   cert,
		Key:    key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start HTTPS listener: %v", err)
	}

	return job, nil
}

func (c *SliverClient) KillJob(ctx context.Context, jobID uint32) (*clientpb.KillJob, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	killJob, err := c.RPCClient.KillJob(ctx, &clientpb.KillJobReq{
		ID: jobID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to kill job: %v", err)
	}

	return killJob, nil
}

func (c *SliverClient) Ls(ctx context.Context, sessionID, path string) (*sliverpb.Ls, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	ls, err := c.RPCClient.Ls(ctx, &sliverpb.LsReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
		},
		Path: path,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}

	return ls, nil
}

func (c *SliverClient) Cd(ctx context.Context, sessionID, path string) (*sliverpb.Pwd, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	pwd, err := c.RPCClient.Cd(ctx, &sliverpb.CdReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
		},
		Path: path,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to change directory: %v", err)
	}

	return pwd, nil
}

func (c *SliverClient) Pwd(ctx context.Context, sessionID string) (*sliverpb.Pwd, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	pwd, err := c.RPCClient.Pwd(ctx, &sliverpb.PwdReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %v", err)
	}

	return pwd, nil
}

func (c *SliverClient) Ps(ctx context.Context, sessionID string) (*sliverpb.Ps, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	ps, err := c.RPCClient.Ps(ctx, &sliverpb.PsReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %v", err)
	}

	return ps, nil
}

func (c *SliverClient) Terminate(ctx context.Context, sessionID string, pid int32, force bool) (*sliverpb.Terminate, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	terminate, err := c.RPCClient.Terminate(ctx, &sliverpb.TerminateReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
		},
		Pid:   pid,
		Force: force,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to terminate process: %v", err)
	}

	return terminate, nil
}

func (c *SliverClient) Execute(ctx context.Context, sessionID, command string) (*sliverpb.Execute, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
	}

	// First, get the session to determine OS type
	session, err := c.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session info: %v", err)
	}

	// Create the request based on the target OS
	var execute *sliverpb.Execute
	if strings.ToLower(session.OS) == "windows" {
		// For Windows, execute the command directly through cmd.exe
		// Use /D to disable AutoRun and /V:OFF to disable delayed variable expansion
		// Use /C to terminate after command completes
		// Use the /u flag for Unicode output in cmd.exe
		execute, err = c.RPCClient.Execute(ctx, &sliverpb.ExecuteReq{
			Request: &commonpb.Request{
				SessionID: sessionID,
			},
			Path:   "cmd.exe",
			Args:   []string{"/D", "/u", "/V:OFF", "/C", command},
			Output: true,
		})

		// Try PowerShell if cmd fails
		if err != nil {
			// Modify the PowerShell command to ensure proper output handling
			// Use -NoProfile to speed up startup and specific output settings
			// Set encoding to UTF8 to avoid encoding issues with multiline output
			psCommand := fmt.Sprintf("$OutputEncoding = [System.Text.Encoding]::UTF8; %s; exit $LASTEXITCODE", command)
			execute, err = c.RPCClient.Execute(ctx, &sliverpb.ExecuteReq{
				Request: &commonpb.Request{
					SessionID: sessionID,
				},
				Path:   "powershell.exe",
				Args:   []string{"-NoProfile", "-NonInteractive", "-OutputFormat", "Text", "-Command", psCommand},
				Output: true,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to execute command with both cmd.exe and powershell.exe: %v", err)
			}
		}
	} else {
		// Unix-like systems (Linux, macOS)
		// Use absolute paths for the shells to avoid any path resolution issues
		const bash = "/bin/bash"
		const sh = "/bin/sh"

		// Try to execute with bash first (most common shell with most features)
		execute, err = c.RPCClient.Execute(ctx, &sliverpb.ExecuteReq{
			Request: &commonpb.Request{
				SessionID: sessionID,
			},
			Path:   bash,
			Args:   []string{"-c", command},
			Output: true,
		})

		// If bash fails, try sh as a fallback
		if err != nil {
			execute, err = c.RPCClient.Execute(ctx, &sliverpb.ExecuteReq{
				Request: &commonpb.Request{
					SessionID: sessionID,
				},
				Path:   sh,
				Args:   []string{"-c", command},
				Output: true,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to execute command with both bash and sh: %v", err)
			}
		}
	}

	return execute, nil
}

func (c *SliverClient) Download(ctx context.Context, sessionID, remotePath string) (*sliverpb.Download, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
	}

	download, err := c.RPCClient.Download(ctx, &sliverpb.DownloadReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
		},
		Path: remotePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %v", err)
	}

	return download, nil
}

func (c *SliverClient) Upload(ctx context.Context, sessionID, remotePath string, data []byte) (*sliverpb.Upload, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
	}

	upload, err := c.RPCClient.Upload(ctx, &sliverpb.UploadReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
		},
		Path: remotePath,
		Data: data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %v", err)
	}

	return upload, nil
}

func (c *SliverClient) Rm(ctx context.Context, sessionID, path string, recursive, force bool) (*sliverpb.Rm, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	rm, err := c.RPCClient.Rm(ctx, &sliverpb.RmReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
		},
		Path:      path,
		Recursive: recursive,
		Force:     force,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to remove file: %v", err)
	}

	return rm, nil
}

func (c *SliverClient) Mkdir(ctx context.Context, sessionID, path string) (*sliverpb.Mkdir, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	mkdir, err := c.RPCClient.Mkdir(ctx, &sliverpb.MkdirReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
		},
		Path: path,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	return mkdir, nil
}

func (c *SliverClient) Kill(ctx context.Context, sessionID string, force bool) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	_, err := c.RPCClient.Kill(ctx, &sliverpb.KillReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
		},
		Force: force,
	})
	if err != nil {
		return fmt.Errorf("failed to kill session: %v", err)
	}

	return nil
}

func (c *SliverClient) GetSession(ctx context.Context, sessionID string) (*clientpb.Session, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	sessions, err := c.RPCClient.GetSessions(ctx, &commonpb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %v", err)
	}

	for _, session := range sessions.Sessions {
		if session.ID == sessionID {
			return session, nil
		}
	}

	return nil, fmt.Errorf("session with ID %s not found", sessionID)
}

func (c *SliverClient) RenameSession(ctx context.Context, sessionID, newName string) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	_, err := c.RPCClient.Rename(ctx, &clientpb.RenameReq{
		SessionID: sessionID,
		Name:      newName,
	})
	if err != nil {
		return fmt.Errorf("failed to rename session: %v", err)
	}

	return nil
}

func (c *SliverClient) Mv(ctx context.Context, sessionID, srcPath, dstPath string) (*sliverpb.Mv, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	mv, err := c.RPCClient.Mv(ctx, &sliverpb.MvReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
		},
		Src: srcPath,
		Dst: dstPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to move file: %v", err)
	}

	return mv, nil
}

// TODO: Cp needs to be implemented
// Protobuf definitions/implementation not found in sliver version v1.5.x
// Will need to update sliver version or adapt to available API
