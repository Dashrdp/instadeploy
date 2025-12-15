.PHONY: build build-agent clean run test help

# Variables
BINARY_NAME=agent
AGENT_DIR=./agent
BUILD_DIR=./bin
GO=go

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-trimpath

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: build-agent ## Build all binaries

build-agent: ## Build the agent binary
	@echo "Building agent..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(AGENT_DIR)
	@echo "Agent binary built at $(BUILD_DIR)/$(BINARY_NAME)"

build-linux: ## Build agent for Linux amd64
	@echo "Building agent for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(AGENT_DIR)
	@echo "Linux agent binary built at $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

build-obfuscated: ## Build obfuscated agent binary (requires garble)
	@echo "Building obfuscated agent..."
	@mkdir -p $(BUILD_DIR)
	@if command -v garble >/dev/null 2>&1; then \
		garble -literals -tiny build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-obfuscated $(AGENT_DIR); \
		echo "Obfuscated agent binary built at $(BUILD_DIR)/$(BINARY_NAME)-obfuscated"; \
	else \
		echo "Error: garble is not installed. Install it with: go install mvdan.cc/garble@latest"; \
		exit 1; \
	fi

run: build-agent ## Build and run the agent
	@echo "Running agent..."
	@echo "Make sure to set SERVER_URL, AGENT_TOKEN, and AGENT_SECRET_TOKEN environment variables"
	$(BUILD_DIR)/$(BINARY_NAME)

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf ./projects
	@echo "Clean complete"

test: ## Run tests
	$(GO) test -v ./...

fmt: ## Format Go code
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

lint: fmt vet ## Run formatting and vetting

deps: ## Download dependencies
	$(GO) mod download
	$(GO) mod tidy

install-garble: ## Install garble for obfuscated builds
	$(GO) install mvdan.cc/garble@latest

