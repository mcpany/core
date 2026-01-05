# Vector Database Upstream

The Vector Database Upstream allows the MCP Any server to connect to vector databases for RAG (Retrieval-Augmented Generation) workflows.

## Configuration

The vector provider is configured as an `upstream_service` with the `vector_service` block.

### Supported Providers

Currently, **Pinecone** is the supported vector database provider.

### Pinecone Configuration

Configure the `pinecone` block within `vector_service`.

| Field         | Type     | Description                                                                 |
| ------------- | -------- | --------------------------------------------------------------------------- |
| `api_key`     | `string` | Your Pinecone API Key.                                                      |
| `environment` | `string` | The environment (e.g., `us-west1-gcp`). Optional for serverless indexes.    |
| `index_name`  | `string` | The name of your Pinecone index.                                            |
| `project_id`  | `string` | The project ID. Optional if `host` is provided.                             |
| `host`        | `string` | The full host URL (e.g., `https://index-project.svc.pinecone.io`).          |

### Example Configuration

```yaml
upstream_services:
  - name: "knowledge-base"
    vector_service:
      pinecone:
        api_key: "${PINECONE_API_KEY}"
        index_name: "my-index"
        host: "https://my-index-12345.svc.pinecone.io"
```

## Tools Exposed

When configured, the vector service automatically exposes standard vector operations as tools:

- `query`: Search for similar vectors.
- `upsert`: Insert or update vectors.
- `delete`: Delete vectors.
- `describe_index_stats`: Get statistics about the index.

## Use Case

Connect an LLM to a vector database to allow it to semantic search over your documents.
