# Grafana Dashboards

Grafana is an open-source visualization and analytics platform. In this Helm chart, Grafana is pre-configured with Prometheus as a data source and includes an auto-provisioned Sock-Shop dashboard.

## Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     Grafana Architecture                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────┐      ┌──────────────────┐                 │
│  │     Grafana      │◀─────│    Prometheus    │                 │
│  │                  │      │   (Data Source)  │                 │
│  │  ┌────────────┐  │      └──────────────────┘                 │
│  │  │ Dashboards │  │                                           │
│  │  │            │  │                                           │
│  │  │ - Sock-Shop│  │                                           │
│  │  │   Overview │  │                                           │
│  │  └────────────┘  │                                           │
│  └──────────────────┘                                           │
│                                                                  │
│  Auto-Provisioned:                                              │
│  ├── /etc/grafana/provisioning/datasources/                     │
│  │   └── datasources.yaml (Prometheus)                          │
│  ├── /etc/grafana/provisioning/dashboards/                      │
│  │   └── dashboards.yaml (Provider config)                      │
│  └── /var/lib/grafana/dashboards/                               │
│      └── sock-shop-overview.json (Dashboard)                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Configuration

### values.yaml

```yaml
monitoring:
  enabled: true
  
  grafana:
    replicas: 1
    image: grafana/grafana:latest
    service:
      type: NodePort
      port: 3000
      nodePort: 31687
```

---

## Deployed Resources

When `monitoring.enabled: true`, the following Grafana resources are created:

| Resource | Name | Purpose |
|----------|------|---------|
| Deployment | `grafana` | Grafana server |
| Service | `grafana` | Exposes Grafana UI |
| ConfigMap | `grafana-datasources` | Auto-configures Prometheus |
| ConfigMap | `grafana-dashboard-provider` | Enables dashboard loading |
| ConfigMap | `grafana-dashboard-sockshop` | Pre-built dashboard JSON |

---

## Auto-Provisioning

### How It Works

Grafana supports **provisioning** - automatically loading configuration from files at startup. This Helm chart uses three ConfigMaps:

#### 1. Data Source Provisioning

`grafana-datasource.yaml`:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-datasources
data:
  datasources.yaml: |
    apiVersion: 1
    datasources:
      - name: Prometheus
        type: prometheus
        url: http://prometheus:9090
        isDefault: true
        editable: true
```

This automatically configures Prometheus as a data source - no manual setup required.

#### 2. Dashboard Provider

`grafana-dashboard-provider.yaml`:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboard-provider
data:
  dashboards.yaml: |
    apiVersion: 1
    providers:
      - name: 'sock-shop'
        folder: 'Sock Shop'
        type: file
        options:
          path: /var/lib/grafana/dashboards
```

This tells Grafana to load dashboards from `/var/lib/grafana/dashboards`.

#### 3. Dashboard JSON

`grafana-dashboard-sockshop.yaml`:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboard-sockshop
data:
  sock-shop-overview.json: |
    {
      "title": "Sock-Shop Overview",
      "panels": [...]
    }
```

Contains the actual dashboard definition.

---

## Accessing Grafana

### Via Minikube (Recommended)

```bash
minikube service grafana -n monitoring
```

### Via Port Forward

```bash
kubectl port-forward svc/grafana -n monitoring 3030:3000
# Open http://localhost:3030
```

### Login Credentials

| Username | Password |
|----------|----------|
| admin | admin |

You'll be prompted to change the password on first login (can be skipped).

---

## Pre-Built Dashboard

### Sock-Shop Overview

Location: **Dashboards** → **Sock Shop** folder → **Sock-Shop Overview**

#### Panels Included:

| Panel | Description | PromQL Query |
|-------|-------------|--------------|
| Service Status | UP/DOWN status for all services | `up{job=~"sock-shop/.*"}` |
| Memory Usage | Memory allocation per service | `go_memstats_alloc_bytes{job=~"sock-shop/.*"}` |
| Goroutines | Active goroutines per service | `go_goroutines{job=~"sock-shop/.*"}` |
| CPU Usage Rate | CPU consumption over time | `rate(process_cpu_seconds_total{job=~"sock-shop/.*"}[5m])` |
| Open File Descriptors | FD usage per service | `process_open_fds{job=~"sock-shop/.*"}` |

---

## Creating Custom Dashboards

### Via UI

1. Click **+** → **Dashboard**
2. Click **Add visualization**
3. Select **Prometheus** as data source
4. Enter PromQL query
5. Configure visualization options
6. Click **Apply**
7. Click **Save dashboard**

### Useful PromQL Queries for Sock-Shop

```promql
# Service availability
up{job=~"sock-shop/.*"}

