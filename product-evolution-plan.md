# Product Evolution Plan

## 1. Updated Roadmap

### Status: Active Development

### Implemented Features

#### Service Types
- [x] [gRPC](docs/features/service_types.md)
- [x] [HTTP](docs/features/service_types.md)
- [x] [OpenAPI](docs/features/service_types.md)
- [x] [GraphQL](docs/features/service_types.md)
- [x] [Stdio](docs/features/service_types.md)
- [x] [MCP-to-MCP Proxy](docs/features/service_types.md)
- [x] [WebSocket](docs/features/service_types.md)
- [x] [WebRTC](docs/features/service_types.md)
- [x] [SQL](docs/features/service_types.md)

#### Authentication & Authorization
- [x] [API Key](docs/features/auth.md)
- [x] [Bearer Token](docs/features/auth.md)
- [x] [OAuth 2.0](docs/features/auth.md)
- [x] [Role-Based Access Control (RBAC)](docs/features/auth.md)

#### Policies
- [x] [Caching](docs/features/policies.md)
- [x] [Rate Limiting](docs/features/policies.md)
- [x] [Resilience](docs/features/policies.md)

#### Observability
- [x] [Distributed Tracing](docs/features/observability.md)
- [x] [Metrics](docs/features/observability.md)
- [x] [Structured Logging](docs/features/observability.md)
- [x] [Audit Logging](docs/features/observability.md)

#### Security
- [x] [Secrets Management](docs/features/security.md)
- [x] [IP Allowlisting](docs/features/security.md)
- [x] [Webhooks](docs/features/security.md)

#### Core
- [x] Dynamic Tool Registration
- [x] Message Bus (NATS, Kafka)
- [x] [Structured Output Transformation](docs/features/transformation.md)
- [x] [Admin Management API](docs/features/admin_api.md)
- [x] [Dynamic Web UI](docs/features/web_ui.md) (Beta)

---

## 2. Top 10 Recommended Features

| Rank | Feature Name | Why it matters | Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **WASM Plugin System** | **Extensibility**: Allows users to write custom transformations/validations safely without recompiling the core server. | High |
| 2 | **Cost & Quota Management** | **Scalability/Business**: Essential for SaaS/multi-tenant deployments to track and limit usage. | Medium |
| 3 | **Client SDKs (Python/TS)** | **UX/Adoption**: Simplifies integration for developers building AI agents and apps. | Medium |
| 4 | **File System Provider** | **Utility**: Provides safe, controlled access to local files, highly requested for local agent workflows. | Low |
| 5 | **Interactive Tool Playground** | **UX**: Allows users to test and debug tools directly within the Web UI before exposing them to agents. | Medium |
| 6 | **External Secrets Integration** | **Security**: Integration with enterprise vaults (HashiCorp Vault, AWS Secrets Manager) instead of just env vars. | Medium |
| 7 | **Audit Log Streaming** | **Compliance**: Stream audit logs to external systems (Splunk, CloudWatch, S3) rather than just local storage. | Medium |
| 8 | **Traffic Splitting / Canary** | **Reliability**: Enable A/B testing of tool backends by splitting traffic between versions. | High |
| 9 | **Service Mesh Integration** | **Scalability**: Native mTLS and observability integration with Istio/Linkerd. | High |
| 10 | **Policy-as-Code (OPA)** | **Security**: Fine-grained, logic-based authorization policies using Open Policy Agent. | High |

---

## 3. Codebase Health

### Strengths
- **Structure**: The project follows standard Go project layout (`pkg/`, `cmd/`, `internal/`).
- **Testing**: Good foundation with `make test` covering unit, integration, and E2E scenarios.
- **Linting**: Comprehensive linting setup with `golangci-lint` and `pre-commit` hooks.
- **Documentation**: `docs/` folder is well-organized, and feature coverage is decent.

### Areas for Improvement
- **Test Coverage**: While tests exist, ensure coverage metrics are monitored and improved, especially for new features like RBAC.
- **Code Comments**: Some complex logic could benefit from more inline documentation (e.g., in `middleware/`).
- **TODOs**: A scan revealed a few `TODO` items (e.g., in `server/pkg/prompt/service.go`) that should be addressed or tracked in issues.
- **Dependency Management**: Ensure `go.mod` and `package.json` dependencies are kept up-to-date to avoid security vulnerabilities.
