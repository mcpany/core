# Design Doc: HITL Middleware (Human-in-the-Loop)

**Status:** Draft
**Created:** 2026-02-07

## 1. Context and Scope
AI agents can perform high-impact actions (e.g., deleting a database, sending a payment) that require human verification. Currently, agents often "hallucinate" success or fail silently when blocked. MCP Any needs a standardized protocol to suspend tool execution and wait for human approval.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a suspension mechanism in the tool execution lifecycle.
    *   Standardize a `CallToolResult` sub-type for "Pending Approval."
    *   Provide an API for users to approve or deny pending actions via the UI.
    *   Support auto-timeout for stale approvals.
*   **Non-Goals:**
    *   Build a complex workflow engine (keep it focused on single-step approvals).
    *   Handle long-running background tasks (stays within the session context).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Financial Controller
*   **Primary Goal:** Approve a $10,000 wire transfer initiated by an autonomous procurement agent.
*   **The Happy Path (Tasks):**
    1.  Agent calls `bank:transfer_funds` with `$10,000`.
    2.  HITL Middleware identifies the tool as "high-impact" based on configuration.
    3.  Middleware suspends the request and returns a `REQUIRES_APPROVAL` status to the client.
    4.  A notification appears in the MCP Any UI.
    5.  The Controller reviews the details and clicks "Approve."
    6.  The Middleware resumes the execution, calling the upstream API.
    7.  The final result is returned to the agent.

## 4. Design & Architecture
*   **System Flow:**
    `Agent` -> `MCP Any` -> `HITL Middleware (Suspend)` -> `Store in SQLite` -> `User Approval (via UI)` -> `Resume Middleware` -> `Upstream`
*   **APIs / Interfaces:**
    *   `POST /v1/approvals/{id}/resolve`: Endpoint for the UI to signal approval/denial.
    *   `GET /v1/approvals`: List pending approvals.
*   **Data Storage/State:** Pending approvals are stored in the embedded SQLite database to survive restarts.

## 5. Alternatives Considered
*   **Agent-Side HITL:** Rejected because it relies on the agent being "honest" and "competent" enough to ask for permission.
*   **Polling:** Rejected in favor of a push-based notification (WebSocket/SSE) for better UX.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Approvals must be signed by an authorized user session.
*   **Observability:** Approval history is part of the "Agent Black Box" and Audit Log.

## 7. Evolutionary Changelog
*   **2026-02-07:** Initial Document Creation.
