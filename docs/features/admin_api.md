# Admin API

The Admin API provides administrative operations for managing the MCP Any server. It is a gRPC service that allows you to register and unregister upstream services, list tools, and clear caches.

## Service Definition

The Admin API is defined in `proto/admin/v1/admin.proto`.

### Methods

#### CreateService

Registers a new upstream service dynamically.

```protobuf
rpc CreateService(CreateServiceRequest) returns (CreateServiceResponse);
```

**Request:** `CreateServiceRequest`
- `service`: `mcpany.config.v1.UpstreamServiceConfig` - The configuration of the service to register.

**Response:** `CreateServiceResponse`
- `service_id`: `string` - The ID of the registered service.
- `tools`: `repeated mcpany.config.v1.ToolDefinition` - List of discovered tools.
- `resources`: `repeated mcpany.config.v1.ResourceDefinition` - List of discovered resources.

#### DeleteService

Unregisters an existing upstream service.

```protobuf
rpc DeleteService(DeleteServiceRequest) returns (DeleteServiceResponse);
```

**Request:** `DeleteServiceRequest`
- `service_id`: `string` - The ID of the service to delete.

**Response:** `DeleteServiceResponse` (Empty)

#### ListServices

Lists all registered upstream services.

```protobuf
rpc ListServices(ListServicesRequest) returns (ListServicesResponse);
```

#### GetService

Retrieves configuration for a specific service.

```protobuf
rpc GetService(GetServiceRequest) returns (GetServiceResponse);
```

#### ListTools

Lists all registered tools across all services.

```protobuf
rpc ListTools(ListToolsRequest) returns (ListToolsResponse);
```

#### GetTool

Retrieves details for a specific tool.

```protobuf
rpc GetTool(GetToolRequest) returns (GetToolResponse);
```

#### ClearCache

Clears the internal cache.

```protobuf
rpc ClearCache(ClearCacheRequest) returns (ClearCacheResponse);
```
