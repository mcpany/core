# Design Doc: CRDT-Native Mailbox Sharding
**Status:** Draft
**Created:** 2026-06-28

## 1. Context and Scope
The transition to horizontal Agent Teams has hit a performance ceiling: the "Mailbox Lock." Current teammate coordination relies on a central, synchronized task list that requires global locks for every "Claim" or "Delegate" operation. In swarms with 10+ teammates, this creates a 2s+ coordination stall, severely degrading the responsiveness of the swarm.

**CRDT-Native Mailbox Sharding** introduces Conflict-Free Replicated Data Types to the coordination bus. It allows teammates to synchronize state asynchronously and merge divergent task lists without global locks, enabling non-blocking coordination for high-density horizontal swarms.

## 2. Goals & Non-Goals
* **Goals:**
    * Eliminate global coordination locks for teammate task lists.
    * Implement CRDT-based state reconciliation for shared mailbox shards.
    * Achieve sub-100ms task claiming latency in swarms of 20+ agents.
    * Support "Optimistic Task Claiming" with automatic conflict resolution.
* **Non-Goals:**
    * Replacing the Shared KV Store (Blackboard) for general-purpose state.
    * Managing inter-framework transport (handled by A2A/Named Pipes).

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** Coordinate 15 specialists in a parallel research mission without coordination stalls.
* **The Happy Path (Tasks):**
    1. Orchestrator initializes a sharded mailbox using the CRDT provider.
    2. Teammate A and Teammate B simultaneously claim two different sub-tasks.
    3. Both teammates update their local shards instantly without waiting for a global lock.
    4. The CRDT logic propagates the claims across the mesh.
    5. A transient conflict occurs where both claim the same task.
    6. CRDT resolution rules (e.g., LWW - Last Write Wins) automatically assign the task to Teammate A.
    7. Teammate B receives the resolution signal and automatically pivots to the next available task.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        T1[Teammate 1] -->|Local Update| S1[Shard 1]
        T2[Teammate 2] -->|Local Update| S2[Shard 2]
        S1 <-->|Gossip Sync| S2
        S1 -->|Merge| State[Converged State]
        S2 -->|Merge| State
    ```
* **APIs / Interfaces:**
    * `Mailbox_ClaimTask_Async(task_id)`: Optimistically claims a task.
    * `Mailbox_Sync_Gossip()`: Background process for state propagation.
* **Data Storage/State:**
    * Task list represented as an **OR-Set** (Observed-Remove Set) or **LWW-Element-Set**.

## 5. Alternatives Considered
* **Hierarchical Locks**: Rejected as they still introduce bottlenecks and increase complexity in flat teammate meshes.
* **Synchronous Raft/Paxos**: Rejected as the 3+ round-trip latency for every claim is too slow for 100ms reasoning turns.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Every CRDT update must be signed with the hardware-attested teammate identity (HAIR).
* **Observability:** Mesh "Convergence Lag" and "Conflict Rates" are monitored in the Lock-Free Coordination Monitor.

## 7. Evolutionary Changelog
* **2026-06-28:** Initial Document Creation.
