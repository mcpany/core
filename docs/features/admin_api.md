# Admin API

The Admin API provides programmatic access to manage and inspect the MCP Any server. It is exposed via gRPC on the same port as the Registration API.

## Features

- **ListServices**: List all registered upstream services.
- **GetService**: Get the configuration of a specific service.
- **ListTools**: List all registered tools.
- **GetTool**: Get the definition of a specific tool.
- **ClearCache**: Clear the internal cache.

## Usage

You can interact with the Admin API using any gRPC client (like `grpcurl` or standard gRPC libraries).

### Example: List Services

```bash
grpcurl -plaintext localhost:50051 mcpany.admin.v1.AdminService/ListServices
```

### Proto Definition

See `proto/admin/v1/admin.proto` for the full service definition.
