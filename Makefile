.PHONY: start dev build test clean docker-up docker-down

# Start the application with all services
start: docker-up
	@echo "🚀 Starting auth API..."
	KAFKA_BROKER=localhost:9092 go run cmd/main.go

# Development mode with hot reload (requires air)
dev: docker-up
	@echo "🔥 Starting in development mode with hot reload..."
	@if command -v air > /dev/null; then \
		echo "Using existing air installation"; \
	else \
		echo "Installing air for hot reload..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@export PATH="$$PATH:$$HOME/go/bin" && KAFKA_BROKER=localhost:9092 air

# Build the application
build:
	@echo "🔨 Building application..."
	go build -o bin/auth-api cmd/main.go

# Run tests
test:
	@echo "🧪 Running tests..."
	go test ./...

# Start Docker services
docker-up:
	@echo "🐳 Starting Docker services..."
	docker compose up -d

# Stop Docker services
docker-down:
	@echo "🛑 Stopping Docker services..."
	docker compose down

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod tidy
	go mod download

# Show logs
logs:
	@echo "📋 Showing logs..."
	docker compose logs -f

# Reset everything (stop services, clean, restart)
reset: docker-down clean docker-up
	@echo "�� Reset complete!" 