# Design Doc: Multi-Host Teammate Synchronization (MHTS)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI agent frameworks like Claude Code are evolving from single-session subagents to collaborative horizontal teams. These teams coordinate through shared task lists and direct messaging. However, current implementations are often bound to a single host or session. MCP Any needs to provide a robust, multi-host coordination layer that ensures state consistency (shared task list) and secure communication (inbox isolation) across disparate agent environments.

## 2. Goals & Non-Goals
* **Goals:**
    * Facilitate synchronization of a shared task list across multiple MCP Any nodes.
    * Use Conflict-Free Replicated Data Types (CRDTs) to ensure non-blocking, eventual consistency.
    * Provide hardware-attested identity verification for teammates joining the swarm.
    * Enable cryptographically isolated "Inbox" fragments for teammate messaging.
* **Non-Goals:**
    * Providing a general-purpose real-time database.
    * Handling low-level network discovery (assumes AMT or similar transport).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Host Swarm Orchestrator
* **Primary Goal:** Coordinate a refactor project across a local Laptop and a remote Build Server using collaborative agents.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes a Team Session on the Laptop.
    2. Laptop node generates a hardware-attested Team Invite.
    3. Build Server node joins the team using the invite and verifies its identity via TPM.
    4. Both nodes synchronize the initial Shared Task List via MHTS.
    5. Laptop agent claims an "API Refactor" task; MHTS propagates the "Claimed" state to the Build Server.
    6. Build Server agent receives a message from the Laptop agent with API schema discoveries, isolated to their specific teammate inbox.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        subgraph Host A
            A1[Agent 1] --> M1[MHTS Provider]
            M1 --> C1[CRDT Task List]
        end
        subgraph Host B
            A2[Agent 2] --> M2[MHTS Provider]
            M2 --> C2[CRDT Task List]
        end
        M1 <==>|Hardware-Attested Sync| M2
    ```
* **APIs / Interfaces:**
    * `mhts.JoinTeam(inviteToken, hardwareProof) -> SessionID`
    * `mhts.ClaimTask(taskID) -> Status`
    * `mhts.PostTeammateMessage(targetAgentID, fragment) -> void`
* **Data Storage/State:**
    * **Task Graph (SQLite + CRDT):** Local persistent store for the replicated task list.
    * **Handshake Registry:** In-memory store of verified teammate identities and their hardware fingerprints.

## 5. Alternatives Considered
* **Centralized SQL Orchestrator:** Rejected because it introduces a single point of failure and coordination bottlenecks (locks).
* **Git-based Coordination:** Rejected due to high latency and lack of fine-grained, per-message isolation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All sync events require hardware-attested lineage proof. Inboxes are fragments-isolated to prevent lateral context probing.
* **Observability:** Integrated with the "Service Mesh Topology Monitor" and a new "Multi-Host Teammate Dashboard."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
