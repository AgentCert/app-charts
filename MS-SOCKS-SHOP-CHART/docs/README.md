# Sock-Shop Helm Chart Documentation

Welcome to the Sock-Shop Helm Chart documentation. This chart deploys a complete microservices demo application with monitoring and observability tools.

## Documentation Index

| Document | Description |
|----------|-------------|
| [Getting Started](getting-started.md) | Installation, commands, and quick start guide |
| [Sock-Shop Application](sock-shop.md) | Microservices architecture and configuration |
| [Prometheus](prometheus.md) | Metrics collection and monitoring setup |
| [Grafana](grafana.md) | Dashboards and visualization |
| [MCP Tools](mcp-tools.md) | Kubernetes log tool for AI/LLM integration |

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Helm Chart                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                 sock-shop namespace                          │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │   │
│  │  │front-end│ │catalogue│ │  carts  │ │ orders  │           │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘           │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │   │
│  │  │ payment │ │shipping │ │  user   │ │rabbitmq │           │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘           │   │
│  │  ┌───────────────────┐ ┌────────────────────────┐           │   │
│  │  │ sockshop-log-tool │ │sockshop-prometheus-tool│           │   │
│  │  │   (MCP Tool)      │ │     (MCP Tool)         │           │   │
│  │  └───────────────────┘ └────────────────────────┘           │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                 monitoring namespace                         │   │
│  │  ┌──────────────┐      ┌─────────────┐                      │   │
│  │  │  Prometheus  │─────▶│   Grafana   │                      │   │
│  │  └──────────────┘      └─────────────┘                      │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

## Quick Start

```bash
# Install from .tgz package (recommended)
cd MS-SOCKS-SHOP-CHART
helm install sock-shop sock-shop-litmus-1.0.0.tgz --create-namespace

# Or install from chart directory (for development)
helm install sock-shop . --create-namespace

# Wait for pods to be ready
kubectl get pods -n sock-shop -w
kubectl get pods -n monitoring -w

# Port-forward all services (run each in a separate terminal)
kubectl port-forward svc/front-end -n sock-shop 8081:80
kubectl port-forward svc/sockshop-log-tool -n sock-shop 8082:8082
kubectl port-forward svc/sockshop-prometheus-tool -n sock-shop 8083:8083
kubectl port-forward svc/prometheus -n monitoring 9090:9090
kubectl port-forward svc/grafana -n monitoring 3000:3000
```

### Service URLs (after port-forward)

| Service | URL | Credentials |
|---------|-----|-------------|
| Sock-Shop Frontend | http://localhost:8081 | — |
| K8s Log Tool (MCP) | http://localhost:8082 | — |
| Prometheus Tool (MCP) | http://localhost:8083 | — |
| Prometheus | http://localhost:9090 | — |
| Grafana | http://localhost:3000 | admin / admin |

```bash
# Uninstall
helm uninstall sock-shop
kubectl delete namespace sock-shop monitoring --ignore-not-found
```

## Requirements

- Kubernetes 1.19+
- Helm 3.0+
- Minikube (for local development)
- Docker Desktop (for Windows)
