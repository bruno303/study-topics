#!/bin/bash
set -e

TAG="latest"

echo "Building"
./gradlew build -x check -x test

set +e
rm build/libs/*-plain.jar 2> /dev/null
set -e

echo "Generating docker image"
docker build . -t bruno303/k8s-test-java:$TAG
docker push bruno303/k8s-test-java:$TAG
