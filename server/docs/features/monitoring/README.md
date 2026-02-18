# Monitoring & Metrics

MCP Any exposes comprehensive metrics in Prometheus format to help you monitor the health and performance of your server.

## Endpoint

Metrics are exposed at the `/metrics` endpoint by default.

```bash
curl http://localhost:8081/metrics
```

> **Note**: The metrics server port is configurable via `--metrics-listen-address` or `MCPANY_METRICS_LISTEN_ADDRESS`.

## Key Metrics

### Tool Execution Metrics

These metrics provide deep insights into tool usage, performance, and token consumption.

| Metric Name | Type | Labels | Description |
| :--- | :--- | :--- | :--- |
| `mcpany_tools_call_total` | Counter | `tool`, `service_id`, `status` (success/error), `error_type` | Total number of tool executions. |
| `mcpany_tools_call_latency_seconds` | Histogram | `tool`, `service_id`, `status`, `error_type` | Latency distribution of tool executions. |
| `mcpany_tools_call_tokens_total` | Counter | `tool`, `service_id`, `direction` (input/output) | Total tokens consumed by tool inputs and outputs. |
| `mcpany_tools_call_in_flight` | Gauge | `tool`, `service_id` | Current number of tool executions in progress. |
| `mcpany_tools_call_input_bytes` | Histogram | `tool`, `service_id` | Size of tool inputs in bytes. |
| `mcpany_tools_call_output_bytes` | Histogram | `tool`, `service_id` | Size of tool outputs in bytes. |

### System & Middleware Metrics

These metrics track the overall health and request flow of the server.

| Metric Name | Type | Description |
| :--- | :--- | :--- |
| `mcpany_middleware_request_total` | Counter | Total number of requests processed by the middleware chain. |
| `mcpany_middleware_request_latency` | Summary | Latency of request processing in the middleware chain. |
| `mcpany_middleware_request_error` | Counter | Total number of failed requests. |
| `mcpany_config_reload_total` | Counter | Total number of configuration reload attempts. |
| `mcpany_config_reload_errors` | Counter | Total number of failed configuration reloads. |

## Integration with Prometheus

To scrape these metrics, add a job to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'mcpany'
    static_configs:
      - targets: ['localhost:8081']
```

## Dashboard

You can visualize these metrics using Grafana. A sample dashboard JSON is available in the `examples/monitoring/grafana` directory (if available).
