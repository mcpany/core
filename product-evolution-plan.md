# Product Evolution Plan

## 1. Updated Roadmap

The following roadmap reflects the reconciled state of the project, with implemented features marked and linked to their documentation.

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

#### Authentication & Security
- [x] [API Key](docs/features/auth.md)
- [x] [Bearer Token](docs/features/auth.md)
- [x] [OAuth 2.0](docs/features/auth.md)
- [x] [Secrets Management](docs/features/security.md)
- [x] [IP Allowlisting](docs/features/security.md)
- [x] [Webhooks](docs/features/security.md)
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

#### Core & UI
- [x] Dynamic Tool Registration
- [x] Message Bus
- [x] [Structured Output Transformation](docs/features/transformation.md)
- [x] [Dynamic Web UI](docs/features/web_ui.md) (Beta)
- [x] Admin Management API (Partial)

### Planned Features (Backlog)
1.  **WASM Plugins**: For sandboxed custom logic.
2.  **File System Provider**: Safe local file access.
3.  **Cost & Quota Management**: User-level limits.
4.  **Client SDKs (Python/TS)**: Idiomatic client libraries.
5.  **Admin API Expansion**: Full CRUD for all resources.

---

## 2. Top 10 Recommended Features

These recommendations focus on enabling enterprise adoption, developer experience, and extensibility.

| #  | Feature Name | Why it matters | Difficulty |
| -- | :--- | :--- | :--- |
| 1  | **Client SDKs (Python/TypeScript)** | **UX**: Drastically reduces friction for developers integrating MCP Any into their apps. Essential for ecosystem growth. | Medium |
| 2  | **Advanced Cost & Quota Management** | **Scalability/Business**: Critical for SaaS/Enterprise use cases to prevent abuse and enable monetization (tier-based access). | Medium |
| 3  | **WASM Plugin System** | **Extensibility**: Allows users to write custom transformations/validations safely without recompiling the core server. | High |
| 4  | **File System Provider (Sandboxed)** | **UX/Utility**: Enables "Agentic" workflows where LLMs can read/write files locally in a controlled manner. | Medium |
| 5  | **Configuration Versioning & GitOps** | **Ops/Scalability**: Treat configuration as code. Auto-reload from Git repositories. | Medium |
| 6  | **Interactive Playground** | **UX**: A UI component to test tools immediately after registration, debugging inputs/outputs visually. | Low |
| 7  | **Cloud Provider Identity Federation** | **Security**: Support AWS SigV4, GCP OIDC, and Azure AD natively to avoid managing long-lived keys. | High |
| 8  | **Comprehensive E2E Testing Framework** | **Reliability**: A framework for users to write tests for their *configurations* to ensure upstream changes don't break tools. | Medium |
| 9  | **Policy-as-Code (OPA/Rego)** | **Security**: More expressive than simple RBAC. Allow complex rules like "User X can only call Tool Y if argument Z < 100". | High |
| 10 | **Marketplace / Config Hub** | **UX**: A centralized registry or CLI command to pull community-maintained configs for popular services (e.g., `mcpany pull github`). | Medium |

---

## 3. Codebase Health & Refactoring

### Current State
The codebase is well-structured with clear separation of concerns (`upstream`, `auth`, `transformer`, `ui`). The use of interfaces for `Upstream` and `Tool` is robust.

### Areas for Improvement

1.  **RBAC Integration**:
    *   **Issue**: `server/pkg/auth/rbac.go` contains the logic, but it needs to be consistently applied across all access points (gRPC, HTTP, Admin API).
    *   **Action**: Implement a unified RBAC middleware that intercepts all incoming requests and enforces policy based on the user context.

2.  **Admin API Completeness**:
    *   **Issue**: `server/pkg/admin/server.go` currently implements read-only operations and cache clearing. Registration is handled separately via `serviceregistry`.
    *   **Action**: Consolidate management into a RESTful Admin API that supports full CRUD (Create, Read, Update, Delete) for services, policies, and users.

3.  **UI/Backend Coupling**:
    *   **Issue**: The UI is a separate Next.js app. Ensure the API contract between the UI and the Backend is stable and versioned (e.g., using the Admin API).
    *   **Action**: Formalize the "Management API" spec (OpenAPI/Protobuf) used by the UI.

4.  **Error Handling & Validation**:
    *   **Issue**: Transformation errors or upstream failures need to be propagated clearly to the LLM/Client.
    *   **Action**: Standardize error responses (MCP standard error codes) and improve validation feedback for configuration files.
