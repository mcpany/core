# Monitoring and Metrics

MCP Any provides detailed metrics to help you monitor the health and performance of your server and the tools it exposes.

## Prometheus Metrics

The server exposes Prometheus-compatible metrics at the `/metrics` endpoint. To enable this, you must configure the `metrics-listen-address` in your configuration or via command-line flags.

### Configuration

**Via CLI Flag:**

```bash
mcpany run --metrics-listen-address :9090 ...
```

**Via Environment Variable:**

```bash
export MCPANY_METRICS_LISTEN_ADDRESS=:9090
mcpany run ...
```

**Via Config File:**

```yaml
global:
  metrics_listen_address: ":9090"
```

### Available Metrics

#### Tool Execution Metrics

These metrics provide insights into the usage, performance, and reliability of your exposed tools.

| Metric Name | Type | Description | Labels |
| :--- | :--- | :--- | :--- |
| `mcp_tool_executions_total` | Counter | Total number of tool executions. | `tool`, `service_id`, `status` (`success`/`error`), `error_type` |
| `mcp_tool_execution_duration_seconds` | Histogram | Duration of tool execution in seconds. | `tool`, `service_id`, `status`, `error_type` |
| `mcp_tool_execution_input_bytes` | Histogram | Size of tool input arguments in bytes. | `tool`, `service_id` |
| `mcp_tool_execution_output_bytes` | Histogram | Size of tool output result in bytes. | `tool`, `service_id` |

**Labels:**

*   `tool`: The name of the tool being executed (e.g., `get_weather`).
*   `service_id`: The ID of the service the tool belongs to.
*   `status`: The outcome of the execution, either `success` or `error`.
*   `error_type`: A more granular classification of errors:
    *   `none`: Success.
    *   `execution_failed`: General execution failure.
    *   `context_canceled`: The request was canceled by the client or server shutdown.
    *   `deadline_exceeded`: The execution timed out.

### Go Runtime Metrics

Standard Go runtime metrics (goroutines, memory usage, GC, etc.) are also exposed by default via the Prometheus Go client.

## Integration with Prometheus

To scrape these metrics with Prometheus, add a job to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'mcpany'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:9090']
```

## Grafana Dashboard

You can visualize these metrics in Grafana. Key charts to create include:

1.  **Tool Request Rate**: `rate(mcp_tool_executions_total[5m])` by `tool`.
2.  **Error Rate**: `rate(mcp_tool_executions_total{status="error"}[5m]) / rate(mcp_tool_executions_total[5m])`.
3.  **Latency P95**: `histogram_quantile(0.95, rate(mcp_tool_execution_duration_seconds_bucket[5m]))`.
4.  **Traffic Volume**: `rate(mcp_tool_execution_input_bytes_sum[5m])` and `rate(mcp_tool_execution_output_bytes_sum[5m])`.
