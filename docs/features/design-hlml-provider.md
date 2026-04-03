# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI agents are increasingly delegated high-privilege tasks that require access to sensitive local tools (e.g., shell commands, database migrations). Standard capability-based tokens are often static and prone to "Identity Squatting" if a subagent process is not correctly terminated.

The Hardware-Locked Mission Lease (HLML) Provider is designed to issue TPM-signed, task-specific capability leases that are automatically revoked by the hardware root upon mission completion. This ensures that permissions are strictly lifecycle-bound and physically non-repudiable.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/SEP) leases for high-risk tool execution.
    * Bind capability lifespan to the specific mission-root task ID.
    * Automate the revocation of leases upon sub-task termination or mission-root exit.
    * Provide a verifiable audit trail of lease issuance and reconciliation.
* **Non-Goals:**
    * Managing low-level OS user permissions (e.g., sudoers).
    * Providing long-term persistent access tokens; all HLML leases are ephemeral.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Supervisor
* **Primary Goal:** Ensure a specialist subagent only has `fs:write` access for the duration of a specific file-refactoring task.
* **The Happy Path (Tasks):**
    1. Parent agent proposes a sub-task requiring sensitive tool access.
    2. HLML Provider generates a TPM-signed lease bound to the sub-task ID and mission-root intent.
    3. The subagent executes the tool; the gateway verifies the hardware lease against the active task state.
    4. Upon sub-task completion, the subagent signals "Mission Success."
    5. The HLML Provider immediately triggers hardware-level revocation of the lease.
    6. The ALR Engine (Automated Lease Reconciliation) audits the event and logs the successful closure.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Agent
        participant HLML as HLML Provider
        participant TPM as Hardware (TPM/SEP)
        participant ALR as ALR Engine

        Agent->>HLML: Request Mission Lease (TaskID, ToolID)
        HLML->>TPM: Sign Ephemeral Capability
        TPM-->>HLML: Hardware-Locked Token
        HLML-->>Agent: Mission Lease
        Note over Agent: Execute Task
        Agent->>HLML: Signal Completion
        HLML->>TPM: Void Lease
        HLML->>ALR: Log Reconciliation
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(missionRoot, taskID, capability) -> LeaseToken`: Mints a hardware-locked lease.
    * `hlml.RevokeLease(leaseToken) -> Status`: Forcefully voids a lease.
    * `hlml.GetActiveLeases(missionRoot) -> []LeaseInfo`: Returns current active hardware binds.
* **Data Storage/State:**
    * **Lease Registry:** In-memory, hardware-attested store of active mission-bound bindings.
    * **Audit Log:** Persistent SQLite store for reconciled lease events.

## 5. Alternatives Considered
* **JWTs with Short TTLs:** Rejected because software-based TTLs can be bypassed by system-clock manipulation. HLML uses monotonic hardware counters.
* **Linux Namespaces/gVisor:** Complementary, not alternative. HLML provides the "Permission Lease" while namespaces provide the "Execution Isolation."

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All leases require hardware attestation of the initiating parent agent.
* **Observability:** Integrated with the "Mission Lease Manager" in the UI for real-time lifecycle tracking.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
