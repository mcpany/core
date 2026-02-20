# Design Doc: HITL Middleware
**Status:** Draft
**Created:** 2025-02-17

## 1. Context and Scope
"Human-in-the-Loop" (HITL) is essential for high-stakes agent actions. Currently, agents often execute tools blindly. This middleware allows MCP Any to pause tool execution and wait for human approval via a UI or CLI.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept "Sensitive" tool calls based on configuration.
    * Suspend the MCP session and notify the user.
    * Resume or abort execution based on user feedback.
* **Non-Goals:**
    * Providing a chat UI for the user (handled by MCP Any UI).
    * Handling multi-user approvals (initially single-user).

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using an autonomous coding agent.
* **Primary Goal:** Approve a "git push" or "rm -rf" command before it runs.
* **The Happy Path (Tasks):**
    1. Agent calls a tool marked as `requires_approval: true`.
    2. HITL Middleware pauses the request and generates a unique `RequestID`.
    3. UI displays a notification: "Agent wants to run 'rm -rf /'. Approve?"
    4. User clicks "Approve".
    5. Middleware resumes the tool call to the upstream service.

## 4. Design & Architecture
* **System Flow:**
    `Tool Call` -> `HITL Middleware` -> `State Store (Suspended)` -> `Wait for Signal` -> `Resume` -> `Upstream`
* **APIs / Interfaces:**
    * `/api/v1/approvals`: GET list of pending approvals, POST to approve/deny.
* **Data Storage/State:**
    * Temporary suspension state in the shared KV store (SQLite).

## 5. Alternatives Considered
* **Agent-side Approval**: Rejected because it relies on the agent being "well-behaved". Enforcement must happen at the gateway.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Approval tokens must be short-lived and cryptographically signed.
* **Observability:** Track approval latency and user response times.

## 7. Evolutionary Changelog
* **2025-02-17:** Initial Document Creation.
