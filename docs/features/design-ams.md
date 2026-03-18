# Design Doc: Asynchronous Mailbox Sharding (AMS)
**Status:** Draft
**Created:** 2026-05-30

## 1. Context and Scope
Enterprise swarms using Claude Code "Agent Teams" are hitting a "Task List Contention" bottleneck. When multiple agents (5+) attempt to simultaneously read/write to the shared task list (Blackboard), reasoning latency increases significantly due to synchronous locking. MCP Any needs a non-blocking coordination mechanism to support horizontal scaling of parallel teams.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement granular, task-bound mailbox shards to eliminate global coordination locks.
    * Provide sub-millisecond state synchronization for parallel teammates.
    * Enable framework-neutral (Claude, OpenClaw, AutoGen) asynchronous coordination.
* **Non-Goals:**
    * Replacing the Shared KV Store for persistent, long-term state.
    * Managing the LLM's internal context window directly.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** Coordinate 10 parallel agents on a complex codebase refactor without "Cognitive Stall" from state contention.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes a mission with AMS enabled.
    2. MCP Any automatically creates addressable mailbox shards for every sub-task in the mission root.
    3. Parallel agents claim shards and synchronize their local state asynchronously via lock-free buffers.
    4. Mission root reconciles shards into the global blackboard upon sub-task completion.

## 4. Design & Architecture
* **System Flow:**
    * The T2T Encryption Bridge is upgraded with an "AMS Router."
    * Incoming messages are routed to specific Shard Buffers based on the `task_id` or `branch_id`.
    * A "Snapshot-and-Merge" background service reconciles shard deltas into the primary Blackboard.
* **APIs / Interfaces:**
    * `ams.v1.CreateShard(mission_id, task_id)`: Initializes a new isolated mailbox shard.
    * `ams.v1.PushDelta(shard_id, delta_object)`: Asynchronously pushes state changes to a shard.
* **Data Storage/State:**
    * In-memory lock-free ring buffers for high-speed messaging.
    * Shard-level SQLite journals for crash recovery.

## 5. Alternatives Considered
* **Optimistic Concurrency Control (OCC):** Rejected because high contention leads to excessive retries in reasoning loops, wasting tokens.
* **Centralized Redis Locking:** Rejected due to network latency overhead in local-first deployments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Shards are cryptographically isolated; an agent can only access shards it is explicitly authorized for via its mission token.
* **Observability:** Provides "Contention Heatmaps" to the UI to identify coordination bottlenecks.

## 7. Evolutionary Changelog
* **2026-05-30:** Initial Document Creation.
