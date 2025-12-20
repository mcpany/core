# Product Evolution Plan

## 1. Updated Roadmap

The roadmap has been reconciled with the current codebase.

### Implemented Features (Verified)

- [x] [**Service Types**](./docs/features/service-types.md) (gRPC, HTTP, OpenAPI, GraphQL, Stdio, Proxy, WebSocket, WebRTC)
- [x] [**Upstream Authentication**](./docs/features/authentication/README.md) (API Key, Bearer Token, OAuth 2.0)
- [x] [**Dynamic Registration**](./docs/features/registration.md) (via gRPC Admin API)
- [x] [**Static Registration**](./docs/features/registration.md) (via YAML/JSON)
- [x] [**Caching**](./docs/features/caching/README.md)
- [x] [**Rate Limiting**](./docs/features/rate-limiting/README.md) (Memory, Redis)
- [x] [**Resilience**](./docs/features/resilience/README.md) (Circuit Breakers, Retries)
- [x] **Deployment** (Helm, Docker)
- [x] [**Health Checks**](./docs/features/health-checks.md)
- [x] [**Schema Validation**](./docs/features/schema-validation.md)
- [x] [**Service Profiles**](./docs/features/profiles_and_policies/README.md)
- [x] **Hot Configuration Reloading**
- [x] [**Secrets Management**](./docs/features/security.md)
- [x] [**Distributed Tracing**](./docs/features/tracing/README.md)
- [x] [**NATS Support**](./docs/features/nats.md)
- [x] [**Kafka Support**](./docs/features/kafka.md)
- [x] [**Automated Documentation Generation**](./docs/features/documentation_generation.md)
- [x] [**IP Allowlisting**](./docs/features/security.md)
- [x] [**Webhooks**](./docs/features/webhooks/README.md)
- [x] [**Audit Logging**](./docs/features/audit_logging.md)
- [x] [**Security Policies**](./docs/features/profiles_and_policies/README.md)
- [x] [**Advanced Authentication**](./docs/features/authentication/README.md)

### High Priority (Next 1-3 Months)

- [ ] **Dynamic UI**: Web-based management console.
- [ ] **RBAC**: Role-Based Access Control.

### Long-Term Goals

- [ ] **WASM Plugin Support**
- [ ] **More Service Types**
- [ ] **MCP Any Config Registry**
- [ ] **Client SDKs**

---

## 2. Top 10 Recommended Features

Based on the current architecture and industry standards, the following features are recommended for immediate implementation.

### 1. Open Policy Agent (OPA) Integration
- **Why it matters (Security)**: Enables complex, fine-grained authorization policies that go beyond simple RBAC (e.g., "Allow access only if user is in group X and time is between 9-5"). Decouples policy logic from code.
- **Implementation Difficulty**: Medium

### 2. Traffic Splitting / Canary Releases
- **Why it matters (Scalability/UX)**: Allows gradual rollout of new upstream service versions (e.g., 10% of traffic to v2). Essential for safe deployments in production environments.
- **Implementation Difficulty**: High

### 3. Admin UI (Dynamic UI)
- **Why it matters (UX)**: A visual interface to view registered services, inspect health, and manage configurations is critical for operator adoption. Currently, everything is API/CLI/Config based.
- **Implementation Difficulty**: High

### 4. Mocking Service
- **Why it matters (UX/Dev)**: Allows developers to define static responses for tools. This is invaluable for testing AI agents without hitting real production APIs or incurring costs.
- **Implementation Difficulty**: Low/Medium

### 5. Data Loss Prevention (DLP)
- **Why it matters (Security)**: Automatically scan request and response payloads for sensitive data (PII, credit cards, API keys) and redact or block them before they reach the LLM or the upstream service.
- **Implementation Difficulty**: High

### 6. gRPC Transcoding
- **Why it matters (UX)**: Allow calling gRPC upstream services using HTTP/JSON from the client side (if the client doesn't support gRPC). Increases interoperability.
- **Implementation Difficulty**: High

### 7. GraphQL Federation
- **Why it matters (Scalability)**: If multiple GraphQL upstreams are registered, federation allows querying them as a single unified graph.
- **Implementation Difficulty**: High

### 8. Automated Security Scanning
- **Why it matters (Security)**: A CLI tool or runtime check that scans `config.yaml` for misconfigurations (e.g., hardcoded secrets, overly permissive policies) and warns the user.
- **Implementation Difficulty**: Medium

### 9. Cost/Usage Budgeting
- **Why it matters (Scalability)**: Track token usage or call counts per user/profile and enforce monthly budgets. Essential for providing MCP Any as a managed service or controlling API costs.
- **Implementation Difficulty**: Medium

### 10. Anomaly Detection
- **Why it matters (Security)**: Use statistical methods or AI to detect unusual traffic patterns (e.g., spike in error rates, unusual tool access sequences) that might indicate an attack or a misbehaving agent.
- **Implementation Difficulty**: High

---

## 3. Codebase Health

The following areas of the codebase require attention to ensure long-term maintainability and stability.

### 1. Middleware Logic Duplication
**Location:** `pkg/mcpserver/server.go`
**Issue:** `toolListFilteringMiddleware`, `resourceListFilteringMiddleware`, and `promptListFilteringMiddleware` contain nearly identical logic for profile-based filtering.
**Recommendation:** Refactor into a generic `GenericListFilteringMiddleware[T any]` or a shared helper function to reduce code duplication and ensure consistent behavior.

### 2. Brittle Notification Workaround
**Location:** `pkg/mcpserver/server.go`
**Issue:** The server uses a dummy resource (`internal-notification-trigger`) to force the Go SDK to emit a notification. This is a hack that relies on internal SDK behavior and may break with SDK updates.
**Recommendation:** Investigate if the Go SDK exposes a proper notification API or contribute a patch to the SDK.

### 3. Heuristic Result Handling in CallTool
**Location:** `pkg/mcpserver/server.go` (`CallTool` method)
**Issue:** The method uses heuristics (checking for `content` or `isError` keys) to determine if a map result should be treated as a `CallToolResult` or raw data. This can lead to ambiguity if a valid tool response coincidentally matches the heuristic.
**Recommendation:** Standardize the internal return types of tools. Tools should return a strictly typed `Result` object rather than `any`, or use a wrapper type to distinguish raw data from structured MCP results.

### 4. Hardcoded "Profiles" Logic
**Location:** `pkg/mcpserver/server.go`
**Issue:** Profile matching logic is embedded directly in the middleware.
**Recommendation:** Extract policy enforcement logic into a dedicated `PolicyEngine` or `AccessControlService` (potentially integrating with the OPA recommendation).

### 5. Concurrency & Error Handling
**Location:** General
**Issue:** While `xsync` is used, some error paths in `RegistrationServer` and `CallTool` return generic errors or log without sufficient context for tracing in a distributed system.
**Recommendation:** Adopt a structured error handling library and ensure all errors are wrapped with stack traces or context where appropriate.
