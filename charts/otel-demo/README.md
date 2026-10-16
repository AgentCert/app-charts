# OpenTelemetry Demo (Astronomy Shop) Helm Chart

Wraps the official [OpenTelemetry Demo](https://github.com/open-telemetry/opentelemetry-helm-charts/tree/main/charts/opentelemetry-demo)
Helm chart as an OCI dependency, pinned to version `0.40.9` — the same
version [ITBench](https://github.com/itbench-hub/ITBench) pins for its SRE
scenarios (`scenarios/sre/project/roles/applications/defaults/main/managers.yaml`).

## Why this chart is structured differently from `sock-shop` / `bookinfo`

OpenTelemetry Demo is a ~20-microservice, multi-language application
distributed upstream only as a Helm chart (no flat manifests to vendor).
Hand-transcribing each service's Deployment the way `sock-shop` and
`bookinfo` do would risk silently wrong images, env vars, or wiring with no
easy way to verify against upstream. Instead this chart declares the
official chart as a dependency (see `Chart.yaml`) and layers ACE's
namespace/litmus/monitoring/mcpTools conventions on top — the values under
the `opentelemetry-demo:` key in `values.yaml` are passed straight through
to the upstream subchart.

Per-component resource limits are carried over unchanged from ITBench's own
proven configuration
(`scenarios/sre/project/roles/applications/templates/helm/otel_demo/values.j2`),
since those are the values ITBench's published scenario results were
measured against.

See the header comment in `values.yaml` for why the bundled
jaeger/grafana/prometheus/opensearch components are disabled, and why
otel-demo's *internal* traces are not currently wired into ACE's Langfuse
instance (which ingests the *agent's* LLM/tool-call traces via LiteLLM, not
arbitrary application OTLP data).

## Install

```bash
helm dependency build ./otel-demo
helm upgrade --install otel-demo ./otel-demo --namespace otel-demo --create-namespace
```
