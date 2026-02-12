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
│  │  ┌───────────────────┐                                      │   │
│  │  │ sockshop-log-tool │  (MCP Tool)                         │   │
│  │  └───────────────────┘                                      │   │
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
# Install
helm install sock-shop . --create-namespace

# Access services (Windows with Minikube)
minikube service front-end -n sock-shop
minikube service grafana -n monitoring
minikube service prometheus -n monitoring

# Uninstall
helm uninstall sock-shop
```

## Requirements

- Kubernetes 1.19+
- Helm 3.0+
- Minikube (for local development)
- Docker Desktop (for Windows)
