package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("InstaDeploy Agent starting...")

	// Update system packages
	log.Println("Updating system packages...")
	if err := UpdateSystem(); err != nil {
		log.Printf("WARNING: System update failed: %v", err)
		log.Println("Agent will continue but system packages may be outdated")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded: ServerURL=%s, Version=%s, Architecture=%s",
		config.ServerURL, config.AgentVersion, config.AgentArchitecture)

	// Check and ensure Docker is installed
	log.Println("Checking Docker installation...")
	if err := EnsureDockerInstalled(); err != nil {
		log.Printf("WARNING: Docker check/installation failed: %v", err)
		log.Println("Agent will continue but deployments will fail without Docker")
	}

	// Create agent
	agent, err := NewAgent(config)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, exiting...")
		os.Exit(0)
	}()

	// Run agent (with automatic reconnection)
	agent.Run()
}

