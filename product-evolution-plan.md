# Product Evolution Plan: MCP Any

**Date:** 2025-05-24
**Author:** Jules (Lead Product Manager & Principal Architect)

This document outlines the strategic roadmap for synchronizing the project's documentation with reality and driving future innovation.

---

## 1. Updated Roadmap

### Status: Active Development

### Implemented Features

#### Service Types
- [x] [gRPC](docs/features/service-types.md)
- [x] [HTTP](docs/features/service-types.md)
- [x] [OpenAPI](docs/features/service-types.md)
- [x] [GraphQL](docs/features/service-types.md)
- [x] [Stdio](docs/features/service-types.md)
- [x] [MCP-to-MCP Proxy](docs/features/service-types.md)
- [x] [WebSocket](docs/features/service-types.md)
- [x] [WebRTC](docs/features/service-types.md)
- [x] [SQL](docs/features/service-types.md)
- [x] [File System Provider](docs/features/filesystem.md)

#### Authentication
- [x] [API Key](docs/features/authentication/README.md)
- [x] [Bearer Token](docs/features/authentication/README.md)
- [x] [OAuth 2.0](docs/features/authentication/README.md)
- [x] [Role-Based Access Control (RBAC)](docs/features/rbac.md)
- [x] [Upstream mTLS](docs/features/security.md) (Client Certificate Authentication)

#### Policies
- [x] [Caching](docs/features/caching/README.md)
- [x] [Rate Limiting](docs/features/rate-limiting/README.md) (Memory & Redis)
- [x] [Resilience](docs/features/resilience/README.md) (Circuit Breakers & Retries)

#### Observability
- [x] [Distributed Tracing](docs/features/tracing/README.md) (OpenTelemetry)
- [x] [Metrics](docs/features/monitoring/README.md)
- [x] [Structured Logging](docs/features/monitoring/README.md)
- [x] [Audit Logging](docs/features/audit_logging.md)

#### Security
- [x] [Secrets Management](docs/features/security.md)
- [x] [IP Allowlisting](docs/features/security.md)
- [x] [Webhooks](docs/features/webhooks/README.md)
- [x] [Data Loss Prevention (DLP)](docs/features/security.md)

#### Core
- [x] [Dynamic Tool Registration & Auto-Discovery](docs/features/dynamic_registration.md)
- [x] [Message Bus (NATS, Kafka)](docs/features/message_bus.md)
- [x] [Structured Output Transformation](docs/features/transformation.md) (JQ/JSONPath)
- [x] [Admin Management API](docs/features/admin_api.md)
- [x] [Dynamic Web UI](docs/features/dynamic-ui.md) (Beta)
- [x] [Health Checks](docs/features/health-checks.md)
- [x] [Schema Validation](docs/features/schema-validation.md)
- [x] [Service Profiles](docs/features/profiles_and_policies/README.md)
- [x] [Semantic Caching](docs/features/caching/README.md) (SQLite/Memory with Vector Embeddings)

### Upcoming Features (High Priority)

#### 1. WASM Plugins
**Why:** Allow users to deploy safe, sandboxed custom logic for transformations or validations without recompiling the server.
**Status:** Planned

#### 2. Cloud Storage Support (S3, GCS)
**Why:** Extend the filesystem provider to support cloud object storage, allowing AI agents to interact with S3/GCS buckets as if they were local directories.
**Status:** Planned

#### 3. Token-Based Quota Management
**Why:** While Rate Limiting is implemented, "Cost" tracking (token usage accounting) and strict token quotas per user/tenant are needed for enterprise billing controls.
**Status:** Planned

#### 4. Kubernetes Operator
**Why:** Simplify deployment and lifecycle management of MCP Any instances in Kubernetes environments, enabling GitOps workflows.
**Status:** Recommended

#### 5. Client SDKs (Python/TS)
**Why:** Provide idiomatic wrappers for connecting to MCP Any, simplifying integration for developers building custom AI agents.
**Status:** Planned

---

## 2. Top 10 Recommended Features

