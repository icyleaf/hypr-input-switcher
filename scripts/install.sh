#!/bin/bash

# Install dependencies
go mod tidy

# Build the application
go build -o hypr-smart-input ./cmd/hypr-smart-input

# Install the application
sudo mv hypr-smart-input /usr/local/bin/

# Print success message
echo "Hypr Smart Input has been installed successfully!"