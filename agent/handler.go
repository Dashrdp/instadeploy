package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"

	"instadeploy/shared"
)

// CommandHandler processes commands received from the control plane
type CommandHandler struct {
	deploymentManager *DeploymentManager
}

// NewCommandHandler creates a new command handler
func NewCommandHandler() *CommandHandler {
	return &CommandHandler{
		deploymentManager: NewDeploymentManager(),
	}
}

// HandleCommand processes a command and returns a response
func (h *CommandHandler) HandleCommand(cmd shared.Command) shared.Response {
	switch cmd.Type {
	case shared.CommandTypeDeploy:
		return h.handleDeploy(cmd)
	case shared.CommandTypeStop:
		return h.handleStop(cmd)
	case shared.CommandTypeStatus:
		return h.handleStatus(cmd)
	case shared.CommandTypeHealthCheck:
		return h.handleHealthCheck(cmd)
	default:
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  fmt.Sprintf("Unknown command type: %s", cmd.Type),
		}
	}
}

// handleDeploy processes a DEPLOY_COMPOSE command
func (h *CommandHandler) handleDeploy(cmd shared.Command) shared.Response {
	var payload shared.DeployPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  fmt.Sprintf("Invalid deploy payload: %v", err),
		}
	}

	log.Printf("Deploying project: %s", payload.ProjectName)

	// Validate project name
	if err := h.deploymentManager.ValidateProjectName(payload.ProjectName); err != nil {
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  fmt.Sprintf("Invalid project name: %v", err),
		}
	}

	// Deploy the compose file
	output, err := h.deploymentManager.DeployCompose(payload.ProjectName, payload.ComposeFileBase64)
	if err != nil {
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  err.Error(),
			Logs:   output,
		}
	}

	return shared.Response{
		JobID:  cmd.ID,
		Status: shared.StatusCompleted,
		Logs:   fmt.Sprintf("Project '%s' deployed successfully.\n\nOutput:\n%s", payload.ProjectName, output),
	}
}

// handleStop processes a STOP_COMPOSE command
func (h *CommandHandler) handleStop(cmd shared.Command) shared.Response {
	var payload shared.StopPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  fmt.Sprintf("Invalid stop payload: %v", err),
		}
	}

	log.Printf("Stopping project: %s", payload.ProjectName)

	// Stop the compose deployment
	output, err := h.deploymentManager.StopCompose(payload.ProjectName)
	if err != nil {
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  err.Error(),
			Logs:   output,
		}
	}

	return shared.Response{
		JobID:  cmd.ID,
		Status: shared.StatusCompleted,
		Logs:   fmt.Sprintf("Project '%s' stopped successfully.\n\nOutput:\n%s", payload.ProjectName, output),
	}
}

// handleStatus processes a STATUS command
func (h *CommandHandler) handleStatus(cmd shared.Command) shared.Response {
	var payload shared.StatusPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  fmt.Sprintf("Invalid status payload: %v", err),
		}
	}

	log.Printf("Getting status for project: %s", payload.ProjectName)

	// Get the deployment status
	output, err := h.deploymentManager.GetStatus(payload.ProjectName)
	if err != nil {
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  err.Error(),
			Logs:   output,
		}
	}

	return shared.Response{
		JobID:  cmd.ID,
		Status: shared.StatusCompleted,
		Logs:   fmt.Sprintf("Status for project '%s':\n\n%s", payload.ProjectName, output),
	}
}

// handleHealthCheck processes a HEALTH_CHECK command
func (h *CommandHandler) handleHealthCheck(cmd shared.Command) shared.Response {
	hostname, _ := os.Hostname()
	
	info := shared.AgentInfo{
		Version:      "1.0.0",
		Architecture: fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH),
		Hostname:     hostname,
	}

	data, _ := json.Marshal(info)

	return shared.Response{
		JobID:  cmd.ID,
		Status: shared.StatusCompleted,
		Logs:   "Agent is healthy and operational",
		Data:   json.RawMessage(data),
	}
}

