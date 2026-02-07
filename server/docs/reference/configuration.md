# MCP Any Configuration Reference

> **Disclaimer:** This document is a reference for all the configuration options available in the `proto/config/v1/config.proto` file. While these settings are defined in the configuration schema, not all of them have been fully implemented in the server logic. Please refer to the project's roadmap for the current implementation status of each feature.

This document provides a comprehensive reference for configuring the MCP Any server. The configuration is defined in the `McpAnyServerConfig` protobuf message and can be provided to the server in YAML or JSON format.

## Using Environment Variables

MCP Any supports the use of environment variables within configuration files to enhance security and portability. You can reference environment variables using the `${VAR_NAME}` syntax. Default values are also supported using the `${VAR_NAME:default_value}` syntax.

### Example

```yaml
upstream_services:
  - name: "my-api"
    http_service:
      address: "https://api.example.com"
    upstream_authentication:
      api_key:
        header_name: "X-API-Key"
        api_key:
          plain_text: "${API_KEY:my-secret-key}"
```

In this example, the `api_key` will be set to the value of the `API_KEY` environment variable. If `API_KEY` is not set, it will default to `my-secret-key`.

## Root Server Configuration (`McpAnyServerConfig`)

The `McpAnyServerConfig` is the top-level configuration object for the entire MCP Any server.

| Field                          | Type                                 | Description                                                                                                                          |
| ------------------------------ | ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------ |
| `global_settings`              | `GlobalSettings`                     | Defines server-wide operational parameters, such as the bind address and log level.                                                  |
| `upstream_services`            | `repeated UpstreamServiceConfig`     | A list of all configured upstream services that MCP Any will proxy to. Each service has its own specific configuration and policies. |
| `upstream_service_collections` | `repeated Collection` | A list of upstream service collections to load from remote sources.                                                                  |

### Use Case and Example

A top-level configuration for an MCP Any server that connects to a local gRPC service and a remote HTTP service.

```yaml
global_settings:
  mcp_listen_address: "0.0.0.0:8080"
  log_level: "LOG_LEVEL_INFO"
upstream_services:
  - name: "user-service"
    grpc_service:
      address: "localhost:50051"
      use_reflection: true
  - name: "weather-api"
    http_service:
      address: "https://api.weather.com"
```

### `Collection`

Defines a collection of upstream services that can be loaded from a remote source.

| Field            | Type                     | Description                                                         |
| ---------------- | ------------------------ | ------------------------------------------------------------------- |
| `name`           | `string`                 | The name of the collection.                                         |
| `http_url`       | `string`                 | The HTTP URL to load the collection from.                           |
| `priority`       | `int32`                  | The priority of the collection. Lower numbers have higher priority. |
| `authentication` | `UpstreamAuthentication` | The authentication to use when fetching the collection.             |

### Use Case and Example

Dynamically load a collection of upstream services from a remote URL. This is useful for managing a large number of services or for updating service configurations without restarting the MCP Any server.

```yaml
upstream_service_collections:
  - name: "shared-services"
    http_url: "https://config.example.com/services.yaml"
    priority: 1
    authentication:
      bearer_token:
        token:
          environment_variable: "CONFIG_SERVER_AUTH_TOKEN"
```

### `GlobalSettings`

Contains server-wide operational parameters.

| Field                | Type         | Description                                                                   |
| -------------------- | ------------ | ----------------------------------------------------------------------------- |
| `mcp_listen_address` | `string`     | The address and port the server should bind to (e.g., "0.0.0.0:8080").        |
| `log_level`          | `enum`       | The logging level for the server. Can be `INFO`, `WARN`, `ERROR`, or `DEBUG`. |
| `log_format`         | `enum`       | The logging format. Can be `text` or `json`.                                  |
| `message_bus`        | `MessageBus` | The message bus configuration.                                                |
| `api_key`            | `string`     | The API key for securing the MCP server.                                      |
| `audit`              | `AuditConfig`| Audit logging configuration.                                                  |
| `db_path`            | `string`     | The path to the SQLite database file.                                         |
| `db_dsn`             | `string`     | The database connection string (DSN).                                         |
| `db_driver`          | `string`     | The database driver (sqlite, postgres).                                       |
| `github_api_url`     | `string`     | GitHub API URL for self-updates (optional).                                   |
| `use_sudo_for_docker`| `bool`       | Whether to use sudo for Docker commands.                                      |
| `dlp`                | `DLPConfig`  | DLP configuration.                                                            |
| `gc_settings`        | `GCSettings` | Garbage Collection configuration.                                             |
| `oidc`               | `OIDCConfig` | OIDC Configuration.                                                           |
| `rate_limit`         | `RateLimitConfig` | Rate limiting configuration for the server.                              |
| `telemetry`          | `TelemetryConfig` | Telemetry configuration.                                                 |
| `profiles`           | `repeated string` | The profiles to enable.                                                  |
| `allowed_ips`        | `repeated string` | The allowed IPs to access the server.                                    |
| `profile_definitions`| `repeated ProfileDefinition` | The definitions of profiles.                                  |
| `middlewares`        | `repeated Middleware` | The list of middlewares to enable and their configuration.             |
| `allowed_file_paths` | `repeated string` | Allowed file paths for validation.                                       |
| `allowed_origins`    | `repeated string` | Allowed origins for CORS.                                                |
| `context_optimizer`  | `ContextOptimizerConfig` | Context Optimizer configuration.                                    |
| `debugger`           | `DebuggerConfig` | Debugger configuration.                                                     |
| `read_only`          | `bool`       | If true, the configuration is read-only.                                      |
| `auto_discover_local`| `bool`       | Whether to auto-discover local services (e.g. Ollama).                        |
| `alerts`             | `AlertConfig`| Alert configuration.                                                          |

