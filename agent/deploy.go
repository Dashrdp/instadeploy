package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// BaseProjectDir is the base directory for all projects
	// In production, this would be /opt/platform/projects
	// For development/testing, we use a local directory
	BaseProjectDir = "./projects"
)

// DeploymentManager handles Docker Compose deployments
type DeploymentManager struct {
	baseDir string
}

// NewDeploymentManager creates a new deployment manager
func NewDeploymentManager() *DeploymentManager {
	return &DeploymentManager{
		baseDir: BaseProjectDir,
	}
}

// ValidateProjectName validates that the project name is safe
func (dm *DeploymentManager) ValidateProjectName(projectName string) error {
	if projectName == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Only allow alphanumeric characters, hyphens, and underscores
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(projectName) {
		return fmt.Errorf("project name contains invalid characters (only alphanumeric, hyphens, and underscores allowed)")
	}

	// Prevent directory traversal
	if strings.Contains(projectName, "..") || strings.Contains(projectName, "/") || strings.Contains(projectName, "\\") {
		return fmt.Errorf("project name contains invalid path characters")
	}

	return nil
}

// GetProjectDir returns the directory for a specific project
func (dm *DeploymentManager) GetProjectDir(projectName string) (string, error) {
	if err := dm.ValidateProjectName(projectName); err != nil {
		return "", err
	}

	projectDir := filepath.Join(dm.baseDir, projectName)
	
	// Ensure the path is within baseDir (prevent directory traversal)
	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve project directory: %w", err)
	}

	absBaseDir, err := filepath.Abs(dm.baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base directory: %w", err)
	}

	if !strings.HasPrefix(absProjectDir, absBaseDir) {
		return "", fmt.Errorf("project directory is outside base directory")
	}

	return absProjectDir, nil
}

// EnsureProjectDir creates the project directory if it doesn't exist
func (dm *DeploymentManager) EnsureProjectDir(projectName string) (string, error) {
	projectDir, err := dm.GetProjectDir(projectName)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create project directory: %w", err)
	}

	return projectDir, nil
}

// WriteComposeFile writes the docker-compose.yml file to the project directory
func (dm *DeploymentManager) WriteComposeFile(projectName string, composeFileBase64 string) (string, error) {
	// Decode base64 content
	composeContent, err := base64.StdEncoding.DecodeString(composeFileBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode compose file: %w", err)
	}

	// Ensure project directory exists
	projectDir, err := dm.EnsureProjectDir(projectName)
	if err != nil {
		return "", err
	}

	// Write the file
	composeFilePath := filepath.Join(projectDir, "docker-compose.yml")
	if err := os.WriteFile(composeFilePath, composeContent, 0644); err != nil {
		return "", fmt.Errorf("failed to write compose file: %w", err)
	}

	return composeFilePath, nil
}

// DeployCompose deploys a docker-compose.yml file
func (dm *DeploymentManager) DeployCompose(projectName string, composeFileBase64 string) (string, error) {
	// Write the compose file
	composeFilePath, err := dm.WriteComposeFile(projectName, composeFileBase64)
	if err != nil {
		return "", err
	}

	projectDir := filepath.Dir(composeFilePath)

	// Execute docker compose up -d
	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("docker compose up failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// StopCompose stops a docker-compose deployment
func (dm *DeploymentManager) StopCompose(projectName string) (string, error) {
	projectDir, err := dm.GetProjectDir(projectName)
	if err != nil {
		return "", err
	}

	// Check if the project directory exists
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return "", fmt.Errorf("project does not exist")
	}

	// Check if docker-compose.yml exists
	composeFilePath := filepath.Join(projectDir, "docker-compose.yml")
	if _, err := os.Stat(composeFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("docker-compose.yml not found")
	}

	// Execute docker compose down
	cmd := exec.Command("docker", "compose", "down")
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("docker compose down failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// GetStatus gets the status of a docker-compose deployment
func (dm *DeploymentManager) GetStatus(projectName string) (string, error) {
	projectDir, err := dm.GetProjectDir(projectName)
	if err != nil {
		return "", err
	}

	// Check if the project directory exists
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return "Project does not exist", nil
	}

	// Check if docker-compose.yml exists
	composeFilePath := filepath.Join(projectDir, "docker-compose.yml")
	if _, err := os.Stat(composeFilePath); os.IsNotExist(err) {
		return "docker-compose.yml not found", nil
	}

	// Execute docker compose ps
	cmd := exec.Command("docker", "compose", "ps")
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("docker compose ps failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

