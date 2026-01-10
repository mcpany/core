# Server Roadmap

## 1. Updated Roadmap

### Status: Active Development

### Implemented Features (Recently Completed)

- [x] **Agent Debugger & Inspector**: Middleware for traffic replay and inspection. [Docs](server/docs/features/debugger.md)
- [x] **Context Optimizer**: Middleware to prevent context bloat. [Docs](server/docs/features/context_optimizer.md)
- [x] **Data Loss Prevention (DLP)**: Redaction of PII in inputs/outputs. [Docs](server/docs/features/dlp.md)
- [x] **Diagnostic "Doctor" API**: `mcpctl` validation and health checks. [Docs](server/docs/features/mcpctl.md)
- [x] **SSO Integration**: OIDC/SAML support. [Docs](server/docs/features/sso.md)
- [x] **Audit Log Export**: Native Splunk and Datadog integration. [Docs](server/docs/features/audit_logging.md)
- [x] **Cost Attribution**: Token-based cost estimation and metrics. [Docs](server/docs/features/rate-limiting/README.md)
- [x] **Distributed Tracing**: OpenTelemetry support. [Docs](server/docs/features/monitoring/tracing)
- [x] **Universal Connector Runtime**: Sidecar for stdio tools. [Docs](server/docs/features/connector_runtime.md)
- [x] **WASM Plugin System**: Runtime for sandboxed plugins. [Docs](server/docs/features/wasm.md)
- [x] **Hot Reload**: Dynamic configuration reloading. [Docs](server/docs/features/hot_reload.md)
- [x] **SQL Upstream**: Expose SQL databases as tools. [Docs](server/docs/features/sql_upstream.md)
- [x] **Webhooks Sidecar**: Context optimization and offloading. [Docs](server/docs/features/webhooks/sidecar.md)
- [x] **Dynamic Tool Registration**: Auto-discovery from OpenAPI/gRPC/GraphQL. [Docs](server/docs/features/dynamic_registration.md)
- [x] **Helm Chart Official Support**: K8s deployment charts. [Docs](server/docs/features/helm.md)
- [x] **Message Bus**: NATS/Kafka integration for events. [Docs](server/docs/features/message_bus.md)
- [x] **Structured Output Transformation**: JQ/JSONPath response shaping. [Docs](server/docs/features/transformation.md)
- [x] **Documentation Generator**: Auto-generate beautiful Markdown/HTML documentation. [Docs](server/docs/features/documentation_generation.md)
- [x] **Pre-flight Config Validation**: Added pre-flight checks for command existence and working directory validity to prevent "silent failures" at runtime.

## 2. Top 10 Recommended Features

These features represent the next logical steps for the product, focusing on Enterprise Readiness, Safety, and Developer Experience.

| Rank | Feature Name | Why it matters | Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **Human-in-the-Loop Approval UI** | **Safety**: Critical for preventing dangerous actions (e.g., `DROP TABLE`) by requiring manual admin approval via the UI before execution. | Low |
| 2 | **Secrets Management Integration** | **Security**: First-class integration with HashiCorp Vault or AWS Secrets Manager to avoid storing API keys in plain text configs or env vars. | Medium |
| 3 | **Interactive OAuth Handler** | **UX/Auth**: Solve the "copy-paste token" friction. Allow users to click "Login" in the UI to authenticate tools like GitHub/Google. | High |
| 4 | **Local LLM "One-Click" Connect** | **Connectivity**: Auto-detect and connect to local inference servers (Ollama, LM Studio) to democratize AI access without cloud costs. | Low |
| 5 | **Tool "Dry Run" Mode** | **DevX**: Allow tools to define a "dry run" logic to validate inputs and permissions without executing side effects. | Medium |
| 6 | **Team Configuration Sync** | **Collaboration**: Allow teams to synchronize `mcpany` configurations and secrets securely, ensuring consistent dev environments. | Medium |
| 7 | **Smart Error Recovery** | **Resilience**: Use an internal LLM loop to analyze tool errors and automatically retry with corrected parameters (Self-Healing). | High |
| 8 | **Canary Tool Deployment** | **Ops**: gradually roll out new tool versions to a subset of users or sessions to catch regressions before they impact everyone. | High |
| 9 | **Compliance Reporting** | **Enterprise**: Automated generation of PDF/CSV reports from Audit Logs for SOC2/GDPR compliance reviews. | Medium |
| 10 | **Advanced Tiered Caching** | **Performance**: Implement a multi-layer cache (Memory -> Redis -> Disk) with configurable eviction policies to reduce upstream costs. | Medium |

## 3. Codebase Health

### Critical Areas (Refactoring Needed)

1.  **Rate Limiting Complexity (`server/pkg/middleware/ratelimit.go`)**
    *   **Status**: ✅ Refactored (Strategy Pattern Implemented).
    *   **Issue**: The current implementation tightly couples business logic with Redis/Memory backend logic.
    *   **Risk**: High cognitive load, difficult to test, hard to add new backends (e.g., Postgres).
    *   **Recommendation**: Extract a `RateLimitStrategy` interface.

2.  **Filesystem Upstream Monolith (`server/pkg/upstream/filesystem`)**
    *   **Issue**: The logic for Local, S3, and GCS providers is often intertwined or shares too much structural code in a way that violates SRP.
    *   **Risk**: Changes to S3 support might break Local file support.
    *   **Recommendation**: Strictly separate providers into distinct packages/structs implementing a common interface.

3.  **Webhook "Prototype" Status (`server/cmd/webhooks`)**
    *   **Issue**: The code appears experimental and lacks the robustness of the main server (error handling, config validation).
    *   **Risk**: If used in production, it could become a single point of failure.
    *   **Recommendation**: Graduate this to `server/pkg/sidecar/webhooks` with full test coverage.

### Warning Areas

1.  **UI Component Duplication**: Some UI components in `ui/src/components` seem to have overlapping responsibilities (e.g., multiple "detail" views). A UI component audit is recommended.
2.  **Test Coverage gaps**: While core logic is tested, cloud providers (S3/GCS) and some new UI features lack comprehensive integration tests.

### Healthy Areas

*   **Core Middleware Pipeline**: The middleware architecture is robust and extensible.
*   **Protocol Implementation**: `server/pkg/mcpserver` cleanly separates protocol details from business logic.
*   **Documentation**: The project has excellent documentation coverage for most features.
