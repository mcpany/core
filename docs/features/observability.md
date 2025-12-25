# Observability

MCP Any includes comprehensive observability features to monitor and debug your server.

## Features

*   **Distributed Tracing**: OpenTelemetry support for tracing requests across services.
*   **Metrics**: Prometheus-compatible metrics for monitoring performance and health.
*   **Structured Logging**: JSON logs for easy parsing and analysis.
*   **Audit Logging**: detailed logs of all actions for compliance and security.

## Configuration

Observability settings are configured in the global `observability` and `audit` sections of the configuration.

### Metrics

To enable metrics, set the `metrics-listen-address` flag or environment variable `MCPANY_METRICS_LISTEN_ADDRESS` (e.g., `:9090`).

### Audit Logging

Audit logging can be configured in the global settings.

```yaml
global:
  audit:
    enabled: true
    output_path: "/var/log/mcpany/audit.log"
    storage_type: "FILE" # or "SQLITE"
    log_arguments: true
    log_results: false
```

*   `enabled`: Enable or disable audit logging.
*   `output_path`: Path to the log file or SQLite database.
*   `storage_type`: `FILE` or `SQLITE`.
*   `log_arguments`: Whether to include tool arguments in the log (careful with sensitive data).
*   `log_results`: Whether to include tool results in the log.
