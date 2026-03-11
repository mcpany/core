# Design Doc: HITL Middleware (Human-in-the-Loop)

**Status:** Draft
**Created:** 2026-03-11

## 1. Context and Scope
Autonomous agents occasionally attempt high-risk operations (e.g., deleting a production database, making unauthorized financial transactions, or pushing code to main without review). While the Policy Firewall can block known bad actions, many actions require human judgment. The HITL (Human-in-the-Loop) Middleware provides a suspension protocol that pauses tool execution and requests explicit user approval.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Pause tool execution when a "high-risk" rule is triggered.
    *   Notify the user via the UI or external channels (e.g., Slack/Email).
    *   Support "Approve," "Deny," and "Modify" (input override) actions by the human.
    *   Ensure state persistence while an agent is waiting for approval.
*   **Non-Goals:**
    *   Replacing automated security policies (HITL is for edge cases requiring judgment).
    *   Implementing the notification delivery system (MCP Any provides the hooks).

## 3. Critical User Journey (CUJ)
*   **User Persona:** DevOps Engineer using an Autonomous SRE Agent.
*   **Primary Goal:** Review and approve a tool call that would resize a production Kubernetes cluster.
*   **The Happy Path (Tasks):**
    1.  The SRE Agent decides to resize the cluster and calls `k8s:resize_cluster`.
    2.  The HITL Middleware identifies this as a "Critical Action."
    3.  Tool execution is suspended; the agent enters a `WAITING_FOR_APPROVAL` state.
    4.  The Developer receives a notification in the MCP Any Dashboard.
    5.  The Developer reviews the parameters and clicks "Approve."
    6.  Tool execution resumes; the cluster is resized.

## 4. Design & Architecture
*   **System Flow:**
    - **Trigger**: Policy Firewall identifies an action as `ACTION_SUSPEND`.
    - **Suspension**: The middleware returns a specific "Pending" status to the agent or holds the request if the transport allows.
    - **Persistence**: The call context and arguments are serialized to the `Shared KV Store`.
    - **Notification**: A message is pushed to the UI via WebSockets.
*   **APIs / Interfaces:**
    - `POST /api/v1/approvals/{id}/respond`: Endpoint for the user to approve/deny.
    - `GET /api/v1/approvals`: List pending approvals.
*   **Data Storage/State:** Suspended calls are stored in SQLite with a TTL.

## 5. Alternatives Considered
*   **Agent-Side Approval**: Asking the agent to ask the user. *Rejected* because the agent itself might be compromised or hallucinating. Security must be enforced at the gateway.
*   **Synchronous Blocking**: Holding the TCP connection open indefinitely. *Rejected* due to timeout issues and resource exhaustion.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Approval tokens must be cryptographically bound to the user session. Denying an action should ideally provide a reason back to the agent to prevent infinite loops.
*   **Observability:** Audit logs must clearly distinguish between agent-initiated actions and human-approved actions.

## 7. Evolutionary Changelog
*   **2026-03-11:** Initial Document Creation.
