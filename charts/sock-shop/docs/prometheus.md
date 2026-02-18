# Prometheus Monitoring

Prometheus is an open-source monitoring and alerting toolkit. In this Helm chart, Prometheus is configured to automatically discover and scrape metrics from all sock-shop services.

## Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     Prometheus Architecture                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────┐                                           │
│  │    Prometheus    │◀──── Scrapes metrics every 15s            │
│  │                  │                                           │
│  │  - TSDB Storage  │                                           │
│  │  - Alert Rules   │                                           │
│  │  - Service Disc. │                                           │
│  └────────┬─────────┘                                           │
│           │                                                      │
│           ▼                                                      │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Scrape Targets                         │   │
│  │                                                           │   │
│  │  sock-shop/front-end     ──▶ :8079/metrics               │   │
│  │  sock-shop/catalogue     ──▶ :80/metrics                 │   │
│  │  sock-shop/carts         ──▶ :80/metrics                 │   │
│  │  sock-shop/orders        ──▶ :80/metrics                 │   │
│  │  sock-shop/payment       ──▶ :80/metrics                 │   │
│  │  sock-shop/shipping      ──▶ :80/metrics                 │   │
│  │  sock-shop/user          ──▶ :80/metrics                 │   │
│  │  sock-shop/queue-master  ──▶ :80/metrics                 │   │
│  │  kubernetes-nodes        ──▶ node metrics                │   │
│  │  kubernetes-pods         ──▶ pod metrics                 │   │
│  │                                                           │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Configuration

### values.yaml

```yaml
monitoring:
  enabled: true
  
  prometheus:
    replicas: 1
    image: prom/prometheus:v2.25.0
    retention: 360h  # Data retention period (15 days)
    service:
      type: NodePort
      port: 9090
      nodePort: 31090
```

### Enable/Disable

```yaml
# Enable monitoring
monitoring:
  enabled: true

# Disable monitoring
monitoring:
  enabled: false
```

---

## Deployed Resources

When `monitoring.enabled: true`, the following resources are created:

| Resource | Name | Purpose |
|----------|------|---------|
| Namespace | `monitoring` | Isolates monitoring components |
| Deployment | `prometheus-deployment` | Prometheus server |
| Service | `prometheus` | Exposes Prometheus UI |
| ConfigMap | `prometheus-configmap` | Scrape configuration |
| ConfigMap | `prometheus-alertrules` | Alert rules |
| ServiceAccount | `prometheus` | K8s API access |
| ClusterRole | `prometheus` | Read cluster resources |
| ClusterRoleBinding | `prometheus` | Bind role to SA |

---

## Scrape Configuration

The Prometheus ConfigMap (`prometheus-configmap.yaml`) defines what to scrape:

### Kubernetes Service Discovery

```yaml
scrape_configs:
  # Scrape Kubernetes endpoints
  - job_name: kubernetes-service-endpoints
    kubernetes_sd_configs:
      - role: endpoints
    relabel_configs:
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        action: drop
        regex: 'false'
      - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name]
        separator: /
        target_label: job

  # Scrape Kubernetes pods
  - job_name: kubernetes-pods
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: 'true'

  # Scrape Kubernetes nodes
  - job_name: kubernetes-nodes
    kubernetes_sd_configs:
      - role: node
```

### Global Settings

```yaml
global:
  scrape_interval: 15s  # How often to scrape targets
```

---

## Accessing Prometheus

### Via Minikube (Recommended)

```bash
minikube service prometheus -n monitoring
```

### Via Port Forward

```bash
kubectl port-forward svc/prometheus -n monitoring 9090:9090
# Open http://localhost:9090
```

### Via NodePort

```bash
# Get minikube IP
minikube ip

# Access Prometheus
# http://<minikube-ip>:31090
```

---

## Using the Prometheus UI

### Expression Browser

1. Navigate to Prometheus UI
2. Enter a PromQL query in the "Expression" field
3. Click "Execute"
4. View results in "Table" or "Graph" tab

### Useful PromQL Queries

#### Service Status

```promql
# Check if services are up (1=up, 0=down)
up{job=~"sock-shop/.*"}

# Count of healthy services
count(up{job=~"sock-shop/.*"} == 1)
```

#### Memory Metrics

```promql
# Memory usage per service
go_memstats_alloc_bytes{job=~"sock-shop/.*"}

# Heap memory in use
go_memstats_heap_inuse_bytes{job=~"sock-shop/.*"}
```

#### CPU Metrics

```promql
# CPU usage rate (5m average)
rate(process_cpu_seconds_total{job=~"sock-shop/.*"}[5m])
```

#### Goroutines

```promql
# Active goroutines per service
go_goroutines{job=~"sock-shop/.*"}
```

#### Process Metrics

```promql
# Open file descriptors
process_open_fds{job=~"sock-shop/.*"}

# Process start time
process_start_time_seconds{job=~"sock-shop/.*"}
```

---

## Checking Targets

### Via UI

1. Go to Prometheus UI
2. Navigate to **Status** → **Targets**
3. View all scrape targets and their status

### Via API

```bash
# Get all targets
curl http://localhost:9090/api/v1/targets

# Using PowerShell
Invoke-RestMethod -Uri "http://localhost:9090/api/v1/targets" | 
  Select-Object -ExpandProperty data | 
  Select-Object -ExpandProperty activeTargets | 
  Select-Object job, health
```

---

## Alert Rules

Alert rules are defined in `prometheus-alertrules.yaml`:

```yaml
groups:
  - name: sock-shop-alerts
    rules:
      - alert: ServiceDown
        expr: up{job=~"sock-shop/.*"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Service {{ $labels.job }} is down"
          
      - alert: HighMemoryUsage
        expr: go_memstats_alloc_bytes{job=~"sock-shop/.*"} > 100000000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage in {{ $labels.job }}"
```

### View Active Alerts

1. Go to Prometheus UI
2. Navigate to **Alerts**
3. View pending and firing alerts

---

## Data Retention

Configure how long Prometheus keeps data:

```yaml
monitoring:
  prometheus:
    retention: 360h  # 15 days
```

The retention is set via the `--storage.tsdb.retention.time` flag in the deployment.

---

## RBAC Permissions

Prometheus needs cluster-wide permissions to discover services and pods:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prometheus
rules:
  - apiGroups: [""]
    resources:
      - nodes
      - nodes/proxy
      - services
      - endpoints
      - pods
    verbs: ["get", "list", "watch"]
```

---

## Troubleshooting

### Targets Not Being Scraped

```bash
# Check Prometheus logs
kubectl logs deployment/prometheus-deployment -n monitoring

# Verify service discovery
kubectl get endpoints -n sock-shop
```

### Prometheus Pod Not Starting

```bash
# Check pod status
kubectl describe pod -l app=prometheus -n monitoring

# Check RBAC
kubectl get clusterrole prometheus
kubectl get clusterrolebinding prometheus
```

### No Metrics from a Service

Ensure the service has the annotation:

```yaml
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "80"
    prometheus.io/path: "/metrics"
```

---

## Integration with Grafana

Prometheus is automatically configured as a data source in Grafana. See [Grafana Documentation](grafana.md) for dashboard setup.

---

## Related Documentation

- [Getting Started](getting-started.md)
- [Grafana](grafana.md) - Visualize Prometheus metrics
- [Sock-Shop](sock-shop.md) - Application being monitored
