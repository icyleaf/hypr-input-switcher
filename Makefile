# Makefile

# Makefile for Hypr Input Switcher Project

.PHONY: all build install clean

# Variables
BINARY_NAME=hypr-input-switcher
SRC_DIR=cmd/hypr-input-switcher/main.go
BUILD_DIR=bin

# Default target
all: build

# Build the application
build:
	@echo "Building the application..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)

# Install the application
install: build
	@echo "Installing the application..."
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
