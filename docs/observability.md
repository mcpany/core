# Observability

MCP Any provides observability features that allow you to monitor the health and performance of the server and its upstream services. This is achieved by integrating the [OpenTelemetry](https://opentelemetry.io/) SDK, which allows you to collect and export metrics and traces to various observability backends.

## Metrics

MCP Any exposes a Prometheus-compatible metrics endpoint at `/metrics`. This endpoint provides a variety of metrics that can be used to monitor the server, including:

*   `mcp_server_requests`: The total number of MCP requests received, with labels for the method and status.
*   `mcp_server_latency`: The latency of MCP requests, with labels for the method and status.

### Configuration

The metrics server can be configured using the following command-line flags or environment variables:

| Flag                     | Environment Variable        | Description                               | Default   |
| ------------------------ | --------------------------- | ----------------------------------------- | --------- |
| `--metrics-listen-address` | `MCPANY_METRICS_LISTEN_ADDRESS` | The listen address for the metrics server. | `:9090`   |

### Scraping with Prometheus

To scrape the metrics with Prometheus, you can add the following job to your `prometheus.yml` configuration file:

```yaml
scrape_configs:
  - job_name: 'mcpany'
    static_configs:
      - targets: ['localhost:9090']
```

This will configure Prometheus to scrape the metrics from the `/metrics` endpoint every 15 seconds. You can then use Grafana or other visualization tools to create dashboards and alerts based on these metrics.
