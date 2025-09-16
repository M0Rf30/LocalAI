#!/bin/bash

# Run the sherpa-onnx backend
# If the binary doesn't exist, try to build it first
if [ ! -f "./sherpa-onnx" ]; then
    echo "sherpa-onnx binary not found, building..."
    make build
fi

if [ -f "./sherpa-onnx" ]; then
    # Set LD_LIBRARY_PATH to use packaged libraries
    export LD_LIBRARY_PATH="$(dirname "$(realpath $0)")/lib:${LD_LIBRARY_PATH:-}"
    ./sherpa-onnx "$@"
else
    echo "Failed to build sherpa-onnx backend"
    exit 1
fi