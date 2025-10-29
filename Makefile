.PHONY: help build run test clean docker-build docker-up docker-down swagger install-tools

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Tools installed successfully"

build: ## Build the application
	@echo "Building application..."
	go build -o bin/api ./cmd/api
	@echo "Build complete: bin/api"

run: ## Run the application
	@echo "Running application..."
	go run ./cmd/api/main.go

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "Coverage report:"
	go tool cover -func=coverage.out

test-coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	swag init -g cmd/api/main.go -o docs
	@echo "Swagger documentation generated in docs/"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf coverage.out coverage.html
	rm -rf docs/
	@echo "Clean complete"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker compose build
	@echo "Docker image built"

docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	docker compose up -d
	@echo "Containers started. Access the app at http://localhost:8080"
	@echo "RabbitMQ Management UI: http://localhost:15672 (guest/guest)"

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	docker compose down
	@echo "Containers stopped"

docker-logs: ## View Docker logs
	docker compose logs -f

docker-restart: ## Restart Docker containers
	@echo "Restarting Docker containers..."
	docker compose restart
	@echo "Containers restarted"

lint: ## Run linters
	@echo "Running linters..."
	go fmt ./...
	go vet ./...
	@echo "Linting complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies updated"

migrate: ## Run database migrations (requires running database)
	@echo "Running migrations..."
	@echo "Migrations are run automatically on application startup"

.DEFAULT_GOAL := help
