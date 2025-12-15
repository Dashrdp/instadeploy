# InstaDeploy Quick Start Guide

Get InstaDeploy up and running in 5 minutes.

## Option 1: Local Development (Recommended for Testing)

### Step 1: Start Control Plane

```bash
cd control-plane
pip install -r requirements.txt
export AGENT_SECRET_TOKEN="test-secret-token"
python run.py
```

The Control Plane will be available at `http://localhost:8000`.

### Step 2: Start Agent

Open a new terminal:

```bash
export SERVER_URL="ws://localhost:8000/ws"
export AGENT_SECRET_TOKEN="test-secret-token"
export AGENT_VERSION="1.0.0"
export AGENT_ARCHITECTURE="amd64"

# Build and run agent
make build-agent
./bin/agent
```

### Step 3: Verify Connection

```bash
curl http://localhost:8000/agents
```

You should see your agent listed as "online".

### Step 4: Deploy Something

Create a test `docker-compose.yml`:

```yaml
version: '3.8'
services:
  web:
    image: nginx:alpine
    ports:
      - "8080:80"
```

Deploy it:

```bash
# Base64 encode the file
COMPOSE_B64=$(base64 -i docker-compose.yml)

# Deploy via API
curl -X POST http://localhost:8000/deploy \
  -H "Content-Type: application/json" \
  -d "{
    \"project_name\": \"my-first-app\",
    \"compose_file_base64\": \"$COMPOSE_B64\"
  }"
```

Check the deployment:

```bash
# Check agent logs to see deployment progress
# Visit http://localhost:8080 to see nginx running
curl http://localhost:8080
```

### Step 5: Manage Your Deployment

Stop the deployment:

```bash
curl -X POST http://localhost:8000/projects/my-first-app/stop \
  -H "Content-Type: application/json" \
  -d '{}'
```

Check status:

```bash
curl http://localhost:8000/projects/my-first-app/status
```

## Option 2: Docker Compose (Control Plane Only)

The Control Plane can run in Docker, but agents should run directly on host machines:

```bash
cd control-plane
docker-compose up -d
```

Then run agents on your VPS instances pointing to the Control Plane URL.

## Option 3: Production Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for production deployment instructions.

## Next Steps

- **API Documentation**: Visit `http://localhost:8000/docs` for interactive API docs
- **Testing**: See [control-plane/TESTING.md](control-plane/TESTING.md) for comprehensive tests
- **Architecture**: Read [ARCHITECTURE.md](ARCHITECTURE.md) to understand the system design

## Troubleshooting

### Agent won't connect

1. Check the `AGENT_SECRET_TOKEN` matches on both sides
2. Verify the `SERVER_URL` is correct (use `ws://` not `wss://` for local testing)
3. Check Control Plane logs for authentication errors

### Deployment fails

1. Check agent logs for Docker errors
2. Verify Docker is running: `docker ps`
3. Check the docker-compose.yml syntax
4. Ensure ports aren't already in use

### Can't see my agent

1. Verify agent is running and connected (check agent logs)
2. Query agents endpoint: `curl http://localhost:8000/agents`
3. Check network connectivity between agent and control plane

## Example Commands

```bash
# List all agents
curl http://localhost:8000/agents

# Deploy from a compose file
curl -X POST http://localhost:8000/deploy \
  -H "Content-Type: application/json" \
  -d "{
    \"project_name\": \"myapp\",
    \"compose_file_base64\": \"$(base64 -i docker-compose.yml)\"
  }"

# Stop a project
curl -X POST http://localhost:8000/projects/myapp/stop \
  -H "Content-Type: application/json" \
  -d '{}'

# Get project status
curl http://localhost:8000/projects/myapp/status

# Health check
curl http://localhost:8000/health
```

## Running the Test Suite

Control Plane tests:

```bash
cd control-plane
python test_api.py
```

This will run through deployment, status, and stop operations automatically.


