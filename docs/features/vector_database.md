# Vector Database Integration

MCP Any provides native support for connecting to Vector Databases, enabling AI agents to perform Semantic Search and RAG (Retrieval-Augmented Generation) workflows directly via MCP.

## Supported Providers

- **Pinecone**: Connect to Pinecone indexes.
- *More coming soon (Milvus, Weaviate).*

## Configuration

To enable a Vector Database upstream, add a `vector_service` entry to your configuration file.

### Pinecone Example

```yaml
upstream_services:
  - name: "pinecone-docs"
    vector_service:
      pinecone:
        api_key: "${PINECONE_API_KEY}"
        index_name: "documentation"
        project_id: "abc1234" # Optional for serverless
        environment: "us-west1-gcp" # Optional for serverless
        # OR specify full host if using serverless
        # host: "https://documentation-abc1234.svc.us-west1-gcp.pinecone.io"
```

## Available Tools

When a Vector Service is registered, the following tools are automatically exposed:

### 1. `query_vectors`

Searches the vector database for vectors similar to the provided query vector.

**Arguments:**
- `vector` (array of numbers): The query embedding vector.
- `top_k` (integer): Number of results to return (default: 10).
- `filter` (object): Metadata filter (provider specific).
- `namespace` (string): The namespace to query.

### 2. `upsert_vectors`

Inserts or updates vectors in the database.

**Arguments:**
- `vectors` (array of objects): List of vectors to upsert. Each object must have:
  - `id` (string): Unique ID of the vector.
  - `values` (array of numbers): The embedding vector.
  - `metadata` (object): Optional metadata.
- `namespace` (string): The namespace to upsert into.

### 3. `delete_vectors`

Deletes vectors from the database.

**Arguments:**
- `ids` (array of strings): List of IDs to delete.
- `namespace` (string): The namespace to delete from.
- `filter` (object): Metadata filter (if deleting by filter).
- `deleteAll` (boolean): If true, deletes all vectors in the namespace.

### 4. `describe_index_stats`

Returns statistics about the index, such as total vector count and dimension.

## Usage with LLMs

AI agents can use these tools to build RAG pipelines.

**Example User Query:** "Find documents about 'caching' in the 'documentation' index."

**Agent Action:**
1. The agent first generates an embedding for "caching" (using an embedding tool or internal capability).
2. The agent calls `pinecone-docs.query_vectors` with the generated vector.
3. The agent receives matching document chunks (stored in metadata).
4. The agent formulates a response based on the retrieved context.
