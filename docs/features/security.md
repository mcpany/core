# Security Features

## IP Allowlisting

MCP Any supports restricting access to the server based on IP addresses. This is useful when the server is exposed on a shared network or to limit access to trusted clients.

### Configuration

You can configure the allowed IP addresses or CIDR ranges in the `global_settings` section of your configuration file.

```yaml
global_settings:
  allowed_ips:
    - "127.0.0.1"       # Allow local access
    - "192.168.1.0/24"  # Allow local network
    - "::1"             # Allow IPv6 local access
```

If `allowed_ips` is empty or not specified, the server accepts connections from any IP address.

### Behavior

-   **HTTP/JSON-RPC**: Incoming HTTP requests are checked against the allowlist. If the client IP is not allowed, the server responds with `403 Forbidden`.
-   **gRPC**: Incoming gRPC calls are checked. If the client IP is not allowed, the call fails with `PermissionDenied`.

**Note:** This feature checks the immediate remote address of the connection (`RemoteAddr`). If you are running `mcpany` behind a reverse proxy or load balancer, `RemoteAddr` will be the IP of the proxy. Ensure your proxy is configured to restrict access or is trusted.

## Secrets Management

MCP Any provides a unified way to handle sensitive information (API keys, passwords, tokens) securely. Instead of hardcoding secrets in your configuration files, you can reference them using various providers.

### Supported Secret Providers

You can use `SecretValue` wherever a secret is expected (e.g., in authentication configurations). The supported providers are:

*   **Plain Text**: (Not recommended for production)
*   **Environment Variable**: Read from an environment variable.
*   **File Path**: Read from a file on the local filesystem.
*   **Remote Content**: Fetch from an HTTP URL.
*   **HashiCorp Vault**: Fetch from a Vault KV secret engine.
*   **AWS Secrets Manager**: Fetch from AWS Secrets Manager.

### Configuration Examples

#### Environment Variable

```yaml
upstream_authentication:
  api_key:
    header_name: "X-API-Key"
    api_key:
      environment_variable: "MY_API_KEY"
```

#### File Path

```yaml
upstream_authentication:
  bearer_token:
    token:
      file_path: "/var/secrets/token.txt"
```

#### HashiCorp Vault

```yaml
upstream_authentication:
  basic_auth:
    username: "admin"
    password:
      vault:
        address: "https://vault.example.com"
        token:
          environment_variable: "VAULT_TOKEN"
        path: "secret/data/myapp/database"
        key: "password"
```

#### AWS Secrets Manager

```yaml
upstream_authentication:
  api_key:
    header_name: "Authorization"
    api_key:
      aws_secret_manager:
        secret_id: "prod/myapp/api-key"
        region: "us-west-2"
        json_key: "api_key_value" # Optional: if secret is JSON
```