### `AuditConfig`

Configuration for audit logging of tool executions.

| Field           | Type     | Description                                                          |
| --------------- | -------- | -------------------------------------------------------------------- |
| `enabled`       | `bool`   | Whether audit logging is enabled.                                    |
| `output_path`   | `string` | The file path to write audit logs to.                                |
| `log_arguments` | `bool`   | Whether to log input arguments (caution: might contain secrets).     |
| `log_results`   | `bool`   | Whether to log output results (caution: might contain sensitive data).|
| `storage_type`  | `StorageType` | The storage type to use (FILE, SQLITE, POSTGRES, WEBHOOK, SPLUNK, DATADOG). |
| `webhook_url`   | `string` | The webhook URL for STORAGE_TYPE_WEBHOOK.                            |
| `webhook_headers`| `map<string, string>` | Additional headers to send with the webhook.           |
| `splunk`        | `SplunkConfig` | Splunk configuration.                                          |
| `datadog`       | `DatadogConfig` | Datadog configuration.                                        |

#### Use Case and Example

Enable audit logging to a file.

```yaml
global_settings:
  audit:
    enabled: true
    output_path: "/var/log/mcpany/audit.log"
    log_arguments: false
    log_results: false
```

### Use Case and Example

```yaml
global_settings:
  mcp_listen_address: "0.0.0.0:8080"
  mcp_basepath: "/mcp/v1"
  log_level: "DEBUG"
```

## Upstream Service Configuration (`UpstreamServiceConfig`)

This is the top-level configuration for a single upstream service that MCP Any will proxy.

| Field                     | Type                     | Description                                                                                   |
| ------------------------- | ------------------------ | --------------------------------------------------------------------------------------------- |
| `id`                      | `string`                 | A UUID to uniquely identify this upstream service configuration, used for bindings.           |
| `name`                    | `string`                 | A unique name for the upstream service. Used for identification, logging, and metrics.        |
| `connection_pool`         | `ConnectionPoolConfig`   | Configuration for the pool of connections to the upstream service.                            |
| `upstream_auth` | `UpstreamAuthentication` | Authentication configuration for MCP Any to use when connecting to the upstream service.      |
| `cache`                   | `CacheConfig`            | Caching configuration to improve performance and reduce load on the upstream.                 |
| `rate_limit`              | `RateLimitConfig`        | Rate limiting to protect the upstream service from being overwhelmed.                         |
| `load_balancing_strategy` | `enum`                   | Strategy for distributing requests among multiple instances of the service.                   |
| `resilience`              | `ResilienceConfig`       | Advanced resiliency features like circuit breakers and retries to handle failures gracefully. |
| `service_config`          | `oneof`                  | The specific configuration for the type of upstream service (gRPC, HTTP, OpenAPI, etc.).      |
| `version`                 | `string`                 | The version of the upstream service, if known (e.g., "v1.2.3").                               |
| `authentication`          | `AuthenticationConfig`   | Authentication configuration for securing access to the MCP Any service (incoming requests).  |
| `disable`                 | `bool`                   | If true, this upstream service is disabled.                                                   |
| `priority`                | `int32`                  | The priority of the service. Lower numbers have higher priority.                              |
| `profiles`                | `repeated Profile`       | A list of profiles this service belongs to. Defaults to `[{name: "default"}]` if empty.       |

### Profiles

Profiles allow you to categorize and selectively enable services, tools, resources, and prompts based on the runtime environment (e.g., "dev", "prod", "staging"). You can start the server with specific profiles enabled using the `--profiles` command-line flag (e.g., `--profiles=dev,staging`).

If a configuration item (service, tool, etc.) has an empty `profiles` list, it is treated as belonging to the "default" profile.
Items are enabled if they share at least one profile with the enabled profiles list.

#### Use Case and Example

Enable "debug-tools" only when the "dev" profile is active.

```yaml
upstream_services:
  - name: "debug-service"
    profiles:
      - name: "dev"
    # Service specific configuration (e.g. http_service)
    http_service:
      # ...
```

To run this service, start the server with `--profiles=dev` (or `--profiles=dev,default`).

### Use Case and Example

A gRPC service with a connection pool, rate limiting, a circuit breaker, and API key authentication for the upstream.

