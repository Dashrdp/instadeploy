package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

// CheckDockerInstalled checks if Docker is installed and available
func CheckDockerInstalled() bool {
	log.Println("CheckDockerInstalled: Checking if Docker is installed...")
	
	cmd := exec.Command("docker", "--version")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		log.Printf("CheckDockerInstalled: Docker not found: %v", err)
		return false
	}
	
	log.Printf("CheckDockerInstalled: Docker found: %s", strings.TrimSpace(string(output)))
	
	// Also check docker compose
	cmd = exec.Command("docker", "compose", "version")
	output, err = cmd.CombinedOutput()
	
	if err != nil {
		log.Printf("CheckDockerInstalled: Docker Compose not found: %v", err)
		return false
	}
	
	log.Printf("CheckDockerInstalled: Docker Compose found: %s", strings.TrimSpace(string(output)))
	return true
}

// InstallDocker attempts to install Docker on the system
func InstallDocker() error {
	log.Println("InstallDocker: Starting Docker installation...")
	
	// Check OS - only support Linux for now
	if runtime.GOOS != "linux" {
		return fmt.Errorf("automatic Docker installation is only supported on Linux (detected: %s)", runtime.GOOS)
	}
	
	// Detect if we're on Ubuntu/Debian (check for apt-get)
	if !commandExists("apt-get") {
		return fmt.Errorf("apt-get not found - automatic installation only supports Ubuntu/Debian")
	}
	
	log.Println("InstallDocker: Detected Debian/Ubuntu system")
	
	// Step 1: Update apt package index
	log.Println("InstallDocker: Step 1/7 - Updating apt package index...")
	if err := runCommand("apt-get", "update", "-y"); err != nil {
		return fmt.Errorf("failed to update apt: %w", err)
	}
	
	// Step 2: Install prerequisites
	log.Println("InstallDocker: Step 2/7 - Installing prerequisites...")
	if err := runCommand("apt-get", "install", "-y", "ca-certificates", "curl", "gnupg"); err != nil {
		return fmt.Errorf("failed to install prerequisites: %w", err)
	}
	
	// Step 3: Create keyrings directory
	log.Println("InstallDocker: Step 3/7 - Creating keyrings directory...")
	if err := runCommand("install", "-m", "0755", "-d", "/etc/apt/keyrings"); err != nil {
		log.Printf("InstallDocker: Warning - failed to create keyrings directory (may already exist): %v", err)
	}
	
	// Step 4: Download Docker's GPG key
	log.Println("InstallDocker: Step 4/7 - Downloading Docker GPG key...")
	cmd := exec.Command("sh", "-c", "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("InstallDocker: GPG key download output: %s", string(output))
		return fmt.Errorf("failed to download Docker GPG key: %w", err)
	}
	
	// Set permissions on the GPG key
	if err := runCommand("chmod", "a+r", "/etc/apt/keyrings/docker.gpg"); err != nil {
		log.Printf("InstallDocker: Warning - failed to set GPG key permissions: %v", err)
	}
	
	// Step 5: Add Docker repository
	log.Println("InstallDocker: Step 5/7 - Adding Docker repository...")
	repoCmd := `echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo $VERSION_CODENAME) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null`
	cmd = exec.Command("sh", "-c", repoCmd)
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("InstallDocker: Repository setup output: %s", string(output))
		return fmt.Errorf("failed to add Docker repository: %w", err)
	}
	
	// Step 6: Update apt again
	log.Println("InstallDocker: Step 6/7 - Updating apt with Docker repository...")
	if err := runCommand("apt-get", "update", "-y"); err != nil {
		return fmt.Errorf("failed to update apt after adding Docker repo: %w", err)
	}
	
	// Step 7: Install Docker packages
	log.Println("InstallDocker: Step 7/7 - Installing Docker packages (this may take a few minutes)...")
	if err := runCommand("apt-get", "install", "-y",
		"docker-ce",
		"docker-ce-cli",
		"containerd.io",
		"docker-buildx-plugin",
		"docker-compose-plugin"); err != nil {
		return fmt.Errorf("failed to install Docker packages: %w", err)
	}
	
	// Start and enable Docker service
	log.Println("InstallDocker: Starting Docker service...")
	if err := runCommand("systemctl", "start", "docker"); err != nil {
		log.Printf("InstallDocker: Warning - failed to start Docker service: %v", err)
	}
	
	if err := runCommand("systemctl", "enable", "docker"); err != nil {
		log.Printf("InstallDocker: Warning - failed to enable Docker service: %v", err)
	}
	
	log.Println("InstallDocker: Docker installation completed successfully!")
	
	// Verify installation
	if CheckDockerInstalled() {
		log.Println("InstallDocker: Docker verification successful!")
		return nil
	}
	
	return fmt.Errorf("Docker installation completed but verification failed")
}

// commandExists checks if a command is available in PATH
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// runCommand executes a command and logs its output
func runCommand(name string, args ...string) error {
	log.Printf("runCommand: Executing: %s %v", name, args)
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	
	if len(output) > 0 {
		log.Printf("runCommand: Output (%d bytes):\n%s", len(output), string(output))
	}
	
	if err != nil {
		log.Printf("runCommand: Command failed: %v", err)
		return err
	}
	
	return nil
}

// UpdateSystem performs system package updates (apt-get update && apt-get upgrade)
func UpdateSystem() error {
	log.Println("UpdateSystem: Starting system package update...")
	
	// Check OS - only support Linux for now
	if runtime.GOOS != "linux" {
		log.Printf("UpdateSystem: Skipping - only supported on Linux (detected: %s)", runtime.GOOS)
		return nil
	}
	
	// Check if apt-get is available
	if !commandExists("apt-get") {
		log.Println("UpdateSystem: Skipping - apt-get not found (only supports Ubuntu/Debian)")
		return nil
	}
	
	log.Println("UpdateSystem: Step 1/2 - Running apt-get update...")
	if err := runCommand("apt-get", "update", "-y"); err != nil {
		return fmt.Errorf("apt-get update failed: %w", err)
	}
	
	log.Println("UpdateSystem: Step 2/2 - Running apt-get upgrade (this may take several minutes)...")
	if err := runCommand("apt-get", "upgrade", "-y"); err != nil {
		return fmt.Errorf("apt-get upgrade failed: %w", err)
	}
	
	log.Println("UpdateSystem: System packages updated successfully!")
	return nil
}

// EnsureDockerInstalled checks if Docker is installed and attempts installation if not
func EnsureDockerInstalled() error {
	if CheckDockerInstalled() {
		return nil
	}
	
	log.Println("EnsureDockerInstalled: Docker not found, attempting automatic installation...")
	
	if err := InstallDocker(); err != nil {
		return fmt.Errorf("failed to install Docker: %w", err)
	}
	
	return nil
}

