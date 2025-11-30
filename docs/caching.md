# Caching

MCP Any provides a caching middleware that can be used to cache the results of tool executions. This can be useful for improving performance and reducing the number of calls to upstream services.

## Configuration

Caching is configured at the service level in the `mcp-any.yaml` configuration file. The following options are available:

- `is_enabled`: A boolean value that indicates whether caching is enabled for the service.
- `ttl`: The time-to-live for cached entries, specified as a duration (e.g., "10s", "5m").

Here is an example of how to configure caching for an HTTP service:

```yaml
upstream_services:
  - name: "test-service"
    http_service:
      address: "http://localhost:8080"
      calls:
        test_call:
          endpoint_path: "/"
          method: "HTTP_METHOD_GET"
      tools:
        - name: "test-tool"
          call_id: "test_call"
    cache:
      is_enabled: true
      ttl: "10s"
```

## Implementation

The caching middleware is implemented in `pkg/middleware/cache.go`. It uses an in-memory cache with a TTL to store the results of tool executions. The cache key is generated from the tool name and the tool inputs.
