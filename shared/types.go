package shared

import "encoding/json"

// CommandType represents the type of command sent from server to agent
type CommandType string

const (
	// CommandTypeDeploy instructs the agent to deploy a docker-compose file
	CommandTypeDeploy CommandType = "DEPLOY_COMPOSE"
	
	// CommandTypeStop instructs the agent to stop a deployment
	CommandTypeStop CommandType = "STOP_COMPOSE"
	
	// CommandTypeStatus requests the status of a deployment
	CommandTypeStatus CommandType = "STATUS"
	
	// CommandTypeHealthCheck requests a health check from the agent
	CommandTypeHealthCheck CommandType = "HEALTH_CHECK"
)

// ResponseStatus represents the status of a command execution
type ResponseStatus string

const (
	// StatusCompleted indicates successful completion
	StatusCompleted ResponseStatus = "COMPLETED"
	
	// StatusFailed indicates failure
	StatusFailed ResponseStatus = "FAILED"
	
	// StatusInProgress indicates ongoing execution
	StatusInProgress ResponseStatus = "IN_PROGRESS"
	
	// StatusQueued indicates the command is queued
	StatusQueued ResponseStatus = "QUEUED"
)

// Command represents a command sent from the control plane to the agent
type Command struct {
	// ID is a unique identifier for this command/job
	ID string `json:"id"`
	
	// Type specifies the command type
	Type CommandType `json:"type"`
	
	// Payload contains command-specific data
	Payload json.RawMessage `json:"payload"`
}

// DeployPayload represents the payload for a DEPLOY_COMPOSE command
type DeployPayload struct {
	// ProjectName is the unique name for this deployment
	ProjectName string `json:"project_name"`
	
	// ComposeFileBase64 is the base64-encoded docker-compose.yml content
	ComposeFileBase64 string `json:"compose_file_base64"`
}

// StopPayload represents the payload for a STOP_COMPOSE command
type StopPayload struct {
	// ProjectName is the name of the project to stop
	ProjectName string `json:"project_name"`
}

// StatusPayload represents the payload for a STATUS command
type StatusPayload struct {
	// ProjectName is the name of the project to check
	ProjectName string `json:"project_name"`
}

// Response represents a response sent from the agent to the control plane
type Response struct {
	// JobID is the ID of the command this response corresponds to
	JobID string `json:"job_id"`
	
	// Status indicates the execution status
	Status ResponseStatus `json:"status"`
	
	// Logs contains execution logs or output
	Logs string `json:"logs"`
	
	// Error contains error message if status is FAILED
	Error string `json:"error,omitempty"`
	
	// Data contains additional response data (optional)
	Data json.RawMessage `json:"data,omitempty"`
}

// AgentInfo represents information about the agent
type AgentInfo struct {
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	Hostname     string `json:"hostname"`
}

