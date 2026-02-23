# Design Doc: HITL (Human-In-The-Loop) Middleware
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
Autonomous agents operating in "Yolo" mode (high-trust execution) can cause irreversible damage if they execute sensitive tools (e.g., `delete_database`, `send_payment`) without confirmation. The HITL Middleware provides a standardized suspension protocol that pauses tool execution until a human explicitly approves the action via the UI or CLI.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept sensitive tool calls and suspend execution.
    * Provide a persistent "Pending Approval" state in the MCP Any database.
    * Notify the UI/user of a pending action requiring approval.
    * Allow the user to "Approve", "Deny", or "Modify" the tool arguments before execution.
* **Non-Goals:**
    * Automating the approval process (that's what the agent is for).
    * Supporting offline approvals (user must be active or eventually respond).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer / Ops Engineer
* **Primary Goal:** Review and approve a database migration tool call before it runs on production.
* **The Happy Path (Tasks):**
    1. Agent decides to call `sql:execute` with a `DROP TABLE` command.
    2. HITL Middleware detects the sensitive command and suspends the call.
    3. Middleware returns a specific `Suspended` status to the agent (holding the session).
    4. UI displays a notification: "Agent requesting approval for sql:execute".
    5. User reviews the SQL, clicks "Approve".
    6. Middleware resumes the execution and returns the result to the agent.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [HITL Middleware] -> (Wait for User) -> [Upstream]`
    * Uses a `Suspension` table in SQLite to track active requests.
    * Leverages a WebSocket event to notify the UI.
* **APIs / Interfaces:**
    * `POST /api/v1/approvals/{id}/approve`
    * `POST /api/v1/approvals/{id}/deny`
* **Data Storage/State:**
    Managed in the `mcpany.db` SQLite database to survive server restarts.

## 5. Alternatives Considered
* **Agent-side confirmation:** Rejected because agents can be "hallucinated" into skipping their own safety checks. Security must be enforced by the infrastructure.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Approval tokens must be cryptographically signed and tied to the specific user session.
* **Observability:** Track "Time to Approval" and "Rejection Rates" to identify bottlenecks in the human loop.

## 7. Evolutionary Changelog
* **2026-02-23:** Initial Document Creation.
