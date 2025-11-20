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

## Resources

- [K3s Documentation](https://docs.k3s.io/)
- [K3s GitHub](https://github.com/k3s-io/k3s)
