# Distributed Tracing

MCP Any supports [OpenTelemetry](https://opentelemetry.io/) for distributed tracing. This allows you to trace requests as they flow from the MCP client, through the MCP Any server, to the upstream services.

## Configuration

Tracing is configured using standard OpenTelemetry environment variables.

### Supported Exporters

- `otlp`: Exports traces via OTLP/HTTP.
- `stdout`: Prints traces to stderr. Useful for debugging and development.
- `none`: Disables tracing (default behavior if no specific configuration is provided).

### Enabling OTLP Tracing

To enable OTLP tracing, set the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable. The server uses the OTLP/HTTP exporter.

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
./mcp-any-server run --config config.yaml
```

You can also specify headers using `OTEL_EXPORTER_OTLP_HEADERS`.

### Enabling Stdout Tracing

To enable stdout tracing, set `OTEL_TRACES_EXPORTER=stdout`.

```bash
OTEL_TRACES_EXPORTER=stdout ./mcp-any-server run --config config.yaml
```

You will see trace output in the logs (stderr).

## Spans

The server instruments the following operations:
- Incoming HTTP requests (if using the HTTP server mode).
- Outgoing HTTP requests to upstream services.

Each span includes standard OpenTelemetry attributes.
