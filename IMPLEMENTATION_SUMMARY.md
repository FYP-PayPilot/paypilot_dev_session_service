# Implementation Summary: Always-On Dev Sessions with Helm Template

## Overview

This implementation introduces always-on dev sessions with Kubernetes integration. Each project now automatically has a dedicated dev session deployed via Helm charts with three exposed services: preview, chat, and VS Code.

## Changes Made

### 1. Session Model Updates (`internal/models/session.go`)

**Added Fields:**
- `ProjectUUID` - Unique identifier for the project (used as K8s namespace)
- `PreviewURL` & `PreviewPath` - Application preview service endpoints
- `ChatURL` & `ChatPath` - AI chat service endpoints  
- `VscodeURL` & `VscodePath` - VS Code web editor endpoints

**Key Changes:**
- Made `ProjectUUID` a unique, required field for identifying projects across services
- Namespace now uses `ProjectUUID` instead of generic naming
- Sessions expire after 1 year (always-on) instead of 24 hours

### 2. Helm Chart Template (`helm/dev-session-template/`)

Created a complete Helm chart for deploying dev session environments:

**Chart Structure:**
```
helm/dev-session-template/
├── Chart.yaml                 # Chart metadata
├── values.yaml                # Default configuration values
├── templates/
│   ├── _helpers.tpl          # Template helpers
│   ├── namespace.yaml         # K8s namespace (project UUID)
│   ├── deployment.yaml        # Main pod with 3 containers
│   ├── service.yaml          # ClusterIP services for each endpoint
│   ├── pvc.yaml              # Persistent storage
│   └── ingress.yaml          # Path-based routing
└── README.md                 # Documentation
```

**Services Deployed:**
1. **Preview Service** - Port 3000, Path `/preview`
   - Application preview and chat endpoint (same container)
2. **Chat Service** - Port 3001, Path `/chat`
   - AI agents communication endpoint
3. **VS Code Service** - Port 8080, Path `/vscode`
   - Web-based code editor

**Key Features:**
- Single namespace per project (using project UUID)
- ClusterIP load balancer service
- Path-based routing via Ingress
- Persistent storage (10Gi default)
- Resource limits (CPU: 2000m, Memory: 4Gi)

### 3. Kubernetes Client (`internal/kubernetes/client.go`)

Implemented Helm-based Kubernetes integration:

**New Methods:**
- `NewClient(log, helmChartPath)` - Initialize K8s client with Helm chart path
- `CreateDevContainer(ctx, projectUUID, projectID, userID)` - Deploy dev session via Helm
- `GetServiceEndpoints(ctx, namespace, releaseName)` - Retrieve service IPs and paths
- `DeleteDevContainer(ctx, projectUUID)` - Uninstall Helm release
- `GetContainerStatus(ctx, namespace, releaseName)` - Check deployment status
- `UpdateContainer(ctx, projectUUID, projectID, userID)` - Update running deployment

**Implementation Details:**
- Uses `helm install` command for deployments
- Executes `kubectl` to fetch service ClusterIPs
- Returns structured `ServiceEndpoints` with URLs and paths
- Automatic namespace creation via `--create-namespace` flag

### 4. Session Handlers (`internal/handlers/session.go`)

**New Endpoint:**
```
GET /api/v1/sessions/project/:project_uuid
```

**Get-or-Create Logic:**
1. Frontend requests session by project UUID
2. If session exists → return existing session
3. If session doesn't exist → create new session:
   - Deploy Helm chart to K8s
   - Populate service endpoints
   - Save to database
   - Return new session

**Updated Endpoints:**
- `POST /sessions` - Now creates K8s resources and populates endpoints
- `DELETE /sessions/:id` - Now deletes Helm releases

### 5. API Integration (`cmd/api/main.go`)

**Changes:**
- Initialize Kubernetes client with Helm chart path
- Pass K8s client to session handler
- Register new get-or-create endpoint
- Graceful degradation if K8s client fails (logs warning, continues without K8s)

### 6. Tests (`internal/kubernetes/client_test.go`)

Added unit tests for:
- Client initialization with custom/default paths
- ServiceEndpoints structure validation

## API Usage Examples

### Get or Create Session by Project UUID

```bash
# Frontend calls this endpoint with project UUID
curl -X GET "http://localhost:8080/api/v1/sessions/project/550e8400-e29b-41d4-a716-446655440000?user_id=123&project_id=456"

# Response includes service endpoints:
{
  "id": 1,
  "project_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "namespace": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "ip_address": "10.96.0.1",
  "preview_url": "http://10.96.0.1/preview",
  "preview_path": "/preview",
  "chat_url": "http://10.96.0.1/chat",
  "chat_path": "/chat",
  "vscode_url": "http://10.96.0.1/vscode",
  "vscode_path": "/vscode"
}
```

