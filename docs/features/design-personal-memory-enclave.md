# Design Doc: Personal Memory Enclave (PME) Manager
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents evolve into long-term teammates, they require access to historical context (episodic memory) to maintain continuity. However, storing this context in centralized cloud databases exposes sensitive user data to model providers and infrastructure operators.

The Personal Memory Enclave (PME) Manager provides a hardware-encrypted, user-local storage solution for agent memory. It ensures that sensitive context remains under the user's physical control while still being searchable and useful for the agent.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide hardware-encrypted (TPM/Secure Enclave) storage for agent episodic memory.
    * Enable local similarity search (Vector DB) without exposing raw text to the LLM.
    * Support "Zero-Exposure Retrieval" where only anonymized snippets or similarity scores are sent to the model.
* **Non-Goals:**
    * Replacing the primary Shared KV Store (Blackboard) for active task state.
    * Providing cloud-based backup/sync (must be done via encrypted P2P tunnels).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Executive
* **Primary Goal:** Use an agent to summarize 6 months of private emails without the summary process exposing the raw emails to the cloud provider's training set.
* **The Happy Path (Tasks):**
    1. The user enables PME in the MCP Any dashboard.
    2. The agent "ingests" local files/emails, which are chunked and embedded locally using a small local model (e.g., FastEmbed).
    3. The embeddings and raw text are stored in a TPM-locked local SQLite/Vector database.
    4. When a summary is requested, the agent queries the PME Manager via a capability token.
    5. The PME Manager performs local retrieval and provides a "Privacy-Preserving Context Window" to the agent.

## 4. Design & Architecture
* **System Flow:**
    [Agent] <--> [PME Middleware] <--> [Local Vector DB (SQLite-vec)] <--> [Hardware Enclave (TPM)]
* **APIs / Interfaces:**
    * `pme.Search(query_embedding, top_k)` -> Returns encrypted/anonymized fragments.
    * `pme.Store(text_chunks, metadata)` -> Encrypts and indexes chunks.
* **Data Storage/State:**
    * SQLite-based vector store located in `$USER_HOME/.mcpany/pme/`.
    * Encryption keys derived from the system TPM.

## 5. Alternatives Considered
* **Homomorphic Encryption:** Rejected due to extreme computational overhead for LLM-scale context.
* **Simple Local Storage:** Rejected as it doesn't protect against host-level malware (TPM-binding is required).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Access to PME shards is gated by mission-root hardware leases.
* **Observability:** Logs will track retrieval volume and latency but never the content of memory shards.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
