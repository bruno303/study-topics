# ESO (External Secrets Operator)

Deploys External Secrets Operator (v2.9.0) + ClusterSecretStores for Doppler.

## Required secrets (one-time, manual)

Created outside Git (do not commit):

- `doppler-token` (ns `eso`) - Doppler service token for `homelab`/`prd`
- `doppler-token-stg` (ns `eso`) - Doppler service token for `homelab`/`stg`
