# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI agent swarms are increasingly performing high-privilege operations such as shell command execution, cloud infrastructure modification, and financial transactions. Traditional time-bound API keys or tokens are insufficient because they persist even after a specific sub-task is finished, allowing a compromised subagent to "squat" on permissions.

The Hardware-Locked Mission Lease (HLML) Provider addresses this by issuing TPM-signed, task-specific capability leases. These leases are cryptographically bound to a unique mission-root fragment and expire automatically upon task completion or mission termination, ensuring that privilege is never persistent.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/SEP) capability leases for high-risk agent operations.
    * Bind every lease to a specific, unique mission-root task ID.
    * Facilitate sub-millisecond, automated lease revocation by the hardware root.
    * Ensure non-repudiable audit trails for all privilege escalations.
* **Non-Goals:**
    * Replacing standard OAuth/OIDC for human-to-system authentication.
    * Managing low-level hardware drivers (uses standard TPM/SEP APIs).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Execute a sensitive database migration sub-task without leaving persistent credentials on the specialist subagent.
* **The Happy Path (Tasks):**
    1. The parent agent delegates a "Database Migration" task to a specialist subagent.
    2. The HLML Provider generates a TPM-signed lease for the `db:migrate` capability, bound to the `task_id`.
    3. The specialist subagent executes the migration using the lease handle.
    4. Upon completion, the subagent signals "Task Finished."
    5. The HLML Provider immediately invalidates the lease at the hardware level and scrubs associated memory shards.
    6. Any subsequent attempt to use the migration capability by the subagent is rejected.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        ParentAgent->>HLML: Request Lease(TaskID, Capability)
        HLML->>TPM: Sign Lease Fragment
        TPM-->>HLML: HLML-Token (Signed)
        HLML-->>SubAgent: Lease Handle
        SubAgent->>Tool: Execute(Handle)
        Tool->>HLML: Verify(Handle)
        HLML-->>Tool: Authorized (TaskID Match)
        SubAgent->>HLML: Terminate Task
        HLML->>TPM: Revoke Token
    ```
* **APIs / Interfaces:**
    * `hlml.MintLease(missionID, taskID, capability) -> LeaseHandle`
    * `hlml.VerifyLease(leaseHandle, requestedAction) -> bool`
    * `hlml.RevokeLease(leaseHandle) error`
* **Data Storage/State:** Lease state is stored in kernel-protected memory and synchronized with the hardware monotonic counter to prevent replay attacks.

## 5. Alternatives Considered
* **Time-Bound JWTs:** Rejected because they cannot be forcefully revoked before expiration without a centralized, high-latency CRL (Certificate Revocation List). HLML provides sub-millisecond hardware-level revocation.
* **Agent-Side Sandboxing:** Rejected because sandboxes can be escaped; HLML ensures that even if a subagent escapes its sandbox, its cryptographic authority is revoked.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All leases require the "Mission-Root" hardware signature as a prerequisite.
* **Observability:** Integrated with the "Mission Lease Manager" in the UI for real-time visualization of active and expired leases.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
