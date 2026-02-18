# Caching

MCP Any provides caching capabilities to improve performance and reduce costs by avoiding redundant tool executions.

## 1. Standard Caching (Exact Match)

Standard caching works by hashing the tool inputs (arguments) and storing the result in an in-memory cache. If the same tool is called with the exact same arguments, the cached result is returned instantly.

### Implementation Details
*   **Storage**: In-Memory (GoCache).
*   **Key Generation**: FNV-1a 128-bit hash of the normalized JSON arguments.
*   **TTL**: Configurable per tool or service. Default is 5 minutes.
*   **Eviction**: Least Recently Used (LRU) based on memory constraints (GoCache default).

### Configuration
You can configure caching in your `config.yaml` for specific tools or services.

```yaml
upstream_services:
  - name: my-service
    config:
      cache:
        enabled: true
        ttl: 300s # 5 minutes
```

> **Note**: Currently, standard caching is **In-Memory only**. Distributed caching (Redis) and persistent disk caching for standard exact matches are planned for a future release (see [Roadmap](../../../roadmap.md), Priority P5).

## 2. Semantic Caching (Similarity Match)

Semantic caching uses vector embeddings to find "similar" requests. This is useful for LLM-based tools or fuzzy searches where the input text might vary slightly but the intent is the same.

### Implementation Details
*   **Provider**: Supports OpenAI, Ollama, and HTTP-based embedding providers.
*   **Vector Store**:
    *   **In-Memory**: Simple, non-persistent.
    *   **SQLite**: Persistent local file.
    *   **Postgres**: Persistent relational database (requires `pgvector` extension).
*   **Threshold**: Configurable similarity threshold (0.0 to 1.0).

### Configuration

To enable semantic caching, set the strategy to `semantic` and configure the provider.

```yaml
upstream_services:
  - name: my-llm-service
    config:
      cache:
        enabled: true
        strategy: "semantic"
        ttl: 3600s
        semantic_config:
          similarity_threshold: 0.95
          persistence_path: "data/vectors.db" # Optional: for SQLite persistence
          openai:
            model: "text-embedding-3-small"
            # api_key: loaded from secrets
```

## Cache Control

Clients can control caching behavior using the `X-Cache-Control` header or MCP request metadata.
*   `no-cache`: Bypasses the cache (force miss).
*   `no-store`: Skips writing to the cache.
