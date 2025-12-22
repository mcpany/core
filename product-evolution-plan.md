# Product Evolution Plan

## Phase 1: Roadmap Reconciliation (Completed)

We have synchronized the project documentation with the current codebase reality.
- **RBAC**: Moved to Implemented.
- **Dynamic Web UI**: Moved to Implemented (Beta).
- **Admin Management API**: Documented and linked.
- **Docs**: Created `docs/features/admin_api.md`.

## Phase 2: Strategic Feature Extraction

We have identified the following Top 10 Recommended Features to drive future innovation.

### Top 10 Recommended Features

| Rank | Feature Name | Why it matters | Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **Strict Domain-Based Egress Control** | **Security**: IP allowlists are brittle. We need domain-based filtering (e.g., `*.github.com`) to safely run untrusted configs. | Medium |
| 2 | **Secret Redaction & Masking** | **Security**: Prevent API keys and PII from leaking into logs, traces, or audit trails. Essential for enterprise adoption. | Medium |
| 3 | **Human-in-the-Loop Approval** | **Security/UX**: For "dangerous" tools (e.g., `DELETE`, `EXEC`), require a human admin to approve the action via the UI before execution. | High |
| 4 | **Data Loss Prevention (DLP)** | **Security**: Scan tool outputs for patterns like Credit Card numbers or AWS keys before returning them to the LLM. | High |
| 5 | **WASM Plugins** | **Scalability**: Allow users to write custom transformation/validation logic in Rust/Go (WASM) without recompiling the server. | High |
| 6 | **Vector Database Integration** | **UX**: Native support for Pinecone/Weaviate to allow agents to "Search Knowledge Base" easily. | Medium |
| 7 | **Advanced Caching (Stale-While-Revalidate)** | **UX/Performance**: Return stale data immediately while fetching fresh data in the background to reduce AI latency. | Low |
| 8 | **Service Mesh Integration** | **Scalability**: Native support for Istio/Linkerd for mTLS and advanced traffic management in Kubernetes. | Medium |
| 9 | **Tenant Isolation (Strong)** | **Security**: Beyond RBAC, ensure physical or logical separation of resources (e.g., separate namespaces/processes) for high-security environments. | High |
| 10 | **Horizontal Scaling / Clustering** | **Scalability**: Leader election and shared state coordination for distributed deployments. | High |

## Phase 3: Codebase Health Report

### 1. Middleware Fail-Safe Logic
- **Issue**: In `server/pkg/middleware/call_policy.go`, if `ServiceInfo` is missing, the code currently allows execution (Fail-Open) or proceeds.
- **Recommendation**: Critical security middleware should Fail-Closed (Deny) if configuration or context is missing.

### 2. Frontend Integration
- **Issue**: The `ui/` directory contains a Next.js application, but it needs tight integration with the `Admin Management API` to be fully functional.
- **Recommendation**: Prioritize binding the UI forms to the `/v1/services` endpoints.

### 3. Test Coverage
- **Issue**: While unit tests exist for middleware, End-to-End (E2E) tests simulating real MCP tool calls with Auth and Policies enabled are needed to prevent regressions.
- **Recommendation**: Add a suite of integration tests in `verification/`.

## Updated Roadmap

### Status: Active Development

### Implemented Features
- **Service Types**: gRPC, HTTP, OpenAPI, GraphQL, Stdio, MCP-to-MCP, WebSocket, WebRTC, SQL
- **Authentication**: API Key, Bearer Token, OAuth 2.0
- **Policies**: Caching, Rate Limiting (Redis), Resilience
- **Observability**: Tracing (OTEL), Metrics, Logging, Audit
- **Security**: Secrets, IP Allowlist, Webhooks, RBAC
- **Core**: Dynamic Registration, Message Bus, Transformations, Admin API, Web UI (Beta)

### Upcoming Features (High Priority)
1.  **Strict Domain-Based Egress Control** (New)
2.  **Secret Redaction & Masking** (New)
3.  **Human-in-the-Loop Approval** (New)
4.  **WASM Plugins**
5.  **File System Provider**
6.  **Cost & Quota Management**
7.  **Client SDKs (Python/TS)**
