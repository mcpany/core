# MCP Any Configuration Reference

> **Disclaimer:** This document is a reference for all the configuration options available in the `proto/config/v1/config.proto` file. While these settings are defined in the configuration schema, not all of them have been fully implemented in the server logic. Please refer to the project's roadmap for the current implementation status of each feature.

This document provides a comprehensive reference for configuring the MCP Any server. The configuration is defined in the `McpAnyServerConfig` protobuf message and can be provided to the server in YAML or JSON format.

## Root Server Configuration (`McpAnyServerConfig`)

The `McpAnyServerConfig` is the top-level configuration object for the entire MCP Any server.

| Field               | Type                             | Description                                                                                                                          |
| ------------------- | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `global_settings`   | `GlobalSettings`                 | Defines server-wide operational parameters, such as the bind address and log level.                                                  |
| `upstream_services` | `repeated UpstreamServiceConfig` | A list of all configured upstream services that MCP Any will proxy to. Each service has its own specific configuration and policies. |
| `upstream_service_collections` | `repeated UpstreamServiceCollection` | A list of upstream service collections to load from remote sources. |

### `UpstreamServiceCollection`

Defines a collection of upstream services that can be loaded from a remote source.

| Field          | Type                     | Description                                                          |
| -------------- | ------------------------ | -------------------------------------------------------------------- |
| `name`         | `string`                 | The name of the collection.                                          |
| `http_url`     | `string`                 | The HTTP URL to load the collection from.                            |
| `priority`     | `int32`                  | The priority of the collection. Lower numbers have higher priority.  |
| `authentication` | `UpstreamAuthentication` | The authentication to use when fetching the collection.              |

### `GlobalSettings`

Contains server-wide operational parameters.

| Field                | Type            | Description                                                                   |
| -------------------- | --------------- | ----------------------------------------------------------------------------- |
| `mcp_listen_address` | `string`        | The address and port the server should bind to (e.g., "0.0.0.0:8080").        |
| `mcp_basepath`       | `string`        | The base path for all MCP API endpoints (e.g., "/mcp/v1").                    |
| `log_level`          | `enum`          | The logging level for the server. Can be `INFO`, `WARN`, `ERROR`, or `DEBUG`. |
| `message_bus`        | `MessageBus`    | The message bus configuration.                                                |

## Upstream Service Configuration (`UpstreamServiceConfig`)

This is the top-level configuration for a single upstream service that MCP Any will proxy.

| Field                     | Type                     | Description                                                                                   |
| ------------------------- | ------------------------ | --------------------------------------------------------------------------------------------- |
| `id`                      | `string`                 | A UUID to uniquely identify this upstream service configuration, used for bindings.           |
| `name`                    | `string`                 | A unique name for the upstream service. Used for identification, logging, and metrics.        |
| `connection_pool`         | `ConnectionPoolConfig`   | Configuration for the pool of connections to the upstream service.                            |
| `upstream_authentication` | `UpstreamAuthentication` | Authentication configuration for MCP Any to use when connecting to the upstream service.      |
| `cache`                   | `CacheConfig`            | Caching configuration to improve performance and reduce load on the upstream.                 |
| `rate_limit`              | `RateLimitConfig`        | Rate limiting to protect the upstream service from being overwhelmed.                         |
| `load_balancing_strategy` | `enum`                   | Strategy for distributing requests among multiple instances of the service.                   |
| `resilience`              | `ResilienceConfig`       | Advanced resiliency features like circuit breakers and retries to handle failures gracefully. |
| `service_config`          | `oneof`                  | The specific configuration for the type of upstream service (gRPC, HTTP, OpenAPI, etc.).      |
| `version`                 | `string`                 | The version of the upstream service, if known (e.g., "v1.2.3").                               |
| `authentication`          | `AuthenticationConfig`   | Authentication configuration for securing access to the MCP Any service (incoming requests).  |
| `disable`                 | `bool`                   | If true, this upstream service is disabled.                                                  |
| `priority`                | `int32`                  | The priority of the service. Lower numbers have higher priority.                             |

### Upstream Service Types

The `service_config` oneof field can contain one of the following service types:

- **`GrpcUpstreamService`**: For gRPC services.
- **`HttpUpstreamService`**: For generic HTTP services.
- **`OpenapiUpstreamService`**: For services defined by an OpenAPI (Swagger) specification.
- **`CommandLineUpstreamService`**: For services that communicate over standard I/O.
- **`McpUpstreamService`**: For proxying another MCP Any instance.

