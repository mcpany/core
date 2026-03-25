# Design Doc: CRDT-Native Mailbox Sharding
**Status:** Draft
**Created:** 2026-03-25

## 1. Context and Scope
The transition toward horizontal "Agent Teams" (e.g., Claude Code) has exposed a critical bottleneck in teammate coordination. Current models often rely on synchronous locks for a shared task list (Mailbox), leading to 2s+ coordination stalls as teams scale. As agents move from linear sessions to parallel teammate meshes, the infrastructure must provide a non-blocking, framework-neutral coordination layer.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a non-blocking, sharded "Mailbox" for parallel teammates.
    * Use Conflict-Free Replicated Data Types (CRDTs) to ensure state convergence without global locks.
    * Support "Auth-before-Claim" task integrity using hardware-attested mesh tokens.
    * Maintain sub-millisecond task claiming and delegation latency.
* **Non-Goals:**
    * Managing the internal reasoning monologues of individual agents.
    * Providing long-term archival of task history (handled by the Blackboard).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate 10 parallel teammates (Claude, OpenClaw, AutoGen) to execute a complex task without coordination deadlocks or performance stalls.
* **The Happy Path (Tasks):**
    1. The Mission-Root agent initializes a sharded Mailbox for the team.
    2. Teammates asynchronously claim tasks from the shared task list via CRDT-native updates.
    3. Conflicts (e.g., two agents claiming the same task) are resolved locally using LWW (Last-Write-Wins) logic and hardware-attested identity priority.
    4. Task state converges across all teammates without any agent waiting for a global lock.
    5. The mission-root agent synthesizes the results from the converged mailbox state.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        T1[Teammate 1] -->|Claim Task| MS[Mailbox Shard]
        T2[Teammate 2] -->|Update Status| MS
        MS -->|CRDT Sync| Mesh[Global Teammate Mesh]
        Mesh -->|Converged State| MR[Mission Root]
        MR -->|Finalize| BB[Blackboard/SQLite]
    ```
* **APIs / Interfaces:**
    * `POST /v1/mailbox/init`: Initialize a sharded mailbox session.
    * `POST /v1/mailbox/claim`: Claim a task using a CRDT-op with hardware-attested token.
    * `GET /v1/mailbox/state`: Retrieve the current converged mailbox state.
* **Data Storage/State:**
    * Active state is held in an in-memory Op-based CRDT (LWW-Element-Set).
    * Hardware-attested identity tokens (SMI) are used for conflict resolution priority.

## 5. Alternatives Considered
* **Centralized Redis/SQL Locks:** Rejected due to the "Mailbox Lock" congestion observed in high-density swarms.
* **Hierarchical Supervisor Delegation:** Rejected as it recreates the supervisor bottleneck for horizontal teams.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All mailbox requests must be signed with a hardware-attested Mesh Token.
* **Observability:** Real-time tracking of "Coordination Stall" and "Conflict Resolution" events.

## 7. Evolutionary Changelog
* **2026-03-25:** Initial Document Creation.
