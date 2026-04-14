# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of "Agent Teams" and high-privilege autonomous operations (e.g., Claude Code), the risk of persistent privilege escalation has become a primary security concern. Standard capability-based tokens are often long-lived or session-bound, making them targets for exfiltration. Hardware-Locked Mission Leases (HLML) address this by tying tool capabilities to a specific, hardware-attested (TPM) lease that is cryptographically bound to a unique mission-root task.

The HLML Provider ensures that privileges are not only "Just-in-Time" but also "Mission-Locked," expiring automatically and irreversibly upon task completion.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, task-specific capability leases for high-risk tool calls.
    * Enforce "Recursive Lease Verification" for subagent meshes.
    * Provide automated, hardware-locked revocation of leases upon mission completion.
    * Maintain a non-repudiable audit trail of lease lineage.
* **Non-Goals:**
    * Managing general user authentication.
    * Replacing software-based policy engines (Rego/CEL); it provides the hardware-enforced root of trust.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious DevSecOps Agent
* **Primary Goal:** Execute a `terraform apply` command using a temporary hardware lease that is revoked immediately after the cloud resource is created.
* **The Happy Path (Tasks):**
    1. The primary agent requests a mission-root lease for a Terraform task.
    2. HLML Provider generates a TPM-signed lease token restricted to Terraform tools.
    3. The agent spawns a specialist subagent to execute the command.
    4. The subagent presents the lease to the tool gateway.
    5. The gateway verifies the hardware signature and the subagent's lineage back to the mission-root.
    6. Terraform executes successfully.
    7. Upon subagent termination, the HLML Provider broadcasts a hardware-bound revocation signal, neutralizing the lease.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -->|Request Lease| B[HLML Provider]
        B -->|Sign with TPM| C[Mission Lease Token]
        C -->|Inherit| D[Subagent A]
        D -->|Present Token| E[Tool Gateway]
        E -->|Verify Signature & Lineage| F[Privileged Tool]
        F -->|Result| D
        D -->|Task Complete| G[Revocation Logic]
        G -->|Invalidate| B
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(missionID, scopes) -> LeaseToken`: Generates a TPM-signed lease.
    * `hlml.VerifyLease(token, subagentLineage) -> bool`: Validates recursive hardware lineage.
    * `hlml.RevokeLease(leaseID)`: Irreversibly invalidates a hardware lease.
* **Data Storage/State:**
    * **Lease Registry:** Encrypted local database (SQLite) tracking active leases and their hardware fingerprints.
    * **Lineage Map:** Graph structure mapping parent-child agent relationships and their inherited leases.

## 5. Alternatives Considered
* **Short-lived JWTs:** Rejected because they are software-bound and can be replayed or exfiltrated. HLML requires hardware-attestation at every hop.
* **Sudo-style HITL:** Rejected for high-frequency autonomous swarms as it causes "Approval Fatigue." HLML provides autonomous security with hardware-enforced boundaries.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Implements "Recursive Verification" to prevent "Lease Hijacking" by rogue subagents.
* **Observability:** Integrated with the "Mission Lease Manager" in the UI for real-time tracking of active vs. expired leases.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
