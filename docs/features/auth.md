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
*   **Enforcement**: The `RBACMiddleware` intercepts requests and enforces policy based on User Roles and Profile requirements.

RBAC logic is enforced globally via middleware, allowing for granular control over tool execution and service access based on the authenticated user's profile and roles.
