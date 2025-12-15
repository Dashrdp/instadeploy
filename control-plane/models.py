"""Database models for Control Plane."""
from datetime import datetime
from typing import Optional
from sqlmodel import SQLModel, Field


class Agent(SQLModel, table=True):
    """Agent database model."""
    
    id: Optional[int] = Field(default=None, primary_key=True)
    hostname: str = Field(index=True, unique=True)
    ip_address: Optional[str] = None
    status: str = Field(default="offline")  # online, offline
    last_seen: datetime = Field(default_factory=datetime.utcnow)
    version: Optional[str] = None
    architecture: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)


class Deployment(SQLModel, table=True):
    """Deployment database model."""
    
    id: Optional[int] = Field(default=None, primary_key=True)
    project_name: str = Field(index=True)
    agent_id: int = Field(foreign_key="agent.id")
    status: str = Field(default="pending")  # pending, running, stopped, failed
    compose_file_hash: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    
    # Deployment metadata
    last_logs: Optional[str] = None
    last_error: Optional[str] = None


class Job(SQLModel, table=True):
    """Job execution tracking."""
    
    id: str = Field(primary_key=True)  # UUID from command
    deployment_id: Optional[int] = Field(default=None, foreign_key="deployment.id")
    agent_id: int = Field(foreign_key="agent.id")
    command_type: str  # DEPLOY_COMPOSE, STOP_COMPOSE, STATUS, etc.
    status: str = Field(default="queued")  # queued, in_progress, completed, failed
    logs: Optional[str] = None
    error: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)


