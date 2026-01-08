# Health Checks

MCP Any provides a robust mechanism to monitor the health of your upstream services and the MCP Any server itself.

## Server Health Endpoints

MCP Any exposes the following endpoints for monitoring the server's own health:

-   `/healthz`: Liveness probe. Returns 200 OK if the server is running.
-   `/readyz`: Readiness probe. Returns 200 OK if the server is ready to accept traffic (e.g. connected to database, message bus).
-   `/v1/health/doctor`: **Doctor API**. Runs a suite of diagnostic checks (internet connectivity, dependencies) and returns a detailed JSON report.

## Upstream Health Checks

By configuring health checks for upstream services, you ensure that traffic is only routed to healthy service instances, improving the overall reliability of your system.

### Supported Protocols

Health checks can be configured for the following upstream service types:

-   **HTTP**: Sends an HTTP request (GET, POST, etc.) and expects a specific status code and optional response body.
-   **gRPC**: Uses the standard [gRPC Health Checking Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md).
-   **WebSocket**: Sends a message and expects a specific response.
-   **WebRTC**: Can perform HTTP or WebSocket checks over the WebRTC channel.
-   **Command Line**: Executes a command and checks the output.

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
        interval: "30s"
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

## Monitoring

Health check status is logged and can be monitored via the metrics exported by the server. When a service fails its health check, it is marked as unhealthy, and requests may fail or be routed to other instances (if load balancing is configured).
