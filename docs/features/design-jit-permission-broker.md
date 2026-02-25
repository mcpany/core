# Design Doc: JIT Permission Broker (Intent-Escalation)

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agents become more autonomous, they frequently encounter security boundaries that they cannot cross without explicit authorization. In a "Zero Trust" environment like MCP Any, this leads to "Permission Deadlock," where an agent stops execution because it lacks a necessary capability (e.g., writing to a specific directory or accessing a sensitive tool). The JIT Permission Broker solves this by allowing agents to request temporary, scoped permission elevation by providing an "Intent Token" that describes the task they are attempting.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Allow agents to request temporary permission elevation for specific tool calls.
    *   Implement "Intent Tokens" that cryptographically link a permission request to a specific high-level task.
    *   Integrate with the HITL Middleware for human approval of elevation requests.
    *   Provide automatic expiration of JIT permissions once a task is complete.
*   **Non-Goals:**
    *   Permanent permission modification (all JIT elevations are ephemeral).
    *   Automatic approval of all requests (must follow policy engine rules).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Autonomous DevOps Agent.
*   **Primary Goal:** Fix a production bug by writing a temporary log file to a restricted directory without pre-configured access.
*   **The Happy Path (Tasks):**
    1.  The Agent attempts to write to `/var/log/app/` and receives a `PERM_DENIED` error from MCP Any.
    2.  The Agent calls the `request_jit_elevation` tool, providing an Intent Token: `{"intent": "debug_prod_issue_123", "resource": "/var/log/app/", "duration": "10m"}`.
    3.  MCP Any suspends the session and triggers a HITL notification to the human operator.
    4.  The operator approves the request via the UI.
    5.  MCP Any grants a temporary capability token to the Agent's session.
    6.  The Agent successfully completes the write operation.
    7.  After 10 minutes, the permission is automatically revoked.

## 4. Design & Architecture
*   **System Flow:**
    - **Request Capture**: The `PermissionBrokerMiddleware` intercepts failed tool calls or explicit elevation requests.
    - **Intent Verification**: The Policy Engine evaluates the request against "Escalation Rules."
    - **Elevation Execution**: If approved, a new `CapabilityToken` is minted and injected into the current `AgentSession`.
*   **APIs / Interfaces:**
    - `POST /v1/permissions/request`: Endpoint for agents to submit JIT requests.
    - `POST /v1/permissions/approve`: Endpoint for HITL approvals.
*   **Data Storage/State:** Ephemeral permissions are stored in the `Shared KV Store` with a TTL (Time-To-Live).

## 5. Alternatives Considered
*   **Broad Initial Permissions**: Granting agents high-level access upfront. *Rejected* due to security risks (Violates Principle of Least Privilege).
*   **Static Configuration Updates**: Requiring a human to update `config.yaml` and restart. *Rejected* as it breaks agent autonomy and real-time responsiveness.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Prevent "Token Smuggling" where one agent passes its JIT token to another unauthorized agent. Tokens must be bound to a specific `SessionID` and `AgentID`.
*   **Observability:** All JIT requests and approvals are logged in the Audit Trail with a reference to the Intent Token.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
