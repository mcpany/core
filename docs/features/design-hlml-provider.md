# Design Doc: Hardware-Locked Mission Leases (HLML) Provider
**Status:** Draft
**Created:** 2026-04-17

## 1. Context and Scope
Agentic "Privilege Escalation" remains a critical threat vector. Once an agent is granted access to a high-privilege tool (like shell execution), it often retains that access for the duration of the session, even after the specific task requiring it is finished. Inspired by Claude Code's "Mission-Bound Hardware Leases," the HLML Provider ensures that tool access is cryptographically bound to a specific mission branch and expires automatically.

## 2. Goals & Non-Goals
* **Goals:**
    * Grant time-bound and task-bound tool leases.
    * Bind leases to hardware-attested (TPM) session tokens.
    * Automatically revoke capabilities upon sub-mission termination or intent drift.
    * Neutralize "Identity Squatting" where specialist agents retain parent-level tokens.
* **Non-Goals:**
    * Managing persistent system-level users or permissions.
    * Providing long-term persistent access.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Allow a "Refactoring Agent" to run shell commands ONLY during the refactoring task, ensuring it cannot use those commands later if it gets compromised or enters a hallucination loop.
* **The Happy Path (Tasks):**
    1. The Mission Root spawns a "Refactoring Subagent" and requests an HLML lease for `run_shell_command`.
    2. MCP Any Hub issues a TPM-signed lease token, valid for 30 minutes and scoped to the `refactor-v1` mission ID.
    3. The Subagent executes shell commands; the HLML Middleware validates the lease and mission ID on every call.
    4. The refactoring task completes; the Subagent signals termination.
    5. The HLML Provider immediately invalidates the lease, even if the 30-minute window hasn't expired.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Root[Mission Root] -->|Request Lease| HLML[HLML Provider]
        HLML -->|TPM-Signed Token| Sub[Subagent]
        Sub -->|Tool Call + Lease| Gateway[MCP Any Gateway]
        Gateway -->|Validate Lease| HLML
        HLML -->|Grant/Deny| Tool[Protected Tool]
    ```
* **APIs / Interfaces:**
    * `POST /lease/grant`: Issues a task-scoped, hardware-bound token.
    * `POST /lease/revoke`: Explicitly invalidates a lease.
* **Data Storage/State:**
    * Active Lease Registry (ALR) mapping mission IDs to lease metadata.
    * Hardware-bound revocation lists.

## 5. Alternatives Considered
* **Time-based TTL (Only)**: Rejected because many tasks complete faster than the TTL, leaving a window of vulnerability.
* **Process-based Isolation**: Useful but insufficient, as an agent can still abuse a tool within its own process if the gateway allows it.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** All high-privilege tool calls mandate a valid, unexpired HLML lease.
* **Observability:** Real-time dashboard showing active leases and their association with specific mission branches.

## 7. Evolutionary Changelog
* **2026-04-17:** Initial Document Creation.
