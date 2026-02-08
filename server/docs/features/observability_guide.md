# Server Observability Guide

**Goal**: Monitor, debug, and audit your MCP Server instances.

## Pain Point: "Why did this tool fail?"

When a tool execution returns an error or hangs, you need visibility into the request lifecycle.

### 1. Tracing (OpenTelemetry)
Tracing allows you to see the "Waterfall" of a request as it passes through middleware, auth, and reaches the upstream.

**Configuration:**
```yaml
global_settings:
  telemetry:
    traces_exporter: "otlp"
    metrics_exporter: "otlp"
    otlp_endpoint: "http://jaeger:4318" # OTLP HTTP
    service_name: "mcp-server"
```

### 2. Live Logs
The server emits structured JSON logs.
- **Debug**: `log_level: debug` shows raw payloads (Warning: PII risk).
- **Error**: Shows stack traces and upstream connection errors.

## Pain Point: "Who ran this dangerous tool?"

**Audit Logging** provides a tamper-evident trail of who did what and when.

### Usage
Enable audit logging to a secure destination:
```yaml
global_settings:
  audit:
    enabled: true
    output_path: "/var/log/mcp/audit.log"
    storage_type: 1 # STORAGE_TYPE_FILE
    # Or send to external system
```

**What is logged?**
- Actor (User ID / API Key)
- Action (`tool.execute`, `service.register`)
- Status (Success/Deny/Fail)
- Metadata (IP, User Agent)