```yaml
upstream_services:
  - name: "product-catalog-service"
    connection_pool:
      max_connections: 100
      max_idle_connections: 10
      idle_timeout: "30s"
    rate_limit:
      is_enabled: true
      requests_per_second: 1000
      burst: 100
    resilience:
      circuit_breaker:
        failure_rate_threshold: 0.5
        open_duration: "5s"
    upstream_auth:
      api_key:
        header_name: "X-API-Key"
        api_key:
          environment_variable: "PRODUCT_CATALOG_API_KEY"
    grpc_service:
      address: "grpc.product-catalog.svc.cluster.local:50051"
      use_reflection: true
```

### Upstream Service Types

The `service_config` oneof field can contain one of the following service types:

- **`GrpcUpstreamService`**: For gRPC services.
- **`HttpUpstreamService`**: For generic HTTP services.
- **`OpenapiUpstreamService`**: For services defined by an OpenAPI (Swagger) specification.
- **`CommandLineUpstreamService`**: For services that communicate over standard I/O.
- **`McpUpstreamService`**: For proxying another MCP Any instance.
- **`GraphqlUpstreamService`**: For GraphQL services.
- **`WebsocketUpstreamService`**: For services that communicate over Websocket.
- **`WebrtcUpstreamService`**: For services that communicate over WebRTC data channels.

#### `GrpcUpstreamService`

| Field               | Type                              | Description                                                     |
| ------------------- | --------------------------------- | --------------------------------------------------------------- |
| `address`           | `string`                          | The address of the gRPC server.                                 |
| `use_reflection`    | `bool`                            | If true, MCP Any will use gRPC reflection to discover services. |
| `tls_config`        | `TLSConfig`                       | TLS configuration for the connection.                           |
| `tools`             | `repeated ToolDefinition`         | Manually defined mappings from MCP tools.                       |
| `health_check`      | `GrpcHealthCheck`                 | Health check configuration.                                     |
| `proto_definitions` | `repeated ProtoDefinition`        | A list of protobuf definitions for the gRPC service.            |
| `proto_collection`  | `repeated ProtoCollection`        | A collection of protobuf files to be discovered.                |
| `resources`         | `repeated ResourceDefinition`     | A list of resources served by this service.                     |
| `calls`             | `map<string, GrpcCallDefinition>` | A map of call definitions, keyed by their unique ID.            |
| `prompts`           | `repeated PromptDefinition`       | A list of prompts served by this service.                       |

##### Use Case and Example

Expose a gRPC service that requires TLS and has its protobuf definitions stored locally.

```yaml
grpc_service:
  address: "user-service.internal:50051"
  tls_config:
    ca_cert_path: "/certs/ca.pem"
  proto_definitions:
    - proto_file:
        file_name: "user.proto"
        file_path: "/proto/user.proto"
```

#### `HttpUpstreamService`

| Field          | Type                              | Description                                          |
| -------------- | --------------------------------- | ---------------------------------------------------- |
| `address`      | `string`                          | The base URL of the HTTP service.                    |
| `tools`        | `repeated ToolDefinition`         | Manually defined mappings from MCP tools.            |
| `health_check` | `HttpHealthCheck`                 | Health check configuration.                          |
| `tls_config`   | `TLSConfig`                       | TLS configuration for the connection.                |
| `resources`    | `repeated ResourceDefinition`     | A list of resources served by this service.          |
| `calls`        | `map<string, HttpCallDefinition>` | A map of call definitions, keyed by their unique ID. |
| `prompts`      | `repeated PromptDefinition`       | A list of prompts served by this service.            |

##### Use Case and Example

Proxy an external HTTP API and add a health check.

```yaml
http_service:
  address: "https://api.example.com"
  health_check:
    url: "https://api.example.com/health"
    expected_code: 200
    interval: "10s"
```

##### `HttpCallDefinition`

Defines how an MCP tool entry maps to a specific HTTP call.

| Field | Type | Description |
| :--- | :--- | :--- |
| `endpoint_path` | `string` | The API path relative to the service address (e.g., `/users`). |
| `method` | `string` | The HTTP method (GET, POST, etc.). |
| `timeout` | `duration` | Timeout for this specific call. |
| `cache` | `CacheConfig` | Call-level cache configuration (overrides service default). |
| `retry_policy` | `RetryConfig` | Call-level retry policy. |

#### `OutputTransformer`

Defines how to parse and transform the upstream service's output.

| Field | Type | Description |
| :--- | :--- | :--- |
| `format` | `enum` | The format of the output (`JSON`, `XML`, `TEXT`, `RAW_BYTES`, `JQ`). |
| `extraction_rules` | `map<string, string>` | Extraction rules for JSON/XML/TEXT. |
| `template` | `string` | Optional Go template to render the extracted/transformed data. |
| `jq_query` | `string` | JQ query to transform the output (when format is `JQ`). |

##### Use Case and Example: JQ Transformation

Transform a complex JSON response using JQ.

```yaml
output_transformer:
  format: "JQ"
  jq_query: ".users[] | select(.active) | .name"
```

