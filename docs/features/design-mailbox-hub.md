# Design Doc: Lock-Free Sharded Mailbox Hub
**Status:** Draft
**Created:** 2026-06-25

## 1. Context and Scope
As agent swarms scale horizontally (e.g., Claude Code Agent Teams), the "Mailbox Lock" has become the primary bottleneck for inter-teammate coordination. Current synchronous locking mechanisms lead to high latency and contention when multiple agents attempt to claim or update tasks simultaneously.

The **Lock-Free Sharded Mailbox Hub** introduces a decentralized coordination model using Conflict-Free Replicated Data Types (CRDTs). By sharding the mailbox at the task level and utilizing lock-free synchronization, it allows teammates from disparate frameworks to coordinate at machine-speed without global wait-states.

## 2. Goals & Non-Goals
* **Goals:**
    * Eliminate global coordination locks for inter-agent task claiming.
    * Implement CRDT-based state synchronization for the teammate mailbox.
    * Support sub-10ms task-claiming latency for swarms of up to 50 agents.
    * Ensure mission-root intent is cryptographically bound to every shard.
* **Non-Goals:**
    * Replacing the Shared KV Store (Blackboard) for long-term state.
    * Implementing agent-specific reasoning logic for task selection.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Infrastructure Architect
* **Primary Goal:** Coordinate a high-density team of 20 agents on a massive codebase migration without performance degradation.
* **The Happy Path (Tasks):**
    1. The mission-root agent initializes the Sharded Mailbox Hub with a task manifest.
    2. Teammates connect and receive their respective mailbox shards.
    3. Multiple teammates simultaneously claim different sub-tasks from the hub.
    4. The hub reconciles claims using CRDT logic, ensuring zero-collision without locking the entire mailbox.
    5. State updates are propagated asynchronously to all interested peers.

## 4. Design & Architecture
* **System Flow:**
    `Teammate A <-> [Task Shard A (CRDT)] <-> Mailbox Hub <-> [Task Shard B (CRDT)] <-> Teammate B`
* **APIs / Interfaces:**
    * `mailbox.v2.ClaimTask(task_id, teammate_id) -> claim_token`
    * `mailbox.v2.SyncShard(shard_id, state_delta) -> void`
* **Data Storage/State:**
    * Shards are maintained as in-memory LWW-Element-Sets (Last-Write-Wins) backed by the Blackboard for persistence.

## 5. Alternatives Considered
* **Distributed Locking (Etcd/Consul):** Rejected due to the high latency floor of network-roundtrip consensus for every task claim.
* **Synchronous Relational Database:** Rejected due to row-level contention and scaling limits under high-frequency agentic updates.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Every mailbox request must be signed with a **Multi-Channel Session Sovereignty (MCSS)** token to prevent teammate impersonation.
* **Observability:** We will track `ConflictRate` and `CoordinationLatencyMs` per shard.

## 7. Evolutionary Changelog
* **2026-06-25:** Initial Document Creation.
