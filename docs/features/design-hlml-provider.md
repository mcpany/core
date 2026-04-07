# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents evolve from linear task execution to complex, autonomous swarms, the risk of persistent privilege escalation becomes a critical failure point. Traditional session-based authentication often grants agents broad access that remains active long after the specific reasoning branch has terminated.

The Hardware-Locked Mission Lease (HLML) Provider addresses this by binding agent capabilities (e.g., filesystem writes, network requests, tool execution) to cryptographically signed, mission-specific leases. These leases are anchored to a Trusted Platform Module (TPM) or Secure Enclave, ensuring that authority is both time-bound and mission-bound.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement TPM-signed mission leases for all high-privilege agent operations.
    * Automate lease revocation upon mission-root task completion or sub-mission termination.
    * Provide hardware-attested proof of lease expiration to the supervisor agent.
* **Non-Goals:**
    * Implementing the hardware root of trust itself (relying on host TPM/SEP).
    * Governing low-risk, read-only operations (e.g., public API reads) unless explicitly configured.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Grant a specialist agent `sudo` access only for a specific 5-minute deployment task without risk of persistent backdoors.
* **The Happy Path (Tasks):**
    1. Parent agent proposes a sub-mission with a defined scope (e.g., "Fix nginx config").
    2. HLML Provider issues a hardware-signed lease fragment tied to the task ID and a 10-minute TTL.
    3. Specialist agent attempts to execute `mcp_any.run_shell_command`.
    4. Gateway validates the shell command against the active HLML lease and hardware signature.
    5. Specialist agent completes task; parent agent signals "Mission End".
    6. HLML Provider broadcasts an immediate hardware revocation signal, invalidating the lease across the mesh.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        Agent->>HLML_Provider: Request Mission Lease (TaskID, Scopes)
        HLML_Provider->>TPM: Sign Lease Fragment (Nonce, MissionHash)
        TPM-->>HLML_Provider: Signed Lease
        HLML_Provider-->>Agent: Capability Token (Lease-Bound)
        Agent->>Gateway: Execute High-Privilege Tool (Token)
        Gateway->>HLML_Provider: Validate Lease
        HLML_Provider-->>Gateway: OK (Hardware Verified)
        Gateway-->>Agent: Tool Result
        Agent->>HLML_Provider: Signal Mission Completion
        HLML_Provider->>ARL: Publish Revocation
    ```
* **APIs / Interfaces:**
    * `POST /v1/leases/request`: Initiate a new mission-bound lease.
    * `POST /v1/leases/revoke`: Forcefully terminate an active lease.
* **Data Storage/State:**
    * Lease metadata stored in a memory-mapped, encrypted region (ZCMB).
    * Revocation signals persisted in the Attestation Revocation List (ARL).

## 5. Alternatives Considered
* **Short-Lived JWTs**: Rejected due to vulnerability to replay attacks and lack of hardware-bound mission anchoring.
* **Kernel-Bound Process Isolation**: Complementary, but rejected as a standalone solution because it doesn't provide the cryptographic provenance required for cross-node swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All leases require monotonic counter validation to prevent replay attacks. Hardware signatures ensure non-repudiation.
* **Observability:** Lease issuance, usage, and revocation are logged to the Hardware-Attested Audit Trail.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
