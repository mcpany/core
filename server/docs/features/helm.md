# Helm Chart Support

Official Helm Charts are available for deploying MCP Any to Kubernetes.

## Installation

```bash
helm repo add mcpany https://mcpany.github.io/charts
helm install my-mcpany mcpany/server -f values.yaml
```

## Configuration

The Helm chart supports configuring:
- Server image and resources
- Configuration via ConfigMap or Secret
- Ingress settings
- Service Account and RBAC
- Sidecars and Init Containers

See the `k8s/helm` directory for more details.
