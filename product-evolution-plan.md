# Product Evolution Plan

## 1. Updated Roadmap

### Current Status

#### Implemented Features

The following features are fully implemented and tested:

- **Service Types**: gRPC, HTTP, OpenAPI, GraphQL, Stdio, MCP-to-MCP Proxy, WebSocket, WebRTC, **SQL**.
- **Upstream Authentication**: API Key, Bearer Token, OAuth 2.0.
- **Registration**: Dynamic (gRPC) and Static (YAML/JSON).
- **Policies**: Caching, Rate Limiting (Memory & Redis), Resilience (Circuit Breakers & Retries).
- **Deployment**: Helm Chart, Docker.
- **Observability**: Distributed Tracing (OpenTelemetry), Metrics, Structured Logging, Audit Logging.
- **Security**: Secrets Management (Env, AWS Secrets Manager), IP Allowlisting, Webhooks (Pre/Post/Transform), Fine-grained Policies.
- **Message Bus**: NATS and Kafka support.
- **Transformation**: JQ, JSONPath, and Go Template support for input/output transformation.

#### High Priority (Next 1-3 Months)

- **Dynamic Web UI**: Build a web-based UI for managing upstream services dynamically.
- **RBAC**: Role-Based Access Control for managing user permissions.
- **Admin Management API**: Expand to support CRUD operations (currently Read-Only/ClearCache).

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
| 2 | **Dynamic Web UI** | **UX**: A visual dashboard to monitor health, view metrics, and manage configurations without editing YAML. | High |
| 3 | **Admin Management API (CRUD)** | **Automation**: Expand the Admin API to support full CRUD operations on services/config at runtime (currently mostly read-only). | Medium |
| 4 | **WASM Plugins** | **Extensibility**: Allow users to deploy safe, sandboxed custom logic for transformations or validations. | High |
| 5 | **File System Provider** | **Utility**: Safe, controlled access to the local file system (read/write/list) as an MCP tool source. | Medium |
| 6 | **Cost & Quota Management** | **Governance**: Track token usage or call counts per user/profile and enforce strict quotas (beyond rate limiting). | Medium |
| 7 | **Client SDKs (Python/TS)** | **DX**: Provide idiomatic wrappers for connecting to MCP Any, handling authentication, and parsing responses. | Medium |
| 8 | **Playground** | **UX**: An interactive web-based playground to test tools and query logic (similar to GraphiQL). | Medium |
| 9 | **Config Versioning & History** | **Ops**: Track changes to configurations over time, allowing rollback and audit trails of *who* changed *what*. | Medium |
| 10 | **Advanced Secrets Integration** | **Security**: Integration with Vault or Azure KeyVault for enterprise-grade secret management beyond AWS/Env. | High |

---

## 3. Codebase Health Report

This section highlights areas of the codebase that require refactoring or attention to ensure stability and maintainability.

### 1. High Complexity in Tool Discovery
The function `createAndRegisterHTTPTools` in `pkg/upstream/http/http.go` has identified high cyclomatic complexity (flagged with `//nolint:gocyclo`). This function handles too many responsibilities (parsing, validation, registration) and should be refactored into smaller, testable sub-components.

### 2. Admin API Limitations
The current `pkg/admin` package implements `ClearCache`, `ListServices`, `GetService`, `ListTools`, and `GetTool`. It lacks **Create**, **Update**, and **Delete** operations, which are prerequisites for the "Dynamic UI".

### 3. Redis Rate Limiter Time Synchronization
The `RedisLimiter` (in `pkg/middleware/ratelimit_redis.go`) relies on client-side timestamps (`now = timeNow().UnixMicro()`) passed to a Lua script. This can lead to inaccuracies if server clocks drift. Future improvements should consider using Redis server time or robust drift mitigation.

### 4. Dependency Management (Google APIs)
The project vendors Google API definitions in `build/googleapis`, which is approximately **105MB**. This manual management leads to repo bloat and version drift. Moving to a managed dependency or automated generation pipeline is highly recommended.
