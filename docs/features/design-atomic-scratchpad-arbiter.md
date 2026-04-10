# Design Doc: Atomic Scratchpad Arbiter
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of horizontal Agent Teams (e.g., Claude Code Agent Teams), multiple subagents frequently coordinate via a shared project-local workspace, often a `.scratchpad` file or directory. As swarms scale to 10+ agents, synchronous file-system locks lead to "Scratchpad Deadlocks," causing significant coordination stalls (5s-10s) or session crashes due to write contention.

MCP Any needs to provide a kernel-level lock management service that mediates access to these shared workspaces. The Atomic Scratchpad Arbiter will act as the authoritative gatekeeper, ensuring that state mutations are atomic, mission-aligned, and non-blocking for high-density parallel swarms.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide sub-millisecond lock acquisition for project-local scratchpads.
    * Implement mission-bound atomic write-access to prevent race conditions.
    * Support "Priority-Aware Queuing" where the Mission-Root can override specialist locks.
    * Detect and resolve circular wait-graphs between teammates.
* **Non-Goals:**
    * Replacing general-purpose version control (Git).
    * Managing remote cloud storage locks (handled by S3/GCS providers).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Enable 12 parallel specialist agents to write architectural updates to a shared `.scratchpad` without deadlocking the session.
* **The Happy Path (Tasks):**
    1. Agent A requests a "Write-Intent" lock for `.scratchpad` via the MCP Any gateway.
    2. The Arbiter verifies Agent A's mission-bound authority and grants a 500ms lease.
    3. Agent B requests a lock while Agent A is writing; the Arbiter places Agent B in a mission-priority queue.
    4. Agent A commits the write; the Arbiter atomically updates the file and releases the lock.
    5. Agent B is instantly notified and acquires the lock with zero-latency handoff.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Agent
        participant Gateway as MCP Any Gateway
        participant Arbiter as Atomic Scratchpad Arbiter
        participant FS as Local Filesystem

        Agent->>Gateway: acquire_lock(path=".scratchpad")
        Gateway->>Arbiter: Check Lease & Priority
        Arbiter->>Arbiter: Update Wait-Graph
        Arbiter-->>Gateway: Lock Granted (Lease ID: 0xAF)
        Gateway-->>Agent: Success
        Agent->>Gateway: atomic_write(Lease ID, Content)
        Gateway->>Arbiter: Validate Lease
        Arbiter->>FS: O_DIRECT Atomic Write
        FS-->>Arbiter: OK
        Arbiter-->>Gateway: Commit Success
        Gateway-->>Agent: OK
    ```
* **APIs / Interfaces:**
    * `mcp.scratchpad.acquire(path, priority, ttl_ms)`
    * `mcp.scratchpad.commit(lease_id, content_hash, delta)`
    * `mcp.scratchpad.status(path)`
* **Data Storage/State:**
    * In-memory lock table with hardware-attested lease IDs.
    * Persistent wait-graph stored in the Shared KV Store (Blackboard).

## 5. Alternatives Considered
* **OS-Level File Locking (`flock`):** Rejected due to lack of mission-awareness and high overhead for high-frequency sub-millisecond writes.
* **Git-based Coordination:** Rejected due to the latency of commits and the complexity of managing 100+ micro-merges per minute.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Locks are only granted to agents with a verified mission-token. The Arbiter redacts sensitive fragments from the scratchpad before granting read-access to low-trust specialists.
* **Observability:** Real-time visualization of the Scratchpad Contention Dashboard, showing lock-wait times and queue depths.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
