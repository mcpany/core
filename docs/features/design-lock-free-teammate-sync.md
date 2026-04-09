# Design Doc: Lock-Free Teammate Synchronization (LFTS)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms scale horizontally (e.g., Claude Code Agent Teams), the Mean Time to Coordinate (MTTC) has surfaced as the primary performance bottleneck. Current architectures rely on centralized coordination locks for the shared task list, leading to "Cognitive Stall" where agents wait for mailbox access. LFTS introduces a decentralized, lock-free synchronization layer to ensure sub-millisecond coordination.

## 2. Goals & Non-Goals
* **Goals:**
    * Eliminate centralized coordination locks for the shared teammate mailbox.
    * Utilize Conflict-Free Replicated Data Types (CRDTs) to ensure eventually consistent state across parallel teammates.
    * Reduce MTTC for swarms of 10+ agents by 90%.
    * Support disconnected operation and merge-on-reconnect for local-first agent meshes.
* **Non-Goals:**
    * Replacing the primary mission-root intent (LFTS manages the *execution* of the intent, not its definition).
    * Providing long-term state persistence (LFTS is for active session coordination).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate 5 specialized agents to refactor a codebase without coordination deadlocks.
* **The Happy Path (Tasks):**
    1. Mission Root initializes the "Refactor Mesh".
    2. 5 Specialist Agents connect to the LFTS Bus.
    3. Agent A identifies a task and broadcasts a "Claim" intent to its local CRDT replica.
    4. LFTS propagates the claim to all teammates via high-speed P2P gossip.
    5. Agent B concurrently identifies a separate task and claims it; the CRDT resolves any potential overlaps based on mission-root priority.
    6. All agents proceed with execution without ever waiting for a global lock.

## 4. Design & Architecture
* **System Flow:**
    `Agent A` -> `Local CRDT Shard` -> `P2P Gossip (Named Pipes/WS)` -> `Agent B Shard`
* **APIs / Interfaces:**
    * `SyncBus`: `Broadcast(update Delta) error`
    * `Mailbox`: `ClaimTask(taskID string) bool`, `DelegateTask(task Task) error`
* **Data Storage/State:**
    * Task list stored as a causal-ordered set (OR-Set) within the CRDT.

## 5. Alternatives Considered
* **Distributed Locking (Raft/Paxos)**: Rejected due to the high latency tax (50ms+) on every coordination event.
* **Centralized Redis Mailbox**: Rejected to maintain "Local Sovereignty" and avoid external dependencies.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: All CRDT updates must be cryptographically signed by the agent's hardware-attested identity.
* **Observability**: Real-time visualization of the "Sync Graph" in the MCP Any UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
