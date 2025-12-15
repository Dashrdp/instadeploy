# Control Plane

The Control Plane manages PaaS agents via WebSocket connections and exposes a REST API for deployments.

## Features

- WebSocket-based agent communication
- REST API for deployment management
- SQLite database for state management
- Agent health monitoring
- Deployment lifecycle management

## Setup

1. Install dependencies:
```bash
pip install -r requirements.txt
```

2. Configure environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Run the server:
```bash
uvicorn main:app --host 0.0.0.0 --port 8000 --reload
```

## Docker

Build and run with Docker:
```bash
docker build -t control-plane .
docker run -p 8000:8000 --env-file .env control-plane
```

## API Endpoints

### Agent Management
- `GET /agents` - List all agents

### Deployment Management
- `POST /deploy` - Deploy a project
- `POST /projects/{name}/stop` - Stop a project
- `GET /projects/{name}/status` - Get project status

### WebSocket
- `WS /ws` - Agent WebSocket connection endpoint

## Authentication

Agents must authenticate using a Bearer token in the `Authorization` header when connecting to the WebSocket endpoint.


