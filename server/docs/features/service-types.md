# Service Types

MCP Any acts as a Universal Adapter, supporting a wide range of upstream service protocols. This allows you to expose diverse systems as standardized MCP tools without writing custom adapter code for each one.

## Supported Protocols

### 1. HTTP (`http_service`)
Connects to standard RESTful or JSON-over-HTTP APIs.
-   **Features**: Authentication, Headers, Query Parameters, JSON Body.
-   **Discovery**: Can be manually configured or auto-discovered if the API follows predictable patterns.

### 2. gRPC (`grpc_service`)
Connects to high-performance gRPC services.
-   **Features**: Protobuf serialization, Reflection (for auto-discovery), TLS.
-   **Discovery**: Uses gRPC Server Reflection to automatically discover available methods and expose them as tools.

### 3. OpenAPI / Swagger (`openapi_service`)
Connects to APIs defined by an OpenAPI Specification (v2 or v3).
-   **Features**: Auto-generates tools from the OpenAPI spec.
-   **Discovery**: Points to a `swagger.json` or `openapi.yaml` URL or file.

### 4. GraphQL (`graphql_service`)
Connects to GraphQL endpoints.
-   **Features**: Execute queries and mutations.
-   **Discovery**: Can introspect the schema to discover available queries and mutations (planned).

### 5. Command Line (`command_line_service`)
Executes local CLI commands or scripts.
-   **Features**: Captures stdout/stderr, supports working directories and environment variables.
-   **Use Cases**: Run Python scripts, shell commands, or binaries as tools.

### 6. MCP-to-MCP Proxy (`mcp_service`)
Connects to another existing MCP server.
-   **Features**: Aggregates tools/resources/prompts from other MCP servers into a single endpoint.
-   **Connection Types**:
    -   **Stdio**: Runs the MCP server as a subprocess.
    -   **HTTP**: Connects to a remote MCP server over SSE (Server-Sent Events).

### 7. WebSocket (`websocket_service`)
Connects to WebSocket servers.
-   **Features**: Send and receive messages.

### 8. WebRTC (`webrtc_service`)
Connects to services via WebRTC data channels.
-   **Features**: Low-latency peer-to-peer communication.

### 9. File System (`filesystem_service`)
Exposes local directories as tools.
-   **Features**: Read, write, list files safely within sandboxed directories.
-   **Use Cases**: Agentic coding, log analysis, local config management.
-   [Detailed Documentation](filesystem.md)

## Configuration Reference

For detailed configuration options for each service type, please refer to the [Configuration Reference](../reference/configuration.md) (if available) or the `upstream_services` section of your `config.yaml`.
