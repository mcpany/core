# Design Doc: HITL Verification Protocol
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the rise of autonomous agents like OpenClaw that can execute shell commands, browser automation, and financial transactions, there is a critical need for a standardized way to pause execution and request human consent. Currently, each agent framework handles human-in-the-loop (HITL) differently, leading to inconsistent security postures. MCP Any aims to solve this by providing a universal HITL Verification Protocol at the gateway level.

## 2. Goals & Non-Goals
* **Goals:**
    * Standardize the JSON-RPC communication between agents, MCP Any, and the user for approval requests.
    * Provide rich context (diffs, risk scores, justifications) to the user during approval.
    * Support asynchronous suspension and resumption of agent sessions.
* **Non-Goals:**
    * Implementing the UI for approval (this is handled by the MCP Any UI or third-party integrations).
    * Replacing existing framework-specific HITL (but providing a bridge to them).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer using an autonomous agent swarm.
* **Primary Goal:** Prevent an agent from accidentally deleting a production database or making an unauthorized API call.
* **The Happy Path (Tasks):**
    1. Agent attempts to call a "high-risk" tool (e.g., `shell_execute` with `rm -rf`).
    2. MCP Any Policy Firewall intercepts the call and identifies it as requiring HITL.
    3. MCP Any suspends the tool call and sends a `verification/request` to the user's registered UI.
    4. User receives a notification with the command, a risk assessment, and a "Justification" from the agent.
    5. User clicks "Approve" in the MCP Any Dashboard.
    6. MCP Any resumes the tool call and returns the execution result to the agent.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [MCP Any Gateway (Policy Engine)] -> (Suspend Call) -> [Notification Service] -> [User UI] -> [Approval] -> [MCP Any Gateway] -> (Resume Call) -> Upstream Tool`
* **APIs / Interfaces:**
    * `mcpany/suspend`: Internal signal to pause a request.
    * `verification/request`: JSON-RPC notification sent to the client/UI.
    * `verification/respond`: JSON-RPC method for the user to submit an approval/denial.
* **Data Storage/State:**
    * Suspended requests are stored in the Shared KV Store (SQLite) with a TTL.

## 5. Alternatives Considered
* **Framework-level HITL**: Rejected because it doesn't provide a central point of governance across different agent types (OpenClaw vs. Claude Code).
* **Always-on Confirmation**: Rejected as it destroys agent autonomy and developer productivity for low-risk tasks.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Requests are signed; approval tokens are single-use and session-bound.
* **Observability:** All HITL events (request, approval, denial, timeout) are logged in the audit trail.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
