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
- [x] **Enhanced Secret Validation**: Strict startup validation for referenced environment variables and secret files. Server will now fail fast if a configured secret (env var or file) does not exist.
- [x] **Robust Transport Error Reporting**: Improved error messages for command-based and Docker transports. Now captures and surfaces `stderr` when a process exits unexpectedly, aiding quick debugging of configuration or runtime errors.
- [x] **Proactive Schema Sanitization**: Automatically fixes common schema issues (like missing `type: object`) in tool definitions to ensure compatibility with strict MCP clients (e.g. Claude Code).
- [x] **Smart Config Error Messages**: Detect and guide users migrating from Claude Desktop configuration format (`mcpServers` vs `upstream_services`). (Friction Fighter)
- [x] **Relative Command Resolution**: Fixed an issue where relative commands in `stdio_connection` failed validation even if they existed in the specified `working_directory`. (Friction Fighter)
- [x] **Config Linting Tool**: Implemented `mcpany lint config.yaml` to detect security issues (plain text secrets, shell injection) and best practice violations before runtime. (Friction Fighter)

## 2. Top 10 Recommended Features

These features represent the next logical steps for the product, focusing on Enterprise Readiness, Safety, and Developer Experience.

| Rank | Feature Name | Why it matters | Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **Tool Poisoning Mitigation** | **Security**: Implement integrity checks for tool definitions to prevent "Rug Pull" attacks where a tool definition changes maliciously after installation. | High |
| 3 | **Interactive OAuth Handler** | **UX/Auth**: Solve the "copy-paste token" friction. Allow users to click "Login" in the UI to authenticate tools like GitHub/Google. (Backend Implemented, UI Pending) | High |
| 4 | **Interactive OAuth UI** | **UX**: Complete the UI implementation for the Interactive OAuth Handler to allow seamless user authentication flows. | Medium |
| 5 | **Local LLM "One-Click" Connect** | **Connectivity**: Auto-detect and connect to local inference servers (Ollama, LM Studio) to democratize AI access without cloud costs. | Low |
| 6 | **Tool "Dry Run" Mode** | **DevX**: Allow tools to define a "dry run" logic to validate inputs and permissions without executing side effects. | Medium |
| 7 | **Team Configuration Sync** | **Collaboration**: Allow teams to synchronize `mcpany` configurations and secrets securely, ensuring consistent dev environments. | Medium |
| 8 | **Smart Error Recovery** | **Resilience**: Use an internal LLM loop to analyze tool errors and automatically retry with corrected parameters (Self-Healing). | High |
| 9 | **Canary Tool Deployment** | **Ops**: gradually roll out new tool versions to a subset of users or sessions to catch regressions before they impact everyone. | High |
| 10 | **Compliance Reporting** | **Enterprise**: Automated generation of PDF/CSV reports from Audit Logs for SOC2/GDPR compliance reviews. | Medium |
| 11 | **Advanced Tiered Caching** | **Performance**: Implement a multi-layer cache (Memory -> Redis -> Disk) with configurable eviction policies to reduce upstream costs. | Medium |
| 12 | **Smart Retry Policies** | **Resilience**: Configurable exponential backoff and jitter for upstream connections to handle transient network failures gracefully. | Medium |
| 13 | **Service Dependency Graph** | **Observability**: Visualize dependencies between services (e.g. Service A calls Tool B in Service C) to detect circular dependencies or bottlenecks. | High |
| 14 | **Schema Validation Playground** | **DevX**: A dedicated UI in the dashboard to paste JSON/YAML and validate it against the server schema with real-time error highlighting. | Low |
| 15 | **Partial Reloads** | **Resilience**: When reloading config dynamically, if one service is invalid, keep the old version running instead of removing it or failing the whole reload (if possible). | High |
| 16 | **Filesystem Health Check** | **Observability**: Add a health check probe for filesystem roots to report status to the UI, not just logs. | Low |
| 17 | **Safe Symlink Traversal** | **Security**: Add configuration options to strictly control symlink traversal policies (allow/deny/internal-only). | Medium |
| 18 | **Multi-Model Advisor** | **Intelligence**: Orchestrate queries across multiple models (e.g. Ollama models) to synthesize insights. | High |
| 19 | **MCP Server Aggregator/Proxy** | **Architecture**: A meta-server capability to discover, configure, and manage multiple downstream MCP servers dynamically. | High |
| 20 | **Preset Service Gallery** | **UX**: A curated list of popular services (like `wttr.in`, `sqlite`, etc.) that can be added via CLI or UI with one click/command. | Medium |
| 21 | **Configuration Migration Tool** | **DevX**: A CLI command to convert `claude_desktop_config.json` to `mcpany` config format. | Low |
| 22 | **Auth Doctor** | **UX**: Diagnostic tool to verify that configured authentication methods match upstream expectations (e.g., detecting if a Bearer token is used where Basic Auth is expected). | Medium |
| 23 | **Dynamic Tool Pruning** | **Performance/Cost**: Feature to filter visible tools based on the current user's role or context to reduce LLM context window usage and costs. | High |
| 24 | **Linter Git Hook** | **DevX**: Provide a pre-commit hook script that automatically runs `mcpany lint` on staged configuration files to prevent committing insecure configs. | Low |
| 25 | **Secret Rotation Helper** | **Ops**: A CLI tool to help rotate secrets by identifying which services are using a specific secret key/path and validating the new secret against the upstream. | Medium |

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
