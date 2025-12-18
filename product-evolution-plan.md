# Product Evolution Plan

This document outlines the strategic plan for synchronizing the project's documentation with reality and driving future innovation.

## 1. Updated Roadmap

### Current Status

#### Implemented Features

The following features are fully implemented and tested:

- **Service Types**: gRPC, HTTP, OpenAPI, GraphQL, Stdio, MCP-to-MCP Proxy, WebSocket, WebRTC.
- **Upstream Authentication**: API Key, Bearer Token, OAuth 2.0.
- **Registration**: Dynamic (gRPC) and Static (YAML/JSON).
- **Policies**: Caching, Rate Limiting (In-Memory), Resilience (Circuit Breakers & Retries).
- **Deployment**: Helm Chart, Docker.
- **Health Checks**: Implemented for all major protocols.
- **Schema Validation**: Built-in JSON Schema validation for config.
- **Observability**: Distributed Tracing (OTLP/Stdout), Logging.
- **Security**: Secrets Management, IP Allowlisting.
- **Eventing**: NATS support, Webhooks (Pre/Post call).
- **Documentation**: Automated Doc Gen.

#### High Priority (Next 1-3 Months)

- **Distributed Rate Limiting (Redis)**: Essential for scaling beyond a single instance.
- **Runtime Argument Validation**: Critical security feature to prevent injection attacks.
- **Audit Logging**: Required for enterprise compliance.
- **Kafka Transport**: For high-throughput async messaging.
- **RBAC**: For multi-tenant security.

## 2. Top 10 Recommended Features

Based on the strategic feature extraction analysis, the following features are recommended for immediate implementation to address security, scalability, and UX gaps.

| Rank | Feature Name | Why it matters | Implementation Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **Distributed Rate Limiting (Redis)** | **Scalability**: Current rate limiting is in-memory (per instance). Distributed limiting is essential for running multiple replicas in production. | Medium |
| 2 | **Runtime Argument Validation** | **Security**: Completes the "Security Policies" feature by enforcing regex checks on tool arguments at runtime, preventing injection attacks or misuse. | Medium |
| 3 | **Kafka Transport Support** | **Scalability**: Completes the message bus vision. Kafka is industry-standard for high-throughput event streaming. | Medium |
| 4 | **Audit Logging** | **Security/Compliance**: Record *who* accessed *what* tool and *when*. Critical for enterprise adoption. | Medium |
| 5 | **Admin/Management API** | **UX/Automation**: Allow programmatic configuration updates (add/remove services) without restarting or editing files directly. Precursor to the Dynamic UI. | High |
| 6 | **Structured Logging (JSON)** | **Observability**: Switch from text logs to JSON for better ingestion by log aggregators (Splunk, ELK). | Low |
| 7 | **RBAC (Role-Based Access Control)** | **Security**: Restrict which users can access which profiles or tools. Essential for multi-tenant environments. | High |
| 8 | **Dynamic UI** | **UX**: A web interface to visualize and manage the server. | High |
| 9 | **WASM Plugins** | **Extensibility**: Allow users to write custom transformation/validation logic in any language (Rust, Go, TS) without recompiling the server. | High |
| 10 | **Formalized Client SDKs** | **Developer Experience**: Provide official libraries (Python/TS) to interact with MCP Any's specific features easily. | Medium |

## 3. Codebase Health Report

This section highlights areas of the codebase that require attention to ensure long-term maintainability and stability.

### Critical Areas

*   **Security Policy Implementation**: The `shouldBlock` function in `pkg/upstream/http/http.go` explicitly comments that argument regex matching is not performed. This leaves a functional gap in the security policy feature where users might expect protection that doesn't exist.
*   **Tool Discovery Complexity**: The `createAndRegisterHTTPTools` function in `pkg/upstream/http/http.go` has high cyclomatic complexity (`//nolint:gocyclo`). This makes it difficult to test and maintain. It should be refactored into smaller, testable components.
*   **Testing Gaps**: While there are test files, "fine-grained request validation" and other edge cases in policy enforcement need rigorous integration tests to ensure security guarantees are met.
*   **Admin API**: The current `pkg/admin` implementation is very minimal (only cache clearing). It needs significant expansion to support full management capabilities.

### Recommendations

1.  **Refactor Tool Registration**: Break down `createAndRegisterHTTPTools` and similar functions in other upstreams.
2.  **Implement Runtime Interceptors**: Use a middleware or interceptor pattern to handle security policies (argument validation) at request time, rather than trying to do it at registration time.
3.  **Expand Admin API**: Build out the Admin API to support adding/removing services dynamically, which is a prerequisite for a robust Dynamic UI.
