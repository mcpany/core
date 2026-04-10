# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI agents operating in autonomous swarms often require high-privilege tool access (e.g., `run_shell_command`, `db_admin_query`) to complete their tasks. Current security models rely on persistent capability tokens, which create a significant risk of lateral movement if a subagent is compromised.

The Hardware-Locked Mission Lease (HLML) Provider addresses this by issuing TPM-signed, task-specific leases that are cryptographically bound to the hardware root and expire automatically upon mission-root completion. This ensures that subagent authority is strictly "Just-in-Time" and non-persistent.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM) capability leases for high-risk tool calls.
    * Bind leases to specific mission-root task IDs.
    * Enforce automated revocation of leases upon subagent termination or mission completion.
    * Neutralize persistent privilege escalation via "Capability Squatting."
* **Non-Goals:**
    * Replacing the per-call Zero-Trust authorization logic.
    * Managing low-level hardware drivers for TPMs.
    * Defining the business logic of mission completion (handled by the Orchestrator).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Compliance Officer
* **Primary Goal:** Ensure that a "DB Specialist" subagent can only access production tables for the duration of a specific optimization task and loses all access immediately after.
* **The Happy Path (Tasks):**
    1. The Supervisor Agent requests a "DB Admin" lease for the Specialist Subagent, referencing Task ID `M-102`.
    2. The HLML Provider generates a TPM-signed lease token containing the capability, hardware fingerprint, and Task ID.
    3. The Specialist Subagent invokes the `db_query` tool, providing the HLML token.
    4. The Tool Gateway verifies the TPM signature and checks that Task `M-102` is still active.
    5. Once the Specialist Subagent signals task completion, the HLML Provider broadcasts a hardware-locked revocation signal.
    6. Subsequent tool calls by the subagent are immediately blocked, even if the session token is still valid.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Supervisor Agent] -->|Request Lease| B[HLML Provider]
        B -->|Sign with TPM| C[TPM / Secure Enclave]
        C -->|Lease Token| B
        B -->|Issue Lease| D[Specialist Subagent]
        D -->|Invoke Tool + Token| E[Tool Gateway]
        E -->|Verify TPM Sig| F[Hardware Root]
        E -->|Check Task Status| G[Mission Registry]
        F & G -->|Authorized| E
        E -->|Execute| H[High-Privilege Tool]
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(capabilities, missionID) -> HLMLToken`: Requests a hardware-locked lease.
    * `hlml.RevokeLease(tokenID)`: Forcefully invalidates a lease.
    * `hlml.CheckIntegrity(tokenID) -> bool`: Verifies TPM signature and mission-binding.
* **Data Storage/State:**
    * **Lease Registry:** SQLite-backed store for active leases and their mission-phase status.
    * **Hardware Root Keys:** Securely managed by the host OS TPM/SEP.

## 5. Alternatives Considered
* **Time-Bound JWTs:** Rejected because they can still be "squatted" until expiration. HLML provides event-driven, hardware-locked revocation.
* **Kernel-Bound FD Passing:** Effective for local files, but HLML scales to remote tool calls and cross-node meshes.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tokens are hardware-bound and origin-locked, preventing replay attacks across disparate devices.
* **Observability:** Lease issuance and revocation events are logged to the `Hardware-Attested Audit Log` and visualized in the `Mission Lease Manager`.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
