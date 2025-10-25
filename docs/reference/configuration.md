# MCP-X Configuration Reference

> **Disclaimer:** This document is a reference for all the configuration options available in the `proto/config/v1/config.proto` file. While these settings are defined in the configuration schema, not all of them have been fully implemented in the server logic. Please refer to the project's roadmap for the current implementation status of each feature.

This document provides a comprehensive reference for configuring the MCP-X server. The configuration is defined in the `McpxServerConfig` protobuf message and can be provided to the server in YAML or JSON format.

## Root Server Configuration (`McpxServerConfig`)

The `McpxServerConfig` is the top-level configuration object for the entire MCP-X server.

| Field               | Type                             | Description                                                                                                                                   |
| ------------------- | -------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `global_settings`   | `GlobalSettings`                 | Defines server-wide operational parameters, such as the bind address and log level.                                                           |
| `upstream_services` | `repeated UpstreamServiceConfig` | A list of all configured upstream services that MCP-X will proxy to. Each service has its own specific configuration and policies.            |

### `GlobalSettings`

Contains server-wide operational parameters.

| Field            | Type     | Description                                                                   |
| ---------------- | -------- | ----------------------------------------------------------------------------- |
| `bind_address`   | `string` | The address and port the server should bind to (e.g., "0.0.0.0:8080").        |
| `mcp_basepath`   | `string` | The base path for all MCP API endpoints (e.g., "/mcp/v1").                    |
| `log_level`      | `enum`   | The logging level for the server. Can be `INFO`, `WARN`, `ERROR`, or `DEBUG`. |
| `protoc_version` | `string` | The version of `protoc` to use for parsing `.proto` files.                    |

## Upstream Service Configuration (`UpstreamServiceConfig`)

This is the top-level configuration for a single upstream service that MCP-X will proxy.

| Field                     | Type                     | Description                                                                                   |
| ------------------------- | ------------------------ | --------------------------------------------------------------------------------------------- |
| `id`                      | `string`                 | A UUID to uniquely identify this upstream service configuration, used for bindings.           |
| `name`                    | `string`                 | A unique name for the upstream service. Used for identification, logging, and metrics.        |
| `connection_pool`         | `ConnectionPoolConfig`   | Configuration for the pool of connections to the upstream service.                            |
| `upstream_authentication` | `UpstreamAuthentication` | Authentication configuration for MCP-X to use when connecting to the upstream service.        |
| `cache`                   | `CacheConfig`            | Caching configuration to improve performance and reduce load on the upstream.                 |
| `rate_limit`              | `RateLimitConfig`        | Rate limiting to protect the upstream service from being overwhelmed.                         |
| `load_balancing_strategy` | `enum`                   | Strategy for distributing requests among multiple instances of the service.                   |
| `resilience`              | `ResilienceConfig`       | Advanced resiliency features like circuit breakers and retries to handle failures gracefully. |
| `service_config`          | `oneof`                  | The specific configuration for the type of upstream service (gRPC, HTTP, OpenAPI, etc.).      |
| `version`                 | `string`                 | The version of the upstream service, if known (e.g., "v1.2.3").                               |
| `authentication`          | `AuthenticationConfig`   | Authentication configuration for securing access to the MCP-X service (incoming requests).    |

### Upstream Service Types

The `service_config` oneof field can contain one of the following service types:

- **`GrpcUpstreamService`**: For gRPC services.
- **`HttpUpstreamService`**: For generic HTTP services.
- **`OpenapiUpstreamService`**: For services defined by an OpenAPI (Swagger) specification.
- **`CommandLineUpstreamService`**: For services that communicate over standard I/O.
- **`McpUpstreamService`**: For proxying another MCP-X instance.

### Service Policies and Advanced Configuration

MCP-X supports several advanced policies that can be applied to upstream services.

#### `ConnectionPoolConfig`

| Field                  | Type       | Description                                                                |
| ---------------------- | ---------- | -------------------------------------------------------------------------- |
| `max_connections`      | `int32`    | The maximum number of simultaneous connections to the upstream service.    |
| `max_idle_connections` | `int32`    | The maximum number of idle connections to keep in the pool.                |
| `idle_timeout`         | `duration` | The duration a connection can remain idle in the pool before being closed. |

#### `RateLimitConfig`

| Field                 | Type     | Description                                                  |
| --------------------- | -------- | ------------------------------------------------------------ |
| `is_enabled`          | `bool`   | Whether rate limiting is enabled.                            |
| `requests_per_second` | `double` | The maximum number of requests allowed per second.           |
| `burst`               | `int64`  | The number of requests that can be allowed in a short burst. |

#### `CacheConfig`

| Field        | Type       | Description                                                   |
| ------------ | ---------- | ------------------------------------------------------------- |
| `is_enabled` | `bool`     | Whether caching is enabled.                                   |
| `ttl`        | `duration` | The duration for which a cached response is considered valid. |
| `strategy`   | `string`   | The caching strategy to use (e.g., "lru", "lfu").             |

#### `ResilienceConfig`

Contains configurations for circuit breakers and retries.

- **`circuit_breaker` (`CircuitBreakerConfig`)**:
  - `failure_rate_threshold`: The failure rate threshold to open the circuit (e.g., 0.5 for 50%).
  - `consecutive_failures`: The number of consecutive failures required to open the circuit.
  - `open_duration`: The duration the circuit remains open before transitioning to half-open.
  - `half_open_requests`: The number of requests to allow in the half-open state to test for recovery.
- **`retry_policy` (`RetryConfig`)**:
  - `number_of_retries`: The number of times to retry a failed request.
  - `base_backoff`: The base duration for the backoff between retries.
  - `max_backoff`: The maximum duration for the backoff.

### Health Checks

MCP-X can perform health checks on upstream services to ensure they are available.

- **`HttpHealthCheck`**: For HTTP-based services.
- **`GrpcHealthCheck`**: For gRPC-based services.
- **`StdioHealthCheck`**: For stdio-based services.

### Authentication

MCP-X supports authentication for both incoming requests and requests to upstream services.

#### `AuthenticationConfig` (Incoming)

Secures access to the MCP-X service itself.

- **`api_key` (`APIKeyAuth`)**: API key in a header or query parameter.
- **`oauth2` (`OAuth2Auth`)**: OAuth 2.0 client credentials flow.

#### `UpstreamAuthentication` (Outgoing)

Authenticates MCP-X with the upstream service.

- **`api_key` (`UpstreamAPIKeyAuth`)**: API key sent in a header.
- **`bearer_token` (`UpstreamBearerTokenAuth`)**: Bearer token in the `Authorization` header.
- **`basic_auth` (`UpstreamBasicAuth`)**: Basic authentication with a username and password.

### TLS Configuration (`TLSConfig`)

Defines TLS settings for connecting to an upstream service.

| Field                  | Type     | Description                                                                               |
| ---------------------- | -------- | ----------------------------------------------------------------------------------------- |
| `server_name`          | `string` | The server name to use for SNI.                                                           |
| `ca_cert_path`         | `string` | Path to the CA certificate file for verifying the server's certificate.                   |
| `client_cert_path`     | `string` | Path to the client certificate file for mTLS.                                             |
| `client_key_path`      | `string` | Path to the client private key file for mTLS.                                             |
| `insecure_skip_verify` | `bool`   | If true, the client will not verify the server's certificate chain. **Use with caution.** |
