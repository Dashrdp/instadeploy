"""Configuration management for Control Plane."""
import os
from typing import Optional


class Config:
    """Application configuration."""
    
    # Agent authentication
    AGENT_SECRET_TOKEN: str = os.getenv("AGENT_SECRET_TOKEN", "default-secret-token")
    
    # Database
    DATABASE_URL: str = os.getenv("DATABASE_URL", "sqlite:///./control_plane.db")
    
    # Server
    HOST: str = os.getenv("HOST", "0.0.0.0")
    PORT: int = int(os.getenv("PORT", "8000"))
    
    # WebSocket settings
    PING_INTERVAL: int = 30  # seconds
    PONG_TIMEOUT: int = 60  # seconds


config = Config()


