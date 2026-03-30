# Design Doc: Teammate Mailbox Integrity Guard (TMIG)
**Status:** Draft
**Created:** 2026-07-12

## 1. Context and Scope
With the rise of horizontal "Agent Teams" (e.g., Claude Code), agents now coordinate via peer-to-peer mailbox systems. These mailboxes enable agents to claim tasks, share state, and delegate sub-problems. However, current implementations often rely on unauthenticated filesystem writes (git-based) or implicit local trust, creating a massive attack surface for lateral movement and intent-hijacking.

TMIG provides a secure mediation layer for inter-agent mailbox coordination. It ensures that every message in a teammate's inbox is cryptographically signed, semantically validated against the mission root, and explicitly authorized by the Team Lead or a security quorum.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Mandate hardware-attested signatures for all inter-agent coordination messages.
    *   Perform real-time semantic analysis to ensure teammate requests align with the parent mission root.
    *   Implement "Auth-before-Read" for all teammate mailboxes.
    *   Provide a tamper-proof audit trail of inter-agent task handoffs.
*   **Non-Goals:**
    *   Replacing the underlying transport (e.g., it will still support Git-based mailboxes via the RAGL adapter).
    *   Managing the content of the agent's internal reasoning (monologue), only the coordination messages.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Delegate a security-sensitive task (e.g., API key rotation) to a subagent without risking lateral movement into host SSH keys.
*   **The Happy Path (Tasks):**
    1.  Team Lead agent generates a task card for the "Security Specialist" subagent.
    2.  Team Lead signs the task card using its MCP Any hardware-attested identity.
    3.  TMIG intercepts the mailbox write, validating the Team Lead's signature and the task's alignment with the mission root.
    4.  Subagent attempts to read its mailbox.
    5.  TMIG verifies the subagent's identity and provides the decrypted, validated task card.
    6.  Subagent attempts to delegate a "shadow task" back to the Lead.
    7.  TMIG detects the lack of parent authorization and blocks the message.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        A[Team Lead Agent] -->|Signed Message| B[TMIG Middleware]
        B -->|Validation & Quorum| C[Encrypted Mailbox Shard]
        D[Teammate Agent] -->|Identity Auth| B
        B -->|Authenticated Read| C
    ```
*   **APIs / Interfaces:**
    *   `POST /mailbox/send`: Secure endpoint for signing and sending inter-agent messages.
    *   `GET /mailbox/receive`: Authenticated retrieval of messages with integrity verification.
    *   `POST /mailbox/attest`: Manual or autonomous quorum attestation for high-risk coordination.
*   **Data Storage/State:**
    *   State is managed in the **Shared KV Store (Blackboard)** using **Intent-Sealed Shards**.

## 5. Alternatives Considered
*   **Plain Git-Locking**: Rejected due to high latency and lack of cryptographic identity validation.
*   **Centralized Orchestration Only**: Rejected because it creates a bottleneck for horizontal swarms and doesn't support the "Agent Teams" peer-to-peer paradigm.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** TMIG enforces per-message attestation. Even if one specialist agent is compromised, it cannot send instructions to others without a valid signature and mission-alignment.
*   **Observability:** All mailbox interactions are logged in the **Teammate Mailbox Monitor** with real-time alerts for blocked injections.

## 7. Evolutionary Changelog
*   **2026-07-12:** Initial Document Creation.
