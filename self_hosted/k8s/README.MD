# K3s Cluster

## Overview

This directory contains configuration and documentation for a K3s cluster setup.

## Deployed Services

The following services are currently deployed in this K3s cluster:

- **Planning Poker** - Team estimation and planning tool
- **Grafana** - Metrics visualization and dashboards
- **Loki** - Log aggregation system
- **Tempo** - Distributed tracing backend
- **Prometheus** - Metrics collection and monitoring
- **Cloudflared** - Cloudflare Tunnel for secure external access
- **Open-Telemetry Collector** - Centralized telemetry data collection and processing pipeline

## Debug Tips

```shell
kubectl run debug-pod --image=busybox -it --rm --restart=Never --namespace <namespace> -- sh
```

## TODO

### Cluster hygiene

- [ ] Remove cert-manager + traefik stripper (failed experiments)
  - [ ] Git: delete `argo/apps/cert-manager.yaml`, `cert-manager/`, `traefik/stripper.yaml`
  - [ ] Git: remove commented-out `strip-prefix` annotations from the 5 ingress files
  - [ ] Git: remove `cert-manager.io` CertificateRequest exclusion from `argo/cluster-config/argocd-cm.yaml`
  - [ ] Cluster: delete `cloudflare-le-prod` ClusterIssuer and `cloudflare-api-token` secret
  - [ ] Cluster: uninstall the cert-manager chart (verify install method with `helm list -A` first)
  - [ ] Cluster: delete `strip-prefix` Middleware (traefik itself stays - K3s-bundled)
- [ ] Provision Grafana datasources via `observability/grafana/values.yaml` (currently UI-configured, lost on fresh install without restic restore)

### Backup / recovery

- [ ] Add etcd snapshot backup (`k3s etcd-snapshot save` or back up `/var/lib/rancher/k3s/server/db`) - covers argocd-cm, configmaps and other cluster state outside `/k3s-volumes`
- [ ] Confirm restic target is off-machine (rclone destination)
- [ ] Run a restore drill (e.g. restore vaultwarden to a temp dir and verify the SQLite opens)

### Config tuning

- [ ] `timeout.reconciliation: 30s` -> default 180s in `argo/cluster-config/argocd-cm.yaml` (reduces constant reconcile log stream + CPU load)
- [ ] Pin `grafana/loki:latest` to a stable release tag in `observability/loki/deployment.yaml`

## Resources

- [K3s Documentation](https://docs.k3s.io/)
- [K3s GitHub](https://github.com/k3s-io/k3s)
