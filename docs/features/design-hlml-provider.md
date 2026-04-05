# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI agents are increasingly being deployed in "Agent Teams" (e.g., Claude Code) where specialist agents are delegated high-privilege tasks like `run_shell_command` or `write_file`. Standard long-lived capability tokens are insecure in this context, as a compromised specialist could retain its privileges indefinitely.

The Hardware-Locked Mission Lease (HLML) Provider is required to issue TPM-signed, task-specific capability leases that expire automatically upon mission-root task completion, ensuring that "Excessive Agency" is contained within a verified temporal and logical boundary.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/Secure Enclave) capability leases for subagent delegations.
    * Bind every lease to a specific mission-root task ID and monotonic counter.
    * Enforce automatic revocation of leases upon task termination or mission-root completion.
    * Support "Just-in-Time" privilege escalation for high-trust tool calls.
* **Non-Goals:**
    * Managing user-level IAM outside the agent gateway.
    * Providing persistent long-term credentials for non-agent services.
    * Replacing existing role-based access control (RBAC); it adds a temporal, mission-bound constraint.

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Grant a "Refactoring Agent" permission to edit files only while its current sub-task is active.
* **The Happy Path (Tasks):**
    1. Parent agent delegates a "Refactor `main.py`" task to a specialist subagent.
    2. HLML Provider intercepts the delegation and requests a hardware-signed lease for the `write_file` capability.
    3. The TPM issues a mission-bound lease token cryptographically linked to Task ID `REF-123`.
    4. Specialist agent receives the lease and executes the file edits.
    5. Parent agent marks task `REF-123` as "Complete."
    6. HLML Provider receives the completion signal and broadcasts a hardware-locked revocation for the `REF-123` lease.
    7. Any subsequent attempts by the specialist to use the lease are blocked by the gateway.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        ParentAgent->>HLML: Delegate(TaskID, Capability)
        HLML->>TPM: SignLease(TaskID, CapHash)
        TPM-->>HLML: MissionLeaseToken
        HLML-->>SubAgent: GrantLease(Token)
        SubAgent->>Gateway: ToolCall(Token, Args)
        Gateway->>HLML: Validate(Token)
        HLML-->>Gateway: OK
        ParentAgent->>HLML: Complete(TaskID)
        HLML->>ARL: Revoke(TokenID)
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(missionRoot, taskID, capability) -> LeaseToken`: Generates a hardware-attested lease.
    * `hlml.ValidateLease(token) -> boolean`: Verifies the token and its mission status.
    * `hlml.RevokeLease(taskID)`: Forcefully expires all leases associated with a task.
* **Data Storage/State:**
    * **Lease Registry:** In-memory store of active mission-bound leases.
    * **TPM State:** Hardware-enforced monotonic counters to prevent lease replay.

## 5. Alternatives Considered
* **Short-Lived JWTs:** Rejected because they rely on logical expiration times. HLML requires hardware-attested task-completion triggers for stronger security.
* **Standard OAuth Scopes:** Rejected because they lack the granularity to bind permissions to a specific reasoning mission branch.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All leases are hardware-locked; exfiltration of the token to another device is neutralized.
* **Observability:** Integrated with the "Mission Lease Manager" in the UI for real-time tracking of active vs. revoked leases.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
