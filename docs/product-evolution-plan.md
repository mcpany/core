# Product Evolution Plan

## 1. Updated Roadmap

### Current Status

#### Implemented Features

The following features are fully implemented and tested:

- **Service Types**: gRPC, HTTP, OpenAPI, GraphQL, Stdio, MCP-to-MCP Proxy, WebSocket, WebRTC.
- **Upstream Authentication**: API Key, Bearer Token, OAuth 2.0.
- **Registration**: Dynamic (gRPC) and Static (YAML/JSON).
- **Policies**: Caching, Rate Limiting (Memory & Redis), Resilience (Circuit Breakers & Retries).
- **Deployment**: Helm Chart, Docker.
- **Observability**: Distributed Tracing (OpenTelemetry), Metrics, Structured Logging, Audit Logging.
- **Security**: Secrets Management, IP Allowlisting, Webhooks, Fine-grained Policies.
- **Message Bus**: NATS support.
- **Advanced Authentication**: Priority-based (Profile > User > Global).

#### High Priority (Next 1-3 Months)

- **Dynamic UI**: Build a web-based UI for managing upstream services dynamically.
- **RBAC**: Role-Based Access Control for managing user permissions.
- **Transport Protocols (Kafka)**: Add support for asynchronous communication via Kafka.

#### Long-Term Goals

- **WASM Plugin Support**: Extensibility via WebAssembly.
- **MCP Any Config Registry**: Public registry for configurations.
- **Client SDKs**: Official libraries for Go, Python, TS.

---

## 2. Top 10 Recommended Features

We have identified the following features as critical for the next phase of product evolution, focusing on Enterprise Readiness and Developer Experience.

| Rank | Feature Name | Why it matters | Implementation Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **RBAC (Role-Based Access Control)** | **Security**: Essential for multi-tenant environments to restrict access to specific profiles or tools based on user roles. | High |
| 2 | **SQL Database Provider** | **Utility**: Allow users to directly expose SQL queries (Postgres, MySQL) as MCP tools without writing intermediate API code. | High |
| 3 | **Admin Management API** | **Automation**: Expand the Admin API to support full CRUD operations on services/config at runtime, enabling the Dynamic UI. | Medium |
| 4 | **Dynamic Web UI** | **UX**: A visual dashboard to monitor health, view metrics, and manage configurations without editing YAML. | High |
| 5 | **Kafka Transport** | **Scalability**: Support event-driven architectures where tools can be triggered by Kafka messages. | Medium |
| 6 | **WASM Plugins** | **Extensibility**: Allow users to deploy safe, sandboxed custom logic for transformations or validations. | High |
| 7 | **File System Provider** | **Utility**: safe, controlled access to the local file system (read/write/list) as an MCP tool source. | Medium |
| 8 | **Cost & Quota Management** | **Governance**: Track token usage or call counts per user/profile and enforce strict quotas (beyond rate limiting). | Medium |
| 9 | **Client SDKs (Python/TS)** | **DX**: Provide idiomatic wrappers for connecting to MCP Any, handling authentication, and parsing responses. | Medium |
| 10 | **Structured Output Transformation** | **DX**: Native support for JQ or JSONPath to transform complex upstream API responses before sending them to the LLM. | Low |

---

## 3. Codebase Health Report

This section highlights areas of the codebase that require refactoring or attention to ensure stability and maintainability.

### 1. High Complexity in Tool Discovery
The function `createAndRegisterHTTPTools` in `pkg/upstream/http/http.go` has identified high cyclomatic complexity. This function handles too many responsibilities (parsing, validation, registration) and should be refactored into smaller, testable sub-components.

### 2. Skeletal Admin API
The current `pkg/admin` package implements only `ClearCache`. To support the "Dynamic UI" roadmap item, this API needs to be significantly expanded to expose internal state, configuration management, and health status.

### 3. Redis Rate Limiter Time Synchronization
The `RedisLimiter` (in `pkg/middleware/ratelimit_redis.go`) relies on client-side timestamps passed to a Lua script. While functional, this can lead to inaccuracies if server clocks drift. Future improvements should consider using Redis server time or robust drift mitigation.

### 4. Dependency Management (Google APIs)
The project appears to vendor Google API definitions in `build/googleapis`. This manual management can lead to version drift and maintenance overhead. Moving to a managed dependency or automated generation pipeline is recommended.
