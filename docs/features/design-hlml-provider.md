# Design Doc: Hardware-Locked Mission Leases (HLML)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI Agent Teams are increasingly operating with high-privilege capabilities (e.g., shell access, production DB writes). Persistent access is a liability; if a specialized subagent is compromised or drifts from its intent, it can perform unauthorized actions. The "Sovereignty Island" problem also prevents these leases from being recognized across different framework nodes.

HLML provides task-specific, TPM-signed capability leases that are automatically revoked by the hardware root upon mission-root task completion. It ensures that "Just-in-Time" agency is both non-repudiable and time-bound.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, task-specific capability leases.
    * Enforce automatic revocation of leases upon sub-mission completion.
    * Neutralize persistent privilege escalation in specialized subagents.
    * Support "Attestation Translation" for cross-framework lease recognition.
* **Non-Goals:**
    * Managing user-level IAM outside the agent mission scope.
    * Replacing standard OS-level permissions (e.g., sudo); it layers on top of them.

## 3. Critical User Journey (CUJ)
* **User Persona:** Specialized DevOps Agent
* **Primary Goal:** Securely execute a sequence of shell commands to deploy a patch, then lose all shell access immediately.
* **The Happy Path (Tasks):**
    1. Parent agent delegates a "Deploy Patch" task to the DevOps subagent.
    2. HLML Provider issues a hardware-attested lease for `run_shell_command` bound to the "Deploy Patch" task ID.
    3. DevOps subagent invokes the shell tool. MCP Any verifies the HLML lease against the active task ID and hardware signature.
    4. DevOps subagent completes the deployment.
    5. Mission-root signals task completion.
    6. HLML Provider forcefully revokes the hardware lease.
    7. Any subsequent shell calls by the subagent are blocked as "Lease Expired."

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Parent as Parent Agent
        participant HLML as HLML Provider
        participant TPM as Hardware TPM
        participant Sub as Subagent
        participant Tool as MCP Tool (Shell)

        Parent->>HLML: Delegate(TaskID, Capability)
        HLML->>TPM: SignLease(TaskID, Capability)
        TPM-->>HLML: SignedLeaseToken
        HLML-->>Sub: LeaseToken
        Sub->>Tool: Call(LeaseToken, Command)
        Tool->>HLML: Verify(LeaseToken)
        HLML-->>Tool: Valid
        Tool-->>Sub: Result
        Sub->>Parent: TaskComplete(TaskID)
        Parent->>HLML: Terminate(TaskID)
        HLML->>HLML: RevokeLease(TaskID)
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(taskID, capabilities) -> LeaseToken`: Generates a TPM-signed lease.
    * `hlml.VerifyLease(leaseToken) -> bool`: Validates the lease against current mission state and hardware signature.
    * `hlml.RevokeLease(taskID)`: Explicitly terminates a lease.
* **Data Storage/State:**
    * **Lease Registry:** In-memory, hardware-synchronized store of active leases and their task bindings.

## 5. Alternatives Considered
* **Time-based JWTs:** Rejected because they don't account for early task completion and can be replayed if exfiltrated. HLML is task-bound and hardware-locked.
* **OS-level ephemeral users:** Rejected due to high overhead and lack of "Agentic Intent" context in audit logs.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All leases are cryptographically bound to the hardware ID of the node and the mission-root intent.
* **Observability:** Integrated with the "Mission Lease Manager" in the UI for real-time tracking of active vs. revoked leases.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
