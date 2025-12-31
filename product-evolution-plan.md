# Product Evolution Plan

## Phase 1: Roadmap Reconciliation

The `docs/roadmap.md` file has been synchronized with the current state of the codebase. We have identified several key features that were implemented but not marked as such, including the CLI validator, DLP, and Hot Reload.

### Completed Items (Moved to "Implemented")
*   **CI/CD Config Validator CLI**: `server/cmd/mcpctl`
*   **Data Loss Prevention (DLP)**: `server/pkg/middleware/dlp.go`
*   **Hot Reload**: `server/pkg/config/watcher.go`
*   **Doc Generator**: `server/pkg/config/doc_generator.go`
*   **SQL Upstream**: `server/pkg/upstream/sql`
*   **Helm Chart**: `server/helm`

## Phase 2: Strategic Feature Extraction

Based on the current architecture and industry standards for AI Gateway/MCP tools, here are the Top 10 Recommended Features for the next development cycle.

### Top 10 New Features

| Priority | Feature Name | Why it matters | Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **Kubernetes Operator** | **Scalability/Ops**: Simplifies Day 2 operations (upgrades, backups, scaling) and enables GitOps. Critical for enterprise adoption. | High |
| 2 | **Client SDKs (Python/TS)** | **UX**: Reducing friction for developers integrating agents is vital. Current raw API/gRPC usage is cumbersome. | Medium |
| 3 | **Cloud Storage Provider (S3/GCS)** | **UX/Capability**: Agents frequently need to read/write files. Extending the filesystem provider to cloud buckets unlocks massive utility. | Medium |
| 4 | **SSO with OIDC/SAML** | **Security**: Enterprise customers require integration with Okta/AzureAD for the Admin UI and RBAC. | Medium |
| 5 | **Token-Based Quota Management** | **Security/Business**: Rate limiting is not enough. Managing "spend" (LLM tokens or API calls) is required for cost control and billing. | Medium |
| 6 | **Browser Automation Provider** | **Capability**: Giving agents the ability to "see" and "interact" with the web is a top user request (Headless Chrome/Playwright). | High |
| 7 | **Interactive Playground 2.0** | **UX**: The current UI is basic. A rich debugger with form generation for tools and history replay will speed up prompt engineering. | Medium |
| 8 | **WASM Plugin System** | **Extensibility**: Allowing users to write custom logic in Rust/Go/TS and run it safely (sandboxed) prevents "middleware sprawl" in the core. | High |
| 9 | **Vector Database Connectors** | **Capability**: While we have semantic caching, direct tools to query Pinecone/Weaviate/Milvus are needed for RAG agents. | Medium |
| 10 | **Audit Log Exporters** | **Observability**: Built-in logging is good, but enterprises need native push support to Splunk, Datadog, or CloudWatch Logs. | Low |

## Phase 3: Codebase Health

### Areas for Improvement

1.  **Upstream Factory Complexity**:
    *   **Location**: `server/pkg/upstream/factory/factory.go`
    *   **Issue**: The factory is a growing switch statement. As we add more providers (S3, Browser, Vectors), this will become a bottleneck and violation of Open/Closed Principle.
    *   **Recommendation**: Implement a dynamic registry pattern for upstream providers similar to how tools are registered.

2.  **Configuration Monolith**:
    *   **Location**: `server/pkg/config`
    *   **Issue**: The configuration loading logic handles files, env vars, GitHub, and more. It is becoming complex to test and maintain.
    *   **Recommendation**: Refactor into distinct `Source` providers and a cleaner merging strategy.

3.  **Documentation Structure**:
    *   **Location**: `docs/features/`
    *   **Issue**: The folder is becoming flat and cluttered.
    *   **Recommendation**: Group by category (e.g., `docs/features/security/`, `docs/features/connectivity/`).

4.  **Testing Strategy**:
    *   **Observation**: We have good "bug repro" tests in `mcpserver`, which is excellent. However, integration tests for complex upstreams (like SQL) rely on specific drivers.
    *   **Recommendation**: Introduce Docker-based integration tests (Testcontainers) to verify SQL, Redis, and other external dependencies reliably in CI.

## Updated Roadmap

Please refer to `docs/roadmap.md` for the live, reconciled roadmap.
