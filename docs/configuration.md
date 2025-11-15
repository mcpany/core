# Configuration Reference

This document provides a detailed reference for all configuration options available in the `config.yaml` file for MCP Any.

## Global Settings

-   **`mcpListenAddress`** (string): The address on which the MCP server listens for JSON-RPC requests. Defaults to `:50050`.
-   **`grpcPort`** (string): The port for the gRPC registration server. If not specified, gRPC registration is disabled.
-   **`logLevel`** (string): The log level for the server. Can be `debug`, `info`, `warn`, or `error`. Defaults to `info`.
-   **`shutdownTimeout`** (duration): The graceful shutdown timeout. Defaults to `5s`.

## Upstream Services

The `upstreamServices` section is a list of backend services that MCP Any will connect to and expose as tools.

### Common Service Configuration

-   **`name`** (string, required): A unique name for the service.
-   **`disable`** (boolean): If `true`, this service will be disabled.
-   **`upstreamAuthentication`**: Configures authentication for the upstream service. See [Authentication](#authentication).
-   **`cache`**: Configures caching for the service. See [Caching](#caching).

### HTTP Service (`httpService`)

-   **`address`** (string, required): The base URL of the HTTP service.
-   **`calls`** (list, required): A list of HTTP calls to be exposed as tools.
    -   **`toolName`** (string, required): The name of the tool.
    -   **`description`** (string): A description of the tool.
    -   **`method`** (string, required): The HTTP method (e.g., `GET`, `POST`).
    -   **`endpointPath`** (string, required): The endpoint path, which can include path parameters (e.g., `/users/{userId}`).

### gRPC Service (`grpcService`)

-   **`address`** (string, required): The address of the gRPC service.
-   **`reflection`**: Configures gRPC reflection.
    -   **`enabled`** (boolean): If `true`, the server will use reflection to discover services and methods.

### OpenAPI Service (`openapiService`)

-   **`address`** (string): The base URL of the OpenAPI service. If not provided, the server will be inferred from the OpenAPI specification.
-   **`spec`**: The OpenAPI specification.
    -   **`path`** (string): The path to the OpenAPI specification file.

### Command-Line Service (`commandLineService`)

-   **`command`** (string, required): The command to be executed.
-   **`timeout`** (duration): A timeout for the command execution.
-   **`workingDirectory`** (string): The working directory for the command.

## Authentication

-   **`apiKey`**:
    -   **`headerName`** (string, required): The name of the HTTP header to send the API key in.
    -   **`apiKey`** (string, required): The API key.
-   **`bearerToken`**:
    -   **`token`** (string, required): The bearer token.
-   **`basicAuth`**:
    -   **`username`** (string, required): The username.
    -   **`password`** (string, required): The password.

## Caching

-   **`isEnabled`** (boolean): If `true`, caching is enabled for the service or tool.
-   **`ttl`** (duration): The time-to-live for cached items.
