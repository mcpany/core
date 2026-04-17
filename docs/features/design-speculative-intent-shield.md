# Design Doc: Speculative Intent Shield (SIS)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of "Predictive Co-reasoning" in agent frameworks (e.g., Gemini CLI), agents are increasingly executing tool calls speculatively before the primary reasoning path is fully committed. This shift drastically reduces latency but introduces a major security risk: "Speculative Shadowing." Speculative results can pollulte the Shared Blackboard or trigger side-effects before they are verified against the user's mission root.

MCP Any needs to provide a secure, transactional buffer for these speculative actions. The Speculative Intent Shield (SIS) ensures that all predictive execution remains isolated and reversible until cryptographic attestation is provided.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide isolated, ephemeral buffers for speculative tool results.
    * Prevent speculative data from reaching the persistent Shared Blackboard without attestation.
    * Enable atomic "Commit" or "Discard" operations for speculative branches.
    * Support hardware-attested validation of the predictive path.
* **Non-Goals:**
    * SIS will not perform the predictive reasoning itself; it is a governance layer for the results.
    * SIS will not prevent side-effects in external systems that do not support transactional rollback.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Performance Architect
* **Primary Goal:** Enable 2x faster agent coordination via predictive execution without risking state corruption.
* **The Happy Path (Tasks):**
    1. The parent agent signals a "Speculative Branch" via a UACO v3.8 header.
    2. MCP Any initializes a SIS Ephemeral Buffer for the session.
    3. The subagent executes a predictive tool call; results are routed to the SIS Buffer.
    4. The primary reasoning path completes and confirms the branch.
    5. The parent agent sends a "Commit" signal with a hardware-attested token.
    6. SIS merges the buffer into the persistent Shared Blackboard and marks the tool call as "Finalized."

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>MCP Any: Tool Call (Speculative=True)
        MCP Any->>SIS: Create Ephemeral Shard
        SIS->>Tool: Execute
        Tool-->>SIS: Result Data
        SIS-->>Agent: Probabilistic Result (Read-Only)
        Agent->>MCP Any: Commit Signal (Signed)
        MCP Any->>SIS: Verify & Merge
        SIS->>Blackboard: Persist State
    ```
* **APIs / Interfaces:**
    * `POST /v1/speculative/commit`: Finalizes a buffered branch.
    * `DELETE /v1/speculative/branch/{id}`: Discards a failed prediction.
* **Data Storage/State:** Speculative results are stored in memory-mapped `memfd` regions, isolated from the primary SQLite Blackboard.

## 5. Alternatives Considered
* **Application-Level Buffering**: Rejected because it relies on the agent framework's honesty. Governance must be enforced at the infrastructure layer.
* **Database Savepoints**: Rejected due to the overhead of long-running transactions in high-density horizontal meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Speculative branches are restricted to "Read-Only" access of the primary Blackboard to prevent leakage.
* **Observability:** SIS metrics will track "Prediction Accuracy" (Commit vs. Discard ratio) to help optimize swarm behavior.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
