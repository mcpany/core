# Monitoring

MCP Any provides built-in observability through a Prometheus metrics endpoint. This allows you to track the performance and health of your MCP server, including request rates, latencies, and error counts for tools and services.

## Configuration

Monitoring is primarily enabled via a command-line flag when starting the server.

- `--metrics-listen-address`: The address to expose Prometheus metrics on (e.g., `:9090`).

If this flag is provided, the server will start a metrics server.

### Service Configuration

While monitoring is enabled globally, the metrics are tagged by **service name** and **tool name**. Therefore, defining meaningful names in your configuration is key to effective monitoring.

```yaml
upstream_services:
  - name: "weather-service" # This name appears in metrics
    http_service:
      address: "https://api.weather.com"
      tools:
        - name: "get_forecast" # This name appears in metrics
```

## Use Case

You want to alert if the error rate of your "weather-service" exceeds 5% or if the P99 latency of "get_forecast" goes above 2 seconds. By enabling the metrics endpoint and scraping it with Prometheus, you can build Grafana dashboards to visualize these metrics and set up Alertmanager rules.

## Available Metrics

### Tool Metrics

- `mcpany_tools_call_total`: Total number of tool calls.
  - Labels: `tool`, `service_id`, `status` (success/error), `error_type`
- `mcpany_tools_call_latency_seconds`: Histogram of tool execution duration in seconds.
  - Labels: `tool`, `service_id`, `status`, `error_type`
- `mcpany_tools_call_input_bytes`: Histogram of tool input size in bytes.
  - Labels: `tool`, `service_id`
- `mcpany_tools_call_output_bytes`: Histogram of tool output size in bytes.
  - Labels: `tool`, `service_id`
- `mcpany_tools_call_tokens_total`: Total number of tokens in tool executions.
  - Labels: `tool`, `service_id`, `direction` (input/output)
- `mcpany_tools_call_in_flight`: Current number of tool executions in flight.
  - Labels: `tool`, `service_id`

### Protocol Metrics

- `mcpany_mcp_operation_duration_seconds`: Duration of MCP operations (e.g., `CallTool`, `ListTools`).
- `mcpany_mcp_operation_total`: Total count of MCP operations.
- `mcpany_mcp_payload_size_bytes`: Size of MCP payloads.
- `mcpany_mcp_operation_tokens_total`: Total tokens in MCP operations.

### Other Metrics

- `mcpany_grpc_connections_opened_total`: Total number of opened gRPC connections.
- `mcpany_grpc_connections_closed_total`: Total number of closed gRPC connections.
- `mcpany_grpc_rpc_started_total`: Total number of started gRPC RPCs.
- `mcpany_grpc_rpc_finished_total`: Total number of finished gRPC RPCs.

## Public API Example

Start the server:

```bash
./mcp-any-server --config config.yaml --metrics-listen-address :9090
```

Scrape the metrics:

```bash
curl http://localhost:9090/metrics
```
