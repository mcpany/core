# MCP Any Server Features

This section details the capabilities of the MCP Any Server.

## Core Config & Operations
- [Configuration Guide](configuration_guide.md) - Best practices for setting up `config.yaml`.
- [Service Types](service-types.md) - Deep dive into HTTP, gRPC, and Stdio upstreams.
- [Security](security.md) - Authentication, DLP, and Secrets.
- [Dynamic Registration](dynamic_registration.md) - Adding services at runtime.

## Observability & Debugging
- [Audit Logging](audit_logging.md) - Compliance and activity tracking.
- [Tracing](tracing/README.md) - Distributed request tracing with OpenTelemetry.
- [Debugger](debugger.md) - Inspecting traffic and replaying requests.
- [Health Checks](health-checks.md) - Monitoring uptime.

## Middleware & Resilience
- [Rate Limiting](rate-limiting/README.md) - Protecting your backend.
- [Context Optimization](context_optimizer.md) - Managing token usage.
- [DLP](dlp.md) - Redacting sensitive data.

## Advanced
- [WASM Plugins](wasm.md) - Extending server logic.
- [Message Bus](message_bus.md) - Event-driven integrations.
