# MCP-X Architecture

## Introduction

MCP-X (Model Context Protocol eXtensions) is a versatile server designed to dynamically register and expose capabilities from various backend services as standardized "Tools." These tools can then be listed and executed through a unified interface. It supports a wide range of service types, including gRPC, RESTful APIs (via OpenAPI), generic HTTP services, and command-line tools that communicate over standard I/O.

## System Diagram

```mermaid
graph TD
    subgraph Client Interaction
        Client -->|List/Execute Tools| McpRouterApi[MCP Router API]
    end

    subgraph Service Registration
        Admin -->|Dynamic Registration| RegistrationApi[Registration API (gRPC)]
        ConfigFile[YAML/JSON Config] -->|Static Registration| McpxServer[MCP-X Server]
    end

    subgraph "MCP-X Server"
        McpxServer --> McpRouterApi
        McpxServer --> RegistrationApi
        McpxServer --> ToolIndex[Tool Index]
        McpxServer --> UpstreamClients[Upstream Clients]
    end

    subgraph Upstream Services
        UpstreamClients -- gRPC --> GrpcService[gRPC Service]
        UpstreamClients -- HTTP --> HttpService[HTTP Service]
        UpstreamClients -- HTTP --> OpenApiService[OpenAPI Service]
        UpstreamClients -- Stdio --> StdioService[Stdio Service]
    end

    style GrpcService fill:#d4edda
    style HttpService fill:#d4edda
    style OpenApiService fill:#d4edda
    style StdioService fill:#d4edda
```

The diagram above illustrates the two main workflows in MCP-X:

1.  **Service Registration**: Services can be registered with the MCP-X server either dynamically via the gRPC-based `Registration API` or statically by defining them in a configuration file that is loaded at startup.
2.  **Tool Execution**: Clients interact with the `MCP Router API` to list and execute the tools that have been made available by the registered services. The MCP-X server then dispatches these calls to the appropriate upstream service.

## Core Concepts

### 1. Upstream Services

An "upstream service" is a backend service that MCP-X can connect to and expose as a set of tools. MCP-X supports several types of upstream services, each with its own configuration:

- **gRPC (`GrpcUpstreamService`)**: Exposes a gRPC service as a set of tools. Can use `.proto` files or gRPC reflection to discover services and methods.
- **OpenAPI (`OpenapiUpstreamService`)**: Exposes a RESTful API as a set of tools by ingesting an OpenAPI (Swagger) specification.
- **HTTP (`HttpUpstreamService`)**: Exposes any HTTP endpoint as a tool.
- **Stdio (`StdioUpstreamService`)**: Wraps any command-line tool that communicates over standard I/O.
- **MCP-X Proxy (`McpUpstreamService`)**: Proxies and re-exposes tools from another MCP-X instance.

### 2. Service Configuration (`proto/config/v1/config.proto`)

The `McpxServerConfig` message is the root configuration for the entire MCP-X server. It contains:

- **`global_settings`**: Server-wide operational parameters, such as the bind address and log level.
- **`upstream_services`**: A list of all configured upstream services that MCP-X can proxy to. Each service has its own specific configuration (e.g., `GrpcUpstreamService`, `OpenapiUpstreamService`).
- **`frontend_services`**: A list of all defined public-facing frontend services.
- **`service_bindings`**: A list of bindings that link a frontend service to a specific upstream service.

### 3. Tool Definition (`proto/config/v1/config.proto`)

A `ToolDefinition` represents a single capability or "tool" offered by a service. It includes:

- **`name`**: The name of the tool, which will be used to invoke it.
- **`description`**: A human-readable description of what the tool does.
- **`input_schema`**: The schema for the input parameters required by the tool.
- **`is_stream`**: Indicates if the tool produces a continuous stream of responses.

## Components

### 1. MCP-X Server (`cmd/server`)

The `cmd/server` package contains the main application logic for the MCP-X server. It is responsible for:

- **Loading Configuration**: Reading the `McpxServerConfig` from a file or other source.
- **Service Registration**: Registering all the upstream services defined in the configuration.
- **Tool Indexing**: Creating and maintaining an index of all available tools.
- **API Server**: Exposing the MCP Router API and the Registration API.

### 2. Parsers

MCP-X uses a set of parsers to discover tools from different types of upstream services:

- **`pkg/grpc/protobufparser`**: Parses `.proto` files and uses gRPC reflection to discover gRPC services and methods.
- **`pkg/openapi/parser`**: Parses OpenAPI specifications to discover RESTful API endpoints.

### 3. API Services

MCP-X exposes two main API services:

- **`McpRouter` (`proto/mcp_router/v1/mcp_router.proto`)**: Allows clients to list and execute tools.
- **`RegistrationService` (`proto/api/v1/registration.proto`)**: Allows backend services to be registered dynamically.

## Tool Execution Flow

1. A user requests to execute a tool by its name, providing the necessary inputs.
2. The MCP-X server looks up the `ToolDefinition` from its index.
3. The server identifies the upstream service that provides the tool.
4. The server dispatches the request to the appropriate client logic for the upstream service type (gRPC, HTTP, etc.).
5. The client logic sends the request to the backend service.
6. The response from the backend service is returned to the user through the MCP-X server.

## Key Design Goals

- **Extensibility**: Easily add support for new service definition types or protocols.
- **Dynamic Nature**: Register and use tools without prior code generation for specific services.
- **Uniform Tool Interaction**: Provide a consistent way for users to list and execute tools, regardless of the underlying service type.
