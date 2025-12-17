# Caching

MCP Any provides a caching middleware that can be used to cache the results of tool executions. This can be useful for improving performance and reducing the number of calls to upstream services, especially for data-intensive or rate-limited APIs.

## Configuration

Caching can be configured at the **service level** (applying to all tools in the service) or at the **call level** (applying to specific tools). Call-level configuration overrides service-level settings.

### Fields

| Field        | Type     | Description                                                    |
| ------------ | -------- | -------------------------------------------------------------- |
| `is_enabled` | `bool`   | Whether caching is enabled for this service.                   |
| `ttl`        | `string` | The time-to-live for cached entries (e.g., "10s", "5m", "1h"). |

### Configuration Snippet

```yaml
upstream_services:
  - name: "cached-weather-service"
    http_service:
      address: "https://api.weather.com"
      tools:
        - name: "get_forecast"
      calls:
        get_forecast:
          cache:
            is_enabled: true
            ttl: "5m" # Overrides service default
    cache:
      is_enabled: true
      ttl: "1h"
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
