# Design Doc: Lock-Free Teammate Coordination (LFTC)
**Status:** Draft | In Review | Approved
**Created:** 2026-03-24

## 1. Context and Scope
As AI agents evolve from hierarchical subagents to horizontal "Agent Teams" (e.g., Claude Code), the coordination bottleneck shifts from the parent-child handoff to the teammate-to-teammate (T2T) interaction. Current models often rely on synchronous locks for a shared task list, leading to "Mailbox Lock" congestion as teams scale. LFTC aims to provide a high-performance, non-blocking coordination layer for horizontal swarms.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a non-blocking "Shared Task List" for parallel teammates.
    * Utilize Conflict-Free Replicated Data Types (CRDTs) to ensure state convergence without global locks.
    * Support "Auth-before-Claim" task integrity using hardware-attested mission tokens.
    * Enable sub-millisecond task claiming and delegation across heterogeneous agent frameworks.
* **Non-Goals:**
    * Replacing the primary reasoning engine of the agents.
    * Managing the internal memory of individual agents (this is handled by Context Sharding).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate a team of 5 specialist agents to refactor a large codebase without coordination deadlocks.
* **The Happy Path (Tasks):**
    1. The mission-root agent initializes a "Shared Task List" in the MCP Any LFTC hub.
    2. Five specialist teammates (Claude, OpenClaw, etc.) join the mesh and receive a sharded view of the task list.
    3. Teammates asynchronously "claim" tasks by providing their hardware-attested mission tokens.
    4. Task state (In-Progress, Completed) propagates across the mesh via CRDT sync without blocking other teammates.
    5. The mission-root agent monitors the sharded task list in real-time to synthesize the final result.

## 4. Design & Architecture
* **System Flow:**
    * Teammates interact with a "Virtual Inbox" (Shared Task List) hosted by MCP Any.
    * State is managed using an Op-based CRDT (e.g., LWW-Element-Set) for task status.
    * Task claiming requires a `Mission-Root-Token` validated against the hardware-attested session.
* **APIs / Interfaces:**
    * `InitializeTaskList(missionId, initialTasks)`
    * `ClaimTask(teammateId, taskId, missionToken)`
    * `UpdateTaskStatus(taskId, newStatus, proofOfWork)`
* **Data Storage/State:** Sharded, in-memory CRDT state with periodic persistence to the Blackboard (SQLite).

## 5. Alternatives Considered
* **Synchronous Redis/SQL Locks**: Rejected due to latency and the risk of "Mailbox Lock" congestion in high-density swarms.
* **Hierarchical Delegation (Subagents)**: Rejected as it doesn't support the emerging "Horizontal Team" paradigm where teammates must be autonomous.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** "Auth-before-Claim" ensures only teammates with a verified lineage can interact with the task list.
* **Observability:** Real-time task-claiming telemetry is exported to the Swarm Monitor.

## 7. Evolutionary Changelog
* **2026-03-24:** Initial Document Creation.
