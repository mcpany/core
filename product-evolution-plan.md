# Product Evolution Plan

## 1. Updated Roadmap

The following is the reconciled roadmap as of today. Completed items are marked with `[x]` and linked to their documentation.

### Service Types (Implemented)
- [x] [gRPC](server/docs/features/service-types.md)
- [x] [HTTP](server/docs/features/service-types.md)
- [x] [OpenAPI](server/docs/features/service-types.md)
- [x] [GraphQL](server/docs/features/service-types.md)
- [x] [Stdio](server/docs/features/service-types.md)
- [x] [MCP-to-MCP Proxy](server/docs/features/service-types.md)
- [x] [WebSocket](server/docs/features/service-types.md)
- [x] [WebRTC](server/docs/features/service-types.md)
- [x] [SQL](server/docs/features/sql_upstream.md)
- [x] [File System Provider](server/docs/features/filesystem.md) (Local, S3, GCS)
- [x] [Vector Database Connector](server/docs/features/vector_database.md) (Pinecone)

### Authentication (Implemented)
- [x] [API Key](server/docs/features/authentication/README.md)
- [x] [Bearer Token](server/docs/features/authentication/README.md)
- [x] [OAuth 2.0](server/docs/features/authentication/README.md)
- [x] [Role-Based Access Control (RBAC)](server/docs/features/rbac.md)
- [x] [Upstream mTLS](server/docs/features/security.md)

### Policies (Implemented)
- [x] [Caching](server/docs/features/caching/README.md)
- [x] [Rate Limiting](server/docs/features/rate-limiting/README.md) (Memory & Redis, Token-based)
- [x] [Resilience](server/docs/features/resilience/README.md)

### Observability (Implemented)
- [x] [Distributed Tracing](server/docs/features/tracing/README.md)
- [x] [Metrics](server/docs/features/monitoring/README.md)
- [x] [Structured Logging](server/docs/features/monitoring/README.md)
- [x] [Audit Logging](server/docs/features/audit_logging.md)
- [x] [Audit Log Export](server/docs/features/audit_logging.md) (Splunk/Datadog)

### Security (Implemented)
- [x] [Secrets Management](server/docs/features/security.md)
- [x] [IP Allowlisting](server/docs/features/security.md)
- [x] [Webhooks](server/docs/features/webhooks/README.md) (Pre/Post call)
- [x] [Data Loss Prevention (DLP)](server/docs/features/security.md)

## 2. Top 10 Recommended Features

Based on the current architecture and market needs, the following features should be prioritized immediately.

| Rank | Feature Name | Why it matters | Implementation Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **Kubernetes Operator** | **Scalability/Ops**: Essential for enterprise adoption to manage deployment, scaling, and configuration via GitOps. | High |
| 2 | **SSO Integration (OIDC/SAML)** | **Security**: Enterprise requirement for managing access to the Admin UI and RBAC without shared credentials. | Medium |
| 3 | **Interactive Playground 2.0** | **UX**: The current UI is "Beta". A robust playground with auto-generated forms for tools will significantly improve developer experience. | Medium |
| 4 | **WASM Plugin System** | **Extensibility/Security**: Allows safe extension of transformation and validation logic without recompiling the server. | High |
| 5 | **Terraform Provider** | **Ops**: "Configuration as Code" for managing MCP resources (Sources, Tools, Policies). | High |
| 6 | **Additional Vector Connectors** | **Feature**: Support for Milvus and Weaviate to enable RAG workflows (Pinecone is already implemented). | Medium |
| 7 | **Multi-Region Federation** | **Scalability**: Link multiple MCP instances to reduce latency and improve availability for global deployments. | High |
| 8 | **Browser Automation Provider** | **Feature**: A high-demand tool capability for agents to read/interact with websites (headless browser). | High |
| 9 | **Jira/Confluence Connector** | **Feature**: Critical integration for enterprise knowledge management and workflow automation. | Medium |
| 10 | **Cost Attribution & Billing** | **Business**: Track token usage and API costs per user/team to enable chargeback models. | Medium |

## 3. Codebase Health Report

### Critical Areas
*   **Rate Limiting Complexity**: The `server/pkg/middleware/ratelimit.go` file contains complex logic mixing local and Redis implementations, along with cost estimation. Refactoring into cleaner, separate strategies (e.g., `TokenBucketLimiter`, `RedisLimiter`, `MemoryLimiter`) would improve testability and maintainability.
*   **Filesystem Provider Monolith**: `server/pkg/upstream/filesystem/upstream.go` handles multiple filesystem types (Local, S3, GCS, Zip, SFTP) in a single `createProvider` function. As new providers are added, this will become unmanageable. Splitting these into distinct packages/factories with a common interface is recommended.
*   **Documentation Scatter**: Documentation is spread across `server/docs/features/` and `README.md`. There is no single source of truth for "How to configure X". A unified documentation structure or site generator (enhancing the existing `doc_generator`) should be prioritized.
*   **Test Coverage for Cloud Providers**: End-to-end integration tests for S3 and GCS are likely missing or mocked in CI due to credential requirements. Implementing a local emulation layer (e.g., MinIO for S3) for CI tests would ensure robustness.
*   **Webhooks "Test" Code**: The `server/cmd/webhooks` directory appears to be a test server but is referenced in examples. It should be clarified if this is a production-ready component or just for testing. If for production, it needs proper structure and tests.

### Recommendations
1.  **Refactor Filesystem Upstream**: Split `upstream.go` into `s3.go`, `gcs.go`, `local.go`, etc., using a factory pattern.
2.  **Consolidate SDKs**: Move `server/pkg/client` to a separate repository (e.g., `mcp-go-sdk`) to encourage public usage and versioning independent of the server.
3.  **Formalize Webhook Server**: If the webhook server is intended for use, move it to `server/cmd/mcp-webhook-sidecar` and polish it.
4.  **Standardize Configuration**: Ensure all features (like Rate Limiting) have consistent configuration patterns in `config.yaml` and corresponding documentation.
