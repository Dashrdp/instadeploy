# InstaDeploy Agent Architecture

## Overview

The InstaDeploy Agent is a lightweight Go application that runs on target VPS instances. It enables remote Docker Compose deployments through a secure WebSocket connection to a control plane, working seamlessly behind NAT/firewalls without requiring inbound port access.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Control Plane (Future)                   │
│                                                              │
│  - WebSocket Server                                          │
│  - User Management                                           │
│  - Deployment Orchestration                                  │
└───────────────────────┬──────────────────────────────────────┘
                        │
                        │ WebSocket (WSS)
                        │ Outbound Connection
                        │
┌───────────────────────▼──────────────────────────────────────┐
│                    Agent (Target VPS)                         │
│                                                              │
│  ┌────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Connection   │  │ Command Handler │  │  Deployment  │ │
│  │    Manager     │──│                 │──│   Manager    │ │
│  │                │  │  - Deploy       │  │              │ │
│  │  - Reconnect   │  │  - Stop         │  │  - Validate  │ │
│  │  - Keepalive   │  │  - Status       │  │  - Write     │ │
│  │  - Auth        │  │  - Health       │  │  - Execute   │ │
│  └────────────────┘  └─────────────────┘  └──────┬───────┘ │
│                                                    │         │
│                                                    │         │
└────────────────────────────────────────────────────┼─────────┘
                                                     │
                                                     │
                                          ┌──────────▼────────┐
                                          │  Docker Engine    │
                                          │                   │
                                          │  - Compose        │
                                          │  - Containers     │
                                          │  - Networks       │
                                          └───────────────────┘
```

## Component Details

### 1. Connection Manager (`connection.go`)

**Responsibilities:**
- Establish and maintain WebSocket connection to control plane
- Handle authentication with Bearer tokens
- Implement reconnection logic with exponential backoff
- Manage ping/pong keepalive mechanism
- Route incoming messages to command handler

**Key Features:**
- **Exponential Backoff**: Starts at 1s, doubles up to 2 minutes
- **Keepalive**: Sends ping every 30s, expects pong within 60s
- **Auto-reconnect**: Infinite retry loop for resilience
- **Graceful shutdown**: Handles SIGTERM/SIGINT

**Configuration:**
```go
initialBackoff     = 1 * time.Second
maxBackoff        = 2 * time.Minute
pingInterval      = 30 * time.Second
pongTimeout       = 60 * time.Second
```

### 2. Command Handler (`handler.go`)

**Responsibilities:**
- Parse incoming JSON commands
- Route commands to appropriate handlers
- Execute deployment operations
- Generate structured responses

**Supported Commands:**

| Command Type | Description | Payload |
|-------------|-------------|---------|
| `DEPLOY_COMPOSE` | Deploy docker-compose.yml | `{project_name, compose_file_base64}` |
| `STOP_COMPOSE` | Stop a deployment | `{project_name}` |
| `STATUS` | Get deployment status | `{project_name}` |
| `HEALTH_CHECK` | Agent health check | `{}` |

**Response Format:**
```json
{
  "job_id": "unique-job-id",
  "status": "COMPLETED|FAILED|IN_PROGRESS",
  "logs": "execution logs",
  "error": "error message if failed",
  "data": {}
}
```

### 3. Deployment Manager (`deploy.go`)

**Responsibilities:**
- Validate project names for security
- Manage project directories
- Write docker-compose.yml files
- Execute Docker Compose commands
- Prevent directory traversal attacks

**Security Features:**
- **Path Validation**: Only alphanumeric, hyphens, underscores
- **Directory Containment**: All projects in `./projects/` (or `/opt/platform/projects/`)
- **Base64 Decoding**: Safely decode compose files
- **Permission Control**: Creates directories with 0755, files with 0644

**File Structure:**
```
./projects/
├── project-1/
│   └── docker-compose.yml
├── project-2/
│   └── docker-compose.yml
└── project-n/
    └── docker-compose.yml
```

### 4. Configuration Manager (`config.go`)

**Responsibilities:**
- Load configuration from environment variables
- Provide sensible defaults
- Validate required settings

**Required Environment Variables:**
- `SERVER_URL`: WebSocket endpoint (e.g., `ws://example.com:8080/ws`)
- `AGENT_TOKEN`: Authentication token
- `AGENT_SECRET_TOKEN`: Optional secret token (defaults to AGENT_TOKEN)

**Auto-detected Values:**
- `AgentVersion`: Hardcoded to "1.0.0"
- `AgentArchitecture`: Detected from runtime (e.g., "linux-amd64")

## Communication Protocol

### Handshake (Agent → Server)

