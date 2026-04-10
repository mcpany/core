# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the rise of Agent Teams and high-privilege autonomous operations, persistent tool access has become a critical vulnerability. If an agent is compromised mid-mission, attackers can exploit dormant capabilities. MCP Any needs a mechanism to bind privilege to the mission lifecycle using hardware-backed guarantees.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, task-specific capability leases.
    * Automatically revoke capabilities upon mission-root completion.
    * Provide cryptographic proof of lease expiration to the security gateway.
* **Non-Goals:**
    * Managing the underlying hardware (TPM/SEP) driver logic.
    * Replacing long-term identity certificates.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Enterprise Admin
* **Primary Goal:** Ensure that a subagent's `run_shell_command` capability is only valid for the duration of a specific `Fix Bug` mission.
* **The Happy Path (Tasks):**
    1. Parent agent initiates a mission and requests an HLML for a specialist subagent.
    2. HLML Provider issues a TPM-signed lease bound to the Mission ID and specific tool scopes.
    3. Specialist subagent invokes the tool; the security gateway verifies the lease's hardware signature and mission status.
    4. Upon mission completion, the parent agent signals the HLML Provider.
    5. HLML Provider invalidates the lease; subsequent tool calls by the specialist are rejected.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        A[Parent Agent] -->|Request Lease| B(HLML Provider)
        B -->|Sign with TPM| C{Mission Lease}
        C -->|Used By| D[Subagent]
        D -->|Tool Call + Lease| E(Security Gateway)
        E -->|Verify Signature & Status| F[Tool Execution]
        A -->|End Mission| B
        B -->|Invalidate| C
    ```
* **APIs / Interfaces:**
    * `IssueLease(mission_id, tools[], duration) -> SignedLease`
    * `InvalidateLease(lease_id) -> Success`
* **Data Storage/State:**
    * Lease registry stored in a hardware-isolated secure vault.

## 5. Alternatives Considered
* **Time-bound JWTs:** Rejected because they can be replayed until expiration even after the task is done.
* **Manual HITL per call:** Rejected due to prohibitive latency in autonomous swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Mandatory TPM-signing ensures leases cannot be forged or modified by software.
* **Observability:** Logs all lease issuance, usage, and revocation events with hardware-attested timestamps.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.

### Update: 2026-07-25 - Atomic Scratchpad Purge Integration
**Context:** Today's market sync revealed CVE-2026-44012 (Scratchpad Leakage) in Claude Code, where subagents can exfiltrate data from shared team workspaces.
**Architecture Adjustment:**
* Implementing **Atomic Lease-Bound Purge (ALBP)** logic within the revocation flow.
* When a Mission Lease is invalidated, the HLML Provider now triggers a kernel-level wipe of all project-local scratchpad fragments (.scratchpad/*) associated with that mission ID.
**Security Impact:** Prevents "Context-Stitching" attacks by ensuring that no intent residue remains in shared workspaces after a mission concludes.
