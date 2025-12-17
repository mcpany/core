# IP Allowlisting

MCP Any provides IP allowlisting to restrict access to the server based on the client's IP address. This is a critical security feature when running the server on a shared network or exposing it to the internet.

## Configuration

You can configure the allowed IP addresses or CIDR ranges in the `global_settings` section of your configuration file.

### Example

```yaml
global_settings:
  # ... other settings ...
  allowed_ips:
    - "127.0.0.1"        # Allow localhost
    - "192.168.1.0/24"   # Allow local network
    - "10.0.0.5"         # Allow specific internal server
    - "::1"              # Allow IPv6 localhost
```

## Behavior

- **Allow All**: If `allowed_ips` is not specified or is empty, the server allows requests from **all IP addresses**.
- **Block**: If `allowed_ips` is specified, any request from an IP address that does not match one of the entries will be rejected with a `403 Forbidden` status code.
- **Matching**: Both single IP addresses (e.g., `192.168.1.1`) and CIDR ranges (e.g., `192.168.1.0/24`) are supported.

## Use Cases

- **Developer Workstation**: Restrict access to `127.0.0.1` to ensure only local processes can call tools.
- **Internal Network**: Allow access only from your corporate VPN or VPC subnet (e.g., `10.0.0.0/8`).
- **CI/CD**: Allow access from specific CI runner IPs.

## Notes

- The check is performed against the `RemoteAddr` of the incoming HTTP request.
- If you are running `mcpany` behind a reverse proxy (like Nginx, Load Balancer, or Kubernetes Ingress), `RemoteAddr` might be the proxy's IP.
