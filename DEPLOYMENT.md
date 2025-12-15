# Deployment Guide

This guide covers deploying the InstaDeploy Agent to production servers.

## Prerequisites

### Target Server Requirements

- Linux OS (Ubuntu 20.04+, Debian 11+, CentOS 8+, etc.)
- Docker Engine installed and running
- Docker Compose plugin/binary available
- Network access to the control plane (outbound connections)
- Minimum 512MB RAM, 1GB+ recommended
- Minimum 5GB disk space

### Control Plane Requirements

- WebSocket server accessible from target servers
- Valid SSL certificate for WSS (recommended for production)
- Token generation and management system

## Installation Methods

### Method 1: Manual Installation

#### Step 1: Install Docker (if not installed)

**Ubuntu/Debian:**
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

**Verify:**
```bash
docker --version
docker compose version
```

#### Step 2: Download Agent Binary

```bash
# Download the agent binary
sudo curl -L https://github.com/yourorg/instadeploy/releases/latest/download/agent-linux-amd64 \
  -o /usr/local/bin/instadeploy-agent

# Make it executable
sudo chmod +x /usr/local/bin/instadeploy-agent
```

#### Step 3: Create Project Directory

```bash
sudo mkdir -p /opt/platform/projects
sudo chmod 755 /opt/platform/projects
```

#### Step 4: Create Systemd Service

Create `/etc/systemd/system/instadeploy-agent.service`:

```ini
[Unit]
Description=InstaDeploy Agent
Documentation=https://github.com/yourorg/instadeploy
After=docker.service
Requires=docker.service
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=10
User=root

# Configuration
Environment="SERVER_URL=wss://control.example.com/api/v1/agents/ws"
Environment="AGENT_SECRET_TOKEN=YOUR_UNIQUE_TOKEN_HERE"

# Agent binary
ExecStart=/usr/local/bin/instadeploy-agent

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=instadeploy-agent

# Security (optional hardening)
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/opt/platform /var/run/docker.sock

[Install]
WantedBy=multi-user.target
```

#### Step 5: Start and Enable Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Start the service
sudo systemctl start instadeploy-agent

# Enable auto-start on boot
sudo systemctl enable instadeploy-agent

# Check status
sudo systemctl status instadeploy-agent
```

#### Step 6: Verify Operation

```bash
# Check logs
sudo journalctl -u instadeploy-agent -f

# You should see:
# - "InstaDeploy Agent starting..."
# - "Configuration loaded: ServerURL=..."
# - "Connecting to wss://..."
# - "Connected successfully!"
```

### Method 2: Cloud-Init (Automated)

For automated deployment on cloud VPS (AWS, DigitalOcean, Hetzner, etc.), use cloud-init:

**cloud-init.yaml:**
```yaml
#cloud-config

# Update packages
package_update: true
package_upgrade: true

# Install Docker
packages:
  - curl
  - ca-certificates

runcmd:
  # Install Docker
  - curl -fsSL https://get.docker.com -o /tmp/get-docker.sh
  - sh /tmp/get-docker.sh
  
  # Download agent binary
  - curl -L https://github.com/yourorg/instadeploy/releases/latest/download/agent-linux-amd64 -o /usr/local/bin/instadeploy-agent
  - chmod +x /usr/local/bin/instadeploy-agent
  
  # Create project directory
  - mkdir -p /opt/platform/projects
  - chmod 755 /opt/platform/projects
  
  # Create systemd service
  - |
    cat > /etc/systemd/system/instadeploy-agent.service << 'EOF'
    [Unit]
    Description=InstaDeploy Agent
    After=docker.service
    Requires=docker.service
    
    [Service]
    Type=simple
    Restart=always
    RestartSec=10
    User=root
    Environment="SERVER_URL=wss://control.example.com/api/v1/agents/ws"
    Environment="AGENT_SECRET_TOKEN=__REPLACE_WITH_UNIQUE_TOKEN__"
    ExecStart=/usr/local/bin/instadeploy-agent
    StandardOutput=journal
    StandardError=journal
    
    [Install]
    WantedBy=multi-user.target
    EOF
  
  # Start service
  - systemctl daemon-reload
  - systemctl enable instadeploy-agent
  - systemctl start instadeploy-agent

# Optional: Set hostname
hostname: agent-node-1
fqdn: agent-node-1.example.com
```

**Usage:**
- AWS EC2: Paste into "User Data" field
- DigitalOcean: Use "User Data" in droplet creation
- Hetzner: Use "Cloud Config" in server creation

### Method 3: Docker Container (Agent in Docker)

If you want to run the agent itself as a Docker container:

**Dockerfile:**
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o agent ./agent

FROM alpine:latest
RUN apk --no-cache add ca-certificates docker-cli
WORKDIR /root/
COPY --from=builder /app/agent .
CMD ["./agent"]
```

**docker-compose.yml:**
```yaml
version: '3.8'

services:
  agent:
    build: .
    restart: unless-stopped
    environment:
      - SERVER_URL=wss://control.example.com/api/v1/agents/ws
      - AGENT_SECRET_TOKEN=${AGENT_SECRET_TOKEN}
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /opt/platform/projects:/opt/platform/projects
```

**Run:**
```bash
export AGENT_SECRET_TOKEN="your-token"
docker compose up -d
```

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SERVER_URL` | Yes | - | WebSocket URL of control plane |
| `AGENT_TOKEN` | Yes | - | Agent authentication token |
| `AGENT_SECRET_TOKEN` | No | `AGENT_TOKEN` | Secret token for auth |

### Security Considerations

#### 1. Use WSS (WebSocket Secure) in Production

```bash
# Good (encrypted)
SERVER_URL=wss://control.example.com/ws

