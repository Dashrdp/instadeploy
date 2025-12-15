# Testing Guide

This guide explains how to test the InstaDeploy Agent before the control plane is implemented.

## Prerequisites

1. **Go installed** (1.22+)
2. **Docker installed and running**
3. **Docker Compose available** (usually included with Docker Desktop or install separately)

## Setup

### 1. Build the Agent

```bash
# Download dependencies
go mod download

# Build the agent binary
make build-agent
```

The binary will be created at `./bin/agent`.

### 2. Create a Test WebSocket Server

Since the control plane doesn't exist yet, you need a simple WebSocket server to test against. Here's a minimal example using Node.js:

**test-server.js:**
```javascript
const WebSocket = require('ws');
const wss = new WebSocket.Server({ port: 8080 });

wss.on('connection', (ws, req) => {
  console.log('Agent connected:', req.headers);
  
  ws.on('message', (message) => {
    console.log('Received from agent:', message.toString());
  });

  // Send a test deployment command after 2 seconds
  setTimeout(() => {
    const composeContent = `version: '3.8'
services:
  nginx:
    image: nginx:alpine
    ports:
      - "8080:80"`;
    
    const command = {
      id: 'test-job-' + Date.now(),
      type: 'DEPLOY_COMPOSE',
      payload: {
        project_name: 'test-nginx',
        compose_file_base64: Buffer.from(composeContent).toString('base64')
      }
    };
    
    ws.send(JSON.stringify(command));
    console.log('Sent deployment command');
  }, 2000);
});

console.log('WebSocket server running on ws://localhost:8080');
```

**Run it:**
```bash
npm install ws
node test-server.js
```

### 3. Run the Agent

In a separate terminal:

```bash
export SERVER_URL="ws://localhost:8080/ws"
export AGENT_TOKEN="test-token"
./bin/agent
```

## Manual Testing with websocat

If you don't want to write a test server, you can use `websocat` for manual testing:

### Install websocat

```bash
# macOS
brew install websocat

# Linux
wget https://github.com/vi/websocat/releases/download/v1.11.0/websocat_amd64-linux -O websocat
chmod +x websocat
```

### Start websocat server

```bash
websocat -s 8080
```

### Start the agent (different terminal)

```bash
export SERVER_URL="ws://localhost:8080"
export AGENT_TOKEN="test-token"
./bin/agent
```

### Send commands (in websocat terminal)

**1. Health Check:**
```json
{"id":"health-1","type":"HEALTH_CHECK","payload":{}}
```

**2. Deploy Command:**
```json
{"id":"deploy-1","type":"DEPLOY_COMPOSE","payload":{"project_name":"test-app","compose_file_base64":"dmVyc2lvbjogJzMuOCcKc2VydmljZXM6CiAgbmdpbng6CiAgICBpbWFnZTogbmdpbng6YWxwaW5lCiAgICBwb3J0czoKICAgICAgLSAiODA4MDo4MCI="}}
```

**3. Status Check:**
```json
{"id":"status-1","type":"STATUS","payload":{"project_name":"test-app"}}
```

**4. Stop Deployment:**
```json
{"id":"stop-1","type":"STOP_COMPOSE","payload":{"project_name":"test-app"}}
```

## Testing with the Example Script

Use the provided example script to generate a test deployment command:

```bash
cd examples
chmod +x test-deployment.sh
./test-deployment.sh
```

This creates `deployment-command.json` which you can send via websocat or your test server.

## Automated Testing

### Unit Tests (Future)

Create test files for each component:

**agent/deploy_test.go:**
```go
package main

import (
	"testing"
)

func TestValidateProjectName(t *testing.T) {
	dm := NewDeploymentManager()
	
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid name", "my-project", false},
		{"valid with underscore", "my_project", false},
		{"invalid with slash", "my/project", true},
		{"invalid with dots", "../project", true},
		{"empty", "", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dm.ValidateProjectName(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateProjectName() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
```

Run tests:
```bash
go test ./...
```

## Integration Testing Scenarios

### Scenario 1: Basic Deployment

1. Start the agent
2. Send a DEPLOY_COMPOSE command with a simple nginx service
3. Verify the response status is COMPLETED
4. Check Docker: `docker ps` should show nginx container
5. Check project directory: `ls -la ./projects/test-app/`
6. Verify compose file exists: `cat ./projects/test-app/docker-compose.yml`

### Scenario 2: Invalid Project Name

1. Send a DEPLOY_COMPOSE with `project_name: "../../etc/passwd"`
2. Verify the response status is FAILED
3. Verify error message mentions "invalid path characters"

### Scenario 3: Stop Deployment

1. Deploy a service (as in Scenario 1)
2. Send a STOP_COMPOSE command
3. Verify response status is COMPLETED
4. Check Docker: `docker ps` should not show the container

### Scenario 4: Reconnection

1. Start the agent
2. Kill the WebSocket server
3. Observe agent logs showing reconnection attempts
4. Restart the WebSocket server
5. Verify agent reconnects successfully

### Scenario 5: Invalid Compose File

1. Send a DEPLOY_COMPOSE with invalid base64 or invalid YAML
2. Verify response status is FAILED
3. Check error message for useful debugging info

## Verification Commands

### Check Running Containers
```bash
docker ps
```

### Check All Containers (including stopped)
```bash
docker ps -a
```

### Check Project Directory
```bash
ls -la ./projects/
ls -la ./projects/test-app/
cat ./projects/test-app/docker-compose.yml
```

### Check Docker Compose Status
```bash
cd ./projects/test-app/
docker compose ps
docker compose logs
```

### Monitor Agent Logs
```bash
./bin/agent 2>&1 | tee agent.log
```

## Troubleshooting Tests

### Agent Won't Connect

**Check:**
- Is the WebSocket server running? `netstat -an | grep 8080`
- Is SERVER_URL correct? `echo $SERVER_URL`
- Check agent logs for error messages

### Docker Commands Fail

**Check:**
- Is Docker running? `docker info`
- Does user have Docker permissions? `docker ps`
- Is Docker Compose installed? `docker compose version`

### Project Directory Issues

**Check:**
- Does `./projects/` directory exist and is writable?
- File permissions: `ls -la ./projects/`
- Disk space: `df -h`

### Base64 Decoding Fails

**Verify base64 encoding:**
```bash
echo "your-base64-string" | base64 -d
```

**Generate correct base64:**
```bash
cat docker-compose.yml | base64
```

## Performance Testing

### Load Test with Multiple Concurrent Deployments

Create a script to send multiple deployment commands rapidly:

```bash
#!/bin/bash
for i in {1..10}; do
  echo "Deploying project-$i"
  # Send deployment command via your WebSocket client
done
```

Monitor:
- Agent memory usage: `ps aux | grep agent`
- Docker resource usage: `docker stats`
- Connection stability
- Response times

## Security Testing

### Test Directory Traversal Prevention

Try these malicious project names:
- `../../../etc/passwd`
- `....//etc/passwd`
- `/etc/passwd`
- `test/../../../etc/passwd`

All should fail with validation errors.

### Test Invalid Tokens

Start agent with wrong token, verify connection is rejected by server.

## Next Steps

Once the control plane is implemented:
1. Test full integration (agent ↔ control plane)
2. Test multi-agent scenarios
3. Test deployment orchestration
4. Test UI interactions
5. Test authentication/authorization
6. Performance testing with 100+ agents

