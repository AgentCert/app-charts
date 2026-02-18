# MCP K8s Log Tool - SockShop

A minimal MCP-style Kubernetes log fetching tool built in Go for SockShop microservices. This HTTP service provides an API to fetch Kubernetes pod logs from any cluster accessible via kubeconfig.

## Features

- 🚀 HTTP API for fetching Kubernetes pod logs
- 🔍 Support for filtering by namespace, pod, and container
- 📊 Configurable log line limits
- 🔄 Streaming logs support
- 🌐 CORS-enabled for browser access
- 🏥 Health check endpoint
- 🐳 Docker containerization
- ☸️ Kubernetes deployment ready
- 🔐 RBAC security configuration

## SockShop Services

This tool can fetch logs from all SockShop microservices:
- `front-end` - Web frontend
- `orders` - Order processing service
- `payment` - Payment processing
- `user` - User management
- `catalogue` - Product catalogue
- `carts` - Shopping cart service
- `shipping` - Shipping service
- `queue-master` - Message queue processor

## Quick Start

### Local Development

1. **Prerequisites:**
   - Go 1.21+
   - Minikube or any Kubernetes cluster with SockShop deployed
   - kubectl configured with cluster access

2. **Install dependencies:**
   ```bash
   cd mcptools/k8s-log-tool
   go mod tidy
   ```

3. **Run the service:**
   ```bash
   go run main.go
   ```

4. **Test the API:**
   - Open browser to http://localhost:8080 for documentation
   - Health check: http://localhost:8080/health
   - Fetch logs: http://localhost:8080/logs?namespace=sock-shop&pod=front-end-xxx

### Using Browser DevTools

1. Open browser to http://localhost:8080
2. Open DevTools (F12)
3. Go to Console tab
4. Test the API:
   ```javascript
   // Fetch logs for front-end pod
   fetch('http://localhost:8080/logs?namespace=sock-shop&pod=front-end-xxx&lines=50')
     .then(response => response.text())
     .then(logs => console.log(logs));
   
   // Health check
   fetch('http://localhost:8080/health')
     .then(response => response.json())
     .then(data => console.log(data));
   ```

## API Endpoints

### GET /logs

Fetch pod logs from Kubernetes cluster.

**Query Parameters:**
- `namespace` (required): Kubernetes namespace (e.g., `sock-shop`)
- `pod` (required): Pod name
- `container` (optional): Specific container name
- `lines` (optional): Number of log lines to fetch (default: 100)
- `follow` (optional): Stream logs (true/false, default: false)

**Examples:**
```bash
# Get front-end logs
curl "http://localhost:8080/logs?namespace=sock-shop&pod=front-end-xxx"

# Get orders service logs with line count
curl "http://localhost:8080/logs?namespace=sock-shop&pod=orders-xxx&lines=50"

# Stream payment service logs
curl "http://localhost:8080/logs?namespace=sock-shop&pod=payment-xxx&follow=true"
```

### GET /health

Health check endpoint returning service status.

### GET /

API documentation (HTML interface).

## Deployment

### Docker

1. **Build image:**
   ```bash
   docker build -t sockshop-log-tool:latest .
   ```

2. **Run locally:**
   ```bash
   docker run -p 8080:8080 \
     -v ~/.kube/config:/root/.kube/config:ro \
     sockshop-log-tool:latest
   ```

### Kubernetes

1. **Build and load image (Minikube):**
   ```bash
   eval $(minikube docker-env)
   docker build -t sockshop-log-tool:latest .
   ```

2. **Deploy:**
   ```bash
   kubectl apply -f deployment.yaml -n sock-shop
   ```

3. **Access the service:**
   ```bash
   kubectl port-forward svc/sockshop-log-tool 8080:8080 -n sock-shop
   ```

## RBAC Permissions

The tool requires the following Kubernetes permissions:
- `get` and `list` on `pods` resources
- `get` on `pods/log` resources

These are configured in the `deployment.yaml` via ClusterRole and ClusterRoleBinding.

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `KUBECONFIG` | `~/.kube/config` | Path to kubeconfig file |

## Security Considerations

- Runs as non-root user (UID 1000)
- Read-only root filesystem
- Minimal RBAC permissions (read-only access to pods/logs)
- No privilege escalation allowed
- All capabilities dropped

## Integration with SockShop

When SockShop is deployed, you can use this tool to monitor all services:

```bash
# List all pods in sock-shop namespace
kubectl get pods -n sock-shop

# Then use the tool to fetch logs
curl "http://localhost:8080/logs?namespace=sock-shop&pod=<pod-name>&lines=100"
```

## Troubleshooting

### Cannot connect to cluster
- Ensure kubeconfig is properly configured
- Check if the cluster is accessible: `kubectl cluster-info`

### Pod not found
- Verify the pod name: `kubectl get pods -n sock-shop`
- Ensure the namespace is correct

### Permission denied
- Check RBAC configuration is applied
- Verify ServiceAccount is properly bound to ClusterRole
