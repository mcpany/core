# Design Doc: JIT Permission Escalation Middleware

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
The current permission model in MCP Any relies on static scopes assigned to sessions or tokens at initialization. As agents become more autonomous and their tasks more complex, this "all-or-nothing" approach either leads to over-privileged agents or frequent task failures. JIT (Just-in-Time) Permission Escalation allows an agent to request additional capabilities dynamically when it encounters a task that requires them, following the principle of least privilege.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enable agents to request temporary capability upgrades (e.g., `fs:write`) during an active session.
    *   Integrate with the Policy Firewall to automate low-risk escalations.
    *   Provide a standardized "Interactive Approval" flow for high-risk escalations via HITL Middleware.
    *   Ensure escalations are time-bound or task-bound (e.g., "Allow for the next 5 minutes").
*   **Non-Goals:**
    *   Permanently modifying the user's base configuration.
    *   Implementing the UI for user approval (this is handled by the HITL component).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Local LLM Swarm Orchestrator.
*   **Primary Goal:** Allow a "Reader" agent to temporarily become a "Writer" to fix a bug it discovered, without manual configuration changes.
*   **The Happy Path (Tasks):**
    1.  Agent attempts to call `write_file` but lacks the `fs:write` capability.
    2.  MCP Any returns a `403 Forbidden` with a `request_escalation_hint`.
    3.  Agent calls `mcpany_request_escalation(capability="fs:write", reason="Updating outdated dependency in package.json")`.
    4.  MCP Any triggers the Policy Firewall; the request is flagged for "User Approval" (HITL).
    5.  User approves the request via the UI.
    6.  MCP Any grants the `fs:write` capability to the session for 10 minutes.
    7.  Agent successfully calls `write_file`.

## 4. Design & Architecture
*   **System Flow:**
    `Agent -> Gateway -> Permission Middleware (Check) -> [Fail] -> Agent -> Gateway -> Escalation Middleware -> Policy Engine -> [HITL] -> Approval -> Session State (Update)`
*   **APIs / Interfaces:**
    - New Tool: `mcpany_request_escalation(capability: string, reason: string, duration_minutes: int)`
    - Internal API: `PUT /sessions/{id}/capabilities`
*   **Data Storage/State:**
    - Temporary capabilities are stored in the session's memory or `Shared KV Store` with an expiration timestamp.

## 5. Alternatives Considered
*   **Pre-emptive Scoping**: Trying to guess all needed permissions. *Rejected* because it's impossible for complex, non-deterministic agent workflows.
*   **Global Admin Mode**: Running the agent with full system access. *Rejected* as it violates all Zero Trust principles.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Escalation requests must be cryptographically signed by the session token. Reasons must be logged for auditability.
*   **Observability:** The UI should show "Active Escalations" and their remaining duration.

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
