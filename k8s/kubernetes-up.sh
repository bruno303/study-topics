#!/bin/bash

echo "Deploy application to k8s"
kubectl apply -k .
echo "Deploy finished"
