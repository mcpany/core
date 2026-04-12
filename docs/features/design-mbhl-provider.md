# Design Doc: Mission-Bound Hardware Lease (MBHL) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents move from single-user CLI tools to autonomous multi-agent teams (e.g., Claude Code Agent Teams), the "Session-Bound" permission model is proving insufficient. A subagent spawned for a specific sub-task (e.g., "Refactor tests") should not inherit the entire parent session's authority indefinitely. Malicious subagents or "Capability Squatters" can exploit persistent permissions to perform unauthorized actions after their primary task is finished.

The Mission-Bound Hardware Lease (MBHL) Provider implements a "Just-in-Time" (JIT) privilege model where capabilities are issued as TPM-signed leases tied to a specific mission-root task ID and reasoning branch.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/SEP), task-specific capability leases.
    * Enforce automatic revocation of leases upon task or sub-mission completion.
    * Provide a cryptographically signed "Lease Manifest" for all high-privilege operations (e.g., FS writes, remote execution).
    * Support "Recursive Lease Pruning" for nested subagent chains.
* **Non-Goals:**
    * Managing low-level filesystem ACLs (MBHL operates at the agent bus layer).
    * Restricting read-only discovery (covered by ZKDB).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Engineer
* **Primary Goal:** Ensure a specialist "File Editor" subagent only has write access for the duration of a single `edit_file` task.
* **The Happy Path (Tasks):**
    1. Parent agent delegates an `edit_file` task to the specialist.
    2. MBHL Provider generates a TPM-signed lease token for the specific file path, bound to the Task ID.
    3. Specialist subagent presents the MBHL token to the MCP Any Filesystem Adapter.
    4. Adapter verifies the hardware signature and task-bound constraint.
    5. Specialist completes the edit; the parent agent signals "Task Done."
    6. MBHL Provider forcefully revokes the lease; subsequent attempts by the specialist to write to the file are interdicted.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Parent Agent] -->|Request Lease| B[MBHL Provider]
        B --> C[TPM Signing Engine]
        C --> D[Hardware-Locked Lease]
        D --> E[Subagent]
        E -->|Present Lease| F[Capability Adapter]
        F -->|Verify| G[TPM Verify]
        G --> H{Valid Task?}
        H -->|Yes| I[Execute Tool]
        H -->|No| J[Interdict]
    ```
* **APIs / Interfaces:**
    * `mbhl.IssueLease(taskID, scope, duration) -> LeaseToken`: Generates a hardware-bound lease.
    * `mbhl.RevokeLease(taskID) -> Success`: Forcefully terminates all leases for a specific task.
    * `mbhl.VerifyLease(token) -> Scope`: Core validation logic for capability adapters.
* **Data Storage/State:**
    * **Lease Registry:** In-memory, hardware-attested store of active leases and their mission-root parentage.

## 5. Alternatives Considered
* **Time-Bound JWTs:** Rejected because they lack hardware-binding and cannot be easily revoked before expiration if the task finishes early.
* **OS-Level Sudo/RBAC:** Rejected as too coarse-grained for granular agentic tasks like "Edit only lines 10-20 of file X."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MBHL is the primary defense against "Capability Squatting" and lateral movement in swarms.
* **Observability:** Integrated with the "Mission Lease Manager" UI for real-time visualization of active leases.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
