# Rate Limiting

Rate limiting protects your upstream services from being overwhelmed by too many requests. It uses a token bucket algorithm to enforce limits on a per-service basis.

## Configuration

Rate limiting is configured within the `rate_limit` block of an upstream service.

### Fields

| Field                 | Type     | Description                                                  |
| --------------------- | -------- | ------------------------------------------------------------ |
| `is_enabled`          | `bool`   | Whether rate limiting is enabled.                            |
| `requests_per_second` | `double` | The maximum number of requests allowed per second.           |
| `burst`               | `int64`  | The number of requests that can be allowed in a short burst. |

### Configuration Snippet

```yaml
upstream_services:
  - name: "limited-service"
    rate_limit:
      is_enabled: true
      requests_per_second: 10.0
      burst: 20
    http_service:
      address: "https://api.example.com"
```

## Use Case

If an upstream API charges you based on request volume or has strict quotas, you can use rate limiting to ensure you never exceed those quotas. For example, `requests_per_second: 10.0` ensures a steady flow of at most 10 requests/sec.

## Public API Example

If a client sends requests faster than the limit, MCP Any will return a `429 Too Many Requests` error (or equivalent gRPC status) for the excess requests.
