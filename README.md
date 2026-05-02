# App Charts

Helm charts and tooling for deploying target applications that AI agents will monitor and manage. These applications serve as the "system under test" for AgentCert evaluations.

## Contents

```
app-charts/
├── charts/
│   └── sock-shop/        # Microservices demo application
└── install-app/          # Go CLI tool for app deployment
```

## Charts

### sock-shop

A cloud-native microservices demo application (Weaveworks Sock Shop) used as the default target for chaos engineering experiments.

**Services included:**
- Frontend (Node.js)
- Catalogue (Go)
- Carts (Java)
- Orders (Java)
- Payment (Go)
- Shipping (Java)
- User (Go)
- Queue (RabbitMQ)
- Database (MongoDB)

```bash
# Install via Helm
helm upgrade --install sock-shop charts/sock-shop \
  --namespace sock-shop --create-namespace
```

## Install-App CLI

A Go binary that installs application charts from within Kubernetes (used by AgentCert server to deploy target apps).

```bash
cd install-app

# Build
make build

# Build Docker image with baked-in charts
make docker-build
```

See [install-app/README.md](install-app/README.md) for details.

## Build

```bash
# Build install-app Docker image (includes all app charts)
cd install-app && make docker-build

# Load into kind cluster
kind load docker-image agentcert/agentcert-install-app:latest --name agentcert
```

## Usage with AgentCert

When running AgentCert experiments:

1. AgentCert server calls `install-app` to deploy the target application
2. Chaos experiments are injected into the application namespace
3. AI agents monitor and respond to the induced failures
4. Traces are collected for certification analysis

## License

MIT License - see [LICENSE](LICENSE)
