# Design Doc: Memory-Mapped Intent Barriers (MMIB) Broker
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms scale horizontally (e.g., Claude Code Agent Teams), the "Mean Time to Coordinate" (MTTC) has become the primary performance bottleneck. Current coordination models rely on network-based mailbox locking or database-backed state synchronization, which introduce multi-second latencies during high-frequency conflict resolution. The "Cognitive Stall" experienced by parallel teammates prevents real-time collaboration.

The Memory-Mapped Intent Barriers (MMIB) Broker is required to provide a kernel-level synchronization primitive using shared memory regions. This enables lock-free, atomic "Intent Checkpoints" that parallel teammates can query and update with sub-millisecond latency, ensuring mesh stability and performance.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a high-performance coordination hub utilizing Linux `memfd_create` and memory-mapped (mmap) regions.
    * Provide lock-free, atomic synchronization for "Intent Barriers" across parallel agent processes.
    * Reduce MTTC for horizontal swarm coordination by 90% compared to network-based mailboxing.
    * Facilitate "Intent Checkpoints" that allow agents to speculatively prepare state while querying peer progress.
* **Non-Goals:**
    * Replacing the persistent Shared KV Store (Blackboard); MMIB is for transient, high-frequency synchronization.
    * Managing inter-node (distributed) memory mapping; MMIB is focused on local multi-process swarms.
    * Providing general-purpose IPC for non-agent applications.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Density Swarm Orchestrator
* **Primary Goal:** Resolve a 10-agent task collision on a shared file without 5s+ coordination locks.
* **The Happy Path (Tasks):**
    1. A lead agent spawns 10 specialist subagents to perform parallel code refactoring.
    2. Each subagent initiates an MMIB session via the Broker, which allocates a shared memory shard.
    3. Before editing a file, Agent A writes an "Intent Lock" to the memory-mapped intent barrier.
    4. Agent B queries the same MMIB region via a sub-millisecond pointer read and sees the lock.
    5. Agent B speculatively switches to a different task branch without triggering a network-based mailbox lock.
    6. Once Agent A completes the write, it atomically updates the Intent Barrier to "Resolved."
    7. All 10 agents receive a hardware-attested signal and resume their respective branches instantly.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        subgraph Kernel Space
            A[memfd_create Shard]
        end
        subgraph MCP Any Gateway
            B[MMIB Broker] -->|mmap| A
            C[Active Intent Monitor] --> B
        end
        subgraph Agent A (Process)
            D[Specialist Agent] -->|mmap read/write| A
        end
        subgraph Agent B (Process)
            E[Specialist Agent] -->|mmap read/write| A
        end
        B <-->|Hardware Attestation| D
        B <-->|Hardware Attestation| E
    ```
* **APIs / Interfaces:**
    * `mmib.CreateBarrier(missionID, shardSize) -> BarrierID`: Allocates a shared memory region.
    * `mmib.GetPointer(barrierID) -> FileDescriptor`: Returns an FD for the mmap-capable region.
    * `mmib.AtomicSignal(barrierID, offset, value)`: Performs a hardware-attested atomic update to the barrier.
* **Data Storage/State:**
    * **Barrier Registry:** In-memory map of active mission-bound barriers and authorized process IDs.
    * **Shared Shards:** Non-persistent, kernel-resident memory segments (`memfd`).

## 5. Alternatives Considered
* **Redis-based Locks:** Rejected due to 5ms-20ms network round-trip overhead and serialization tax.
* **Standard POSIX Semaphores:** Rejected because they lack "Agentic Awareness" and hardware-attested intent-binding.
* **SQLite Write-Ahead Logging (WAL):** Rejected due to disk I/O bottlenecks during high-frequency (1000+ per sec) updates.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Access to shared memory FDs is restricted via `ScmRights` (FD-passing) and hardware-attested mission tokens. Only authorized subagents can map the region.
* **Observability:** Integrated with the "Shared-Shard Race Detector" in the UI for real-time visualization of barrier contention and atomic signals.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
