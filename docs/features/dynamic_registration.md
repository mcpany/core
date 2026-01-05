# Dynamic Tool Registration & Auto-Discovery

MCP Any supports dynamic registration of tools from upstream services. This means you don't need to manually define every tool; the server can discover them from:

- **gRPC Reflection**: Automatically discovers methods from gRPC services supporting reflection.
- **OpenAPI / Swagger**: Parses OpenAPI specifications to register REST endpoints as tools.
- **MCP-to-MCP**: Proxies tools from other MCP servers.

## Runtime Registration

Services can be registered at runtime without restarting the server using the **Admin API**. This allows for dynamic reconfiguration of the server.

The Admin API is a gRPC service defined in `proto/admin/v1/admin.proto`.

For more details on configuring these services, see:

- [Service Types](../../server/docs/features/service-types.md)
- [Configuration Reference](../../server/docs/reference/configuration.md)
