# Admin Management API

**Status**: Implemented

The Admin Management API allows for runtime configuration and management of the MCP Any server. It exposes both a gRPC-based API (via gRPC-Gateway) and a RESTful Configuration API.

## RESTful Configuration API

The Configuration API is available at `/api/v1/services`. It supports standard CRUD operations for managing upstream services.

### Endpoints

*   **List Services**
    *   `GET /api/v1/services`
    *   Returns a list of all registered upstream services.

*   **Create Service**
    *   `POST /api/v1/services`
    *   Body: `UpstreamServiceConfig` JSON object.
    *   Registers a new upstream service.

*   **Get Service**
    *   `GET /api/v1/services/{name}`
    *   Returns the configuration for a specific service.

*   **Update Service**
    *   `PUT /api/v1/services/{name}`
    *   Body: `UpstreamServiceConfig` JSON object.
    *   Updates an existing service configuration.

*   **Delete Service**
    *   `DELETE /api/v1/services/{name}`
    *   Unregisters and deletes the service configuration.

## Service Registration API (gRPC / gRPC-Gateway)

This API focuses on operational service registration and status checks, primarily used by the `mcp-cli` or internal components.

### Endpoints

- `POST /v1/services/register`: Register a new service dynamically.
- `POST /v1/services/unregister`: Unregister a service.
- `GET /v1/services`: List all services.
- `GET /v1/services/{service_name}`: Get details of a specific service.
- `GET /v1/services/{service_name}/status`: Get status/metrics of a service.

## Usage

Requests to the Admin API generally require Authentication (e.g., API Key).

```bash
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/services
```
