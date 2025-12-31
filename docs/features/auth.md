# Authentication

MCP Any supports various upstream authentication methods to securely connect to backend services, as well as incoming authentication to secure the MCP server itself.

## Supported Authentication Methods

### Upstream Authentication (Outgoing)

When MCP Any connects to upstream services, it can use the following methods:

*   **API Key**: Pass an API key in a header.
*   **Bearer Token**: Use a Bearer token in the `Authorization` header.
*   **OAuth 2.0**: Support for Client Credentials flow.
*   **Basic Auth**: Username and password.
*   **mTLS**: Mutual TLS for secure communication.

### Server Authentication (Incoming)

To secure access to the MCP Any server:

*   **API Key**: Clients must provide an API key in a header or query parameter.
*   **OAuth 2.0 / OIDC**: Validate JWT tokens from an OIDC issuer. This leverages OIDC Discovery to configure the provider.

## Configuration

Authentication is configured in the `upstream_services` section (for outgoing) or `global_settings` / `authentication` (for incoming).

For detailed configuration options and examples, see [Configuration Reference](../../server/docs/reference/configuration.md#authentication).

## Role-Based Access Control (RBAC)

MCP Any implements a core RBAC engine to manage user permissions.

*   **Roles**: Define sets of permissions.
*   **Users**: Assigned one or more roles.
*   **Enforcement**: The `RBACEnforcer` checks if a user has the required role to access a resource or perform an action.

### Middleware

A unified RBAC middleware is available in `pkg/middleware/rbac.go` to enforce role requirements on HTTP endpoints.

```go
rbac := middleware.NewRBACMiddleware()
mux.Handle("/admin", rbac.RequireRole("admin")(adminHandler))
```

This middleware relies on the request context having roles populated (e.g., via `auth.ContextWithRoles`).
