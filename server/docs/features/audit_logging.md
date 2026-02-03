# Audit Logging

MCP Any provides a built-in audit logging capability to record details about every tool execution. This is essential for compliance, security monitoring, and debugging.

## Overview

When enabled, the audit logger intercepts every tool execution request and records structured data about the event to a specified file.

The audit log captures:
- **Timestamp**: When the execution started.
- **Tool Name**: The name of the tool being executed.
- **User ID**: The ID of the authenticated user (if available).
- **Profile ID**: The ID of the active profile (if available).
- **Duration**: How long the execution took.
- **Arguments**: The input arguments provided to the tool (optional).
- **Result**: The result returned by the tool (optional).
- **Error**: Any error that occurred during execution.

## Configuration

Audit logging is configured in the `GlobalSettings` section of your MCP Any configuration file.

### Example Configuration

```yaml
global_settings:
  audit:
    enabled: true
    output_path: "audit.log"
    log_arguments: true
    log_results: false
```

### Webhook Configuration

You can also configure audit logs to be sent to a webhook endpoint.

```yaml
global_settings:
  audit:
    enabled: true
    storage_type: "STORAGE_TYPE_WEBHOOK"
    webhook_url: "https://audit-collector.example.com/v1/logs"
    webhook_headers:
      Authorization: "Bearer my-token"
      X-Environment: "production"
    log_arguments: true
```

### Splunk Configuration

```yaml
global_settings:
  audit:
    enabled: true
    storage_type: "STORAGE_TYPE_SPLUNK"
    splunk:
      hec_url: "https://splunk.example.com:8088/services/collector/event"
      token: "your-hec-token"
      source: "mcpany"
      sourcetype: "_json"
      index: "main"
```

### Datadog Configuration

```yaml
global_settings:
  audit:
    enabled: true
    storage_type: "STORAGE_TYPE_DATADOG"
    datadog:
      api_key: "your-datadog-api-key"
      site: "datadoghq.com" # or datadoghq.eu, etc.
      service: "mcpany-prod"
      tags:
        env: "production"
        region: "us-east-1"
```

### Configuration Options

| Option | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `enabled` | `bool` | `false` | Enables or disables audit logging. |
| `storage_type` | `enum` | `FILE` | The storage backend to use: `FILE`, `SQLITE`, `POSTGRES`, `WEBHOOK`, `SPLUNK`, or `DATADOG`. |
| `output_path` | `string` | `""` | Path to the log file (for `FILE`) or database connection string/path (for `SQLITE`/`POSTGRES`). |
| `webhook_url` | `string` | `""` | The URL to send POST requests to (for `WEBHOOK`). |
| `webhook_headers` | `map` | `{}` | HTTP headers to include in the webhook request (for `WEBHOOK`). |
| `splunk` | `object` | `nil` | Configuration for Splunk HEC (for `SPLUNK`). |
| `datadog` | `object` | `nil` | Configuration for Datadog Logs (for `DATADOG`). |
| `log_arguments` | `bool` | `false` | If true, logs the input arguments. **Warning:** May log sensitive data. |
| `log_results` | `bool` | `false` | If true, logs the execution result. **Warning:** May log sensitive data. |

**Note on Webhook Performance:** The webhook storage implementation uses an **asynchronous, buffered worker pool**. Tool execution is not blocked by webhook latency. Logs are batched (default batch size: 10) or flushed periodically (every 1 second) by background workers.

## Log Format

Audit logs are written as newline-delimited JSON (NDJSON). Each line represents a single tool execution event.

### Example Log Entry

```json
{
  "timestamp": "2023-10-27T10:00:00.123Z",
  "tool_name": "weather_get_forecast",
  "user_id": "alice",
  "profile_id": "prod",
  "duration": "150ms",
  "duration_ms": 150,
  "arguments": {
    "city": "London"
  },
  "result": {
    "temperature": 15,
    "unit": "C"
  }
}
```

## Security Considerations

- **Sensitive Data**: By default, `log_arguments` and `log_results` are disabled. Enable them with caution, as they may expose API keys, PII, or other sensitive information handled by your tools.
- **File Permissions**: Ensure that the `output_path` is writable by the MCP Any server process and readable only by authorized personnel.
