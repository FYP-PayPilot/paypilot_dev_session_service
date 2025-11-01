# Deployment Guide - Dev Session Service

This guide explains how to deploy the Dev Session Service to Kubernetes on Digital Ocean using GitHub Actions and Helm.

## Prerequisites

- Digital Ocean Kubernetes Cluster
- GitHub repository with Actions enabled
- Docker registry (GitHub Container Registry is used by default)
- Helm 3.x installed locally (for manual deployments)
- kubectl configured locally (for manual deployments)

## GitHub Secrets Configuration

You need to configure the following secrets in your GitHub repository:

### Required Secrets

Navigate to: `Settings > Secrets and variables > Actions > New repository secret`

1. **DIGITALOCEAN_ACCESS_TOKEN**
   - Your Digital Ocean API token
   - Get it from: https://cloud.digitalocean.com/account/api/tokens
   - Required for: Accessing your Kubernetes cluster

2. **DO_CLUSTER_NAME**
   - Your Digital Ocean Kubernetes cluster name
   - Find it in: Digital Ocean Dashboard > Kubernetes
   - Example: `paypilot-k8s-cluster`

3. **DB_PASSWORD**
   - PostgreSQL database password
   - This will be used for the PostgreSQL instance in Kubernetes
   - Example: `your-secure-db-password-here`

4. **RABBITMQ_PASSWORD**
   - RabbitMQ password for the shared instance
   - Get this from your RabbitMQ admin in the default namespace
   - Default is often `guest` but use your actual password

5. **RABBITMQ_HOST** (Optional - has default)
   - RabbitMQ service host
   - Default: `rabbitmq.default.svc.cluster.local`
   - Only set if your RabbitMQ is at a different location

## CI/CD Pipeline Overview

The GitHub Actions workflow (`.github/workflows/deploy.yaml`) performs:

1. **Build Stage**
   - Checks out code
   - Builds Docker image with commit SHA tag
   - Pushes to GitHub Container Registry (ghcr.io)
   - Tags: `branch-name-{short-sha}`, `latest` (for main branch)

2. **Deploy Stage** (only on main/develop branches)
   - Connects to Digital Ocean Kubernetes cluster
   - Creates `dev-session-service` namespace
   - Creates Docker registry pull secret
   - Deploys/upgrades using Helm with the new image tag
   - Verifies deployment rollout

## Automatic Deployment

Deployments trigger automatically on:
- Push to `main` branch (production)
- Push to `develop` branch (staging)
- Manual workflow dispatch from GitHub Actions tab

## Manual Deployment

### 1. Configure kubectl for Digital Ocean

```bash
# Install doctl
# macOS: brew install doctl
# Linux: snap install doctl
# Windows: Download from GitHub releases

# Authenticate
doctl auth init

# Get cluster name
doctl kubernetes cluster list

# Configure kubectl
doctl kubernetes cluster kubeconfig save YOUR_CLUSTER_NAME
```

### 2. Create Namespace

```bash
kubectl create namespace dev-session-service
```

### 3. Create Secrets

```bash
# Create Docker registry secret (if using private registry)
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=YOUR_GITHUB_USERNAME \
  --docker-password=YOUR_GITHUB_TOKEN \
  --namespace=dev-session-service

# Database and RabbitMQ secrets are created by Helm
```

### 4. Deploy with Helm

```bash
# From repository root
helm upgrade --install dev-session-service ./helm/dev-session-service \
  --namespace dev-session-service \
  --create-namespace \
  --set image.repository=ghcr.io/YOUR_ORG/YOUR_REPO \
  --set image.tag=latest \
  --set secrets.dbPassword=YOUR_DB_PASSWORD \
  --set secrets.rabbitmqPassword=YOUR_RABBITMQ_PASSWORD \
  --set rabbitmq.host=rabbitmq.default.svc.cluster.local \
  --wait
```

## Helm Values Customization

Edit `helm/dev-session-service/values.yaml` to customize:

### Image Configuration
```yaml
image:
  repository: ghcr.io/your-org/dev-session-service
  tag: "latest"
  pullPolicy: IfNotPresent
```

### Resources
```yaml
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi
```

### Database Storage
```yaml
database:
  storage:
    size: 10Gi
    storageClassName: "do-block-storage"  # Digital Ocean specific
```

### Environment Variables
```yaml
env:
  SERVER_PORT: "8080"
  SERVER_MODE: "release"
  LOG_LEVEL: "info"
```

