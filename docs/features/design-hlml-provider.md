# Design Doc: Hardware-Locked Mission Leases (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
AI agent swarms often suffer from "Persistent Privilege Over-reach," where a subagent retains capabilities (like filesystem write or network access) long after its specific task is complete. This creates a massive attack surface if a subagent is compromised during a multi-day mission.

The Hardware-Locked Mission Leases (HLML) Provider addresses this by issuing TPM-signed, task-specific capability leases. These leases are cryptographically bound to a unique "Mission Root" and automatically expire upon task completion, ensuring that privilege is strictly lifecycle-bound.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, time-bound capability leases for high-risk tool calls.
    * Mandate automatic revocation of privileges upon mission-root task completion.
    * Provide cryptographic proof of mission-bound authority for inter-agent coordination.
    * Support "Mission Continuity" by allowing lease resumption across system reboots via the MRCP.
* **Non-Goals:**
    * Replacing general-purpose IAM systems for human users.
    * Managing model-layer safety (HLML is a transport and execution gate).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Swarm Orchestrator
* **Primary Goal:** Grant a "Temporary File System Write" capability to a specialist subagent that expires the moment the refactoring task is done.
* **The Happy Path (Tasks):**
    1. The parent agent requests a "Refactoring Lease" for the subagent.
    2. The HLML Provider generates a TPM-signed token bound to `TaskID: REFACTOR_001`.
    3. The subagent invokes the `fs:write` tool using this lease.
    4. MCP Any validates the lease against the hardware root and active mission status.
    5. The subagent completes the task and signals `TASK_DONE`.
    6. The HLML Provider immediately invalidates the lease; any further `fs:write` attempts by that subagent are rejected.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Parent[Parent Agent] -->|Request Lease| HLML[HLML Provider]
        HLML -->|Sign with TPM| Lease[TPM-Signed Lease]
        Lease -->|Authorize| Tool[High-Risk Tool]
        Tool -->|Callback| Status[Mission Status Monitor]
        Status -->|Task Done| HLML
        HLML -->|Revoke| Lease
    ```
* **APIs / Interfaces:**
    * `hlml.IssueLease(missionRoot, capabilities, duration) -> LeaseID`: Generates a hardware-attested lease.
    * `hlml.ValidateLease(leaseID, toolCall) -> boolean`: Verifies the lease and mission alignment.
    * `hlml.RevokeMission(missionID)`: Forcefully expires all leases associated with a mission.
* **Data Storage/State:**
    * Leases are stored in the Hardware-Locked Secure Enclave (TPM) and mirrored in the `Mission-Root Continuity Provider (MRCP)` for persistence.

## 5. Alternatives Considered
* **Time-Based JWTs:** Rejected because they don't account for the "Mission Lifecycle." A task might finish much faster than the JWT expiry, leaving a window of vulnerability.
* **Manual HITL Revocation:** Rejected as it doesn't scale for high-frequency autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All leases are hardware-locked. Even if the subagent's memory is dumped, the TPM-bound keys cannot be exfiltrated.
* **Observability:** Visualized in the "Mission Lease Manager" dashboard.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
