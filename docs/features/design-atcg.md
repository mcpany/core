# Design Doc: Agent Team Coordination Guard (ATCG)
**Status:** Draft
**Created:** 2026-07-13

## 1. Context and Scope
The general availability of "Agent Teams" in frameworks like Claude Code has shifted the coordination paradigm from hierarchical delegation to horizontal, peer-to-peer collaboration. However, current coordination mechanisms (git-based locks, shared mailboxes) often lack cryptographic intent validation, making them vulnerable to "Teammate Impersonation" and "State Injection" if one specialist agent is compromised.

The **Agent Team Coordination Guard (ATCG)** is a high-speed security service that provides intent-bound mailbox shards for horizontal meshes. It ensures that every teammate-to-teammate coordination event is hardware-attested and linked to a verified mission role.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide cryptographically isolated, task-bound mailbox shards.
    * Enforce "Auth-before-Claim" for teammate task acquisition.
    * Perform real-time intent validation for all inter-agent messages.
    * Support lock-free, CRDT-based state synchronization for parallel teams.
* **Non-Goals:**
    * Managing hierarchical subagent handoffs (handled by A2A Messaging Hub).
    * Orchestrating the high-level task breakdown (handled by the "Team Lead" agent).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate a team of 3 agents (Coder, Auditor, Deployer) on a shared repository without allowing the "Coder" to inject malicious tasks into the "Deployer"'s mailbox.
* **The Happy Path (Tasks):**
    1. Team Lead agent creates a "Team Mesh Session" in MCP Any.
    2. ATCG generates session-bound, hardware-attested "Mission Tokens" for each teammate role.
    3. Teammates authenticate with ATCG using their tokens.
    4. Coder publishes a "PR Review" task to the shared mailbox.
    5. ATCG validates the message intent against the Coder's role.
    6. Auditor claims the task; ATCG verifies the Auditor's hardware token before granting access.
    7. All coordination is logged and anchored to the Mission Root.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        TeammateA[Teammate A] -->|Signed Message| ATCG[ATCG]
        ATCG -->|Intent/Role Check| Shard[Mailbox Shard]
        Shard -->|Validated State| TeammateB[Teammate B]
        ATCG -->|Attestation| TPM[Host TPM]
    ```
* **APIs / Interfaces:**
    * `mesh/mailbox/publish`: Publish a message to a task-bound shard.
    * `mesh/mailbox/claim`: Securely acquire a task from the shard.
    * `mesh/mailbox/sync`: Conflict-free state synchronization for the teammate mesh.
* **Data Storage/State:**
    * Granular shards stored in an in-memory CRDT (Conflict-Free Replicated Data Type) store for lock-free performance.

## 5. Alternatives Considered
* **Git-Based Locking:** Rejected due to high latency and susceptibility to filesystem-level racing/tampering.
* **Global coordination locks:** Rejected because they cause "Coordination Stall" as teams scale horizontally.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All mesh coordination is cryptographically bound to the user's hardware-attested approval.
* **Observability:** Integrated with the "Service Mesh Topology Monitor."

## 7. Evolutionary Changelog
* **2026-07-13:** Initial Document Creation.