# Memory usage by service
go_memstats_alloc_bytes{job=~"sock-shop/.*"}

# Heap memory usage
go_memstats_heap_inuse_bytes{job=~"sock-shop/.*"}

# Goroutines count
go_goroutines{job=~"sock-shop/.*"}

# CPU usage (5m rate)
rate(process_cpu_seconds_total{job=~"sock-shop/.*"}[5m])

# GC pause duration
rate(go_gc_duration_seconds_sum{job=~"sock-shop/.*"}[5m])

# Process uptime
time() - process_start_time_seconds{job=~"sock-shop/.*"}
```

---

## Importing Dashboards

### From Grafana.com

1. Go to **Dashboards** → **New** → **Import**
2. Enter Dashboard ID and click **Load**
3. Select **Prometheus** as data source
4. Click **Import**

### Recommended Dashboard IDs

| ID | Name | Description |
|----|------|-------------|
| 315 | Kubernetes Cluster Monitoring | Cluster overview |
| 8588 | Kubernetes Deployment Statefulset | Workload metrics |
| 6417 | Kubernetes Cluster (Prometheus) | Node and pod metrics |

---

## Configuring Alerts

### Create Alert Rule

1. Edit a panel
2. Go to **Alert** tab
3. Click **Create alert rule from this panel**
4. Configure conditions:
   - Query: `up{job=~"sock-shop/.*"} == 0`
   - Condition: When query returns results
   - Duration: 1m
5. Configure notifications
6. Save

### Example Alert

```yaml
Alert: Service Down
Query: up{job=~"sock-shop/.*"} == 0
For: 1 minute
Message: Service {{ $labels.job }} is not responding
```

---

## Grafana Deployment Configuration

The deployment mounts three ConfigMaps:

```yaml
volumeMounts:
  - mountPath: /etc/grafana/provisioning/datasources
    name: grafana-datasources
  - mountPath: /etc/grafana/provisioning/dashboards
    name: grafana-dashboard-provider
  - mountPath: /var/lib/grafana/dashboards
    name: grafana-dashboards

volumes:
  - configMap:
      name: grafana-datasources
    name: grafana-datasources
  - configMap:
      name: grafana-dashboard-provider
    name: grafana-dashboard-provider
  - configMap:
      name: grafana-dashboard-sockshop
    name: grafana-dashboards
```

---

## Troubleshooting

### Dashboard Not Showing

```bash
# Check Grafana logs
kubectl logs deployment/grafana -n monitoring

# Verify ConfigMaps exist
kubectl get configmap -n monitoring

# Check provisioning directory
kubectl exec -it deployment/grafana -n monitoring -- ls -la /var/lib/grafana/dashboards/
```

### Data Source Not Working

1. Go to **Configuration** → **Data Sources**
2. Click **Prometheus**
3. Click **Test**
4. Check the URL is `http://prometheus:9090`

### No Data in Panels

1. Verify Prometheus is running:
   ```bash
   kubectl get pods -n monitoring
   ```
2. Check if Prometheus has targets:
   - Access Prometheus UI
   - Go to **Status** → **Targets**

### Grafana Pod Keeps Restarting

```bash
# Check pod events
kubectl describe pod -l app=grafana -n monitoring

# Check if ConfigMaps are valid
kubectl get configmap grafana-datasources -n monitoring -o yaml
```

---

## Persistence Note

⚠️ **Warning**: The current configuration uses `emptyDir` for storage. Dashboard changes made through the UI will be lost when the pod restarts.

For persistence, consider:
1. Using the provisioned ConfigMap dashboards (recommended)
2. Adding a PersistentVolumeClaim for `/var/lib/grafana`

---

## Related Documentation

- [Getting Started](getting-started.md)
- [Prometheus](prometheus.md) - Data source configuration
- [Sock-Shop](sock-shop.md) - Application being monitored
