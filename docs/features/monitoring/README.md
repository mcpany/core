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

The following metrics are exposed by the server. Note that the prefix `mcpany_` is applied by default.

| Metric Name | Type | Labels | Description |
| :--- | :--- | :--- | :--- |
| `mcpany_tools_call_total` | Counter | `tool`, `service_id` | Total number of tool calls initiated. |
| `mcpany_tools_call_errors` | Counter | `tool`, `service_id` | Total number of tool calls that resulted in an error. |
| `mcpany_tools_call_latency` | Histogram | `tool`, `service_id` | Latency of tool calls (in seconds/microseconds depending on sink config). |
| `mcpany_tools_list_total` | Counter | none | Total number of tools/list requests received. |
| `mcpany_config_reload_total` | Counter | none | Total number of configuration reload events. |
| `mcpany_config_reload_errors` | Counter | none | Total number of configuration reload failures. |
| `mcpany_grpc_connections_opened_total` | Counter | none | Total number of opened gRPC connections. |
| `mcpany_grpc_connections_closed_total` | Counter | none | Total number of closed gRPC connections. |
| `mcpany_grpc_rpc_started_total` | Counter | none | Total number of started gRPC RPCs. |
| `mcpany_grpc_rpc_finished_total` | Counter | none | Total number of finished gRPC RPCs. |
| `mcpany_http_connections_opened_total` | Counter | none | Total number of opened HTTP connections. |
| `mcpany_http_connections_closed_total` | Counter | none | Total number of closed HTTP connections. |

## Public API Example

Start the server:

```bash
./mcp-any-server --config config.yaml --metrics-listen-address :9090
```

Scrape the metrics:

```bash
curl http://localhost:9090/metrics
```
