# Design Doc: Teammate Task-List Arbiter
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the release of parallel "Agent Teams" in frameworks like Claude Code, the agentic workflow is shifting from sequential sub-delegation to horizontal, multi-teammate meshes. In this topology, multiple agents operate in parallel, claiming and executing tasks from a shared list.

However, this horizontal shift introduces significant coordination and security risks:
- **Race Conditions:** Multiple teammates claiming the same task.
- **State Inconsistency:** Parallel teammates mutating shared state (Blackboard) without mission-root reconciliation.
- **Teammate Spoofing:** Rogue subagents injecting malicious "completed" signals into the task list.

MCP Any needs a centralized, authoritative arbiter to provide secure, lock-free task coordination for these horizontal meshes.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a hardware-attested, lock-free "Task Claiming" API for inter-agent teammate meshes.
    * Enforce mission-root intent alignment on all task claims and completions.
    * Maintain a cryptographically signed "Chain of Agency" for the entire horizontal teammate mesh.
* **Non-Goals:**
    * Orchestrating the actual reasoning or logic within the teammates (handled by the respective frameworks).
    * Providing a general-purpose project management UI for humans.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Coordinate 5 specialized agents to refactor a massive codebase in parallel without state collisions or unauthorized task-claiming.
* **The Happy Path (Tasks):**
    1. The Lead Agent initializes a "Mesh Mission Root" and publishes a set of tasks to the Arbiter.
    2. Teammate A (Security Specialist) sends a `ClaimTask` request with its hardware-attested session token.
    3. The Arbiter validates the token against the Mission Root and grants an "Agency Lease" for the specific task.
    4. Teammate A executes the task and sends a `CompleteTask` signal with a signed reasoning trace.
    5. The Arbiter verifies the trace, updates the shared mission state, and marks the task as closed.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Lead[Lead Agent] -- Publish Tasks --> Arbiter[Teammate Task Arbiter]
        Arbiter -- Hardware-Attested Task List --> T1[Teammate 1]
        Arbiter -- Hardware-Attested Task List --> T2[Teammate 2]
        T1 -- ClaimTask(token) --> Arbiter
        Arbiter -- GrantLease(task_id, expiry) --> T1
        T1 -- CompleteTask(trace) --> Arbiter
        Arbiter -- UpdateState --> Blackboard[(Shared Blackboard)]
    ```
* **APIs / Interfaces:**
    - `POST /v1/mesh/tasks/publish`: Lead agent initializes the list.
    - `POST /v1/mesh/tasks/claim`: Teammate claims a task.
    - `POST /v1/mesh/tasks/complete`: Teammate marks completion with trace.
* **Data Storage/State:**
    - Task states are persisted in the SQLite Blackboard with row-level security bound to the mission-root ID.

## 5. Alternatives Considered
* **Framework-Specific Coordination:** Rejected as it lacks cross-framework interoperability (e.g., a Claude teammate and an OpenClaw auditor).
* **Distributed Locking (Redis/Etcd):** Rejected for local-first environments due to complexity and lack of hardware-attestation integration.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All `ClaimTask` requests require hardware-attested tokens. The Arbiter enforces monotonic sequence counters to prevent "Mailbox Replay" attacks.
* **Observability:** Every task claim, failure, and completion is logged to the Mesh Audit Log with full stylometric tracing.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation. Addressing parallel "Agent Teams" horizontal coordination risks.
