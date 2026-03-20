# Design Doc: Federated Vector Blackboard (FVB)

## Objective
Upgrade the existing "Shared KV Store" (Blackboard) into a multi-modal Federated Vector Blackboard (FVB). The FVB will act as the unified, cross-framework memory layer for heterogeneous agent swarms, supporting semantic search, multimodal embeddings, and Intent-Sealed Sharding for Zero-Trust data isolation.

## Background
Currently, agent frameworks utilize rudimentary string-based or localized KV stores for memory. This causes "Memory Fragmentation" when swarms operate across framework boundaries (e.g., Claude communicating with OpenClaw). As agents expand to multimodal inputs and complex RAG scenarios, a simple KV store cannot support the required semantic search capabilities. Moreover, storing shared state without intent-based access control leads to context dumping and exfiltration risks.

## High-Level Design
The FVB will replace the embedded SQLite Blackboard with a lightweight, embedded vector database (e.g., integrated Milvus Lite or Chroma architecture adapted for MCP Any). It will introduce a standardized embedding schema so all connected frameworks can push and query memory uniformly.

### Core Components
1.  **Vector Store Engine:** The underlying embedded storage layer optimized for high-speed similarity search on multi-dimensional vectors.
2.  **Cross-Framework Embedding Normalizer:** A middleware component that takes raw context (text, image descriptors, audio transcripts) and generates standardized embeddings (or uses a provided embedding model) before storing them in the FVB.
3.  **Intent-Sealed Sharding Manager:** A security layer that partitions the vector space. Memory shards are cryptographically bound to specific "Mission Roots." An agent can only query or write to a shard if it possesses a hardware-attested token proving its lineage to that specific mission.

## Detailed Design

### Standardized Memory Object
All memory entries will conform to a standardized schema:
- `id`: UUID.
- `mission_root_id`: The cryptographic hash of the parent mission.
- `agent_id`: The hardware-attested identity of the authoring agent.
- `timestamp`: Monotonic creation time.
- `type`: Enum (text, image, audio_transcript).
- `vector`: The N-dimensional embedding.
- `metadata`: JSON payload for exact-match filtering (e.g., tags, tool_name).
- `payload`: The actual raw data (or a reference pointer if too large).

### Intent-Sealed Queries
When an agent submits a query (e.g., "Find previous errors related to database connection"), the FVB will:
1.  Embed the query string.
2.  Validate the agent's current `mission_root_id`.
3.  Perform a vector similarity search *strictly partitioned* by the validated `mission_root_id`.
4.  Return the top-K results.

### Write Operations & Splicing Defense
Writes to the FVB must pass through the Semantic Integrity Bridge. The MMSE (Multi-Modal Sanitization Engine) will scan the raw payload for adversarial perturbations or embedded prompt injections before generating the final vector and storing the record.

## Security Considerations
- **Context Smuggling:** Subagents might try to encode malicious instructions into the `metadata` payload. The MMSE must sanitize all textual metadata fields.
- **Shard Collision:** Cryptographic hashes for `mission_root_id` must be collision-resistant (SHA-256 minimum) to prevent unauthorized agents from brute-forcing access to parallel shards.

## Rollout Strategy
1.  **Phase 1 (Alpha):** Implement the Vector Store Engine side-by-side with the legacy KV store. Support basic text-to-text semantic search.
2.  **Phase 2 (Beta):** Introduce the Intent-Sealed Sharding Manager. Migrate all internal MCP Any routing state to the FVB.
3.  **Phase 3 (GA):** Expose the FVB capability to external agents via the MCP protocol. Deprecate the legacy Shared KV Store.
