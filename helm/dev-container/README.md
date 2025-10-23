# Dev Container Helm Chart

This Helm chart deploys development containers for the no-code app generator in Kubernetes.

## Overview

Each dev session gets its own containerized environment with:
- Isolated workspace with persistent storage
- Dedicated resources (CPU/Memory)
- Network isolation per namespace
- Health monitoring and auto-restart

## Installation

### Prerequisites

- Kubernetes 1.20+
- Helm 3.0+
- kubectl configured to access your cluster

### Install a dev container

```bash
helm install dev-session-123 ./helm/dev-container \
  --set project.id=456 \
  --set user.id=789 \
  --namespace project-456 \
  --create-namespace
```

### Upgrade a dev container

```bash
helm upgrade dev-session-123 ./helm/dev-container \
  --set project.id=456 \
  --set user.id=789 \
  --namespace project-456
```

### Uninstall a dev container

```bash
helm uninstall dev-session-123 --namespace project-456
```

## Configuration

The following table lists the configurable parameters and their default values:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Container image repository | `paypilot/dev-container` |
| `image.tag` | Container image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `8080` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `512Mi` |
| `resources.requests.cpu` | CPU request | `250m` |
| `resources.requests.memory` | Memory request | `256Mi` |
| `persistence.enabled` | Enable persistent storage | `true` |
| `persistence.size` | Size of persistent volume | `5Gi` |
| `persistence.mountPath` | Mount path for workspace | `/workspace` |
| `project.id` | Project ID | `""` |
| `user.id` | User ID | `""` |

## Usage Example

Deploy a dev container for project 123 and user 456:

```bash
# Create namespace for the project
kubectl create namespace project-123

# Install dev container
helm install dev-session-abc ./helm/dev-container \
  --set project.id=123 \
  --set user.id=456 \
  --set image.repository=myregistry/dev-container \
  --set image.tag=v1.0.0 \
  --set persistence.size=10Gi \
  --namespace project-123

# Check the deployment
kubectl get pods -n project-123

# Access the dev container
kubectl port-forward -n project-123 svc/dev-session-abc-dev-container 8080:8080
```

## Architecture

```
┌─────────────────────────────────────────┐
│           Kubernetes Cluster             │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │      Namespace: project-123         │ │
│  │                                     │ │
│  │  ┌──────────────────────────────┐  │ │
│  │  │   Dev Container Pod          │  │ │
│  │  │  - project-id: 123           │  │ │
│  │  │  - user-id: 456              │  │ │
│  │  │  - Port: 8080                │  │ │
│  │  │  - Volume: /workspace (5Gi)  │  │ │
│  │  └──────────────────────────────┘  │ │
│  │            ↓                        │ │
│  │  ┌──────────────────────────────┐  │ │
│  │  │   Service (ClusterIP)        │  │ │
│  │  └──────────────────────────────┘  │ │
│  │            ↓                        │ │
│  │  ┌──────────────────────────────┐  │ │
│  │  │   PVC (5Gi)                  │  │ │
│  │  └──────────────────────────────┘  │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

## Integration with Session Service

The dev session service automatically manages these Helm deployments:

1. **Create Session**: Deploys a new Helm release with project/user context
2. **Get Session**: Queries Kubernetes for pod status
3. **Delete Session**: Uninstalls the Helm release and cleans up resources

## Monitoring

The dev containers expose a health endpoint at `/health` for:
- Liveness probes (restarts on failure)
- Readiness probes (removes from service on failure)
- External health checks

## Security

- Containers run as non-root user (UID 1000)
- Capabilities are dropped by default
- Resource limits prevent resource exhaustion
- Network policies can be applied per namespace
