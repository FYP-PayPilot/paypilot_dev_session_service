# GitHub Actions Workflows

This directory contains GitHub Actions workflows for continuous integration and deployment.

## Workflows

### CI Pipeline (`ci.yml`)

**Triggers:**
- Push to `main`, `develop`, or any `copilot/*` branch
- Pull requests to `main` or `develop`

**Jobs:**

1. **Lint** - Code quality checks
   - Verifies code formatting with `go fmt`
   - Runs `go vet` for suspicious constructs
   - Runs `staticcheck` for advanced static analysis

2. **Test** - Unit and integration tests
   - Runs all tests with race detection (`-race`)
   - Generates code coverage reports
   - Uploads coverage to Codecov

3. **Build** - Build verification
   - Compiles the application
   - Uploads binary artifacts (7-day retention)

4. **Docker** - Container build
   - Builds Docker image
   - Uses GitHub Actions cache for faster builds

5. **Security** - Security scanning
   - Runs Gosec security scanner
   - Uploads SARIF results to GitHub Security

6. **Integration** - End-to-end tests
   - Spins up PostgreSQL and RabbitMQ services
   - Runs integration tests against real services

### CodeQL Analysis (`codeql.yml`)

**Triggers:**
- Push to `main` or `develop`
- Pull requests to `main` or `develop`
- Weekly schedule (Monday at midnight)

**Purpose:**
- Deep security and quality analysis
- Identifies potential vulnerabilities
- Results available in GitHub Security tab

### Dependency Review (`dependency-review.yml`)

**Triggers:**
- Pull requests to `main` or `develop`

**Purpose:**
- Reviews dependency changes for security issues
- Fails on high-severity vulnerabilities
- Comments summary on pull requests

### Release Pipeline (`release.yml`)

**Triggers:**
- Version tags (e.g., `v1.0.0`)

**Jobs:**

1. **Release** - Multi-platform builds
   - Builds binaries for Linux (amd64, arm64)
   - Builds binaries for macOS (amd64, arm64)
   - Builds binaries for Windows (amd64)
   - Creates GitHub release with binaries
   - Auto-generates release notes

2. **Docker Release** - Container publishing
   - Builds multi-platform Docker images (amd64, arm64)
   - Pushes to Docker Hub with semantic versioning
   - Tags: `latest`, `v1.0.0`, `v1.0`, `v1`, SHA

## Configuration

### Required Secrets

For Docker publishing on releases:
- `DOCKER_USERNAME` - Docker Hub username
- `DOCKER_PASSWORD` - Docker Hub password or access token

### Optional Secrets

For Codecov integration:
- `CODECOV_TOKEN` - Codecov upload token (not required for public repos)

## Running Locally

Most CI checks can be run locally:

```bash
# Linting
go fmt ./...
go vet ./...

# Tests
go test -v -race ./...

# Build
go build -o bin/api ./cmd/api

# Docker
docker build -t paypilot-dev-session-service .
```

## Maintenance

### Updating Dependencies

GitHub Actions versions are specified in the workflows. To update:

1. Check for new versions at the action's repository
2. Update the version in the workflow file
3. Test the changes in a pull request

### Adding New Jobs

When adding new jobs:

1. Define the job in the appropriate workflow
2. Add dependencies with `needs:` if required
3. Test in a pull request before merging
4. Update this README with the new job description

## Monitoring

### Status Badges

Status badges are included in the main README:
- CI Pipeline status
- CodeQL analysis status
- Go Report Card grade

### Notifications

GitHub Actions will:
- Email workflow failures to the commit author
- Display status on pull requests
- Post status checks on commits

### Viewing Results

- **Workflow runs**: Actions tab in GitHub repository
- **Security findings**: Security tab → Code scanning alerts
- **Test coverage**: Codecov dashboard (if configured)
