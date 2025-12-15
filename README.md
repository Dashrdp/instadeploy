# InstaDeploy

A lightweight PaaS (Platform as a Service) system for deploying Docker applications. InstaDeploy consists of:

- **Control Plane**: Python-based FastAPI server that manages agents and deployments via REST API
- **Agent**: Go-based agent that runs on target VPS instances and executes Docker Compose deployments

The Control Plane and Agents communicate via WebSocket for real-time command execution and status updates.

## 🏗️ Architecture

```
┌─────────────────┐
│  REST API       │
│  (User/CI/CD)   │
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│  Control Plane      │
│  (Python/FastAPI)   │
│  - WebSocket Hub    │
│  - Job Queue        │
│  - SQLite DB        │
└──────────┬──────────┘
           │
           │ WebSocket
           │
    ┌──────┴──────┬──────────┬──────────┐
    ▼             ▼          ▼          ▼
┌────────┐   ┌────────┐  ┌────────┐  ┌────────┐
│ Agent  │   │ Agent  │  │ Agent  │  │ Agent  │
│ (Go)   │   │ (Go)   │  │ (Go)   │  │ (Go)   │
│ VPS 1  │   │ VPS 2  │  │ VPS 3  │  │ VPS N  │
└────────┘   └────────┘  └────────┘  └────────┘
```

## 🚀 Quick Start

### Prerequisites

**Control Plane:**
- Python 3.11+
- SQLite (included)

**Agent:**
- Go 1.22+ (for building)
- Docker and Docker Compose
- Linux, macOS, or Windows

### Setup Control Plane

```bash
# Clone the repository
git clone https://github.com/yourorg/instadeploy
cd instadeploy/control-plane

# Install dependencies
pip install -r requirements.txt

# Set environment variables
export AGENT_SECRET_TOKEN="your-secret-token"

# Run the control plane
python run.py
```

The Control Plane will start on `http://localhost:8000`. Visit `http://localhost:8000/docs` for API documentation.

### Build Agent

```bash
# Navigate to project root
cd ..

# Download dependencies
go mod download

# Build the agent
make build-agent
```

### Run Agent

```bash
# Set environment variables
export SERVER_URL="ws://localhost:8000/ws"
export AGENT_SECRET_TOKEN="your-secret-token"
export AGENT_VERSION="1.0.0"
export AGENT_ARCHITECTURE="amd64"

# Start the agent
./bin/agent
```

The agent will connect to the Control Plane and be ready to receive deployment commands.

### Test the System

**Terminal 1 - Control Plane:**
```bash
cd control-plane
python run.py
```

**Terminal 2 - Agent:**
```bash
export SERVER_URL="ws://localhost:8000/ws"
export AGENT_SECRET_TOKEN="your-secret-token"
export AGENT_VERSION="1.0.0"
export AGENT_ARCHITECTURE="amd64"
./bin/agent
```

**Terminal 3 - Deploy:**
```bash
# Check connected agents
curl http://localhost:8000/agents

# Deploy a test project
curl -X POST http://localhost:8000/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "project_name": "test-app",
    "compose_file_base64": "'"$(base64 -i examples/docker-compose-test.yml)"'"
  }'
```

## 📋 Features

### Control Plane
- ✅ **REST API**: Simple HTTP API for deployments, status checks, and management
- ✅ **WebSocket Hub**: Manages persistent connections to multiple agents
- ✅ **Multi-agent support**: Deploy to any connected agent in your fleet
- ✅ **Job tracking**: Track deployment jobs with unique IDs and status updates
- ✅ **Database persistence**: SQLite database for agents, deployments, and jobs
- ✅ **Interactive docs**: Auto-generated Swagger/OpenAPI documentation
- ✅ **Health monitoring**: Track agent online/offline status and last seen timestamps

