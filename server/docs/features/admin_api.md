# Admin Management API

The Admin Management API provides a set of gRPC endpoints to inspect and manage the internal state of the MCP Any server. This is useful for building dashboards, debugging, and monitoring the server's configuration and registered tools.

## Service Definition

The Admin API is exposed as a gRPC service defined in `proto/admin/v1/admin.proto`.

### Endpoints

#### `ListServices`

Returns a list of all currently registered upstream services.

- **Request**: `ListServicesRequest` (empty)
- **Response**: `ListServicesResponse` containing a list of `ServiceState` (which includes `config`, `status`, and `error`).

#### `GetService`

Returns the configuration for a specific service by its ID.

- **Request**: `GetServiceRequest` containing `service_id`.
- **Response**: `GetServiceResponse` containing `UpstreamServiceConfig`.

#### `ListTools`

Returns a list of all registered tools across all services.

- **Request**: `ListToolsRequest` (empty)
- **Response**: `ListToolsResponse` containing a list of `Tool`.

#### `GetTool`

Returns the definition of a specific tool by its name.

- **Request**: `GetToolRequest` containing `tool_name`.
- **Response**: `GetToolResponse` containing `Tool`.

#### `ClearCache`

Clears the global cache (if caching is enabled).

- **Request**: `ClearCacheRequest` (empty)
- **Response**: `ClearCacheResponse` (empty)

## Usage

You can interact with the Admin API using any gRPC client, such as `grpcurl` or by generating a client in your preferred language using the provided protobuf definition.

### Example with `grpcurl`

Assuming the gRPC server is running on `localhost:50051`:

```bash
# List all services
grpcurl -plaintext localhost:50051 mcpany.admin.v1.AdminService/ListServices

# List all tools
grpcurl -plaintext localhost:50051 mcpany.admin.v1.AdminService/ListTools
```
