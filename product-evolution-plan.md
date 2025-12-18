# Product Evolution Plan

This document outlines the strategic plan for synchronizing the project's documentation with reality and driving future innovation.

## 1. Updated Roadmap

### Current Status

#### Implemented Features

The following features are fully implemented and tested:

- **Service Types**:
  - gRPC (with reflection)
  - HTTP
  - OpenAPI
  - GraphQL
  - Stdio
  - MCP-to-MCP Proxy
  - WebSocket
- **Upstream Authentication**:
  - API Key
  - Bearer Token
  - OAuth 2.0
- **Dynamic Registration**: Services can be registered at runtime via the gRPC Registration API.
- **Static Registration**: Services can be registered at startup via a YAML or JSON configuration file.
- **Advanced Service Policies**:
  - Caching (`CacheConfig`)
  - Rate Limiting (`RateLimitConfig`)
  - Resilience (Circuit Breakers & Retries) (`ResilienceConfig`)
- **Deployment**:
  - Official Helm Chart
  - Docker Container
- [x] **Health Checks**: Implement health check endpoints for upstream services (HTTP, gRPC, WebSocket, WebRTC, Command Line).
- [x] **Schema Validation**: Integrate JSON Schema to validate configuration files before loading.
- [x] **Service Profiles**: Categorize and selectively enable services using profiles (`--profiles` flag).
- **Configuration**:
  - Hot Configuration Reloading
- [x] [**Secrets Management**](./docs/features/security.md): Secure handling of sensitive data (API keys, passwords) using Vault, AWS Secrets Manager, or Env Vars.
- [x] [**Distributed Tracing**](./docs/features/tracing/README.md): Integrate OpenTelemetry for tracing requests across services.
- [x] [**Transport Protocols (NATS)**](./docs/features/nats.md): Support for NATS as a message bus.
- [x] [**Automated Documentation Generation**](./docs/features/documentation_generation.md): Generate markdown documentation for registered tools directly from the configuration.
- [x] [**IP Allowlisting**](./docs/features/security.md): Restrict access to specific IP addresses/CIDRs.

#### Configured but Not Yet Implemented

The following features are defined in the configuration schema (`proto/config/v1/config.proto`) but are **not yet implemented** in the server logic:

- **Advanced Authentication**:
  - [x] Incoming request authentication (Profile > User > Global Priority)

### High Priority (Next 1-3 Months)

- [ ] **Dynamic UI**: Build a web-based UI for managing upstream services dynamically.
- [ ] **RBAC**: Role-Based Access Control for managing user permissions.
- [ ] **Enhanced Metrics**: Provide more granular metrics for tool usage, performance, and error rates.
- [ ] **Security Policies**: Implement fine-grained request validation policies (runtime argument validation).
- [ ] **WASM Plugin Support**: Allow extending functionality using WebAssembly plugins for custom logic.
- [ ] **Transport Protocols (Kafka)**: Add support for asynchronous communication via Kafka.
- [ ] **Client SDKs**: Develop official Client SDKs (Go, Python, TS) to interact with MCP Any programmatically.

## 2. Top 10 Recommended Features

Based on the strategic feature extraction analysis, the following features are recommended for immediate implementation to address security, scalability, and UX gaps.

| Rank | Feature Name | Why it matters | Implementation Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **OTLP Exporter Support** | **Observability**: Currently, tracing only supports `stdout`. Enabling OTLP allows integration with standard tools like Jaeger/Honeycomb for real cross-service visibility. | Low |
| 2 | **Distributed Rate Limiting (Redis)** | **Scalability**: Current rate limiting is in-memory (per instance). Distributed limiting is essential for running multiple replicas in production. | Medium |
| 3 | **Runtime Argument Validation** | **Security**: Completes the "Security Policies" feature by enforcing regex checks on tool arguments at runtime, preventing injection attacks or misuse. | Medium |
| 4 | **Kafka Transport Support** | **Scalability**: Completes the message bus vision. Kafka is industry-standard for high-throughput event streaming. | Medium |
| 5 | **Audit Logging** | **Security/Compliance**: Record *who* accessed *what* tool and *when*. Critical for enterprise adoption. | Medium |
| 6 | **Admin/Management API** | **UX/Automation**: Allow programmatic configuration updates (add/remove services) without restarting or editing files directly. Precursor to the Dynamic UI. | High |
| 7 | **Structured Logging (JSON)** | **Observability**: Switch from text logs to JSON for better ingestion by log aggregators (Splunk, ELK). | Low |
| 8 | **Webhooks for Tool Events** | **Extensibility**: Trigger external systems (Slack, PagerDuty) when specific tools are called or fail. | Medium |
| 9 | **RBAC (Role-Based Access Control)** | **Security**: Restrict which users can access which profiles or tools. Essential for multi-tenant environments. | High |
| 10 | **WASM Plugins** | **Extensibility**: Allow users to write custom transformation/validation logic in any language (Rust, Go, TS) without recompiling the server. | High |

## 3. Codebase Health Report

This section highlights areas of the codebase that require attention to ensure long-term maintainability and stability.

### Critical Areas

*   **Telemetry Configuration**: The `pkg/telemetry/tracing.go` file hardcodes the `stdout` exporter. While the OTLP libraries are present in `go.mod`, they are not wired up. This limits observability in production environments.
*   **Security Policy Implementation**: The `shouldBlock` function in `pkg/upstream/http/http.go` contains a comment indicating that argument regex matching cannot be done at registration time. This confirms that runtime validation is missing, leaving a gap in the security policy feature.
*   **Tool Discovery Complexity**: The `createAndRegisterHTTPTools` function in `pkg/upstream/http/http.go` has high cyclomatic complexity (`//nolint:gocyclo`). This makes it difficult to test and maintain. It should be refactored into smaller, testable components.
*   **Testing Gaps**: While there are test files, "fine-grained request validation" and other edge cases in policy enforcement need rigorous integration tests to ensure security guarantees are met.

### Recommendations

1.  **Refactor Tool Registration**: Break down `createAndRegisterHTTPTools` and similar functions in other upstreams.
2.  **Wire up OTLP**: Modify `pkg/telemetry` to respect `OTEL_EXPORTER_OTLP_ENDPOINT` and initialize the OTLP exporter.
3.  **Implement Runtime Interceptors**: Use a middleware or interceptor pattern to handle security policies (argument validation) at request time, rather than trying to do it at registration time.
