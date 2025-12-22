# Product Evolution Plan

## 1. Updated Roadmap

The `ROADMAP.md` (located at `docs/roadmap.md`) has been reconciled with the codebase.

### Implemented Features (Moved from Planned)
*   **Dynamic Web UI (Beta)**: Now marked as implemented (Beta). Source located in `ui/`.
*   **Admin Management API**: Now marked as implemented. Source located in `server/pkg/admin`.
*   **Role-Based Access Control (RBAC)**: Confirmed as implemented (Code exists in `server/pkg/auth/rbac.go`), though pending full integration into middleware.

### Current Roadmap Snapshot

#### Service Types
*   gRPC, HTTP, OpenAPI, GraphQL, Stdio, MCP-to-MCP Proxy, WebSocket, WebRTC, SQL

#### Security & Auth
*   API Key, Bearer Token, OAuth 2.0, Secrets Management, IP Allowlisting, Webhooks, RBAC

#### Policies & Observability
*   Caching, Rate Limiting, Resilience, Tracing, Metrics, Logging, Audit

#### Core
*   Dynamic Tool Registration, Message Bus, Transformation

## 2. Top 10 Recommended Features

These features are selected based on industry standards for API gateways and MCP servers, addressing gaps in the current implementation.

| Feature Name | Why it matters | Difficulty |
| :--- | :--- | :--- |
| **1. WASM Plugins** | **Extensibility/Security**. Allows users to write custom transformation/validation logic in any language (Rust/Go/TS) and run it safely in a sandbox. Critical for "Edge" logic. | High |
| **2. File System Provider** | **Utility**. A core use case for MCP is interacting with local files (coding agents). Needs strict sandboxing and allow-listing. | Medium |
| **3. Persistent State / KV Store** | **Statefulness**. MCP servers are often stateless, but agents need "memory". A simple KV store provider (backed by Redis/Disk) allows tools to save state between calls. | Medium |
| **4. Cost & Quota Management** | **Enterprise**. Beyond rate limiting. Track "token usage" (if proxied) or "call counts" per user/tenant for billing or strict quota enforcement. | Medium |
| **5. Client SDKs (Python/TS)** | **Adoption**. While MCP is standard, idiomatic SDKs make it easier to build clients *for* MCP Any or consume it. | Medium |
| **6. Interactive Playground** | **UX**. Complete the "Coming Soon" feature in the Web UI. Allow users to construct MCP calls and see results/traces in real-time. | Low |
| **7. Async/Long-running Tool Support** | **Scalability**. Some tools take minutes. Support 202-Accepted patterns or "Job" based MCP tools where the client polls for result. | High |
| **8. Secret Rotation / Dynamic Secrets** | **Security**. Integrate with Vault or AWS Secrets Manager to fetch secrets dynamically instead of static config, and support rotation. | High |
| **9. Agent Framework Integrations** | **Ecosystem**. Official adapters or guides for LangChain, AutoGen, and LlamaIndex. "How to use MCP Any with LangChain". | Low |
| **10. Container/Docker Provider** | **Sandboxing**. A provider that spins up ephemeral Docker containers to run CLI tools, ensuring total isolation for dangerous tools. | High |

## 3. Codebase Health

### Areas for Improvement

1.  **RBAC Middleware Integration**:
    *   **Issue**: `server/pkg/auth/rbac.go` contains the `RBACEnforcer` logic, but it is currently **not wired** into the global `AuthMiddleware` or `Registry`. The `AuthMiddleware` verifies identity but does not enforce role checks.
    *   **Recommendation**: Create a dedicated `RBACMiddleware` in `server/pkg/middleware` that utilizes `RBACEnforcer` and checks against configured policies before allowing tool execution.

2.  **Client Package Ambiguity**:
    *   **Issue**: `server/pkg/client` exists but appears to be an internal abstraction for connecting to *upstream* services (gRPC/HTTP), not a user-facing SDK.
    *   **Recommendation**: Rename `server/pkg/client` to `server/pkg/upstream_client` or similar to avoid confusion. Develop actual client SDKs in separate repositories or top-level folders.

3.  **Documentation Coverage**:
    *   **Issue**: While the structure is good, many feature docs are likely stubs. `AGENTS.md` was missing in root (found in `server/`).
    *   **Recommendation**: Ensure `docs/` is the single source of truth. Move `server/AGENTS.md` content to `docs/developer_guide.md` or keep it but reference it clearly.

4.  **Test Coverage**:
    *   **Issue**: RBAC tests exist but are unit tests for the struct. Integration tests ensuring a user with `role: viewer` cannot access `admin` tools are needed.
    *   **Recommendation**: Add integration tests in `server/tests` covering the full authz flow.
