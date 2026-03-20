# Resilience

Resilience features help your MCP server handle failures in upstream services gracefully. The primary mechanisms supported are **Retry Policy** and **Circuit Breaker**.

## Configuration

Resilience is configured within the `resilience` block.

### Retry Policy Fields

| Field                    | Type     | Description                                                               |
| ------------------------ | -------- | ------------------------------------------------------------------------- |
| `number_of_retries`      | `int32`  | The maximum number of retry attempts before failing.                      |
| `base_backoff`           | `string` | The initial backoff duration between retries (e.g., "1s").                |
| `max_backoff`            | `string` | The maximum backoff duration for exponential backoff (e.g., "30s").       |

### Circuit Breaker Fields

| Field                    | Type     | Description                                                               |
| ------------------------ | -------- | ------------------------------------------------------------------------- |
| `consecutive_failures`   | `int32`  | The number of consecutive failures that causes the circuit to open.       |
| `open_duration`          | `string` | How long the circuit remains open before trying to recover (e.g., "10s"). |

### Configuration Snippet

```yaml
upstream_services:
  - name: "unstable-service"
    resilience:
      retry_policy:
        number_of_retries: 3
        base_backoff: "1s"
      circuit_breaker:
        consecutive_failures: 5
        open_duration: "5s"
    http_service:
      address: "https://unstable.example.com"
```

## Use Case

If an upstream service starts failing, continuing to send requests wastes resources and slows down your server. A retry policy will attempt to recover from transient failures automatically. A circuit breaker will detect consistent failure rates and "open", immediately failing subsequent requests locally for a set duration (`open_duration`), giving the upstream service time to recover.

## Public API Example

When the circuit is open, MCP Any will return an error indicating the service is unavailable, without attempting to contact the upstream.
