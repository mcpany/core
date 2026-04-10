# Design Doc: Mission-Bound Hardware Leases (MBHL) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents gain more autonomy, the risk of persistent privilege escalation has become a critical threat. Traditional capability-based tokens are often session-bound, meaning if an agent session is compromised, the attacker retains all granted privileges until the session expires. The Azure DevOps MCP bypass (CVE-2026-32211) highlighted that authentication and authorization must be more granular and physically tied to the mission lifecycle.

The Mission-Bound Hardware Leases (MBHL) Provider (formerly HLML) solves this by issuing TPM-signed, mission-fragment-specific capability leases. These leases are physically revoked by the hardware root at the end of a specific mission, ensuring that high-privilege access is strictly time-bound and non-persistent.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM/SEP) leases for high-privilege tool access.
    * Bind leases to specific mission-root fragments and task IDs.
    * Enforce automated physical revocation of leases upon mission completion or termination.
    * Integrate with the MLE Gateway for mandatory lease verification on high-impact tools.
* **Non-Goals:**
    * Managing low-level TPM driver implementation (assumes existing hardware abstraction).
    * Replacing standard session tokens for low-risk operations.
    * Providing user-level identity management (it manages agentic capability leases).

## 3. Critical User Journey (CUJ)
* **User Persona:** DevSecOps Engineer
* **Primary Goal:** Ensure that an autonomous "Refactor Agent" can only use `run_shell_command` for the duration of a specific code refactoring task, and that access is revoked immediately after the PR is created.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a "Refactor" mission and requests an MBHL for the `run_shell_command` capability.
    2. MBHL Provider verifies the mission-root intent and requests a TPM-signed lease.
    3. The TPM generates a lease token cryptographically bound to the mission fragment.
    4. The agent executes the refactoring using the leased capability.
    5. Once the PR is created, the mission-root sends a "Mission Complete" signal.
    6. MBHL Provider notifies the TPM to revoke the lease.
    7. Any subsequent attempts to use the lease are blocked at the hardware level.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -->|Request Lease| B[MBHL Provider]
        B -->|Attestation Request| C[Hardware Root (TPM/SEP)]
        C -->|Issue Signed Lease| B
        B -->|Grant Lease| D[Specialist Agent]
        D -->|Tool Call + Lease| E[MLE Gateway]
        E -->|Verify Lease| B
        B -->|Hardware Validation| C
        E -- Valid --> F[High-Privilege Tool]
        A -->|Mission End| B
        B -->|Revoke| C
    ```
* **APIs / Interfaces:**
    * `mbhl.RequestLease(missionToken, capability, duration) -> LeaseToken`: Requests a hardware-signed lease.
    * `mbhl.VerifyLease(leaseToken) -> bool`: Validates the lease against the hardware root.
    * `mbhl.RevokeLease(leaseToken) -> void`: Forcefully terminates a lease.
* **Data Storage/State:**
    * **Lease Registry:** Secure, in-memory registry of active leases and their mission-root bindings.
    * **Hardware Monotonic Counters:** Used to prevent lease replay attacks.

## 5. Alternatives Considered
* **Short-Lived JWTs:** Rejected because they are software-bound and susceptible to "Time-of-Check to Time-of-Use" (TOCTOU) exploits if the agent process is hijacked but the clock is manipulated. MBHL is hardware-enforced.
* **Manual HITL:** Rejected due to the latency requirements of high-frequency machine-speed swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MBHL is a foundational component of the Zero-Trust mesh. It ensures that "Least Privilege" is enforced temporally as well as logically.
* **Observability:** Integrated with the "Mission Lease Manager" in the UI for real-time tracking of active hardware leases.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation (Transitioned from HLML concept).
