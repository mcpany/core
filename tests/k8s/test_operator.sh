#!/bin/bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# Configuration
CLUSTER_NAME="mcp-e2e"
KIND_IMAGE="kindest/node:v1.29.2"
NAMESPACE="mcp-system"
# Use a production-like tag (matching Chart.yaml appVersion or similar)
TAG="1.0.0"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Starting K8s E2E Tests...${NC}"

# 1. Cleanup previous runs
echo "Cleaning up previous runs..."
if helm status mcpany -n "$NAMESPACE" &> /dev/null; then
    helm uninstall mcpany -n "$NAMESPACE" --wait || true
fi
if kubectl get ns "$NAMESPACE" &> /dev/null; then
    kubectl delete ns "$NAMESPACE" --wait || true
fi
# Optional: could delete cluster, but reusing is faster
# kind delete cluster --name "$CLUSTER_NAME"

# 2. Check prerequisites
for cmd in kind kubectl helm docker; do
    if ! command -v $cmd &> /dev/null; then
        echo -e "${RED}Error: $cmd is not installed${NC}"
        exit 1
    fi
done

# 3. Create Kind Cluster
if kind get clusters | grep -q "^$CLUSTER_NAME$"; then
    echo "Cluster $CLUSTER_NAME already exists."
else
    echo "Creating Kind cluster $CLUSTER_NAME..."
    iptables -P FORWARD ACCEPT
    kind create cluster --name "$CLUSTER_NAME" --image "$KIND_IMAGE" --wait 2m
fi

# Ensure kubectl context is set
kubectl cluster-info --context "kind-$CLUSTER_NAME"

# 4. Build Images (Locally) with Production Tag
echo "Building Docker images with tag $TAG..."
# We assume we are in the root of the repo
docker build -t mcpany/server:$TAG -f server/docker/Dockerfile.server .
docker build -t mcpany/operator:$TAG -f operator/Dockerfile .

# 5. Load Images into Kind
echo "Loading images into Kind..."
kind load docker-image mcpany/server:$TAG --name "$CLUSTER_NAME"
kind load docker-image mcpany/operator:$TAG --name "$CLUSTER_NAME"

# 6. Install Helm Chart
echo "Installing Helm chart..."
# Create namespace
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Upgrade/Install using the local image (Never pull)
helm upgrade --install mcpany server/helm/mcpany \
    --namespace "$NAMESPACE" \
    --set image.repository=mcpany/server \
    --set image.tag=$TAG \
    --set image.pullPolicy=Never \
    --set operator.enabled=true \
    --set operator.image.repository=mcpany/operator \
    --set operator.image.tag=$TAG \
    --set operator.image.pullPolicy=Never \
    --wait \
    --timeout 3m

echo -e "${GREEN}Deployment successful!${NC}"

# 7. Verify Pods
echo "Verifying pods..."
kubectl get pods -n "$NAMESPACE"
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=mcpany -n "$NAMESPACE" --timeout=60s

# 8. Cleanup (Optional, can be skipped for debugging)
if [ "$1" == "--cleanup" ]; then
    echo "Cleaning up..."
    kind delete cluster --name "$CLUSTER_NAME"
fi

echo -e "${GREEN}E2E Tests Passed!${NC}"
