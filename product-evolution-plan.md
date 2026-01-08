# Product Evolution Plan

## 1. Updated Roadmap

### Status: Active Development

### Implemented Features (Recently Completed)

- [x] **Agent Debugger & Inspector**: Middleware for traffic replay and inspection.
- [x] **Context Optimizer**: Middleware to prevent context bloat.
- [x] **Diagnostic "Doctor" API**: `mcpctl` validation and health checks.
- [x] **SSO Integration**: OIDC/SAML support via `server/pkg/middleware/sso.go`.
- [x] **Prompt Injection Guardrails**: Security middleware.
- [x] **WASM Plugin System**: Runtime for sandboxed plugins.
- [x] **Universal Connector Runtime**: Sidecar for stdio tools.
- [x] **Terraform Provider**: Initial provider implementation.

## 2. Top 10 Recommended Features

| Rank | Feature Name | Why it matters | Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **Kubernetes Operator V2** | **Scalability/Ops**: The current operator is skeletal. A robust operator is essential for enterprise adoption, enabling automated deployment, scaling, and lifecycle management. | High |
| 2 | **Browser Automation Provider** | **Feature**: "Read Webpage" is a critical user journey for agents. Integrating Playwright allows agents to interact with live web content safely. | High |
| 3 | **Multi-Region Federation** | **Scalability**: Enables low-latency access to tools by linking multiple MCP Any instances across different geographical regions. | High |
| 4 | **Active-Active High Availability** | **Reliability**: Critical for production environments to ensure zero-downtime upgrades and tolerance to node failures. | High |
| 5 | **Disaster Recovery Playbook** | **Ops**: Automated backup and restore procedures for state and configuration (to S3/GCS) are non-negotiable for enterprise SLAs. | Medium |
| 6 | **Dynamic Secret Rotation** | **Security**: Integration with Vault or AWS Secrets Manager to automatically rotate credentials without restarting servers reduces attack surface. | High |
| 7 | **Downstream mTLS** | **Security**: Enforcing mutual TLS for agents connecting to MCP Any ensures a Zero Trust architecture. | Medium |
| 8 | **Just-In-Time (JIT) Access** | **Security**: Allows temporary privilege elevation for specific tools (e.g., "Grant Write access for 1 hour"), enforcing least privilege. | High |
| 9 | **Audit Log Export** | **Compliance**: Real-time pushing of audit logs to SIEMs (Splunk, Datadog) is required for security auditing in large orgs. | Medium |
| 10 | **Cost Attribution** | **Observability**: Tracking token usage and "cost" per user/team allows for chargeback models and resource optimization. | Medium |

## 3. Codebase Health

### Critical Areas

- **Rate Limiting Complexity**: `server/pkg/middleware/ratelimit.go` mixes local memory and Redis logic. This makes it hard to test and extend.
- **Filesystem Provider Monolith**: `server/pkg/upstream/filesystem/upstream.go` handles too many types (Local, S3, GCS) in one place.
- **Test Coverage for Cloud Providers**: S3/GCS tests are largely mocked or missing. Integration tests with local emulators (MinIO) are needed.
- **Webhooks "Test" Code**: `server/cmd/webhooks` needs to be formalized into a proper sidecar if it's intended for production use.
- **SDK Consolidation**: `server/pkg/client` should ideally be in a separate repository to be used by other Go clients without pulling in the whole server.

### Recommendations

1.  **Refactor Rate Limiting**: Split into `RateLimiterStrategy` interface with `LocalStrategy` and `RedisStrategy` implementations.
2.  **Refactor Filesystem Upstream**: Use a Factory pattern to separate Local, S3, and GCS implementations.
3.  **Formalize Webhook Server**: Polish `server/cmd/webhooks` as a Sidecar.
4.  **Standardize Configuration**: Ensure consistent configuration patterns across all modules.
