# Deployment Checklist ✓

Use this checklist to ensure everything is configured correctly before deploying.

## Pre-Deployment Checklist

### 1. Digital Ocean Setup
- [ ] Digital Ocean account is active
- [ ] Kubernetes cluster is created and running
- [ ] Cluster name is known (needed for secrets)
- [ ] `doctl` CLI is installed (for manual deployments)

### 2. GitHub Repository Setup
- [ ] Code is pushed to GitHub
- [ ] GitHub Actions is enabled
- [ ] Repository has GitHub Container Registry enabled

### 3. RabbitMQ Verification
- [ ] RabbitMQ is running in the `default` namespace
- [ ] RabbitMQ service name is `rabbitmq` (or update values.yaml)
- [ ] RabbitMQ credentials are known

### 4. GitHub Secrets Configuration
Configure in: **Settings → Secrets and variables → Actions**

- [ ] **DIGITALOCEAN_ACCESS_TOKEN**
  - Generated from Digital Ocean dashboard
  - Has Read + Write permissions
  - Format: `dop_v1_...`

- [ ] **DO_CLUSTER_NAME**
  - Matches your cluster name exactly
  - Example: `paypilot-k8s-cluster`

- [ ] **DB_PASSWORD**
  - Strong password generated (16+ chars)
  - Saved securely for future reference
  - Will be used for PostgreSQL in Kubernetes

- [ ] **RABBITMQ_PASSWORD**
  - Password from RabbitMQ in default namespace
  - Default is `guest` or your custom password

- [ ] **RABBITMQ_HOST** (Optional)
  - Only set if different from default
  - Default: `rabbitmq.default.svc.cluster.local`

### 5. Configuration Review

#### values.yaml Review
- [ ] Image repository is correct
  ```yaml
  image:
    repository: ghcr.io/YOUR_ORG/dev-session-service
  ```

- [ ] Resource limits are appropriate
  ```yaml
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
  ```

- [ ] Database storage size is sufficient
  ```yaml
  database:
    storage:
      size: 10Gi
  ```

- [ ] RabbitMQ host is correct
  ```yaml
  rabbitmq:
    host: "rabbitmq.default.svc.cluster.local"
  ```

#### Dockerfile Review
- [ ] Dockerfile exists and is correct
- [ ] Build command is: `go build -o main ./cmd/api`
- [ ] Exposed port is 8080

### 6. Testing Before Deploy
- [ ] Code compiles successfully locally
  ```bash
  go build ./cmd/api
  ```

- [ ] Docker image builds successfully
  ```bash
  docker build -t test-image .
  ```

- [ ] All tests pass
  ```bash
  go test ./...
  ```

## Deployment Steps

### Automatic Deployment (Recommended)
- [ ] All secrets are configured
- [ ] Changes committed to git
- [ ] Push to `main` or `develop` branch
  ```bash
  git add .
  git commit -m "Deploy to Kubernetes"
  git push origin main
  ```

- [ ] Monitor GitHub Actions workflow
  - Go to: **Actions** tab
  - Watch "Build, Push and Deploy" workflow
  - Verify both jobs complete successfully

### Manual Deployment (Alternative)
- [ ] kubectl is configured
  ```bash
  doctl kubernetes cluster kubeconfig save YOUR_CLUSTER_NAME
  ```

- [ ] Helm is installed (v3.14+)
  ```bash
  helm version
  ```

- [ ] Deploy with Helm
  ```bash
  helm upgrade --install dev-session-service ./helm/dev-session-service \
    --namespace dev-session-service \
    --create-namespace \
    --set image.repository=ghcr.io/YOUR_ORG/YOUR_REPO \
    --set image.tag=latest \
    --set secrets.dbPassword=YOUR_DB_PASSWORD \
    --set secrets.rabbitmqPassword=YOUR_RABBITMQ_PASSWORD \
    --wait
  ```

## Post-Deployment Verification

### 1. Check Deployment Status
- [ ] Namespace is created
  ```bash
  kubectl get namespace dev-session-service
  ```

- [ ] Pods are running
  ```bash
  kubectl get pods -n dev-session-service
  # Should show: Running status for all pods
  ```

- [ ] Services are created
  ```bash
  kubectl get svc -n dev-session-service
  ```

- [ ] StatefulSet is ready (PostgreSQL)
  ```bash
  kubectl get statefulset -n dev-session-service
  ```

