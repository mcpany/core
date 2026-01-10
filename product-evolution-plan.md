# Product Evolution Plan

**Date:** 2025-05-15
**Version:** 2.0.0
**Status:** Active

## 1. Updated Roadmap

This roadmap consolidates the Server and UI roadmaps into a single view of the project's trajectory.

### ✅ Completed Features

The following features have been successfully implemented and verified in the codebase.

**Core Server & Middleware**
- [x] **Agent Debugger & Inspector**: Traffic replay and inspection. [Docs](server/docs/features/debugger.md)
- [x] **Context Optimizer**: Middleware to prevent context bloat. [Docs](server/docs/features/context_optimizer.md)
- [x] **Data Loss Prevention (DLP)**: Redaction of PII in inputs/outputs. [Docs](server/docs/features/dlp.md)
- [x] **Diagnostic "Doctor" API**: `mcpctl` validation. [Docs](server/docs/features/mcpctl.md)
- [x] **SSO Integration**: OIDC/SAML support. [Docs](server/docs/features/sso.md)
- [x] **Audit Log Export**: Native Splunk/Datadog integration. [Docs](server/docs/features/audit_logging.md)
- [x] **Cost Attribution**: Token-based cost estimation. [Docs](server/docs/features/rate-limiting/README.md)
- [x] **Distributed Tracing**: OpenTelemetry support. [Docs](server/docs/features/monitoring/tracing) *(Note: Basic tracing implemented in `server/pkg/telemetry`)*

**Runtime & Plugins**
- [x] **Universal Connector Runtime**: Sidecar for stdio tools. [Docs](server/docs/features/connector_runtime.md)
- [x] **WASM Plugin System**: Sandboxed plugins. [Docs](server/docs/features/wasm.md)
- [x] **Hot Reload**: Dynamic config reloading. [Docs](server/docs/features/hot_reload.md)
- [x] **Dynamic Tool Registration**: Auto-discovery (OpenAPI/gRPC). [Docs](server/docs/features/dynamic_registration.md)
- [x] **Webhooks Sidecar**: Offloading and optimization. [Docs](server/docs/features/webhooks/sidecar.md)
- [x] **SQL Upstream**: SQL database support. [Docs](server/docs/features/sql_upstream.md)
- [x] **Message Bus**: NATS/Kafka integration. [Docs](server/docs/features/message_bus.md)

**User Interface (Dashboard)**
- [x] **Network Topology Visualization**: Interactive graph of the ecosystem. [Docs](server/docs/features/dynamic-ui.md)
- [x] **Middleware Visualization**: Drag-and-drop pipeline management. [Docs](server/docs/features/middleware_visualization.md)
- [x] **Real-time Log Streaming UI**: Live audit logs. [Docs](server/docs/features/log_streaming_ui.md)
- [x] **Granular Metrics**: Real-time stats. [Docs](server/docs/monitoring.md)
- [x] **Theme Builder**: Dark/Light mode support. [Docs](server/docs/features/theme_builder.md)
- [x] **Mobile Optimization**: Responsive layout. [Docs](server/docs/mobile-view.md)
- [x] **Service Management**: UI-based service config. [Docs](server/docs/features/dynamic-ui.md)

**DevOps & Tools**
- [x] **Helm Chart Official Support**: K8s charts. [Docs](server/docs/features/helm.md)
- [x] **Documentation Generator**: Auto-docs. [Docs](server/docs/features/documentation_generation.md)
- [x] **Structured Output Transformation**: JQ/JSONPath. [Docs](server/docs/features/transformation.md)

---

## 2. Top 10 Recommended Features (Strategic Feature Extraction)

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

---

## 3. Codebase Health Report

This section highlights areas of the codebase that require refactoring or attention to ensure long-term maintainability.

### 🔴 Critical Areas (Refactoring Needed)

1.  **Rate Limiting Complexity (`server/pkg/middleware/ratelimit.go`)**
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

### 🟡 Warning Areas

1.  **UI Component Duplication**: Some UI components in `ui/src/components` seem to have overlapping responsibilities (e.g., multiple "detail" views). A UI component audit is recommended.
2.  **Test Coverage gaps**: While core logic is tested, cloud providers (S3/GCS) and some new UI features lack comprehensive integration tests.

### 🟢 Healthy Areas

*   **Core Middleware Pipeline**: The middleware architecture is robust and extensible.
*   **Protocol Implementation**: `server/pkg/mcpserver` cleanly separates protocol details from business logic.
*   **Documentation**: The project has excellent documentation coverage for most features.