| Rank | Feature Name | Why it matters | Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **WASM Plugin System** | **Scalability/Security:** Enables users to extend functionality (transformers, custom protocols) safely without modifying the core codebase. Critical for ecosystem growth. | High |
| 2 | **Cloud Storage Provider (S3/GCS)** | **UX/Feature:** Modern AI agents operate in the cloud. Accessing S3/GCS buckets as native filesystems is a baseline expectation for enterprise data access. | Medium |
| 3 | **Kubernetes Operator** | **Scalability/Ops:** Essential for enterprise adoption. Automates deployment, updates, and scaling of MCP Any clusters. | High |
| 4 | **Client SDKs (Python & TS)** | **UX/DX:** Drastically improves Developer Experience. Current raw HTTP/gRPC usage is error-prone. SDKs drive adoption. | Medium |
| 5 | **Audit Log Exporters** | **Security/Compliance:** Enterprises require logs in their SIEM (Splunk, Datadog). Storing in local SQLite is insufficient for production. | Medium |
| 6 | **SSO Integration (OIDC/SAML)** | **Security:** Mandatory for enterprise environments to manage access to the Admin UI and services via Okta/Auth0/Google. | Medium |
| 7 | **Interactive Playground 2.0** | **UX:** A robust UI with auto-generated forms to test tools (like Swagger UI/GraphiQL) makes it easier for developers to understand and debug their agents. | Medium |
| 8 | **Terraform Provider** | **Scalability/Ops:** Infrastructure as Code (IaC) is the standard. Users want to define Sources and Policies in HCL, not just YAML on disk. | High |
| 9 | **Database Connectors with RAG** | **Feature:** Simply proxying SQL is not enough. Native RAG (Indexing data into vector stores automatically) turns databases into "Knowledge Bases" for Agents. | High |
| 10 | **Downstream mTLS (Zero Trust)** | **Security:** Ensures that only authorized Agents (clients) can connect to the MCP Server, completing the Zero Trust loop. | Medium |

---

## 3. Codebase Health

### Areas Requiring Attention

#### 1. Token Estimation Logic (`server/pkg/middleware/ratelimit.go`)
*   **Issue:** The current `estimateTokenCost` function uses a crude heuristic (`char_count / 4`). This is inaccurate for different languages and tokenizers.
*   **Recommendation:** Abstract the tokenizer interface. Integrate a proper BPE tokenizer (e.g., tiktoken-go) or allow calling an external tokenization service.
*   **Severity:** Medium (Affects billing/rate-limit accuracy)

#### 2. Internal Storage Dependencies (`server/pkg/storage/sqlite`)
*   **Issue:** Critical stateful components like Audit Logging and Semantic Caching heavily rely on SQLite. This creates a bottleneck for High Availability (HA) deployments where disk is ephemeral or shared.
*   **Recommendation:** Introduce a generic `Storage` interface with implementations for PostgreSQL and MySQL. This allows the system to scale horizontally.
*   **Severity:** High (Blocks proper HA clustering)

#### 3. Data Loss Prevention (DLP) (`server/pkg/middleware/dlp.go`)
*   **Issue:** DLP currently relies on simple Regex patterns. This is prone to false positives and negatives.
*   **Recommendation:** Make the DLP provider pluggable to support external DLP services (e.g., Google DLP API, AWS Macie) or more advanced local NLP models.
*   **Severity:** Low (Acceptable for V1, but limits Enterprise Compliance)

#### 4. Redis Client Management
*   **Issue:** Multiple components (Rate Limiter, Bus) seem to manage their own Redis connections/pools.
*   **Recommendation:** Centralize Redis client management in a `pkg/redis` or `pkg/infrastructure` module to share pools and configuration (timeouts, TLS settings).
*   **Severity:** Low (Code cleanliness/Resource optimization)

#### 5. Test Coverage for Edge Cases
*   **Issue:** Presence of "repro" tests (`logging_bug_repro_test.go`) suggests reactive bug fixing.
*   **Recommendation:** Increase proactive integration testing, specifically for concurrency and error handling in the `mcpserver` core.
*   **Severity:** Medium
