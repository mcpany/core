# Design Doc: Remote Dispatch Attestation Provider
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
With the stabilization of Claude Code "Remote Control" and "Dispatch" v3.2.0, agents are increasingly operating as persistent background workers (non-interactive mode). This transition shifts the security frontier from point-in-time human approval to continuous, hardware-locked lease management. MCP Any needs to act as the authoritative issuer of these leases to ensure that headless dispatch workers remain bound to their specific mission-root intent.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed Mission-Bound Hardware Leases (MBHL) for background agent tasks.
    * Provide automated revocation of high-privilege tool access upon mission completion or lease expiration.
    * Maintain a verifiable audit trail of lease lifecycle for compliance (SOC2/GDPR).
* **Non-Goals:**
    * Managing the underlying lifecycle of the background worker process itself (handled by Claude Code/OpenClaw).
    * Providing a full Zero-Knowledge proof system (handled by the PPRP Validator).

## 3. Critical User Journey (CUJ)
* **User Persona:** Headless DevSecOps Swarm Orchestrator
* **Primary Goal:** Authorize a background "Dispatch" worker to perform limited security patches without exposing host-level sudo access permanently.
* **The Happy Path (Tasks):**
    1. The parent agent requests a "Remote Dispatch" session for a sub-task.
    2. MCP Any validates the request against the mission-root manifest.
    3. The Remote Dispatch Attestation Provider issues a TPM-signed MBHL with a 30-minute expiry and specific tool scopes (e.g., `fs:write:/src/security`).
    4. The background worker utilizes the lease to call authorized tools via MCP Any.
    5. Upon task completion, the worker signals termination, and MCP Any forcefully revokes the lease.

## 4. Design & Architecture
* **System Flow:**
    [Subagent Request] -> [Dispatch Attestation Provider] -> [TPM Signature Service] -> [Issued MBHL]
    [Tool Call + MBHL] -> [MCP Any Policy Engine] -> [Execution]
* **APIs / Interfaces:**
    * `POST /v1/dispatch/lease/request`: Request a new mission-bound lease.
    * `POST /v1/dispatch/lease/revoke`: Manual revocation of a lease.
    * `GET /v1/dispatch/lease/status`: Check validity and scope of an active lease.
* **Data Storage/State:**
    Lease metadata is stored in the hardware-bound SQLite "Blackboard," with signatures verified against the local hardware security module (TPM/SEP).

## 5. Alternatives Considered
* **Time-Based JWTs:** Rejected because they can be replayed if exfiltrated; MBHLs are hardware-bound and non-exportable.
* **Manual HITL for Every Call:** Rejected due to "Approval Fatigue" in high-frequency background swarms.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** MBHLs are strictly scoped; even if a subagent is compromised, it cannot use the lease to access tools outside the predefined fragment.
* **Observability:** Every lease issuance and revocation event is logged to the "Command Traceability Provider" (CTP).

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