# Bad (unencrypted, only for testing)
SERVER_URL=ws://control.example.com/ws
```

#### 2. Generate Strong Tokens

```bash
# Generate a secure token
openssl rand -base64 32
# Output: 8H3jKl9mP2qR5sT7vW9xZ0aB1cD3eF4g

# Use this as AGENT_SECRET_TOKEN
```

#### 3. Firewall Configuration

The agent only needs **outbound** connections. No inbound ports required.

**UFW (Ubuntu/Debian):**
```bash
# Allow Docker (if needed)
sudo ufw allow 2375/tcp
sudo ufw allow 2376/tcp

# Default: deny incoming, allow outgoing
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw enable
```

#### 4. Run with Limited Permissions

While the agent needs Docker access, you can limit other permissions:

```bash
# Create dedicated user (optional)
sudo useradd -r -s /bin/false instadeploy
sudo usermod -aG docker instadeploy

# Update systemd service to use this user
User=instadeploy
```

## Monitoring

### System Logs

```bash
# View logs in real-time
sudo journalctl -u instadeploy-agent -f

# View last 100 lines
sudo journalctl -u instadeploy-agent -n 100

# View logs from last hour
sudo journalctl -u instadeploy-agent --since "1 hour ago"
```

### Health Check

Create a simple health check script:

**check-agent.sh:**
```bash
#!/bin/bash

if systemctl is-active --quiet instadeploy-agent; then
  echo "Agent is running"
  exit 0
else
  echo "Agent is NOT running"
  exit 1
fi
```

### Resource Monitoring

```bash
# CPU and Memory usage
ps aux | grep instadeploy-agent

# Detailed stats
systemctl status instadeploy-agent

# Docker container stats (for deployed apps)
docker stats
```

## Troubleshooting

### Agent Won't Start

**Check service status:**
```bash
sudo systemctl status instadeploy-agent
```

**Check logs:**
```bash
sudo journalctl -u instadeploy-agent -n 50
```

**Common issues:**
- Missing environment variables
- Docker not running: `sudo systemctl start docker`
- Binary not executable: `sudo chmod +x /usr/local/bin/instadeploy-agent`
- Invalid SERVER_URL format

### Connection Issues

**Test connectivity:**
```bash
# Test WebSocket endpoint (requires websocat)
websocat wss://control.example.com/api/v1/agents/ws
```

**Check DNS:**
```bash
nslookup control.example.com
ping control.example.com
```

**Check firewall:**
```bash
sudo iptables -L -n
sudo ufw status
```

### Docker Permission Issues

**Error: "Cannot connect to Docker daemon"**

```bash
# Check Docker is running
sudo systemctl status docker

# Check socket permissions
ls -la /var/run/docker.sock

# Add user to docker group
sudo usermod -aG docker $USER

# Restart agent
sudo systemctl restart instadeploy-agent
```

### Disk Space Issues

**Check disk usage:**
```bash
df -h
du -sh /opt/platform/projects/*
```

**Clean up Docker:**
```bash
# Remove unused containers, images, and volumes
docker system prune -a --volumes
```

## Updating the Agent

### Method 1: Manual Update

```bash
# Download new version
sudo curl -L https://github.com/yourorg/instadeploy/releases/latest/download/agent-linux-amd64 \
  -o /usr/local/bin/instadeploy-agent

# Make executable
sudo chmod +x /usr/local/bin/instadeploy-agent

# Restart service
sudo systemctl restart instadeploy-agent
```

### Method 2: Automated Update (Future)

The agent will support self-update via command from control plane:

```json
{
  "id": "update-1",
  "type": "SELF_UPDATE",
  "payload": {
    "version": "1.1.0",
    "download_url": "https://releases.example.com/agent-1.1.0-linux-amd64"
  }
}
```

## Uninstallation

### Remove Agent

```bash
# Stop and disable service
sudo systemctl stop instadeploy-agent
sudo systemctl disable instadeploy-agent

# Remove service file
sudo rm /etc/systemd/system/instadeploy-agent.service

# Remove binary
sudo rm /usr/local/bin/instadeploy-agent

# Remove project data (optional, be careful!)
sudo rm -rf /opt/platform/projects

# Reload systemd
sudo systemctl daemon-reload
```

### Remove Docker (if desired)

```bash
# Remove Docker packages (Ubuntu/Debian)
sudo apt-get remove docker docker-engine docker.io containerd runc
sudo apt-get purge docker-ce docker-ce-cli containerd.io

# Remove Docker data
sudo rm -rf /var/lib/docker
sudo rm -rf /var/lib/containerd
```

## Best Practices

1. **Use SSL/TLS**: Always use `wss://` in production
2. **Rotate Tokens**: Regularly rotate AGENT_SECRET_TOKEN
3. **Monitor Logs**: Set up log aggregation (ELK, Loki, etc.)
4. **Backup**: Backup `/opt/platform/projects` regularly
5. **Resource Limits**: Set Docker resource limits for deployments
6. **Updates**: Keep agent and Docker updated
7. **Testing**: Test deployments in staging before production
8. **Monitoring**: Use Prometheus/Grafana for metrics

## Production Checklist

- [ ] Docker installed and running
- [ ] Agent binary downloaded and executable
- [ ] Systemd service configured with correct tokens
- [ ] Service started and enabled
- [ ] Logs showing successful connection
- [ ] WSS (not WS) for encrypted communication
- [ ] Firewall configured (outbound only)
- [ ] Monitoring/alerting set up
- [ ] Backup strategy in place
- [ ] Update procedure documented
- [ ] Rollback plan prepared

