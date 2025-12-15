# Control Plane Implementation Summary

This document summarizes the implementation of the InstaDeploy Control Plane.

## Overview

The Control Plane is a Python-based FastAPI application that manages a fleet of agents via WebSocket connections and exposes a REST API for deployment management.

## Technology Stack

- **FastAPI**: Modern, fast web framework for building APIs
- **SQLModel**: SQL database ORM with Pydantic integration
- **SQLite**: Lightweight database for persistence
- **WebSockets**: Real-time bidirectional communication with agents
- **Pydantic**: Data validation and settings management
- **Uvicorn**: ASGI server for running the application

## Project Structure

```
control-plane/
├── __init__.py           # Package initialization
├── main.py               # FastAPI application and endpoints
├── config.py             # Configuration management
├── database.py           # Database setup and session management
├── models.py             # SQLModel database models
├── schemas.py            # Pydantic schemas for API/protocol
├── manager.py            # WebSocket connection manager
├── run.py                # Application entry point
├── requirements.txt      # Python dependencies
├── Dockerfile            # Docker build configuration
├── docker-compose.yml    # Docker Compose setup
├── test_api.py           # API test script
├── README.md             # Control Plane documentation
├── API.md                # API reference documentation
└── TESTING.md            # Testing guide
```

## Core Components

### 1. Database Models (`models.py`)

Three main models:

- **Agent**: Tracks connected agents
  - id, hostname, ip_address, status, last_seen, version, architecture
  
- **Deployment**: Tracks project deployments
  - id, project_name, agent_id, status, compose_file_hash, logs, error
  
- **Job**: Tracks individual command executions
  - id, deployment_id, agent_id, command_type, status, logs, error

### 2. Protocol Schemas (`schemas.py`)

Matches the Go agent protocol:

- **Command Types**: DEPLOY_COMPOSE, STOP_COMPOSE, STATUS, HEALTH_CHECK
- **Response Status**: QUEUED, IN_PROGRESS, COMPLETED, FAILED
- **Payloads**: DeployPayload, StopPayload, StatusPayload
- **Messages**: Command, Response

### 3. Connection Manager (`manager.py`)

Manages WebSocket connections:

- **connect()**: Register new agent connection
- **disconnect()**: Remove agent connection
- **send_command()**: Send command to specific agent
- **receive_response()**: Receive response from agent
- **is_connected()**: Check agent connection status
- **get_connected_agents()**: List connected agent IDs

### 4. FastAPI Application (`main.py`)

#### WebSocket Endpoint

**WS /ws**: Agent connection endpoint

- Authenticates agents via Bearer token
- Extracts agent metadata from headers
- Creates/updates agent in database
- Listens for responses and updates job/deployment status
- Handles disconnections and marks agents offline

#### REST API Endpoints

