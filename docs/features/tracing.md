# Distributed Tracing

MCP Any supports Distributed Tracing via OpenTelemetry (OTLP). This allows you to trace requests as they flow through the MCP Any server to your upstream services.

## Configuration

To enable tracing, add the `tracing` block to your `global_settings` in the configuration file.

```yaml
global_settings:
  tracing:
    enabled: true
    endpoint: "localhost:4317" # The OTLP gRPC endpoint of your collector
    insecure: true # Set to true if your collector does not use TLS
```

## How it works

1.  **Incoming Requests**: When `mcpany` receives a request (via HTTP or gRPC), it extracts the trace context from the headers (if present) or starts a new trace.
2.  **Internal Processing**: Spans are created for internal operations.
3.  **Outgoing Requests**: When `mcpany` calls an upstream service (via HTTP or gRPC), it injects the trace context into the request headers.

## Supported Exporters

Currently, only the **OTLP gRPC** exporter is supported. This allows you to send traces to any OTLP-compatible backend, such as:

-   Jaeger
-   Zipkin (via OpenTelemetry Collector)
-   Prometheus (via OpenTelemetry Collector)
-   Datadog, Honeycomb, etc. (via OpenTelemetry Collector)
