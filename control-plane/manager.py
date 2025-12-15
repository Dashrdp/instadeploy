"""WebSocket connection manager for agents."""
import json
import logging
from typing import Dict, Optional
from fastapi import WebSocket
from schemas import Command, Response

logger = logging.getLogger(__name__)


class ConnectionManager:
    """Manages WebSocket connections from agents."""
    
    def __init__(self):
        # Map agent_id -> WebSocket connection
        self.active_connections: Dict[int, WebSocket] = {}
        # Map agent_id -> hostname for logging
        self.agent_hostnames: Dict[int, str] = {}
    
    async def connect(self, agent_id: int, hostname: str, websocket: WebSocket):
        """Register a new agent connection."""
        await websocket.accept()
        self.active_connections[agent_id] = websocket
        self.agent_hostnames[agent_id] = hostname
        logger.info(f"Agent connected: id={agent_id}, hostname={hostname}")
    
    def disconnect(self, agent_id: int):
        """Unregister an agent connection."""
        if agent_id in self.active_connections:
            del self.active_connections[agent_id]
            hostname = self.agent_hostnames.get(agent_id, "unknown")
            logger.info(f"Agent disconnected: id={agent_id}, hostname={hostname}")
            if agent_id in self.agent_hostnames:
                del self.agent_hostnames[agent_id]
    
    async def send_command(self, agent_id: int, command: Command) -> bool:
        """Send a command to a specific agent.
        
        Returns:
            True if sent successfully, False if agent not connected.
        """
        if agent_id not in self.active_connections:
            logger.error(f"Agent {agent_id} not connected")
            return False
        
        websocket = self.active_connections[agent_id]
        try:
            # Convert command to JSON
            command_json = command.model_dump_json()
            await websocket.send_text(command_json)
            logger.info(f"Sent command to agent {agent_id}: type={command.type}, id={command.id}")
            return True
        except Exception as e:
            logger.error(f"Failed to send command to agent {agent_id}: {e}")
            self.disconnect(agent_id)
            return False
    
    async def receive_response(self, agent_id: int, websocket: WebSocket) -> Optional[Response]:
        """Receive a response from an agent.
        
        Returns:
            Response object or None if connection closed/error.
        """
        try:
            data = await websocket.receive_text()
            response_dict = json.loads(data)
            response = Response(**response_dict)
            logger.info(f"Received response from agent {agent_id}: job_id={response.job_id}, status={response.status}")
            return response
        except Exception as e:
            logger.error(f"Error receiving response from agent {agent_id}: {e}")
            return None
    
    def is_connected(self, agent_id: int) -> bool:
        """Check if an agent is connected."""
        return agent_id in self.active_connections
    
    def get_connected_agents(self) -> list[int]:
        """Get list of connected agent IDs."""
        return list(self.active_connections.keys())
    
    async def broadcast(self, command: Command):
        """Broadcast a command to all connected agents."""
        for agent_id in self.active_connections.keys():
            await self.send_command(agent_id, command)


# Global connection manager instance
connection_manager = ConnectionManager()


