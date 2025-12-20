# Health Status Monitoring

MCP Any provides a built-in Health Status monitoring system that periodically checks the health of all registered upstream services. This status is exposed via the Admin API, allowing administrators and the Dynamic UI to visualize service health in real-time.

## How it works

1.  **Service Registration**: When an upstream service is registered (via config or dynamic registration), it is automatically registered with the Health Manager.
2.  **Health Checks**: The Health Manager runs a background loop (default every 30 seconds) that executes the configured health check for each service.
    *   If a service has a specific `health_check` configuration (e.g., HTTP endpoint, gRPC health check), that check is used.
    *   If no specific check is configured, a default TCP connection check is performed.
3.  **Status Storage**: The result of the health check (Healthy/Unhealthy/Degraded) and any error messages are stored in memory.
4.  **Admin API**: The `ListServices` and `GetService` Admin API methods include the current `ServiceStatus`, `last_error`, and `last_check_time` in their response.

## Configuration

Health checks are configured within the `upstream_services` block in your `config.yaml`.

Example HTTP Service with Health Check:

```yaml
upstream_services:
  - name: "weather-service"
    http_service:
      address: "https://api.weather.com"
      health_check:
        url: "https://api.weather.com/health"
        expected_code: 200
        timeout: "5s"
```

## Admin API Usage

You can retrieve the health status using the Admin gRPC API.

**List Services:**

```protobuf
rpc ListServices(ListServicesRequest) returns (ListServicesResponse);
```

Response includes `ServiceState`:

```protobuf
message ServiceState {
  mcpany.config.v1.UpstreamServiceConfig config = 1;
  ServiceStatus status = 2; // UNKNOWN, HEALTHY, UNHEALTHY, DEGRADED
  string last_error = 3;
  int64 last_check_time = 4;
}
```
