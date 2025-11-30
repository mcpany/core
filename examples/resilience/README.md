# Resilience Example

This example demonstrates how to configure resilience features, such as circuit breakers and retry policies, for an upstream HTTP service.

## Configuration

The `config.yaml` file in this directory configures an HTTP service with a circuit breaker and a retry policy.

### Circuit Breaker

The circuit breaker is configured to open if the failure rate exceeds 60% over a minimum of 3 requests. When the circuit is open, all requests to the service will fail fast for a duration of 5 seconds. After 5 seconds, the circuit will transition to a half-open state and allow 2 requests to pass through. If these requests are successful, the circuit will close and normal operation will resume.

### Retry Policy

The retry policy is configured to retry failed requests up to 3 times with an exponential backoff. The base backoff is 1 second and the maximum backoff is 5 seconds.

## Running the Example

To run this example, you will need to start the MCP Any server with the `config.yaml` file in this directory.

```bash
make run ARGS="--config-paths examples/resilience/config.yaml"
```

You will also need to start a local HTTP server that simulates a failing service. You can use the `http-echo-server` in the `docker` directory for this purpose.

```bash
docker compose up --build http-echo-server
```

Once the servers are running, you can send requests to the `resilient-http-service` and observe the behavior of the circuit breaker and retry policy.
