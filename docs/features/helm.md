# Helm Chart

**Status**: Implemented

MCP Any provides an official Helm chart for deploying the server to Kubernetes clusters. This chart encapsulates best practices for production deployments, including scalability, observability, and security configuration.

## Location

The Helm chart is located in `server/helm/`.

## Features

*   **Configurable**: Easy override of all configuration values via `values.yaml`.
*   **Scalable**: Supports HPA (Horizontal Pod Autoscaler).
*   **Observability**: Built-in support for Prometheus metrics and OpenTelemetry sidecars.
*   **Security**: Pod Security Policies and non-root execution by default.

## Installation

```bash
helm install mcp-any ./server/helm
```
