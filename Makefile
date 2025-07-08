.PHONY: all build clean test fmt lint run install

# Variables
BINARY_NAME := battop
GO := go
MAIN_PATH := cmd/battop/main.go
BUILD_DIR := build
VERSION := 0.3.0
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "development")
DATE := $(shell date -u +"%Y-%m-%d")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build for release (with optimizations)
release:
	@echo "Building release version..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -trimpath -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run the application
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run with verbose logging
run-verbose: build
	./$(BUILD_DIR)/$(BINARY_NAME) -verbose

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS) $(MAIN_PATH)

# Update dependencies
deps:
	@echo "Updating dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Cross-compilation targets
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

build-all: build-linux build-darwin

# Development helpers
dev:
	@echo "Starting development mode with auto-reload..."
	@which air > /dev/null || (echo "air not installed. Install with: go install github.com/cosmtrek/air@latest" && exit 1)
	air

# Show help
help:
	@echo "Available targets:"
	@echo "  make build      - Build the binary"
	@echo "  make release    - Build optimized release version"
	@echo "  make run        - Build and run the application"
	@echo "  make run-verbose- Build and run with verbose logging"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make test       - Run tests"
	@echo "  make fmt        - Format code"
	@echo "  make lint       - Run linter"
	@echo "  make install    - Install binary to GOPATH/bin"
	@echo "  make deps       - Update dependencies"
	@echo "  make build-all  - Build for all platforms"
	@echo "  make dev        - Start development mode with auto-reload"
	@echo "  make help       - Show this help message"