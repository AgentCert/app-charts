# Getting Started Guide

This guide walks you through installing and using the Sock-Shop Helm chart.

## Prerequisites

### Required Software

| Software | Version | Purpose |
|----------|---------|---------|
| Kubernetes | 1.19+ | Container orchestration |
| Helm | 3.0+ | Package manager |
| Minikube | Latest | Local K8s cluster |
| Docker Desktop | Latest | Container runtime |
| kubectl | Latest | K8s CLI |

### Verify Installation

```bash
# Check Kubernetes
kubectl version

# Check Helm
helm version

# Check Minikube
minikube version
```

---

## Installation

### Step 1: Start Minikube

```bash
# Start minikube with sufficient resources
minikube start --memory=4096 --cpus=2

# Verify cluster is running
minikube status
```

### Step 2: Clone/Navigate to Chart Directory

```bash
cd MS-SOCKS-SHOP-CHART
```

### Step 3: Install the Helm Chart

```bash
# Install with default values
helm install sock-shop . --create-namespace

# Or install with custom values
helm install sock-shop . --create-namespace -f custom-values.yaml
```

### Step 4: Wait for Pods to be Ready

```bash
# Check sock-shop pods
kubectl get pods -n sock-shop -w

# Check monitoring pods
kubectl get pods -n monitoring -w
```

### Step 5: Access Services

On Windows with Docker driver, use `minikube service`:

```bash
# Sock-Shop Frontend
minikube service front-end -n sock-shop

# Grafana (admin/admin)
minikube service grafana -n monitoring

# Prometheus
minikube service prometheus -n monitoring

# MCP Log Tool
minikube service sockshop-log-tool -n sock-shop
```

---

## Useful Commands

### Helm Commands

```bash
# List all releases
helm list --all-namespaces

# Get release status
helm status sock-shop

# Upgrade release
helm upgrade sock-shop .

# Rollback to previous version
helm rollback sock-shop 1

# Uninstall release
helm uninstall sock-shop

# Uninstall and clean up namespaces
helm uninstall sock-shop
kubectl delete namespace sock-shop monitoring --ignore-not-found
```

### Kubernetes Commands

```bash
# Get all resources in sock-shop
kubectl get all -n sock-shop

# Get all resources in monitoring
kubectl get all -n monitoring

# Get pod logs
kubectl logs <pod-name> -n sock-shop

# Describe pod for troubleshooting
kubectl describe pod <pod-name> -n sock-shop

# Execute command in pod
kubectl exec -it <pod-name> -n sock-shop -- /bin/sh

# Port forward (may not work reliably on Windows)
kubectl port-forward svc/front-end -n sock-shop 8081:80
```

### Minikube Commands

```bash
# Get minikube IP
minikube ip

# Open service in browser
minikube service <service-name> -n <namespace>

# List all services
minikube service list

# SSH into minikube node
minikube ssh

# View minikube dashboard
minikube dashboard

# Stop minikube
minikube stop

# Delete minikube cluster
minikube delete
```

---

## Configuration

### Enable/Disable Components

Edit `values.yaml`:

```yaml
# Enable sock-shop application
sockShop:
  enabled: true

# Enable monitoring (Prometheus + Grafana)
monitoring:
  enabled: true

# Enable MCP tools
mcpTools:
  k8sLogTool:
    enabled: true

# Enable LitmusChaos (optional)
litmus:
  enabled: false
```

### Custom Values File

Create `custom-values.yaml`:

```yaml
sockShop:
  frontEnd:
    replicas: 2
    
monitoring:
  prometheus:
    retention: 720h
    
  grafana:
    replicas: 1
```

Install with custom values:

```bash
helm install sock-shop . -f custom-values.yaml --create-namespace
```

---

## Troubleshooting

### Pods Not Starting

```bash
# Check pod status
kubectl get pods -n sock-shop

# Check pod events
kubectl describe pod <pod-name> -n sock-shop

# Check logs
kubectl logs <pod-name> -n sock-shop
```

### Port Forward Not Working (Windows)

Use `minikube service` instead:

```bash
# Instead of: kubectl port-forward svc/grafana -n monitoring 3030:3000
# Use:
minikube service grafana -n monitoring
```

### Cluster Unreachable

```bash
# Restart minikube
minikube stop
minikube start

# Or reset cluster
minikube delete
minikube start
```

### Clean Up Everything

```bash
# Uninstall helm release
helm uninstall sock-shop

# Delete namespaces
kubectl delete namespace sock-shop monitoring --ignore-not-found

# Delete cluster-wide resources
kubectl delete clusterrole sockshop-log-tool prometheus --ignore-not-found
kubectl delete clusterrolebinding sockshop-log-tool prometheus --ignore-not-found
```

---

## Next Steps

- [Sock-Shop Application](sock-shop.md) - Learn about the microservices
- [Prometheus](prometheus.md) - Configure metrics collection
- [Grafana](grafana.md) - Set up dashboards
- [MCP Tools](mcp-tools.md) - Use the log tool API