### Agent
- ✅ **Outbound-only communication**: Works behind NAT/firewalls, no inbound ports required
- ✅ **Docker Compose support**: Deploy applications using docker-compose.yml files
- ✅ **Secure authentication**: Token-based authentication with Bearer tokens
- ✅ **Auto-reconnection**: Exponential backoff reconnection logic
- ✅ **Path validation**: Security measures to prevent directory traversal attacks
- ✅ **Command types**: Deploy, Stop, Status, Health Check
- ✅ **Real-time logs**: Streams deployment logs back to control plane

## 📖 Documentation

### Control Plane
- **[control-plane/README.md](control-plane/README.md)**: Control Plane setup and usage
- **[control-plane/API.md](control-plane/API.md)**: Complete API reference
- **[control-plane/TESTING.md](control-plane/TESTING.md)**: Testing guide for Control Plane

### Agent
- **[ARCHITECTURE.md](ARCHITECTURE.md)**: Detailed architecture and design decisions
- **[TESTING.md](TESTING.md)**: Complete testing guide with examples
- **[DEPLOYMENT.md](DEPLOYMENT.md)**: Production deployment guide

## 🔧 How It Works

1. **Agent Connection**: Agents establish outbound WebSocket connections to the Control Plane
2. **Job Creation**: User/CI sends deployment request to Control Plane REST API
3. **Command Routing**: Control Plane routes command to target agent via WebSocket
4. **Execution**: Agent executes Docker Compose command on local Docker daemon
5. **Status Updates**: Agent sends real-time status updates back to Control Plane
6. **Persistence**: All jobs, deployments, and agent status persisted in database

### Key Components

**Control Plane:**
1. **FastAPI Application**: HTTP server with REST endpoints
2. **WebSocket Manager**: Manages active agent connections
3. **Database Layer**: SQLModel for agent, deployment, and job persistence
4. **Protocol Schemas**: Pydantic models for type-safe communication

**Agent:**
1. **Connection Manager**: Handles WebSocket connection, reconnection, and keepalive
2. **Command Handler**: Processes incoming commands and routes to appropriate handlers
3. **Deployment Manager**: Manages Docker Compose deployments with security validation
4. **Configuration**: Loads settings from environment variables

## 🔧 Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `SERVER_URL` | Yes | WebSocket URL (e.g., `ws://localhost:8080/ws`) |
| `AGENT_TOKEN` | Yes | Authentication token |
| `AGENT_SECRET_TOKEN` | No | Secret token (defaults to AGENT_TOKEN) |

### Example Configuration

```bash
export SERVER_URL="ws://control.example.com:8080/ws"
export AGENT_TOKEN="agent-token-123"
export AGENT_SECRET_TOKEN="secret-token-456"
```

## 🔐 Security

- **Path Validation**: Project names validated against `^[a-zA-Z0-9_-]+$`
- **Directory Isolation**: All projects isolated to `./projects/` (or `/opt/platform/projects/`)
- **No Inbound Ports**: Agent only makes outbound connections
- **Token Authentication**: Bearer token on every connection
- **Docker Socket**: Requires appropriate permissions (root or docker group)

## 📦 Building

```bash
# Standard build
make build-agent

# Build for Linux (cross-compile)
make build-linux

# Build obfuscated binary (requires garble)
make install-garble
make build-obfuscated

# Clean build artifacts
make clean
```

## 🧪 Testing

```bash
# Run tests
make test

# Format and lint
make lint

# Run test server and agent
# Terminal 1:
cd test-server && go run server.go

# Terminal 2:
export SERVER_URL="ws://localhost:8080/ws"
export AGENT_TOKEN="test-token"
./bin/agent
```

See [TESTING.md](TESTING.md) for comprehensive testing scenarios.

## 📡 Communication Protocol

### Handshake (Agent → Server)

Headers sent during WebSocket upgrade:
```
Authorization: Bearer <AGENT_SECRET_TOKEN>
X-Agent-Version: 1.0.0
X-Agent-Architecture: linux-amd64
X-Agent-Hostname: hostname
```

### Command (Server → Agent)

