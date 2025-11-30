# Resilience

MCP Any provides resilience features to help you build robust and reliable services.

## Retry Policy

You can configure a retry policy for any upstream service to automatically retry failed requests. This can help to improve the reliability of your services by making them more resilient to transient failures.

### Configuration

To configure a retry policy, you can add a `retry_policy` to the `resilience` section of your service configuration. The following options are available:

*   `number_of_retries`: The number of times to retry a failed request.
*   `base_backoff`: The base duration for the backoff between retries. The backoff duration is calculated using the formula `base_backoff * (2 ** (retry_number - 1))`.
*   `max_backoff`: The maximum duration for the backoff.

### Example

The following example shows how to configure a retry policy that retries a failed request up to 3 times with an exponential backoff:

```yaml
upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "https://api.example.com"
      calls:
        - operationId: "get_user"
          description: "Get user by ID"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/{userId}"
    resilience:
      retry_policy:
        number_of_retries: 3
        base_backoff: 10ms
        max_backoff: 1s
```