#### `GrpcUpstreamService`

| Field              | Type                              | Description                                                    |
| ------------------ | --------------------------------- | -------------------------------------------------------------- |
| `address`          | `string`                          | The address of the gRPC server.                                |
| `use_reflection`   | `bool`                            | If true, MCP Any will use gRPC reflection to discover services.|
| `tls_config`       | `TLSConfig`                       | TLS configuration for the connection.                          |
| `tools`            | `repeated ToolDefinition`         | Manually defined mappings from MCP tools.                      |
| `health_check`     | `GrpcHealthCheck`                 | Health check configuration.                                    |
| `proto_definitions`| `repeated ProtoDefinition`        | A list of protobuf definitions for the gRPC service.           |
| `proto_collection` | `repeated ProtoCollection`        | A collection of protobuf files to be discovered.               |
| `resources`        | `repeated ResourceDefinition`     | A list of resources served by this service.                    |
| `calls`            | `map<string, GrpcCallDefinition>`   | A map of call definitions, keyed by their unique ID.           |
| `prompts`          | `repeated PromptDefinition`       | A list of prompts served by this service.                      |

#### `HttpUpstreamService`

| Field          | Type                              | Description                                                    |
| -------------- | --------------------------------- | -------------------------------------------------------------- |
| `address`      | `string`                          | The base URL of the HTTP service.                              |
| `tools`        | `repeated ToolDefinition`         | Manually defined mappings from MCP tools.                      |
| `health_check` | `HttpHealthCheck`                 | Health check configuration.                                    |
| `tls_config`   | `TLSConfig`                       | TLS configuration for the connection.                          |
| `resources`    | `repeated ResourceDefinition`     | A list of resources served by this service.                    |
| `calls`        | `map<string, HttpCallDefinition>`   | A map of call definitions, keyed by their unique ID.           |
| `prompts`      | `repeated PromptDefinition`       | A list of prompts served by this service.                      |

#### `OpenapiUpstreamService`

| Field          | Type                                | Description                                                    |
| -------------- | ----------------------------------- | -------------------------------------------------------------- |
| `address`      | `string`                            | The base URL of the API.                                       |
| `openapi_spec` | `string`                            | The OpenAPI specification content.                             |
| `health_check` | `HttpHealthCheck`                   | Health check configuration.                                    |
| `tls_config`   | `TLSConfig`                         | TLS configuration for the connection.                          |
| `tools`        | `repeated ToolDefinition`           | Overrides for calls discovered from the spec.                  |
| `resources`    | `repeated ResourceDefinition`       | A list of resources served by this service.                    |
| `calls`        | `map<string, OpenAPICallDefinition>`  | A map of call definitions, keyed by their unique ID.           |
| `prompts`      | `repeated PromptDefinition`         | A list of prompts served by this service.                      |

#### `CommandLineUpstreamService`

| Field                   | Type                                      | Description                                                    |
| ----------------------- | ----------------------------------------- | -------------------------------------------------------------- |
| `command`               | `string`                                  | The command to execute the service.                            |
| `working_directory`     | `string`                                  | The working directory for the command.                         |
| `tools`                 | `repeated ToolDefinition`                 | Manually defined mappings from MCP tools.                      |
| `health_check`          | `CommandLineHealthCheck`                  | Health check configuration.                                    |
| `cache`                 | `CacheConfig`                             | Caching configuration.                                         |
| `container_environment` | `ContainerEnvironment`                    | Container environment to run the command in.                   |
| `timeout`               | `duration`                                | Timeout for the command execution.                             |
| `resources`             | `repeated ResourceDefinition`             | A list of resources served by this service.                    |
| `calls`                 | `map<string, CommandLineCallDefinition>`    | A map of call definitions, keyed by their unique ID.           |
| `prompts`               | `repeated PromptDefinition`               | A list of prompts served by this service.                      |

#### `McpUpstreamService`

