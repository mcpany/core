# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of Agent Teams and horizontal swarms (e.g., Claude Code v3.2), session-based privilege is no longer granular enough. A compromised subagent in a long-running session can "squat" on capabilities (like shell access) long after its specific task is done.

The HLML Provider implements Mission-Bound Hardware Leases. Capabilities are tied to a TPM-signed lease that is cryptographically restricted to a specific mission-root task ID and automatically expires upon task completion. This ensures "Just-in-Time" agency that is non-repudiable and hardware-enforced.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, task-specific capability leases.
    * Automatically revoke capabilities via the hardware root upon mission-root task completion.
    * Neutralize persistent privilege escalation in specialist agents.
    * Provide hardware-attested proof of lease expiration.
* **Non-Goals:**
    * Replacing session-level authentication.
    * Managing non-hardware-bound permissions.
    * Providing real-time capability rotation for low-risk tools.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Ensure a specialist subagent only has `run_shell_command` access for the exact duration of a "Build & Test" task.
* **The Happy Path (Tasks):**
    1. Parent agent delegates a "Build" task to a specialist subagent.
    2. HLML Provider generates a TPM-signed lease for `build-task-789` granting `sh:exec`.
    3. The specialist subagent performs the build using the lease.
    4. Upon completion, the subagent returns the result to the parent.
    5. The mission root signals task completion to the HLML Provider.
    6. The HLML Provider invalidates the lease in hardware memory; any further attempts to use the lease trigger a security fault.
    7. A "Lease Revocation Receipt" is logged for auditing.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -->|Authorize Task| B[HLML Provider]
        B -->|Issue Lease| C[TPM / Secure Enclave]
        C -->|Sign| D[Lease Token]
        D --> E[Specialist Subagent]
        E -->|Call Tool w/ Lease| F[MCP Any Gateway]
        F -->|Verify w/ TPM| C
        A -->|Signal Completion| B
        B -->|Evict Key| C
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(taskID, capabilities) -> LeaseToken`: Generates a TPM-bound lease.
    * `hlml.ValidateLease(leaseToken) -> bool`: Checks hardware validity and mission state.
    * `hlml.RevokeLease(taskID) -> Receipt`: Forcefully clears hardware-backed privileges.
* **Data Storage/State:**
    * **Lease Registry:** TPM-bound, volatile memory region for active keys.
    * **Mission Metadata:** Intent-bound mapping of task IDs to lease signatures.

## 5. Alternatives Considered
* **Time-Bound JWTs:** Rejected because they can be "replayed" or "squatted" until expiration. HLML allows for immediate, hardware-enforced revocation upon task event signals.
* **Manual User Approval (HITL):** Rejected as the primary mechanism due to "Approval Fatigue"; HLML provides the automated safety needed for full delegation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Leases are mathematically restricted to the subset of the caller's existing authority.
* **Observability:** Integrated with the "Mission Lease Manager" UI for visual tracking of lease lifetimes.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
