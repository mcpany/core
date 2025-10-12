# MCP-XY Helm Chart

This Helm chart deploys the [MCP-XY](https://github.com/mcpxy/core) server to a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+

## Installing the Chart

To install the chart with the release name `my-mcpxy`:

```bash
helm install my-mcpxy .
```

The chart will deploy MCP-XY with a default configuration. You can customize the installation by providing your own `values.yaml` file.

## Configuration

The following table lists the configurable parameters of the MCP-XY chart and their default values.

| Parameter | Description | Default |
| --- | --- | --- |
| `replicaCount` | Number of MCP-XY replicas. | `1` |
| `image.repository` | MCP-XY image repository. | `mcpxy/server` |
| `image.pullPolicy` | MCP-XY image pull policy. | `IfNotPresent` |
| `image.tag` | MCP-XY image tag. | `""` (defaults to chart's appVersion) |
| `service.type` | Kubernetes service type. | `ClusterIP` |
| `service.jsonrpcPort` | JSON-RPC port. | `50050` |
| `service.grpcPort` | gRPC port. | `50051` |
| `config` | MCP-XY configuration in YAML format. | See `values.yaml` |

You can specify your configuration in a `values.yaml` file and install the chart with it:

```bash
helm install my-mcpxy . -f my-values.yaml
```

## Uninstalling the Chart

To uninstall/delete the `my-mcpxy` deployment:

```bash
helm uninstall my-mcpxy
```