#### `OpenapiUpstreamService`

| Field          | Type                                 | Description                                          |
| -------------- | ------------------------------------ | ---------------------------------------------------- |
| `address`      | `string`                             | The base URL of the API.                             |
| `openapi_spec` | `string`                             | The OpenAPI specification content.                   |
| `health_check` | `HttpHealthCheck`                    | Health check configuration.                          |
| `tls_config`   | `TLSConfig`                          | TLS configuration for the connection.                |
| `tools`        | `repeated ToolDefinition`            | Overrides for calls discovered from the spec.        |
| `resources`    | `repeated ResourceDefinition`        | A list of resources served by this service.          |
| `calls`        | `map<string, OpenAPICallDefinition>` | A map of call definitions, keyed by their unique ID. |
| `prompts`      | `repeated PromptDefinition`          | A list of prompts served by this service.            |

##### Use Case and Example

Automatically discover and expose an OpenAPI-defined service.

```yaml
openapi_service:
  address: "https://petstore.swagger.io/v2"
  openapi_spec: |
    swagger: "2.0"
    info:
      version: "1.0.0"
      title: "Swagger Petstore"
    # ... rest of the spec
```

#### `CommandLineUpstreamService`

| Field                    | Type                                     | Description                                           |
| ------------------------ | ---------------------------------------- | ----------------------------------------------------- |
| `command`                | `string`                                 | The command to execute the service.                   |
| `working_directory`      | `string`                                 | The working directory for the command.                |
| `tools`                  | `repeated ToolDefinition`                | Manually defined mappings from MCP tools.             |
| `health_check`           | `CommandLineHealthCheck`                 | Health check configuration.                           |
| `cache`                  | `CacheConfig`                            | Caching configuration.                                |
| `container_environment`  | `ContainerEnvironment`                   | Container environment to run the command in.          |
| `timeout`                | `duration`                               | Timeout for the command execution.                    |
| `communication_protocol` | `enum`                                   | Protocol for communicating with the command (`JSON`). |
| `local`                  | `bool`                                   | If true, execute locally instead of in a container.   |
| `resources`              | `repeated ResourceDefinition`            | A list of resources served by this service.           |
| `calls`                  | `map<string, CommandLineCallDefinition>` | A map of call definitions, keyed by their unique ID.  |
| `prompts`                | `repeated PromptDefinition`              | A list of prompts served by this service.             |

##### Use Case and Example

Wrap a command-line tool, run it in a container, and cache the results.

```yaml
command_line_service:
  command: "python /scripts/process.py"
  working_directory: "/data"
  container_environment:
    image: "python:3.9-slim"
    volumes:
      "/local/data": "/data"
  cache:
    is_enabled: true
    ttl: "1h"
```

##### `CommandLineCallDefinition`

Defines argument mapping for a command-line tool.

| Field | Type | Description |
| :--- | :--- | :--- |
| `args` | `repeated string` | Arguments to pass to the command. Supports templates (e.g. `{{arg}}`). |
| `timeout` | `duration` | Timeout for this specific execution. |
| `cache` | `CacheConfig` | Call-level cache configuration (overrides service default). |

#### `McpUpstreamService`

| Field                 | Type                             | Description                                               |
| --------------------- | -------------------------------- | --------------------------------------------------------- |
| `connection_type`     | `oneof`                          | The connection details for the upstream MCP service.      |
| `tool_auto_discovery` | `bool`                           | If true, auto-discover and proxy all tools from upstream. |
| `tools`               | `repeated ToolDefinition`        | Overrides for calls discovered from the service.          |
| `resources`           | `repeated ResourceDefinition`    | A list of resources served by this service.               |
| `calls`               | `map<string, MCPCallDefinition>` | A map of call definitions, keyed by their unique ID.      |
| `prompts`             | `repeated PromptDefinition`      | A list of prompts served by this service.                 |

##### Use Case and Example

Proxy another MCP Any instance and automatically discover its tools.

```yaml
mcp_service:
  http_connection:
    http_address: "mcp-internal.example.com:8080"
  tool_auto_discovery: true
```

#### Proxying CLI-based MCP Servers (Python/Node)

You can also proxy MCP servers that run as a local command (stdio) or in a container. This is useful for using the growing ecosystem of MCP servers.

##### Example: Python MCP Server

```yaml
mcp_service:
  stdio_connection:
    command: "python"
    args:
      - "main.py"
    working_directory: "/path/to/server"
    env:
      MY_ENV:
        plain_text: "value"
  tool_auto_discovery: true
```

##### Example: NPX MCP Server (Puppeteer)

