# Design Doc: Hardware-Locked Mission Leases (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The emergence of "Registry Persistence" exploits and the disclosure of unauthorized lateral movement in distributed swarms reveal that time-bound privilege is no longer sufficient. High-privilege operations (e.g., shell execution, filesystem writes) must be cryptographically anchored to a specific, hardware-attested mission-root task.

The Hardware-Locked Mission Leases (HLML) Provider ensures that subagent capabilities are issued as TPM-signed, task-specific leases that expire automatically and irrecoverably upon the completion of the authorized mission fragment.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/SEP) capability leases bound to specific mission IDs.
    * Enforce irrecoverable expiration of leases upon task completion or mission-root termination.
    * Neutralize "Privilege Squatting" by specialist agents.
    * Support hardware-locked re-attestation for mission continuity.
* **Non-Goals:**
    * Managing network-layer permissions (handled by AMT/LOWA).
    * Providing long-term persistent storage for agent secrets.

## 3. Critical User Journey (CUJ)
* **User Persona:** DevOps Automation Swarm
* **Primary Goal:** Authorize a subagent to execute a single `terraform apply` command without granting persistent shell access.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a "Cloud Infrastructure Update" mission.
    2. HLML Provider issues a TPM-signed lease for the `run_shell_command` tool, bound to `mission_id: CI-99` and `allowlist: ["terraform apply"]`.
    3. Specialist subagent receives the lease and executes the command.
    4. Upon command completion, the subagent signals task success.
    5. HLML Provider invalidates the hardware-locked lease.
    6. Any subsequent attempt by the subagent to use `run_shell_command` is blocked by the TPM-resident enforcement layer.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant PA as Parent Agent
        participant HLML as HLML Provider
        participant TPM as Hardware Enclave
        participant SA as Specialist Agent
        PA->>HLML: Request Lease(TaskID, Capability)
        HLML->>TPM: Sign Lease(TaskID, Capability, Nonce)
        TPM-->>HLML: Hardware-Locked Token
        HLML-->>SA: Provision Lease
        SA->>SA: Execute Task
        SA->>HLML: Terminate Task
        HLML->>TPM: Invalidate Lease(Nonce)
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(missionID, capabilityManifest) -> LeaseToken`: Generates a hardware-signed lease.
    * `hlml.VerifyLease(leaseToken) -> bool`: Enclave-bound validation of lease validity.
    * `hlml.RevokeLease(missionID)`: Explicitly terminates all leases for a mission branch.
* **Data Storage/State:**
    * **Lease Registry:** Enclave-resident monotonic counter and hash-map of active nonces.

## 5. Alternatives Considered
* **JWT-based Time-Limited Tokens:** Rejected because they can be replayed until expiry and lack hardware-bound non-repudiation.
* **Unix-level Sudo Leases:** Rejected because they do not understand agentic mission boundaries and are difficult to rotate at sub-second frequencies.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Leases are cryptographically tied to the unique hardware signature of the executing node.
* **Observability:** Integrated with the "Mission Lease Manager" UI for real-time visualization of active and expired leases.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
