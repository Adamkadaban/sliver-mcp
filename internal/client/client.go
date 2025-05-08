package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	data, err := ioutil.ReadFile(configPath)
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

	generate, err := c.RPCClient.Generate(ctx, &clientpb.GenerateReq{
		Config: config,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate implant: %v", err)
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

func (c *SliverClient) Execute(ctx context.Context, sessionID, command string) (*sliverpb.Execute, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
	}

	execute, err := c.RPCClient.Execute(ctx, &sliverpb.ExecuteReq{
		Request: &commonpb.Request{
			SessionID: sessionID,
		},
		Path: command,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %v", err)
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
