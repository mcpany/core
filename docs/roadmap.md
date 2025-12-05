# 🗺️ Roadmap

This document outlines the current status and future direction of the MCP Any project.

## Current Status

### Implemented Features

The following features are fully implemented and tested:

- **Service Types**:
  - gRPC (with reflection)
  - HTTP
  - OpenAPI
  - GraphQL
  - Stdio
  - MCP-to-MCP Proxy
- **Upstream Authentication**:
  - API Key
  - Bearer Token
  - OAuth 2.0
- **Dynamic Registration**: Services can be registered at runtime via the gRPC Registration API.
- **Static Registration**: Services can be registered at startup via a YAML or JSON configuration file.
- **Advanced Service Policies**:
  - Caching (`CacheConfig`)
  - Rate Limiting (`RateLimitConfig`)
  - Resilience (Circuit Breakers & Retries) (`ResilienceConfig`)
- **Deployment**:
  - Official Helm Chart

### Configured but Not Yet Implemented

The following features are defined in the configuration schema (`proto/config/v1/config.proto`) but are **not yet implemented** in the server logic:

- **Advanced Authentication**:
  - Incoming request authentication (`AuthenticationConfig`)
- **Service Health Checks**:
  - `HttpHealthCheck`
  - `GrpcHealthCheck`
  - `StdioHealthCheck`
  - `WebsocketHealthCheck`
  - `WebRTCHealthCheck`

For a complete list of all available configuration options, please see the [Configuration Reference](./reference/configuration.md).

## High Priority (Next 1-3 Months)

- [ ] **Hot Configuration Reloading**: Allow updating configuration without restarting the server to ensure zero downtime.
- [ ] **Secret Management Integration**: Support sourcing secrets from Vault, AWS Secrets Manager, and other providers directly in configuration.
- [ ] **Schema Validation**: Integrate CUE or JSON Schema to validate configuration files before loading.
- [ ] **Distributed Tracing**: Integrate with OpenTelemetry to provide end-to-end visibility into request flows across microservices.
- [ ] **Automated Documentation Generation**: Generate markdown documentation for registered tools directly from the configuration.
- [ ] **Enhanced Metrics**: Provide more granular metrics for tool usage, performance, and error rates.
- [ ] **IP Allowlisting & Security Policies**: Implement fine-grained security policies, including IP allowlisting and request validation.
- [ ] **WASM Plugin Support**: Allow extending functionality using WebAssembly plugins for custom logic.
- [ ] **Transport Protocols (NATS/Kafka)**: Add support for asynchronous communication via NATS and Kafka.
- [ ] **Client SDKs**: Develop official Client SDKs (Go, Python, TS) to interact with MCP Any programmatically.

## Ongoing Goals

- [ ] **Implement Health Checking**: Build the logic for performing service health checks and routing traffic accordingly.
- [ ] **Implement Advanced Authentication**: Add support for OAuth 2.0 and incoming request authentication.
- [ ] **Expand Test Coverage**: Increase unit and integration test coverage for all existing and new features.
- [ ] **Improve Error Handling**: Enhance error messages and provide more context for debugging.

## Long-Term Goals (6-12+ Months)

- [ ] **Add Support for More Service Types**: Extend the server to support additional protocols, such as WebSockets.
- [ ] **Implement a Web-Based UI**: Create a user interface for easier management and monitoring of the server.
