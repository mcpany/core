# Design Doc: Policy Firewall Engine
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
As AI agents become more autonomous, the risk of them performing unauthorized or dangerous actions increases. Existing MCP servers often lack granular control over what an agent can actually do once a tool is called. The Policy Firewall Engine (PFE) aims to provide a robust, programmable security layer that intercepts all tool calls and evaluates them against a set of user-defined policies.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a Zero-Trust middleware for all tool calls.
    * Support declarative policy languages (Rego/CEL).
    * Enable granular field-level inspection of tool arguments.
    * Support "Block," "Allow," and "Alert" actions.
* **Non-Goals:**
    * Implementing the LLM itself.
    * Managing upstream API keys (handled by Upstream Adapters).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-conscious Enterprise Admin
* **Primary Goal:** Prevent an autonomous agent from calling `delete_resource` on any production database without manual approval.
* **The Happy Path (Tasks):**
    1. Admin defines a Rego policy that checks if the tool name is `delete_resource` and the `env` argument is `production`.
    2. Admin registers the policy in `mcpany` via config.
    3. An agent attempts to call `delete_resource` with `env: production`.
    4. PFE intercepts the call, evaluates the policy, and blocks the request.
    5. PFE returns a clear security error to the agent and logs the attempt.

## 4. Design & Architecture
* **System Flow:**
    `Agent Request` -> `MCP Core` -> **`Policy Firewall Engine (Middleware)`** -> `Rego Evaluator` -> `Decision (Allow/Block)` -> `Upstream Adapter` (if Allowed)
* **APIs / Interfaces:**
    * `PolicyEngineInterface`: `Evaluate(context, request) (Decision, error)`
* **Data Storage/State:**
    * Policies are stored as `.rego` or `.yaml` files and loaded into an in-memory evaluator for high performance.

## 5. Alternatives Considered
* **Hardcoded Policies:** Rejected due to lack of flexibility for complex enterprise needs.
* **Upstream-only Validation:** Rejected because it requires modifying every upstream service; PFE provides a centralized "Gatekeeper" approach.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Policies are enforced at the core, meaning no tool call can bypass the firewall regardless of the adapter used.
* **Observability:** Every policy evaluation is logged with a trace ID, allowing for auditability and debugging of blocked actions.

## 7. Evolutionary Changelog
* **2026-02-22:** Initial Document Creation.
