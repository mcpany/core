# Design Doc: Differential Privacy Shard (DPS) Manager
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of "Context-Stitching" and "Side-Channel Shard Leakage," logical isolation of RAG shards is no longer sufficient. Specialized subagents can often "probe" the boundaries of shared teammate shards to exfiltrate mission-root constraints or user PII that was inadvertently included in the metadata.

MCP Any must move to an active state management model where local RAG shards are semantically sanitized using differential privacy techniques before being exposed to the teammate mesh.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement real-time differential privacy scrubbing for local RAG metadata.
    * Provide a semantic sanitization layer for the Universal Multimodal Memory Bus (UMMB).
    * Prevent "Shard-Probing" attacks via high-entropy reasoning requests.
    * Ensure that "Intent Residue" cannot be reconstructed from fragmented shard access.
* **Non-Goals:**
    * Modifying the raw vector embeddings themselves.
    * Replacing the primary vector database (e.g., Pinecone, Milvus).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Share a code-base RAG shard with 3 subagents without exposing internal `.env` metadata or private developer reasoning.
* **The Happy Path (Tasks):**
    1. The agent requests a context shard mount via the Shard Manager.
    2. DPS Manager intercepts the shard metadata stream.
    3. DPS Manager performs semantic analysis to identify high-entropy PII or intent fragments.
    4. DPS Manager applies Laplacian noise to metadata timestamps and redacts sensitive keys.
    5. The "Diff-Private" shard is mounted to the teammate mailbox.
    6. Subagent reasoning traces are monitored for "Shard-Boundary Probing" anomalies.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Vector DB] -->|Raw Shard| B(DPS Manager)
        B -->|Semantic Analysis| C[Reasoning-Aware Redaction Hub]
        C -->|Scrubbed Fragments| B
        B -->|Noise Injection| D[Differential Privacy Engine]
        D -->|Sanitized Shard| E[Teammate Mailbox]
    ```
* **APIs / Interfaces:**
    * `POST /v1/context/shards/sanitize`: Endpoint for semantic metadata scrubbing.
    * `GET /v1/security/privacy-score`: Real-time audit of shard entropy levels.
* **Data Storage/State:**
    * Privacy policies are stored as hardware-locked manifests (HAMM-compliant).

## 5. Alternatives Considered
* **Static Redaction:** Rejected because it cannot handle the novel instruction-injection patterns found in "Deceptive Context" markdown.
* **Full Shard Encryption:** Rejected because it prevents agents from performing necessary similarity searches on the metadata.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All sanitization rules must be hardware-attested. A failure in the DPS Engine must trigger an immediate "Shard Unmount" signal.
* **Observability:** integrated with the Visual Attention Dashboard to show users which metadata fragments were redacted.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
