#!/bin/bash

CURDIR=$(dirname "$(realpath $0)")

# Create the package directory if it doesn't exist
mkdir -p $CURDIR/package

# Check if the binary exists
if [ ! -f "sherpa-onnx" ]; then
    echo "Error: sherpa-onnx binary not found. Please build it first."
    exit 1
fi

# Copy the binary to package directory
cp sherpa-onnx $CURDIR/package/

# Copy the library files if they exist
if [ -d "backend-assets/lib" ]; then
    mkdir -p $CURDIR/package/lib
    cp -r backend-assets/lib/* $CURDIR/package/lib/
fi

# Copy run script
cp -rfv run.sh $CURDIR/package/

echo "Packaging complete"
ls -liah $CURDIR/package/