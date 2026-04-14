# Design Doc: Mesh-Aware Deadlock Resolver (MADR)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms become more heterogeneous and parallel (e.g., Claude Code Agent Teams interacting with OpenClaw specialists), they frequently rely on shared project-local workspaces such as scratchpads or shared files for coordination. A critical "Scaling Bottleneck" has emerged where agents from different frameworks attempt to lock the same resource simultaneously, leading to "Negotiation Deadlocks" that can stall the mission for seconds or minutes.

The Mesh-Aware Deadlock Resolver (MADR) is required to act as a kernel-level arbiter that detects circular lock dependencies and applies mission-aligned resolution policies to ensure swarm liveness.

## 2. Goals & Non-Goals
* **Goals:**
    * Proactively identify circular task and resource dependencies across disparate agent frameworks.
    * Resolve deadlocks using priority-weighted mission-root rules.
    * Provide atomic, lease-bound access to shared project scratchpads.
    * Log deadlock events for later "Swarm Sanity" analysis.
* **Non-Goals:**
    * Replacing framework-specific internal scheduling.
    * Managing distributed database locks (it focuses on project-local shared state).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Prevent a deadlock where Agent A (Claude) waits for a scratchpad lock held by Agent B (OpenClaw), while Agent B waits for a tool result from Agent A.
* **The Happy Path (Tasks):**
    1. Agent A and Agent B both request a write-lock on `.scratchpad`.
    2. MADR intercepts both requests and maps them to its internal "Wait-Graph."
    3. MADR detects a potential circular dependency based on the agents' active intents.
    4. MADR evaluates the "Mission Priority" of both agents (authorized via HLML).
    5. MADR grants the lock to the higher-priority agent and issues a "Wait-with-Backoff" signal to the other.
    6. If a hard deadlock is already formed, MADR forcefully revokes the lock from the lower-priority agent and triggers an "Atomic State Rollback."

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent A] -->|Lock Req| C[MADR Arbiter]
        B[Agent B] -->|Lock Req| C
        C -->|Analyze| D(Wait-Graph)
        D -->|Cycle Detected| E{Conflict Resolver}
        E -->|High Priority| A
        E -->|Pre-empt| B
        A -->|Release| C
    ```
* **APIs / Interfaces:**
    * `madr.AcquireLock(resourceID, missionToken) -> LockID`: Requests an atomic lock.
    * `madr.ReleaseLock(lockID)`: Releases the held resource.
    * `madr.DetectDeadlock() -> Graph`: Internal background service for graph analysis.
* **Data Storage/State:**
    * **Wait-Graph Store:** In-memory graph of active lock holders and waiters.
    * **Priority Registry:** Hardware-attested list of mission-root priorities.

## 5. Alternatives Considered
* **Framework-Specific Locking:** Rejected because it doesn't work across frameworks (Claude can't see OpenClaw's internal locks).
* **Simple Timeouts:** Rejected because they lead to "Thundering Herd" problems and inefficient resource usage.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Lock requests must be accompanied by an HLML-signed mission token to prevent "Lock Denial of Service" attacks by rogue subagents.
* **Observability:** Deadlock resolutions are visualized in the "Swarm Topology Monitor" UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
