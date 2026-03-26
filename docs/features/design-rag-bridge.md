# Design Doc: OpenAI-Compatible RAG Bridge
**Status:** Draft
**Created:** 2026-07-08

## 1. Context and Scope
As agent ecosystems standardize, the need for interoperability between different RAG (Retrieval-Augmented Generation) frameworks and model providers has become critical. OpenAI's `/v1/embeddings` and `/v1/models` APIs have emerged as de facto standards. MCP Any needs to act as a universal proxy that allows any OpenAI-compatible client (like OpenClaw specialists) to discover and utilize embedding models and reasoning tools hosted across a heterogeneous MCP mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a standardized `/v1/embeddings` endpoint that proxies requests to MCP-based embedding tools.
    * Implement `/v1/models` to allow discovery of available MCP tools as "models" for reasoning.
    * Ensure seamless integration with OpenAI-native RAG clients.
    * Maintain mission-root security and intent-binding during RAG operations.
* **Non-Goals:**
    * Implementing a custom embedding model (we proxy to existing tools).
    * Providing long-term vector storage (this is handled by specialized MCP servers).

## 3. Critical User Journey (CUJ)
* **User Persona:** RAG System Architect
* **Primary Goal:** Use a standard OpenAI client library to generate embeddings from a local MCP-hosted model.
* **The Happy Path (Tasks):**
    1. The architect configures an OpenAI client to point to the MCP Any RAG Bridge URL.
    2. The client calls `/v1/models` to list available embedding tools.
    3. The client sends a POST request to `/v1/embeddings` with a text snippet.
    4. The RAG Bridge identifies the appropriate MCP tool, validates the request against the active mission intent, and proxies the call.
    5. The MCP tool returns the vector, which the bridge re-formats into a standard OpenAI response.
    6. The client successfully ingests the embedding vector for its local retrieval logic.

## 4. Design & Architecture
* **System Flow:**
    * External Client -> RAG Bridge (HTTP) -> Intent Validator -> Tool Registry -> MCP Server.
* **APIs / Interfaces:**
    * `POST /v1/embeddings`: Maps `model` parameter to MCP `tool_name`.
    * `GET /v1/models`: Filters and returns MCP tools with `capability:embedding`.
* **Data Storage/State:** Transient request-response mapping; uses existing Tool Registry for capability discovery.

## 5. Alternatives Considered
* **Framework-Specific Adapters**: Rejected due to the high maintenance cost of supporting 10+ different RAG SDKs.
* **Direct Model Access**: Rejected because it bypasses the Zero-Trust security and mission-root pinning provided by MCP Any.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All proxy requests must carry a hardware-attested mission token.
* **Observability:** Embeddings requests are logged in the Swarm Observability Hub for performance and cost tracking.

## 7. Evolutionary Changelog
* **2026-07-08:** Initial Document Creation.
