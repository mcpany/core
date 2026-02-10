# Authentication

The `mcpany` server supports flexible authentication mechanisms for both incoming requests (securing your MCP server) and outgoing requests (authenticating with upstream services).

## Configuration

Incoming authentication is configured under `authentication`. Outgoing authentication is configured under `upstream_auth` for each service.

### Incoming Authentication

To secure access to a specific service exposed by MCP Any, you can use the following methods:

#### 1. API Key

```yaml
upstream_services:
  - name: "secure-service"
    authentication:
      api_key:
        param_name: "X-Mcp-Api-Key"
        in: "HEADER" # Options: HEADER, QUERY, COOKIE
        key_value: "my-secret-key"
```

#### 2. Basic Auth

```yaml
upstream_services:
  - name: "secure-service"
    authentication:
      basic_auth:
        username: "admin"
        password_hash: "$2a$10$..." # Bcrypt hash
```

#### 3. OAuth2 / OIDC

```yaml
upstream_services:
  - name: "secure-service"
    authentication:
      oidc:
        issuer: "https://accounts.google.com"
        audience: ["my-client-id"]
```

#### 4. Trusted Header

Useful when running behind a proxy that handles authentication.

```yaml
upstream_services:
  - name: "secure-service"
    authentication:
      trusted_header:
        header_name: "X-Forwarded-User"
```

### Outgoing Authentication

To authenticate with an upstream service, you can use:

#### 1. Bearer Token

```yaml
upstream_services:
  - name: "secure-upstream"
    upstream_auth:
      bearer_token:
        token:
          environment_variable: "UPSTREAM_TOKEN"
    http_service:
      address: "https://api.secure.com"
```

#### 2. API Key

```yaml
upstream_services:
  - name: "secure-upstream"
    upstream_auth:
      api_key:
        param_name: "apikey"
        in: "QUERY"
        value:
          environment_variable: "API_KEY"
```

#### 3. Basic Auth

```yaml
upstream_services:
  - name: "secure-upstream"
    upstream_auth:
      basic_auth:
        username: "user"
        password:
          environment_variable: "PASSWORD"
```

#### 4. OAuth2 (Client Credentials)

```yaml
upstream_services:
  - name: "secure-upstream"
    upstream_auth:
      oauth2:
        client_id:
          environment_variable: "CLIENT_ID"
        client_secret:
          environment_variable: "CLIENT_SECRET"
        token_url: "https://oauth.provider.com/token"
        scopes: "read write"
```

## Real World Example: IPInfo

This example demonstrates how to configure an upstream service (`ipinfo.io`) that requires an API key.

### Configuration

```yaml
upstream_services:
  - name: "ipinfo"
    http_service:
      address: "https://ipinfo.io"
      tools:
        - name: "get_ip_info"
          ignore_arguments: true
          http:
             endpoint_path: "/json"
             method: "GET"
    upstream_auth:
      bearer_token:
        token:
          environment_variable: "IPINFO_API_KEY"
```

### Verification

Set the environment variable and start the server:

```bash
export IPINFO_API_KEY="your_actual_token_here"
./bin/mcp-any --config config.yaml
```