**GET /**
- Root endpoint with service info

**GET /health**
- Health check with connected agent count

**GET /agents**
- List all registered agents

**POST /deploy**
- Deploy a project to an agent
- Creates deployment and job records
- Sends DEPLOY_COMPOSE command to agent

**POST /projects/{name}/stop**
- Stop a running project
- Sends STOP_COMPOSE command to agent

**GET /projects/{name}/status**
- Get project status
- Sends STATUS command to agent

## Authentication

Agents authenticate using Bearer tokens in the WebSocket handshake:

```
Authorization: Bearer <AGENT_SECRET_TOKEN>
```

The token is validated against the `AGENT_SECRET_TOKEN` environment variable.

## Configuration

All configuration via environment variables:

- `AGENT_SECRET_TOKEN`: Secret token for agent authentication
- `DATABASE_URL`: SQLite database URL (default: sqlite:///./control_plane.db)
- `HOST`: Server host (default: 0.0.0.0)
- `PORT`: Server port (default: 8000)

## Database Schema

### agents
```sql
CREATE TABLE agent (
    id INTEGER PRIMARY KEY,
    hostname VARCHAR UNIQUE NOT NULL,
    ip_address VARCHAR,
    status VARCHAR NOT NULL,
    last_seen DATETIME NOT NULL,
    version VARCHAR,
    architecture VARCHAR,
    created_at DATETIME NOT NULL
);
```

### deployments
```sql
CREATE TABLE deployment (
    id INTEGER PRIMARY KEY,
    project_name VARCHAR NOT NULL,
    agent_id INTEGER NOT NULL,
    status VARCHAR NOT NULL,
    compose_file_hash VARCHAR,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    last_logs TEXT,
    last_error TEXT,
    FOREIGN KEY (agent_id) REFERENCES agent(id)
);
```

### jobs
```sql
CREATE TABLE job (
    id VARCHAR PRIMARY KEY,
    deployment_id INTEGER,
    agent_id INTEGER NOT NULL,
    command_type VARCHAR NOT NULL,
    status VARCHAR NOT NULL,
    logs TEXT,
    error TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (deployment_id) REFERENCES deployment(id),
    FOREIGN KEY (agent_id) REFERENCES agent(id)
);
```

## Communication Flow

### Agent Connection Flow

1. Agent connects to `/ws` with authentication headers
2. Control Plane validates Bearer token
3. Control Plane extracts agent info from headers
4. Control Plane creates/updates agent record in database
5. Connection registered in ConnectionManager
6. Control Plane listens for responses

### Deployment Flow

1. User/CI sends POST request to `/deploy`
2. Control Plane validates request
3. Control Plane finds target agent (specified or first available)
4. Control Plane creates Deployment and Job records
5. Control Plane creates Command with base64-encoded compose file
6. Control Plane sends Command via WebSocket to agent
7. Agent receives command and starts deployment
8. Agent sends back Response messages (IN_PROGRESS, COMPLETED/FAILED)
9. Control Plane updates Job and Deployment status in database

### Status Update Flow

1. Agent sends Response via WebSocket
2. Control Plane receives and parses Response
3. Control Plane looks up Job by job_id
4. Control Plane updates Job status, logs, error
5. If Job has associated Deployment, updates Deployment status too
6. Changes committed to database

## Error Handling

- **Agent not found**: HTTP 404
- **Agent not connected**: HTTP 503
- **Command send failure**: HTTP 503, job marked as failed
- **WebSocket errors**: Logged, connection closed, agent marked offline
- **Database errors**: Logged, transaction rolled back

## Security Features

- **Token authentication**: All agent connections authenticated
- **Path validation**: Agent validates project names (in Go agent)
- **Input validation**: Pydantic schemas validate all inputs
- **No direct shell access**: All commands via Docker Compose

## Monitoring & Observability

- **Health endpoint**: `/health` returns status and connected agent count
- **Agent status**: Online/offline tracking with last_seen timestamps
- **Job tracking**: All commands tracked with unique IDs
- **Logging**: Structured logging for all operations
- **Database persistence**: Full audit trail of deployments and jobs

## Testing

See [TESTING.md](TESTING.md) for complete testing guide.

Quick test:
```bash
python test_api.py
```

## Deployment

### Development
```bash
python run.py
```

### Docker
```bash
docker build -t control-plane .
docker run -p 8000:8000 -e AGENT_SECRET_TOKEN=secret control-plane
```

### Docker Compose
```bash
docker-compose up -d
```

## API Documentation

When running, visit:
- Swagger UI: http://localhost:8000/docs
- ReDoc: http://localhost:8000/redoc

## Future Enhancements

Possible improvements:

1. **PostgreSQL support**: For production scale
2. **TLS/WSS**: Encrypted WebSocket connections
3. **Multi-tenancy**: Support multiple users/organizations
4. **RBAC**: Role-based access control
5. **Webhooks**: Notify external systems on events
6. **Metrics**: Prometheus metrics export
7. **UI Dashboard**: Web UI for management
8. **Agent health checks**: Periodic health check commands
9. **Deployment rollback**: Rollback to previous versions
10. **Secrets management**: Secure secrets injection

## Compatibility

The Control Plane is fully compatible with the existing Go agent implementation, using the same protocol defined in `shared/types.go`.

## Performance

- **Async I/O**: FastAPI/Uvicorn for high concurrency
- **Connection pooling**: SQLModel handles database connections
- **Non-blocking**: WebSocket operations don't block API requests
- **Scalability**: Can handle hundreds of concurrent agent connections

## License

Same as parent project (see [LICENSE](../LICENSE)).


