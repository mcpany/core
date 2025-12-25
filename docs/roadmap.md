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

- [x] Dynamic Tool Registration
- [x] Message Bus (NATS, Kafka)
- [x] [Structured Output Transformation](features/transformation.md) (JQ/JSONPath)
- [x] [Admin Management API](features/admin_api.md)
- [x] [Dynamic Web UI](features/web_ui.md) (Beta)

### Transformation

- [x] Structured Output Transformation (JQ/JSONPath)

## Upcoming Features (High Priority)

### 1. WASM Plugins

**Why:** Allow users to deploy safe, sandboxed custom logic for transformations or validations.
**Status:** Planned

### 2. Cloud Storage Support (S3, GCS)

**Why:** Extend filesystem capabilities to cloud object storage.
**Status:** Planned

### 3. Cost & Quota Management

**Why:** Track token usage or call counts per user/profile and enforce strict quotas.
**Status:** Planned

### 4. Client SDKs (Python/TS)

**Why:** Provide idiomatic wrappers for connecting to MCP Any.
**Status:** Planned

## Deprecated / Obsolete

- (None currently identified)
