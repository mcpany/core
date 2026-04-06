# Design Doc: Hardware-Locked Mission Lease (HLML) Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As multi-agent swarms perform high-privilege operations (e.g., shell command execution, sensitive file access), the risk of "Privilege Residue" or "Lease-Squatting" has become a primary attack vector. If an agent retains a capability after its specific sub-task is completed, a compromised specialist or a "Ghost" process can perform unauthorized actions.

The HLML Provider is designed to move MCP Any from session-based permissions to mission-bound, hardware-attested leases. Every high-risk capability is granted as a TPM-signed lease that is cryptographically anchored to a specific intent and must be explicitly "Finalized" via a completion handshake.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Bind high-privilege tool access to specific, hardware-attested sub-mission IDs.
    *   Enforce atomic lease revocation upon task completion via a completion handshake.
    *   Neutralize "Lease-Squatting" by requiring re-attestation for any goal shift.
    *   Provide a verifiable audit trail of lease lifecycle (Grant -> Use -> Finalize).
*   **Non-Goals:**
    *   Managing basic read-only access for low-risk tools (handled by standard scopes).
    *   Replacing the primary user-level attestation (HLML operates at the subagent/task level).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator
*   **Primary Goal:** Grant a specialist agent 60 seconds of `run_shell_command` access for a specific bug-fix task, ensuring the access is revoked the instant the bug is fixed.
*   **The Happy Path (Tasks):**
    1.  The Mission Root identifies a task and requests an HLML lease for the `bug-fix-shell` subagent.
    2.  MCP Any issues a TPM-signed lease, cryptographically bound to the task-intent and a 60s timeout.
    3.  The subagent performs authorized shell commands; every call is validated against the HLML lease.
    4.  The subagent completes the task and issues a "Task Finalized" signal.
    5.  The HACH (Completion Handshake) Provider verifies the final state and atomically revokes the lease.
    6.  The subagent attempts a subsequent shell command; the HLML Provider rejects it as the lease is closed.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph TD
        A[Mission Root] -->|Request Lease| B[HLML Provider]
        B -->|Sign with TPM| C[Mission-Bound Lease]
        C --> D[Subagent Execution]
        D -->|Tool Call + Lease| E[Validation Gate]
        E -->|Verify Intent| F[MCP Tool]
        D -->|Completion Signal| G[HACH Provider]
        G -->|Atomic Revocation| B
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/lease/grant`: Request a new mission-bound hardware lease.
    *   `POST /v1/lease/finalize`: Submit a completion token to close a lease.
    *   `GET /v1/lease/status/{lease_id}`: Check the remaining lifetime and intent-binding of a lease.
*   **Data Storage/State:**
    *   Lease state is managed in the **Memory-Mapped Lease Buffer**, with hardware-backed monotonic counters to prevent replay.

## 5. Alternatives Considered
*   **Time-based TTL only:** Rejected because a fast subagent could still be hijacked within the remaining TTL window. HLML requires *both* intent-binding and atomic finalization.
*   **OS-level process killing:** Rejected as too blunt; we want to revoke the *capability*, not necessarily terminate the specialist agent if it has other low-risk monitoring tasks.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** "Lease-Bound Intent Anchoring" ensures that even if a TPM key is leaked, it is only valid for the specific, immutable goal defined in the lease.
*   **Observability:** Visualized in the "Mission Lease Manager" in the UI, showing active leases, their intents, and countdown to revocation.

## 7. Evolutionary Changelog
*   **2026-07-25:** Initial Document Creation.
