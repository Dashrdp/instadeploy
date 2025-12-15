#!/usr/bin/env python3
"""Run the Control Plane server."""
import uvicorn
from config import config

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host=config.HOST,
        port=config.PORT,
        reload=True,  # Enable auto-reload for development
        log_level="info"
    )


