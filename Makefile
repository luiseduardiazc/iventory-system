.PHONY: help build run test clean docker-up docker-down lint deps

help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application"
	@echo "  test        - Run tests"
	@echo "  docker-up   - Start Docker services"
	@echo "  docker-down - Stop Docker services"
	@echo "  lint        - Run linters"
	@echo "  clean       - Clean build artifacts"
	@echo "  deps        - Install dependencies"

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed!"

build:
	@echo "Building API..."
	@go build -o bin/api cmd/api/main.go
	@echo "Build complete!"

run: build
	@echo "Starting API server..."
	@./bin/api

test:
	@echo "Running tests..."
	@go test -v -race ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	@echo "Running linters..."
	@golangci-lint run ./...

docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d
	@echo "Waiting for services to be healthy..."
	@sleep 5
	@docker-compose ps

docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down

docker-logs:
	@docker-compose logs -f

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

.DEFAULT_GOAL := help
