# Design Doc: HITL Middleware
**Status:** Draft
**Created:** 2026-02-24

## 1. Context and Scope
"Human-in-the-Loop" (HITL) is essential for safety-critical agent operations. When an agent proposes an action that is irreversible or high-risk (e.g., spending money, deleting production data), the system must pause and wait for explicit human approval.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept "tagged" tool calls and suspend execution.
    * Notify a human user via UI or Webhook.
    * Resume or Cancel the tool call based on user input.
    * Provide a timeout mechanism for expired approval requests.
* **Non-Goals:**
    * Building a complex task management system.
    * Managing human workforce scheduling.

## 3. Critical User Journey (CUJ)
* **User Persona:** Financial Controller
* **Primary Goal:** Approve a $1,000 transaction proposed by a procurement agent.
* **The Happy Path (Tasks):**
    1. Agent calls `process_payment(amount: 1000)`.
    2. HITL Middleware detects `risk: high` tag.
    3. Middleware suspends the request and generates a unique `approval_id`.
    4. UI displays a notification to the Controller.
    5. Controller clicks "Approve".
    6. Middleware resumes the tool call with the original parameters.

## 4. Design & Architecture
* **System Flow:**
    `Request` -> `Risk Evaluator` -> `Suspend & Store State` -> `Wait for Signal` -> `Resume`.
* **APIs / Interfaces:**
    * `GET /api/v1/approvals`: List pending approvals.
    * `POST /api/v1/approvals/{id}/{action}`: Approve or Reject.
* **Data Storage/State:**
    Pending requests are stored in the "Blackboard" (SQLite) to survive restarts.

## 5. Alternatives Considered
* **Agent-Level Safety:** Rejected because it relies on the agent being "honest" about its safety constraints.
* **Synchronous UI Blocking:** Rejected; the backend must support asynchronous, multi-user approval flows.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Approval endpoints require strict RBAC.
* **Observability:** Approval time and the approving user ID are captured in audit logs.

## 7. Evolutionary Changelog
* **2026-02-24:** Initial Document Creation.
