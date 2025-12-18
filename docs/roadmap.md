# 🗺️ Roadmap

This document outlines the current status and future direction of the MCP Any project.

## Current Status

### Implemented Features

The following features are fully implemented and tested:

- [x] [**Service Types**](./features/service-types.md):
  - gRPC (with reflection)
  - HTTP
  - OpenAPI
  - GraphQL
  - Stdio
  - MCP-to-MCP Proxy
  - WebSocket
  - WebRTC
- **Upstream Authentication**:
  - API Key
  - Bearer Token
  - OAuth 2.0
- **Dynamic Registration**: Services can be registered at runtime via the gRPC Registration API.
- **Static Registration**: Services can be registered at startup via a YAML or JSON configuration file.
- **Advanced Service Policies**:
  - [x] [Caching](./features/caching/README.md) (`CacheConfig`)
  - [x] [Rate Limiting](./features/rate-limiting/README.md) (`RateLimitConfig`)
  - [x] [Resilience](./features/resilience/README.md) (Circuit Breakers & Retries) (`ResilienceConfig`)
- **Deployment**:
  - Official Helm Chart
  - Docker Container
- [x] [**Health Checks**](./features/health-checks.md): Implement health check endpoints for upstream services (HTTP, gRPC, WebSocket, WebRTC, Command Line).
- [x] [**Schema Validation**](./features/schema-validation.md): Integrate JSON Schema to validate configuration files before loading.
- [x] [**Service Profiles**](./features/profiles_and_policies/README.md): Categorize and selectively enable services using profiles (`--profiles` flag).
- **Configuration**:
  - Hot Configuration Reloading
- [x] [**Secrets Management**](./features/security.md): Secure handling of sensitive data (API keys, passwords) using Vault, AWS Secrets Manager, or Env Vars.
- [x] [**Distributed Tracing**](./features/tracing/README.md): Integrate OpenTelemetry for tracing requests across services.
- [x] [**Transport Protocols (NATS)**](./features/nats.md): Support for NATS as a message bus.
- [x] [**Automated Documentation Generation**](./features/documentation_generation.md): Generate markdown documentation for registered tools directly from the configuration.
- [x] [**IP Allowlisting**](./features/security.md): Restrict access to specific IP addresses/CIDRs.
- [x] [**Webhooks**](./features/webhooks/README.md): Pre-call and Post-call hooks for validation and transformation.

### Configured but Not Yet Implemented

The following features are defined in the configuration schema (`proto/config/v1/config.proto`) but are **not yet implemented** in the server logic:

- **Advanced Authentication**:
  - [x] Incoming request authentication (Profile > User > Global Priority)

For a complete list of all available configuration options, please see the [Configuration Reference](./reference/configuration.md).

## High Priority (Next 1-3 Months)

- [ ] **Distributed Rate Limiting (Redis)**: Current rate limiting is in-memory. Redis support is needed for distributed deployments.
- [ ] **Security Policies**: Implement fine-grained request validation policies (runtime argument validation).
- [ ] **Dynamic UI**: Build a web-based UI for managing upstream services dynamically.
- [ ] **RBAC**: Role-Based Access Control for managing user permissions.
- [ ] **Audit Logging**: Record who accessed what tool and when.
- [ ] **Transport Protocols (Kafka)**: Add support for asynchronous communication via Kafka.

## Ongoing Goals

- [x] **Implement Advanced Authentication**: Add support for OAuth 2.0 and incoming request authentication.
  - [x] Standardized `AuthenticationConfig` for Users and Profiles.
  - [x] Priority-based authentication (Profile > User > Global).
- [ ] **Expand Test Coverage**: Increase unit and integration test coverage for all existing and new features.
- [ ] **Improve Error Handling**: Enhance error messages and provide more context for debugging.

## Long-Term Goals (6-12+ Months)

- [ ] **WASM Plugin Support**: Allow extending functionality using WebAssembly plugins for custom logic.
- [ ] **Add Support for More Service Types**: Extend the server to support additional protocols.
- [ ] **Implement a Web-Based UI**: Create a user interface for easier management and monitoring of the server.
- [ ] **MCP Any Config Registry**: A public registry where users can publish, subscribe to, and auto-update MCP configurations.
- [ ] **Client SDKs**: Develop official Client SDKs (Go, Python, TS).
