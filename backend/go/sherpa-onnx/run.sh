#!/bin/bash

# Run the sherpa-onnx backend
# If the binary doesn't exist, try to build it first
if [ ! -f "./sherpa-onnx" ]; then
    echo "sherpa-onnx binary not found, building..."
    make build
fi

if [ -f "./sherpa-onnx" ]; then
    ./sherpa-onnx "$@"
else
    echo "Failed to build sherpa-onnx backend"
    exit 1
fi