## Monitoring Deployment

### Check Deployment Status

```bash
# Watch deployment rollout
kubectl rollout status deployment/dev-session-service-dev-session-service -n dev-session-service

# Get all resources
kubectl get all -n dev-session-service

# Check pods
kubectl get pods -n dev-session-service

# View logs
kubectl logs -f deployment/dev-session-service-dev-session-service -n dev-session-service
```

### Debug Issues

```bash
# Describe pod for events
kubectl describe pod POD_NAME -n dev-session-service

# Check service endpoints
kubectl get endpoints -n dev-session-service

# Test database connectivity
kubectl exec -it POD_NAME -n dev-session-service -- sh
```

## Architecture

### Components Deployed

1. **Application Deployment**
   - Go service container
   - Environment variables from ConfigMap
   - Secrets mounted for sensitive data
   - Health checks configured

2. **PostgreSQL StatefulSet**
   - Persistent volume for data
   - Single replica with persistent storage
   - Headless service for stable network identity

3. **Services**
   - Application ClusterIP service (port 8080)
   - PostgreSQL headless service (port 5432)

4. **RabbitMQ**
   - Uses shared instance in `default` namespace
   - Connected via: `rabbitmq.default.svc.cluster.local:5672`

### Network Architecture

```
Internet
   │
   ├─> [Ingress/LoadBalancer] (optional)
   │
   └─> dev-session-service namespace
       ├─> dev-session-service (ClusterIP:8080)
       │   └─> Application Pods
       │       ├─> Connects to: PostgreSQL (local)
       │       └─> Connects to: RabbitMQ (default namespace)
       │
       └─> postgres (Headless:5432)
           └─> PostgreSQL StatefulSet
```

## Rollback

### Using Helm

```bash
# List releases
helm list -n dev-session-service

# View history
helm history dev-session-service -n dev-session-service

# Rollback to previous version
helm rollback dev-session-service -n dev-session-service

# Rollback to specific revision
helm rollback dev-session-service 2 -n dev-session-service
```

### Using kubectl

```bash
# Rollback deployment
kubectl rollout undo deployment/dev-session-service-dev-session-service -n dev-session-service

# Rollback to specific revision
kubectl rollout undo deployment/dev-session-service-dev-session-service --to-revision=2 -n dev-session-service
```

## Cleanup

```bash
# Remove Helm release (keeps namespace)
helm uninstall dev-session-service -n dev-session-service

# Delete namespace (removes everything)
kubectl delete namespace dev-session-service

# Note: PersistentVolumeClaims might need manual deletion
kubectl delete pvc --all -n dev-session-service
```

## Troubleshooting

### Image Pull Errors
- Verify `ghcr-secret` is created correctly
- Ensure GitHub token has `read:packages` permission
- Check image repository name matches exactly

### Database Connection Issues
- Check PostgreSQL pod is running: `kubectl get pods -n dev-session-service`
- Verify secrets are created: `kubectl get secrets -n dev-session-service`
- Check logs: `kubectl logs POD_NAME -n dev-session-service`

### RabbitMQ Connection Issues
- Verify RabbitMQ is running in default namespace: `kubectl get pods -n default | grep rabbitmq`
- Test connectivity: `kubectl exec -it POD_NAME -n dev-session-service -- ping rabbitmq.default.svc.cluster.local`
- Check RabbitMQ credentials in secrets

### Pod CrashLoopBackOff
- Check logs: `kubectl logs POD_NAME -n dev-session-service`
- Describe pod: `kubectl describe pod POD_NAME -n dev-session-service`
- Verify health check endpoints are working

## Security Best Practices

1. **Secrets Management**
   - Never commit secrets to git
   - Use GitHub encrypted secrets
   - Consider using external secret managers (Vault, Sealed Secrets)

2. **Image Security**
   - Use specific image tags (commit SHA) instead of `latest`
   - Scan images for vulnerabilities
   - Use minimal base images (alpine)

3. **Network Policies**
   - Consider implementing NetworkPolicies
   - Restrict inter-namespace communication
   - Use TLS for external communications

4. **RBAC**
   - Use least privilege service accounts
   - Implement proper RBAC rules
   - Regularly audit permissions

## Support

For issues or questions:
1. Check pod logs
2. Review GitHub Actions logs
3. Verify all secrets are configured
4. Check Digital Ocean cluster health