```json
{
  "id": "job-123",
  "type": "DEPLOY_COMPOSE",
  "payload": {
    "project_name": "my-app",
    "compose_file_base64": "dmVyc2lvbjogJzMuOCc..."
  }
}
```

### Response (Agent → Server)

```json
{
  "job_id": "job-123",
  "status": "COMPLETED",
  "logs": "Deployment successful",
  "error": ""
}
```

### Command Types

| Command | Description | Payload |
|---------|-------------|---------|
| `DEPLOY_COMPOSE` | Deploy docker-compose.yml | `{project_name, compose_file_base64}` |
| `STOP_COMPOSE` | Stop deployment | `{project_name}` |
| `STATUS` | Get deployment status | `{project_name}` |
| `HEALTH_CHECK` | Agent health check | `{}` |

## 🚢 Deployment

### Systemd Service

```bash
# Copy binary
sudo cp bin/agent /usr/local/bin/instadeploy-agent

# Create service file
sudo nano /etc/systemd/system/instadeploy-agent.service
```

```ini
[Unit]
Description=InstaDeploy Agent
After=docker.service

[Service]
Type=simple
Environment="SERVER_URL=wss://control.example.com/ws"
Environment="AGENT_SECRET_TOKEN=your-token"
ExecStart=/usr/local/bin/instadeploy-agent
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
# Start service
sudo systemctl daemon-reload
sudo systemctl enable instadeploy-agent
sudo systemctl start instadeploy-agent
```

See [DEPLOYMENT.md](DEPLOYMENT.md) for production deployment guide.

## 📁 Project Structure

```
instadeploy/
├── agent/               # Agent source code
│   ├── main.go         # Entry point
│   ├── config.go       # Configuration
│   ├── connection.go   # WebSocket connection
│   ├── handler.go      # Command handler
│   └── deploy.go       # Deployment logic
├── shared/             # Shared types/protocol
│   └── types.go        # Protocol definitions
├── test-server/        # Test WebSocket server
│   ├── server.go       # Test server
│   └── README.md       # Test server docs
├── examples/           # Example files
│   ├── docker-compose-test.yml
│   └── test-deployment.sh
├── go.mod
├── Makefile
├── README.md
├── ARCHITECTURE.md     # Architecture details
├── TESTING.md         # Testing guide
└── DEPLOYMENT.md      # Deployment guide
```

## 🛠️ Development

### Make Commands

```bash
make help              # Show all available commands
make build-agent       # Build agent binary
make build-linux       # Build for Linux amd64
make run              # Build and run agent
make test             # Run tests
make clean            # Clean build artifacts
make lint             # Format and vet code
```

## 🐛 Troubleshooting

### Connection Issues

```bash
# Check SERVER_URL
echo $SERVER_URL

# Test WebSocket endpoint
websocat ws://localhost:8080/ws

# Check agent logs
./bin/agent 2>&1 | tee agent.log
```

### Docker Issues

```bash
# Check Docker is running
docker info

# Check Docker Compose
docker compose version

# Check permissions
docker ps
```

### Deployment Issues

```bash
# Check project directory
ls -la ./projects/

# Check compose file
cat ./projects/my-app/docker-compose.yml

# Check containers
docker ps -a
```

## 📄 License

MIT License - see LICENSE file for details

## 🤝 Contributing

Contributions welcome! Please read CONTRIBUTING.md first.

## 📞 Support

- Issues: https://github.com/yourorg/instadeploy/issues
- Documentation: https://docs.instadeploy.com
- Discord: https://discord.gg/instadeploy

## 🗺️ Roadmap

- [x] Basic agent implementation
- [x] Docker Compose support
- [x] Reconnection logic
- [x] Security validation
- [ ] Control plane implementation
- [ ] Self-update capability
- [ ] Traefik integration
- [ ] Multi-architecture support (ARM)
- [ ] Real-time log streaming
- [ ] Metrics collection

## 🙏 Acknowledgments

Built with:
- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [Docker](https://www.docker.com/)
- Go 1.22+

