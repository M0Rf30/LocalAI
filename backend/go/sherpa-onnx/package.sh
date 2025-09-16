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

# Copy sherpa-onnx libraries from Go module cache
# Find the Go module cache path for sherpa-onnx-go-linux
# First try with GOPATH, then fallback to default Go module cache
if [ -z "$GOPATH" ]; then
    GO_MOD_CACHE="$HOME/go/pkg/mod"
else
    GO_MOD_CACHE="$GOPATH/pkg/mod"
fi

SHERPA_LIB_DIR=$(find "$GO_MOD_CACHE"/github.com/k2-fsa/sherpa-onnx-go-linux* -name "libsherpa-onnx-c-api.so" -exec dirname {} \; 2>/dev/null | head -n 1)
if [ -n "$SHERPA_LIB_DIR" ]; then
    echo "Found sherpa-onnx libraries at: $SHERPA_LIB_DIR"
    mkdir -p $CURDIR/package/lib
    # Copy all sherpa-onnx libraries with read permissions
    cp -f "$SHERPA_LIB_DIR"/* $CURDIR/package/lib/
    chmod 644 $CURDIR/package/lib/*
else
    echo "Warning: sherpa-onnx libraries not found in Go module cache"
fi

# Copy run script
cp -rfv run.sh $CURDIR/package/

echo "Packaging complete"
ls -liah $CURDIR/package/