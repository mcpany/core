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
