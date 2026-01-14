.PHONY: build test lint clean install run help

# Binary name
BINARY_NAME=kost
OUTPUT_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build parameters
BUILD_FLAGS=-v
LDFLAGS=-ldflags "-s -w"

all: test build

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(OUTPUT_DIR)
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME) ./cmd/kost

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	$(GOLINT) run ./...

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) -s -w .

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(OUTPUT_DIR)
	rm -rf out
	rm -f coverage.out coverage.html

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

install: build ## Install the binary
	@echo "Installing $(BINARY_NAME)..."
	cp $(OUTPUT_DIR)/$(BINARY_NAME) /usr/local/bin/

run: build ## Build and run
	@echo "Running $(BINARY_NAME)..."
	./$(OUTPUT_DIR)/$(BINARY_NAME)

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

.DEFAULT_GOAL := help
