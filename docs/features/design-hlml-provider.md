# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agents move toward autonomous task execution in sensitive environments, the risk of persistent privilege escalation has become a primary concern. Static API keys or long-lived tokens are vulnerable to exfiltration and "Credential Squatting."

The Hardware-Locked Mission Lease (HLML) Provider addresses this by issuing TPM-signed, task-specific capability leases that are cryptographically bound to a unique mission-root fragment. These leases expire automatically upon mission completion or detecting unauthorized task patterns.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, mission-bound capability leases for high-risk tools (e.g., shell, DB access).
    * Enforce automated revocation based on mission-root termination signals.
    * Detect and neutralize "Lease-Squatting" exploits via semantic occupancy monitoring.
    * Provide a non-repudiable audit trail for privilege usage.
* **Non-Goals:**
    * Managing low-level OS user permissions (it operates at the agent bus layer).
    * Providing a general-purpose identity provider for human users.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious DevSecOps Agent
* **Primary Goal:** Execute a database schema migration without leaving persistent access tokens in the environment.
* **The Happy Path (Tasks):**
    1. Agent receives a "Mission Root" for a schema migration.
    2. Agent requests a lease for the `db_admin` capability from the HLML Provider.
    3. HLML Provider verifies the mission intent and issues a TPM-signed lease token.
    4. Agent executes the migration tool using the lease.
    5. MCP Any validates the lease signature and mission-binding before each tool call.
    6. Once the migration is complete, the mission-root broadcasts a termination signal.
    7. HLML Provider forcefully revokes the lease across the mesh.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -->|Authorize| B[HLML Provider]
        B -->|Issue Lease| C[Subagent]
        C -->|Tool Call + Lease| D[MCP Any Gateway]
        D -->|Verify Signature| E[TPM/Secure Enclave]
        D -->|Check Occupancy| F[Lease-Squatting Monitor]
        F -->|Validate Intent| A
        D -->|Invoke| G[Secure Tool]
    ```
* **APIs / Interfaces:**
    * `hlml.RequestLease(missionToken, capability) -> LeaseToken`: Requests a hardware-attested lease.
    * `hlml.RevokeLease(leaseID) -> Status`: Forcefully terminates a lease.
    * `hlml.ReportTask(leaseID, taskEntropy) -> void`: Submits occupancy signals for anti-squatting analysis.
* **Data Storage/State:**
    * **Lease Registry:** Encrypted local store of active leases and their mission bounds.
    * **Entropy Baseline:** Statistical model of "normal" task patterns for specific capabilities.

## 5. Alternatives Considered
* **Short-Lived JWTs:** Rejected because they can be replayed if exfiltrated during their valid window. HLML requires per-call hardware attestation.
* **Sudo-style Approvals:** Rejected due to "Approval Fatigue" in high-speed autonomous swarms. HLML provides autonomous governance based on pre-signed manifests.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Integrates with the "Lease-Squatting Detection Middleware" to prevent temporal occupancy attacks.
* **Observability:** Leases and their expiration events are visualized in the "Mission Lease Manager" UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
