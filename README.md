# PayPilot Dev Session Service

[![CI Pipeline](https://github.com/villageFlower/paypilot_dev_session_service/actions/workflows/ci.yml/badge.svg)](https://github.com/villageFlower/paypilot_dev_session_service/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/villageFlower/paypilot_dev_session_service)](https://goreportcard.com/report/github.com/villageFlower/paypilot_dev_session_service)

A microservice for managing development sessions in a no-code app generator. This service manages containerized development environments in Kubernetes, keeping dev containers alive and listening for frontend changes.

## Overview

This service is part of a no-code application generator platform. It manages development sessions for projects by:
- Creating isolated dev containers in Kubernetes namespaces
- Managing session lifecycle (create, monitor, delete)
- Tracking active development sessions per project and user
- Using Helm charts to deploy and manage dev containers
- Integrating with other microservices for user and project management

## Tech Stack

- **Language**: Go 1.21+
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin)
- **Database**: PostgreSQL 15+
- **ORM**: [GORM](https://gorm.io/) with AutoMigrate
- **Message Queue**: RabbitMQ
- **Configuration**: [Viper](https://github.com/spf13/viper)
- **Logging**: [Zap](https://github.com/uber-go/zap)
- **API Documentation**: [Swagger](https://github.com/swaggo/swag) (swaggo/swag)
- **Testing**: Standard Go testing + [Testify](https://github.com/stretchr/testify)
- **Containerization**: Docker & Docker Compose

## Project Structure

```
.
├── cmd/
│   └── api/              # Application entrypoint
│       └── main.go
├── configs/              # Configuration files
│   └── config.yaml
├── docs/                 # Swagger documentation (generated)
├── internal/             # Private application code
│   ├── database/         # Database connection and migrations
│   ├── handlers/         # HTTP request handlers
│   ├── middleware/       # HTTP middleware
│   ├── messaging/        # RabbitMQ messaging
│   └── models/          # Data models
├── pkg/                  # Public library code
│   ├── config/          # Configuration management
│   └── logger/          # Logging utilities
├── .env.example          # Environment variables template
├── .gitignore
├── docker-compose.yml    # Docker Compose configuration
├── Dockerfile           # Docker image definition
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
├── Makefile            # Build automation
└── README.md
```

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose (for containerized setup)
- PostgreSQL 15+ (if running locally without Docker)
- RabbitMQ (if running locally without Docker)

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/villageFlower/paypilot_dev_session_service.git
cd paypilot_dev_session_service
```

### 2. Using Docker Compose (Recommended)

The easiest way to get started is using Docker Compose, which sets up PostgreSQL, RabbitMQ, and the application:

```bash
# Start all services
make docker-up

# View logs
make docker-logs

# Stop all services
make docker-down
```

The application will be available at:
- API: http://localhost:8080
- Swagger UI: http://localhost:8080/swagger/index.html
- RabbitMQ Management: http://localhost:15672 (guest/guest)

### 3. Running Locally

If you prefer to run the application locally:

#### Install dependencies

```bash
make deps
```

#### Configure environment

Copy the example environment file and modify as needed:

```bash
cp .env.example .env
```

Update the configuration in `configs/config.yaml` or set environment variables.

#### Start PostgreSQL and RabbitMQ

Ensure PostgreSQL and RabbitMQ are running and accessible with the configured credentials.

#### Run the application

```bash
make run
```

Or build and run:

```bash
make build
./bin/api
```

## Development

### Install development tools

```bash
make install-tools
```

This installs:
- `swag` for generating Swagger documentation

### Generate Swagger documentation

```bash
make swagger
```

### Run tests

```bash
make test
```

### Generate test coverage report

```bash
make test-coverage
```

This generates an HTML coverage report at `coverage.html`.

### Run linters

```bash
make lint
```

### Build the application

```bash
make build
```

## API Endpoints

### Health Check

- `GET /api/v1/health` - Health check endpoint

### Dev Sessions

- `POST /api/v1/sessions` - Create a new dev session (deploys a container)
- `GET /api/v1/sessions` - List all sessions with filtering (user_id, project_id, status)
- `GET /api/v1/sessions/:id` - Get a specific session
- `DELETE /api/v1/sessions/:id` - Delete a session (stops the container)

**Query Parameters for List:**
- `user_id` - Filter sessions by user ID
- `project_id` - Filter sessions by project ID  
- `status` - Filter by status (pending, running, stopped, error)
- `page` - Page number for pagination
- `page_size` - Number of items per page

### API Documentation

Access the interactive Swagger UI at: http://localhost:8080/swagger/index.html

## Configuration

Configuration can be provided via:

1. **Config file**: `configs/config.yaml`
2. **Environment variables**: Override any config value using uppercase with underscores (e.g., `SERVER_PORT`, `DB_HOST`)

### Key Configuration Options

```yaml
server:
  port: 8080              # Server port
  mode: debug             # Gin mode: debug, release, test

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: paypilot_dev
  sslmode: disable

rabbitmq:
  host: localhost
  port: 5672
  user: guest
  password: guest
  vhost: /
  exchange: paypilot_exchange
  queue: paypilot_queue

log:
  level: debug            # Log level: debug, info, warn, error
  encoding: json          # Log format: json, console
```

## Kubernetes & Helm Integration

This service manages dev containers in Kubernetes using Helm charts:

### Helm Chart

The `helm/dev-container` chart deploys isolated dev environments with:
- Dedicated namespace per project
- Persistent storage for project files
- Resource limits (CPU/Memory)
- Health monitoring and auto-restart
- Service exposure for frontend communication

### Deploy a Dev Container

```bash
# Install dev container for project 123, user 456
helm install dev-session-abc ./helm/dev-container \
  --set project.id=123 \
  --set user.id=456 \
  --namespace project-123 \
  --create-namespace
```

See [helm/dev-container/README.md](helm/dev-container/README.md) for full documentation.

### Architecture

```
┌──────────────────────────────────────────────────┐
│           Dev Session Service (This)              │
│  - Manages session records in PostgreSQL         │
│  - Creates/deletes Helm releases                 │
│  - Tracks container status                       │
└────────────────┬─────────────────────────────────┘
                 │
                 ↓ Helm Commands
┌──────────────────────────────────────────────────┐
│           Kubernetes Cluster                      │
│                                                   │
│  ┌────────────────────────────────────────────┐  │
│  │  Namespace: project-123                    │  │
│  │  - Dev Container Pod (user 456)            │  │
│  │  - PVC (5Gi workspace)                     │  │
│  │  - Service (ClusterIP)                     │  │
│  └────────────────────────────────────────────┘  │
│                                                   │
│  ┌────────────────────────────────────────────┐  │
│  │  Namespace: project-789                    │  │
│  │  - Dev Container Pod (user 101)            │  │
│  │  - PVC (5Gi workspace)                     │  │
│  │  - Service (ClusterIP)                     │  │
│  └────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────┘
```

## Database Migrations

Database migrations are handled automatically by GORM's AutoMigrate feature. When the application starts, it will:

1. Connect to the database
2. Automatically create or update tables based on the model definitions
3. Preserve existing data

The main model is `Session` which tracks:
- `user_id` - Integer reference to user (managed by another microservice)
- `project_id` - Integer reference to project (managed by another microservice)
- `container_name` - Name of the Kubernetes pod
- `namespace` - Kubernetes namespace
- `status` - Container status (pending, running, stopped, error)
- `token` - Unique session token
- `expires_at` - Session expiration time

## Testing

The project uses Go's standard testing framework with Testify for assertions.

### Run all tests

```bash
make test
```

### Run tests with coverage

```bash
make test-coverage
```

### Writing tests

Tests should be placed alongside the code they test with a `_test.go` suffix:

```
internal/
  models/
    user.go
    user_test.go
    session.go
    session_test.go
```

## Docker

### Build Docker image

```bash
make docker-build
```

### Start services with Docker Compose

```bash
make docker-up
```

### Stop services

```bash
make docker-down
```

### View logs

```bash
make docker-logs
```

## Makefile Commands

Run `make help` to see all available commands:

```bash
make help
```

Available commands:
- `build` - Build the application
- `run` - Run the application
- `test` - Run tests
- `test-coverage` - Generate test coverage report
- `swagger` - Generate Swagger documentation
- `clean` - Clean build artifacts
- `docker-build` - Build Docker image
- `docker-up` - Start Docker containers
- `docker-down` - Stop Docker containers
- `docker-logs` - View Docker logs
- `docker-restart` - Restart Docker containers
- `lint` - Run linters
- `deps` - Download dependencies
- `install-tools` - Install development tools

## CI/CD Pipeline

The repository includes comprehensive GitHub Actions workflows for continuous integration and deployment:

### CI Pipeline (`.github/workflows/ci.yml`)

Runs on every push and pull request:

1. **Lint** - Code formatting and static analysis
   - `go fmt` - Ensures code is properly formatted
   - `go vet` - Examines Go source code and reports suspicious constructs
   - `staticcheck` - Advanced static analysis

2. **Test** - Unit and integration tests
   - Runs all tests with race detection
   - Generates code coverage reports
   - Uploads coverage to Codecov

3. **Build** - Application build verification
   - Builds the application binary
   - Uploads build artifacts

4. **Docker** - Container build verification
   - Builds Docker image
   - Uses layer caching for faster builds

5. **Security** - Security scanning
   - Runs Gosec security scanner
   - Outputs security scan results

6. **Integration Tests** - End-to-end testing
   - Spins up PostgreSQL and RabbitMQ services
   - Runs integration tests against real services

### Release Pipeline (`.github/workflows/release.yml`)

Triggered when pushing version tags (e.g., `v1.0.0`):

1. **Multi-platform Builds**
   - Builds binaries for Linux (amd64, arm64)
   - Builds binaries for macOS (amd64, arm64)
   - Builds binaries for Windows (amd64)

2. **GitHub Release**
   - Creates GitHub release with binaries
   - Auto-generates release notes

3. **Docker Release**
   - Builds multi-platform Docker images
   - Pushes to Docker Hub with semantic versioning tags
   - Tags: `latest`, `v1.0.0`, `v1.0`, `v1`, and SHA

### Status Badges

The README includes status badges showing:
- CI Pipeline status
- Go Report Card grade

### Setting Up CI/CD

The workflows are ready to use. For Docker publishing on releases:

1. Add Docker Hub credentials to repository secrets:
   - `DOCKER_USERNAME` - Your Docker Hub username
   - `DOCKER_PASSWORD` - Your Docker Hub password or access token

2. Create a release:
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.

## Support

For support, please open an issue in the GitHub repository.