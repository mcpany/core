# Design Doc: Hardware-Locked Lease (MBHL) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the release of Claude Code v3.2.0, "Mission-Bound Hardware Leases" (MBHL) has become the industry standard for securing high-privilege agent operations. Traditional session-based permissions are too broad and persist longer than necessary. MBHL ensures that sensitive capabilities (e.g., shell access, remote file writes) are cryptographically bound to a specific mission task and are automatically revoked by the hardware root of trust (TPM/Secure Enclave) upon completion.

MCP Any must implement a MBHL Provider to maintain its position as the indispensable infrastructure layer for enterprise-grade autonomous swarms.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement TPM-signed, task-specific capability leases.
    * Automate the lifecycle of high-privilege access, ensuring zero-latency revocation upon task completion.
    * Provide a hardware-attested audit trail for all leased capability usage.
    * Compatibility with the UACO v3.2 and MBHL standards.
* **Non-Goals:**
    * Managing the model's internal reasoning logic.
    * Replacing existing framework-level task management; it provides the security enforcement layer for those tasks.

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Ensure a subagent's `run_shell_command` capability is strictly limited to the "Build & Test" phase and cannot be used for lateral movement.
* **The Happy Path (Tasks):**
    1. Parent agent proposes a "Build & Test" sub-task via the Universal Agent Bus.
    2. MBHL Provider issues a TPM-signed lease for `run_shell_command` specifically for the `task_uuid`.
    3. Specialist agent executes the command over an attested tunnel; the MBHL Provider validates the lease against the active task state.
    4. Upon task completion, the specialist agent returns the result.
    5. MBHL Provider receives the completion signal and immediately invalidates the hardware lease.
    6. Any subsequent attempt to use the shell command with that lease is rejected by the hardware root.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Task Initiation] --> B[MBHL Provider]
        B --> C[TPM/Enclave Lease Signing]
        C --> D[Capability Grant]
        D --> E[Tool Execution]
        E --> F[Lease Validation]
        F --> G[Task Completion]
        G --> H[Automated Hardware Revocation]
    ```
* **APIs / Interfaces:**
    * `mbhl.IssueLease(taskID string, capabilities []string) (LeaseToken, error)`
    * `mbhl.ValidateLease(token LeaseToken, currentTaskID string) bool`
    * `mbhl.RevokeLease(taskID string) error`
* **Data Storage/State:**
    * **Lease Registry:** A kernel-bound, hardware-protected memory region for tracking active task-to-capability mappings.

## 5. Alternatives Considered
* **Time-Bound Leases (TTL):** Rejected because mission durations are non-deterministic; TTLs lead to either premature failure or "zombie" privileges.
* **Hierarchical Intent Leases (HIL):** MBHL is an evolution of HIL, moving from software-defined hierarchical trust to hardware-locked task boundaries.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Leases are non-transferable and tied to a unique hardware-attested identity.
* **Observability:** Visualized in the "Mission Lease Manager" dashboard, showing real-time allocation and revocation events.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
