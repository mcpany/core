# Design Doc: Policy Firewall Engine
**Status:** In Review
**Created:** 2026-02-18

## 1. Context and Scope
As AI agents become more autonomous, they are being granted access to increasingly sensitive tools (e.g., database deletion, financial transactions, shell execution). Relying solely on the LLM's "internal alignment" is insufficient for enterprise-grade security.

The **Policy Firewall Engine** is a middleware layer for MCP Any that intercepts all tool calls and evaluates them against a set of declarative policies. It ensures that tool execution is permitted based on the current context, user role, and session state, effectively creating a "Zero Trust" environment for agents.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a centralized engine for validating tool calls before execution.
    * Support declarative policy languages (Rego/Open Policy Agent).
    * Enable context-aware decisions (e.g., "Allow `delete_user` only if the user was created in the last 1 hour").
    * Log all policy evaluations for auditing and compliance.
* **Non-Goals:**
    * Replacing upstream authentication (e.g., API keys).
    * Modifying the tool's internal logic.
    * Providing a general-purpose LLM guardrail (this is specifically for tool calls).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise Security Architect
* **Primary Goal:** Prevent an agent from performing high-risk actions without explicit approval or matching strict criteria.
* **The Happy Path (Tasks):**
    1. The Architect defines a Rego policy that blocks `rm -rf` commands in the Command Adapter.
    2. An AI Agent attempts to call the `execute_command` tool with `rm -rf /`.
    3. MCP Any intercepts the request and sends the tool name and arguments to the Policy Firewall Engine.
    4. The Engine evaluates the policy and returns a `DENY` decision.
    5. MCP Any returns a protocol-compliant error to the agent, explaining the policy violation.
    6. The event is logged in the Audit Log with the specific policy rule that triggered the denial.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph LR
        Agent[AI Agent] -->|Tool Call| Core[MCP Any Core]
        Core -->|Intercept| PFE[Policy Firewall Engine]
        PFE -->|Fetch Rules| Store[Policy Store]
        PFE -->|Decision: ALLOW/DENY| Core
        Core -->|If ALLOW| Adapter[Upstream Adapter]
        Core -->|If DENY| Agent
    ```
* **APIs / Interfaces:**
    * `Check(ctx, ToolCallRequest) -> PolicyDecision`
    * `policy_id`, `rule_name`, `decision` (ALLOW/DENY/HITL), `reason`.
* **Data Storage/State:**
    * Policies stored as `.rego` files or in a dedicated SQLite table for dynamic updates.

## 5. Alternatives Considered
* **Hardcoded Rules in Adapters**: Rejected because it's not scalable and lacks flexibility.
* **LLM-based Validation**: Rejected because it's non-deterministic and subject to prompt injection.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The engine itself must be isolated. Policies should be cryptographically signed to prevent tampering.
* **Observability:** Every decision is logged. Metrics track the number of denied requests per agent/tool.

## 7. Evolutionary Changelog
* **2026-02-18:** Initial Document Creation.
