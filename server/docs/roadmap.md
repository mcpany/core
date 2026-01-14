# Roadmap

This document outlines the current status and future plans for MCP Any.

## Status: Active Development

## Strategic Context (Jan 2026)

### Market Research & Alignment
Based on a review of the MCP ecosystem (mcp.so, LobeHub, GitHub, Docker), we identified key opportunities:

- **"Debugging is Hell"**: Developers need "Traffic Replay" and "Agent Debugger" (Solution: Observability++).
- **Security & Trust**: Prompt Injection and Data Exfiltration risks (Solution: Guardrails & Granular Scopes).
- **Tool Discovery & Config**: Manual config is error-prone (Solution: K8s Operator & Terraform Provider).

### GitHub Issue Insights
- **Context Bloat**: Token explosions in large repos (Action: Context Optimization).
- **Installation Failures**: Environment mismatches (Action: Doctor API).
- **Security**: Command injection fears (Action: WASM Sandboxing).

## Feature Priorities (Jan 2026)

### Top 10 Recommended Features

| Rank | Feature Name | Why it matters | Difficulty |
| :--- | :--- | :--- | :--- |
| 1 | **Kubernetes Operator V2** | **Scalability/Ops**: Robust automation for enterprise deployment. | High |
| 2 | **Browser Automation Provider** | **Feature**: High-demand capability (Playwright) for live web interaction. | High |
| 3 | **Multi-Region Federation** | **Scalability**: Link instances for low latency. | High |
| 4 | **Active-Active High Availability** | **Reliability**: Zero-downtime upgrades and failure tolerance. | High |
| 5 | **Disaster Recovery Playbook** | **Ops**: Automated backup/restore. | Medium |
| 6 | **Dynamic Secret Rotation** | **Security**: Integration with Vault/AWS Secrets Manager. | High |
| 7 | **Downstream mTLS** | **Security**: Zero Trust for agents. | Medium |
| 8 | **Just-In-Time (JIT) Access** | **Security**: Temporary privilege elevation. | High |
| 9 | **Audit Log Export** | **Compliance**: Real-time push to Splunk/Datadog. | Medium |
| 10 | **Cost Attribution** | **Observability**: Track token usage/cost per user. | Medium |


### Feature Gap & Technical Feasibility

| Feature | Status | Feasibility / Strategy |
| :--- | :--- | :--- |
| **Browser Automation** | Missing | **High**: Implement `server/pkg/upstream/browser` using Playwright-go. |
| **K8s Operator V2** | In Progress | **High**: Enhance `k8s/operator` with CRDs and controller logic. |

## Implemented Features

### Service Types

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

### Authentication

- [x] [API Key](features/authentication/README.md)
- [x] [Bearer Token](features/authentication/README.md)
- [x] [OAuth 2.0](features/authentication/README.md)
- [x] [Role-Based Access Control (RBAC)](features/rbac.md)
- [x] [Upstream mTLS](features/security.md) (Client Certificate Authentication)
- [x] [SSO Integration (OIDC/SAML)](features/sso.md)

### Policies

- [x] [Caching](features/caching/README.md)
- [x] [Rate Limiting](features/rate-limiting/README.md) (Memory & Redis)
- [x] [Resilience](features/resilience/README.md) (Circuit Breakers & Retries)
- [x] [Prompt Injection Guardrails](features/guardrails.md)

### Observability

- [x] [Distributed Tracing](features/tracing/README.md) (OpenTelemetry)
- [x] [Metrics](features/monitoring/README.md)
- [x] [Structured Logging](features/monitoring/README.md)
- [x] [Audit Logging](features/audit_logging.md) (File, SQL, Splunk, Datadog)
- [x] [Context Consumption Metrics](features/monitoring/README.md) (Token Counting)
- [x] [Cost Attribution](features/rate-limiting/README.md) (Token Estimation)
- [x] [Agent Debugger & Inspector](features/debugger.md)
- [x] [Context Optimizer Middleware](features/context_optimizer.md)

### Security

- [x] [Secrets Management](features/security.md)
- [x] [IP Allowlisting](features/security.md)
- [x] [Webhooks](features/webhooks/README.md) (Pre/Post-Call)
- [x] [Webhooks Sidecar](features/webhooks/sidecar.md) (Context Optimization)
- [x] [Data Loss Prevention (DLP)](features/security.md)

### Core

- [x] [Dynamic Tool Registration & Auto-Discovery](features/dynamic_registration.md)
- [x] [Message Bus (NATS, Kafka)](features/message_bus.md)
- [x] [Structured Output Transformation](features/transformation.md) (JQ/JSONPath)
- [x] [Admin Management API](features/admin_api.md)
- [x] [Dynamic Web UI](features/dynamic-ui.md) (Beta)
- [x] [Health Checks](features/health-checks.md)
- [x] [Schema Validation](features/schema-validation.md)
- [x] [Service Profiles](features/profiles_and_policies/README.md)
- [x] [Semantic Caching](features/caching/README.md) (SQLite/Memory with Vector Embeddings)
- [x] [WASM Plugin System](features/wasm.md)
- [x] [Universal Connector Runtime](features/connector_runtime.md)
- [x] [Terraform Provider](features/terraform.md)

