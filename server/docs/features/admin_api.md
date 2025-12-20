# Admin Management API

The Admin Management API provides a set of gRPC endpoints to inspect and manage the internal state of the MCP Any server, including dynamic service registration, caching, and tool inspection.

These endpoints are part of the `RegistrationService` defined in `proto/api/v1/registration.proto`.

## Service Definition

### Endpoints

#### `RegisterService`

Registers a new upstream service dynamically.

- **RPC**: `RegisterService`
- **HTTP**: `POST /v1/services/register`
- **Request**: `RegisterServiceRequest` containing `UpstreamServiceConfig`.
- **Response**: `RegisterServiceResponse` containing the service key and discovered tools.

#### `UpdateService`

Updates an existing upstream service configuration.

- **RPC**: `UpdateService`
- **HTTP**: `PUT /v1/services/update`
- **Request**: `UpdateServiceRequest` containing `UpstreamServiceConfig`.
- **Response**: `UpdateServiceResponse` containing the updated `UpstreamServiceConfig`.

#### `UnregisterService`

Unregisters (removes) an upstream service.

- **RPC**: `UnregisterService`
- **HTTP**: `POST /v1/services/unregister`
- **Request**: `UnregisterServiceRequest` containing `service_name`.
- **Response**: `UnregisterServiceResponse` (empty message).

#### `ListServices`

Returns a list of all currently registered upstream services.

- **RPC**: `ListServices`
- **HTTP**: `GET /v1/services`
- **Request**: `ListServicesRequest` (empty)
- **Response**: `ListServicesResponse` containing a list of `UpstreamServiceConfig`.

#### `GetService`

Returns the configuration for a specific service.

- **RPC**: `GetService`
- **HTTP**: `GET /v1/services/{service_name}`
- **Request**: `GetServiceRequest` containing `service_name`.
- **Response**: `GetServiceResponse` containing `UpstreamServiceConfig`.

#### `ListTools`

Returns a list of all registered tools across all services.

- **RPC**: `ListTools`
- **HTTP**: `GET /v1/tools`
- **Request**: `ListToolsRequest` (empty)
- **Response**: `ListToolsResponse` containing a list of `Tool`.

#### `GetTool`

Returns the definition of a specific tool by its name.

- **RPC**: `GetTool`
- **HTTP**: `GET /v1/tools/{tool_name}`
- **Request**: `GetToolRequest` containing `tool_name`.
- **Response**: `GetToolResponse` containing `Tool`.

#### `ClearCache`

Clears the global cache (if caching is enabled).

- **RPC**: `ClearCache`
- **HTTP**: `POST /v1/cache/clear`
- **Request**: `ClearCacheRequest` (empty)
- **Response**: `ClearCacheResponse` (empty)

## Usage

You can interact with the API using any gRPC client or HTTP client (if the gateway is enabled).

### Example with `grpcurl`

Assuming the gRPC server is running on `localhost:50051`:

```bash
# List all services
grpcurl -plaintext localhost:50051 mcpany.api.v1.RegistrationService/ListServices

# Get a service
grpcurl -plaintext -d '{"service_name": "my-service"}' localhost:50051 mcpany.api.v1.RegistrationService/GetService

# List all tools
grpcurl -plaintext localhost:50051 mcpany.api.v1.RegistrationService/ListTools

# Clear Cache
grpcurl -plaintext localhost:50051 mcpany.api.v1.RegistrationService/ClearCache
```
