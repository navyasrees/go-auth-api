#!/bin/bash

echo "🚀 Starting Auth API..."

# Start Docker services
echo "🐳 Starting Docker services..."
docker-compose up -d

# Wait a moment for services to be ready
sleep 3

# Start the Go application
echo "🔥 Starting Go application..."
go run cmd/main.go 