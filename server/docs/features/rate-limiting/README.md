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
| `storage`             | `enum`   | The storage backend to use: `STORAGE_MEMORY` (default) or `STORAGE_REDIS`. |
| `redis`               | `object` | Redis connection details (required if storage is `STORAGE_REDIS`). |
| `tool_limits`         | `map`    | Tool-specific rate limits. Key is the tool name, value is a RateLimitConfig object. |

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

### Token-Based Rate Limiting

In addition to request-based limiting, you can limit based on the number of "tokens" (e.g., words or characters) in the request arguments. This is useful for controlling costs when sending data to LLMs.

To enable token-based limiting, set `cost_metric` to `COST_METRIC_TOKENS`. The system will estimate the token count of the request arguments.

```yaml
upstream_services:
  - name: "token-limited-service"
    rate_limit:
      is_enabled: true
      requests_per_second: 1000.0 # Limit in tokens per second
      burst: 5000
      cost_metric: COST_METRIC_TOKENS
    http_service:
      address: "https://api.example.com"
```

### Distributed Rate Limiting (Redis)

By default, rate limiting is handled in-memory. To support distributed deployments (multiple replicas), you can configure a Redis backend.

```yaml
upstream_services:
  - name: "limited-service-redis"
    rate_limit:
      is_enabled: true
      requests_per_second: 100.0
      burst: 50
      storage: STORAGE_REDIS
      redis:
        address: "redis-host:6379"
        password: "optional-password"
        db: 0
    http_service:
      address: "https://api.example.com"
```

## Use Case

If an upstream API charges you based on request volume or has strict quotas, you can use rate limiting to ensure you never exceed those quotas. For example, `requests_per_second: 10.0` ensures a steady flow of at most 10 requests/sec.

## Execution Order

Rate limiting is applied **after** caching. This means that:
1.  If a request is found in the cache, it is returned immediately and **does not** count towards the rate limit.
2.  If a request is not in the cache (cache miss), it passes through the rate limiter.
    - If allowed, it proceeds to the upstream tool.
    - If blocked, an error is returned.

## Metrics

The following metrics are exposed for rate limiting:

| Metric Name | Type | Labels | Description |
| :--- | :--- | :--- | :--- |
| `rate_limit_requests_total` | Counter | `service_id`, `limit_type`, `status` | Total number of requests processed by the rate limiter. `status` can be `allowed` or `blocked`. `limit_type` indicates the scope (e.g., `service`, `tool`). |

## Public API Example

If a client sends requests faster than the limit, MCP Any will return an error indicating that the rate limit has been exceeded for the excess requests.

## Tutorial: Verifying Rate Limiting with Gemini CLI

This tutorial guides you through setting up a rate-limited service and verifying it using the Gemini CLI.

### Prerequisites
- `mcp-any` binary installed (or run via `go run`)
- `gemini` CLI tool installed

### 1. Create a Configuration File
Create a file named `tutorial_config.yaml` with the following content (or use the provided `server/docs/features/rate-limiting/tutorial_config.yaml`):

```yaml
upstream_services:
  - name: "httpbin-rate-limited"
    http_service:
      address: "https://httpbin.org"
      calls:
        get-call:
          endpoint_path: "/get"
          method: "HTTP_METHOD_GET"
      tools:
        - name: "get"
          call_id: "get-call"
    rate_limit:
      is_enabled: true
      # Strict limit: 1 request per second
      requests_per_second: 1.0
      burst: 1
```

### 2. Start the MCP Any Server
Run the server with the configuration file:

```bash
# Assuming you are in the root of the repo
go run ./cmd/server run --config-path ./server/docs/features/rate-limiting/tutorial_config.yaml --mcp-listen-address :8080 --metrics-listen-address :8081
```

The server should start and listen on port `8080`.

### 3. Verify Rate Limiting with Gemini CLI
Open a new terminal to run the Gemini CLI.

**Step 3a: Run a single request (Allowed)**
```bash
gemini run httpbin-rate-limited.get
```
*Expected Output:* JSON response from httpbin.

**Step 3b: Trigger Rate Limit (Blocked)**
Run multiple requests in rapid succession.
```bash
for i in {1..5}; do gemini run httpbin-rate-limited.get; echo "Request $i done"; done
```

*Expected Output:*
- The first request should succeed.
- Subsequent requests (within the same second) should fail with an error similar to:
  > Error: rate limit exceeded for service httpbin-rate-limited

### 4. Observe Metrics
Visit the metrics endpoint (default: `http://localhost:8081/metrics`) to see the counters increase.

```bash
curl http://localhost:8081/metrics | grep rate_limit
```

*Expected Output:*
```
# HELP rate_limit_requests_total Total number of requests processed by the rate limiter
# TYPE rate_limit_requests_total counter
rate_limit_requests_total{service_id="httpbin-rate-limited",status="allowed"} 1.0
rate_limit_requests_total{service_id="httpbin-rate-limited",status="blocked"} 4.0
```
