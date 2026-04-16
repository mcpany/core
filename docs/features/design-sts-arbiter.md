# Design Doc: Shared Teammate Scratchpad (STS) Arbiter
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In high-density horizontal agent swarms (e.g., Claude Code Agent Teams), parallel teammates often need to collaborate on intermediate file artifacts or code fragments. Writing directly to the primary project directory can lead to pollution and accidental corruption of the stable source tree.

The Shared Teammate Scratchpad (STS) Arbiter is required to provide a volatile, sharded filesystem fragment where teammates can collaborate with hardware-attested atomic write-access and conflict resolution.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide isolated, volatile "Scratchpad" shards for parallel missions.
    * Enforce hardware-attested atomic write-access to prevent race conditions.
    * Facilitate real-time conflict resolution between concurrent teammate writes.
    * Ensure scratchpad data is purged automatically upon mission completion.
* **Non-Goals:**
    * Providing long-term persistent storage; scratchpads are volatile.
    * Replacing the primary Project-Local Snapshot Sync (PLSS); it serves as a "Whiteboard" for uncommitted changes.

## 3. Critical User Journey (CUJ)
* **User Persona:** Parallel Teammate Orchestrator
* **Primary Goal:** 5 agents concurrently refining a single documentation file in a scratchpad without overwriting each other's work.
* **The Happy Path (Tasks):**
    1. Parent agent initializes a "Mission Scratchpad" via the STS Arbiter.
    2. Agents A, B, and C mount the scratchpad shard.
    3. Agent A attempts to write a new paragraph.
    4. STS Arbiter grants an "Atomic Write Lock" to Agent A.
    5. Agent B attempts to edit the same line; STS Arbiter queues the request or triggers a "Conflict Alert".
    6. Agent A commits the change; the lock is released.
    7. Agent B's change is merged or rejected based on priority rules.
    8. Once the final version is reached, it is promoted to the host filesystem.
    9. Scratchpad is securely wiped.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        A[Agent A] --> B[STS Arbiter]
        C[Agent B] --> B
        B --> D[Volatile Shard Store]
        B --> E[Lock Manager]
        B --> F[Conflict Resolver]
    ```
* **APIs / Interfaces:**
    * `sts.CreateScratchpad(missionID) -> ShardID`: Provision a new volatile workspace.
    * `sts.WriteFragment(shardID, path, content, attestation) -> Status`: Securely updates a file fragment.
    * `sts.Promote(shardID, hostPath) -> Success`: Merges scratchpad content to the host.
* **Data Storage/State:**
    * **Volatile Buffer (tmpfs):** High-speed, in-memory filesystem for scratchpad content.

## 5. Alternatives Considered
* **Standard Linux Sockets:** Rejected because they don't provide a shared filesystem abstraction.
* **Global Blackboard (KV):** Useful for small metadata, but inefficient for large file artifacts.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Write access is strictly tied to the hardware-attested mission lineage.
* **Observability:** Scratchpad write-latency and contention events are monitored in the "Scratchpad Contention Monitor".

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
