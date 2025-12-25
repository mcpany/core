# Product Evolution Plan

## 1. Updated Roadmap

The following roadmap reflects the current state of the codebase and the strategic direction for MCP Any.

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
- [x] [File System Provider](docs/features/filesystem.md)

#### Authentication
- [x] [API Key](docs/features/auth.md)
- [x] [Bearer Token](docs/features/auth.md)
- [x] [OAuth 2.0](docs/features/auth.md)
- [x] [Role-Based Access Control (RBAC)](docs/features/auth.md)

#### Policies
- [x] [Caching](docs/features/policies.md)
- [x] [Rate Limiting](docs/features/policies.md) (Memory & Redis)
- [x] [Resilience](docs/features/policies.md) (Circuit Breakers & Retries)

#### Observability
- [x] [Distributed Tracing](docs/features/observability.md) (OpenTelemetry)
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
- [x] [Structured Output Transformation](docs/features/transformation.md) (JQ/JSONPath)
- [x] [Admin Management API](docs/features/admin_api.md)
- [x] [Dynamic Web UI](docs/features/web_ui.md) (Beta)

### Upcoming Features (High Priority)

1.  **Client SDKs (Python/TS)**
2.  **WASM Plugins**
3.  **Cloud Storage Support (S3, GCS)**
4.  **Cost & Quota Management**

## 2. Top 10 Recommended Features

These recommendations are based on a gap analysis of the current product capabilities vs. industry standards for enterprise-grade API gateways and LLM infrastructure.

| Rank | Feature Name | Why it matters | Implementation Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **Client SDKs (Python/TypeScript)** | Accelerates adoption by providing developers with idiomatic, typed wrappers for interacting with MCP Any. Critical for DX. | Medium |
| 2 | **WASM Plugins** | Enables users to extend functionality (transformations, validators) safely without recompiling the server. Key for extensibility. | High |
| 3 | **Cloud Storage Support (S3, GCS)** | Extends the Filesystem provider to support cloud storage, allowing LLMs to interact with enterprise data lakes. | Medium |
| 4 | **Cost & Quota Management** | Essential for SaaS / multi-tenant deployments to control costs and monetize usage (tokens/requests per tenant). | Medium |
| 5 | **Vector Database Integration** | Enables RAG (Retrieval-Augmented Generation) workflows directly within the gateway, turning it into a "Semantic Gateway". | Medium |
| 6 | **Advanced Identity Federation** | Support for OIDC and SAML 2.0 to integrate with enterprise Identity Providers (Okta, Entra ID) beyond simple OAuth2. | Medium |
| 7 | **Human-in-the-Loop Workflows** | Adds a layer of safety by requiring human approval for sensitive tool executions (e.g., database writes, emails). | Medium |
| 8 | **Code Sandbox / Remote Code Execution** | Allows LLMs to execute code snippets (Python/JS) safely. High value for data analysis agents but requires strict isolation. | High |
| 9 | **Terraform / IaC Provider** | Allows infrastructure teams to manage MCP Any configuration (services, policies) using GitOps and Infrastructure as Code. | Medium |
| 10 | **Prompt Management Registry** | A centralized store for managing and versioning system prompts and agent instructions, decoupled from code. | Low |

## 3. Codebase Health Report

### Overview
The codebase appears robust, following standard Go project layouts (`server/pkg/...`). The separation of concerns between `upstream`, `transformer`, and `middleware` is clean.

### Key Observations
*   **Dependency Management**: The project uses `go.work`, indicating a multi-module workspace. This is good for separation but requires careful management of dependency versions.
*   **Filesystem Provider**: The recent implementation of `server/pkg/upstream/filesystem` correctly uses `spf13/afero` for filesystem abstraction. This is a best practice and significantly eases the path to adding Cloud Storage (S3/GCS) support later.
*   **Testing**: There is a significant number of tests (`grep -r "test" server/pkg | wc -l` returned ~4500 lines). `upstream_test.go` exists in most packages, which is a good sign.
*   **TODOs**: There are a moderate number of `TODO` comments in the codebase. These should be reviewed, particularly those in critical paths like `upstream` and `auth`.
*   **Documentation**: The `docs/` folder is well-structured. Keeping `ROADMAP.md` and `features/` in sync with code is crucial, as addressed in this update.
*   **Python Virtual Environment**: There are traces of a Python virtual environment in `build/venv`, likely for testing or pre-commit hooks. Ensure these do not leak into production builds.

### Recommendations for Refactoring
*   **Consolidate Upstream Interfaces**: Ensure all upstreams strictly adhere to the `Upstream` interface. As more types (like SQL, Filesystem) are added, verify that the interface abstraction holds up without excessive type casting.
*   **Standardize Error Handling**: Review error returns across upstreams to ensure consistent error codes/types are returned to the client (e.g., distinguishing between "upstream unavailable" vs "invalid arguments").
*   **Audit "Planned" Stubs**: The code contains several stubs or comments for planned features (e.g., in filesystem). These should either be implemented or clearly marked as "Not Implemented" in the API responses to avoid user confusion.
