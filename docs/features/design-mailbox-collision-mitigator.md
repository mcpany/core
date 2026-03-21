# Design Doc: Mailbox Collision Mitigator
**Status:** Draft
**Created:** 2026-06-21

## 1. Context and Scope
With the introduction of "Agent Teams" in Claude Code and other multi-agent frameworks, horizontal coordination has become a primary performance bottleneck. Current implementations often rely on synchronous file-locking (e.g., in a shared `mailbox/` directory), which leads to significant "Mailbox Lock" contention as the swarm size grows beyond 3-5 agents.

The Mailbox Collision Mitigator (MCM) evolves MCP Any into a high-performance coordination bus. It replaces synchronous filesystem locks with a lock-free, CRDT-based arbitration layer, allowing parallel teammates to claim tasks and exchange messages with sub-100ms latency.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a Lock-Free Mesh Arbiter (LFMA) using Conflict-Free Replicated Data Types (CRDTs).
    * Support asynchronous task claiming and state synchronization for horizontal teammate meshes.
    * Provide a unified API for frameworks (Claude, OpenClaw, AutoGen) to exchange mailbox messages without blocking.
* **Non-Goals:**
    * Replacing the primary reasoning engine of the agents.
    * Managing long-term state persistence (this remains the role of the Blackboard).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate 10+ parallel subagents on a complex task (e.g., full-stack refactor) without coordination stalls.
* **The Happy Path (Tasks):**
    1. Parent agent posts 10 sub-tasks to the Shared Task List.
    2. 10 subagents simultaneously attempt to "claim" different tasks.
    3. LFMA resolves claims using deterministic CRDT logic, ensuring each task is assigned to exactly one agent without any agent waiting on a file lock.
    4. Subagents post status updates to their task-bound mailbox shards.
    5. Parent agent receives a real-time, non-blocking aggregate view of the swarm's progress.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A1[Agent 1] -- Claim Task A --> B{LFMA}
        A2[Agent 2] -- Claim Task B --> B
        B -- Merged State --> C[Shared Task List]
        B -- Resolved Conflict --> A1
        B -- Resolved Conflict --> A2
    ```
* **APIs / Interfaces:**
    * `POST /v1/mesh/claim`: Atomically claim a task ID using CRDT resolution.
    * `POST /v1/mesh/mailbox/send`: Asynchronously send a message to a teammate's shard.
* **Data Storage/State:**
    * In-memory CRDT state with periodic checkpoints to the SQLite Blackboard.

## 5. Alternatives Considered
* **Redis-based Locking**: Rejected for local-first use cases due to infrastructure overhead.
* **NATS/Message Queue**: Considered, but CRDTs provide better offline-first and decentralized resilience for swarm coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Every claim and message must be signed with the agent's hardware-attested session token (TLSB).
* **Observability:** Coordination latency and claim conflict rates are visualized in the "Mailbox Shard Monitor."

## 7. Evolutionary Changelog
* **2026-06-21:** Initial Document Creation.
