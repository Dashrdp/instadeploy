"""Control Plane FastAPI application."""
import logging
import uuid
from datetime import datetime
from contextlib import asynccontextmanager
from typing import Optional

from fastapi import FastAPI, WebSocket, WebSocketDisconnect, HTTPException, Depends, Header
from fastapi.responses import JSONResponse
from sqlmodel import Session, select

from config import config
from database import init_db, get_session, engine
from models import Agent, Deployment, Job
from schemas import (
    Command, Response, CommandType, ResponseStatus,
    DeployRequest, DeployResponse, DeployPayload,
    StopRequest, StopResponse, StopPayload,
    StatusResponse, StatusPayload,
    AgentResponse, AgentListResponse, AgentInfo
)
from manager import connection_manager

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan events."""
    # Startup
    logger.info("Starting Control Plane...")
    init_db()
    logger.info("Database initialized")
    yield
    # Shutdown
    logger.info("Shutting down Control Plane...")


app = FastAPI(
    title="InstaDeploy Control Plane",
    description="Control Plane for managing PaaS agents",
    version="1.0.0",
    lifespan=lifespan
)


# Authentication helper
def verify_agent_token(authorization: Optional[str] = Header(None)) -> bool:
    """Verify agent authentication token."""
    if not authorization:
        return False
    
    # Expected format: "Bearer <token>"
    parts = authorization.split()
    if len(parts) != 2 or parts[0].lower() != "bearer":
        return False
    
    token = parts[1]
    return token == config.AGENT_SECRET_TOKEN


# WebSocket endpoint for agents
@app.websocket("/ws")
async def websocket_endpoint(
    websocket: WebSocket,
    authorization: Optional[str] = Header(None),
    x_agent_version: Optional[str] = Header(None),
    x_agent_architecture: Optional[str] = Header(None),
    x_agent_hostname: Optional[str] = Header(None)
):
    """WebSocket endpoint for agent connections."""
    
    # Verify authentication
    if not verify_agent_token(authorization):
        logger.warning("Agent connection rejected: invalid token")
        await websocket.close(code=1008, reason="Unauthorized")
        return
    
    # Extract agent info from headers
    hostname = x_agent_hostname or "unknown"
    version = x_agent_version or "unknown"
    architecture = x_agent_architecture or "unknown"
    
    # Get client IP
    client_host = websocket.client.host if websocket.client else None
    
    logger.info(f"Agent connecting: hostname={hostname}, version={version}, arch={architecture}, ip={client_host}")
    
    # Register or update agent in database
    with Session(engine) as session:
        # Check if agent exists
        statement = select(Agent).where(Agent.hostname == hostname)
        agent = session.exec(statement).first()
        
        if agent:
            # Update existing agent
            agent.status = "online"
            agent.last_seen = datetime.utcnow()
            agent.version = version
            agent.architecture = architecture
            agent.ip_address = client_host
            session.add(agent)
            session.commit()
            session.refresh(agent)
            logger.info(f"Updated existing agent: id={agent.id}, hostname={hostname}")
        else:
            # Create new agent
            agent = Agent(
                hostname=hostname,
                ip_address=client_host,
                status="online",
                version=version,
                architecture=architecture,
                last_seen=datetime.utcnow()
            )
            session.add(agent)
            session.commit()
            session.refresh(agent)
            logger.info(f"Registered new agent: id={agent.id}, hostname={hostname}")
        
        agent_id = agent.id
    
    # Connect to WebSocket manager
    await connection_manager.connect(agent_id, hostname, websocket)
    
    try:
        # Listen for responses from agent
        while True:
            response = await connection_manager.receive_response(agent_id, websocket)
            
            if response is None:
                # Connection closed or error
                break
            
            # Update job status in database
            with Session(engine) as session:
                statement = select(Job).where(Job.id == response.job_id)
                job = session.exec(statement).first()
                
                if job:
                    job.status = response.status.value.lower()
                    job.logs = response.logs
                    job.error = response.error
                    job.updated_at = datetime.utcnow()
                    session.add(job)
                    
                    # Update deployment if exists
                    if job.deployment_id:
                        statement = select(Deployment).where(Deployment.id == job.deployment_id)
                        deployment = session.exec(statement).first()
                        if deployment:
                            if response.status == ResponseStatus.COMPLETED:
                                if job.command_type == CommandType.DEPLOY_COMPOSE.value:
                                    deployment.status = "running"
                                elif job.command_type == CommandType.STOP_COMPOSE.value:
                                    deployment.status = "stopped"
                            elif response.status == ResponseStatus.FAILED:
                                deployment.status = "failed"
                            
                            deployment.last_logs = response.logs
                            deployment.last_error = response.error
                            deployment.updated_at = datetime.utcnow()
                            session.add(deployment)
                    
                    session.commit()
                    logger.info(f"Updated job {response.job_id}: status={response.status}")
    
    except WebSocketDisconnect:
        logger.info(f"Agent {agent_id} ({hostname}) disconnected")
    except Exception as e:
        logger.error(f"Error in WebSocket handler for agent {agent_id}: {e}")
    finally:
        # Mark agent as offline
        connection_manager.disconnect(agent_id)
        
        with Session(engine) as session:
            statement = select(Agent).where(Agent.id == agent_id)
            agent = session.exec(statement).first()
            if agent:
                agent.status = "offline"
                agent.last_seen = datetime.utcnow()
                session.add(agent)
                session.commit()


# REST API Endpoints

@app.get("/")
async def root():
    """Root endpoint."""
    return {
        "service": "InstaDeploy Control Plane",
        "version": "1.0.0",
        "status": "running"
    }


@app.get("/agents", response_model=AgentListResponse)
async def list_agents(session: Session = Depends(get_session)):
    """List all agents."""
    statement = select(Agent)
    agents = session.exec(statement).all()
    
    agent_responses = [
        AgentResponse(
            id=agent.id,
            hostname=agent.hostname,
            ip_address=agent.ip_address,
            status=agent.status,
            last_seen=agent.last_seen,
            version=agent.version,
            architecture=agent.architecture
        )
        for agent in agents
    ]
    
    return AgentListResponse(agents=agent_responses, total=len(agent_responses))


@app.post("/deploy", response_model=DeployResponse)
async def deploy_project(
    request: DeployRequest,
    session: Session = Depends(get_session)
):
    """Deploy a project to an agent."""
    
    # Find target agent
    if request.agent_id:
        statement = select(Agent).where(Agent.id == request.agent_id)
        agent = session.exec(statement).first()
        if not agent:
            raise HTTPException(status_code=404, detail="Agent not found")
    else:
        # Use first available online agent
        statement = select(Agent).where(Agent.status == "online")
        agent = session.exec(statement).first()
        if not agent:
            raise HTTPException(status_code=503, detail="No agents available")
    
    # Check if agent is connected
    if not connection_manager.is_connected(agent.id):
        raise HTTPException(status_code=503, detail=f"Agent {agent.id} is not connected")
    
    # Create deployment record
    deployment = Deployment(
        project_name=request.project_name,
        agent_id=agent.id,
        status="pending"
    )
    session.add(deployment)
    session.commit()
    session.refresh(deployment)
    
    # Create job record
    job_id = str(uuid.uuid4())
    job = Job(
        id=job_id,
        deployment_id=deployment.id,
        agent_id=agent.id,
        command_type=CommandType.DEPLOY_COMPOSE.value,
        status="queued"
    )
    session.add(job)
    session.commit()
    
    # Create and send command
    payload = DeployPayload(
        project_name=request.project_name,
        compose_file_base64=request.compose_file_base64
    )
    command = Command(
        id=job_id,
        type=CommandType.DEPLOY_COMPOSE,
        payload=payload.model_dump()
    )
    
    success = await connection_manager.send_command(agent.id, command)
    
    if not success:
        # Update job status
        job.status = "failed"
        job.error = "Failed to send command to agent"
        session.add(job)
        session.commit()
        raise HTTPException(status_code=503, detail="Failed to send command to agent")
    
    return DeployResponse(
        job_id=job_id,
        deployment_id=deployment.id,
        status="queued",
        message=f"Deployment queued on agent {agent.hostname}"
    )


@app.post("/projects/{project_name}/stop", response_model=StopResponse)
async def stop_project(
    project_name: str,
    request: StopRequest,
    session: Session = Depends(get_session)
):
    """Stop a project."""
    
    # Find deployment
    statement = select(Deployment).where(
        Deployment.project_name == project_name
    ).order_by(Deployment.created_at.desc())
    deployment = session.exec(statement).first()
    
    if not deployment:
        raise HTTPException(status_code=404, detail="Project not found")
    
    agent_id = request.agent_id or deployment.agent_id
    
    # Check if agent is connected
    if not connection_manager.is_connected(agent_id):
        raise HTTPException(status_code=503, detail=f"Agent {agent_id} is not connected")
    
    # Create job record
    job_id = str(uuid.uuid4())
    job = Job(
        id=job_id,
        deployment_id=deployment.id,
        agent_id=agent_id,
        command_type=CommandType.STOP_COMPOSE.value,
        status="queued"
    )
    session.add(job)
    session.commit()
    
    # Create and send command
    payload = StopPayload(project_name=project_name)
    command = Command(
        id=job_id,
        type=CommandType.STOP_COMPOSE,
        payload=payload.model_dump()
    )
    
    success = await connection_manager.send_command(agent_id, command)
    
    if not success:
        job.status = "failed"
        job.error = "Failed to send command to agent"
        session.add(job)
        session.commit()
        raise HTTPException(status_code=503, detail="Failed to send command to agent")
    
    return StopResponse(
        job_id=job_id,
        status="queued",
        message=f"Stop command queued for project {project_name}"
    )


@app.get("/projects/{project_name}/status", response_model=StatusResponse)
async def get_project_status(
    project_name: str,
    session: Session = Depends(get_session)
):
    """Get project status."""
    
    # Find deployment
    statement = select(Deployment).where(
        Deployment.project_name == project_name
    ).order_by(Deployment.created_at.desc())
    deployment = session.exec(statement).first()
    
    if not deployment:
        raise HTTPException(status_code=404, detail="Project not found")
    
    agent_id = deployment.agent_id
    
    # Check if agent is connected
    if not connection_manager.is_connected(agent_id):
        raise HTTPException(status_code=503, detail=f"Agent {agent_id} is not connected")
    
    # Create job record
    job_id = str(uuid.uuid4())
    job = Job(
        id=job_id,
        deployment_id=deployment.id,
        agent_id=agent_id,
        command_type=CommandType.STATUS.value,
        status="queued"
    )
    session.add(job)
    session.commit()
    
    # Create and send command
    payload = StatusPayload(project_name=project_name)
    command = Command(
        id=job_id,
        type=CommandType.STATUS,
        payload=payload.model_dump()
    )
    
    success = await connection_manager.send_command(agent_id, command)
    
    if not success:
        job.status = "failed"
        job.error = "Failed to send command to agent"
        session.add(job)
        session.commit()
        raise HTTPException(status_code=503, detail="Failed to send command to agent")
    
    return StatusResponse(
        job_id=job_id,
        status="queued",
        logs=f"Status check queued for project {project_name}"
    )


@app.get("/health")
async def health_check():
    """Health check endpoint."""
    return {
        "status": "healthy",
        "connected_agents": len(connection_manager.get_connected_agents())
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host=config.HOST, port=config.PORT)


