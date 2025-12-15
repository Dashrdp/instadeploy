# Test WebSocket Server

This is a simple WebSocket server for testing the InstaDeploy Agent before the control plane is implemented.

## Features

- Accepts agent connections with authentication headers
- Automatically sends test commands:
  1. HEALTH_CHECK
  2. DEPLOY_COMPOSE (nginx + redis)
  3. STATUS check
- Logs all agent responses
- Maintains connection with periodic pings

## Usage

### Build and Run

```bash
cd test-server
go mod download
go build -o test-server
./test-server
```

The server will start on `ws://localhost:8080/ws`.

### Run the Agent

In a separate terminal:

```bash
cd ..
export SERVER_URL="ws://localhost:8080/ws"
export AGENT_TOKEN="test-token"
./bin/agent
```

### What to Expect

1. Agent connects with headers
2. Server logs connection details
3. After 2 seconds: Health check command sent
4. After 5 seconds: Deploy command sent (nginx + redis)
5. After 15 seconds: Status check sent
6. Agent executes commands and sends responses
7. Server logs all responses

### Example Output

**Server logs:**
```
Test WebSocket server starting on :8080
Waiting for agent connections at ws://localhost:8080/ws
Authorization header: Bearer test-token
Agent connected - Version: 1.0.0, Architecture: darwin-arm64, Hostname: myhost
WebSocket connection established
Sending HEALTH_CHECK command...
Received from agent: {"job_id":"health-123","status":"COMPLETED",...}
Sending DEPLOY_COMPOSE command...
Received from agent: {"job_id":"deploy-456","status":"IN_PROGRESS",...}
Received from agent: {"job_id":"deploy-456","status":"COMPLETED",...}
```

**Agent logs:**
```
InstaDeploy Agent starting...
Connecting to ws://localhost:8080/ws...
Connected successfully!
Received command: ID=health-123, Type=HEALTH_CHECK
Sent response: JobID=health-123, Status=COMPLETED
Received command: ID=deploy-456, Type=DEPLOY_COMPOSE
Deploying project: test-nginx-redis
Sent response: JobID=deploy-456, Status=COMPLETED
```

## Verify Deployment

After the agent deploys the services:

```bash
# Check running containers
docker ps

# Check project directory
ls -la ./projects/test-nginx-redis/

# Check compose file
cat ./projects/test-nginx-redis/docker-compose.yml

# Test nginx
curl http://localhost:8081

# Test redis
redis-cli -p 6380 ping
```

## Cleanup

Stop the deployment:

```bash
cd ./projects/test-nginx-redis/
docker compose down
```

Or use the test server to send a STOP command (modify server.go to add stop command).

