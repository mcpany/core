# Milvus Vector Database

The `milvus` vector database provider allows you to connect to a [Milvus](https://milvus.io/) instance or Zilliz Cloud and expose it as a set of MCP tools.

## Configuration

To use Milvus, you need to configure it in your `config.yaml` file under the `vector_service` section.

```yaml
upstream_services:
  - name: my-milvus
    vector_service:
      milvus:
        address: "localhost:19530" # The address of your Milvus server
        collection_name: "my_collection" # The name of the collection to use
        database_name: "default" # Optional, defaults to "default"
        username: "root" # Optional
        password: "Milvus" # Optional
        api_key: "your-api-key" # Optional, for Zilliz Cloud
        use_tls: false # Optional
```

## Tools

When a Milvus service is registered, the following tools are automatically available:

### `query_vectors`

Searches for vectors similar to the provided query vector.

*   **Input**:
    *   `vector` (array of numbers): The query vector.
    *   `top_k` (integer): The number of results to return (default: 10).
    *   `filter` (object): Optional metadata filter (e.g., `{"category": "news"}`).
    *   `namespace` (string): Optional partition name.

*   **Output**:
    *   `matches` (array): List of matching vectors with IDs and scores.

### `upsert_vectors`

Inserts or updates vectors in the collection.

*   **Input**:
    *   `vectors` (array of objects): List of vectors to upsert. Each object must have:
        *   `id` (string): Unique identifier.
        *   `values` (array of numbers): The vector embedding.
        *   `metadata` (object): Optional metadata fields.
    *   `namespace` (string): Optional partition name.

*   **Output**:
    *   `upserted_count` (integer): Number of vectors processed.

### `delete_vectors`

Deletes vectors from the collection.

*   **Input**:
    *   `ids` (array of strings): List of IDs to delete.
    *   `filter` (object): Optional metadata filter to delete by query.
    *   `namespace` (string): Optional partition name.

*   **Output**:
    *   `success` (boolean): Whether the operation was successful.

### `describe_index_stats`

Retrieves statistics about the collection.

*   **Input**:
    *   `filter` (object): Optional filter.

*   **Output**:
    *   `stats` (object): Collection statistics.

## Prerequisites

*   A running Milvus instance (v2.3+ recommended).
*   A collection created with appropriate schema (ID field, vector field, and metadata fields).
