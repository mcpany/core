# Admin Management API

The Admin Management API provides programmatic access to manage and monitor the MCP Any server.

> **Status**: Partial Implementation. currently supports read-only operations for introspection and cache management. Full CRUD for service registration is available via the Registration API.

## Features

### 1. Service Inspection
Retrieve the list of currently registered upstream services and their configurations.

- **List Services**: Get all active services.
- **Get Service**: Retrieve details for a specific service by ID.

### 2. Tool Introspection
View all tools exposed by the aggregated MCP server.

- **List Tools**: See all available tools across all services.
- **Get Tool**: Inspect the schema and definition of a specific tool.

### 3. Cache Management
- **Clear Cache**: Invalidate the global cache (useful during development or after configuration changes).

## Future Scope

The Admin API is planned to be expanded to support:
- **Full CRUD**: Create, Update, and Delete services at runtime without restarting.
- **User Management**: Manage users, roles, and API keys.
- **Policy Configuration**: Dynamically adjust rate limits and access controls.
