#!/bin/bash

set -e

APP_NAME="$1"
if [[ -z "$APP_NAME" ]]; then
    echo "You must pass the app name"
fi

cd hello-world-app

## handle namespace creation
echo "Checking kubernetes namespace"
EXISTING_APP_NAMESPACE=$(kubectl get namespaces -o=jsonpath="{.items[?(@.metadata.name==\"$APP_NAME\")].metadata.name}")
echo "$EXISTING_APP_NAMESPACE"
if [[ -z "$EXISTING_APP_NAMESPACE" ]]; then
    echo "Namespace for $APP_NAME not found! Creating"
    kubectl create namespace "$APP_NAME"
    echo "Namespace $APP_NAME created"
fi

# handle docker image creation
echo "Saving new docker image"
DOCKER_IMAGE_VERSION=latest
docker build . -t bruno303/$APP_NAME:$DOCKER_IMAGE_VERSION
docker push bruno303/$APP_NAME:$DOCKER_IMAGE_VERSION

echo "applying kustomization"
cp .env ./deploy/.env # think how to improve this
kubectl apply -k ./deploy
rm ./deploy/.env