### 2. Check Application Health
- [ ] Deployment rollout completed
  ```bash
  kubectl rollout status deployment/dev-session-service-dev-session-service -n dev-session-service
  ```

- [ ] Application logs show no errors
  ```bash
  kubectl logs -f deployment/dev-session-service-dev-session-service -n dev-session-service
  ```

- [ ] Health endpoint responds (if configured)
  ```bash
  kubectl port-forward svc/dev-session-service-dev-session-service 8080:8080 -n dev-session-service
  # Then: curl http://localhost:8080/health
  ```

### 3. Check Database Connection
- [ ] PostgreSQL pod is running
  ```bash
  kubectl get pods -n dev-session-service | grep postgres
  ```

- [ ] Database logs show successful startup
  ```bash
  kubectl logs -f statefulset/dev-session-service-dev-session-service-postgres -n dev-session-service
  ```

- [ ] Application connects to database successfully
  - Check application logs for database connection messages

### 4. Check RabbitMQ Connection
- [ ] RabbitMQ is accessible from the namespace
  ```bash
  kubectl exec -it POD_NAME -n dev-session-service -- ping rabbitmq.default.svc.cluster.local
  ```

- [ ] Application connects to RabbitMQ successfully
  - Check application logs for RabbitMQ connection messages

### 5. Verify Secrets
- [ ] Secrets exist
  ```bash
  kubectl get secrets -n dev-session-service
  ```

- [ ] Docker registry secret exists
  ```bash
  kubectl get secret ghcr-secret -n dev-session-service
  ```

## Common Issues Checklist

### If Pods are Not Starting
- [ ] Check pod events
  ```bash
  kubectl describe pod POD_NAME -n dev-session-service
  ```

- [ ] Check image pull errors
  - Verify ghcr-secret is created
  - Check image name is correct

- [ ] Check resource limits
  - Cluster has enough resources

### If Database Connection Fails
- [ ] PostgreSQL pod is running
- [ ] DB_PASSWORD secret is set correctly
- [ ] Service name is correct in configmap

### If RabbitMQ Connection Fails
- [ ] RabbitMQ is running in default namespace
- [ ] RABBITMQ_PASSWORD is correct
- [ ] Network policy allows cross-namespace traffic

### If Deployment Takes Too Long
- [ ] Check image pull progress
- [ ] Check init containers (if any)
- [ ] Check persistent volume provisioning

## Rollback Plan

If something goes wrong:

- [ ] Rollback using Helm
  ```bash
  helm rollback dev-session-service -n dev-session-service
  ```

- [ ] Or rollback deployment
  ```bash
  kubectl rollout undo deployment/dev-session-service-dev-session-service -n dev-session-service
  ```

- [ ] Check previous version
  ```bash
  helm history dev-session-service -n dev-session-service
  ```

## Cleanup (If Needed)

To completely remove the deployment:

- [ ] Uninstall Helm release
  ```bash
  helm uninstall dev-session-service -n dev-session-service
  ```

- [ ] Delete PVCs (persistent data)
  ```bash
  kubectl delete pvc --all -n dev-session-service
  ```

- [ ] Delete namespace
  ```bash
  kubectl delete namespace dev-session-service
  ```

## Documentation Reference

- [ ] Read **DEPLOYMENT_SUMMARY.md** for quick overview
- [ ] Read **DEPLOYMENT.md** for detailed guide
- [ ] Read **.github/SECRETS_SETUP.md** for secrets setup
- [ ] Review **values.yaml** for all configuration options

## Security Review

- [ ] Secrets are not committed to git
- [ ] Strong passwords are used
- [ ] GitHub token has minimal required permissions
- [ ] Digital Ocean token is stored securely
- [ ] Network policies are considered (optional)
- [ ] Resource limits prevent resource exhaustion

## Monitoring & Maintenance

- [ ] Set up log aggregation (optional)
- [ ] Configure alerting (optional)
- [ ] Plan for regular backups of PostgreSQL
- [ ] Document rollback procedures
- [ ] Schedule regular security updates

---

## Final Check Before Deploy

✅ All GitHub secrets configured
✅ values.yaml reviewed and updated
✅ Code builds and tests pass
✅ RabbitMQ is accessible
✅ Digital Ocean cluster is healthy
✅ Documentation reviewed

**Ready to deploy!** 🚀

```bash
git push origin main
```

Then monitor at: https://github.com/YOUR_ORG/YOUR_REPO/actions
