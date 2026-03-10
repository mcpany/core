# Design Doc: Autonomous Loop Governor

**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
With the introduction of autonomous patterns like the "Ralph-Loop" (automated `/continue` injection) in tools like Claude Code, AI agents can now enter long-running execution loops without human intervention. While this boosts productivity, it introduces risks of "Token Runaway," infinite recursion, and unexpected costs. MCP Any needs to provide an infrastructure-level "Circuit Breaker" to govern these loops and ensure they operate within safe, user-defined boundaries.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a "Circuit Breaker" for autonomous agent loops.
    * Enforce per-session budgets for tool calls, token usage, and execution time.
    * Provide a standardized "Heartbeat" mechanism for agents to attest their progress.
    * Require manual user intervention (via HITL) when safety thresholds are exceeded.
* **Non-Goals:**
    * Implementing the agent's loop logic itself.
    * Replacing the agent's internal state management.

## 3. Critical User Journey (CUJ)
* **User Persona:** Developer using an autonomous coding agent.
* **Primary Goal:** Ensure a long-running "Refactor Project" task doesn't spend more than $50 or run for more than 30 minutes without a status check.
* **The Happy Path (Tasks):**
    1. User starts a session with an "Autonomous Budget" (e.g., max 50 tool calls, $10 token limit).
    2. Agent enters a `ralph-loop` to perform the refactor.
    3. MCP Any tracks every tool call and its estimated cost.
    4. At the 40th call (80% threshold), MCP Any injects a warning into the tool output.
    5. At the 50th call, MCP Any suspends the session and triggers a "Heartbeat Attestation" request via the UI.
    6. User reviews the progress and grants an additional budget of 20 calls.
    7. Agent resumes and completes the task.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> MCP Any (Governance Middleware) -> Tool Call Execution`
    - The `GovernanceMiddleware` intercepts all JSON-RPC requests.
    - It checks the `SessionContext` for remaining budget (tokens, calls, time).
    - If the budget is exceeded, it returns a standardized `GOVERNANCE_SUSPENDED` error code, which the `HITL Middleware` uses to trigger a user approval flow.
* **APIs / Interfaces:**
    - **Session Config**: New fields `budget_tool_calls`, `budget_tokens`, `budget_duration`.
    - **Governance Header**: `X-MCP-Loop-Heartbeat` (used by agents to signal intent to continue).
* **Data Storage/State:**
    - Session budgets and real-time consumption are tracked in the `Shared KV Store` (Blackboard) for durability across agent restarts.

## 5. Alternatives Considered
* **Agent-Side Budgeting**: Relying on the agent to track its own budget. *Rejected* because agents can hallucinate or fail to enforce their own limits during a runaway loop.
* **Cloud-Side Billing Alerts**: Using cloud provider alerts. *Rejected* because they are often too slow (hours behind) and don't provide granular per-task control.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Budget settings must be immutable after session initialization unless authorized by a high-privilege user token.
* **Observability:** The UI must display a "Fuel Gauge" showing the current budget consumption of the active loop.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
