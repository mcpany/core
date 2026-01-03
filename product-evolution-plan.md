# Product Evolution Plan

## 1. Updated Roadmap

The following is the reconciled roadmap as of today. Completed items are marked with `[x]` and linked to their documentation.

### Service Types (Implemented)
- [x] [gRPC](features/service-types.md)
- [x] [HTTP](features/service-types.md)
- [x] [OpenAPI](features/service-types.md)
- [x] [GraphQL](features/service-types.md)
- [x] [Stdio](features/service-types.md)
- [x] [MCP-to-MCP Proxy](features/service-types.md)
- [x] [WebSocket](features/service-types.md)
- [x] [WebRTC](features/service-types.md)
- [x] [SQL](features/sql_upstream.md)
- [x] [File System Provider](features/filesystem.md) (Local, S3, GCS)

### Authentication (Implemented)
- [x] [API Key](features/authentication/README.md)
- [x] [Bearer Token](features/authentication/README.md)
- [x] [OAuth 2.0](features/authentication/README.md)
- [x] [Role-Based Access Control (RBAC)](features/rbac.md)
- [x] [Upstream mTLS](features/security.md)

### Policies (Implemented)
- [x] [Caching](features/caching/README.md)
- [x] [Rate Limiting](features/rate-limiting/README.md) (Memory & Redis, Token-based)
- [x] [Resilience](features/resilience/README.md)

### Observability (Implemented)
- [x] [Distributed Tracing](features/tracing/README.md)
- [x] [Metrics](features/monitoring/README.md)
- [x] [Structured Logging](features/monitoring/README.md)
- [x] [Audit Logging](features/audit_logging.md)

### Security (Implemented)
- [x] [Secrets Management](features/security.md)
- [x] [IP Allowlisting](features/security.md)
- [x] [Webhooks](features/webhooks/README.md) (Pre/Post call)
- [x] [Data Loss Prevention (DLP)](features/security.md)

## 2. Top 10 Recommended Features

Based on the current architecture and market needs, the following features should be prioritized immediately.

| Rank | Feature Name | Why it matters | Implementation Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **Kubernetes Operator** | **Scalability/Ops**: Essential for enterprise adoption to manage deployment, scaling, and configuration via GitOps. | High |
| 2 | **SSO Integration (OIDC/SAML)** | **Security**: Enterprise requirement for managing access to the Admin UI and RBAC without shared credentials. | Medium |
| 3 | **Audit Log Export (Splunk/Datadog)** | **Security/Compliance**: Enterprises need to ship logs to their SIEM. Current SQLite/Postgres audit is good but needs export capabilities. | Low |
| 4 | **Interactive Playground 2.0** | **UX**: The current UI is "Beta". A robust playground with auto-generated forms for tools will significantly improve developer experience. | Medium |
| 5 | **WASM Plugin System** | **Extensibility/Security**: Allows safe extension of transformation and validation logic without recompiling the server. | High |
| 6 | **Terraform Provider** | **Ops**: "Configuration as Code" for managing MCP resources (Sources, Tools, Policies). | High |
| 7 | **Vector Database Connector** | **Feature**: Native support for vector stores (Pinecone, Milvus, Weaviate) to enable RAG workflows directly via MCP. | Medium |
| 8 | **Multi-Region Federation** | **Scalability**: Link multiple MCP instances to reduce latency and improve availability for global deployments. | High |
| 9 | **Browser Automation Provider** | **Feature**: A high-demand tool capability for agents to read/interact with websites (headless browser). | High |

## 3. Codebase Health Report

### Critical Areas
*   **Rate Limiting Complexity**: The `server/pkg/middleware/ratelimit.go` file contains complex logic mixing local and Redis implementations. Refactoring into cleaner interfaces would improve maintainability.
*   **Filesystem Provider Monolith**: `server/pkg/upstream/filesystem/upstream.go` is becoming large. S3, GCS, and Local logic should be separated into distinct providers or packages to avoid "god objects".
*   **Documentation Scatter**: Documentation is spread across `docs/features/` and `README.md`. A unified documentation site generator (using the existing `doc_generator` feature) should be standard.
*   **Test Coverage**: While unit tests exist, end-to-end integration tests for cloud providers (S3/GCS) are likely mocked or missing in CI due to credential requirements. Ensuring hermetic tests for these is crucial.
*   **Webhooks "Test" Code**: `server/cmd/webhooks` appears to be a test server but is referenced in examples. It should be clarified if this is a production-ready component or just for testing.

### Recommendations
1.  **Refactor Filesystem Upstream**: Split `upstream.go` into `s3.go`, `gcs.go`, `local.go` with a common interface.
2.  **Consolidate SDKs**: Move `server/pkg/client` to a separate repository (e.g., `mcp-go-sdk`) to encourage public usage and versioning independent of the server.
3.  **Formalize Webhook Server**: If the webhook server is intended for use, move it to `server/cmd/mcp-webhook-sidecar` and polish it.
