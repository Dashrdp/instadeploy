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
	log.Printf("handleDeploy: Processing DEPLOY_COMPOSE command, JobID=%s", cmd.ID)
	
	var payload shared.DeployPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		log.Printf("handleDeploy: Failed to unmarshal payload: %v", err)
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  fmt.Sprintf("Invalid deploy payload: %v", err),
		}
	}

	log.Printf("handleDeploy: Deploying project: %s (base64 length: %d)", payload.ProjectName, len(payload.ComposeFileBase64))

	// Validate project name
	if err := h.deploymentManager.ValidateProjectName(payload.ProjectName); err != nil {
		log.Printf("handleDeploy: Project name validation failed: %v", err)
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  fmt.Sprintf("Invalid project name: %v", err),
		}
	}
	log.Printf("handleDeploy: Project name validation passed for '%s'", payload.ProjectName)

	// Deploy the compose file
	log.Printf("handleDeploy: Starting deployment process for '%s'", payload.ProjectName)
	output, err := h.deploymentManager.DeployCompose(payload.ProjectName, payload.ComposeFileBase64)
	if err != nil {
		log.Printf("handleDeploy: Deployment FAILED for '%s': %v", payload.ProjectName, err)
		log.Printf("handleDeploy: Error output:\n%s", output)
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  err.Error(),
			Logs:   output,
		}
	}

	log.Printf("handleDeploy: Deployment SUCCESSFUL for '%s'", payload.ProjectName)
	return shared.Response{
		JobID:  cmd.ID,
		Status: shared.StatusCompleted,
		Logs:   fmt.Sprintf("Project '%s' deployed successfully.\n\nOutput:\n%s", payload.ProjectName, output),
	}
}

// handleStop processes a STOP_COMPOSE command
func (h *CommandHandler) handleStop(cmd shared.Command) shared.Response {
	log.Printf("handleStop: Processing STOP_COMPOSE command, JobID=%s", cmd.ID)
	
	var payload shared.StopPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		log.Printf("handleStop: Failed to unmarshal payload: %v", err)
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  fmt.Sprintf("Invalid stop payload: %v", err),
		}
	}

	log.Printf("handleStop: Stopping project: %s", payload.ProjectName)

	// Stop the compose deployment
	output, err := h.deploymentManager.StopCompose(payload.ProjectName)
	if err != nil {
		log.Printf("handleStop: Stop FAILED for '%s': %v", payload.ProjectName, err)
		log.Printf("handleStop: Error output:\n%s", output)
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  err.Error(),
			Logs:   output,
		}
	}

	log.Printf("handleStop: Stop SUCCESSFUL for '%s'", payload.ProjectName)
	return shared.Response{
		JobID:  cmd.ID,
		Status: shared.StatusCompleted,
		Logs:   fmt.Sprintf("Project '%s' stopped successfully.\n\nOutput:\n%s", payload.ProjectName, output),
	}
}

// handleStatus processes a STATUS command
func (h *CommandHandler) handleStatus(cmd shared.Command) shared.Response {
	log.Printf("handleStatus: Processing STATUS command, JobID=%s", cmd.ID)
	
	var payload shared.StatusPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		log.Printf("handleStatus: Failed to unmarshal payload: %v", err)
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  fmt.Sprintf("Invalid status payload: %v", err),
		}
	}

	log.Printf("handleStatus: Getting status for project: %s", payload.ProjectName)

	// Get the deployment status
	output, err := h.deploymentManager.GetStatus(payload.ProjectName)
	if err != nil {
		log.Printf("handleStatus: Status check FAILED for '%s': %v", payload.ProjectName, err)
		log.Printf("handleStatus: Error output:\n%s", output)
		return shared.Response{
			JobID:  cmd.ID,
			Status: shared.StatusFailed,
			Error:  err.Error(),
			Logs:   output,
		}
	}

	log.Printf("handleStatus: Status check SUCCESSFUL for '%s'", payload.ProjectName)
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

