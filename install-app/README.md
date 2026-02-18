# Install App - Helm Chart Installer

A containerized Go CLI tool to install Helm charts from a pre-packaged repository into your Kubernetes cluster.

## Overview

This tool provides a secure way to deploy Helm charts by packaging them directly into a Docker image. Instead of pointing to external repositories at runtime, all charts are bundled into the container during build time.

## Features

- **Pre-packaged charts**: All Helm charts are included in the Docker image
- **Secure**: No need to expose chart repositories at runtime
- **Simple CLI**: Easy-to-use command-line interface
- **Flexible**: Supports custom values, namespaces, and Helm options
- **Kubernetes-ready**: Includes kubectl and helm binaries

## Quick Start

### Building the Image

```bash
# Build with default settings
make build

# Build with custom registry/tag
IMAGE_REGISTRY=ghcr.io IMAGE_REPO=agentcert/sock-shop IMAGE_TAG=v1.0.0 make build
```

### Pushing to Registry

```bash
# Push to registry
make push

# Build and push in one command
make build-push
```

### Installing Charts

#### Using Docker

```bash
# Install sock-shop chart
docker run --rm \
  -v ~/.kube/config:/kubeconfig:ro \
  --network host \
  agentcert-install-app:latest \
  -folder sock-shop \
  -namespace sock-shop \
  -kubeconfig /kubeconfig

# Dry-run to see what would be installed
docker run --rm \
  -v ~/.kube/config:/kubeconfig:ro \
  --network host \
  agentcert-install-app:latest \
  -folder sock-shop \
  -dry-run \
  -kubeconfig /kubeconfig
```

#### Using Make

```bash
# Install using local kubeconfig
make install-local FOLDER=sock-shop NAMESPACE=sock-shop

# Dry-run
make install-local FOLDER=sock-shop DRY_RUN=true
```

#### In Kubernetes (as a Job)

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: install-sock-shop
  namespace: default
spec:
  template:
    spec:
      serviceAccountName: helm-installer
      containers:
      - name: installer
        image: agentcert-install-app:latest
        args:
          - "-folder"
          - "sock-shop"
          - "-namespace"
          - "sock-shop"
          - "-create-namespace"
      restartPolicy: Never
  backoffLimit: 3
```

## CLI Options

| Flag | Description | Default |
|------|-------------|---------|
| `-folder` | Name of the folder containing Helm chart (required) | - |
| `-release` | Helm release name | folder name |
| `-namespace` | Kubernetes namespace | `default` |
| `-charts-path` | Base path where charts are located | `/charts` |
| `-values` | Path to custom values file | - |
| `-set` | Set values (key=value,key2=value2) | - |
| `-dry-run` | Simulate installation | `false` |
| `-wait` | Wait for resources to be ready | `true` |
| `-timeout` | Timeout for installation | `5m` |
| `-create-namespace` | Create namespace if not exists | `true` |
| `-upgrade` | Upgrade if release exists | `false` |
| `-kubeconfig` | Path to kubeconfig file | - |
| `-context` | Kubernetes context to use | - |

## Examples

### Basic Installation

```bash
# Install with all defaults
install-app -folder sock-shop

# Install into specific namespace
install-app -folder sock-shop -namespace production

# Custom release name
install-app -folder sock-shop -release my-shop -namespace production
```

### Advanced Installation

```bash
# With custom values file
install-app -folder sock-shop -values /path/to/values.yaml

# Override specific values
install-app -folder sock-shop -set image.tag=v2.0.0,replicas=3

# Upgrade existing installation
install-app -folder sock-shop -upgrade -namespace sock-shop

# Specific Kubernetes context
install-app -folder sock-shop -context production-cluster
```

### Dry Run

```bash
# Preview what will be installed
install-app -folder sock-shop -dry-run
```

## Available Charts

The following charts are pre-packaged in the image:

| Chart | Description |
|-------|-------------|
| `sock-shop` | Sock Shop microservices demo application with monitoring |

## Building with Custom Charts

To add your own charts to the image, modify the Dockerfile:

```dockerfile
# Add your charts
COPY my-charts/ /charts/my-charts/
```

Then rebuild:

```bash
make build-no-cache
```

## Security Considerations

1. **Chart Immutability**: Charts are baked into the image at build time, ensuring version consistency
2. **No External Dependencies**: Runtime doesn't require access to Helm repositories
3. **Non-root User**: Container runs as non-root user (UID 1000)
4. **RBAC**: When running as a Kubernetes Job, use appropriate ServiceAccount with minimal permissions

### Required RBAC Permissions

Create a ServiceAccount with Helm installer permissions:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: helm-installer
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: helm-installer
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: helm-installer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: helm-installer
subjects:
- kind: ServiceAccount
  name: helm-installer
  namespace: default
```

> **Note**: The above RBAC is permissive. In production, scope permissions to only the resources needed.

## Troubleshooting

### Common Issues

**1. Connection refused to Kubernetes API**
```
Error: Kubernetes cluster unreachable
```
Solution: Ensure `--network host` is used when running locally, or proper ServiceAccount when running in-cluster.

**2. Chart not found**
```
Error: chart folder not found: /charts/my-chart
```
Solution: Verify the chart is included in the image. Run `make list-charts` to see available charts.

**3. Permission denied**
```
Error: create: failed to create
```
Solution: Ensure the ServiceAccount or kubeconfig has appropriate permissions.

## Development

### Building Locally

```bash
# Build Go binary
go build -o install-app .

# Run tests
make test

# Lint code
make lint
```

### Adding New Features

1. Modify `main.go` with new functionality
2. Update Dockerfile if new dependencies are needed
3. Update README with new options
4. Build and test

## License

Apache License 2.0
