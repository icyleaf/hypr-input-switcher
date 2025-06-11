# Makefile

# Makefile for Hypr Input Switcher Project

.PHONY: build install clean test

# Build variables
BINARY_NAME=hypr-input-switcher
BUILD_DIR=bin
CONFIG_DIR=configs

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/hypr-input-switcher

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

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Run with development config
run-dev: build
	@echo "Running with development config..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --config=./$(CONFIG_DIR)/default.yaml --log-level=debug --watch