This example runs the [`@modelcontextprotocol/server-puppeteer`](https://github.com/modelcontextprotocol/servers) using `npx`.

```yaml
mcp_service:
  stdio_connection:
    command: "npx"
    args:
      - "-y"
      - "@modelcontextprotocol/server-puppeteer"
  tool_auto_discovery: true
```

##### Verification with Gemini CLI

To verify these configurations, you can use the `@google/gemini-cli`.

1.  **Start your MCP Any server** with the config.
2.  **Add the server** to Gemini CLI:
    ```bash
    npx -y @google/gemini-cli mcp add --transport http mcpany-server http://localhost:8080/mcp/v1
    ```
    *(Adjust the URL if your server listens on a different port/path)*

3.  **Interact**:
    ```bash
    npx -y @google/gemini-cli -p "Use the puppeteer tool to take a screenshot of google.com"
    ```

#### `GraphqlUpstreamService`

| Field          | Type                                 | Description                                |
| -------------- | ------------------------------------ | ------------------------------------------ |
| `address`      | `string`                             | The endpoint URL of the GraphQL service.   |
| `calls`        | `map<string, GraphqlCallDefinition>` | A map of call definitions to expose tools. |
| `prompts`      | `repeated PromptDefinition`          | A list of prompts served by this service.  |
| `health_check` | `WebsocketHealthCheck`               | Health check configuration.                |

##### Use Case and Example

Register a GraphQL service and customize the selection set for a query.

```yaml
upstream_services:
  - name: "my-graphql-service"
    graphql_service:
      address: "http://localhost:8080/graphql"
      calls:
        user:
          selection_set: "{ id name }"
    upstream_auth:
      api_key:
        header_name: "X-API-Key"
        api_key:
          plain_text: "my-secret-key"
```

- **`WebsocketUpstreamService`**: For services that communicate over Websocket.
- **`WebrtcUpstreamService`**: For services that communicate over WebRTC data channels.

#### `WebsocketUpstreamService`

| Field        | Type                                   | Description                                          |
| ------------ | -------------------------------------- | ---------------------------------------------------- |
| `address`    | `string`                               | The URL of the Websocket service.                    |
| `tools`      | `repeated ToolDefinition`              | Manually defined mappings from MCP tools.            |
| `tls_config` | `TLSConfig`                            | TLS configuration for the connection.                |
| `resources`  | `repeated ResourceDefinition`          | A list of resources served by this service.          |
| `calls`      | `map<string, WebsocketCallDefinition>` | A map of call definitions, keyed by their unique ID. |
| `prompts`    | `repeated PromptDefinition`            | A list of prompts served by this service.            |

##### Use Case and Example

Connect to a real-time data streaming service over a secure Websocket.

```yaml
websocket_service:
  address: "wss://streaming.example.com/data"
  tls_config:
    server_name: "streaming.example.com"
```

#### `WebrtcUpstreamService`

| Field        | Type                                | Description                                          |
| ------------ | ----------------------------------- | ---------------------------------------------------- |
| `address`    | `string`                            | The URL of the WebRTC signaling service.             |
| `tools`      | `repeated ToolDefinition`           | Manually defined mappings from MCP tools.            |
| `tls_config` | `TLSConfig`                         | TLS configuration for the signaling connection.      |
| `resources`  | `repeated ResourceDefinition`       | A list of resources served by this service.          |
| `calls`      | `map<string, WebrtcCallDefinition>` | A map of call definitions, keyed by their unique ID. |
| `prompts`    | `repeated PromptDefinition`         | A list of prompts served by this service.            |

##### Use Case and Example

Expose a WebRTC service for real-time communication, connecting to its signaling server.

```yaml
webrtc_service:
  address: "https://signaling.example.com"
```

### Service Policies and Advanced Configuration

MCP Any supports several advanced policies that can be applied to upstream services.

#### `ConnectionPoolConfig`

| Field                  | Type       | Description                                                                |
| ---------------------- | ---------- | -------------------------------------------------------------------------- |
| `max_connections`      | `int32`    | The maximum number of simultaneous connections to the upstream service.    |
| `max_idle_connections` | `int32`    | The maximum number of idle connections to keep in the pool.                |
| `idle_timeout`         | `duration` | The duration a connection can remain idle in the pool before being closed. |

##### Use Case and Example

Manage connections to a high-traffic database proxy.

```yaml
connection_pool:
  max_connections: 200
  max_idle_connections: 20
  idle_timeout: "60s"
```

#### `RateLimitConfig`

| Field                 | Type     | Description                                                  |
| --------------------- | -------- | ------------------------------------------------------------ |
| `is_enabled`          | `bool`   | Whether rate limiting is enabled.                            |
| `requests_per_second` | `double` | The maximum number of requests allowed per second.           |
| `burst`               | `int64`  | The number of requests that can be allowed in a short burst. |

##### Use Case and Example

Protect a public API from excessive use.

```yaml
rate_limit:
  is_enabled: true
  requests_per_second: 100
  burst: 20
```

#### `CacheConfig`

| Field        | Type       | Description                                                   |
| ------------ | ---------- | ------------------------------------------------------------- |
| `is_enabled` | `bool`     | Whether caching is enabled.                                   |
| `ttl`        | `duration` | The duration for which a cached response is considered valid. |

**Priority Note:** When defined at the call level (e.g., inside `HttpCallDefinition.cache`), this configuration **overrides** the service-level cache settings.

##### Use Case and Example

Cache responses from a slow, data-intensive service.

```yaml
cache:
  is_enabled: true
  ttl: "5m"
```

#### `ResilienceConfig`

Contains configurations for circuit breakers and retries.

##### Use Case and Example

Improve the reliability of a connection to a occasionally unstable upstream service.

```yaml
resilience:
  circuit_breaker:
    failure_rate_threshold: 0.5
    consecutive_failures: 3
    open_duration: "10s"
    half_open_requests: 5
  retry_policy:
    number_of_retries: 3
    base_backoff: "100ms"
    max_backoff: "30s"
```

- **`circuit_breaker` (`CircuitBreakerConfig`)**:
  - `failure_rate_threshold`: The failure rate threshold to open the circuit (e.g., 0.5 for 50%).
  - `consecutive_failures`: The number of consecutive failures required to open the circuit.
  - `open_duration`: The duration the circuit remains open before transitioning to half-open.
  - `half_open_requests`: The number of requests to allow in the half-open state to test for recovery.
- **`retry_policy` (`RetryConfig`)**:
  - `number_of_retries`: The number of times to retry a failed request.
  - `base_backoff`: The base duration for the backoff between retries.
  - `max_backoff`: The maximum duration for the backoff.
  - `max_elapsed_time`: The maximum total time to spend retrying.

#### `Call Policy`

Call Policies allow you to fine-tune which calls are allowed or denied based on the request's properties. These policies are evaluated before any request is forwarded to the upstream service.

| Field             | Type                    | Description                                |
| ----------------- | ----------------------- | ------------------------------------------ |
| `default_action`  | `enum`                  | The default action if no rules match (`ALLOW` or `DENY`). |
| `rules`           | `repeated CallPolicyRule` | A list of rules to apply. The first matching rule determines the action. |

##### `CallPolicyRule`

| Field            | Type     | Description                                                         |
| ---------------- | -------- | ------------------------------------------------------------------- |
| `action`         | `enum`   | The action to take if the rule matches (`ALLOW` or `DENY`).         |
| `name_regex`     | `string` | Regex to match the tool or call name. Empty means match all.        |
| `argument_regex` | `string` | Regex to match request arguments (JSON stringified). Empty means match all. |
| `url_regex`      | `string` | Regex to match endpoint path or URL.                                |
| `call_id_regex`  | `string` | Regex to match the call ID. Empty means match all.                  |

##### Use Case and Example

Deny all calls to the "delete_user" tool, and allow everything else.

```yaml
call_policies:
  - default_action: ALLOW
    rules:
      - action: DENY
        name_regex: "^delete_user$"
```

#### `Hooks` (Pre-Call and Post-Call)

Hooks allow you to execute custom logic before (`pre_call_hooks`) or after (`post_call_hooks`) a tool is executed. This is useful for validation, auditing, transformation, or enforcing complex policies.

| Field   | Type     | Description             |
| ------- | -------- | ----------------------- |
| `name`  | `string` | A name for the hook.    |
| `webhook`| `WebhookConfig` | Configuration for an external webhook. |
| `call_policy` | `CallPolicy` | A call policy to enforce (only for `pre_call_hooks`). |

##### `WebhookConfig`

| Field            | Type       | Description                                      |
| ---------------- | ---------- | ------------------------------------------------ |
| `url`            | `string`   | The URL of the webhook service.                  |
| `timeout`        | `duration` | The timeout for the webhook request.             |
| `webhook_secret` | `string`   | A secret shared with the webhook for HMAC validation (optional). |

##### Use Case and Example

Validate tool arguments using an external webhook before execution.

```yaml
pre_call_hooks:
  - name: "argument-validator"
    webhook:
      url: "http://policy-engine.internal/validate"
      timeout: "500ms"
```

#### `ContainerEnvironment`

| Field     | Type                  | Description                                                                           |
| --------- | --------------------- | ------------------------------------------------------------------------------------- |
| `name`    | `string`              | The name of the container.                                                            |
| `image`   | `string`              | The image to use for the container.                                                   |
| `volumes` | `map<string, string>` | The volumes to mount into the container, with destination as key and source as value. |

### Health Checks

MCP Any can perform health checks on upstream services to ensure they are available.

#### `HttpHealthCheck`

| Field                             | Type       | Description                                                                   |
| --------------------------------- | ---------- | ----------------------------------------------------------------------------- |
| `url`                             | `string`   | The full URL to send the health check request to.                             |
| `expected_code`                   | `int32`    | The expected HTTP status code for a successful health check. Defaults to 200. |
| `expected_response_body_contains` | `string`   | A substring that must be present in the response body for the check to pass.  |
| `interval`                        | `duration` | How often to perform the health check.                                        |
| `timeout`                         | `duration` | The timeout for each health check attempt.                                    |

#### `GrpcHealthCheck`

| Field               | Type       | Description                                                     |
| ------------------- | ---------- | --------------------------------------------------------------- |
| `service`           | `string`   | The gRPC service name to check (e.g., "grpc.health.v1.Health"). |
| `method`            | `string`   | The gRPC method to call.                                        |
| `request`           | `string`   | A JSON string representing the request message.                 |
| `expected_response` | `string`   | A JSON string representing the expected response message.       |
| `insecure`          | `bool`     | Set to true if connecting to the gRPC service without TLS.      |
| `interval`          | `duration` | How often to perform the health check.                          |
| `timeout`           | `duration` | The timeout for each health check attempt.                      |

#### `CommandLineHealthCheck`

| Field                        | Type       | Description                                                         |
| ---------------------------- | ---------- | ------------------------------------------------------------------- |
| `method`                     | `string`   | The method or command to send to the command line service.          |
| `prompt`                     | `string`   | The input/prompt to send to the service.                            |
| `expected_response_contains` | `string`   | A substring expected in the service's output for the check to pass. |
| `interval`                   | `duration` | How often to perform the health check.                              |
| `timeout`                    | `duration` | The timeout for each health check attempt.                          |

#### `WebsocketHealthCheck`

| Field                        | Type       | Description                                                           |
| ---------------------------- | ---------- | --------------------------------------------------------------------- |
| `message`                    | `string`   | The message to send to the websocket service for the health check.    |
| `expected_response_contains` | `string`   | A substring expected in the service's response for the check to pass. |
| `interval`                   | `duration` | How often to perform the health check.                                |
| `timeout`                    | `duration` | The timeout for each health check attempt.                            |

### Authentication

MCP Any supports authentication for both incoming requests (securing access to the MCP Any service itself) and outgoing requests (authenticating with upstream services).

#### `AuthenticationConfig` (Incoming)

Configures the authentication method for incoming requests to the MCP Any server.

| Field     | Type         | Description                                     |
| --------- | ------------ | ----------------------------------------------- |
| `api_key` | `APIKeyAuth` | API key in a header or query parameter.         |
| `oauth2`  | `OAuth2Auth` | OAuth 2.0 client credentials or JWT validation. |

##### Use Case and Example

Secure the MCP Any server with API key authentication.

```yaml
authentication:
  api_key:
    param_name: "X-Mcp-Api-Key"
    in: "HEADER"
    key_value: "my-secret-key"
```

##### `APIKeyAuth`

| Field        | Type     | Description                                                     |
| ------------ | -------- | --------------------------------------------------------------- |
| `param_name` | `string` | The name of the parameter carrying the key (e.g., "X-API-Key"). |
| `in`         | `enum`   | Where the API key is located (`HEADER` or `QUERY`).             |
| `key_value`  | `string` | The actual API key value.                                       |

##### `OAuth2Auth`

| Field               | Type     | Description                                      |
| ------------------- | -------- | ------------------------------------------------ |
| `token_url`         | `string` | The URL to the OAuth 2.0 token endpoint.         |
| `authorization_url` | `string` | The URL to the OAuth 2.0 authorization endpoint. |
| `scopes`            | `string` | A space-delimited list of scopes.                |
| `issuer_url`        | `string` | The URL of the JWT issuer for token validation.  |
| `audience`          | `string` | The audience for JWT token validation.           |

#### `UpstreamAuthentication` (Outgoing)

Configures the authentication method for MCP Any to use when connecting to an upstream service.

| Field          | Type                      | Description                                        |
| -------------- | ------------------------- | -------------------------------------------------- |
| `api_key`      | `UpstreamAPIKeyAuth`      | API key sent in a header.                          |
| `bearer_token` | `UpstreamBearerTokenAuth` | Bearer token in the `Authorization` header.        |
| `basic_auth`   | `UpstreamBasicAuth`       | Basic authentication with a username and password. |
| `oauth2`       | `UpstreamOAuth2Auth`      | OAuth 2.0 client credentials flow.                 |

##### Use Case and Example

Authenticate with an upstream service using the OAuth 2.0 client credentials flow.

```yaml
upstream_auth:
  oauth2:
    token_url: "https://auth.example.com/oauth2/token"
    client_id:
      environment_variable: "UPSTREAM_CLIENT_ID"
    client_secret:
      file_path: "/secrets/upstream_client_secret"
    scopes: "read:data write:data"
```

##### Use Case and Example with Vault

Authenticate with an upstream service using an API key stored in HashiCorp Vault.

```yaml
upstream_auth:
  api_key:
    header_name: "X-API-Key"
    api_key:
      vault:
        address: "https://vault.example.com"
        token: "s.1234567890abcdef"
        path: "secret/data/my-app"
        key: "api_key"
```

##### Use Case and Example with AWS Secrets Manager

Authenticate using an API key stored in AWS Secrets Manager.

```yaml
upstream_auth:
  api_key:
    header_name: "X-API-Key"
    api_key:
      aws_secret_manager:
        secret_id: "my-app/api-keys"
        json_key: "my-api-key"
        region: "us-west-2"
```

##### `UpstreamAPIKeyAuth`

| Field         | Type          | Description                                  |
| ------------- | ------------- | -------------------------------------------- |
| `header_name` | `string`      | The name of the header to carry the API key. |
| `api_key`     | `SecretValue` | The API key value, managed as a secret.      |

##### `UpstreamBearerTokenAuth`

| Field   | Type          | Description                            |
| ------- | ------------- | -------------------------------------- |
| `token` | `SecretValue` | The bearer token, managed as a secret. |

##### `UpstreamBasicAuth`

| Field      | Type          | Description                            |
| ---------- | ------------- | -------------------------------------- |
| `username` | `string`      | The username for basic authentication. |
| `password` | `SecretValue` | The password, managed as a secret.     |

##### `UpstreamOAuth2Auth`

| Field           | Type          | Description                               |
| --------------- | ------------- | ----------------------------------------- |
| `token_url`     | `string`      | The URL to the OAuth 2.0 token endpoint.  |
| `client_id`     | `SecretValue` | The client ID for the OAuth 2.0 flow.     |
| `client_secret` | `SecretValue` | The client secret for the OAuth 2.0 flow. |
| `scopes`        | `string`      | A space-delimited list of scopes.         |

##### `SecretValue`

The `SecretValue` message provides a secure way to manage sensitive information like API keys, passwords, and tokens. It can be defined in one of the following ways:

| Field                  | Type            | Description                                                    |
| ---------------------- | --------------- | -------------------------------------------------------------- |
| `plain_text`           | `string`        | The secret value as a plain text string. **(Not Recommended)** |
| `environment_variable` | `string`        | The name of an environment variable containing the secret.     |
| `file_path`            | `string`        | The path to a file containing the secret.                      |
| `remote_content`       | `RemoteContent` | Fetches the secret from a remote URL.                          |
| `vault`                | `VaultSecret`   | Fetches the secret from a HashiCorp Vault instance.            |
| `aws_secret_manager`   | `AwsSecretManagerSecret` | Fetches the secret from AWS Secrets Manager.                   |

##### `VaultSecret`

| Field     | Type     | Description                                                          |
| --------- | -------- | -------------------------------------------------------------------- |
| `address` | `string` | The address of the Vault server (e.g., "https://vault.example.com"). |
| `token`   | `string` | The token to authenticate with Vault.                                |
| `path`    | `string` | The path to the secret in Vault (e.g., "secret/data/my-app/db").     |
| `key`     | `string` | The key of the secret to retrieve from the path.                     |

##### `AwsSecretManagerSecret`

| Field           | Type     | Description                                                                 |
| --------------- | -------- | --------------------------------------------------------------------------- |
| `secret_id`     | `string` | The name or ARN of the secret.                                              |
| `json_key`      | `string` | Optional: The key to extract from the secret JSON.                          |
| `version_stage` | `string` | Optional: The version stage (defaults to AWSCURRENT).                       |
| `version_id`    | `string` | Optional: The version ID.                                                   |
| `region`        | `string` | Optional: The AWS region. If not set, uses environment or profile defaults. |
| `profile`       | `string` | Optional: The AWS profile to use.                                           |

### TLS Configuration (`TLSConfig`)

Defines TLS settings for connecting to an upstream service.

| Field                  | Type     | Description                                                                               |
| ---------------------- | -------- | ----------------------------------------------------------------------------------------- |
| `server_name`          | `string` | The server name to use for SNI.                                                           |
| `ca_cert_path`         | `string` | Path to the CA certificate file for verifying the server's certificate.                   |
| `client_cert_path`     | `string` | Path to the client certificate file for mTLS.                                             |
| `client_key_path`      | `string` | Path to the client private key file for mTLS.                                             |
| `insecure_skip_verify` | `bool`   | If true, the client will not verify the server's certificate chain. **Use with caution.** |

### Use Case and Example

Connect to an upstream service that requires mutual TLS (mTLS) authentication.

```yaml
tls_config:
  server_name: "secure.internal.service"
  ca_cert_path: "/etc/ssl/certs/ca-bundle.crt"
  client_cert_path: "/etc/ssl/private/client.pem"
  client_key_path: "/etc/ssl/private/client.key"
```

## Defining Prompts

MCP Any allows you to define and execute prompts directly from your configuration files. This is useful for integrating with AI models and other services that require dynamic, template-based inputs.

| Field         | Type                     | Description                            |
| ------------- | ------------------------ | -------------------------------------- |
| `name`        | `string`                 | The name of the prompt.                |
| `description` | `string`                 | A description of what the prompt does. |
| `messages`    | `repeated PromptMessage` | The list of messages in the prompt.    |

### Use Case and Example

Here's an example of how to define a prompt in any service configuration (e.g., `http_service`):

```yaml
upstream_services:
  - name: "my-prompt-service"
    http_service:
      address: "https://api.example.com"
      prompts:
        - name: "my-prompt"
          description: "A sample prompt"
          messages:
            - role: "user"
              content:
                text: "Hello, {{name}}!"
```

You can then execute this prompt by sending a `prompts/get` request to the server.
