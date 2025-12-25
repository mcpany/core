# Admin Management API

**Status**: Implemented

The Admin Management API allows for runtime configuration and management of the MCP Any server.

## Features

- **Service Registration**: Register new services dynamically.
- **Service Unregistration**: Remove existing services.
- **Status Checks**: Monitor the health and status of registered services.

## Endpoints

- `POST /v1/services/register`: Register a new service.
- `POST /v1/services/unregister`: Unregister a service.
- `GET /v1/services`: List all registered services.
- `GET /v1/services/{service_name}`: Get details of a specific service.
- `GET /v1/services/{service_name}/status`: Get the status (tools, metrics) of a specific service.

## Usage

### Register a Service

```bash
curl -X POST http://localhost:8080/v1/services/register \
  -H "Content-Type: application/json" \
  -d '{
    "config": {
      "name": "my-service",
      "type": "http",
      "http": {
        "base_url": "https://api.example.com"
      }
    }
  }'
```

### Unregister a Service

```bash
curl -X POST http://localhost:8080/v1/services/unregister \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "my-service"
  }'
```

### Get Service Status

```bash
curl http://localhost:8080/v1/services/my-service/status
```