## Developer Experience & Tooling

- [x] [CI/CD Config Validator CLI](features/mcpctl.md) (includes Doctor API)
- [x] [Doc Generator](features/doc_generator.md)
- [x] [Hot Reload](features/hot_reload.md)
- [x] [Helm Chart Official Support](features/helm.md)

## Critical User Journeys (Upcoming)

### Enterprise & Operations

1.  **Kubernetes Operator**: Automate deployment, scaling, and lifecycle management of MCP Any instances in K8s. (Partially implemented in `k8s/operator`)
2.  **Multi-Region Federation**: Link multiple MCP Any instances across regions for low-latency tool access.
3.  **Active-Active High Availability**: Support leaderless clustering for zero-downtime upgrades and failure tolerance.
4.  **Disaster Recovery Playbook**: Automated backup/restore of state and configuration to S3/GCS.
5.  **Dynamic Secret Rotation**: Integration with HashiCorp Vault / AWS Secrets Manager for zero-touch secret handling.
6.  **Automated Dependency Updates**: "Dependabot" for MCP Tools - auto-update tool definitions when upstreams change.
7.  **Service Mesh Sidecar Mode**: Run MCP Any as a lightweight sidecar for application pods.

### Security & Compliance

8.  **Downstream mTLS**: Enforce mutual TLS for agents collecting to MCP Any (Zero Trust).
9.  **Just-In-Time (JIT) Access**: Temporary elevation of privileges for specific tools (e.g., "Grant Write access for 1 hour").
10. **Fine-Grained ABAC**: Attribute-Based Access Control (e.g. "Only allow production tools during business hours").
12. **Tool Signature Verification**: Enforce that loaded WASM/Binary tools are signed by a trusted key.
13. **Vulnerability Scanning Integration**: Auto-scan registered tool container images for CVEs.
14. **Policy dry-run mode**: Test new security policies on traffic without blocking (shadow mode).
15. **Compliance Reports**: Generate PDF reports of user activity for SOC2/ISO audits.

### Observability & Insights

16. **Custom Dashboards**: Drag-and-drop UI to create dashboards from MCP metrics.
17. **Alerting Rules Integration**: Built-in Prometheus alerting rules for high error rates or latency.
19. **Request/Response Replay**: "TiVo" for tool interactions - replay past requests for debugging.
20. **Distributed Tracing Sampling Control**: Dynamic sampling rates based on tenant or error-rate.
21. **SLO Management**: Define and track Service Level Objectives (availability, latency) within the UI.
22. **Semantic Search over Logs**: Use embeddings to search audit logs (e.g., "Show me all database drops").
23. **Tool Usage Analytics**: Heatmaps of most used tools and arguments.
24. **Anomaly Detection**: ML-based detection of unusual tool usage patterns.
25. **Webhook Notifications**: Slack/PagerDuty alerts for critical system events.

### Connectivity & Integration

26. **Salesforce Integration**: Official connector for CRM data.
27. **Jira/Confluence Connector**: Read/Write tickets and docs.
28. **Slack/Discord Bot Gateway**: Expose tools directly as chat commands.
29. **Email Server Gateway**: Trigger tools via inbound email (SMTP/IMAP).
30. **Browser Automation Provider**: Headless browser tool for "Read Webpage" capabilities.
31. **GraphQL Subscriptions**: Support real-time data push from GraphQL upstreams.
32. **Binary Protocol Support**: Protobuf over WebSocket for high-performance low-bandwidth agents.
33. **Edge Computing Support**: Optimized build for Cloudflare Workers / AWS Lambda.
34. **Air-Gapped Mode**: Full offline operation with bundled dependencies and local docs.

### Developer Experience

35. **Enhanced Configuration Validation**: Implement strict schema validation (using JSON Schema) to catch structure errors like `service_config` wrapper usage at parsing time.
36. **Interactive `mcp init` CLI**: A wizard to generate valid configuration files interactively, reducing copy-paste errors from docs.

## Codebase Health Report

### Critical Areas

- **Rate Limiting Complexity**: `server/pkg/middleware/ratelimit.go` mixes local/Redis logic. Needs refactoring into strategies.
- **Filesystem Provider Monolith**: `server/pkg/upstream/filesystem/upstream.go` handles too many types. Split into factory pattern.
- **Test Coverage for Cloud Providers**: S3/GCS tests are missing/mocked. Need local emulation (MinIO).
- **Webhooks "Test" Code**: `server/cmd/webhooks` needs formalization if intended for production (Sidecar pattern).
- **SDK Consolidation**: `server/pkg/client` should ideally be in a separate repository to be used by other Go clients without pulling in the whole server.

### Recommendations

1.  **Refactor Filesystem Upstream**: Split `upstream.go`.
2.  **Refactor Rate Limiting**: Split into `RateLimiterStrategy` interface.
3.  **Formalize Webhook Server**: Polish `server/cmd/webhooks` as a Sidecar.
4.  **Standardize Configuration**: Consistent config patterns (Done: fixed documentation/error handling for `service_config`).
5.  **Consolidate SDKs**: Move `server/pkg/client` to separate repo.
