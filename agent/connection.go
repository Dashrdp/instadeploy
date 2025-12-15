package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"instadeploy/shared"
)

const (
	// Initial backoff duration
	initialBackoff = 1 * time.Second
	// Maximum backoff duration
	maxBackoff = 2 * time.Minute
	// Backoff multiplier
	backoffMultiplier = 2
	// Ping interval for keepalive
	pingInterval = 30 * time.Second
	// Pong timeout
	pongTimeout = 60 * time.Second
)

// Agent represents the agent connection and state
type Agent struct {
	config   *Config
	conn     *websocket.Conn
	hostname string
	handler  *CommandHandler
}

// NewAgent creates a new agent instance
func NewAgent(config *Config) (*Agent, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return &Agent{
		config:   config,
		hostname: hostname,
		handler:  NewCommandHandler(),
	}, nil
}

// Connect establishes a WebSocket connection to the control plane
func (a *Agent) Connect() error {
	headers := http.Header{}
	headers.Add("Authorization", fmt.Sprintf("Bearer %s", a.config.AgentSecretToken))
	headers.Add("X-Agent-Version", a.config.AgentVersion)
	headers.Add("X-Agent-Architecture", a.config.AgentArchitecture)
	headers.Add("X-Agent-Hostname", a.hostname)

	log.Printf("Connecting to %s...", a.config.ServerURL)
	
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(a.config.ServerURL, headers)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	a.conn = conn
	log.Println("Connected successfully!")
	
	return nil
}

// Run starts the agent with reconnection logic
func (a *Agent) Run() {
	backoff := initialBackoff

	for {
		err := a.Connect()
		if err != nil {
			log.Printf("Connection failed: %v. Retrying in %v...", err, backoff)
			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * backoffMultiplier)
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		// Reset backoff on successful connection
		backoff = initialBackoff

		// Start message handling
		err = a.handleMessages()
		if err != nil {
			log.Printf("Connection error: %v. Reconnecting...", err)
		}

		// Close the connection if it's still open
		if a.conn != nil {
			a.conn.Close()
			a.conn = nil
		}

		// Wait before reconnecting
		time.Sleep(backoff)
	}
}

// handleMessages processes incoming messages from the control plane
func (a *Agent) handleMessages() error {
	// Set up ping/pong for keepalive
	a.conn.SetReadDeadline(time.Now().Add(pongTimeout))
	a.conn.SetPongHandler(func(string) error {
		a.conn.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	// Start ping ticker
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	// Channel for errors
	errChan := make(chan error, 1)

	// Goroutine for sending pings
	go func() {
		for range ticker.C {
			if err := a.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				errChan <- fmt.Errorf("ping failed: %w", err)
				return
			}
		}
	}()

	// Main message loop
	for {
		select {
		case err := <-errChan:
			return err
		default:
			_, message, err := a.conn.ReadMessage()
			if err != nil {
				return fmt.Errorf("read error: %w", err)
			}

			// Process the message in a goroutine to avoid blocking
			go a.processCommand(message)
		}
	}
}

// processCommand processes a single command from the control plane
func (a *Agent) processCommand(message []byte) {
	var cmd shared.Command
	if err := json.Unmarshal(message, &cmd); err != nil {
		log.Printf("Failed to unmarshal command: %v", err)
		a.sendResponse(shared.Response{
			JobID:  "unknown",
			Status: shared.StatusFailed,
			Error:  fmt.Sprintf("Invalid command format: %v", err),
		})
		return
	}

	log.Printf("Received command: ID=%s, Type=%s", cmd.ID, cmd.Type)

	// Send initial in-progress response
	a.sendResponse(shared.Response{
		JobID:  cmd.ID,
		Status: shared.StatusInProgress,
		Logs:   "Command received, processing...",
	})

	// Execute the command
	response := a.handler.HandleCommand(cmd)
	
	// Send final response
	a.sendResponse(response)
}

// sendResponse sends a response back to the control plane
func (a *Agent) sendResponse(response shared.Response) {
	if a.conn == nil {
		log.Println("Cannot send response: connection is nil")
		return
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	err = a.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Printf("Failed to send response: %v", err)
	} else {
		log.Printf("Sent response: JobID=%s, Status=%s", response.JobID, response.Status)
	}
}

