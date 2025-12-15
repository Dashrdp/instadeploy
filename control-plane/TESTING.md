# Testing the Control Plane

This guide explains how to test the Control Plane with the Go Agent.

## Prerequisites

1. Python 3.11+ installed
2. Go Agent built and ready to run
3. Docker and Docker Compose installed (for deployments)

## Setup

1. Install Python dependencies:
```bash
cd control-plane
pip install -r requirements.txt
```

2. Set environment variables:
```bash
export AGENT_SECRET_TOKEN="my-secret-token"
export DATABASE_URL="sqlite:///./control_plane.db"
```

3. Start the Control Plane:
```bash
python run.py
```

The server will start on `http://localhost:8000`.

## Configure the Agent

Update your agent environment variables to connect to the Control Plane:

```bash
export SERVER_URL="ws://localhost:8000/ws"
export AGENT_SECRET_TOKEN="my-secret-token"
export AGENT_VERSION="1.0.0"
export AGENT_ARCHITECTURE="amd64"
```

Start the agent:
```bash
cd ../agent
go run .
```

## Test the Connection

1. Check if the agent is connected:
```bash
curl http://localhost:8000/agents
```

Expected response:
```json
{
  "agents": [
    {
      "id": 1,
      "hostname": "your-hostname",
      "ip_address": "127.0.0.1",
      "status": "online",
      "last_seen": "2024-01-01T00:00:00",
      "version": "1.0.0",
      "architecture": "amd64"
    }
  ],
  "total": 1
}
```

## Test Deployment

1. Create a test docker-compose.yml:
```yaml
version: '3.8'
services:
  nginx:
    image: nginx:alpine
    ports:
      - "8080:80"
```

2. Base64 encode the file:
```bash
base64 -i test-compose.yml > compose.b64
# Or use Python:
python -c "import base64; print(base64.b64encode(open('test-compose.yml', 'rb').read()).decode())"
```

3. Deploy via API:
```bash
curl -X POST http://localhost:8000/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "project_name": "test-nginx",
    "compose_file_base64": "YOUR_BASE64_STRING_HERE"
  }'
```

Expected response:
```json
{
  "job_id": "uuid-here",
  "deployment_id": 1,
  "status": "queued",
  "message": "Deployment queued on agent hostname"
}
```

4. Check agent logs to see the deployment process.

5. Verify the deployment:
```bash
curl http://localhost:8080
```

You should see the Nginx welcome page.

## Test Stop

Stop the deployment:
```bash
curl -X POST http://localhost:8000/projects/test-nginx/stop \
  -H "Content-Type: application/json" \
  -d '{}'
```

## Test Status

Get project status:
```bash
curl http://localhost:8000/projects/test-nginx/status
```

## API Documentation

Once the server is running, visit:
- Interactive API docs: http://localhost:8000/docs
- Alternative docs: http://localhost:8000/redoc

## Health Check

Check if the Control Plane is healthy:
```bash
curl http://localhost:8000/health
```

Expected response:
```json
{
  "status": "healthy",
  "connected_agents": 1
}
```

## Troubleshooting

### Agent not connecting

1. Check the agent logs for connection errors
2. Verify the `AGENT_SECRET_TOKEN` matches on both sides
3. Ensure the WebSocket URL is correct (ws:// not wss:// for local testing)

### Deployment fails

1. Check agent logs for docker compose errors
2. Verify Docker is running on the agent machine
3. Check the docker-compose.yml syntax

### Database errors

1. Delete the database file and restart: `rm control_plane.db && python run.py`
2. Check file permissions on the database directory


