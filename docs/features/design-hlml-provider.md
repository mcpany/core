# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI agents in distributed swarms often operate with persistent high-privilege access, creating a massive attack surface. If a specialist subagent is compromised, it can use its long-lived credentials to exfiltrate data or modify the host system outside its intended task. Standard time-bound JWTs are insufficient because they lack hardware-bound non-repudiability.

The Hardware-Locked Mission Lease (HLML) Provider addresses this by issuing TPM-signed, mission-bound capability leases that expire automatically upon task completion and are cryptographically tied to the agent's specific execution environment.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/Secure Enclave) capability leases for agent tasks.
    * Bind leases to specific mission-root fragments to prevent lateral privilege movement.
    * Automate lease revocation upon task completion signals.
    * Support Segregated State Lease (SSL) to isolate mission-local memory shards.
* **Non-Goals:**
    * Replacing general-purpose OIDC or IAM systems.
    * Managing human user authentication; HLML is strictly for Non-Human Identities (NHI).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Orchestrator
* **Primary Goal:** Grant a subagent temporary access to write to a specific repository folder without exposing the entire disk.
* **The Happy Path (Tasks):**
    1. Orchestrator requests a "WriteLease" for a subagent tied to Mission-Root ID `mr-99`.
    2. HLML Provider generates a TPM-signed lease containing the resource scope (`fs:write:./docs/`) and Mission ID.
    3. The subagent presents this lease to the MCP Any Filesystem Tool.
    4. The Tool verifies the lease signature and Mission ID against the hardware root.
    5. Subagent completes the task and signals `MissionComplete`.
    6. HLML Provider invalidates the lease and triggers SSL-cleanup to purge the associated scratchpad memory.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Orchestrator] -->|Request Lease| B[HLML Provider]
        B -->|Sign with TPM| C[Mission Lease]
        C --> D[Subagent]
        D -->|Present Lease| E[MCP Any Tool]
        E -->|Verify| B
        D -->|Task Done| B
        B -->|Revoke| F[SSL Memory Shard]
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(missionRoot, capabilities, duration) -> LeaseToken`
    * `hlml.VerifyLease(leaseToken, currentAction) -> Bool`
    * `hlml.TerminateMission(missionRoot) -> Status`
* **Data Storage/State:**
    * **Lease Registry:** SQLite-backed store of active leases and their cryptographic hardware bindings.

## 5. Alternatives Considered
* **Short-lived JWTs:** Rejected because they can be "replayed" from any environment if stolen. HLML requires the physical TPM present on the original node.
* **Unix Sudo with timeouts:** Rejected as it doesn't scale to distributed swarms or high-density containerized agents.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All lease requests must originate from a verified mission root.
* **Observability:** Integrated with the "Mission Lease Manager" UI for real-time tracking of lease lifetimes and revocation events.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