Headers sent during WebSocket upgrade:
```
Authorization: Bearer <AGENT_SECRET_TOKEN>
X-Agent-Version: 1.0.0
X-Agent-Architecture: linux-amd64
X-Agent-Hostname: server-hostname
```

### Command Flow

1. **Server sends command** (JSON over WebSocket)
2. **Agent acknowledges** with IN_PROGRESS status
3. **Agent executes** the command
4. **Agent responds** with COMPLETED or FAILED status

### Example: Deploy Command

**Server → Agent:**
```json
{
  "id": "deploy-123",
  "type": "DEPLOY_COMPOSE",
  "payload": {
    "project_name": "grafana-stack",
    "compose_file_base64": "dmVyc2lvbjogJzMuOCcK..."
  }
}
```

**Agent → Server (Acknowledgment):**
```json
{
  "job_id": "deploy-123",
  "status": "IN_PROGRESS",
  "logs": "Command received, processing..."
}
```

**Agent → Server (Completion):**
```json
{
  "job_id": "deploy-123",
  "status": "COMPLETED",
  "logs": "Project 'grafana-stack' deployed successfully.\n\nOutput:\n[+] Running 3/3..."
}
```

## Security Considerations

### 1. Path Traversal Prevention
- Project names are validated against regex: `^[a-zA-Z0-9_-]+$`
- Absolute paths are resolved and checked against base directory
- Attempts to escape project directory are rejected

### 2. Authentication
- Bearer token authentication on every connection
- Token validation happens at control plane (future)
- Tokens stored in environment variables, never hardcoded

### 3. Network Security
- **No inbound ports**: Agent initiates all connections
- **WSS recommended**: Use TLS for production (`wss://`)
- **Firewall friendly**: Works behind NAT/firewalls

### 4. Docker Socket Access
- Agent requires access to `/var/run/docker.sock`
- Runs with appropriate permissions (typically root or docker group)
- Only executes `docker compose` commands, not arbitrary Docker operations

## Deployment

### Development Setup
```bash
# Set environment variables
export SERVER_URL="ws://localhost:8080/ws"
export AGENT_TOKEN="test-token"

# Build and run
make build-agent
./bin/agent
```

### Production Deployment (Systemd)
```ini
[Unit]
Description=InstaDeploy Agent
After=docker.service
Requires=docker.service

[Service]
Type=simple
Environment="SERVER_URL=wss://control.example.com/ws"
Environment="AGENT_SECRET_TOKEN=secure-token-here"
ExecStart=/usr/local/bin/agent
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Cloud-Init Script (Future)
The control plane will generate cloud-init scripts to automatically:
1. Install Docker
2. Download agent binary
3. Create systemd service with unique token
4. Start and enable the service

## Future Enhancements

### Planned Features
1. **Self-Update**: Download and install new agent versions remotely
2. **Log Streaming**: Real-time log streaming from containers
3. **Metrics Collection**: CPU, memory, disk usage reporting
4. **Traefik Integration**: Automatic proxy configuration
5. **Multi-architecture**: ARM support for Raspberry Pi, etc.
6. **Encryption**: End-to-end encryption for compose files

### Control Plane Integration
Once the control plane is implemented:
- Agent registration and authentication
- Multi-agent management dashboard
- Deployment scheduling and rollback
- Monitoring and alerting
- User access control

## Development

### Project Structure
```
instadeploy/
├── agent/
│   ├── main.go           # Entry point
│   ├── config.go         # Configuration management
│   ├── connection.go     # WebSocket connection
│   ├── handler.go        # Command processing
│   └── deploy.go         # Deployment logic
├── shared/
│   └── types.go          # Shared protocol types
├── examples/
│   ├── docker-compose-test.yml
│   └── test-deployment.sh
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Building
```bash
# Standard build
make build-agent

# Linux build (for deployment)
make build-linux

# Obfuscated build (requires garble)
make install-garble
make build-obfuscated
```

### Testing
```bash
# Run tests
make test

# Format and lint
make lint
```

## Troubleshooting

### Connection Issues
- **Check SERVER_URL**: Ensure it's reachable and correct protocol (ws/wss)
- **Verify tokens**: AGENT_TOKEN must match control plane expectations
- **Firewall**: Ensure outbound WebSocket connections are allowed

### Deployment Failures
- **Docker installed**: Verify Docker and Docker Compose are available
- **Permissions**: Agent needs access to Docker socket
- **Compose syntax**: Validate docker-compose.yml before sending

### Resource Issues
- **Disk space**: Check `/opt/platform/projects/` has sufficient space
- **Docker resources**: Ensure Docker daemon has adequate resources
- **Connection limits**: Monitor open file descriptors and sockets

