# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI agents in horizontal meshes (e.g., Claude Code Agent Teams) often require high-privilege capabilities (e.g., sudo, shell access) to complete specific tasks. Traditional time-bound tokens are insufficient because they persist after a task is finished, creating a window for "Capability Squatting" by rogue subagents.

The Hardware-Locked Mission Lease (HLML) Provider is needed to issue TPM-signed, task-specific leases that are cryptographically tied to a mission-root fragment and expire automatically upon task completion, following the Claude Code v3.2.0 standard.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM) capability leases for high-privilege operations.
    * Bind leases to specific, granular Mission-Root Task IDs.
    * Ensure "Atomic Revocation" immediately upon task completion or sub-mission termination.
    * Support hardware-locked non-repudiation for all leased actions.
* **Non-Goals:**
    * Replacing general session authentication.
    * Managing low-level hardware drivers outside of TPM/Secure Enclave interactions.
    * Providing persistent identity (handled by SMI/FSI).

## 3. Critical User Journey (CUJ)
* **User Persona:** Secure Operations Orchestrator
* **Primary Goal:** Grant a specialist agent one-time shell access to deploy a specific Docker container, with zero residual access after completion.
* **The Happy Path (Tasks):**
    1. Parent agent identifies a deployment task and requests a shell lease for a subagent.
    2. HLML Provider generates a TPM-signed lease token bound to the `task_id: deploy_container_x`.
    3. Specialist subagent receives the lease and invokes the shell tool.
    4. MCP Any validates the lease against the hardware root and the active task context.
    5. Subagent completes the deployment and signals task completion.
    6. HLML Provider atomically revokes the lease; subsequent shell calls by the subagent are interdicted.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Parent Agent] --> B(HLML Provider)
        B --> C{TPM/Secure Enclave}
        C -->|Sign Lease| B
        B --> D[Subagent]
        D -->|Present Lease| E(Capability Gate)
        E -->|Verify with HLML| F[High-Privilege Tool]
        D -->|Task Done| B
        B -->|Revoke| E
    ```
* **APIs / Interfaces:**
    * `hlml.RequestLease(targetCapability, taskID) -> LeaseToken`: Requests a hardware-locked lease.
    * `hlml.VerifyLease(token, currentContext) -> bool`: Validates a token against active mission state.
    * `hlml.TerminateTask(taskID)`: Atomically revokes all leases associated with a task.
* **Data Storage/State:**
    * **Lease Registry:** In-memory, hardware-attested store of active leases and their task bindings.

## 5. Alternatives Considered
* **Time-Bound JWTs:** Rejected because they don't support task-completion-based revocation without heavy polling. They remain valid for their duration even if the task is finished.
* **Manual User Approval (HITL) for every call:** Rejected due to "Approval Fatigue" and inability to scale in automated background workers (Dispatch mode).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All leases are hardware-bound; exfiltration of the token is useless without access to the specific node's TPM.
* **Observability:** Integrated with the "Mission Lease Manager" in the UI for real-time tracking of lease lifetimes.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
