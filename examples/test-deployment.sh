#!/bin/bash

# This script demonstrates how to send a deployment command to the agent
# It encodes a docker-compose.yml file in base64 and creates a JSON command

# Read the docker-compose file and encode it in base64
COMPOSE_FILE="./docker-compose-test.yml"
if [ ! -f "$COMPOSE_FILE" ]; then
    echo "Error: $COMPOSE_FILE not found"
    exit 1
fi

COMPOSE_BASE64=$(base64 < "$COMPOSE_FILE")

# Create the JSON command
cat <<EOF > deployment-command.json
{
  "id": "test-job-$(date +%s)",
  "type": "DEPLOY_COMPOSE",
  "payload": {
    "project_name": "test-nginx-redis",
    "compose_file_base64": "$COMPOSE_BASE64"
  }
}
EOF

echo "Created deployment-command.json"
echo "This JSON can be sent to the agent via WebSocket"
echo ""
cat deployment-command.json | jq .

