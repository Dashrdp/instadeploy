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

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded: ServerURL=%s, Version=%s, Architecture=%s",
		config.ServerURL, config.AgentVersion, config.AgentArchitecture)

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

