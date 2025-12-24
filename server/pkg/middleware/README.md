# Semantic Cache Middleware

The Semantic Cache middleware provides semantic caching capabilities for MCP tools. Unlike exact matching, semantic caching uses embeddings (vector representations of text) to find semantically similar queries.

## Features

- **Semantic Matching**: Uses cosine similarity to find cached results for similar inputs.
- **Persistence**: Supports persisting cache entries to SQLite to survive server restarts.
- **In-Memory Speed**: Loads vectors into memory for fast search while keeping data persisted.
- **Configurable**: Adjustable similarity threshold and TTL.

## Configuration

To enable semantic caching for a tool or service, configure the `cache` section in your upstream service configuration.

```yaml
cache:
  is_enabled: true
  ttl: "1h"
  strategy: "semantic"
  semantic_config:
    provider: "openai"
    model: "text-embedding-3-small"
    similarity_threshold: 0.95
    api_key:
      plain_text: "sk-..."
```

## Storage

The semantic cache uses a SQLite database for persistence. By default, it uses the global `db_path` configured in `config.yaml` (default: `mcpany.db`).

The middleware automatically creates a `semantic_cache_entries` table in the database.

## Supported Providers

- **OpenAI**: Uses OpenAI's embeddings API.

## Implementation Details

- **Write-Through**: New entries are written to both memory and SQLite.
- **Load-On-Startup**: On server startup, valid (non-expired) entries are loaded from SQLite into memory.
- **Background Cleanup**: A background process periodically removes expired entries from the database.
