# Design Doc: Dynamic Permission Escalation Middleware

**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
Modern AI agent swarms often operate under a "static trust" model where an agent either has full access to a tool or none at all. This violates the principle of least privilege, especially when a subagent is delegated a task that requires temporary access to high-risk tools (e.g., `git push`, `kubernetes:delete-pod`). OpenClaw's recent move towards "Just-in-Time" (JIT) permissions highlights the need for a middleware that can manage dynamic trust transitions based on intent and human-in-the-loop (HITL) approval.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a "Request for Elevation" protocol where agents can ask for temporary access to a tool.
    *   Integrate with `HITL Middleware` to trigger user approval for elevation requests.
    *   Support cryptographically signed "Intent Tokens" to authorize elevation.
    *   Enforce time-bound and session-bound tool access.
*   **Non-Goals:**
    *   Managing permanent user identity (handled by external auth).
    *   Implementing the business logic of the tools themselves.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Agent Orchestrator.
*   **Primary Goal:** Allow a "Junior Coder" agent to request permission to merge a PR only after its code has been reviewed by a "Senior Auditor" agent.
*   **The Happy Path (Tasks):**
    1.  The Junior Coder calls `git_merge_pr`.
    2.  MCP Any intercepts the call and identifies it as "High Risk."
    3.  MCP Any returns a `PERMISSION_DENIED_ELEVATION_REQUIRED` error with an `elevation_request_id`.
    4.  The agent (or its orchestrator) submits an `elevation_request` with a signed approval from the Senior Auditor.
    5.  MCP Any validates the signature and grants temporary (e.g., 5-minute) access to `git_merge_pr` for that specific session.
    6.  The Junior Coder retries the call and succeeds.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: `EscalationMiddleware` checks the tool call against the current session's `CapabilityManifest`.
    - **Elevation Request**: A new JSON-RPC method `mcp_request_elevation` is introduced.
    - **Validation**: The middleware validates the provided "Intent Token" (JWT/Ed25519) or waits for a `HITL` approval.
    - **Manifest Update**: Upon approval, the session's in-memory `CapabilityManifest` is temporarily patched.
*   **APIs / Interfaces:**
    - `tools/call` (Interception point)
    - `mcp/request_elevation(tool_name string, reasoning string, token string)`
*   **Data Storage/State:** Temporary capabilities are stored in the session state (Shared KV Store) with an expiration TTL.

## 5. Alternatives Considered
*   **Pre-authorizing all tools**: Too risky for complex swarms.
*   **Requiring human approval for every call**: Too slow and creates user fatigue.
*   **Static Role-Based Access Control (RBAC)**: Too rigid for the dynamic nature of agentic delegation.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is a core Zero Trust feature. It ensures that "Trust is Earned, Not Given."
*   **Observability:** The UI must clearly show "Pending Elevation Requests" and a history of who authorized what elevation and why.

## 7. Evolutionary Changelog
*   **2026-03-03:** Initial Document Creation.
