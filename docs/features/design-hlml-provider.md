# Design Doc: Hardware-Locked Mission Leases (HLML)
**Status:** Draft
**Created:** 2026-07-25

## 1. Context and Scope
As AI agents move from single-step tool calls to complex, multi-day missions, persistent privilege becomes a critical security liability. Current models grant broad "sudo" or "network" access for the duration of a session, which can be hijacked if a specialist subagent is compromised.

Hardware-Locked Mission Leases (HLML) solve this by binding capability grants to specific mission-root task IDs and cryptographically enforcing their expiration via a Trusted Platform Module (TPM). This ensures that even if an agent session remains active, its access to sensitive tools like `run_shell_command` is automatically revoked once its assigned task is marked complete in the Mission Manifest.

## 2. Goals & Non-Goals
* **Goals:**
    * Issue TPM-signed, task-specific capability leases for high-risk operations.
    * Automatically revoke leases upon mission-root task completion or timeout.
    * Provide non-repudiable audit trails of lease issuance and consumption.
    * Integrate with Claude Code v3.2.0 MBHL standard.
* **Non-Goals:**
    * Replacing OS-level user permissions (HLML sits on top as an agentic policy layer).
    * Managing human user login sessions (focused on Non-Human Identity agency).

## 3. Critical User Journey (CUJ)
* **User Persona:** Local LLM Swarm Orchestrator
* **Primary Goal:** Grant a "Refactor Specialist" subagent temporary access to write to the `/src` directory only for the duration of its refactoring task.
* **The Happy Path (Tasks):**
    1. Parent agent defines a task "Refactor auth logic" in the Mission Manifest.
    2. Parent requests an HLML lease for `fs:write:/src` for the specialist, bound to the Task ID.
    3. MCP Any verifies the request against user policy and generates a TPM-signed lease token.
    4. Specialist agent consumes the token to perform writes via the FileSystem tool.
    5. Specialist signals task completion or hits a 30-minute timeout.
    6. MCP Any's HLML Provider invalidates the token, and subsequent tool calls are denied.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        ParentAgent->>HLMLProvider: Request Lease(TaskID, Capability, Timeout)
        HLMLProvider->>TPM: Sign Lease Fragment
        TPM-->>HLMLProvider: SignedToken
        HLMLProvider-->>ParentAgent: LeaseHandle
        ParentAgent->>SubAgent: Delegate(LeaseHandle)
        SubAgent->>ToolGateway: CallTool(LeaseHandle, Args)
        ToolGateway->>HLMLProvider: Validate(LeaseHandle)
        HLMLProvider-->>ToolGateway: OK
        ToolGateway->>Tool: Execute
        SubAgent->>MissionManifest: MarkComplete(TaskID)
        HLMLProvider->>HLMLProvider: Revoke(LeaseHandle)
    ```
* **APIs / Interfaces:**
    * `POST /v1/leases/issue`: Requests a new mission-bound lease.
    * `GET /v1/leases/verify`: Internal endpoint for the Tool Gateway to check lease status.
    * `DELETE /v1/leases/revoke`: Explicitly terminates a lease.
* **Data Storage/State:**
    * Leases are stored in the hardware-attested "Sovereignty Vault" (Secure SQLite).
    * Active handles are kept in kernel-bound memory for sub-millisecond validation.

## 5. Alternatives Considered
* **Time-only Leases**: Rejected because they don't account for early task completion, leaving a window of vulnerability.
* **Session-wide Tokens**: Rejected as they provide too broad a blast radius if a subagent is compromised.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tokens are hardware-bound; cloning the token to another process or device will result in an integrity failure.
* **Observability:** Every lease request, issuance, and revocation is logged in the Audit Trail with the associated Mission Root ID.

## 7. Evolutionary Changelog
* **2026-07-25:** Initial Document Creation.