| Field                 | Type                            | Description                                                    |
| --------------------- | ------------------------------- | -------------------------------------------------------------- |
| `connection_type`     | `oneof`                         | The connection details for the upstream MCP service.           |
| `tool_auto_discovery` | `bool`                          | If true, auto-discover and proxy all tools from upstream.      |
| `tools`               | `repeated ToolDefinition`       | Overrides for calls discovered from the service.               |
| `resources`           | `repeated ResourceDefinition`   | A list of resources served by this service.                    |
| `calls`               | `map<string, MCPCallDefinition>`  | A map of call definitions, keyed by their unique ID.           |
| `prompts`             | `repeated PromptDefinition`     | A list of prompts served by this service.                      |
- **`WebsocketUpstreamService`**: For services that communicate over Websocket.
- **`WebrtcUpstreamService`**: For services that communicate over WebRTC data channels.

#### `WebsocketUpstreamService`

| Field       | Type                                | Description                                           |
| ----------- | ----------------------------------- | ----------------------------------------------------- |
| `address`   | `string`                            | The URL of the Websocket service.                     |
| `tools`     | `repeated ToolDefinition`           | Manually defined mappings from MCP tools.             |
| `tls_config`| `TLSConfig`                         | TLS configuration for the connection.                 |
| `resources` | `repeated ResourceDefinition`       | A list of resources served by this service.           |
| `calls`     | `map<string, WebsocketCallDefinition>` | A map of call definitions, keyed by their unique ID.  |
| `prompts`   | `repeated PromptDefinition`         | A list of prompts served by this service.             |

#### `WebrtcUpstreamService`

| Field       | Type                              | Description                                           |
| ----------- | --------------------------------- | ----------------------------------------------------- |
| `address`   | `string`                          | The URL of the WebRTC signaling service.              |
| `tools`     | `repeated ToolDefinition`         | Manually defined mappings from MCP tools.             |
| `tls_config`| `TLSConfig`                       | TLS configuration for the signaling connection.       |
| `resources` | `repeated ResourceDefinition`     | A list of resources served by this service.           |
| `calls`     | `map<string, WebrtcCallDefinition>` | A map of call definitions, keyed by their unique ID.  |
| `prompts`   | `repeated PromptDefinition`       | A list of prompts served by this service.             |

### Service Policies and Advanced Configuration

MCP Any supports several advanced policies that can be applied to upstream services.

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

MCP Any can perform health checks on upstream services to ensure they are available.

#### `HttpHealthCheck`

| Field                             | Type       | Description                                                                |
| --------------------------------- | ---------- | -------------------------------------------------------------------------- |
| `url`                             | `string`   | The full URL to send the health check request to.                          |
| `expected_code`                   | `int32`    | The expected HTTP status code for a successful health check. Defaults to 200.|
| `expected_response_body_contains` | `string`   | A substring that must be present in the response body for the check to pass. |
| `interval`                        | `duration` | How often to perform the health check.                                     |
| `timeout`                         | `duration` | The timeout for each health check attempt.                                 |

#### `GrpcHealthCheck`

| Field               | Type       | Description                                                                |
| ------------------- | ---------- | -------------------------------------------------------------------------- |
| `service`           | `string`   | The gRPC service name to check (e.g., "grpc.health.v1.Health").            |
| `method`            | `string`   | The gRPC method to call.                                                   |
| `request`           | `string`   | A JSON string representing the request message.                            |
| `expected_response` | `string`   | A JSON string representing the expected response message.                  |
| `insecure`          | `bool`     | Set to true if connecting to the gRPC service without TLS.                 |
| `interval`          | `duration` | How often to perform the health check.                                     |
| `timeout`           | `duration` | The timeout for each health check attempt.                                 |

#### `CommandLineHealthCheck`

| Field                        | Type       | Description                                                                |
| ---------------------------- | ---------- | -------------------------------------------------------------------------- |
| `method`                     | `string`   | The method or command to send to the command line service.                 |
| `prompt`                     | `string`   | The input/prompt to send to the service.                                   |
| `expected_response_contains` | `string`   | A substring expected in the service's output for the check to pass.        |
| `interval`                   | `duration` | How often to perform the health check.                                     |
| `timeout`                    | `duration` | The timeout for each health check attempt.                                 |

### Authentication

MCP Any supports authentication for both incoming requests (securing access to the MCP Any service itself) and outgoing requests (authenticating with upstream services).

#### `AuthenticationConfig` (Incoming)

Configures the authentication method for incoming requests to the MCP Any server.

