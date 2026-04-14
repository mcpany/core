# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agent swarms become more autonomous and distributed, the risk of "Capability Squatting"—where a subagent retains access to sensitive tools (like shell execution or database writes) beyond its immediate task—has become a critical vulnerability. Currently, capability tokens are often session-bound, providing too much lateral movement for a potentially compromised or hallucinating specialist agent.

The HLML Provider aims to solve this by anchoring agent agency to a hardware-attested, task-specific lease. This ensures that every high-privilege operation is cryptographically bound to a unique mission-root fragment and automatically expires upon task completion or mission-root termination.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, time-bound, and task-scoped capability leases.
    * Enforce absolute lease expiration upon mission-root signal or task exit.
    * Provide non-repudiable audit logs of lease issuance and usage.
    * Integrate with OpenClaw v3.6.1 and Claude Code v3.2.0 standards.
* **Non-Goals:**
    * Managing the underlying hardware (TPM/Secure Enclave) directly (delegated to low-level platform drivers).
    * Defining the high-level security policies (delegated to the Policy Firewall).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Ensure a specialist "Database Migrator" agent can only write to the production database for the duration of a specific verified migration script.
* **The Happy Path (Tasks):**
    1. The supervisor agent initiates a "Migration Mission" and requests an HLML for the specialist.
    2. The HLML Provider verifies the supervisor's mission-root intent and the specialist's reputation.
    3. The HLML Provider issues a TPM-signed lease bound to `db:write` and the specific Migration Task ID.
    4. The specialist executes the migration; every call is validated against the hardware-locked lease.
    5. Upon script completion, the supervisor signals task end; the HLML Provider forcefully revokes the lease across the mesh.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        participant S as Supervisor Agent
        participant H as HLML Provider (TPM-Bound)
        participant P as Policy Firewall
        participant T as Specialist Agent
        participant R as Resource (e.g., DB)

        S->>H: Request Lease (Task_ID, Capability)
        H->>P: Validate Request vs Mission Root
        P-->>H: Approved
        H->>H: Generate TPM-Signed Lease Token
        H-->>S: Lease Token
        S->>T: Delegate Task with Lease
        T->>R: Call Resource + Lease Token
        R->>H: Verify Hardware Signature & Scope
        H-->>R: Valid
        R-->>T: Success
        S->>H: Task Complete / Revoke
        H->>H: Mark Lease Expired in Local ARL
    ```
* **APIs / Interfaces:**
    * `POST /v1/leases`: Request a new mission-bound lease.
    * `GET /v1/leases/{id}/verify`: Verify a lease signature and current status.
    * `DELETE /v1/leases/{id}`: Forcefully revoke a lease.
* **Data Storage/State:**
    * Leases are stored in an encrypted SQLite sidecar, with the master key held in the hardware enclave.
    * Active leases are mirrored in a high-speed, in-memory Attestation Revocation List (ARL).

## 5. Alternatives Considered
* **Short-lived JWTs:** Rejected because they are not hardware-bound and can be replayed if exfiltrated during their (even short) lifetime.
* **Standard MCP Scopes:** Rejected as they lack the temporal "mission-bound" lifecycle and cryptographic non-repudiation required for production swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All lease verification occurs at the "Pipe" or "Resource" layer, independent of the agent's internal state.
* **Observability:** Every lease lifecycle event (Request, Issue, Usage, Revocation) is logged with a hardware-attested timestamp.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
