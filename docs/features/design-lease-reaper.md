# Design Doc: Kernel-Resident Lease Reaper
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As swarms move toward multi-node ephemeral meshes, subagents are frequently spawned and terminated across disparate physical devices. Ephemeral hardware leases (HLML) are used to grant time-bound capabilities. However, when subagents terminate abruptly (e.g., due to tunnel failure or process crash), these leases often remain active, leading to "Capability Squatting" and potential security leaks.

The Kernel-Resident Lease Reaper is required to perform real-time, kernel-level monitoring of subagent liveness and automatically revoke orphaned hardware leases to maintain mesh stability and security.

## 2. Goals & Non-Goals
* **Goals:**
    * Monitor subagent process liveness at the kernel level (PID/Container tracking).
    * Automatically trigger HLML lease revocation upon subagent termination.
    * Neutralize "Capability Squatting" in high-density distributed meshes.
    * Provide audit logs for all automated lease reclamations.
* **Non-Goals:**
    * Managing the initial issuance of leases (handled by HLML Provider).
    * Monitoring non-agentic system processes.

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Security Auditor
* **Primary Goal:** Ensure that a specialist agent's access to a sensitive database tool is revoked immediately if its container crashes, even if the parent agent is busy.
* **The Happy Path (Tasks):**
    1. Specialist Agent is spawned on Node B with a task-bound HLML lease.
    2. Kernel-Resident Lease Reaper registers the PID/Container ID associated with the lease.
    3. Specialist Agent process crashes unexpectedly due to an OOM error.
    4. Lease Reaper detects the process exit signal from the kernel.
    5. Reaper immediately signals the hardware root (TPM) to invalidate the lease.
    6. Auditor sees the "Automated Reclamation" event in the security logs.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Subagent Process] --- B[Kernel Events/eBPF]
        B --> C[Lease Reaper]
        C --> D{Process Alive?}
        D -- No --> E[Revoke HLML Lease]
        E --> F[TPM Hardware Root]
        C --- G[Lease-to-PID Map]
    ```
* **APIs / Interfaces:**
    * `reaper.RegisterLease(leaseID, pid)`: Links a lease to a physical process.
    * `reaper.HealthCheck(leaseID) -> Status`: Manual query for lease liveness.
* **Data Storage/State:**
    * **Lease-to-PID Map:** Kernel-resident or high-priority userspace map for fast lookup during reaper cycles.

## 5. Alternatives Considered
* **Userspace Heartbeats:** Rejected because heartbeats can fail in high-latency meshes, leading to false revocations. Kernel-level exit monitoring is deterministic.
* **Pure Time-Based Expiration:** Rejected because it leaves a window of vulnerability between process crash and lease expiration.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The Reaper operates as a high-privilege service with direct access to the TPM revocation bus.
* **Observability:** Integrated with the "Subagent Reaper Dashboard" in the UI.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
