# MCP Any Helm Chart

This Helm chart deploys the [MCP Any](https://github.com/mcpany/core) server to a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+

## Installing the Chart

To install the chart with the release name `my-mcpany`:

```bash
helm install my-mcpany .
```

The chart will deploy MCP Any with a default configuration. You can customize the installation by providing your own `values.yaml` file.

## Configuration

The following table lists the configurable parameters of the MCP Any chart and their default values.

| Parameter             | Description                           | Default                               |
| --------------------- | ------------------------------------- | ------------------------------------- |
| `replicaCount`        | Number of MCP Any replicas.           | `1`                                   |
| `image.repository`    | MCP Any image repository.             | `mcpany/server`                       |
| `image.pullPolicy`    | MCP Any image pull policy.            | `IfNotPresent`                        |
| `image.tag`           | MCP Any image tag.                    | `""` (defaults to chart's appVersion) |
| `service.type`        | Kubernetes service type.              | `ClusterIP`                           |
| `service.jsonrpcPort` | JSON-RPC port.                        | `50050`                               |
| `service.grpcPort`    | gRPC port.                            | `50051`                               |
| `config`              | MCP Any configuration in YAML format. | See `values.yaml`                     |

You can specify your configuration in a `values.yaml` file and install the chart with it:

```bash
helm install my-mcpany . -f my-values.yaml
```

## Uninstalling the Chart

To uninstall/delete the `my-mcpany` deployment:

```bash
helm uninstall my-mcpany
```
