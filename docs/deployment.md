# 🚀 Deploying MCP-XY

This guide provides instructions for deploying the `mcpxy` server using Docker Compose for local development and Helm for Kubernetes environments.

## Docker Compose

The included `docker-compose.yml` file provides a simple way to run `mcpxy` and a sample `http-echo-server` to demonstrate inter-service communication in a containerized environment.

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

### Running the Services

1.  **Navigate to the root of the repository.**

2.  **Build and start the services:**
    ```bash
    docker-compose up --build
    ```
    This command builds the Docker images for both the `mcpxy` server and the echo server and starts them in the background. The `mcpxy` server is configured via `docker/config.docker.yaml` to automatically discover the echo server.

3.  **Verify the setup:**
    Once the services are running, you can call the `echo` tool from the `http-echo-server` through the `mcpxy` JSON-RPC API:
    ```bash
    curl -X POST -H "Content-Type: application/json" \
      -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "docker-http-echo/-/echo", "arguments": {"message": "Hello from Docker!"}}, "id": 1}' \
      http://localhost:50050
    ```
    You should receive a response echoing your message.

4.  **Shut down the services:**
    ```bash
    docker-compose down
    ```

## Helm

The Helm chart in the `helm/mcpxy` directory allows you to deploy `mcpxy` to a Kubernetes cluster.

### Prerequisites

- [Kubernetes](https://kubernetes.io/docs/setup/)
- [Helm](https://helm.sh/docs/intro/install/)

### Installation

1.  **Navigate to the root of the repository.**

2.  **Install the chart:**
    ```bash
    helm install mcpxy helm/mcpxy
    ```
    This command deploys `mcpxy` to your Kubernetes cluster with the default configuration.

### Configuration

You can customize the deployment by creating a `values.yaml` file and passing it to the `helm install` command. For example, to change the service type to `LoadBalancer`, you would create a file with the following content:

```yaml
# my-values.yaml
service:
  type: LoadBalancer
```

Then, install the chart with your custom values:
```bash
helm install mcpxy helm/mcpxy -f my-values.yaml
```

For a full list of configurable options, see the `values.yaml` file in the `helm/mcpxy` directory.

### Upgrading the Deployment

To upgrade an existing deployment with new configuration, use the `helm upgrade` command:
```bash
helm upgrade mcpxy helm/mcpxy -f my-values.yaml
```

### Uninstalling the Deployment

To remove the `mcpxy` deployment from your cluster, use the `helm uninstall` command:
```bash
helm uninstall mcpxy
```