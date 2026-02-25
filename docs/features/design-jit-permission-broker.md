# Design Doc: JIT Permission Broker

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
Current AI agent deployments rely on static "capability tokens" assigned at the start of a session. However, in complex multi-agent swarms, a subagent may discover a need for a specific tool (e.g., `fs:write` to a new directory) that wasn't pre-authorized. This leads to "Permission Deadlock," where the agent fails because it cannot proceed without human intervention. The JIT Permission Broker solves this by allowing agents to request "Intent-Based Escalation" in real-time.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enable real-time permission escalation based on validated intent.
    *   Integrate with the Policy Firewall to evaluate risk scores for escalation requests.
    *   Provide a "Temporary Grant" mechanism that automatically expires after the task is completed.
    *   Reduce Human-in-the-Loop (HITL) friction for low-risk escalations.
*   **Non-Goals:**
    *   Allowing permanent permission changes without human approval.
    *   Replacing the Zero-Trust foundation (it extends it with dynamic logic).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Autonomous DevOps Agent.
*   **Primary Goal:** The agent is fixing a bug and needs to write to a log directory it doesn't have access to. It needs to acquire this permission without waiting for a human.
*   **The Happy Path (Tasks):**
    1.  The agent attempts to call `fs:write` and receives a "Permission Denied" error with a challenge from MCP Any.
    2.  The agent calls the `request_escalation` tool, providing the intent ("Writing debug logs for bug fix X") and the required capability.
    3.  The JIT Permission Broker evaluates the request against the current session's risk budget.
    4.  Since the intent aligns with the high-level task and the risk is low, the Broker grants a temporary token.
    5.  The agent successfully executes the `fs:write` call.

## 4. Design & Architecture
*   **System Flow:**
    - **Escalation Request**: An agent-facing tool `mcpany_request_escalation`.
    - **Risk Evaluator**: A middleware that uses LLM-based intent analysis and Rego policies to score the request.
    - **Token Vending**: If approved, a short-lived capability token is injected into the session's context.
*   **APIs / Interfaces:**
    - `tools/request_escalation(intent: string, capability: string)`
    - Internal `PolicyFirewall.EvaluateEscalation(context, request)`
*   **Data Storage/State:** Escalation history and temporary tokens are stored in the `Shared KV Store`.

## 5. Alternatives Considered
*   **Pre-authorizing All Tools**: Giving agents broad permissions. *Rejected* due to security risks (Zero-Trust violation).
*   **Mandatory HITL**: Always asking a human. *Rejected* due to latency and scalability issues in autonomous swarms.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The primary risk is "Agent Hijacking" where an agent is tricked into requesting an escalation for a malicious purpose. Mitigation: Use "Risk Budgets" (e.g., an agent can only escalate $X$ amount of risk per hour) and strict intent-to-task mapping.
*   **Observability:** All escalations (granted or denied) are logged in the Audit Trail and visualized in the Security Dashboard.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
