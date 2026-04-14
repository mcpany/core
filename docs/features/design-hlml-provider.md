# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The persistent nature of tool permissions in current agent frameworks (e.g., granting a subagent `run_shell_command` for the duration of a session) creates a significant attack surface for "Capability Squatting" and "Credential Shadowing." If a specialist subagent is compromised or hallucinates, it can reuse high-trust permissions outside its intended task.

The Hardware-Locked Mission Lease (HLML) Provider addresses this by moving from session-bound permissions to task-specific, TPM-signed leases that expire automatically upon task completion.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, mission-bound capability leases for high-risk operations.
    * Ensure leases are mathematically and cryptographically restricted to a specific task ID and temporal window.
    * Automate the revocation of capabilities once the mission-root reports task completion.
    * Neutralize "Capability Squatting" by rogue or orphaned subagents.
* **Non-Goals:**
    * Managing low-level hardware drivers for TPMs.
    * Replacing the base Policy Firewall; HLML provides the *lease* that the Firewall validates.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious DevSecOps Engineer
* **Primary Goal:** Grant a "Security Auditor" agent permission to read `/etc/hosts` only during the "Vulnerability Scan" phase of a CI/CD mission.
* **The Happy Path (Tasks):**
    1. Parent agent initiates the "Vulnerability Scan" task.
    2. HLML Provider generates a TPM-signed lease for `fs:read:/etc/hosts` bound to `TaskID: Scan-99`.
    3. The Security Auditor agent receives the lease and attempts to read the file.
    4. Policy Firewall validates the request against the HLML lease and hardware signature.
    5. Auditor completes the scan and reports task completion to the mission root.
    6. HLML Provider receives the completion signal and marks the lease as "Revoked" in the hardware enclave.
    7. Any subsequent attempts by the Auditor to use that lease are blocked at the hardware level.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -->|Start Task| B[HLML Provider]
        B -->|Request Signature| C[TPM / Secure Enclave]
        C -->|TPM-Signed Lease| B
        B -->|Issue Lease| D[Specialist Agent]
        D -->|Tool Call + Lease| E[Policy Firewall]
        E -->|Verify Hardware Sig| C
        E -->|Execute| F[MCP Server]
        A -->|End Task| B
        B -->|Invalidate Lease| C
    ```
* **APIs / Interfaces:**
    * `hlml.CreateLease(agentID, taskID, scope, duration) -> Lease`: Generates a hardware-signed lease.
    * `hlml.ValidateLease(lease, request) -> bool`: Verifies signature and task-bound constraints.
    * `hlml.RevokeLease(taskID)`: Forcefully invalidates all leases associated with a task.
* **Data Storage/State:**
    * **Lease Registry:** In-memory, hardware-backed store of active lease hashes and their associated task metadata.

## 5. Alternatives Considered
* **Time-Bound JWTs:** Rejected because they can be replayed or used outside the intended "Task" if the agent process is hijacked. HLML binds the permission to the hardware session and specific mission branch.
* **Manual User Approval (HITL):** Rejected as the primary mechanism due to "Approval Fatigue," though it remains the fallback for lease creation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** HLML is the cornerstone of "Lifecycle-Bound Agency." It ensures that even if an agent's memory is dumped, its high-trust leases are useless once the task is over.
* **Observability:** Leases are tracked in the "Mission Lease Manager" UI, showing real-time countdowns to expiration.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation based on Claude Code v3.2.0 MBHL patterns.
