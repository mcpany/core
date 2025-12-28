# Roadmap

This document outlines the current status and future plans for MCP Any.

## Status: Active Development

## Implemented Features

### Service Types

- [x] [gRPC](features/service_types.md)
- [x] [HTTP](features/service_types.md)
- [x] [OpenAPI](features/service_types.md)
- [x] [GraphQL](features/service_types.md)
- [x] [Stdio](features/service_types.md)
- [x] [MCP-to-MCP Proxy](features/service_types.md)
- [x] [WebSocket](features/service_types.md)
- [x] [WebRTC](features/service_types.md)
- [x] [SQL](features/service_types.md)
- [x] [File System Provider](features/filesystem.md)

### Authentication

- [x] [API Key](features/auth.md)
- [x] [Bearer Token](features/auth.md)
- [x] [OAuth 2.0](features/auth.md)
- [x] [Role-Based Access Control (RBAC)](features/auth.md)
- [x] [Upstream mTLS](../server/docs/features/security.md) (Client Certificate Authentication)

### Policies

- [x] [Caching](features/policies.md)
- [x] [Rate Limiting](features/policies.md) (Memory & Redis)
- [x] [Resilience](features/policies.md) (Circuit Breakers & Retries)

### Observability

- [x] [Distributed Tracing](features/observability.md) (OpenTelemetry)
- [x] [Metrics](features/observability.md)
- [x] [Structured Logging](features/observability.md)
- [x] [Audit Logging](features/observability.md)

### Security

- [x] [Secrets Management](features/security.md)
- [x] [IP Allowlisting](features/security.md)
- [x] [Webhooks](features/security.md)

### Core

- [x] [Dynamic Tool Registration & Auto-Discovery](features/dynamic_registration.md)
- [x] [Message Bus (NATS, Kafka)](features/message_bus.md)
- [x] [Structured Output Transformation](features/transformation.md) (JQ/JSONPath)
- [x] [Admin Management API](features/admin_api.md)
- [x] [Dynamic Web UI](features/web_ui.md) (Beta)
- [x] [Health Checks](features/health_checks.md)
- [x] [Schema Validation](features/schema_validation.md)
- [x] [Service Profiles](features/service_profiles.md)
- [x] [Semantic Caching](../server/docs/caching.md) (SQLite/Memory with Vector Embeddings)

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

## Deprecated / Obsolete

- (None currently identified)
