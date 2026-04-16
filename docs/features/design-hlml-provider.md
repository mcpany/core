# Design Doc: Hardware-Locked Mission Leases (HLML)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of autonomous subagents and specialist teammates, the risk of persistent privilege escalation has become a critical vulnerability. If a subagent is granted high-risk capabilities (e.g., shell access) for a specific task, those permissions often linger beyond the task's completion, allowing for "Mission Squatting" or unauthorized lateral movement.

Hardware-Locked Mission Leases (HLML) are required to provide time-bound, task-specific, and hardware-attested capability leases that expire automatically upon mission-root task completion.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement TPM-signed, task-specific capability leases for all high-risk subagent delegations.
    * Ensure automatic revocation of privileges by the hardware root once the mission-root task is completed.
    * Provide cryptographic proof of mission-bound authority for all leased capabilities.
* **Non-Goals:**
    * Managing persistent user-level permissions; it focus on session-bound agent agency.
    * Replacing existing Zero-Trust (LOWA) protocols; it provides the temporal lease layer on top.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Orchestrator
* **Primary Goal:** Grant a specialist agent "Write" access to a specific file for 5 minutes, ensuring it cannot touch other files or retain access after the edit.
* **The Happy Path (Tasks):**
    1. Parent agent requests a "Write Lease" for the subagent, specifying the Target File and Mission ID.
    2. HLML Provider generates a TPM-signed lease token bound to the Hardware ID and Mission ID.
    3. Subagent receives the lease and executes the tool call through the HLML Gateway.
    4. The Gateway verifies the lease's TPM signature and mission-root alignment.
    5. Upon completion of the sub-task or mission-root signal, the HLML Provider broadcasts a revocation signal.
    6. The hardware root (TPM) invalidates the lease, and any subsequent attempts to use the capability fail.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Parent Agent] --> B[HLML Provider]
        B --> C[TPM / Secure Enclave]
        C --> D[Hardware-Signed Lease]
        D --> E[Subagent]
        E --> F[HLML Gateway]
        F --> G[Target Tool/Resource]
        B -.->|Revocation Signal| F
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(missionID, capabilities, duration) -> LeaseID`: Requests a hardware-signed lease.
    * `hlml.VerifyLease(leaseID) -> Status`: Validates the lease against the hardware root.
    * `hlml.RevokeLease(leaseID) -> Success`: Forcefully invalidates a lease.
* **Data Storage/State:**
    * **Lease Registry:** In-memory store of active leases, indexed by Mission ID and Hardware ID.

## 5. Alternatives Considered
* **Time-Bound JWTs:** Rejected because they can be replayed or leaked. HLML requires hardware-locked (TPM) binding to the specific device.
* **Process-Level Sandboxing:** Complimentary, but insufficient for managing capabilities across a distributed mesh.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Leases are cryptographically tied to the verified mission root and hardware.
* **Observability:** Leases are tracked in the "Sovereignty Audit Log" for real-time compliance monitoring.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
