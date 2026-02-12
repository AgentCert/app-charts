# MCP Prometheus Tool for SockShop

A lightweight HTTP API that wraps Prometheus queries for AI/LLM integration via the Model Context Protocol (MCP).

## Overview

This tool provides a simple REST API to query Prometheus metrics for the SockShop microservices application. It's designed to be used by AI assistants (like Claude, ChatGPT) to monitor and analyze service health.

## Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /` | API documentation |
| `GET /health` | Health check (also checks Prometheus connectivity) |
| `GET /query` | Execute instant PromQL query |
| `GET /query_range` | Execute range PromQL query |
| `GET /targets` | Get scrape targets and health |
| `GET /alerts` | Get active alerts |
| `GET /metadata` | Get metric metadata |

## Quick Start

### Build

```bash
docker build -t sockshop-prometheus-tool:latest .
```

### Load into Minikube

```bash
minikube image load sockshop-prometheus-tool:latest
```

### Deploy via Helm

The tool is automatically deployed when enabled in `values.yaml`:

```yaml
mcpTools:
  prometheusTool:
    enabled: true
```

### Access

```bash
minikube service sockshop-prometheus-tool -n sock-shop
```

## API Examples

### Check Service Status

```bash
curl "http://localhost:8083/query?query=up{job=~\"sock-shop/.*\"}"
```

### Get Memory Usage

```bash
curl "http://localhost:8083/query?query=go_memstats_alloc_bytes{job=~\"sock-shop/.*\"}"
```

### Get CPU Trend (Last Hour)

```bash
curl "http://localhost:8083/query_range?query=rate(process_cpu_seconds_total{job=~\"sock-shop/.*\"}[5m])&step=60s"
```

### Check Targets

```bash
curl "http://localhost:8083/targets"
```

### Get Active Alerts

```bash
curl "http://localhost:8083/alerts"
```

### Get Metric Metadata

```bash
curl "http://localhost:8083/metadata?metric=go_goroutines"
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8083` | HTTP server port |
| `PROMETHEUS_URL` | `http://prometheus.monitoring.svc.cluster.local:9090` | Prometheus server URL |

## Architecture

```
AI Assistant ──▶ Prometheus Tool ──▶ Prometheus Server ──▶ Sock-Shop Metrics
    (MCP)          (HTTP API)          (PromQL Engine)        (Go/Java apps)
```
