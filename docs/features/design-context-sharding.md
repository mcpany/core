# Design Doc: Live Context Sharding Middleware
**Status:** Draft
**Created:** 2026-03-27

## 1. Context and Scope
As agent swarms grow in complexity, the monolithic transfer of context becomes a significant bottleneck, leading to "Token Tax" and "Cognitive Stall." OpenClaw's Live Context Sharding (LCS) addresses this by partitioning state into granular shards. MCP Any needs a middleware that can manage the lifecycle of these shards, allowing agents to mount/unmount sub-state on-demand.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Shard Manager" that tracks addressable context fragments.
    * Provide APIs for agents to dynamically load/unload shards based on current "Intent Scope."
    * Integrate with the BSH Gateway for zero-copy shard transfer.
    * Maintain a "Virtual Context Map" that ensures consistency across sharded state.
* **Non-Goals:**
    * Managing the persistence of shards (handled by the Shared KV Store).
    * Providing the logic for shard partitioning (handled by the Agent Framework).

## 3. Critical User Journey (CUJ)
* **User Persona:** Deep Swarm Orchestrator
* **Primary Goal:** Minimize token usage for a subagent performing a specialized code refactoring task within a large codebase.
* **The Happy Path (Tasks):**
    1. Parent Agent identifies the "Refactoring" intent.
    2. Shard Manager identifies the relevant code shards and dependency context.
    3. Subagent is initialized with only the "Active Shard" mounted.
    4. Subagent requests an "Ad-hoc Mount" for a related library shard when needed.
    5. Shard Manager validates the mount request against the Intent Chain and provides sub-millisecond access.

## 4. Design & Architecture
* **System Flow:**
    `Intent` -> `Shard Resolver` -> `Virtual Context Map` -> `BSH Mount` -> `Agent Runtime`
* **APIs / Interfaces:**
    * `ShardManager` Interface: `Mount(shardID string, agentID string) error`, `Unmount(shardID string) error`
    * `ContextMap`: A real-time registry of active mounts and shard lineage.
* **Data Storage/State:**
    * Shards are stored as binary blobs in the BSH State Buffer, indexed by the Shard Manager.

## 5. Alternatives Considered
* **Monolithic BSH Transfer**: Rejected due to high memory overhead and latency in deep chains.
* **Framework-Side Sharding**: Rejected because it prevents cross-framework state sharing (e.g., OpenClaw subagent using an AutoGen shard).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Shard mounts must be authorized against the "Relational PoI" of the agent session.
* **Observability:** Shard lifecycle events (mount/unmount/hit/miss) are logged to the "Binary Handoff Performance Monitor."

## 7. Evolutionary Changelog
* **2026-03-27:** Initial Document Creation.
### Update: 2026-03-28 - Atomic State Rollbacks
**Context:** Today's research on OpenClaw's Atomic State Rollbacks (ASR) confirms the need for swarm-wide checkpoints.
**Architecture Adjustment:**
* Introducing `Checkpoint(sessionID string) (checkpointID string, error)` and `Rollback(checkpointID string) error` to the Shard Manager.
* The Virtual Context Map now supports "Temporal Snapshots," allowing the gateway to revert the entire sharded state of a swarm to a previous valid checkpoint.
**Security Impact:** Prevents "Swarm Sanity" loss and context poisoning by rogue sub-specialists.
