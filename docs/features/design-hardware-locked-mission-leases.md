# Design Doc: Hardware-Locked Mission Leases (HLML)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
Autonomous agents frequently require temporary access to high-privilege tools (e.g., shell, write access to root directories). Current "Static Scoping" models grant these permissions for the entire session, creating a risk of "Capability Squatting" where a compromised specialist agent retains access after its specific task is done.

Hardware-Locked Mission Leases (HLML) solve this by binding capabilities to specific, hardware-attested mission-root fragments. Privileges are issued as TPM-signed leases that expire automatically upon task completion or mission-root termination.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue time-bound, task-specific capability leases tied to a hardware identity.
    * Mandate recursive lease revocation, ensuring that termination of a parent lease cascades to all sub-missions.
    * Prevent persistent privilege escalation by neutralizing tokens post-mission.
    * Provide a hardware-signed audit trail of all lease lifecycle events.
* **Non-Goals:**
    * Managing system-level user permissions (e.g., Linux sudoers).
    * Providing long-term persistent access tokens.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Orchestrator
* **Primary Goal:** Grant a "DevOps Subagent" temporary access to run a deployment script without leaving the `deploy` capability active after the PR is merged.
* **The Happy Path (Tasks):**
    1. The orchestrator defines a "Mission Fragment" for deployment.
    2. HLML Provider issues a TPM-signed lease for the `deploy` tool, bound to the fragment ID.
    3. The subagent executes the deployment.
    4. Upon merging the PR, the parent agent signals mission completion.
    5. The HLML Provider instantly revokes the lease across the mesh.

## 4. Design & Architecture
* **System Flow:**
    `[Mission Root] -> [HLML Issuer] -> (TPM-Signed Lease) -> [Specialist Agent] -> [Privileged Tool]`
* **APIs / Interfaces:**
    * `GrantLease(Capability, MissionFragmentID) -> SignedLeaseToken`
    * `RevokeMission(MissionRootID) -> RevocationSignal`
* **Data Storage/State:**
    * Active leases are tracked in the Mission-Root Registry and mirrored to hardware monotonic counters to prevent replay.

## 5. Alternatives Considered
* **Short-lived JWTs:** Rejected because they lack hardware-bound non-repudiation and cannot be revoked instantly mesh-wide without a complex CRL.
* **Dynamic RBAC:** Rejected due to the overhead of updating global policies for high-frequency sub-missions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** "Zero-Privilege by Default"; capabilities are only "leased" during active reasoning phases.
* **Observability:** Integrated with the Mission Lease Manager UI for real-time visualization of active leases and automated expiration status.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
