# Design Doc: Mission-Bound Hardware Lease (MBHL) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
The proliferation of high-privilege AI agents operating in autonomous swarms has exposed a critical vulnerability: persistent privilege escalation. Currently, once an agent is granted a capability (e.g., shell access), it often retains that access for the duration of a session, or longer. The Mission-Bound Hardware Lease (MBHL) Provider addresses this by binding capabilities to specific, hardware-attested mission-root tasks.

This system ensures that high-privilege tools are only accessible when the agent is actively working on a verified task and that access is revoked automatically by the hardware root (TPM) upon task completion or mission termination.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, task-specific capability leases.
    * Enforce cryptographic binding between a lease and a specific Mission-Root Task ID.
    * Automate the revocation of leases upon task completion or timeout.
    * Provide a non-repudiable audit trail of hardware-locked privilege usage.
* **Non-Goals:**
    * Managing low-level TPM driver interactions (delegated to the hardware abstraction layer).
    * Replacing identity providers; it works in conjunction with FSI and SMI.
    * Defining the tasks themselves (managed by the mission root).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Architect
* **Primary Goal:** Ensure a subagent refactoring a specific module can only execute shell commands for that specific refactoring task.
* **The Happy Path (Tasks):**
    1. The Mission-Root agent defines a task "Refactor Auth Module" and requests a shell lease for a specialist subagent.
    2. The MBHL Provider generates a TPM-signed lease containing the Task ID and the shell capability.
    3. The specialist subagent presents the lease to the MBHL-aware Command Adapter.
    4. The Adapter verifies the lease against the TPM root and checks the current Mission State to ensure Task ID "Refactor Auth Module" is still active.
    5. The subagent executes authorized shell commands.
    6. Upon task completion, the Mission Root signals the MBHL Provider.
    7. The MBHL Provider invalidates the lease in the hardware-locked session state.
    8. Subsequent attempts by the specialist to use the shell tool are rejected.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        MR[Mission Root] -->|Request Lease| MBHL[MBHL Provider]
        MBHL -->|Sign with TPM| TPM[Hardware TPM/Enclave]
        MBHL -->|Issue Lease| SA[Subagent]
        SA -->|Execute with Lease| CA[Command Adapter]
        CA -->|Verify Hardware Signature| MBHL
        CA -->|Check Task Status| BC[Blackboard/Context]
        CA -->|Execute| OS[Host System]
    ```
* **APIs / Interfaces:**
    * `mbhl.IssueLease(taskID, capabilities, ttl) -> LeaseToken`: Generates a hardware-signed lease.
    * `mbhl.VerifyLease(leaseToken) -> (valid, metadata)`: Validates the token and returns bound task info.
    * `mbhl.RevokeLease(taskID)`: Explicitly terminates a lease.
* **Data Storage/State:**
    * **Lease Registry:** Encrypted local SQLite store tracking active lease fingerprints and their bound task IDs.
    * **Hardware State:** Monotonic counters in the TPM to prevent lease replay.

## 5. Alternatives Considered
* **Time-Bound JWTs:** Rejected because they can be exfiltrated and used on other machines. MBHL requires local hardware attestation.
* **Standard RBAC:** Rejected because it is too coarse-grained and does not provide automatic, task-linked revocation.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory origin validation (LOWA) is required to request a lease.
* **Observability:** Integrated with the "Mission Lease Manager" in the UI for real-time tracking of active vs. expired leases.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
