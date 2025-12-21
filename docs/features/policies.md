# Policies

MCP Any provides powerful policies to control traffic, enhance performance, and ensure resilience.

## Available Policies

*   **Caching**: Cache tool responses to reduce latency and upstream load.
*   **Rate Limiting**: Limit the number of requests to upstream services (Memory & Redis).
*   **Resilience**:
    *   **Circuit Breakers**: Prevent cascading failures.
    *   **Retries**: Automatically retry failed requests.

## Configuration

Policies can be applied globally or per service.

See `docs/reference/configuration.md` for detailed configuration options.
