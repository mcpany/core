# Design Doc: Kernel-Level Lock-Arbiter
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of "Agent Teams" in Claude Code, the limitation of mailbox-based locking for task coordination has become a critical performance bottleneck. High-density swarms suffer from "Mailbox Lock" congestion, where multiple agents waiting on filesystem-level locks enter prolonged wait cycles. The Kernel-Level Lock-Arbiter provides a high-performance, non-blocking alternative managed directly by the MCP Any gateway.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a lock-free task claiming system using Conflict-Free Replicated Data Types (CRDTs).
    * Provide sub-millisecond task synchronization across parallel teammates.
    * Support hardware-attested identity verification for all task claims.
    * Enable cross-framework teammate coordination (Claude Code, OpenClaw, AutoGen).
* **Non-Goals:**
    * Replacing the agents' internal task planning logic.
    * Providing general-purpose distributed locking (focused on Agent Task Lists).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate a team of 10+ specialist agents on a complex refactoring task without coordination stalls.
* **The Happy Path (Tasks):**
    1. The mission-root agent initializes a task list via the MCP Any Gateway.
    2.Teammates subscribe to the task-mesh via A2A-authenticated channels.
    3. An agent claims a task by issuing an atomic CRDT mutation request.
    4. The Lock-Arbiter validates the hardware-attested token and merges the mutation.
    5. The updated state is broadcast to all teammates instantly, allowing them to adjust their planning loops.

## 4. Design & Architecture
* **System Flow:**
    [Teammate A] <-> [A2A/mTLS] <-> [Lock-Arbiter (CRDT Logic)] <-> [Shared Task Mesh]
* **APIs / Interfaces:**
    * `ClaimTask(taskId, missionToken)` -> `ack/nack`
    * `YieldTask(taskId, reason)`
    * `SyncState()` -> `TaskMeshView`
* **Data Storage/State:** Sharded in-memory CRDT state with atomic logging to the Blackboard (SQLite).

## 5. Alternatives Considered
* **Distributed Redis Locking:** Rejected due to the overhead of external dependencies and network latency.
* **Kernel-Bound Named Pipes:** Effective for transport but lacks the state-merging capabilities of CRDTs.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All mutations must be signed by a hardware-attested mission-root token.
* **Observability:** Track "Time to Coordinate" (TTC) metrics in the Swarm Topology Monitor.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
