# Admin Management API

MCP Any provides an Admin API for managing services and configurations dynamically at runtime.

## Service Management

### Register a Service

Dynamically register a new service (upstream) without restarting the server.

**Endpoint:** `POST /v1/services/register`

**Request Body:**

```json
{
  "service": {
    "id": "my-service",
    "type": "HTTP",
    "config": {
      "baseUrl": "https://api.example.com"
    }
  }
}
```

### Unregister a Service

Remove a service at runtime.

**Endpoint:** `POST /v1/services/unregister`

**Request Body:**

```json
{
  "serviceId": "my-service"
}
```

## Future Enhancements

The Admin API will be expanded to support full CRUD operations on all configuration aspects, including:
- Policies (Rate Limits, Caching)
- Authentication Providers
- User Management (RBAC)
