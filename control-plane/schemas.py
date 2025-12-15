"""Pydantic schemas for API and WebSocket protocol."""
from enum import Enum
from typing import Optional, Any
from pydantic import BaseModel, Field
from datetime import datetime


# Command Types (matching Go agent)
class CommandType(str, Enum):
    """Command types sent from control plane to agent."""
    DEPLOY_COMPOSE = "DEPLOY_COMPOSE"
    STOP_COMPOSE = "STOP_COMPOSE"
    STATUS = "STATUS"
    HEALTH_CHECK = "HEALTH_CHECK"


# Response Status (matching Go agent)
class ResponseStatus(str, Enum):
    """Response status from agent."""
    COMPLETED = "COMPLETED"
    FAILED = "FAILED"
    IN_PROGRESS = "IN_PROGRESS"
    QUEUED = "QUEUED"


# Command and Payload Schemas
class DeployPayload(BaseModel):
    """Payload for DEPLOY_COMPOSE command."""
    project_name: str
    compose_file_base64: str


class StopPayload(BaseModel):
    """Payload for STOP_COMPOSE command."""
    project_name: str


class StatusPayload(BaseModel):
    """Payload for STATUS command."""
    project_name: str


class Command(BaseModel):
    """Command sent from control plane to agent."""
    id: str
    type: CommandType
    payload: Any  # Will be one of the payload types above


class Response(BaseModel):
    """Response sent from agent to control plane."""
    job_id: str
    status: ResponseStatus
    logs: str
    error: Optional[str] = None
    data: Optional[Any] = None


# API Request/Response Schemas
class DeployRequest(BaseModel):
    """Request to deploy a project."""
    project_name: str = Field(..., description="Unique project name")
    compose_file_base64: str = Field(..., description="Base64-encoded docker-compose.yml")
    agent_id: Optional[int] = Field(None, description="Target agent ID (optional, will use first available)")


class DeployResponse(BaseModel):
    """Response from deploy request."""
    job_id: str
    deployment_id: int
    status: str
    message: str


class StopRequest(BaseModel):
    """Request to stop a project."""
    agent_id: Optional[int] = Field(None, description="Target agent ID (optional)")


class StopResponse(BaseModel):
    """Response from stop request."""
    job_id: str
    status: str
    message: str


class StatusResponse(BaseModel):
    """Response from status request."""
    job_id: str
    status: str
    logs: str


class AgentResponse(BaseModel):
    """Agent information response."""
    id: int
    hostname: str
    ip_address: Optional[str]
    status: str
    last_seen: datetime
    version: Optional[str]
    architecture: Optional[str]


class AgentListResponse(BaseModel):
    """List of agents response."""
    agents: list[AgentResponse]
    total: int


# Agent Registration
class AgentInfo(BaseModel):
    """Agent information from handshake headers."""
    hostname: str
    version: str
    architecture: str
    ip_address: Optional[str] = None


