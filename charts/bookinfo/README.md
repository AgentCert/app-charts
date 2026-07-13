# Bookinfo Helm Chart

Vendors Istio's Bookinfo sample application (details, ratings, reviews-v1/v2/v3,
productpage) from [istio/istio](https://github.com/istio/istio) tag `1.30.2` —
the same tag [ITBench](https://github.com/itbench-hub/ITBench) pins for its SRE
scenarios (`scenarios/sre/project/roles/applications/defaults/main/managers.yaml`).

## Deviation from upstream: no service mesh

Upstream Bookinfo exposes `productpage` through an Istio-managed Gateway API
`Gateway`/`HTTPRoute` (`gatewayClassName: istio`). ACE's cluster stack does not
provision Istio (same as `sock-shop`, which is also mesh-free), so this chart
exposes `productpage` directly via a plain, values-configurable `Service`
instead — mirroring how `sock-shop`'s `front-end-service.yaml` works. Bookinfo's
per-version Services (`reviews-v1/v2/v3`, `productpage-v1`, `ratings-v1`,
`details-v1`) are vendored as-is since they don't depend on the mesh.

Consequence: ITBench fault types that specifically target the Istio layer
(mTLS enforcement, authorization-policy faults, ambient-mode faults) are not
reproducible against this chart until Istio is added as a cluster
prerequisite. That is out of scope for this chart and tracked separately.

## Install

```bash
helm upgrade --install bookinfo ./bookinfo --namespace book-info --create-namespace
```

Disable the load generator if needed:

```bash
helm upgrade --install bookinfo ./bookinfo --namespace book-info --create-namespace --set bookInfo.loadGenerator.enabled=false
```
