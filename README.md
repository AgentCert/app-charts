<div align="center">

# app-charts

**Helm charts and installer image for the *system under test* in AgentCert experiments.**

Where [`agent-charts`](../agent-charts) deploys the AI agent, this repository deploys
the **target application** that the agent monitors and remediates: a microservices stack
(Sock Shop), full Prometheus/Grafana observability, two MCP servers that expose
Kubernetes and Prometheus tooling to agents, and optional Litmus chaos experiment
templates.

![Helm](https://img.shields.io/badge/Helm-v3.14-0F1689?style=flat-square&logo=helm)
![Kubernetes](https://img.shields.io/badge/Kubernetes-1.29-326CE5?style=flat-square&logo=kubernetes)
![Sock Shop](https://img.shields.io/badge/Sock_Shop-microservices-9B5DE5?style=flat-square)
![License](https://img.shields.io/badge/License-MIT-lightgrey?style=flat-square)

</div>

---

## Table of Contents

- [What's in here](#whats-in-here)
- [The `sock-shop` chart](#the-sock-shop-chart)
  - [Application services](#application-services)
  - [Observability stack](#observability-stack)
  - [MCP tooling for agents](#mcp-tooling-for-agents)
  - [Optional chaos experiments](#optional-chaos-experiments)
- [The `install-app` CLI image](#the-install-app-cli-image)
- [Installing](#installing)
- [`app-charts` vs `agent-charts`](#app-charts-vs-agent-charts)
- [How it fits into AgentCert](#how-it-fits-into-agentcert)
- [License](#license)

---

## What's in here

```
app-charts/
├── charts/
│   └── sock-shop/                       # Single Helm chart, v0.1.0 (appVersion 1.0.0)
│       ├── Chart.yaml
│       ├── values.yaml                  # All toggles in one place (no values-<env>.yaml)
│       └── templates/
│           ├── namespaces.yaml          # sock-shop, litmus, monitoring
│           ├── sock-shop/               # 13 Deployments + 13 Services
│           ├── monitoring/              # Prometheus, Grafana, metrics-server, kube-state-metrics
│           ├── mcptools/                # kubernetes-mcp-server, prometheus-mcp-server
│           ├── litmus/                  # chaos-exporter (optional)
│           └── chaos-experiments/       # catalogue-pod-delete, orders-pod-delete (optional)
│
├── install-app/                         # Go CLI + Docker image used by ChaosCenter
│   ├── main.go
│   ├── Dockerfile                       # alpine + helm 3.14 + kubectl 1.29, charts baked in
│   ├── Makefile                         # build / push / install / lint / test
│   ├── build-install-app.sh             # CI: timestamp tag + minikube load + env sync
│   ├── build-and-deploy-app-chart.sh    # Legacy build-and-deploy wrapper
│   └── README.md                        # Detailed CLI reference (263 lines)
│
├── LICENSE
└── README.md
```

---

## The `sock-shop` chart

A single, large Helm chart that brings up the **entire benchmark world**: the target
microservices app, the metrics it exports, the dashboards that visualise them, and the
MCP servers that let an AI agent reason about all of the above. Each piece is gated by
an `enabled` flag in `values.yaml` — no separate env-specific values files.

### Application services

A bundled deploy of the Weaveworks **Sock Shop** demo across 13 services / 13
ClusterIP+LoadBalancer Service pairs:

| Service | Image | Notes |
|---|---|---|
| `front-end` | `weaveworksdemos/front-end:0.3.12` | LoadBalancer 80 → 8079, NodePort 30001 |
| `catalogue` | `weaveworksdemos/catalogue:0.3.5` | Go |
| `catalogue-db` | `weaveworksdemos/catalogue-db:0.3.0` | MySQL |
| `carts` | `weaveworksdemos/carts:0.4.8` | Java, `-Xmx384m`, 500m CPU |
| `carts-db` | `mongo` | Ephemeral storage 2 Gi limit |
| `orders` | `weaveworksdemos/orders:0.4.7` | Java, `-Xmx384m`, 500m CPU |
| `orders-db` | `mongo` | |
| `payment` | `weaveworksdemos/payment:0.4.3` | Go |
| `shipping` | `weaveworksdemos/shipping:0.4.8` | Java, `-Xmx384m`, 500m CPU |
| `user` | `weaveworksdemos/user:0.4.7` | Go |
| `user-db` | `weaveworksdemos/user-db:0.4.0` | |
| `queue-master` | `weaveworksdemos/queue-master:0.3.1` | 500m CPU |
| `rabbitmq` | `rabbitmq:3.6.8` | |

All deployed into the `sock-shop` namespace by default.

### Observability stack

Deployed into the `monitoring` namespace when `prometheus.enabled / grafana.enabled` are
on:

| Component | Image | Exposure |
|---|---|---|
| Prometheus | `prom/prometheus:v2.25.0` | NodePort `31090`, 360 h retention, kubelet proxy on |
| Grafana | `grafana/grafana:latest` | NodePort `31687`, pre-loaded Sock Shop dashboard ConfigMap |
| metrics-server | `registry.k8s.io/metrics-server/metrics-server:v0.7.2` | required for pod-cpu-hog/HPA experiments |
| kube-state-metrics | `registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.13.0` | required for cluster-state queries |

### MCP tooling for agents

This is the **bridge between the target app and the agent under test**. Two MCP servers
expose typed tools the agent calls during a scan:

| Server | Image | Exposure | What it exposes |
|---|---|---|---|
| `kubernetes-mcp-server` | `quay.io/containers/kubernetes_mcp_server:latest` | ClusterIP `:8081` | Pod / Deployment / Service / Event / Log read tools |
| `prometheus-mcp-server` | `agentcert/prometheus-mcp-server:latest` | NodePort `31083` | PromQL query tools against the deployed Prometheus |

> Note: `values.yaml` documents that in some AgentCert deployments the MCP servers are
> wired into the `litmus-exp` namespace by the framework instead — controlled by
> [`AgentCert`](../AgentCert) at experiment-creation time, not here.

### Optional chaos experiments

`templates/chaos-experiments/` ships pre-built Litmus chaos engines for two surfaces:

- `catalogue-pod-delete` — recurring pod kill against the `catalogue` deployment.
- `orders-pod-delete` — same against `orders`.

Both expose `totalChaosDuration`, `chaosInterval`, and `force` via values. The
`chaos-exporter` (`litmuschaos/chaos-exporter:1.13.3`) can be enabled separately to
publish chaos-event metrics into Prometheus.

---

## The `install-app` CLI image

Charts in this repo are not consumed directly by AgentCert — the platform calls a
**baked-in Docker image** (`agentcert/agentcert-install-app`) that bundles the CLI **and
every chart under `charts/`** into a single artifact.

What it does, in order:

1. Parses flags: `--folder`, `--namespace`, `--release`, `--values`, `--set`, `--dry-run`,
   `--wait`, `--timeout`, `--kubeconfig`, `--context`.
2. Validates that `/charts/<folder>` is a real Helm chart.
3. Runs `helm upgrade --install` idempotently against the supplied context/kubeconfig.
4. Returns non-zero on rollout failure (uses `kubectl rollout status` instead of
   `helm --wait` to dodge Helm v3.14 rate-limiter quirks).

Build / push:

```bash
cd install-app

make build                       # → agentcert/agentcert-install-app:latest
make build-no-cache              # full rebuild
make push                        # to registry
make build-push                  # build + push
make tag NEW_TAG=v1.0.0
make run                         # docker run … with kubeconfig mount
make install FOLDER=sock-shop NAMESPACE=sock-shop   # one-shot install
make list-charts                 # introspect what's baked into the image
make lint / make test
```

The Dockerfile is multi-stage — `golang:1.21` builder → `alpine:3.19` runtime with
`helm v3.14.0` + `kubectl v1.29.0` + non-root user (UID 1000). The whole `charts/` tree
is copied at build time into `/charts/`.

`build-install-app.sh` adds the same CI niceties as the install-agent script: timestamped
tag, minikube load, env-var sync into a running `litmusportal-server`.

See [`install-app/README.md`](install-app/README.md) for the full flag reference, RBAC
requirements, and troubleshooting.

---

## Installing

### Via Helm directly (development)

```bash
helm upgrade --install sock-shop charts/sock-shop \
  --namespace sock-shop --create-namespace \
  -f charts/sock-shop/values.yaml \
  --set prometheus.enabled=true \
  --set grafana.enabled=true
```

UI: `http://<node>:30001` &nbsp;·&nbsp; Prometheus: `:31090` &nbsp;·&nbsp; Grafana: `:31687`.

### Via the installer image

```bash
docker run --rm \
  -v ~/.kube/config:/home/appuser/.kube/config:ro \
  agentcert/agentcert-install-app:latest \
  --folder sock-shop --namespace sock-shop --release sock-shop
```

### Via AgentCert ChaosCenter (production path)

1. In the UI, **AppHub → Add Hub** pointing at this repository.
2. ChaosCenter triggers the `install-app` container as part of the scenario's Argo
   workflow.
3. The chaos faults defined in [`chaos-charts`](../chaos-charts) are then run against
   the freshly-deployed app.

---

## `app-charts` vs `agent-charts`

| | `app-charts` *(this repo)* | [`agent-charts`](../agent-charts) |
|---|---|---|
| Deploys | The target application (Sock Shop) + observability + MCP tools | The AI agent(s) under test (flash-agent, k8s-agent) |
| Namespace | `sock-shop`, `monitoring` | `flash-agent`, `k8s-agent` |
| Role in experiment | System under test | System doing the testing |
| Installer image | `agentcert/agentcert-install-app` | `agentcert/agentcert-install-agent` |

Both repos use the same install pattern (Go CLI + baked-in charts in an Alpine+Helm
container) so AgentCert's subscriber only needs to know one calling convention.

---

## How it fits into AgentCert

```
ChaosCenter → install-app  (this repo)  → Sock Shop + monitoring + MCP servers
            → install-agent (agent-charts) → flash-agent + sidecar
            → Argo workflow (chaos-charts) → pod-delete / network-loss / cpu-hog …
                                                       │
                                                       ▼
                                              Langfuse traces → certifier
```

| Component | Role |
|---|---|
| [`AgentCert`](../AgentCert) | Calls `install-app` to deploy the SUT |
| [`chaos-charts`](../chaos-charts) | Injects faults against this deployment |
| [`agent-charts`](../agent-charts) | Deploys the agent; agent's `MCP_URLS` points at the MCP servers deployed *here* |
| [`certifier`](../certifier) | Consumes the resulting traces |

---

## License

MIT — see [LICENSE](LICENSE).
