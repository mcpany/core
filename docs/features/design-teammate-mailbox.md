# Design Doc: Standardized Teammate Mailbox (STM)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The transition from hierarchical subagent models to horizontal "Agent Teams" (e.g., Claude Code) has exposed critical performance bottlenecks in traditional centralized locking mechanisms. When multiple agents attempt to claim tasks from a shared list simultaneously, git-based or database locks often lead to "Cognitive Stall," where agents wait seconds for a coordination signal. MCP Any must provide a high-performance, sharded coordination bus that allows parallel teammates to synchronize state and claim tasks with sub-millisecond latency.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a sharded, peer-to-peer messaging protocol for inter-agent coordination.
    * Utilize Conflict-Free Replicated Data Types (CRDTs) to allow non-blocking updates to the shared task list.
    * Provide "Intent-Bound" mailbox isolation to prevent cross-mission state pollution.
    * Support hardware-attested identity verification for all mailbox participants.
* **Non-Goals:**
    * Replacing existing persistent storage like the Shared KV Store (STM is for transient coordination).
    * Managing the execution of agent tasks (it only manages the coordination of claiming them).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator.
* **Primary Goal:** Coordinate 5 parallel agents to refactor a codebase without hitting coordination deadlocks.
* **The Happy Path (Tasks):**
    1. Orchestrator spawns 5 "Teammate" agents.
    2. Agents initialize their local `STM Mailbox` via MCP Any.
    3. Agent A discovers a bug and adds it to the `Shared Task List` (CRDT-based).
    4. Agents B and C both see the new task; Agent B claims it by marking it in its local CRDT shard.
    5. The claim propagates instantly to Agent C's shard; Agent C automatically pivots to the next available task.
    6. All coordination occurs without a global "Task Lock" being held.

## 4. Design & Architecture
* **System Flow:**
    `Agent` <-> `STM Shard` <-> `Gossip Protocol` <-> `Teammate Shards`
* **APIs / Interfaces:**
    * `MailboxService`: `PostMessage(teammateID, payload)`, `Subscribe(topic)`
    * `TaskArbiter`: `ClaimTask(taskID)`, `ReleaseTask(taskID)`
* **Data Storage/State:**
    * In-memory CRDT shards for active coordination state, periodically snapshotted to the Durable Mission Continuity Provider.

## 5. Alternatives Considered
* **Centralized Redis Lock**: Rejected due to high latency in local-first, distributed environments and single-point-of-failure risks.
* **Git-based coordination**: Rejected due to the 1s+ latency of disk I/O and git operations, causing "Cognitive Stall".

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All messages are cryptographically signed by the hardware-attested agent identity. Intent-Bound isolation ensures agents only see messages relevant to their authorized mission branch.
* **Observability**: The `Teammate Mailbox Inspector` provides a real-time visual trace of coordination messages and task claims.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
