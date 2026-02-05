# Caching

MCP Any provides a caching middleware that can be used to cache the results of tool executions. This can be useful for improving performance and reducing the number of calls to upstream services, especially for data-intensive or rate-limited APIs.

## Configuration

Caching can be configured at the **service level** (applying to all tools in the service) or at the **call level** (applying to specific tools). Call-level configuration overrides service-level settings.

### Fields

| Field        | Type     | Description                                                    |
| ------------ | -------- | -------------------------------------------------------------- |
| `is_enabled` | `bool`   | Whether caching is enabled for this service.                   |
| `ttl`        | `string` | The time-to-live for cached entries (e.g., "10s", "5m", "1h"). |
| `strategy`   | `string` | The caching strategy to use (default: exact match).            |
| `semantic_config` | `object` | Configuration for semantic caching (see below).               |

### Configuration Snippet

```yaml
upstream_services:
  - name: "cached-weather-service"
    http_service:
      address: "https://api.weather.com"
      tools:
        - name: "get_forecast"
          description: "Get weather for a location"
          call_id: "get_weather"
          input_schema:
            type: "object"
            properties:
              location:
                type: "string"
            required: ["location"]
      calls:
        get_weather:
          method: HTTP_METHOD_GET
          endpoint_path: "/weather"
          parameters:
            - schema:
                name: "location"
                type: STRING
                is_required: true
    cache:
      is_enabled: true
      ttl: "5m"
```

## Use Case

Imagine you have a weather service that users query frequently. The weather forecast doesn't change every second, so querying the upstream API for every user request is inefficient and might consume your API rate limits.

By enabling caching with a TTL of 5 minutes, MCP Any will serve repeated requests for the same location from its internal memory, significantly reducing latency and upstream load.

## Public API Example

When using the MCP Any server with this configuration, you can call the tool as usual. The caching happens transparently.

**Request 1 (Miss):**
Client -> MCP Any -> Upstream API -> MCP Any (Cache Store) -> Client

**Request 2 (Hit):**
Client -> MCP Any (Cache Hit) -> Client

## Verification

You can verify that caching is working by using the `gemini` CLI or checking the server logs.

### Using Gemini CLI

1.  **Start the MCP Any server** with your configuration.
2.  **Add the server to Gemini CLI**:
    ```bash
    gemini mcp add my-server http://localhost:50050/mcp
    ```
3.  **Run a prompt** that triggers the tool:
    ```bash
    gemini -p "What is the weather in London?"
    ```
    *First run:* You should see the request being processed by the upstream service (e.g., in server logs).
4.  **Run the prompt again**:
    ```bash
    gemini -p "What is the weather in London?"
    ```
    *Second run:* The response should be faster, and you should NOT see a new request to the upstream service in the logs.

### Checking Server Logs

Enable debug logging in your configuration (`log_level: "debug"`) or set the environment variable `MCPANY_LOG_LEVEL=debug`.

**Cache Miss (First Request):**
```text
level=DEBUG msg="Cache miss" key=...
level=INFO msg="Call upstream" ...
```

**Cache Hit (Second Request):**
```text
level=DEBUG msg="Cache hit" key=...
```

### Automated Verification

Run the E2E verification test:

1.  Build the server binary:
    ```bash
    make build
    ```
2.  Run the test:
    ```bash
    go test -v -count=1 -tags=e2e ./server/docs/features/caching
    ```

### Metrics

The caching middleware exposes the following Prometheus metrics on the configured metrics port (default: 9091):

- `mcpany_cache_hits`: Counter of cache hits, labeled by `service` and `tool`.
- `mcpany_cache_misses`: Counter of cache misses, labeled by `service` and `tool`.
- `mcpany_cache_errors`: Counter of cache errors, labeled by `service` and `tool`.
