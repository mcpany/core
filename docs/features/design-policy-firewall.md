# Design Doc: Policy Firewall (Zero-Trust Tool Guard)

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As AI agents gain more autonomy, the risk of unauthorized or destructive actions increases. Standard API keys are insufficient because they lack granularity—an agent with a GitHub token might accidentally delete a repository when it only meant to read a file. The Policy Firewall provides a Rego/CEL based middleware layer that inspects every tool call against fine-grained security policies before execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide granular, capability-based access control for all MCP tool calls.
    * Support dynamic policy updates without server restarts.
    * Enable "Virtual Patching" by blocking known vulnerable tool patterns.
    * Support "Intent-Aware" scoping (e.g., allow `write` only if the user's high-level intent matches).
* **Non-Goals:**
    * Replacing upstream service authentication (e.g., OAuth).
    * Implementing an LLM-based firewall (this is a deterministic rule engine).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Engineer.
* **Primary Goal:** Prevent an autonomous agent from performing `DELETE` operations on production databases while allowing `SELECT` on specific tables.
* **The Happy Path (Tasks):**
    1. Security Engineer defines a Rego policy: `deny { input.tool == "db_query"; input.args.sql.contains("DELETE"); input.env == "production" }`.
    2. Engineer uploads the policy to MCP Any via the CLI/API.
    3. An agent attempts to call `db_query` with a `DELETE` statement.
    4. The Policy Firewall intercepts the call, evaluates the Rego policy, and returns a `403 Forbidden` error to the agent.
    5. The attempt is logged in the Audit Trail for review.

## 4. Design & Architecture
* **System Flow:**
    - **Interception**: Every `tools/call` request is routed through the `PolicyMiddleware`.
    - **Evaluation**: The middleware extracts `tool_name`, `arguments`, `user_id`, and `context_headers` and passes them to the OPA (Open Policy Agent) or CEL engine.
    - **Action**: Based on the policy result (`allow`, `deny`, `suspend_for_approval`), the middleware either proceeds, blocks, or triggers the HITL flow.
* **APIs / Interfaces:**
    - `POST /api/v1/policies`: Upload new policy definitions.
    - `GET /api/v1/policies/evaluate`: Dry-run a tool call against current policies.
* **Data Storage/State:** Policies are stored in the internal SQLite database and cached in memory for high-performance evaluation.

## 5. Alternatives Considered
* **Hardcoded Rules in Config**: Simple, but not flexible enough for complex enterprise requirements.
* **LLM Guardrails**: Too slow and non-deterministic for a core infrastructure layer.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The firewall itself must be secured. Only authorized admins can update policies.
* **Observability:** Every policy evaluation (hit/miss/block) is recorded with full context for SOC2 compliance.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation. Integration with Vulnerability Intelligence Middleware planned.
