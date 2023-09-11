#!/bin/bash

set -e

echo "Creating kubernetes cluster with kind"
kind create cluster -n local-test-cluster

CLUSTER_IP=$(kubectl get nodes -o=jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
echo "Kind k8s cluster configured. IP: $CLUSTER_IP"

echo "Configuring kong api-gateway"
cd kong
set +e
rm kong.yml 2>/dev/null
set -e
export K8S_CLUSTER_IP="$CLUSTER_IP"
envsubst < kong-template.yml > kong.yml
docker compose up -d
cd ..
echo "Kong configured"
