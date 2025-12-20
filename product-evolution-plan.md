# Product Evolution Plan

## 1. Updated Roadmap

### Implemented Features

The following features are fully implemented and tested:

- [x] [**Service Types**](./features/service-types.md): gRPC, HTTP, OpenAPI, GraphQL, Stdio, MCP-to-MCP Proxy, WebSocket, WebRTC.
- **Upstream Authentication**: API Key, Bearer Token, OAuth 2.0.
- **Registration**: Dynamic (gRPC) and Static (YAML/JSON).
- **Advanced Service Policies**:
  - [x] [Caching](./features/caching/README.md)
  - [x] [Rate Limiting](./features/rate-limiting/README.md) (Memory & Redis)
  - [x] [Resilience](./features/resilience/README.md)
- **Deployment**: Helm Chart, Docker.
- [x] [**Health Checks**](./features/health-checks.md)
- [x] [**Schema Validation**](./features/schema-validation.md)
- [x] [**Service Profiles**](./features/profiles_and_policies/README.md)
- **Configuration**: Hot Reloading.
- [x] [**Secrets Management**](./features/security.md)
- [x] [**Distributed Tracing**](./features/tracing/README.md)
- [x] [**Transport Protocols (NATS)**](./features/nats.md)
- [x] [**Transport Protocols (Kafka)**](./features/kafka.md)
- [x] [**Automated Documentation Generation**](./features/documentation_generation.md)
- [x] [**IP Allowlisting**](./features/security.md)
- [x] [**Webhooks**](./features/webhooks/README.md)
- [x] [**Audit Logging**](./features/audit_logging.md)
- [x] [**Security Policies**](./features/profiles_and_policies/README.md)
- [x] [**Advanced Authentication**](./features/authentication/README.md)
- [x] [**Admin Management API**](./features/admin_api.md)

### High Priority (Next 1-3 Months)

- **Dynamic UI**: Build a web-based UI for managing upstream services dynamically.
- **RBAC**: Role-Based Access Control for managing user permissions.

### Ongoing Goals

- **Expand Test Coverage**: Increase unit and integration test coverage.
- **Improve Error Handling**: Enhance error messages and context.

### Long-Term Goals (6-12+ Months)

- **WASM Plugin Support**: Custom logic via WebAssembly.
- **Add Support for More Service Types**: Extend supported protocols.
- **MCP Any Config Registry**: Public registry for configurations.
- **Client SDKs**: Official Go, Python, TS libraries.

---

## 2. Top 10 Recommended Features

We have identified the following features as critical for the next phase of product evolution, focusing on Enterprise Readiness and Developer Experience.

| Rank | Feature Name | Why it matters | Implementation Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **RBAC (Role-Based Access Control)** | **Security**: Essential for multi-tenant environments. Defines roles and granular permissions beyond simple profiles. | High |
| 2 | **Dynamic Web UI** | **UX**: Completing the `ui/` scaffold to provide a visual dashboard for monitoring and configuration. | High |
| 3 | **WASM Plugin System** | **Extensibility**: Allow users to write custom transformations and policies in any language that compiles to WASM. | High |
| 4 | **Client SDKs (Python/TS)** | **DX**: Provide idiomatic libraries for consuming MCP Any from client applications. | Medium |
| 5 | **Metrics Export** | **Observability**: Export metrics to Prometheus/OpenTelemetry (beyond current Tracing). | Medium |
| 6 | **Audit Log Exporters** | **Compliance**: Push audit logs to external systems like S3, CloudWatch, or Splunk (currently file/stdout). | Low |
| 7 | **Kubernetes Operator** | **Ops**: Native K8s integration to manage MCP Any configurations as CRDs. | Medium |
| 8 | **Policy-as-Code (OPA)** | **Security**: Integrate Open Policy Agent for declarative policy enforcement. | Medium |
| 9 | **Canary/Blue-Green Routing** | **Reliability**: Advanced traffic splitting for upstream services to support safe rollouts. | Medium |
| 10 | **MCP Config Registry** | **Ecosystem**: A centralized place to share and discover configurations for popular APIs. | High |

---

## 3. Codebase Health Report

This section highlights areas of the codebase that require refactoring or attention to ensure stability and maintainability.

### 1. High Complexity in HTTP Tool Discovery
The `createAndRegisterHTTPTools` function in `pkg/upstream/http/http.go` has high cyclomatic complexity. It handles parsing, validation, and registration logic in a single large block. **Recommendation**: Refactor into smaller, testable sub-functions.

### 2. UI Implementation Pending
The `ui/` directory currently contains a starter template ("Firebase Studio"). **Recommendation**: This needs to be developed into the full "Dynamic UI" featureset, connecting to the Admin API.

### 3. Cleanup of Deprecated/Backup Files
Files like `pkg/resource/resource_test.go.bak` exist in the repository. **Recommendation**: Remove backup files to maintain repository hygiene.

### 4. Client Package Maturity
The `pkg/client` package exists but appears internal. **Recommendation**: Formalize the public API for the Go SDK and separate it from internal usage if necessary.

### 5. Regression Test Integration
Several `bug_repro_test.go` files exist (e.g., in `pkg/mcpserver`, `pkg/upstream/http`). **Recommendation**: Ensure these are integrated into the main test suite and not just ad-hoc reproduction scripts.

### 6. Dependency Management (Google APIs)
The project appears to vendor Google API definitions in `build/googleapis`. **Recommendation**: Verify if this is necessary or if managed dependencies can be used to avoid drift.
