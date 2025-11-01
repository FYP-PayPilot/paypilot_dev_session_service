# Digital Ocean Container Registry Setup

## ✅ Configuration Updated

Your deployment now uses **Digital Ocean Container Registry** instead of GitHub Container Registry.

---

## 🎯 What Changed

### Registry Details
- **Registry URL**: `registry.digitalocean.com`
- **Registry Name**: `paypilot`
- **Full Image Path**: `registry.digitalocean.com/paypilot/dev-session-service:tag`

### Secrets Updated
The workflow now uses:
- `DIGITALOCEAN_REGISTRY_TOKEN` for Docker registry authentication
- Secret name in K8s: `docr-secret` (Digital Ocean Container Registry secret)

---

## 🔐 GitHub Secrets Required

You need to add these secrets to GitHub:

### 1. DIGITALOCEAN_ACCESS_TOKEN
**Purpose**: Access to Kubernetes cluster  
**Value**: Your DO API token with Kubernetes access

### 2. DIGITALOCEAN_REGISTRY_TOKEN  
**Purpose**: Push/pull images to/from DO Container Registry  
**Value**: Your DO API token with registry access

> **💡 Tip**: You can use the **same token** for both secrets, or create separate tokens for better security.

### 3. Other Secrets (same as before)
- `DO_CLUSTER_NAME`
- `DB_PASSWORD`
- `RABBITMQ_PASSWORD`
- `RABBITMQ_HOST` (optional)

---

## 📝 Setting Up the Token

### Option A: Use Single Token (Easiest)

1. Go to https://cloud.digitalocean.com/account/api/tokens
2. Click **"Generate New Token"**
3. Name: `GitHub Actions - K8s & Registry`
4. Scopes: **Read + Write** ✅
5. Copy the token

Then add it to GitHub secrets as:
- `DIGITALOCEAN_ACCESS_TOKEN` = `dop_v1_...`
- `DIGITALOCEAN_REGISTRY_TOKEN` = `dop_v1_...` (same token)

### Option B: Separate Tokens (More Secure)

Create two separate tokens:

**Token 1 - Kubernetes Access**
1. Name: `GitHub Actions - Kubernetes`
2. Scopes: Read + Write
3. Add as `DIGITALOCEAN_ACCESS_TOKEN`

**Token 2 - Registry Access**
1. Name: `GitHub Actions - Registry`
2. Scopes: Read + Write
3. Add as `DIGITALOCEAN_REGISTRY_TOKEN`

---

## 🔍 Verify Your Registry

### Check Registry Exists

1. Go to Digital Ocean Dashboard
2. Navigate to **Container Registry** in sidebar
3. Verify you have a registry named `paypilot`

### If Registry Doesn't Exist

Create it:
```bash
# Using doctl CLI
doctl registry create paypilot --subscription-tier-slug basic

# Or create via Digital Ocean Dashboard
# Container Registry → Create Registry → Name: paypilot
```

---

## 🚀 How It Works

### Build & Push Flow

```bash
1. Code pushed to GitHub
   ↓
2. GitHub Actions builds Docker image
   ↓
3. Logs in to registry.digitalocean.com using DIGITALOCEAN_REGISTRY_TOKEN
   ↓
4. Pushes image to registry.digitalocean.com/paypilot/dev-session-service:main-abc1234
   ↓
5. Creates K8s secret 'docr-secret' with registry credentials
   ↓
6. Deploys to K8s, which pulls from DO registry using docr-secret
```

---

## 📊 Image Naming

### Format
```
registry.digitalocean.com/paypilot/dev-session-service:main-abc1234
│                        │        │                    │
│                        │        │                    └─ Tag (branch-commit)
│                        │        └─ Repository name
│                        └─ Registry name
└─ Registry URL
```

### Examples by Branch
- Main: `registry.digitalocean.com/paypilot/dev-session-service:main-abc1234`
- Develop: `registry.digitalocean.com/paypilot/dev-session-service:develop-def5678`

---

## 🔍 Verification

### After Deployment

