# Role-Based Access Control (RBAC)

RBAC allows you to manage user permissions by assigning roles to users.

## Overview

The RBAC implementation is located in `server/pkg/middleware/rbac.go` and `server/pkg/auth/rbac.go`.

It provides:
- **RBACEnforcer**: Checks if a user has a specific role.
- **Middleware**: HTTP middleware to enforce role requirements on endpoints.

## Configuration

RBAC is configured as part of the authentication configuration.

### User Configuration

Users are assigned roles in their configuration:

```yaml
users:
  - id: "admin-user"
    roles: ["admin"]
    authentication:
      api_key: "admin-secret-key"
  - id: "viewer-user"
    roles: ["viewer"]
    authentication:
      api_key: "viewer-secret-key"
```

### Profile Configuration

You can restrict access to specific profiles based on roles:

```yaml
global_settings:
  profile_definitions:
    - name: "prod-profile"
      required_roles: ["admin"]
```

## Middleware Usage

The `RBACMiddleware` provides methods to enforce role requirements on HTTP handlers.

### `RequireRole`

Requires the user to have a specific role.

```go
rbacMiddleware := middleware.NewRBACMiddleware()
handler := rbacMiddleware.RequireRole("admin")(myHandler)
```

### `RequireAnyRole`

Requires the user to have at least one of the specified roles.

```go
rbacMiddleware := middleware.NewRBACMiddleware()
handler := rbacMiddleware.RequireAnyRole("admin", "editor")(myHandler)
```

## Context Integration

The RBAC system integrates with the request context to retrieve user roles.

- **`auth.ContextWithRoles(ctx, roles)`**: Adds roles to the context.
- **`auth.RolesFromContext(ctx)`**: Retrieves roles from the context.

This allows downstream handlers to make decisions based on the user's roles.
