# Design Doc: Just-In-Time (JIT) Permission Broker

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agents become more autonomous, they frequently encounter tasks that require higher privileges than their default security scope allows. Static permission models lead to "Agent Deadlock," where an agent stalls because it lacks the necessary tool access. The JIT Permission Broker provides a mechanism for agents to request temporary, task-scoped permission elevations that are verified against context and optionally approved by a human (HITL).

## 2. Goals & Non-Goals
*   **Goals:**
    *   Allow agents to request temporary access to specific tools or scopes.
    *   Verify elevation requests against the current "Intent" and "Task Context."
    *   Integrate with HITL Middleware for high-risk elevations.
    *   Automatically revoke elevated permissions after a task is completed or a timeout is reached.
*   **Non-Goals:**
    *   Permanent permission changes.
    *   Replacing the Policy Firewall (JIT works *with* the firewall).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Autonomous DevOps Agent.
*   **Primary Goal:** Temporarily elevate permissions to restart a production service after diagnosing a failure, without having permanent 'admin' access.
*   **The Happy Path (Tasks):**
    1.  Agent identifies that it needs to call `restart_service`, but its current token only allows `read_logs`.
    2.  Agent calls the `request_elevation` tool provided by MCP Any, specifying the required tool and the justification.
    3.  The JIT Broker analyzes the justification and the recent tool call history (via Trace IDs).
    4.  If the risk is medium/high, the Broker triggers a HITL notification to the human operator.
    5.  The Human approves the request via the UI.
    6.  The Broker issues a temporary capability token valid for 5 minutes.
    7.  The Agent executes `restart_service` successfully.
    8.  The token expires and permissions revert to default.

## 4. Design & Architecture
*   **System Flow:**
    - **Request**: Agents call a built-in `mcp_any_request_elevation` tool.
    - **Verification**: The Broker checks the `Policy Firewall` for "Allowable JIT Scopes."
    - **HITL Integration**: If required, the request is pushed to the `HITL Middleware` queue.
    - **Token Issuance**: A short-lived, JWT-based capability token is injected into the agent's session context.
*   **APIs / Interfaces:**
    - `POST /api/jit/request`: Internal endpoint for elevation requests.
    - `GET /api/jit/status/:id`: Check status of an elevation request.
*   **Data Storage/State:** Active JIT elevations are tracked in the `Shared KV Store` with an expiration timestamp.

## 5. Alternatives Considered
*   **Static Over-Provisioning**: Giving agents more permissions than they need. *Rejected* due to security risks (Zero Trust violation).
*   **Manual Config Updates**: Requiring a human to edit `config.yaml` every time. *Rejected* because it breaks agent autonomy and speed.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Elevation is the highest-risk action. Must be logged in the Audit Trail with full context. Anti-spoofing checks must ensure the request comes from the authorized agent session.
*   **Observability:** The UI must clearly highlight active elevations and their remaining duration.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
