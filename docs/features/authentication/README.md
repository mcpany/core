# Authentication

MCP Any supports comprehensive authentication mechanisms for both incoming requests (securing your MCP server) and outgoing requests (authenticating with upstream services). These are configured **per upstream service**.

## Configuration

Incoming authentication is configured under `authentication`. Outgoing authentication is configured under `upstream_authentication`.

### Incoming Authentication

To secure access to a specific service exposed by MCP Any:

```yaml
upstream_services:
  - name: "secure-service"
    authentication:
      api_key:
        param_name: "X-Mcp-Api-Key"
        in: "HEADER"
        key_value: "my-secret-key"
```

### Outgoing Authentication

To authenticate with an upstream service:

```yaml
upstream_services:
  - name: "secure-upstream"
    upstream_authentication:
      bearer_token:
        token:
          environment_variable: "UPSTREAM_TOKEN"
    http_service:
      address: "https://api.secure.com"
```

## Use Case

**Incoming**: You want to prevent unauthorized users from calling tool X.
**Outgoing**: Upstream API Y requires an API key or an OAuth token.

## Public API Example

Clients calling `secure-service` must provide the configured authentication (e.g., adding `X-Mcp-Api-Key` header).
