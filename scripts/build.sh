#!/bin/bash

# Build the Go application
echo "Building the Hypr Smart Input application..."

# Set the Go module path
export GO111MODULE=on

# Build the application
go build -o hypr-smart-input ./cmd/hypr-smart-input

# Check if the build was successful
if [ $? -eq 0 ]; then
    echo "Build successful! Executable created: hypr-smart-input"
else
    echo "Build failed. Please check the errors above."
    exit 1
fi