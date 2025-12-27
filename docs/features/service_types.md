# Service Types

MCP Any supports a wide range of service types to expose various backends as MCP tools.

## Supported Service Types

*   **gRPC**: Register services from `.proto` files or by using gRPC reflection.
*   **OpenAPI**: Ingest OpenAPI (Swagger) specifications to expose RESTful APIs as tools.
*   **HTTP**: Expose any HTTP endpoint as a tool.
*   **GraphQL**: Expose a GraphQL API as a set of tools.
*   **Stdio**: Interact with command-line tools.
*   **MCP-to-MCP Proxy**: Proxy and re-expose tools from another MCP Any instance.
*   **WebSocket**: Connect to WebSocket endpoints.
*   **WebRTC**: Connect via WebRTC.
*   **SQL**: (Beta) Connect to SQL databases and expose queries as tools.
*   **Bundle**: Register a bundled MCP service (e.g. a zip file containing a Python script and dependencies).

## Configuration

Service types are configured in the `upstream_services` section of the configuration file.

See `docs/reference/configuration.md` for detailed configuration options.
