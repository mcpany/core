# Admin Management API

The Admin Management API allows you to programmatically manage services and configurations in MCP Any.

## Features

*   **Service Registration**: Register new services dynamically without restarting the server.
*   **Service Unregistration**: Unregister existing services.
*   **Configuration**: View and update server configuration.

## API Endpoints

### Services

*   `POST /v1/services/register`: Register a new service.
*   `POST /v1/services/unregister`: Unregister an existing service.

## Usage

You can use standard HTTP clients to interact with the Admin API. Ensure you have the necessary permissions/authentication if configured.

```bash
curl -X POST http://localhost:8080/v1/services/register \
  -H "Content-Type: application/json" \
  -d '{"name": "my-service", ...}'
```
