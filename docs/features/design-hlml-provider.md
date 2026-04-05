# Design Doc: Hardware-Locked Mission Leases (HLML)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Current AI agent frameworks often grant broad, long-lived permissions to tools (e.g., shell access, database writes) during a session. This creates a significant "Persistence Gap" where a compromised specialist subagent can continue to execute unauthorized actions after its specific task is finished.

The Hardware-Locked Mission Lease (HLML) provider solves this by binding capability tokens to specific, hardware-attested mission-root fragments. These leases are TPM-signed and task-specific, ensuring that privileges are automatically revoked by the hardware root once the agent signals task completion or when a mission-bound timeout is reached.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/Secure Enclave) capability leases bound to specific mission IDs.
    * Enforce automated revocation of leases upon task termination or mission-root expiration.
    * Neutralize persistent privilege escalation by ensuring subagents cannot "squat" on high-trust capabilities.
    * Provide non-repudiable audit logs of lease issuance and reclamation.
* **Non-Goals:**
    * Managing the underlying OS-level permissions (e.g., sudoers files); HLML acts as the application-layer gatekeeper.
    * Replacing the Ephemeral Privilege Manager (EPM); HLML provides the hardware-locked attestation layer for EPM leases.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise DevOps
* **Primary Goal:** Grant a "Refactor Specialist" agent temporary shell access to only one repository for exactly the duration of a PR generation.
* **The Happy Path (Tasks):**
    1. Parent agent delegates a "Refactor" task to a subagent via the A2A hub.
    2. HLML Provider intercepts the delegation and generates a TPM-signed "Mission Lease" for the `run_shell_command` tool.
    3. The lease is bound to the unique `TaskID` and `MissionRootID`.
    4. Subagent executes necessary shell commands; each call is validated against the hardware-signed lease.
    5. Subagent submits the PR and signals task completion.
    6. HLML Provider receives the completion signal and immediately broadcasts a hardware-locked revocation signal to the AMT Broker and Tool Gateway.
    7. Any subsequent attempts by the subagent to use shell tools are blocked.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] --> B[A2A Hub]
        B --> C[HLML Provider]
        C --> D[TPM Signature Engine]
        D --> E[Hardware-Locked Lease]
        E --> F[Tool Gateway]
        F --> G[Revocation Signal]
        G --> C
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(missionID, toolScope, duration) -> LeaseToken`: Generates a TPM-signed lease.
    * `hlml.ValidateLease(leaseToken, missionID) -> bool`: Hardware-verified check for active privileges.
    * `hlml.RevokeLease(missionID)`: Forcefully expires all leases bound to a mission.
* **Data Storage/State:**
    * **Lease Registry:** SQLite-backed store of active lease metadata, linked to hardware monotonic counters.

## 5. Alternatives Considered
* **Time-Bound JWTs:** Rejected because they can be "replayed" if exfiltrated before expiration. HLML requires hardware-bound attestation for every validation check.
* **Purely OS-Level Sandboxing:** Rejected because it lacks the semantic awareness of "Mission Roots" and task-specific lifecycles.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Leases are cryptographically tied to the environment ID (EAP) to prevent "Trace Replay" across disparate sessions.
* **Observability:** Visualized in the "Mission Lease Manager" dashboard with real-time status and expiration countdowns.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
