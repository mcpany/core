# MCP Any Kubernetes Operator

The MCP Any Kubernetes Operator manages `MCPServer` resources, automating the deployment and configuration of MCP Any servers on Kubernetes.

## Features

- **Automated Deployment**: Creates and manages Kubernetes Deployments for MCP Servers.
- **Service Exposure**: Automatically creates Kubernetes Services to expose the servers.
- **Configuration Management**: Mounts configuration files from ConfigMaps.
- **Reconciliation**: Ensures the state of the cluster matches the desired state defined in the CR.

## Usage

### Prerequisites

- Kubernetes cluster
- `kubectl` installed
- MCP Any Operator deployed

### Creating an MCP Server

Define an `MCPServer` custom resource:

```yaml
apiVersion: mcp.any/v1alpha1
kind: MCPServer
metadata:
  name: weather-service
  namespace: default
spec:
  # Number of replicas
  replicas: 2

  # Image to use
  image: ghcr.io/mcpany/server:latest

  # Name of the ConfigMap containing config.yaml
  configMap: weather-config

  # Service configuration
  serviceType: ClusterIP # Can be ClusterIP, NodePort, or LoadBalancer
  servicePort: 8080      # Port to expose the service on (default: 8080)
```

Apply it to the cluster:

```bash
kubectl apply -f mcpserver.yaml
```

The Operator will:
1. Create a Deployment with 2 replicas.
2. Mount the `weather-config` ConfigMap.
3. Create a Service named `weather-service` exposing port 8080.

### Configuration Options

| Field | Description | Default |
|-------|-------------|---------|
| `replicas` | Number of pods to run | `1` |
| `image` | Container image | `ghcr.io/mcpany/server:latest` |
| `configMap` | ConfigMap containing `config.yaml` | **Required** |
| `serviceType` | Kubernetes Service Type | `ClusterIP` |
| `servicePort` | Port exposed by the Service | `8080` |

## Development

1. **Run tests**: `go test ./controllers/...`
2. **Build manager**: `go build -o bin/manager main.go`
