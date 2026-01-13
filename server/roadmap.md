# Server Roadmap

## 1. Updated Roadmap

### Status: Active Development

### Implemented Features (Recently Completed)

- [x] **Upstream Service Diagnostics**: Improved error reporting in the UI for failed upstream connections (e.g., connection refused, auth failure).
- [x] **Config Environment Variable Validation**: Strict validation for missing environment variables in configuration files. [PR](#)
- [x] **Agent Debugger & Inspector**: Middleware for traffic replay and inspection. [Docs](server/docs/features/debugger.md)
- [x] **Context Optimizer**: Middleware to prevent context bloat. [Docs](server/docs/features/context_optimizer.md)
- [x] **Data Loss Prevention (DLP)**: Redaction of PII in inputs/outputs. [Docs](server/docs/features/dlp.md)
- [x] **Diagnostic "Doctor" API**: `mcpctl` validation and health checks. [Docs](server/docs/features/mcpctl.md)
  - _Update_: CLI implementation completed. Now supports connection verification and detailed health reporting.
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
- [x] **Secrets Management Integration**: First-class integration with HashiCorp Vault or AWS Secrets Manager. [Docs](server/docs/reference/configuration.md#upstreamauthentication-outgoing)
- [x] **Resilient Configuration Loading**: Better error reporting for invalid configs and keeping services in the list with error state.
- [x] **Pre-flight Config Validation**: Added pre-flight checks for command existence and working directory validity to prevent "silent failures" at runtime.
- [x] **Strict Startup Validation**: Server now fails fast with descriptive errors if the initial configuration contains invalid services, preventing "zombie services" (Friction Fighter).
- [x] **Filesystem Path Validation**: Proactive validation of configured `root_paths` for Filesystem services to warn about missing directories on startup.
- [x] **Detailed Subprocess Error Reporting**: Captured stderr from failed MCP subprocess launches to expose actionable errors (e.g., command not found) instead of generic connection failures. (Friction Fighter)

## 2. Top 10 Recommended Features

These features represent the next logical steps for the product, focusing on Enterprise Readiness, Safety, and Developer Experience.

| Rank | Feature Name | Why it matters | Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **Human-in-the-Loop Approval UI** | **Safety**: Critical for preventing dangerous actions (e.g., `DROP TABLE`) by requiring manual admin approval via the UI before execution. | Low |
| 2 | **Tool Poisoning Mitigation** | **Security**: Implement integrity checks for tool definitions to prevent "Rug Pull" attacks where a tool definition changes maliciously after installation. | High |
| 3 | **Interactive OAuth Handler** | **UX/Auth**: Solve the "copy-paste token" friction. Allow users to click "Login" in the UI to authenticate tools like GitHub/Google. | High |
| 4 | **Local LLM "One-Click" Connect** | **Connectivity**: Auto-detect and connect to local inference servers (Ollama, LM Studio) to democratize AI access without cloud costs. | Low |
| 5 | **Tool "Dry Run" Mode** | **DevX**: Allow tools to define a "dry run" logic to validate inputs and permissions without executing side effects. | Medium |
| 6 | **Team Configuration Sync** | **Collaboration**: Allow teams to synchronize `mcpany` configurations and secrets securely, ensuring consistent dev environments. | Medium |
| 7 | **Smart Error Recovery** | **Resilience**: Use an internal LLM loop to analyze tool errors and automatically retry with corrected parameters (Self-Healing). | High |
| 8 | **Canary Tool Deployment** | **Ops**: gradually roll out new tool versions to a subset of users or sessions to catch regressions before they impact everyone. | High |
| 9 | **Compliance Reporting** | **Enterprise**: Automated generation of PDF/CSV reports from Audit Logs for SOC2/GDPR compliance reviews. | Medium |
| 10 | **Advanced Tiered Caching** | **Performance**: Implement a multi-layer cache (Memory -> Redis -> Disk) with configurable eviction policies to reduce upstream costs. | Medium |
| 11 | **Smart Retry Policies** | **Resilience**: Configurable exponential backoff and jitter for upstream connections to handle transient network failures gracefully. | Medium |
| 12 | **Service Dependency Graph** | **Observability**: Visualize dependencies between services (e.g. Service A calls Tool B in Service C) to detect circular dependencies or bottlenecks. | High |
| 13 | **Schema Validation Playground** | **DevX**: A dedicated UI in the dashboard to paste JSON/YAML and validate it against the server schema with real-time error highlighting. | Low |
| 14 | **Config Linting Tool** | **DevX**: A standalone CLI command (e.g., `mcpany lint config.yaml`) to validate configuration without starting the server, useful for CI/CD pipelines. | Low |
| 15 | **Partial Reloads** | **Resilience**: When reloading config dynamically, if one service is invalid, keep the old version running instead of removing it or failing the whole reload (if possible). | High |
| 16 | **Filesystem Health Check** | **Observability**: Add a health check probe for filesystem roots to report status to the UI, not just logs. | Low |
| 17 | **Safe Symlink Traversal** | **Security**: Add configuration options to strictly control symlink traversal policies (allow/deny/internal-only). | Medium |
| 18 | **Windows Path Handling** | **UX**: Add explicit path normalization for Windows environments to handle backslash/forward slash discrepancies in configuration automatically. | Low |
| 19 | **Configuration Migration Tool** | **DevX**: CLI command to automatically upgrade old configuration files to the latest schema version. | Medium |

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