| Field     | Type         | Description                                     |
| --------- | ------------ | ----------------------------------------------- |
| `api_key` | `APIKeyAuth` | API key in a header or query parameter.         |
| `oauth2`  | `OAuth2Auth` | OAuth 2.0 client credentials or JWT validation. |

##### `APIKeyAuth`

| Field       | Type       | Description                                                 |
| ----------- | ---------- | ----------------------------------------------------------- |
| `param_name`| `string`   | The name of the parameter carrying the key (e.g., "X-API-Key"). |
| `in`        | `enum`     | Where the API key is located (`HEADER` or `QUERY`).         |
| `key_value` | `string`   | The actual API key value.                                   |

##### `OAuth2Auth`

| Field               | Type     | Description                                               |
| ------------------- | -------- | --------------------------------------------------------- |
| `token_url`         | `string` | The URL to the OAuth 2.0 token endpoint.                  |
| `authorization_url` | `string` | The URL to the OAuth 2.0 authorization endpoint.          |
| `scopes`            | `string` | A space-delimited list of scopes.                         |
| `issuer_url`        | `string` | The URL of the JWT issuer for token validation.           |
| `audience`          | `string` | The audience for JWT token validation.                    |

#### `UpstreamAuthentication` (Outgoing)

Configures the authentication method for MCP Any to use when connecting to an upstream service.

| Field          | Type                        | Description                                     |
| -------------- | --------------------------- | ----------------------------------------------- |
| `api_key`      | `UpstreamAPIKeyAuth`        | API key sent in a header.                       |
| `bearer_token` | `UpstreamBearerTokenAuth`   | Bearer token in the `Authorization` header.     |
| `basic_auth`   | `UpstreamBasicAuth`         | Basic authentication with a username and password.|
| `oauth2`       | `UpstreamOAuth2Auth`        | OAuth 2.0 client credentials flow.              |

##### `UpstreamAPIKeyAuth`

| Field         | Type          | Description                                           |
| ------------- | ------------- | ----------------------------------------------------- |
| `header_name` | `string`      | The name of the header to carry the API key.          |
| `api_key`     | `SecretValue` | The API key value, managed as a secret.               |

##### `UpstreamBearerTokenAuth`

| Field   | Type          | Description                                           |
| ------- | ------------- | ----------------------------------------------------- |
| `token` | `SecretValue` | The bearer token, managed as a secret.                |

##### `UpstreamBasicAuth`

| Field      | Type          | Description                                           |
| ---------- | ------------- | ----------------------------------------------------- |
| `username` | `string`      | The username for basic authentication.                |
| `password` | `SecretValue` | The password, managed as a secret.                    |

##### `UpstreamOAuth2Auth`

| Field           | Type          | Description                                           |
| --------------- | ------------- | ----------------------------------------------------- |
| `token_url`     | `string`      | The URL to the OAuth 2.0 token endpoint.              |
| `client_id`     | `SecretValue` | The client ID for the OAuth 2.0 flow.                 |
| `client_secret` | `SecretValue` | The client secret for the OAuth 2.0 flow.             |
| `scopes`        | `string`      | A space-delimited list of scopes.                     |

##### `SecretValue`

The `SecretValue` message provides a secure way to manage sensitive information like API keys, passwords, and tokens. It can be defined in one of the following ways:

| Field                  | Type            | Description                                                     |
| ---------------------- | --------------- | --------------------------------------------------------------- |
| `plain_text`           | `string`        | The secret value as a plain text string. **(Not Recommended)**    |
| `environment_variable` | `string`        | The name of an environment variable containing the secret.      |
| `file_path`            | `string`        | The path to a file containing the secret.                       |
| `remote_content`       | `RemoteContent` | Fetches the secret from a remote URL.                           |

### TLS Configuration (`TLSConfig`)

Defines TLS settings for connecting to an upstream service.

| Field                  | Type     | Description                                                                               |
| ---------------------- | -------- | ----------------------------------------------------------------------------------------- |
| `server_name`          | `string` | The server name to use for SNI.                                                           |
| `ca_cert_path`         | `string` | Path to the CA certificate file for verifying the server's certificate.                   |
| `client_cert_path`     | `string` | Path to the client certificate file for mTLS.                                             |
| `client_key_path`      | `string` | Path to the client private key file for mTLS.                                             |
| `insecure_skip_verify` | `bool`   | If true, the client will not verify the server's certificate chain. **Use with caution.** |
