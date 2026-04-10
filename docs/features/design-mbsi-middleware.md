# Design Doc: Mission-Bound Scratchpad Isolation (MBSI) Middleware
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
In horizontal Agent Teams (e.g., Claude Code), teammates often share a project-local "scratchpad" file (e.g., `.scratchpad`) to coordinate transient state. However, the lack of atomic synchronization and mission-bound isolation has led to "Scratchpad Pollution," where teammates from different mission branches or unauthorized subagents overwrite critical coordination metadata, causing state corruption.

The MBSI Middleware is required to provide kernel-level, mission-bound atomic locks and isolation for all writes to shared team workspaces. It ensures that only teammates with a verified, hardware-attested mission-root lineage can modify specific scratchpad fragments.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement mission-bound atomic locking for shared project-local files.
    * Enforce lineage-aware access control for all scratchpad writes.
    * Neutralize "Scratchpad Pollution" by isolating mission branches at the filesystem level.
    * Provide real-time "Conflict Resolution" for parallel teammate writes.
* **Non-Goals:**
    * Replacing the Shared KV Store (Blackboard); it specifically manages unstructured file-based scratchpads.
    * Managing remote cloud storage; focus is on project-local coordination files.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a rogue subagent from corrupting the shared coordination file used by the primary agent team.
* **The Happy Path (Tasks):**
    1. A primary agent team starts a "Deployment" mission.
    2. The MBSI Middleware initializes a mission-bound lock for the `.scratchpad` file.
    3. Teammate A requests a write lock to update the deployment status.
    4. MBSI verifies Teammate A's hardware-attested mission token and grants an atomic lock.
    5. A rogue subagent (unauthorized branch) attempts to overwrite the `.scratchpad`.
    6. MBSI detects the lineage mismatch and interdicts the write, alerting the primary supervisor.
    7. Teammate A completes the write, and the lock is released.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Agent[Agent Teammate] -->|Write Request| Gateway[MBSI Middleware]
        Gateway -->|Verify Lineage| Auth[TPM/Lineage Provider]
        Auth -->|Verified| Lock[Atomic Lock Manager]
        Lock -->|Execute Write| FS[Project Scratchpad]
        Auth -->|Deny| Alert[Security Auditor]
    ```
* **APIs / Interfaces:**
    * `mbsi.AcquireLock(filePath, missionToken) -> LockID`: Requests a mission-bound lock.
    * `mbsi.CommitWrite(lockID, data) -> Status`: Atomically writes data and releases the lock.
    * `mbsi.InspectSovereignty(filePath) -> MissionLineage`: Returns the mission root currently owning the file.
* **Data Storage/State:**
    * **Kernel-Bound Lock Registry:** Tracks active file locks and their mission-root owners.
    * **Lineage Cache:** Local cache of verified parent-child agent tokens.

## 5. Alternatives Considered
* **Standard OS File Locking (flock):** Rejected because it lacks "Agentic Awareness" and cannot distinguish between mission branches.
* **Application-Level SQLite Blackboard:** While robust, many existing agent frameworks (Claude Code, OpenClaw) rely on natural-language markdown scratchpads. MBSI secures these legacy patterns.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All locks are cryptographically bound to the TPM. "Echo-Immune" monotonic timestamps are used to prevent replay attacks on the lock manager.
* **Observability:** Integrated with the "Scratchpad Integrity Dashboard" to visualize atomic locks and blocked pollution attempts.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
