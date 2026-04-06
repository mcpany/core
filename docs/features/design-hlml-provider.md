# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI agent swarms are increasingly operating with high-privilege tools (e.g., shell access, production database writes). Standard session-based tokens are often too broad and long-lived, creating a significant security risk if an agent session is hijacked. The emergence of Claude Code's "Mission-Bound Hardware Leases" (MBHL) highlights a shift toward per-task, hardware-bound authorization.

The Hardware-Locked Mission Lease (HLML) Provider is required to issue and manage TPM-signed, task-specific capability leases that expire automatically upon mission-root completion, ensuring the principle of least privilege is enforced at the hardware level.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM) leases for specific agent capabilities.
    * Bind leases to a unique "Mission-Root" and a specific task ID.
    * Enforce automatic lease revocation upon task completion or mission termination.
    * Support "Zero-Knowledge Lease Auditing" where an auditor can verify lease validity without accessing raw context.
* **Non-Goals:**
    * Replacing general authentication (OAuth/JWT); HLML adds a hardware-bound authorization layer.
    * Managing the LLM's internal reasoning process directly.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Security Swarm Orchestrator
* **Primary Goal:** Grant an autonomous subagent temporary `sudo` access to fix a server configuration without leaving a persistent privileged session.
* **The Happy Path (Tasks):**
    1. Parent agent identifies a task requiring `sudo` and delegates it to a specialist subagent.
    2. Subagent requests a mission lease from the HLML Provider, specifying the mission-root token and the `sudo:fix_config` capability.
    3. HLML Provider verifies the request against the "Mission Manifest" and generates a TPM-signed lease token.
    4. The specialist agent calls the tool, providing the lease token.
    5. The tool validator verifies the hardware signature and checks the lease's monotonic counter.
    6. Once the task is marked "Complete" on the shared Blackboard, the HLML Provider broadcasts a revocation signal.
    7. Any subsequent attempts to use the same lease token are rejected by the hardware root.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Agent] -->|Request Lease| B[HLML Provider]
        B -->|Verify Manifest| C[Mission Manifest]
        B -->|Sign Token| D[TPM / Secure Enclave]
        D -->|Lease Token| A
        A -->|Execute Tool| E[Tool Validator]
        E -->|Check Hardware Proof| D
        E -->|Verify Task Context| F[Blackboard]
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(missionRoot, capability, taskID) -> LeaseToken`: Requests a new hardware-bound lease.
    * `hlml.ValidateLease(leaseToken) -> bool`: Verifies the hardware signature and mission-binding.
    * `hlml.RevokeLease(taskID)`: Forcefully expires a lease upon task completion.
* **Data Storage/State:**
    * **Lease Registry:** SQLite-backed store for active lease metadata, keyed by TPM monotonic counters.
    * **Task-Lease Mapping:** Linkages between active tasks on the Blackboard and their associated hardware leases.

## 5. Alternatives Considered
* **Short-Lived JWTs:** Rejected because JWTs are stored in memory and can be exfiltrated. HLML tokens are bound to the hardware's TPM, making them non-transferable to other devices.
* **Manual HITL Approval for Every Call:** Rejected due to "Approval Fatigue" and the latency requirements of autonomous swarms. HLML provides automated, hardware-locked safety.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All leases are cryptographically linked to the initiating user's hardware identity.
* **Observability:** Integrated with the "Mission Lease Manager" in the UI for real-time tracking of active vs. expired leases.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
