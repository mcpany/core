# Distributed Tracing

MCP Any supports [OpenTelemetry](https://opentelemetry.io/) for distributed tracing. This allows you to trace requests as they flow from the MCP client, through the MCP Any server, to the upstream services.

## Configuration

Tracing is configured using standard OpenTelemetry environment variables.

To enable tracing, set the `OTEL_TRACES_EXPORTER` environment variable.

### Supported Exporters

- `stdout`: Prints traces to stderr. Useful for debugging and development.
- `none`: Disables tracing (default behavior if no specific configuration is provided).

**Note:** Currently, `stdout` is the primary supported exporter for demonstration purposes. OTLP support is planned.

### Example

Start the server with stdout tracing enabled:

```bash
OTEL_TRACES_EXPORTER=stdout ./mcp-any-server run --config config.yaml
```

You will see trace output in the logs (stderr).

## Spans

The server instruments the following operations:
- Incoming HTTP requests (if using the HTTP server mode).
- Outgoing HTTP requests to upstream services.

Each span includes standard OpenTelemetry attributes.
