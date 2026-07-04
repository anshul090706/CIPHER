#!/usr/bin/env bash

# Basic build script for the CIPHER project
# Usage: ./scripts/build.sh

set -e

echo "==> Building CIPHER binaries..."

# Create bin directory if it doesn't exist
mkdir -p bin

# Build all binaries in the cmd directory
go build -o bin/ ./cmd/...

echo "==> Build complete! Binaries are located in the bin/ directory."
