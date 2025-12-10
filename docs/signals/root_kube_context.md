# Root Kubernetes Context

## What this is

This signal detects when your current Kubernetes context has cluster-admin or root-level privileges. This is determined by checking if your context can list namespaces cluster-wide, which typically requires elevated permissions.

Working with cluster-admin privileges is risky because mistakes can affect the entire cluster and all applications running on it.

## Why this matters

**Operational Risk**:
- **Accidental deletions**: Can delete critical resources across all namespaces
- **Service outages**: Mistakes can take down production workloads
- **Data loss**: Can delete persistent volumes and stateful data
- **Cluster damage**: Can modify or delete cluster-level resources

**Security Risk**:
- **Credential exposure**: Cluster-admin credentials are high-value targets
- **Lateral movement**: Attackers with cluster-admin can access all workloads
- **Privilege escalation**: Can create privileged pods to escape containers
- **Secret access**: Can read all secrets in all namespaces

**Examples of dangerous operations**:
```bash
# Accidentally delete all pods in all namespaces
kubectl delete pods --all --all-namespaces  # Oops!

# Delete production namespace
kubectl delete namespace production  # Meant to delete dev!

# Modify cluster-wide resources
kubectl delete clusterrole system:node  # Breaks the cluster
```

## How to remediate

### Switch to a less privileged context

**List available contexts**:
```bash
# See all contexts
kubectl config get-contexts

# See current context
kubectl config current-context
```

**Switch to namespace-scoped context**:
```bash
# Switch to a context with limited permissions
kubectl config use-context dev-context

# Or create a new context for specific namespace
kubectl config set-context dev --cluster=my-cluster --namespace=dev --user=dev-user
kubectl config use-context dev
```

**Verify limited permissions**:
```bash
# This should fail if you don't have cluster-admin
kubectl get namespaces
# Error: namespaces is forbidden

# But this should work in your namespace
kubectl get pods
```

### Create namespace-scoped contexts

**Create a context for each namespace**:
```bash
# For development namespace
kubectl config set-context dev \
  --cluster=my-cluster \
  --namespace=development \
  --user=dev-user

# For staging namespace
kubectl config set-context staging \
  --cluster=my-cluster \
  --namespace=staging \
  --user=staging-user

# Switch to dev context
kubectl config use-context dev
```

**Set default namespace** for current context:
```bash
# Set default namespace
kubectl config set-context --current --namespace=development

# Verify
kubectl config view --minify | grep namespace:
```

### Use RBAC to limit permissions

**Create a Role** (namespace-scoped):
```yaml
# dev-role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: development
  name: dev-role
rules:
- apiGroups: ["", "apps", "batch"]
  resources: ["pods", "deployments", "jobs", "services"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

**Create a RoleBinding**:
```yaml
# dev-rolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: dev-rolebinding
  namespace: development
subjects:
- kind: User
  name: dev-user
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: dev-role
  apiGroup: rbac.authorization.k8s.io
```

**Apply the RBAC**:
```bash
kubectl apply -f dev-role.yaml
kubectl apply -f dev-rolebinding.yaml
```

### Use separate kubeconfig files

**Create separate kubeconfig for admin**:
```bash
# Copy current kubeconfig to admin-specific file
cp ~/.kube/config ~/.kube/config-admin

# Use regular kubeconfig by default
export KUBECONFIG=~/.kube/config

# Only use admin config when absolutely necessary
export KUBECONFIG=~/.kube/config-admin
```

**Use kubeconfig per environment**:
```bash
# Development
export KUBECONFIG=~/.kube/config-dev

# Staging
export KUBECONFIG=~/.kube/config-staging

# Production (admin only)
export KUBECONFIG=~/.kube/config-prod
```

### Add safety checks

**Require confirmation for cluster-admin operations**:
```bash
# Add to ~/.bashrc or ~/.zshrc
kubectl() {
  local context=$(command kubectl config current-context)

  # Check if using admin context
  if [[ "$context" == *"admin"* ]] || [[ "$context" == *"prod"* ]]; then
    echo "⚠️  WARNING: Using admin/production context: $context"

    # Require confirmation for destructive operations
    if [[ "$1" == "delete" ]] || [[ "$1" == "apply" ]]; then
      read -p "Are you sure? (yes/no): " confirm
      if [[ "$confirm" != "yes" ]]; then
        echo "Aborted."
        return 1
      fi
    fi
  fi

  command kubectl "$@"
}
```

**Use kubectl plugins** for safety:
```bash
# Install kubectl-safe plugin
kubectl krew install safe

# Use safe mode
kubectl safe delete pod my-pod
```

### Platform-specific considerations

**GKE (Google Kubernetes Engine)**:
```bash
# Get credentials with limited permissions
gcloud container clusters get-credentials my-cluster \
  --region=us-central1 \
  --project=my-project

# Don't use --admin flag unless necessary
```

**EKS (Amazon Elastic Kubernetes Service)**:
```bash
# Get credentials
aws eks update-kubeconfig --name my-cluster --region us-west-2

# Use IAM roles for fine-grained permissions
# Don't use cluster-admin role
```

**AKS (Azure Kubernetes Service)**:
```bash
# Get credentials
az aks get-credentials --resource-group my-rg --name my-cluster

# Use --admin flag only when necessary
# az aks get-credentials --resource-group my-rg --name my-cluster --admin
```

### Best practices

1. **Use least privilege**:
   ```bash
   # Default to namespace-scoped contexts
   # Only use cluster-admin when absolutely necessary
   ```

2. **Separate contexts by environment**:
   ```bash
   # dev-context for development
   # staging-context for staging
   # prod-context for production (limited access)
   ```

3. **Use namespaces** to isolate workloads:
   ```bash
   # Create namespace per team/project
   kubectl create namespace team-a
   kubectl create namespace team-b
   ```

4. **Audit cluster-admin usage**:
   ```bash
   # Enable audit logging
   # Monitor who uses cluster-admin and when
   ```

5. **Use admission controllers**:
   ```yaml
   # Prevent privileged pods
   # Enforce resource limits
   # Require security contexts
   ```

6. **Color-code your prompt** by context:
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   KUBE_PS1_SYMBOL_COLOR=red  # For prod
   KUBE_PS1_SYMBOL_COLOR=yellow  # For staging
   KUBE_PS1_SYMBOL_COLOR=green  # For dev
   ```

7. **Use tools like kubectx/kubens**:
   ```bash
   # Install kubectx and kubens
   brew install kubectx

   # Quick context switching
   kubectx dev
   kubens development
   ```

### If you must use cluster-admin

1. **Document why** you need cluster-admin
2. **Time-box the access** (use for specific task, then switch back)
3. **Use dry-run** for testing:
   ```bash
   kubectl apply -f manifest.yaml --dry-run=client
   kubectl apply -f manifest.yaml --dry-run=server
   ```
4. **Review changes** before applying:
   ```bash
   kubectl diff -f manifest.yaml
   ```
5. **Have a rollback plan**
6. **Notify team** before making cluster-wide changes


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_ROOT_KUBE_CONTEXT=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
