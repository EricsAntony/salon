#!/bin/bash

# Production build script for salon services
# This ensures salon-shared is available during build

set -e

echo "Building salon services for production..."

# Build salon-service
echo "Building salon-service..."
cd salon-service
go mod tidy
go build -o ../bin/salon-service ./cmd
cd ..

# Build user-service  
echo "Building user-service..."
cd user-service
go mod tidy
go build -o ../bin/user-service ./cmd
cd ..

echo "Build complete! Binaries in ./bin/"
ls -la bin/
