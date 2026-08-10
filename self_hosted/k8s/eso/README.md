# ESO (External Secrets Operator)

Deploys External Secrets Operator (v2.9.0) + ClusterSecretStores for Doppler.

## Bootstrap (one-time)

The `secretstores` and `clustersecretstores` CRDs are too large (>262KB) for ArgoCD's
client-side apply, so they are excluded from this app and must be applied once:

```shell
kubectl apply --server-side --force-conflicts -f crds/
```

Required before the ClusterSecretStores/ExternalSecrets can sync.

## Required secrets (one-time, manual)

Created outside Git (do not commit):

- `doppler-token` (ns `eso`) - Doppler service token for `homelab`/`prd`
- `doppler-token-stg` (ns `eso`) - Doppler service token for `homelab`/`stg`
