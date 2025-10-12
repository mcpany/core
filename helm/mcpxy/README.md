# MCP-XY Helm Chart

This Helm chart deploys the [MCP-XY](https://github.com/mcpxy/core) server to a Kubernetes cluster.

## Introduction

This chart bootstraps an `mcpxy` deployment on a Kubernetes cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.12+
- Helm 3.0+

## Installing the Chart

To install the chart with the release name `mcpxy`:

```bash
helm install mcpxy .
```

The command deploys `mcpxy` on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `mcpxy` deployment:

```bash
helm uninstall mcpxy
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the `mcpxy` chart and their default values.

| Parameter | Description | Default |
| --- | --- | --- |
| `replicaCount` | Number of `mcpxy` replicas. | `1` |
| `image.repository` | Image repository. | `mcpxy/server` |
| `image.pullPolicy` | Image pull policy. | `IfNotPresent` |
| `image.tag` | Image tag. Overrides the chart's `appVersion`. | `""` |
| `service.type` | Kubernetes service type. | `ClusterIP` |
| `service.jsonrpcPort` | Port for the JSON-RPC service. | `50050` |
| `service.grpcPort` | Port for the gRPC registration service. | `50051` |
| `config` | The `mcpxy` configuration file content. | See `values.yaml` |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```bash
helm install mcpxy . --set service.type=LoadBalancer
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
helm install mcpxy . -f values.yaml
```

For more detailed deployment instructions, see the [deployment guide](../../docs/deployment.md).