# 🗺️ Roadmap

This document outlines the current status and future direction of the MCP Any project.

## Current Status

### Implemented Features

The following features are fully implemented and tested:

- **Service Types**:
  - gRPC (with reflection)
  - HTTP
  - OpenAPI
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

## Short-Term Goals (Next 1-3 Months)

Our immediate focus is on implementing the features that are already defined in the configuration schema.

- [x] **Implement Caching and Rate Limiting**: Build the server-side logic to enforce the `CacheConfig` and `RateLimitConfig` policies.
- [x] **Implement Resilience Policies**: Build the server-side logic to enforce `ResilienceConfig` policies.
- [ ] **Implement Health Checking**: Build the logic for performing service health checks and routing traffic accordingly.
- [ ] **Implement Advanced Authentication**: Add support for OAuth 2.0 and incoming request authentication.
- [ ] **Expand Test Coverage**: Increase unit and integration test coverage for all existing and new features.
- [ ] **Improve Error Handling**: Enhance error messages and provide more context for debugging.

## Long-Term Goals (6-12+ Months)

- [ ] **Add Support for More Service Types**: Extend the server to support additional protocols, such as GraphQL and WebSockets.
- [ ] **Implement a Web-Based UI**: Create a user interface for easier management and monitoring of the server.
- [ ] **Official Helm Chart**: Provide an official Helm chart for easy deployment to Kubernetes.
- [ ] **Distributed Tracing**: Integrate with systems like OpenTelemetry to provide better observability.
