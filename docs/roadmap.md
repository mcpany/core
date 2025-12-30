# Roadmap

This document outlines the current status and future plans for MCP Any.

## Status: Active Development

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
- [x] [SQL](features/service-types.md)
- [x] [File System Provider](features/filesystem.md)

### Authentication

- [x] [API Key](features/authentication/README.md)
- [x] [Bearer Token](features/authentication/README.md)
- [x] [OAuth 2.0](features/authentication/README.md)
- [x] [Role-Based Access Control (RBAC)](features/rbac.md)
- [x] [Upstream mTLS](features/security.md) (Client Certificate Authentication)

### Policies

- [x] [Caching](features/caching/README.md)
- [x] [Rate Limiting](features/rate-limiting/README.md) (Memory & Redis)
- [x] [Resilience](features/resilience/README.md) (Circuit Breakers & Retries)

### Observability

- [x] [Distributed Tracing](features/tracing/README.md) (OpenTelemetry)
- [x] [Metrics](features/monitoring/README.md)
- [x] [Structured Logging](features/monitoring/README.md)
- [x] [Audit Logging](features/audit_logging.md)

### Security

- [x] [Secrets Management](features/security.md)
- [x] [IP Allowlisting](features/security.md)
- [x] [Webhooks](features/webhooks/README.md)
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

## Upcoming Features (High Priority)

### 1. WASM Plugins

**Why:** Allow users to deploy safe, sandboxed custom logic for transformations or validations without recompiling the server.
**Status:** Planned

### 2. Cloud Storage Support (S3, GCS)

**Why:** Extend the filesystem provider to support cloud object storage, allowing AI agents to interact with S3/GCS buckets as if they were local directories.
**Status:** Planned

### 3. Token-Based Quota Management

**Why:** While Rate Limiting is implemented, "Cost" tracking (token usage accounting) and strict token quotas per user/tenant are needed for enterprise billing controls.
**Status:** Planned

### 4. Kubernetes Operator

**Why:** Simplify deployment and lifecycle management of MCP Any instances in Kubernetes environments, enabling GitOps workflows.
**Status:** Recommended

### 5. Client SDKs (Python/TS)

**Why:** Provide idiomatic wrappers for connecting to MCP Any, simplifying integration for developers building custom AI agents.
**Status:** Planned

## Critical User Journeys (Upcoming)

### Enterprise & Operations

1.  **Kubernetes Operator**: Automate deployment, scaling, and lifecycle management of MCP Any instances in K8s.
2.  **Terraform Provider**: Manage MCP resources (Sources, Tools, Policies) via "Configuration as Code".
3.  **Helm Chart Official Support**: Provide a production-ready Helm chart with auto-scaling and monitoring presets. (Helm Chart exists, need verification of full support).
4.  **Multi-Region Federation**: Link multiple MCP Any instances across regions for low-latency tool access.
5.  **Active-Active High Availability**: Support leaderless clustering for zero-downtime upgrades and failure tolerance.
6.  **Disaster Recovery Playbook**: Automated backup/restore of state and configuration to S3/GCS.
7.  **Dynamic Secret Rotation**: Integration with HashiCorp Vault / AWS Secrets Manager for zero-touch secret handling.
8.  **[x] CI/CD Config Validator CLI**: A standalone binary to validate `config.yaml` in pipelines before deployment.
9.  **Automated Dependency Updates**: "Dependabot" for MCP Tools - auto-update tool definitions when upstreams change.
10. **Service Mesh Sidecar Mode**: Run MCP Any as a lightweight sidecar for application pods.

### Security & Compliance

11. **[x] Data Loss Prevention (DLP)**: Middleware to redact PII (emails, SSNs) from logs and tool inputs/outputs.
12. **Downstream mTLS**: Enforce mutual TLS for agents collecting to MCP Any (Zero Trust).
13. **SSO with SAML/OIDC**: Enterprise identity integration for the Admin UI and RBAC.
14. **Just-In-Time (JIT) Access**: Temporary elevation of privileges for specific tools (e.g., "Grant Write access for 1 hour").
15. **Audit Log Export**: Push audit logs to Splunk, Datadog to Cloud Logging in real-time.
16. **Fine-Grained ABAC**: Attribute-Based Access Control (e.g. "Only allow production tools during business hours").
17. **Tool Signature Verification**: Enforce that loaded WASM/Binary tools are signed by a trusted key.
18. **Vulnerability Scanning Integration**: Auto-scan registered tool container images for CVEs.
19. **Policy dry-run mode**: Test new security policies on traffic without blocking (shadow mode).
20. **Compliance Reports**: Generate PDF reports of user activity for SOC2/ISO audits.

### Observability & Insights

21. **Custom Dashboards**: Drag-and-drop UI to create dashboards from MCP metrics.
22. **Cost Attribution**: Track token usage and "cost" per user/team/project.
23. **Alerting Rules Integration**: Built-in Prometheus alerting rules for high error rates or latency.
24. **Request/Response Replay**: "TiVo" for tool interactions - replay past requests for debugging.
25. **Distributed Tracing Sampling Control**: Dynamic sampling rates based on tenant or error-rate.
26. **SLO Management**: Define and track Service Level Objectives (availability, latency) within the UI.
27. **Semantic Search over Logs**: Use embeddings to search audit logs (e.g., "Show me all database drops").
28. **Tool Usage Analytics**: Heatmaps of most used tools and arguments.
29. **Anomaly Detection**: ML-based detection of unusual tool usage patterns.
30. **Webhook Notifications**: Slack/PagerDuty alerts for critical system events.

### Developer Experience & Core

31. **WASM Plugin System**: Safe, sandboxed custom transformers and checkers.
32. **Cloud Storage Provider (S3/GCS)**: Treat buckets as filesystems for agents.
33. **Interactive Playground 2.0**: Enhanced UI to test tools with auto-generated forms and mock data.
34. **Client SDKs (Python/TS)**: Idiomatic client libraries for agents.
35. **Local Emulator**: CLI command to run a lightweight in-memory MCP server for dev.
36. **Language Server Protocol (LSP)**: IDE support for `config.yaml` editing (auto-complete tools).
37. **Hot Reload**: Reload configuration without restarting the server process.
38. **Type-Safe Tool Chaining**: Define "Workflows" where output of Tool A feeds Tool B (Server-side).
39. **Mock Provider**: A built-in provider that returns static responses for testing agents.
40. **Doc Generator**: Generate static HTML documentation site from registered tools.

### Connectivity & Integration

41. **Database Connectors (SQL)**: Native support for PostgreSQL/MySQL as "Sources" (RAG included).
42. **Salesforce Integration**: Official connector for CRM data.
43. **Jira/Confluence Connector**: Read/Write tickets and docs.
44. **Slack/Discord Bot Gateway**: Expose tools directly as chat commands.
45. **Email Server Gateway**: Trigger tools via inbound email (SMTP/IMAP).
46. **Browser Automation Provider**: Headless browser tool for "Read Webpage" capabilities.
47. **GraphQL Subscriptions**: Support real-time data push from GraphQL upstreams.
48. **Binary Protocol Support**: Protobuf over WebSocket for high-performance low-bandwidth agents.
49. **Edge Computing Support**: Optimized build for Cloudflare Workers / AWS Lambda.
50. **Air-Gapped Mode**: Full offline operation with bundled dependencies and local docs.
