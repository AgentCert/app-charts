# Sock-Shop Litmus Helm Chart

A Helm chart for deploying the Sock-Shop microservices application with LitmusChaos integration for chaos engineering experiments.

## Overview

This Helm chart deploys:
- **Sock-Shop Application**: A complete microservices demo application (WeaveWorks)
- **Monitoring Stack**: Prometheus and Grafana for observability
- **Litmus Chaos Exporter**: For exporting chaos metrics
- **Chaos Experiments**: Pre-configured pod-delete experiments (optional)

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- LitmusChaos 3.x installed on your cluster (in `litmus` namespace)
- Monitoring stack (Prometheus/Grafana) installed separately (in `monitoring` namespace)
- Minikube, Docker Desktop, or any Kubernetes cluster

## Quick Start Installation

### Step 1: Lint the Helm Chart

Validate the chart for any errors:

```bash
helm lint "C:\Work\Infosys\Repos\sock-shop-litmus-chart"
```

### Step 2: Check for Existing Deployment

Check if sock-shop namespace already exists:

```bash
kubectl get namespaces | findstr sock-shop
```

If it exists, check running pods:

```bash
kubectl get pods -n sock-shop
```

### Step 3: Delete Existing Deployment (if needed)

```bash
kubectl delete namespace sock-shop
```

Wait for namespace deletion to complete:

```bash
kubectl get namespaces -w
```

Press `Ctrl+C` once `sock-shop` disappears.

### Step 4: Configure values.yaml

Since LitmusChaos and Monitoring are managed separately, ensure they are disabled in `values.yaml`:

```yaml
litmus:
  enabled: false

monitoring:
  enabled: false
```

### Step 5: Install the Helm Chart

```bash
helm install sock-shop "C:\Work\Infosys\Repos\sock-shop-litmus-chart" --create-namespace --namespace sock-shop
```

### Step 6: Verify Installation

Check Helm release:

```bash
helm list -n sock-shop
```

Check all pods are running:

```bash
kubectl get pods -n sock-shop
```

Check services:

```bash
kubectl get svc -n sock-shop
```

## Accessing the Application

### Port Forwarding

Open separate terminal windows for each service:

| Service | Command | URL |
|---------|---------|-----|
| Sock-Shop Frontend | `kubectl port-forward svc/front-end -n sock-shop 8081:80` | http://localhost:8081 |
| Prometheus | `kubectl port-forward svc/prometheus -n monitoring 9090:9090` | http://localhost:9090 |
| Grafana | `kubectl port-forward svc/grafana -n monitoring 3000:3000` | http://localhost:3000 |

### Sock-Shop Frontend

```bash
kubectl port-forward svc/front-end -n sock-shop 8081:80
```

Then open http://localhost:8081 in your browser.

### Grafana Dashboard

```bash
kubectl port-forward svc/grafana -n monitoring 3000:3000
```

Open http://localhost:3000 (default credentials: admin/admin)

### Prometheus

```bash
kubectl port-forward svc/prometheus -n monitoring 9090:9090
```

Open http://localhost:9090

## Useful Commands Reference

```bash
# View all Helm releases
helm list -A

# Upgrade existing release
helm upgrade sock-shop "C:\Work\Infosys\Repos\sock-shop-litmus-chart" -n sock-shop

# Uninstall release
helm uninstall sock-shop -n sock-shop

# View pod logs
kubectl logs <pod-name> -n sock-shop

# Describe pod for troubleshooting
kubectl describe pod <pod-name> -n sock-shop
```

## Advanced Installation

### Key Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `sockShop.enabled` | Enable Sock-Shop deployment | `true` |
| `monitoring.enabled` | Enable Prometheus/Grafana | `true` |
| `litmus.enabled` | Enable Litmus components | `true` |
| `chaosExperiments.enabled` | Enable chaos experiments | `false` |

### Customizing Values

Create a `custom-values.yaml`:

```yaml
sockShop:
  frontEnd:
    replicas: 2
    service:
      type: NodePort

monitoring:
  grafana:
    service:
      nodePort: 32000

chaosExperiments:
  enabled: true
  cataloguePodDelete:
    enabled: true
    totalChaosDuration: "60"
```

