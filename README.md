# Sliver MCP

A Model-Code-Plugin (MCP) for interacting with the [Sliver](https://github.com/bishopfox/sliver) C2 framework. This MCP allows Large Language Models to interact with Sliver's functionality through well-defined tools.

## Features

- Session management (list, interact, kill)
- Implant generation and management
- C2 profile management
- File operations
- Shell execution
- Process manipulation
- Network operations

## Installation

```bash
# Clone the repository
git clone https://github.com/adamkadaban/sliver-mcp.git
cd sliver-mcp

# Build the MCP
go build -o sliver-mcp ./cmd/sliver-mcp
```

## Usage

### Configuration

Before using the MCP, you need to configure the Sliver client by providing a path to your Sliver configuration.

```bash
./sliver-mcp --config /path/to/sliver/config.json
```

By default, the MCP uses the standard input/output for communication. You can also use the SSE transport:

```bash
./sliver-mcp --transport sse
```

### Tools

The MCP exposes the following tools:

#### Session Management
- `listSessions`: List all active Sliver sessions
- `interactSession`: Interact with a specific session
- `killSession`: Terminate a specific session
- `background`: Background the current session

#### Implant Generation
- `generateImplant`: Generate a new Sliver implant
- `listImplants`: List generated implants
- `regenerateImplant`: Regenerate an existing implant

#### C2 Profile Management
- `listProfiles`: List available C2 profiles
- `createProfile`: Create a new C2 profile
- `deleteProfile`: Delete an existing profile

#### File Operations
- `download`: Download a file from a session
- `upload`: Upload a file to a session
- `ls`: List files on the remote system
- `rm`: Remove files on the remote system
- `mkdir`: Create directories on the remote system

#### Shell Operations
- `executeCommand`: Execute a shell command
- `getShell`: Get an interactive shell

#### Process Management
- `ps`: List processes on the remote system
- `kill`: Kill a process on the remote system
- `execute`: Execute a binary on the remote system

#### Network Operations
- `netstat`: List network connections
- `portscan`: Scan for open ports
- `tcpdump`: Capture network traffic

## Security Considerations

- Store Sliver credentials securely
- Implement proper access controls
- Sanitize all inputs to prevent command injection
- Log all operations for audit purposes
- Consider implementing rate limiting for sensitive operations
- Ensure secure handling of implant generation artifacts