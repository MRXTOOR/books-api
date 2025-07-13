#!/bin/sh
set -e

# Wait for all pods to be ready
kubectl wait --for=condition=Ready pod --all --timeout=120s

# Check service exists
kubectl get svc books-api

# Get ClusterIP and check endpoint
SERVICE_IP=$(kubectl get svc books-api -o jsonpath='{.spec.clusterIP}')
curl -sf "http://$SERVICE_IP:80/api/v1/books" || (echo "Service not responding" && exit 1)

echo "Smoke tests passed!" 