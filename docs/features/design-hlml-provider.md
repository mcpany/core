# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Headless AI agents and long-running swarms often inherit capabilities that persist beyond the duration of a specific task. This "Privilege Squatting" presents a significant security risk, as a compromised specialist agent could retain access to high-privilege tools (e.g., shell, filesystem) indefinitely. Claude Code v3.2.0 has introduced Mission-Bound Hardware Leases (MBHL) as a solution. MCP Any needs to implement a universal HLML Provider to standardize this lifecycle-bound agency across all connected frameworks.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, task-specific capability leases for high-risk tool calls.
    * Enforce automatic revocation of leases upon sub-mission or task completion.
    * Provide hardware-attested proof of mission-root lineage for every lease.
    * Synchronize with the Deterministic Mission Lifecycle Controller for resource reclamation.
* **Non-Goals:**
    * Replacing existing A2A authentication; it adds a temporal/lifecycle layer to it.
    * Managing tool-specific permissions (handled by the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Ensure a subagent spawned for a specific refactoring task cannot access the shell after the refactoring is complete.
* **The Happy Path (Tasks):**
    1. Parent agent spawns a Refactoring Subagent via A2A.
    2. HLML Provider issues a hardware-attested lease for `run_shell_command` bound to the `RefactorTaskID`.
    3. The Subagent executes shell commands within the scope of the task.
    4. Upon `RefactorTaskID` completion signal, the HLML Provider broadcasts a revocation event.
    5. The Hardware Root (TPM) invalidates the lease signature.
    6. Subsequent attempts by the Subagent to use the shell are interdicted by the gateway.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] --> B[HLML Provider]
        B --> C[TPM Key Mint]
        C --> D[Hardware Lease Token]
        D --> E[Subagent]
        E --> F{Tool Gateway}
        F -->|Validate Token| G[TPM Verification]
        G -->|Expired/Invalid| H[Access Denied]
        G -->|Valid| I[Execute Tool]
    ```
* **APIs / Interfaces:**
    * `hlml.MintLease(missionToken, taskID, scopes) -> LeaseToken`
    * `hlml.RevokeLease(taskID)`
    * `hlml.ValidateLease(leaseToken) -> bool`
* **Data Storage/State:**
    * **Lease Registry:** SQLite-backed store for tracking active task-to-lease mappings.
    * **TPM State:** Hardware-bound monotonic counters to ensure lease uniqueness.

## 5. Alternatives Considered
* **Time-Bound JWTs:** Rejected because they don't account for early task completion and can be exfiltrated. HLML requires hardware-bound verification.
* **Standard RBAC:** Insufficient for the dynamic, short-lived nature of autonomous sub-missions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Leases are cryptographically tied to both the hardware ID and the mission-root lineage.
* **Observability:** Integrated with the "Mission Lease Manager" UI for real-time tracking of lease lifetimes.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
