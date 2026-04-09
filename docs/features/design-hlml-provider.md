# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As agent swarms evolve from simple sessions into multi-node P2P meshes, the risk of "Persistent Privilege Escalation" increases. Specialist agents often inherit high-risk capabilities (e.g., shell access, production DB writes) that remain active long after their specific sub-task is finished. MCP Any needs a mechanism to enforce the principle of least privilege in the dimension of *time*.

The HLML Provider implements task-bound, hardware-attested capability leases. Inspired by Claude Code's MBHL standard, this system ensures that subagent agency is cryptographically restricted to the lifecycle of a specific user-authorized task.

## 2. Goals & Non-Goals
* **Goals:**
    *   Bind tool capabilities to hardware-attested (TPM/SEP) mission-root fragments.
    *   Implement "Just-in-Time" privilege escalation with automated, non-repudiable expiration.
    *   Provide sub-millisecond revocation of leases upon sub-mission termination.
* **Non-Goals:**
    *   Replacing the primary Policy Firewall (HLML acts as a temporal overlay).
    *   Managing user-level authentication (assumes a verified Mission Root).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Swarm Architect
* **Primary Goal:** Delegate a sensitive deployment task to a subagent swarm without leaving persistent "Shell" access on the production nodes.
* **The Happy Path (Tasks):**
    1.  The Mission Root initiates a "Deploy WebApp" task.
    2.  MCP Any issues a TPM-signed HLML lease for `exec:kubectl` bound to Task ID `7782`.
    3.  The Specialist Agent executes the deployment.
    4.  Upon the `exit(0)` signal of the deployment task, the HLML Provider broadcasts a revocation signal to the mesh.
    5.  Subsequent attempts by the Specialist to use `exec:kubectl` are blocked by hardware-level lease expiry.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        MissionRoot->>HLML: Request Lease (TaskID, Capability)
        HLML->>TPM: Sign Lease Fragment
        TPM-->>HLML: signed_lease_v1
        HLML->>Specialist: Issue Lease
        Specialist->>Gateway: Call Tool (signed_lease_v1)
        Gateway->>HLML: Verify Lease Persistence
        HLML-->>Gateway: OK
        Specialist->>Gateway: Signal Task Complete
        Gateway->>HLML: Terminate Lease
        HLML->>ARL: Broadcast Expiry
    ```
* **APIs / Interfaces:**
    *   `POST /v1/lease/issue`: Requires Mission-Root Token and Task Manifest.
    *   `DELETE /v1/lease/revoke`: Manual or automated revocation.
* **Data Storage/State:**
    *   Leases are stored in kernel-bound protected memory (HLES) and indexed by Task ID in the Blackboard.

## 5. Alternatives Considered
* **Time-Based TTLs**: Rejected because task durations are non-deterministic; leases might expire mid-task or persist too long.
* **Pure Software-Bound Tokens**: Rejected due to the risk of token exfiltration in compromised subagent process environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All lease transitions are hardware-attested; "Lease Hijacking" is mitigated by binding the lease to the specialist's hardware-enclave ID.
* **Observability:** Leases are tracked in the **Mission Lease Manager** UI, showing real-time countdowns to task completion and automated expiration logs.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
