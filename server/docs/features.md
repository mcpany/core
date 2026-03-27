# MCP Any Server Features

This section details the capabilities of the MCP Any Server.

## Core Config & Operations
- [Configuration Guide](reference/configuration.md) - Best practices for setting up `config.yaml`.
- [Service Types](features/service-types.md) - Deep dive into HTTP, gRPC, and Stdio upstreams.
- [Security](features/security.md) - Authentication, DLP, and Secrets.
- [Dynamic Registration](features/dynamic_registration.md) - Adding services at runtime.

## Observability & Debugging
- [Audit Logging](features/audit_logging.md) - Compliance and activity tracking.
- [Tracing](features/tracing/) - Distributed request tracing with OpenTelemetry.
- [Debugger](features/debugger.md) - Inspecting traffic and replaying requests.
- [Health Checks](features/health-checks.md) - Monitoring uptime.

## Middleware & Resilience
- [Rate Limiting](features/rate-limiting/) - Protecting your backend.
- [Context Optimization](features/context_optimizer.md) - Managing token usage.
- [DLP](features/dlp.md) - Redacting sensitive data.

## Advanced
- [WASM Plugins](features/wasm.md) - Extending server logic.
- [Message Bus](features/message_bus.md) - Event-driven integrations.
