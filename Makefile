# Makefile

# Makefile for Hypr Input Switcher Project

.PHONY: prepare-icons build install clean test release snapshot

# Build variables
BINARY_NAME=hypr-input-switcher
BUILD_DIR=build
CONFIG_DIR=configs

# Version information
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build flags
LDFLAGS = -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
all: build

# Prepare icons for embedding (used during development)
prepare-icons:
	@echo "Preparing icons for embedding..."
	@chmod +x scripts/prepare-icons.sh
	@./scripts/prepare-icons.sh

# Build the binary
build: prepare-icons
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/hypr-input-switcher

# Install the binary and config files
install: build
	@echo "Installing $(BINARY_NAME)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@sudo mkdir -p /usr/share/hypr-input-switcher
	@sudo cp $(CONFIG_DIR)/default.yaml /usr/share/hypr-input-switcher/
	@echo "Installation complete!"

# Install for development (local user)
install-dev: build
	@echo "Installing $(BINARY_NAME) for development..."
	@mkdir -p ~/.local/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) ~/.local/bin/
	@mkdir -p ~/.local/share/hypr-input-switcher
	@cp $(CONFIG_DIR)/default.yaml ~/.local/share/hypr-input-switcher/
	@echo "Development installation complete!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf dist/
	@rm -rf bin/
	@rm -rf internal/notification/icons/

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Run with development config
run-dev: clean build
	@echo "Running with development config..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --config=./$(CONFIG_DIR)/default.yaml --log-level=debug --watch

# Create a snapshot release (for testing)
snapshot:
	@echo "Creating snapshot release..."
	@goreleaser build --snapshot --clean

# Create a full release (requires goreleaser)
release:
	@echo "Creating release..."
	@goreleaser release --clean

# Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"

# Check embedded icons status
check-icons:
	@echo "Checking embedded icons..."
	@if [ -d "internal/notification/icons" ]; then \
		echo "Found $(shell ls internal/notification/icons | wc -l) icon files:"; \
		ls -la internal/notification/icons/; \
	else \
		echo "No icons directory found. Run 'make prepare-icons' first."; \
	fi

# Development build (including icons)
dev: prepare-icons
	@go run ./cmd/hypr-input-switcher --log-level debug
