# Traffic Policy Management

MCP Any provides enterprise-grade traffic management features to ensure the stability and reliability of your upstream services.

![Traffic Policy Configuration](../screenshots/traffic-policy.png)

## Features

### Rate Limiting
Protect your upstream services from excessive load by configuring rate limits.
- **Requests Per Second (RPS)**: The average number of requests allowed per second.
- **Burst**: The maximum number of requests allowed in a short burst.

### Connection Pooling
Manage the connection resources used to communicate with upstream services.
- **Max Connections**: The maximum number of simultaneous connections.
- **Max Idle Connections**: The number of idle connections to keep open.
- **Idle Timeout**: How long a connection can remain idle before being closed.

### Resilience
Ensure your system handles failures gracefully.
- **Global Timeout**: The maximum duration for a request before it is cancelled.
- **Retry Policy**: Configure automatic retries for failed requests.
  - **Max Retries**: Number of retry attempts.
  - **Base/Max Backoff**: Delay between retries.
- **Circuit Breaker**: Fail fast when an upstream service is down to prevent cascading failures.
  - **Failure Threshold**: The percentage of failures that triggers the circuit breaker.
  - **Open Duration**: How long to wait before testing the connection again.
