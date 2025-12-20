# Admin API

The Admin API provides capabilities to manage and monitor the MCP Any server.

## Service Health Monitoring

The Admin API `ListServices` method now returns the runtime health status of each service.

### Status Enum

- `SERVICE_STATUS_UNKNOWN`: Initial state.
- `SERVICE_STATUS_HEALTHY`: The service is reachable and responding correctly to health checks.
- `SERVICE_STATUS_UNHEALTHY`: The service failed the last health check.
- `SERVICE_STATUS_DEGRADED`: (Reserved for future use)

### Service State

The response includes `service_states` field containing:
- `config`: The static configuration of the service.
- `status`: The current health status.
- `last_error`: Error message if status is UNHEALTHY.
- `last_check_time`: Timestamp of the last health check.

The `services` field is deprecated and returns only the configuration.

Health checks are performed periodically (every 30 seconds) in the background.
