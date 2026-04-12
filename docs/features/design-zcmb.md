# Design Doc: Zero-Copy Memory Broker (ZCMB)
**Status:** Draft
**Created:** 2026-07-03

## 1. Context and Scope
As AI agent swarms evolve from linear execution to high-frequency parallel coordination, the latency of state transfer becomes a critical bottleneck. Traditional JSON-based or even Protobuf-based Binary State Handoffs (BSH) incur significant serialization and transport overhead, leading to "Cognitive Stall" in deep specialist meshes.

The Zero-Copy Memory Broker (ZCMB) addresses this by providing hardware-locked, shared memory regions where multiple specialist agents can directly read and write reasoning traces. This eliminates the serialization tax and enables sub-millisecond state sharing for local swarms, while maintaining strict mission-bound security boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide sub-millisecond state synchronization between specialist agents.
    * Eliminate serialization/deserialization overhead for large reasoning traces.
    * Enforce hardware-attested memory boundary isolation.
    * Provide mission-bound access control for shared memory regions.
* **Non-Goals:**
    * Providing zero-copy transport for remote/cloud-based agents (out of scope for shared memory).
    * Replacing the Blackboard (Shared KV Store) for persistent, long-term state.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Enable high-frequency refinement loops between a Dev Agent and an Auditor Agent without transport-induced latency.
* **The Happy Path (Tasks):**
    1. The primary agent initiates a high-intensity mission.
    2. MCP Any allocates a hardware-locked shared memory region (ZCMB Shard) for the mission.
    3. The Dev Agent and Auditor Agent are granted time-bound, mission-attested handles to the shard.
    4. The Dev Agent writes its reasoning trace directly to the shard.
    5. The Auditor Agent reads the trace instantly and appends its feedback in-place.
    6. Upon mission completion, the ZCMB Shard is scrubbed and reclaimed.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] --> B[ZCMB Broker]
        B --> C[Allocated Shared Memory Region]
        C --> D[Specialist Agent A]
        C --> E[Specialist Agent B]
        B -.-> F[Hardware TPM/Enclave]
        F -.-> C[Boundary Locking]
    ```
* **APIs / Interfaces:**
    * `RequestZCMBShard(ctx, missionID) (Handle, error)`
    * `AttachAgentToShard(ctx, shardID, agentID, permission) error`
    * `RevokeShardAccess(ctx, shardID, agentID) error`
* **Data Storage/State:** State is managed in volatile, kernel-mapped shared memory buffers, synchronized via the ZCMB Broker and backed by TPM-bound session keys.

## 5. Alternatives Considered
* **BSH (Binary State Handoffs):** Rejected for high-frequency local swarms due to the 50ms+ serialization tax per hop.
* **Global Blackboard (SQLite):** Rejected for reasoning traces due to disk I/O and locking contention in parallel teams.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** We use hardware-bound Inode pinning and memory-mapping locks to prevent "Memory-Mapped Escape" attacks. Agents are only granted access to shards matching their hardware-attested mission token.
* **Observability:** Memory pressure and shard utilization are tracked via the ZCMB Telemetry Sink.

## 7. Evolutionary Changelog
* **2026-07-03:** Initial Document Creation.
* **2026-07-25:** Evolution to Kinetic Memory Sovereignty.
    * **Context:** Today's research on OpenClaw v3.7.0 reveals a shift toward "Kinetic State Handoffs" to eliminate the serialization tax in high-density meshes.
    * **Architecture Adjustment:** Transitioning the ZCMB from local-only shared memory to an **RDMA-Aware Transport Adapter**. This enables zero-copy performance across physical device boundaries in the mesh.
    * **Security Impact:** Introducing **Kernel-Level Syscall Gating (KLSG)** to the memory broker. This ensures that even with RDMA-level access, high-risk syscalls (e.g., `mmap`) are hardware-attested against the mission-root lifecycle.
