# Resilience

MCP Any provides resilience features to help you build robust and reliable services. These features include circuit breakers and retry policies.

## Circuit Breaker

The circuit breaker is a state machine that monitors the health of an upstream service and prevents you from making requests to it when it is unhealthy. This can help to prevent cascading failures and reduce the load on a failing service.

The circuit breaker has three states:

*   **Closed:** The circuit is closed and all requests are passed through to the upstream service.
*   **Open:** The circuit is open and all requests to the upstream service fail fast.
*   **Half-Open:** The circuit is in a transitional state and a limited number of requests are allowed to pass through to the upstream service to test for recovery.

You can configure the circuit breaker with the following parameters:

*   `failureRateThreshold`: The failure rate threshold at which the circuit will open.
*   `openDuration`: The duration the circuit will remain open before transitioning to the half-open state.
*   `halfOpenRequests`: The number of requests to allow in the half-open state to test for recovery.

Here is an example of how to configure a circuit breaker for an HTTP service:

```yaml
upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "https://api.example.com"
      # ...
    resilience:
      circuitBreaker:
        failureRateThreshold: 0.6
        openDuration: "5s"
        halfOpenRequests: 2
```

## Retry Policy

The retry policy allows you to automatically retry failed requests to an upstream service. This can help to improve the reliability of your service by masking transient failures.

You can configure the retry policy with the following parameters:

*   `numberOfRetries`: The number of times to retry a failed request.
*   `baseBackoff`: The base duration for the backoff between retries.
*   `maxBackoff`: The maximum duration for the backoff.

Here is an example of how to configure a retry policy for an HTTP service:

```yaml
upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "https://api.example.com"
      # ...
    resilience:
      retryPolicy:
        numberOfRetries: 3
        baseBackoff: "1s"
        maxBackoff: "5s"
```
