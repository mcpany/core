# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Current AI agent privilege models often rely on session-persistent tokens that, if leaked, allow for unauthorized tool execution beyond the intended task. As swarms become more autonomous, the risk of "Privilege Squatting" by specialist subagents is increasing.

The Hardware-Locked Mission Lease (HLML) Provider addresses this by issuing TPM-signed, task-specific capability leases. These leases are cryptographically bound to a specific mission-root task and expire automatically, ensuring that subagent agency is strictly time-bound and non-repudiable.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, task-specific capability leases for subagent tool access.
    * Enforce automatic lease revocation upon mission-root task completion.
    * Provide a non-repudiable audit trail of privilege escalation and de-escalation.
    * Integrate with the Mission-Root Continuity Provider (MRCP) for lease state persistence.
* **Non-Goals:**
    * Managing the underlying LLM token budgets (handled by the RBF).
    * Providing long-term persistent credentials for human users.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Developer
* **Primary Goal:** Ensure a "File-Searching" subagent cannot execute "File-Deletion" tools even if it manages to inherit a broader capability set from its parent.
* **The Happy Path (Tasks):**
    1. Parent agent requests a sub-mission for file searching.
    2. HLML Provider issues a lease limited to `read_file` and `list_files` tools, signed by the host TPM.
    3. The subagent attempts to call `delete_file`.
    4. The HLML middleware intercepts the call, verifies the lease doesn't contain the `delete_file` capability, and blocks the execution.
    5. Once the search task is complete, the HLML Provider broadcasts a revocation signal, and the lease is invalidated.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -->|Authorize Task| B[HLML Provider]
        B -->|Sign Lease| C[TPM / Secure Enclave]
        C -->|Issued Lease| D[Subagent Session]
        D -->|Tool Call + Lease| E[Capability Guard]
        E -->|Validate| F[Tool Execution]
    ```
* **APIs / Interfaces:**
    * `hlml.RequestLease(taskID, capabilities) -> LeaseToken`: Requests a hardware-signed lease.
    * `hlml.ValidateLease(leaseToken, action) -> Bool`: Validates a tool call against the active lease.
    * `hlml.RevokeLease(taskID)`: Forcefully invalidates all leases associated with a task.
* **Data Storage/State:**
    * **Lease Registry:** In-memory, high-speed store for active leases.
    * **TPM Key Store:** Hardware-locked storage for signing keys.

## 5. Alternatives Considered
* **JWT-based Scoped Tokens:** Rejected because they can be replayed or stolen from memory. HLML requires hardware-bound attestation for every validation step.
* **Static Permission Allow-Lists:** Rejected because they don't handle the dynamic, hierarchical nature of autonomous agent swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Leases are mission-bound and non-transferable between agents.
* **Observability:** Lease lifecycle events are logged to the "Sovereignty Audit Log."

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
