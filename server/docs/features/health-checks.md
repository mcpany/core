# Health Checks

MCP Any provides a robust mechanism to monitor the health of your upstream services. By configuring health checks, you ensure that traffic is only routed to healthy service instances, improving the overall reliability of your system.

## Supported Protocols

Health checks can be configured for the following upstream service types:

-   **HTTP**: Sends an HTTP request (GET, POST, etc.) and expects a specific status code and optional response body.
-   **gRPC**: Uses the standard [gRPC Health Checking Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md).
-   **WebSocket**: Sends a message and expects a specific response.
-   **WebRTC**: Can perform HTTP or WebSocket checks over the WebRTC channel.
-   **Command Line**: Executes a command and checks the output.
-   **Filesystem**: Checks if configured root paths exist and are accessible (Automatic).

## Configuration

Health checks are defined within the `upstream_services` configuration block.

### HTTP Health Check

```yaml
upstream_services:
  - name: "my-http-service"
    http_service:
      address: "https://api.example.com"
      health_check:
        url: "https://api.example.com/health"
        expected_code: 200
        method: "GET"
        timeout: "5s"
        # interval: "30s" # Currently, the system enforces a global health check interval (default 30s). This field is reserved for future per-service overrides.
```

### gRPC Health Check

```yaml
upstream_services:
  - name: "my-grpc-service"
    grpc_service:
      address: "localhost:50051"
      health_check:
        service: "grpc.health.v1.Health"
        method: "Check"
```

### WebSocket Health Check

```yaml
upstream_services:
  - name: "my-websocket-service"
    websocket_service:
      address: "ws://localhost:8080"
      health_check:
        url: "ws://localhost:8080/health"
        message: "ping"
        expected_response_contains: "pong"
```

### Command Line Health Check

```yaml
upstream_services:
  - name: "my-cli-service"
    command_line_service:
      command: "python3"
      health_check:
        method: "-c"
        prompt: "print('alive')"
        expected_response_contains: "alive"
```

### Filesystem Health Check

Filesystem health checks are enabled automatically for local filesystem services. They verify that all configured `root_paths` exist and are accessible.

```yaml
upstream_services:
  - name: "my-files"
    filesystem_service:
      root_paths:
        "/data": "/var/lib/data"
      # Health check is automatic
```

## Monitoring

The server runs a background health check loop (default every 30 seconds) for all registered services.

-   **Real-time Status**: Visible in the Dashboard.
-   **History**: A rolling history of health status is maintained in-memory for visualization.
-   **On-Demand**: Health checks can also be triggered manually via the Connection Diagnostics tool.
-   **Metrics**: Health check status and latency are exported via Prometheus metrics.
