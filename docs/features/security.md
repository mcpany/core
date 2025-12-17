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
