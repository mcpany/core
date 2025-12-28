# Health Checks

MCP Any provides a robust health check system to ensure upstream services are available before routing requests to them.

## Features

- **Proactive Monitoring**: periodically checks the health of upstream services.
- **Circuit Breaking**: Automatically disables unhealthy services to prevent cascading failures.
- **Multiple Protocols**: Supports HTTP, gRPC, TCP, and command-line health checks.

For detailed configuration options, see [Health Checks](../../server/docs/features/health-checks.md).
