# Dev Session Template Helm Chart

This Helm chart deploys individual dev session environments for the PayPilot no-code app generator platform.

## Overview

Each dev session includes:
- A dedicated Kubernetes namespace (using project UUID)
- A single pod with three exposed services:
  - **Preview Service** - Application preview endpoint
  - **Chat Service** - AI agents communication endpoint  
  - **VS Code Service** - Web-based code editor endpoint
- Path-based routing via Ingress
- Persistent storage for workspace files
- Resource limits for CPU and memory

## Installation

```bash
# Install a dev session for a specific project
helm install dev-session-<project-uuid> ./helm/dev-session-template \
  --set project.uuid=<project-uuid> \
  --set project.id=<project-id> \
  --set user.id=<user-id> \
  --create-namespace
```

## Configuration

Key configuration values:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `project.uuid` | Project UUID (used as namespace) | `""` |
| `project.id` | Project ID | `0` |
| `user.id` | User ID | `0` |
| `service.preview.port` | Preview service port | `3000` |
| `service.preview.path` | Preview path for ingress | `/preview` |
| `service.chat.port` | Chat service port | `3001` |
| `service.chat.path` | Chat path for ingress | `/chat` |
| `service.vscode.port` | VS Code service port | `8080` |
| `service.vscode.path` | VS Code path for ingress | `/vscode` |
| `storage.enabled` | Enable persistent storage | `true` |
| `storage.size` | Storage size | `10Gi` |
| `resources.limits.cpu` | CPU limit | `2000m` |
| `resources.limits.memory` | Memory limit | `4Gi` |

## Services

The chart creates multiple ClusterIP services:

1. **LoadBalancer Service** (`-lb`) - Main entry point
2. **Preview Service** (`-preview`) - Application preview on port 3000
3. **Chat Service** (`-chat`) - AI chat endpoint on port 3001
4. **VS Code Service** (`-vscode`) - Web editor on port 8080

## Path-Based Routing

When ingress is enabled, services are accessible via paths:
- `https://<host>/preview` → Preview Service
- `https://<host>/chat` → Chat Service
- `https://<host>/vscode` → VS Code Service

## Uninstallation

```bash
helm uninstall dev-session-<project-uuid> -n <project-uuid>
kubectl delete namespace <project-uuid>
```
