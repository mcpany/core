# Connection Pooling

Connection pooling allows MCP Any to maintain a pool of open connections to upstream services, reducing the overhead of establishing a new connection for every request. This is critical for high-throughput services.

## Configuration

Connection pooling is configured within the `connection_pool` block of an upstream service.

### Fields

| Field                  | Type     | Description                                                                |
| ---------------------- | -------- | -------------------------------------------------------------------------- |
| `max_connections`      | `int`    | The maximum number of simultaneous connections to the upstream service.    |
| `max_idle_connections` | `int`    | The maximum number of idle connections to keep in the pool.                |
| `idle_timeout`         | `string` | The duration a connection can remain idle in the pool before being closed. |

### Configuration Snippet

```yaml
upstream_services:
  - name: "pooled-database-service"
    connection_pool:
      max_connections: 100
      max_idle_connections: 10
      idle_timeout: "30s"
    http_service:
      address: "https://db-proxy.internal"
```

## Use Case

When connecting to a database proxy or a legacy backend that is sensitive to the number of concurrent connections, you can use connection pooling to limit the load. For example, setting `max_connections: 100` ensures that MCP Any will never open more than 100 connections to that service, queuing excess requests.

## Public API Example

The pooling behavior is internal. Clients simply make tool calls, and MCP Any manages the connections transparently.
