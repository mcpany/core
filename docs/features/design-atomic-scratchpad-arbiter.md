# Design Doc: Atomic Scratchpad Arbiter
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of horizontal "Agent Teams" (e.g., Claude Code), multiple specialist agents often collaborate via shared, project-local workspaces known as scratchpads (e.g., `.scratchpad`). Today's findings reveal that high-frequency parallel writes to these files result in a 12% state corruption rate due to the lack of atomic locking at the filesystem level.

The Atomic Scratchpad Arbiter is needed to act as the kernel-level lock manager for these shared resources. It ensures that parallel teammates can synchronize state without lost updates, race conditions, or "Ghost Fragment" pollution.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide mission-bound atomic write-access to project-local scratchpad files.
    * Implement a "Snapshot-and-Merge" reconciliation model for parallel teammate updates.
    * Enforce Zero-Trust isolation between different mission-root scratchpads.
    * Neutralize "Teammate Deadlock" during high-contention cycles.
* **Non-Goals:**
    * Managing remote cloud storage (RAR focused on local FS).
    * Providing long-term archival of scratchpad history (handled by PLSS).
    * Modifying the content of the scratchpad (only manages access/atomicity).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate a "Writer" subagent and a "Linter" subagent editing the same scratchpad without data loss.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a mission and initializes a `.scratchpad` for the team.
    2. The "Writer" subagent requests a write-lock for the scratchpad.
    3. The Arbiter grants the lock and provides a mission-bound session token.
    4. The "Writer" performs its edit and commits the fragment.
    5. Simultaneously, the "Linter" requests a lock.
    6. The Arbiter queues the request and provides the "Linter" with a speculatively updated snapshot.
    7. Once the "Writer" commit is verified, the Arbiter grants the "Linter" the lock and merges its changes atomically.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Teammate A] -->|Write Request| B[Scratchpad Arbiter]
        C[Teammate B] -->|Write Request| B
        B --> D{Lock Available?}
        D -- Yes --> E[Grant Mission-Bound Lock]
        D -- No --> F[Speculative Snapshot & Queue]
        E --> G[Atomic Commit to .scratchpad]
        G --> H[Notify Queue]
        H --> B
    ```
* **APIs / Interfaces:**
    * `arbiter.AcquireLock(scratchpadPath string, missionID string) -> LockToken`
    * `arbiter.CommitAtomic(token LockToken, fragment string) -> bool`
* **Data Storage/State:**
    * **Lock Registry:** In-memory, kernel-resident table of active scratchpad locks.
    * **Wait-Graph:** Real-time graph of teammates waiting for specific resource locks.

## 5. Alternatives Considered
* **Git-Based Coordination:** Rejected due to excessive latency (>2s per commit) which causes "Cognitive Stall."
* **Shared-Nothing Architecture:** Rejected because horizontal swarms require a "Blackboard" for efficient coordination.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Locks are cryptographically bound to the hardware-attested mission root.
* **Observability:** Scratchpad contention and write-latency metrics are visualized in the `Mesh Resilience Status Hub`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
