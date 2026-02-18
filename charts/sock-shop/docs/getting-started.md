# Getting Started Guide

This guide walks you through installing and using the Sock-Shop Helm chart. It is designed as a **step-by-step onboarding guide** for new team members.

---

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

### Step 2: Install the Helm Chart

You can install either from the **chart directory** (for development) or from the **pre-packaged `.tgz` file** (recommended for new users and consistent deployments).

#### Option A: Install from `.tgz` Package (Recommended)

The `.tgz` file is a self-contained Helm package that bundles all templates, values, and metadata. This is the preferred method for consistent, versioned deployments.

```bash
# Navigate to the chart directory
cd MS-SOCKS-SHOP-CHART

# Install from the .tgz package
helm install sock-shop sock-shop-litmus-1.0.0.tgz --create-namespace
```

> **Note**: If the `.tgz` file is not present or you've made chart changes, rebuild it:
> ```bash
> helm package .
> ```
> This will create `sock-shop-litmus-<version>.tgz` in the current directory.

#### Option B: Install from Chart Directory (For Development)

Use this when you are actively modifying templates or values and want to test changes immediately.

```bash
cd MS-SOCKS-SHOP-CHART

# Install from current directory
helm install sock-shop . --create-namespace

# Or install with custom values
helm install sock-shop . --create-namespace -f custom-values.yaml
```

### Step 3: Verify Installation

```bash
# Check Helm release is deployed
helm list -A

# Expected output:
# NAME       NAMESPACE  REVISION  STATUS    CHART                    APP VERSION
# sock-shop  default    1         deployed  sock-shop-litmus-1.0.0   1.0.0
```

### Step 4: Wait for Pods to be Ready

```bash
# Watch sock-shop pods (Ctrl+C to stop watching)
kubectl get pods -n sock-shop -w

# Watch monitoring pods
kubectl get pods -n monitoring -w
```

> **Note**: Java-based services (`carts`, `orders`, `shipping`) take **2-5 minutes** to start due to JVM initialization. All other pods should be ready within 30-60 seconds.

#### Verify All Pods Are Running

```bash
# Quick status check
kubectl get pods -n sock-shop
kubectl get pods -n monitoring
```

Expected: **15 pods** in `sock-shop` and **2 pods** in `monitoring`, all with `STATUS: Running`.

---

## Step 5: Port-Forward & Access Services

Since we run on Minikube, use `kubectl port-forward` to access the services from your local machine.

### Start All Port-Forwards

Run each command in a **separate terminal** (each one will keep running in the foreground):

```bash
# Terminal 1 - Sock-Shop Frontend (Web UI)
kubectl port-forward svc/front-end -n sock-shop 8081:80

# Terminal 2 - MCP K8s Log Tool (REST API)
kubectl port-forward svc/sockshop-log-tool -n sock-shop 8082:8082

# Terminal 3 - MCP Prometheus Tool (REST API)
kubectl port-forward svc/sockshop-prometheus-tool -n sock-shop 8083:8083

# Terminal 4 - Prometheus (Monitoring UI)
kubectl port-forward svc/prometheus -n monitoring 9090:9090

# Terminal 5 - Grafana (Dashboards UI)
kubectl port-forward svc/grafana -n monitoring 3000:3000
```

### Service URLs & Credentials

Once port-forwards are running, access the services at:

| Service | Local URL | Type | Credentials |
|---------|-----------|------|-------------|
| **Sock-Shop Frontend** | [http://localhost:8081](http://localhost:8081) | Web UI | None (open) |
| **K8s Log Tool** | [http://localhost:8082](http://localhost:8082) | REST API | None |
| **Prometheus Tool** | [http://localhost:8083](http://localhost:8083) | REST API | None |
| **Prometheus** | [http://localhost:9090](http://localhost:9090) | Web UI | None |
| **Grafana** | [http://localhost:3000](http://localhost:3000) | Web UI | `admin` / `admin` |

### What Each Service Does

| Service | Description | How to Test |
|---------|-------------|-------------|
| **Sock-Shop Frontend** | The e-commerce web store. Browse socks, add to cart, place orders. | Open in browser, click through the shop |
| **K8s Log Tool** | REST API to fetch Kubernetes pod logs. Used by AI/MCP integrations. | `curl http://localhost:8082/health` |
| **Prometheus Tool** | REST API to query Prometheus metrics. Used by AI/MCP integrations. | `curl http://localhost:8083/health` |
| **Prometheus** | Time-series metrics database with query UI. | Open in browser, run query: `up` |
| **Grafana** | Pre-configured dashboards for Sock-Shop metrics. | Login → Dashboards → Sock-Shop Overview |

### Quick API Tests

```bash
# Test K8s Log Tool health
curl http://localhost:8082/health

# Test K8s Log Tool - fetch front-end logs
curl "http://localhost:8082/logs?namespace=sock-shop&pod=<pod-name>&lines=50"

# Test Prometheus Tool health
curl http://localhost:8083/health

# Test Prometheus Tool - query a metric
curl "http://localhost:8083/query?query=up"

# Test Prometheus Tool - list scrape targets
curl http://localhost:8083/targets

# Get pod names (to use with log tool)
kubectl get pods -n sock-shop -o custom-columns="POD_NAME:.metadata.name" --no-headers
```

### Alternative: Use `minikube service` (if port-forward is unreliable)

```bash
# Opens service in default browser via minikube tunnel
minikube service front-end -n sock-shop
minikube service grafana -n monitoring
minikube service prometheus -n monitoring
```

---

## Useful Commands

### Helm Commands

```bash
# List all releases
helm list --all-namespaces

# Get release status
helm status sock-shop

# Upgrade release (after chart changes)
helm upgrade sock-shop .

# Upgrade from .tgz (after repackaging)
helm package .
helm upgrade sock-shop sock-shop-litmus-1.0.0.tgz

# Rollback to previous version
helm rollback sock-shop 1

# Uninstall release
helm uninstall sock-shop

# Uninstall and clean up namespaces
helm uninstall sock-shop
kubectl delete namespace sock-shop monitoring --ignore-not-found
```

### Repackaging the Chart

If you make changes to templates or `values.yaml`, rebuild the `.tgz`:

```bash
cd MS-SOCKS-SHOP-CHART

# Remove old package
Remove-Item sock-shop-litmus-*.tgz   # PowerShell
# rm sock-shop-litmus-*.tgz          # Bash/Linux

# Build new package
helm package .

# Upgrade running release with new package
helm upgrade sock-shop sock-shop-litmus-1.0.0.tgz
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

> **Java pods (carts, orders, shipping)** will restart 2-3 times during initial startup — this is normal. The JVM takes several minutes to initialize on resource-constrained minikube clusters. Wait 5 minutes before considering them unhealthy.

### Port-Forward Issues

If a port-forward stops working or you get `bind: address already in use`:

```bash
# Kill all existing port-forward processes (Windows PowerShell)
taskkill /F /IM kubectl.exe 2>$null

# Wait a moment, then restart port-forwards
Start-Sleep -Seconds 2
kubectl port-forward svc/front-end -n sock-shop 8081:80
```

If port-forward is consistently unreliable, use `minikube service`:

```bash
minikube service front-end -n sock-shop
minikube service grafana -n monitoring
```

### Port Forward Not Working (Windows)

Use `minikube service` instead:

```bash
# Instead of: kubectl port-forward svc/grafana -n monitoring 3000:3000
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