Install with custom values:

```bash
helm install sock-shop-demo . -f custom-values.yaml
```

## Running Chaos Experiments

### Using LitmusChaos UI

1. Open LitmusChaos UI (typically at http://localhost:8185)
2. Create a new Environment
3. Register the Infrastructure (your Minikube cluster)
4. Create a new Chaos Experiment:
   - Select "pod-delete" fault
   - Target: `sock-shop` namespace
   - App Label: `name=catalogue` or `name=orders`
5. Run the experiment and observe in Grafana

### Using Helm Chart (Pre-configured)

Enable chaos experiments in values:

```yaml
chaosExperiments:
  enabled: true
  cataloguePodDelete:
    enabled: true
```

```bash
helm upgrade sock-shop-demo . --set chaosExperiments.enabled=true --set chaosExperiments.cataloguePodDelete.enabled=true
```

### Manual Chaos Engine

```bash
kubectl apply -f - <<EOF
apiVersion: litmuschaos.io/v1alpha1
kind: ChaosEngine
metadata:
  name: catalogue-chaos
  namespace: litmus
spec:
  engineState: 'active'
  appinfo:
    appns: 'sock-shop'
    applabel: 'name=catalogue'
    appkind: 'deployment'
  chaosServiceAccount: litmus-admin
  experiments:
    - name: pod-delete
      spec:
        components:
          env:
            - name: TOTAL_CHAOS_DURATION
              value: '30'
            - name: CHAOS_INTERVAL
              value: '10'
EOF
```

## Onboarding to LitmusChaos UI

To use this Helm chart with LitmusChaos UI:

1. **Install the Chart**: Deploy the sock-shop application
2. **Open LitmusChaos UI**: Navigate to your LitmusChaos portal
3. **Create Environment**: 
   - Go to Environments → New Environment
   - Name: `sock-shop-env`
   - Type: Production/Non-Production
4. **Register Infrastructure**:
   - Use the generated manifest or litmusctl
   - Ensure infrastructure connects successfully
5. **Create Chaos Workflow**:
   - Select target application (sock-shop namespace)
   - Choose faults (pod-delete, container-kill, etc.)
   - Set schedule and run

## Monitoring Chaos Impact

1. Configure Grafana Datasource:
   - Add Prometheus: `http://prometheus.monitoring.svc.cluster.local:9090`

2. Import Dashboard:
   - Use the provided dashboard JSON from the original sock-shop demo

3. Observe metrics during chaos:
   - Request latency
   - Error rates
   - Pod restarts

## Cleanup

```bash
# Uninstall the Helm release
helm uninstall sock-shop-demo

# Delete namespaces (if needed)
kubectl delete ns sock-shop monitoring
```

## Troubleshooting

### Pods not starting

```bash
kubectl describe pod <pod-name> -n sock-shop
kubectl logs <pod-name> -n sock-shop
```

### Chaos experiment fails

```bash
kubectl describe chaosengine <engine-name> -n litmus
kubectl get chaosresult -n litmus
```

### No metrics in Grafana

1. Check Prometheus targets: http://localhost:9090/targets
2. Verify chaos-exporter is running: `kubectl get pods -n litmus`

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────┐    │
│  │               sock-shop namespace                    │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │    │
│  │  │front-end│ │catalogue│ │  carts  │ │ orders  │   │    │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘   │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │    │
│  │  │ payment │ │shipping │ │  user   │ │rabbitmq │   │    │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘   │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌──────────────────────┐  ┌──────────────────────────┐    │
│  │   litmus namespace   │  │  monitoring namespace    │    │
│  │  ┌────────────────┐  │  │  ┌──────────┐ ┌───────┐ │    │
│  │  │ chaos-exporter │  │  │  │prometheus│ │grafana│ │    │
│  │  └────────────────┘  │  │  └──────────┘ └───────┘ │    │
│  └──────────────────────┘  └──────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## License

This chart is provided under the Apache 2.0 License.
