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

- `mcpany_tools_call_total`: Total number of tool calls (global).
- `mcpany_tool_<tool_name>_call_total`: Total number of tool calls for a specific tool.
- `mcpany_tools_call_latency`: Latency of tool calls (global).
- `mcpany_tool_<tool_name>_call_latency`: Latency of tool calls for a specific tool.
- `mcpany_tools_list_total`: Total number of tools/list requests.
- `mcpany_config_reload_total`: Configuration reload events.
- `mcpany_grpc_connections_opened_total`: Total number of opened gRPC connections.
- `mcpany_grpc_connections_closed_total`: Total number of closed gRPC connections.
- `mcpany_grpc_rpc_started_total`: Total number of started gRPC RPCs.
- `mcpany_grpc_rpc_finished_total`: Total number of finished gRPC RPCs.
- `mcpany_tool_execution_total`: Total number of tool executions (labels: tool, service_id, status, error_type).
- `mcpany_tool_execution_duration`: Histogram of tool execution duration.
- `mcpany_tool_execution_input_bytes`: Histogram of input payload sizes in bytes.
- `mcpany_tool_execution_output_bytes`: Histogram of output payload sizes in bytes.

## Public API Example

Start the server:

```bash
./mcp-any-server --config config.yaml --metrics-listen-address :9090
```

Scrape the metrics:

```bash
curl http://localhost:9090/metrics
```