Check the image being used:
```bash
kubectl get deployment dev-session-service-dev-session-service \
  -n dev-session-service \
  -o jsonpath='{.spec.template.spec.containers[0].image}'

# Should output:
# registry.digitalocean.com/paypilot/dev-session-service:main-abc1234
```

### Check Registry Secret
```bash
kubectl get secret docr-secret -n dev-session-service

# Should show:
# NAME          TYPE                             DATA   AGE
# docr-secret   kubernetes.io/dockerconfigjson   1      5m
```

### View Images in Registry
```bash
# Using doctl CLI
doctl registry repository list-tags dev-session-service

# Should show tags like:
# main-abc1234
# develop-def5678
# latest
```

---

## 🐛 Troubleshooting

### Image Pull Error: "authentication required"

**Solution**: Verify the K8s secret is created correctly
```bash
kubectl describe secret docr-secret -n dev-session-service

# Check the .dockerconfigjson field exists
```

**Fix**: Re-run the workflow or manually create the secret:
```bash
kubectl create secret docker-registry docr-secret \
  --docker-server=registry.digitalocean.com \
  --docker-username=YOUR_DO_API_TOKEN \
  --docker-password=YOUR_DO_API_TOKEN \
  --namespace=dev-session-service
```

### Build Error: "denied: access forbidden"

**Solution**: Verify `DIGITALOCEAN_REGISTRY_TOKEN` has write access
1. Check token in GitHub secrets is correct
2. Verify token has Read + Write permissions in DO
3. Check registry name is `paypilot` (matches env.REGISTRY_NAME in workflow)

### Image Not Found

**Solution**: Check registry and repository exist
```bash
# List repositories
doctl registry repository list

# Should show:
# NAME                LATEST TAG    UPDATED AT
# dev-session-service main-abc1234  2024-01-01T10:00:00Z
```

---

## 💰 Cost Note

Digital Ocean charges for Container Registry based on:
- **Storage**: Amount of data stored
- **Bandwidth**: Data transfer (egress)

**Basic Plan**: $5/month includes:
- 500 MB storage
- 500 GB transfer

**Pro Plan**: $20/month includes:
- 5 GB storage
- 5 TB transfer

> **💡 Tip**: Images are typically 50-200 MB each. Clean up old images regularly to save costs.

---

## 🧹 Cleaning Up Old Images

### Manual Cleanup
```bash
# List all tags
doctl registry repository list-tags dev-session-service

# Delete old tag
doctl registry repository delete-tag dev-session-service main-old123

# Run garbage collection to free space
doctl registry garbage-collection start
```

### Automated Cleanup (Optional)

Add this to your registry settings:
1. Go to Container Registry → Settings
2. Enable **Auto-Delete**: Remove images older than X days
3. Set retention: Keep last 10 images

---

## 🔐 Security Best Practices

1. **Use Separate Tokens**: One for K8s, one for registry
2. **Rotate Tokens**: Change tokens every 90 days
3. **Limit Scope**: Use tokens with minimum required permissions
4. **Monitor Access**: Check registry access logs in DO dashboard
5. **Enable 2FA**: Secure your Digital Ocean account

---

## 📝 Summary

✅ **Registry**: registry.digitalocean.com  
✅ **Registry Name**: paypilot  
✅ **Image Path**: registry.digitalocean.com/paypilot/dev-session-service  
✅ **Secret Name**: docr-secret (in K8s)  
✅ **GitHub Secret**: DIGITALOCEAN_REGISTRY_TOKEN  

**You're all set!** The workflow will now push to and pull from your DO Container Registry automatically. 🚀

---

## 🔗 Related Files

- `.github/workflows/deploy.yaml` - Updated for DO registry
- `.github/SECRETS_TEMPLATE.txt` - Secret configuration template
- `helm/dev-session-service/values.yaml` - Default image repository

## 📚 Documentation

- [DO Container Registry Docs](https://docs.digitalocean.com/products/container-registry/)
- [DO API Tokens](https://docs.digitalocean.com/reference/api/create-personal-access-token/)
- [doctl Registry Commands](https://docs.digitalocean.com/reference/doctl/reference/registry/)
