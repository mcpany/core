# Resilience

Resilience features help your MCP server handle failures in upstream services gracefully. The primary mechanism supported is the **Circuit Breaker**.

## Configuration

Resilience is configured within the `resilience` block.

### Circuit Breaker Fields

| Field                  | Type      | Description                                                               |
| ---------------------- | --------- | ------------------------------------------------------------------------- |
| `consecutive_failures` | `integer` | The number of consecutive failures that causes the circuit to open.       |
| `open_duration`        | `string`  | How long the circuit remains open before trying to recover (e.g., "10s"). |

### Configuration Snippet

```yaml
upstream_services:
  - name: "unstable-service"
    resilience:
      circuit_breaker:
        consecutive_failures: 5
        open_duration: "5s"
    http_service:
      address: "https://unstable.example.com"
```

## Use Case

If an upstream service starts failing 100% of requests (e.g., it's down), continuing to send requests wastes resources and slows down your server. A circuit breaker will detect these consecutive failures and "open", immediately failing subsequent requests locally for a set duration (`open_duration`), giving the upstream service time to recover.

## Public API Example

When the circuit is open, MCP Any will return an error indicating the service is unavailable, without attempting to contact the upstream.
