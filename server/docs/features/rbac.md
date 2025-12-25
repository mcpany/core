# Role-Based Access Control (RBAC)

RBAC allows you to manage user permissions by assigning roles to users.

## Overview

The RBAC implementation is located in `server/pkg/middleware/rbac.go` and `server/pkg/auth/rbac.go`.

It provides:
- **RBACEnforcer**: Checks if a user has a specific role.
- **Middleware**: HTTP middleware to enforce role requirements on endpoints.

## Configuration

RBAC is configured as part of the authentication configuration.

(Documentation in progress - refer to code for current implementation details)
