package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for testing
	},
}

const testComposeFile = `version: '3.8'

services:
  nginx:
    image: nginx:alpine
    ports:
      - "8081:80"
    restart: unless-stopped

  redis:
    image: redis:alpine
    ports:
      - "6380:6379"
    restart: unless-stopped
`

type Command struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type DeployPayload struct {
	ProjectName       string `json:"project_name"`
	ComposeFileBase64 string `json:"compose_file_base64"`
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Check authorization header
	authHeader := r.Header.Get("Authorization")
	log.Printf("Authorization header: %s", authHeader)

	version := r.Header.Get("X-Agent-Version")
	arch := r.Header.Get("X-Agent-Architecture")
	hostname := r.Header.Get("X-Agent-Hostname")

	log.Printf("Agent connected - Version: %s, Architecture: %s, Hostname: %s", version, arch, hostname)

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	log.Println("WebSocket connection established")

	// Read messages from agent
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Read error: %v", err)
				return
			}
			log.Printf("Received from agent: %s", string(message))
		}
	}()

	// Send test commands
	time.Sleep(2 * time.Second)

	// 1. Health check
	log.Println("Sending HEALTH_CHECK command...")
	healthCmd := Command{
		ID:      fmt.Sprintf("health-%d", time.Now().Unix()),
		Type:    "HEALTH_CHECK",
		Payload: json.RawMessage(`{}`),
	}
	if err := conn.WriteJSON(healthCmd); err != nil {
		log.Printf("Failed to send health check: %v", err)
		return
	}

	time.Sleep(3 * time.Second)

	// 2. Deploy command
	log.Println("Sending DEPLOY_COMPOSE command...")
	deployPayload := DeployPayload{
		ProjectName:       "test-nginx-redis",
		ComposeFileBase64: base64.StdEncoding.EncodeToString([]byte(testComposeFile)),
	}
	payloadJSON, _ := json.Marshal(deployPayload)

	deployCmd := Command{
		ID:      fmt.Sprintf("deploy-%d", time.Now().Unix()),
		Type:    "DEPLOY_COMPOSE",
		Payload: payloadJSON,
	}
	if err := conn.WriteJSON(deployCmd); err != nil {
		log.Printf("Failed to send deploy command: %v", err)
		return
	}

	time.Sleep(10 * time.Second)

	// 3. Status check
	log.Println("Sending STATUS command...")
	statusPayload := map[string]string{"project_name": "test-nginx-redis"}
	statusPayloadJSON, _ := json.Marshal(statusPayload)

	statusCmd := Command{
		ID:      fmt.Sprintf("status-%d", time.Now().Unix()),
		Type:    "STATUS",
		Payload: statusPayloadJSON,
	}
	if err := conn.WriteJSON(statusCmd); err != nil {
		log.Printf("Failed to send status command: %v", err)
		return
	}

	// Keep connection alive
	for {
		time.Sleep(30 * time.Second)
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			log.Printf("Ping failed: %v", err)
			return
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	log.Println("Test WebSocket server starting on :8080")
	log.Println("Waiting for agent connections at ws://localhost:8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

