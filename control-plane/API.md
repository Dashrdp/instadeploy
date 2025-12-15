# Control Plane API Documentation

The Control Plane exposes a REST API for managing agents and deployments, and a WebSocket endpoint for agent connections.

## Base URL

```
http://localhost:8000
```

## Authentication

### Agent Authentication (WebSocket)
Agents must authenticate using a Bearer token in the WebSocket handshake:

```
Authorization: Bearer <AGENT_SECRET_TOKEN>
```

Additional headers required:
- `X-Agent-Version`: Agent version string
- `X-Agent-Architecture`: Agent architecture (e.g., "amd64", "arm64")
- `X-Agent-Hostname`: Agent hostname

## WebSocket Protocol

### Endpoint
```
WS /ws
```

### Message Format

**Command (Control Plane → Agent):**
```json
{
  "id": "uuid",
  "type": "DEPLOY_COMPOSE|STOP_COMPOSE|STATUS|HEALTH_CHECK",
  "payload": {
    // Command-specific data
  }
}
```

**Response (Agent → Control Plane):**
```json
{
  "job_id": "uuid",
  "status": "COMPLETED|FAILED|IN_PROGRESS|QUEUED",
  "logs": "command output",
  "error": "error message if failed",
  "data": null
}
```

## REST API Endpoints

### Health Check

**GET /health**

Check Control Plane health and connected agents.

**Response:**
```json
{
  "status": "healthy",
  "connected_agents": 2
}
```

### List Agents

**GET /agents**

Get all registered agents.

**Response:**
```json
{
  "agents": [
    {
      "id": 1,
      "hostname": "agent-1",
      "ip_address": "192.168.1.100",
      "status": "online",
      "last_seen": "2024-01-01T12:00:00",
      "version": "1.0.0",
      "architecture": "amd64"
    }
  ],
  "total": 1
}
```

### Deploy Project

**POST /deploy**

Deploy a Docker Compose project to an agent.

**Request:**
```json
{
  "project_name": "my-app",
  "compose_file_base64": "base64-encoded-docker-compose.yml",
  "agent_id": 1  // optional, uses first available if not specified
}
```

**Response:**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "deployment_id": 1,
  "status": "queued",
  "message": "Deployment queued on agent agent-1"
}
```

**Status Codes:**
- `200`: Success
- `404`: Agent not found
- `503`: No agents available or agent not connected

### Stop Project

**POST /projects/{project_name}/stop**

Stop a running project.

**Path Parameters:**
- `project_name`: Name of the project to stop

**Request:**
```json
{
  "agent_id": 1  // optional, uses deployment's agent if not specified
}
```

**Response:**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440001",
  "status": "queued",
  "message": "Stop command queued for project my-app"
}
```

**Status Codes:**
- `200`: Success
- `404`: Project not found
- `503`: Agent not connected

### Get Project Status

**GET /projects/{project_name}/status**

Get the status of a project.

**Path Parameters:**
- `project_name`: Name of the project

**Response:**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440002",
  "status": "queued",
  "logs": "Status check queued for project my-app"
}
```

**Status Codes:**
- `200`: Success
- `404`: Project not found
- `503`: Agent not connected

## Command Types

### DEPLOY_COMPOSE

Deploy a Docker Compose file.

**Payload:**
```json
{
  "project_name": "my-app",
  "compose_file_base64": "base64-encoded-content"
}
```

### STOP_COMPOSE

Stop a Docker Compose deployment.

**Payload:**
```json
{
  "project_name": "my-app"
}
```

### STATUS

Get the status of a deployment.

**Payload:**
```json
{
  "project_name": "my-app"
}
```

### HEALTH_CHECK

Request a health check from the agent.

**Payload:** Empty object `{}`

## Response Status Values

- `QUEUED`: Command has been queued
- `IN_PROGRESS`: Command is currently executing
- `COMPLETED`: Command completed successfully
- `FAILED`: Command failed

## Example Usage

### Python

```python
import requests
import base64

# Deploy a project
with open('docker-compose.yml', 'rb') as f:
    compose_content = base64.b64encode(f.read()).decode()

response = requests.post('http://localhost:8000/deploy', json={
    'project_name': 'my-app',
    'compose_file_base64': compose_content
})

print(response.json())
```

### curl

```bash
# List agents
curl http://localhost:8000/agents

# Deploy
curl -X POST http://localhost:8000/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "project_name": "my-app",
    "compose_file_base64": "'"$(base64 -i docker-compose.yml)"'"
  }'

# Stop
curl -X POST http://localhost:8000/projects/my-app/stop \
  -H "Content-Type: application/json" \
  -d '{}'

# Status
curl http://localhost:8000/projects/my-app/status
```

### JavaScript/TypeScript

```typescript
// Deploy a project
const response = await fetch('http://localhost:8000/deploy', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    project_name: 'my-app',
    compose_file_base64: btoa(composeFileContent)
  })
});

const result = await response.json();
console.log(result);
```

## Interactive Documentation

Visit these URLs when the server is running for interactive API documentation:

- Swagger UI: `http://localhost:8000/docs`
- ReDoc: `http://localhost:8000/redoc`


