# Authentication

MCP Any supports various upstream authentication methods to securely connect to backend services.

## Supported Authentication Methods

*   **API Key**: Pass an API key in a header or query parameter.
*   **Bearer Token**: Use a Bearer token in the Authorization header.
*   **OAuth 2.0**: Support for Client Credentials and other flows.
*   **Basic Auth**: Username and password.
*   **mTLS**: Mutual TLS for secure communication.

## Configuration

Authentication is configured per service in the `upstream_services` section.

See `docs/reference/configuration.md` for detailed configuration options.

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
