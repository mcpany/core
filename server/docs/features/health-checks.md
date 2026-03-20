# Health Checks

MCP Any provides a robust mechanism to monitor the health of your upstream services. By configuring health checks, you ensure that traffic is only routed to healthy service instances, improving the overall reliability of your system.

## Supported Protocols

Health checks can be configured for the following upstream service types:

-   **HTTP**: Sends an HTTP request (GET, POST, etc.) and expects a specific status code and optional response body.
-   **gRPC**: Uses the standard [gRPC Health Checking Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md).
-   **WebSocket**: Sends a message and expects a specific response.
-   **WebRTC**: Can perform HTTP or WebSocket checks over the WebRTC channel.
-   **MCP Service**: Checks connectivity to the upstream MCP service (HTTP or Stdio).
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
        interval: "30s" # Optional: Interval for background health checks.
        timeout: "5s"
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
        interval: "10s"
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
        interval: "15s"
```

### WebRTC Health Check

```yaml
upstream_services:
  - name: "my-webrtc-service"
    webrtc_service:
      address: "http://localhost:8080/signal"
      health_check:
        # WebRTC checks rely on the signaling channel (HTTP or WebSocket)
        http:
          url: "http://localhost:8080/health"
          expected_code: 200
```

### MCP Service Health Check

```yaml
upstream_services:
  - name: "my-mcp-service"
    mcp_service:
      http_connection:
        http_address: "http://localhost:8000/mcp"
      # Health check implicitly checks the connection
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
        interval: "60s"
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

Health check status is logged and can be monitored via the metrics exported by the server. When a service fails its health check (performed on-demand, e.g., via diagnostics or at startup), it is marked as unhealthy, and requests may fail or be routed to other instances (if load balancing is configured).
