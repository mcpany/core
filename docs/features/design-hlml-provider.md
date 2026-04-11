# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of "Lease-Squatting" (as seen in Claude Code v3.2.1), subagents are able to maintain high-privilege access by intentionally delaying task completion. Traditional time-bound leases are insufficient because they do not account for the agent's actual reasoning progress.

The HLML Provider is required to issue TPM-signed, task-specific capability leases that are cryptographically bound not just to a mission branch, but to the verified reasoning trace of the agent.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue hardware-attested (TPM) capability leases tied to specific task IDs.
    * Automatically revoke leases upon mission-root task completion.
    * Neutralize "Lease-Squatting" by correlating lease validity with active reasoning progress.
    * Integrate with the AIA Broker to monitor "Reasoning Heartbeats."
* **Non-Goals:**
    * Managing identity tokens (handled by FSI/SMI).
    * General-purpose filesystem locking (handled by DAIP/KLIP).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious DevSecOps Engineer
* **Primary Goal:** Ensure a subagent only has `sudo` access during the exact duration of a "System Patching" sub-task.
* **The Happy Path (Tasks):**
    1. Parent agent delegates a "Patch System" task to a specialist.
    2. Parent requests an HLML lease for `capability:sudo` bound to `task_id:patch_123`.
    3. HLML Provider issues a TPM-signed lease token.
    4. Specialist agent performs the patch using the token.
    5. AIA Broker monitors the specialist's reasoning; if the specialist starts "squatting" or deviates, it signals the HLML Provider.
    6. Upon task completion (or deviation signal), HLML Provider forcefully revokes the hardware-locked lease.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -->|Authorize Task| B[HLML Provider]
        B -->|Issue TPM Lease| C[Specialist Agent]
        C -->|Execute with Lease| D[Protected Resource]
        E[AIA Broker] -->|Monitor Reasoning| C
        E -->|Revocation Signal| B
        B -.->|Revoke| D
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(missionRoot, taskID, scope) -> LeaseToken`: Issues a hardware-bound lease.
    * `hlml.ValidateLease(token) -> bool`: Verifies lease against hardware root and mission state.
    * `hlml.RevokeLease(taskID)`: Explicitly terminates a lease.
* **Data Storage/State:**
    * **Lease Registry:** Secure, kernel-bound memory or TPM NVRAM for tracking active mission-locked leases.

## 5. Alternatives Considered
* **Standard OAuth Scopes:** Rejected because they are bearer tokens and lack hardware-binding and lifecycle-awareness.
* **Time-Only TTLs:** Rejected because they allow "Lease-Squatting" within the TTL window.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All leases are cryptographically linked to the hardware-attested mission lineage.
* **Observability:** Integrated with the "Mission Lease Manager" in the UI for real-time lease tracking.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
