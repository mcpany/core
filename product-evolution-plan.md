# Product Evolution Plan

## 1. Updated Roadmap

The following is the reconciled roadmap as of today. Completed items have been verified against the codebase.

### Current Status

#### Implemented Features
- [x] [**Service Types**](./server/docs/features/service-types.md): gRPC, HTTP, OpenAPI, GraphQL, Stdio, WebSocket, WebRTC.
- [x] [**Dynamic UI**](./server/docs/features/dynamic-ui.md): Web-based interface for managing services (`ui/`).
- [x] [**RBAC**](./server/docs/features/rbac.md): Role-Based Access Control (`server/pkg/middleware/rbac.go`).
- [x] [**Transport Protocols (Kafka)**](./server/docs/features/kafka.md): Add support for asynchronous communication via Kafka.
- [x] [**Transport Protocols (NATS)**](./server/docs/features/nats.md): Support for NATS as a message bus.
- [x] [**Caching**](./server/docs/features/caching/README.md).
- [x] [**Rate Limiting**](./server/docs/features/rate-limiting/README.md).
- [x] [**Resilience**](./server/docs/features/resilience/README.md).
- [x] [**Health Checks**](./server/docs/features/health-checks.md).
- [x] [**Schema Validation**](./server/docs/features/schema-validation.md).
- [x] [**Service Profiles**](./server/docs/features/profiles_and_policies/README.md).
- [x] [**Secrets Management**](./server/docs/features/security.md).
- [x] [**Distributed Tracing**](./server/docs/features/tracing/README.md).
- [x] [**Automated Documentation Generation**](./server/docs/features/documentation_generation.md).
- [x] [**IP Allowlisting**](./server/docs/features/security.md).
- [x] [**Webhooks**](./server/docs/features/webhooks/README.md).
- [x] [**Advanced Authentication**](./server/docs/features/authentication/README.md).

#### Ongoing Goals
- [ ] **Expand Test Coverage**: Increase unit and integration test coverage for all existing and new features.
- [ ] **Improve Error Handling**.

#### Long-Term Goals
- [ ] **WASM Plugin Support**.
- [ ] **More Service Types**.
- [ ] **Config Registry**.
- [ ] **Client SDKs** (End-user facing).

---

## 2. Top 10 Recommended Features

Based on the analysis of current capabilities and industry standards for AI/Infrastructure tools, here are the top 10 recommended features to prioritize:

1.  **Vector Search Integration (RAG Support)**
    *   *Why*: Retrieval-Augmented Generation is the primary use case for MCP. Supporting vector databases (Pinecone, Weaviate, pgvector) as first-class citizens will exponentially increase utility.
    *   *Difficulty*: Medium

2.  **OIDC / SSO Support (Incoming Auth)**
    *   *Why*: Enterprise adoption requires integration with existing identity providers (Google, Okta, Azure AD) rather than just API keys.
    *   *Difficulty*: Medium

3.  **Kubernetes Operator**
    *   *Why*: To truly be "Ops Friendly", we need a native way to manage MCP Any instances and configurations via CRDs in K8s environments.
    *   *Difficulty*: High

4.  **S3 / Blob Storage Adapter**
    *   *Why*: A generic file operations tool (Upload/Download/List) for S3-compatible storage is a very common requirement for AI agents.
    *   *Difficulty*: Low

5.  **PII Redaction / Data Masking**
    *   *Why*: Security and compliance (GDPR/CCPA). Automatically redact sensitive info (emails, credit cards) from logs and traces.
    *   *Difficulty*: Medium

6.  **Interactive Tool Playground (UI)**
    *   *Why*: Improve Developer Experience (DX). Allow users to test tools directly from the UI (like Swagger UI) before connecting an agent.
    *   *Difficulty*: Medium

7.  **Cost / Token Usage Analytics**
    *   *Why*: Business intelligence. Track token usage per user/profile to enable chargeback or quota management.
    *   *Difficulty*: Medium

8.  **Structured Log Streaming**
    *   *Why*: Production readiness. Direct integration with Splunk, Datadog, or CloudWatch for logs (beyond just file/stdout).
    *   *Difficulty*: Low

9.  **Declarative Config Validation CLI**
    *   *Why*: CI/CD. A standalone CLI command to validate `config.yaml` syntax and schema without starting the server.
    *   *Difficulty*: Low

10. **Webhooks Sink (Event Source)**
    *   *Why*: Enable event-driven workflows. Allow external services to trigger MCP tools via webhooks (reverse of current Webhooks feature).
    *   *Difficulty*: Medium

---

## 3. Codebase Health

### Overview
The codebase is well-structured and modular (`pkg/` layout). However, there are some areas requiring attention.

### Issues
1.  **Linting**:
    *   There are ~102 `nolint` directives. While sometimes necessary, excessive use can hide real issues. A review is recommended.

2.  **TODOs**:
    *   Only 2 TODOs found in `server/pkg`. This is surprisingly low, which might indicate either very clean code or a lack of inline tracking for tech debt.
    *   A massive amount of TODOs (1900+) were found in `build/` (dependencies), which is expected and fine, but we should ensure our own code is tracked.

### Refactoring Candidates
*   **Protobuf Parser**: This area is complex (reflecting on proto files) and critical for gRPC support.
*   **Tool Execution Logic**: The core execution engine is a critical path and should be kept clean.

### Conclusion
The project is in a strong position feature-wise, but stability (tests) needs to be addressed before embarking on the "Kubernetes Operator" or "WASM" features.
