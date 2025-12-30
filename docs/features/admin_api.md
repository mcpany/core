# Admin Management API

**Status**: Implemented

The Admin Management API allows for runtime configuration and management of the MCP Any server. It primarily exposes a **Service Registration API** via gRPC, which is also accessible via HTTP (REST) using gRPC-Gateway.

## Service Registration API

This API is used to register, unregister, and query upstream services.

### Endpoints (HTTP)

All endpoints accept JSON bodies and return JSON responses.

*   **Register Service**
    *   `POST /v1/services/register`
    *   **Body**: `RegisterServiceRequest`
        *   `config`: `UpstreamServiceConfig` object.
    *   **Description**: Registers a new upstream service dynamically.

*   **Unregister Service**
    *   `POST /v1/services/unregister`
    *   **Body**: `UnregisterServiceRequest`
        *   `service_name`: The name of the service to unregister.
        *   `namespace`: (Optional) The namespace of the service.
    *   **Description**: Unregisters and removes a service.

*   **List Services**
    *   `GET /v1/services`
    *   **Response**: `ListServicesResponse`
        *   `services`: Array of `UpstreamServiceConfig`.
    *   **Description**: Returns a list of all registered upstream services.

*   **Get Service**
    *   `GET /v1/services/{service_name}`
    *   **Response**: `GetServiceResponse`
        *   `service`: `UpstreamServiceConfig`.
    *   **Description**: Returns the configuration for a specific service.

*   **Get Service Status**
    *   `GET /v1/services/{service_name}/status`
    *   **Response**: `GetServiceStatusResponse`
        *   `tools`: Array of `ToolDefinition`.
        *   `metrics`: Map of metrics.
    *   **Description**: Returns the status and discovered tools/resources for a service.

### gRPC Service

The underlying gRPC service is `mcpany.api.v1.RegistrationService`.

```protobuf
service RegistrationService {
    rpc RegisterService(RegisterServiceRequest) returns (RegisterServiceResponse);
    rpc UnregisterService(UnregisterServiceRequest) returns (UnregisterServiceResponse);
    rpc ListServices(ListServicesRequest) returns (ListServicesResponse);
    rpc GetService(GetServiceRequest) returns (GetServiceResponse);
    rpc GetServiceStatus(GetServiceStatusRequest) returns (GetServiceStatusResponse);
    rpc RegisterTools(RegisterToolsRequest) returns (RegisterToolsResponse);
    rpc InitiateOAuth2Flow(InitiateOAuth2FlowRequest) returns (InitiateOAuth2FlowResponse);
}
```

## Additional Endpoints

*   **Register Tools**
    *   `POST /v1/services/tools/register`
    *   **Body**: `RegisterToolsRequest`
        *   `service_name`: Name of the service.
        *   `namespace`: Namespace of the service.
        *   `tools`: Array of `ToolDefinition`.
    *   **Description**: Registers tools for an existing service.

*   **Initiate OAuth2 Flow**
    *   `POST /v1/services/oauth2/initiate`
    *   **Body**: `InitiateOAuth2FlowRequest`
        *   `service_id`: ID of the service.
        *   `namespace`: Namespace.
    *   **Description**: Initiates an OAuth2 flow for a service.

## Usage

Requests to the Admin API generally require Authentication (e.g., API Key) if configured.

**Example: List Services**

```bash
curl -H "X-API-Key: your-api-key" http://localhost:8080/v1/services
```

**Example: Register Service**

```bash
curl -X POST http://localhost:8080/v1/services/register \
  -H "Content-Type: application/json" \
  -d '{
    "config": {
      "name": "my-service",
      "http": {
        "baseUrl": "http://my-api.com"
      }
    }
  }'
```
