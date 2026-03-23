<!-- markdownlint-disable MD013 MD030 MD032 MD022 MD007 MD033 MD031 MD004 MD024 MD026 MD012 MD003 MD029 MD040 MD009 -->
# Design Doc: Lock-Free Mesh Coordination (LFMC)
**Status:** Draft
**Created:** 2026-06-18

## 1. Context and Scope
With the rise of horizontal "Agent Teams" (Claude Code), the traditional "Mailbox Lock" pattern—where only one agent can access the shared task list at a time—has become a major performance bottleneck. In complex refactors, parallel teammates are spending up to 40% of their time waiting for a lock to clear. MCP Any must provide a high-performance, lock-free coordination layer for the "Universal Agent Mesh" that allows teammates to claim, delegate, and synchronize tasks asynchronously.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a CRDT-based shared task list for lock-free coordination.
    *   Support high-frequency task claiming and state synchronization.
    *   Ensure mission-root consistency across parallel teammates.
*   **Non-Goals:**
    *   Replacing the primary Blackboard for structured, long-term state.
    *   Providing global ordering for all events (we focus on *eventual consistency*).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Agent Team Lead (Claude Code)
*   **Primary Goal:** Delegate three independent sub-tasks to teammates simultaneously without blocking on a shared mailbox lock.
*   **The Happy Path (Tasks):**
    1.  The Team Lead creates a task list in the LFMC Hub.
    2.  The LFMC Hub initializes a CRDT (LWW-Element-Set) for the mission.
    3.  Teammate A claims "API Layer," Teammate B claims "Migrations," and Teammate C claims "Tests" in parallel.
    4.  The LFMC Hub reconciles the claims asynchronously; no global lock is held.
    5.  Teammates work independently and update their task status in the mesh.
    6.  The Team Lead synthesizes the results, observing a consistent view of the completed tasks.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph LR
        A[Teammate A] --> B[LFMC Shard]
        C[Teammate B] --> B
        D[Teammate C] --> B
        B --> E{CRDT Reconciler}
        E --> F[Consistent Task List]
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/mesh/task/claim`: Asynchronously claim a task from the shared list.
    *   `GET /v1/mesh/state`: Retrieve the current eventually-consistent state of the mesh.
*   **Data Storage/State:**
    *   Task states are stored in memory-mapped CRDT buffers for sub-millisecond access.

## 5. Alternatives Considered
*   **Redis-based Locking:** Rejected because it introduces a central point of failure and network-level latency that blocks high-speed teammate coordination.
*   **Message Queues:** Rejected because they don't provide a shared, convergent state for teammates to "observe" their peers' progress.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** All mesh interactions must be signed with a hardware-attested teammate identity (SMI).
*   **Observability:** The "Lock-Free Mesh Arbiter" in the UI will visualize real-time task claiming and CRDT convergence.

## 7. Evolutionary Changelog
*   **2026-06-18:** Initial Document Creation.
### Update: [2026-06-19] - Sovereign Sharding for Semantic Integrity
**Context:** Claude Code v2.2.0-rc1 previews revealed "Semantic Smearing," where parallel teammates over-write intent fragments in adjacent shards.
**Architecture Adjustment:** * Introducing **Sovereign Sharding** in Section 4.
* Shards are now cryptographically bound to the Mission-Root intent via HAIL.
* Implementing a "One-Way Intent Flow" policy to prevent back-propagation of subagent drift into parent shards.
**Security Impact:** Prevents malicious or hallucinating subagents from corrupting the primary mission reasoning path.
