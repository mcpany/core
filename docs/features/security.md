# Security

Security is a core design principle of MCP Any.

## Security Features

*   **Secrets Management**: Securely handle sensitive configuration values.
*   **IP Allowlisting**: Restrict access to the server based on IP address.
*   **Webhooks**: Trigger external actions on specific events.
*   **Fine-grained Policies**: Control which tools and resources are accessible.
*   **Role-Based Access Control (RBAC)**: (Planned) Manage user permissions.

## Configuration

Security features are configured in the global `security` section and per-service policies.

### UI Configuration

Global security settings like DLP (Data Loss Prevention) can be managed via the UI:

1. Navigate to **Settings > General**.
2. Toggle the **Enable DLP** switch.
3. Click **Save Changes**.

![Security Settings UI](../ui/screenshots/settings_general.png)
