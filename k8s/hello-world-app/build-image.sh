#!/bin/bash

docker build . -t bruno303/k8s-test-app:0.0.2

docker push bruno303/k8s-test-app:0.0.2