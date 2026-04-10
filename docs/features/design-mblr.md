# Design Doc: Mission-Bound Lease Reaper (MBLR)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the introduction of Mission-Bound Hardware Leases (MBHL) in Claude Code and Sovereign Node Tunneling (SNT) in OpenClaw, AI agents can now hold cryptographically-bound privileges for specific tasks. However, "Lease Squatting" has emerged as a critical vulnerability where subagents or compromised processes fail to release these leases after task completion, leading to resource starvation and potential persistent privilege escalation.

The Mission-Bound Lease Reaper (MBLR) provides a hardware-locked enforcement layer that monitors the mission-root lifecycle and forcefully reclaims orphaned leases. It ensures that any capability issued via a TPM-signed lease is strictly time-bound and mission-aligned, providing a definitive end to autonomous agency.

## 2. Goals & Non-Goals
* **Goals:**
    * Forcefully revoke MBHLs upon detection of mission completion or subagent abandonment.
    * Implement a "Termination Proof" protocol requiring hardware-attested signals for graceful lease release.
    * Integrate with TPM/Secure Enclaves to perform hardware-level privilege invalidation.
    * Neutralize "Lease Squatting" in high-density horizontal swarms.
* **Non-Goals:**
    * Managing software-only sessions not backed by hardware leases.
    * Modifying the internal logic of agent reasoning engines.

## 3. Critical User Journey (CUJ)
* **User Persona:** High-Security Swarm Administrator
* **Primary Goal:** Prevent a rogue subagent from retaining `run_shell_command` privileges after its assigned debugging task is finished.
* **The Happy Path (Tasks):**
    1. A subagent is issued an MBHL for `run_shell_command` bound to Task-402.
    2. The subagent completes the task but attempts to "squat" on the lease to perform unauthorized scanning.
    3. The MBLR monitors the Shared Task List and detects that Task-402 is marked as completed by the Mission Root.
    4. The MBLR challenges the subagent for a hardware-attested Termination Proof.
    5. The subagent fails to provide a valid proof within the 500ms grace period.
    6. The MBLR triggers a TPM command to invalidate the lease's cryptographic material.
    7. Subsequent attempts by the subagent to use the lease are rejected by the hardware root.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -->|Mark Task Done| B[Shared Task List]
        C[MBLR Daemon] -->|Poll| B
        C -->|Challenge| D[Subagent]
        D -->|Fail Proof| C
        C -->|Revoke| E[TPM/Hardware Root]
        E -->|Invalidate| F[MBHL Storage]
        F -->|Reject Call| G[Tool Execution]
    ```
* **APIs / Interfaces:**
    * `mblr.RegisterLease(leaseID, taskID, tpmContext)`: Binds a hardware lease to a mission task.
    * `mblr.RequestTerminationProof(leaseID) -> Challenge`: Issues a challenge to a subagent.
    * `mblr.ForceRevoke(leaseID)`: Direct hardware-level revocation.
* **Data Storage/State:**
    * **Lease Sentinel Map**: Kernel-bound mapping of `Lease-ID -> {Hardware-Context, Task-State}`.

## 5. Alternatives Considered
* **Software-only Reaping:** Rejected because compromised agents can bypass software gates or ignore SIGTERM. Hardware-level invalidation is non-bypassable.
* **Short-lived TTLs:** Used as a fallback, but insufficient for tasks with variable durations. Active reaping based on task completion is more efficient.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Revocation signals are only accepted from the hardware root or the Mission Root's authenticated proxy.
* **Observability:** Integrated with the "Mission Lease Manager" UI for real-time visualization of lease status and reaping events.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation. Evolving from the Active Subagent Reaper to support hardware-locked MBHL reclamation.
