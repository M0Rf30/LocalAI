#!/bin/bash

# Create the assets directory if it doesn't exist
mkdir -p backend-assets

# Check if the binary exists
if [ ! -f "sherpa-onnx" ]; then
    echo "Error: sherpa-onnx binary not found. Please build it first."
    exit 1
fi

# Copy the binary
cp sherpa-onnx backend-assets/

# Copy the library files if they exist
if [ -d "backend-assets/lib" ]; then
    # Create target directory if it doesn't exist
    mkdir -p /tmp/localai/backend-assets/lib/ 2>/dev/null || true
    # Copy library files
    cp -r backend-assets/lib/* /tmp/localai/backend-assets/lib/ 2>/dev/null || true
fi

echo "Packaging complete"