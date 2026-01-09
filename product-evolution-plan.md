# Product Evolution Plan

## 1. Updated Roadmap

### Status: Active Development

### Implemented Features (Recently Completed)

- [x] **Agent Debugger & Inspector**: Middleware for traffic replay and inspection. [Docs](server/docs/features/debugger.md)
- [x] **Context Optimizer**: Middleware to prevent context bloat. [Docs](server/docs/features/context_optimizer.md)
- [x] **Diagnostic "Doctor" API**: `mcpctl` validation and health checks. [Docs](server/docs/features/mcpctl.md)
- [x] **SSO Integration**: OIDC/SAML support. [Docs](server/docs/features/sso.md)
- [x] **Audit Log Export**: Native Splunk and Datadog integration. [Docs](server/docs/features/audit_logging.md)
- [x] **Cost Attribution**: Token-based cost estimation and metrics. [Docs](server/docs/features/rate-limiting/README.md)
- [x] **Universal Connector Runtime**: Sidecar for stdio tools. [Docs](server/docs/features/connector_runtime.md)
- [x] **WASM Plugin System**: Runtime for sandboxed plugins. [Docs](server/docs/features/wasm.md)
- [x] **Hot Reload**: Dynamic configuration reloading. [Docs](server/docs/features/hot_reload.md)
- [x] **SQL Upstream**: Expose SQL databases as tools. [Docs](server/docs/features/sql_upstream.md)
- [x] **Webhooks Sidecar**: Context optimization and offloading. [Docs](server/docs/features/webhooks/sidecar.md)
- [x] **Dynamic Tool Registration**: Auto-discovery from OpenAPI/gRPC/GraphQL. [Docs](server/docs/features/dynamic_registration.md)
- [x] **Helm Chart Official Support**: K8s deployment charts. [Docs](server/docs/features/helm.md)
- [x] **Message Bus**: NATS/Kafka integration for events. [Docs](server/docs/features/message_bus.md)
- [x] **Structured Output Transformation**: JQ/JSONPath response shaping. [Docs](server/docs/features/transformation.md)

## 2. Top 10 Recommended Features

| Rank | Feature Name | Why it matters | Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **Kubernetes Operator V2** | **Scalability/Ops**: The current `k8s/operator` is skeletal. To support enterprise scale, we need a robust operator with CRDs for defining MCP Servers, handling automated updates, and managing secrets. | High |
| 2 | **Browser Automation Provider** | **Feature**: A "Read Webpage" capability is essential for modern agents. Integrating Playwright as an upstream provider will allow safe, headless interaction with dynamic web content. | High |
| 3 | **Multi-Region Federation** | **Scalability**: For global deployments, linking multiple MCP Any instances (Core-to-Core) reduces latency by routing tool calls to the nearest available provider. | High |
| 4 | **Active-Active High Availability** | **Reliability**: Production environments demand zero downtime. Implementing a leaderless clustering model ensures the system survives node failures and supports rolling upgrades. | High |
| 5 | **Disaster Recovery Playbook** | **Ops**: We need automated tooling to snapshot the server state and configuration to cloud storage (S3/GCS) to meet enterprise SLA requirements for recovery time. | Medium |
| 6 | **Dynamic Secret Rotation** | **Security**: Hardcoded secrets are a risk. Integrating with HashiCorp Vault or AWS Secrets Manager will allow credentials to rotate automatically without server restarts. | High |
| 7 | **Downstream mTLS** | **Security**: To achieve a Zero Trust architecture, we must enforce mutual TLS authentication for all agents connecting to the MCP Any server. | Medium |
| 8 | **Just-In-Time (JIT) Access** | **Security**: Implement temporary privilege elevation (e.g., "Grant Write access for 1 hour") to enforce the principle of least privilege for sensitive tools. | High |
| 9 | **Persistent Vector Store** | **Core**: While SQLite vector search is good for testing, production use requires connectors for dedicated vector databases like Milvus, Pinecone, or pgvector. | Medium |
| 10 | **SDK Consolidation** | **DevX**: The `server/pkg/client` package is coupled to the server. Extracting it into a standalone Go SDK repository will simplify integration for third-party developers. | Medium |

## 3. Codebase Health

### Critical Areas

- **Rate Limiting Complexity**: The current implementation in `server/pkg/middleware/ratelimit.go` tightly couples in-memory logic with Redis commands. This structure makes unit testing difficult and prevents easy addition of new backends.
- **Filesystem Provider Monolith**: `server/pkg/upstream/filesystem/upstream.go` currently implements logic for Local, S3, and GCS backends in a single file. This violation of the Single Responsibility Principle makes the code brittle and hard to maintain.
- **Test Coverage for Cloud Providers**: There is a significant gap in testing for S3 and GCS integrations. We rely on mocks or manual testing; introducing integration tests using local emulators (like MinIO or fake-gcs-server) is critical.
- **Webhooks "Test" Code**: The `server/cmd/webhooks` directory contains code that looks like a prototype. If this is intended for production use as a sidecar, it needs to be formalized with proper configuration, logging, and error handling.
- **SDK Consolidation**: As mentioned in the recommendations, `server/pkg/client` should be moved to its own repository to allow clients to import the SDK without pulling in the entire server dependency tree.

### Recommendations

1.  **Refactor Rate Limiting**: Introduce a `RateLimiterStrategy` interface and implement distinct `LocalStrategy` and `RedisStrategy` structs.
2.  **Refactor Filesystem Upstream**: Adopt a Factory pattern to instantiate specific implementations for Local, S3, and GCS, separating their logic into distinct files.
3.  **Formalize Webhook Server**: graduate `server/cmd/webhooks` from a prototype to a fully supported sidecar component.
4.  **Standardize Configuration**: Audit configuration structures across modules to ensure consistent naming conventions and validation logic.
