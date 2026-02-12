# MCP Tools - Kubernetes Log Tool

The MCP (Model Context Protocol) Kubernetes Log Tool is a lightweight HTTP API that allows AI assistants and LLMs to fetch pod logs from the Kubernetes cluster.

## Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    MCP Log Tool Architecture                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────┐                                           │
│  │   AI Assistant   │                                           │
│  │   (Claude, etc)  │                                           │
│  └────────┬─────────┘                                           │
│           │ HTTP Request                                         │
│           ▼                                                      │
│  ┌──────────────────┐      ┌──────────────────┐                 │
│  │  sockshop-log-   │─────▶│ Kubernetes API   │                 │
│  │      tool        │      │     Server       │                 │
│  │                  │      └────────┬─────────┘                 │
│  │  GET /logs       │               │                           │
│  │  GET /health     │               ▼                           │
│  └──────────────────┘      ┌──────────────────┐                 │
│                            │   Pod Logs       │                 │
│                            │                  │                 │
│                            │ - front-end      │                 │
│                            │ - catalogue      │                 │
│                            │ - carts          │                 │
│                            │ - orders         │                 │
│                            │ - ...            │                 │
│                            └──────────────────┘                 │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Purpose

The MCP Log Tool enables:

1. **AI-Powered Debugging**: Allow LLMs to fetch and analyze pod logs
2. **Automated Troubleshooting**: Enable AI assistants to diagnose issues
3. **Context Gathering**: Provide runtime context for AI decision-making
4. **Integration with MCP**: Works with Model Context Protocol servers

---

## Configuration

### values.yaml

```yaml
mcpTools:
  k8sLogTool:
    enabled: true
    name: sockshop-log-tool
    replicas: 1
    image: sockshop-log-tool:latest
    resources:
      requests:
        memory: "64Mi"
        cpu: "50m"
      limits:
        memory: "128Mi"
        cpu: "100m"
    service:
      type: ClusterIP
      port: 8082
      targetPort: 8082
```

### Enable/Disable

```yaml
# Enable MCP tools
mcpTools:
  k8sLogTool:
    enabled: true

# Disable MCP tools
mcpTools:
  k8sLogTool:
    enabled: false
```

---

## Deployed Resources

When `mcpTools.k8sLogTool.enabled: true`:

| Resource | Name | Purpose |
|----------|------|---------|
| Deployment | `sockshop-log-tool` | API server |
| Service | `sockshop-log-tool` | Exposes HTTP API |
| ServiceAccount | `sockshop-log-tool` | K8s API access |
| ClusterRole | `sockshop-log-tool` | Read pod logs |
| ClusterRoleBinding | `sockshop-log-tool` | Bind role to SA |

---

## API Reference

### Base URL

```
http://sockshop-log-tool:8082
```

Or via minikube:
```bash
minikube service sockshop-log-tool -n sock-shop
```

### Endpoints

#### GET /

Returns API documentation and available endpoints.

**Response:**
```json
{
  "name": "Kubernetes Log Tool",
  "version": "1.0.0",
  "endpoints": {
    "/": "API documentation",
    "/health": "Health check",
    "/logs": "Fetch pod logs"
  }
}
```

#### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2026-02-12T10:00:00Z"
}
```

#### GET /logs

Fetch logs from a specific pod.

**Query Parameters:**

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| namespace | Yes | - | Kubernetes namespace |
| pod | Yes | - | Pod name |
| container | No | (first) | Container name |
| lines | No | 100 | Number of lines to fetch |
| previous | No | false | Get logs from previous instance |

**Example Request:**
```bash
curl "http://localhost:8082/logs?namespace=sock-shop&pod=front-end-67896bfb95-4v26w&lines=50"
```

**Response:**
```json
{
  "namespace": "sock-shop",
  "pod": "front-end-67896bfb95-4v26w",
  "container": "front-end",
  "lines": 50,
  "logs": "2026-02-12T10:00:00Z Starting server...\n2026-02-12T10:00:01Z Listening on port 8079..."
}
```

**Error Response:**
```json
{
  "error": "pod not found",
  "message": "Pod 'invalid-pod' not found in namespace 'sock-shop'"
}
```

---

## Usage Examples

### Fetch Front-End Logs

```bash
# Get pod name
POD=$(kubectl get pods -n sock-shop -l name=front-end -o jsonpath='{.items[0].metadata.name}')

