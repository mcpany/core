# Design Doc: Hardware-Locked Mission Leases (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Autonomous subagents (e.g., specialists in a Claude Code Agent Team) often inherit broad permissions from their parent, leading to "Privilege Squatting" where a compromised specialist can execute unauthorized actions long after its specific task is complete.

The Hardware-Locked Mission Leases (HLML) Provider is required to issue cryptographically signed, task-specific capability leases that are bound to the hardware root (TPM) and automatically expire upon mission completion.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed capability tokens for specific subagent tasks.
    * Enforce monotonic lease expiration based on hardware clock signals.
    * Facilitate "Recursive Revocation" where terminating a parent mission automatically invalidates all child leases.
    * Provide a verifiable audit trail of lease issuance and reclamation.
* **Non-Goals:**
    * Managing user-level authentication (handled by LOWA).
    * Providing long-term persistent storage permissions.
    * Replacing the Policy Firewall; it provides the temporal and hardware binding for policies.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Grant a "File Refactoring" subagent write access only to the `/src/utils` directory for a maximum of 30 minutes.
* **The Happy Path (Tasks):**
    1. Parent agent requests a Mission Lease for a new specialist subagent.
    2. HLML Provider generates a lease manifest specifying the scope (`fs:write:/src/utils`) and duration (1800s).
    3. The manifest is signed by the local TPM and issued as a "Hardware-Locked Lease."
    4. The subagent presents the lease to the MCP Any gateway for each tool call.
    5. The gateway verifies the TPM signature and checks the hardware clock for expiration.
    6. Once the subagent signals "Task Complete," the HLML Provider forcefully revokes the lease.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Parent Agent] --> B[HLML Provider]
        B --> C[TPM Root of Trust]
        C --> B
        B --> D[Hardware-Locked Lease]
        D --> E[Gateway Enforcement]
        E --> F[Tool Execution]
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(scope, duration, parentMissionID) -> LeaseToken`: Issues a new hardware-bound lease.
    * `hlml.VerifyLease(leaseToken) -> bool`: Validates signature and expiration via hardware root.
    * `hlml.RevokeMission(missionID)`: Forcefully expires all leases associated with a mission.
* **Data Storage/State:**
    * **Lease Registry:** SQLite-backed store for active lease metadata, linked to the TPM's monotonic counters.

## 5. Alternatives Considered
* **Standard JWTs with Short TTL:** Rejected because they are software-only and can be replayed if the system clock is manipulated. HLML requires hardware-bound temporal integrity.
* **Linux Capabilities (setcap):** Too low-level and does not support the granular "Agentic Scopes" (e.g., specific file paths or API endpoints) required by MCP Any.

## 6. Cross-Concern Issues
* **Security (Zero Trust):** Integrates with the ARL Provider for mesh-wide revocation.
* **Observability:** Leases are tracked in the "Hardware Lease Manager" UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