### Create Session Manually

```bash
curl -X POST "http://localhost:8080/api/v1/sessions" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "project_id": 456,
    "project_uuid": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

## Helm Deployment Example

Manual Helm deployment (for testing):

```bash
helm install dev-session-550e8400 ./helm/dev-session-template \
  --set project.uuid=550e8400-e29b-41d4-a716-446655440000 \
  --set project.id=456 \
  --set user.id=123 \
  --create-namespace
```

Verify deployment:

```bash
kubectl get pods -n 550e8400-e29b-41d4-a716-446655440000
kubectl get services -n 550e8400-e29b-41d4-a716-446655440000
```

## Architecture

```
┌─────────────────────────────────────────────┐
│        Frontend (Project UUID)              │
└──────────────────┬──────────────────────────┘
                   │ GET /sessions/project/{uuid}
                   ↓
┌─────────────────────────────────────────────┐
│     Dev Session Service (This Service)      │
│  - Check if session exists for project_uuid │
│  - If not, create via Helm                  │
│  - Return session with service endpoints    │
└──────────────────┬──────────────────────────┘
                   │ helm install/kubectl get
                   ↓
┌─────────────────────────────────────────────┐
│         Kubernetes Cluster                  │
│                                             │
│  Namespace: {project-uuid}                  │
│  ┌───────────────────────────────────────┐ │
│  │  Pod: dev-session-{uuid}              │ │
│  │    - Preview container (port 3000)    │ │
│  │    - Chat container (port 3001)       │ │
│  │    - VS Code container (port 8080)    │ │
│  └───────────────────────────────────────┘ │
│  ┌───────────────────────────────────────┐ │
│  │  Services (ClusterIP)                 │ │
│  │    - LoadBalancer service             │ │
│  │    - Preview service                  │ │
│  │    - Chat service                     │ │
│  │    - VS Code service                  │ │
│  └───────────────────────────────────────┘ │
│  ┌───────────────────────────────────────┐ │
│  │  Ingress (Path-based routing)         │ │
│  │    /preview → Preview service         │ │
│  │    /chat → Chat service               │ │
│  │    /vscode → VS Code service          │ │
│  └───────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
```

## Database Schema Changes

The `sessions` table now includes:

```sql
ALTER TABLE sessions ADD COLUMN project_uuid VARCHAR UNIQUE NOT NULL;
ALTER TABLE sessions ADD COLUMN preview_url VARCHAR;
ALTER TABLE sessions ADD COLUMN preview_path VARCHAR;
ALTER TABLE sessions ADD COLUMN chat_url VARCHAR;
ALTER TABLE sessions ADD COLUMN chat_path VARCHAR;
ALTER TABLE sessions ADD COLUMN vscode_url VARCHAR;
ALTER TABLE sessions ADD COLUMN vscode_path VARCHAR;
```

GORM AutoMigrate will handle this automatically on application startup.

## Configuration

No additional configuration is required. The service will:
- Use default Helm chart path: `./helm/dev-session-template`
- Deploy to K8s using project UUID as namespace
- Store service endpoints in the database automatically

## Testing

Run tests:
```bash
make test
```

Build application:
```bash
make build
```

Run locally (without K8s):
```bash
make run
# Service will start, K8s integration will log a warning but continue
```

## Future Enhancements

Potential improvements:
1. Add health checks to monitor pod status
2. Implement auto-scaling based on resource usage
3. Add metrics collection for service endpoints
4. Support custom Helm values per project
5. Implement session cleanup/garbage collection
6. Add authentication for service endpoints
7. Support multiple environments (dev, staging, prod)

## Security Considerations

- Project UUIDs are used as namespaces for isolation
- Service endpoints are currently internal (ClusterIP)
- Ingress should be configured with TLS in production
- Consider adding authentication to service endpoints
- Resource limits prevent resource exhaustion

## Troubleshooting

**Session creation fails:**
- Check Helm is installed: `helm version`
- Check kubectl access: `kubectl cluster-info`
- Check Helm chart path exists: `ls ./helm/dev-session-template`

**Services not accessible:**
- Check pod status: `kubectl get pods -n {project-uuid}`
- Check service endpoints: `kubectl get svc -n {project-uuid}`
- Check logs: `kubectl logs -n {project-uuid} {pod-name}`

**Database migration issues:**
- Ensure PostgreSQL is running and accessible
- Check logs for GORM migration errors
- Manually verify schema: `\d sessions` in psql
