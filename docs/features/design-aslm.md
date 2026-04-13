# Design Doc: Atomic Shard Lock-Manager (ASLM)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent teams transition to shared memory coordination (Claude Code v3.3.0), the risk of "Shard Corruption" due to race conditions has become a primary bottleneck. Parallel teammates attempting to update reasoning anchors or task states in memory-mapped shards can overwrite each other without atomic synchronization. The ASLM provides a kernel-level locking service to ensure mission-root consistency in high-density teammate meshes.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide mission-bound atomic locks for memory-mapped teammate shards.
    * Support real-time race condition detection and automated conflict resolution.
    * Implement "Priority-Weighted Locking" where the Mission-Root can override subagent locks.
    * Reduce coordination latency in shared-memory environments to <5ms.
* **Non-Goals:**
    * Managing the underlying memory-mapping (handled by the framework or OS).
    * Providing a general-purpose distributed lock manager (this is specific to agentic shards).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent two teammates from simultaneously editing the same reasoning anchor in a shared memory shard.
* **The Happy Path (Tasks):**
    1. Teammate A requests a "Write Lock" for Shard Segment [0x100-0x200] via ASLM.
    2. ASLM verifies the teammate's identity and mission-root authority.
    3. ASLM grants the lock and pins the memory region.
    4. Teammate B attempts to write to the same segment; ASLM returns a `Wait/Conflict` signal.
    5. Teammate A completes the write and releases the lock.
    6. ASLM notifies Teammate B and triggers a "Differential Update" to their local view.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Teammate A] --> B[ASLM Proxy]
        C[Teammate B] --> B
        B --> D{Atomic Lock Table}
        D -- Conflict --> E[Wait-Graph Resolver]
        D -- Success --> F[Memory Segment]
    ```
* **APIs / Interfaces:**
    * `AcquireLock(ShardID, Range, Priority)`
    * `ReleaseLock(LockID)`
    * `DetectConflict(ShardID)`
* **Data Storage/State:** Lock table is stored in `memfd`-backed shared memory for zero-copy access.

## 5. Alternatives Considered
* **Git-based Locking**: Rejected due to high disk I/O latency (500ms+) in parallel swarms.
* **Mutex-based User-space Locks**: Rejected as they are prone to deadlocks and lack "Mission-Root" priority awareness.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Every lock request requires hardware-attested identity proof.
* **Observability:** "Lock Contention Heatmaps" are surfaced in the Swarm Topology Monitor.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