# Fetch logs via MCP tool
curl "http://localhost:8082/logs?namespace=sock-shop&pod=$POD&lines=100"
```

### Fetch Logs from Previous Crashed Container

```bash
curl "http://localhost:8082/logs?namespace=sock-shop&pod=carts-55ff946dbc-wlxlt&previous=true"
```

### Fetch Logs from Specific Container

```bash
curl "http://localhost:8082/logs?namespace=sock-shop&pod=catalogue-db-774d5c9867-l6j49&container=catalogue-db"
```

### Using PowerShell

```powershell
# Fetch logs
$response = Invoke-RestMethod -Uri "http://localhost:8082/logs?namespace=sock-shop&pod=front-end-67896bfb95-4v26w&lines=50"
$response.logs
```

---

## Integration with AI Assistants

### MCP Server Configuration

The log tool can be used as a backend for MCP servers:

```json
{
  "mcpServers": {
    "k8s-logs": {
      "command": "node",
      "args": ["mcp-k8s-server.js"],
      "env": {
        "LOG_TOOL_URL": "http://sockshop-log-tool:8082"
      }
    }
  }
}
```

### Example AI Workflow

1. AI receives error report about sock-shop
2. AI calls `/logs` endpoint to fetch relevant pod logs
3. AI analyzes logs and identifies root cause
4. AI suggests remediation steps

---

## RBAC Permissions

The tool requires read access to pod logs across namespaces:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: sockshop-log-tool
rules:
  - apiGroups: [""]
    resources: ["pods", "pods/log"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list"]
```

---

## Source Code

The tool is written in Go. Source code is located in:

```
mcptools/k8s-log-tool/
├── main.go          # Main application
├── go.mod           # Go module definition
├── Dockerfile       # Container build
├── deployment.yaml  # K8s manifest (reference)
└── README.md        # Tool documentation
```

### Building the Image

```bash
cd mcptools/k8s-log-tool

# Build locally
docker build -t sockshop-log-tool:latest .

# Load into minikube
minikube image load sockshop-log-tool:latest
```

---

## Accessing the Tool

### Via Minikube (Recommended)

```bash
minikube service sockshop-log-tool -n sock-shop
```

### Via Port Forward

```bash
kubectl port-forward svc/sockshop-log-tool -n sock-shop 8082:8082
# Then access http://localhost:8082
```

---

## Troubleshooting

### Tool Not Starting

```bash
# Check pod status
kubectl get pods -n sock-shop -l app=sockshop-log-tool

# Check logs
kubectl logs deployment/sockshop-log-tool -n sock-shop

# Check RBAC
kubectl auth can-i get pods/log --as=system:serviceaccount:sock-shop:sockshop-log-tool
```

### Permission Denied Errors

Ensure ClusterRole and ClusterRoleBinding are created:

```bash
kubectl get clusterrole sockshop-log-tool
kubectl get clusterrolebinding sockshop-log-tool
```

### Image Pull Errors

If using local image:

```bash
# Load image into minikube
minikube image load sockshop-log-tool:latest

# Verify image is available
minikube ssh "docker images | grep sockshop-log-tool"
```

---

## Security Considerations

⚠️ **Warning**: This tool provides read access to pod logs, which may contain sensitive information.

Recommendations:
1. Restrict network access to the tool
2. Use RBAC to limit which pods can be queried
3. Don't expose the service externally in production
4. Consider adding authentication for production use

---

## Related Documentation

- [Getting Started](getting-started.md)
- [Sock-Shop](sock-shop.md) - Application logs being accessed
- [Prometheus](prometheus.md) - Metrics monitoring
