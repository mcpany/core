# Design Doc: Policy Firewall Engine
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
As agents become more autonomous, the risk of "Prompt Injection" leading to unauthorized or dangerous tool calls increases. Current MCP implementations lack a granular, policy-driven layer to intercept and validate tool calls before they reach the upstream service. The Policy Firewall Engine (PFE) addresses this by providing a Rego or CEL-based middleware that evaluates every tool call against a set of security policies.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept all `tools/call` requests.
    * Evaluate request parameters against declarative policies (e.g., "Allow `git push` only to `origin/main`").
    * Support both Rego (Open Policy Agent) and CEL (Common Expression Language).
    * Provide detailed audit logs for policy violations.
* **Non-Goals:**
    * Implementing identity management (authentication).
    * Modifying tool responses (this is handled by the DLP/Context Optimizer).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Operations Engineer (SecOps)
* **Primary Goal:** Prevent an AI agent from deleting production database tables even if it's tricked by a prompt injection.
* **The Happy Path (Tasks):**
    1. SecOps defines a policy in `policies/production.rego` that blocks `DROP TABLE` commands for the `postgres` service.
    2. The agent receives a malicious prompt and attempts to call `sql_query` with `DROP TABLE users;`.
    3. MCP Any intercepts the call.
    4. The PFE evaluates the query parameter against the Rego policy.
    5. The policy returns a `deny` decision.
    6. MCP Any returns a structured error to the agent: `Security policy violation: DROP TABLE is not allowed.`
    7. An alert is logged in the audit stream.

## 4. Design & Architecture
* **System Flow:**
    `Agent` -> `MCP Any Core` -> `Middleware: Policy Firewall` -> `Policy Evaluator (Rego/CEL)` -> `Upstream Adapter`
* **APIs / Interfaces:**
    New configuration block in `config.yaml`:
    ```yaml
    policies:
      - name: "sql-safety"
        engine: "rego"
        source: "path/to/policy.rego"
        selector: "service == 'db-prod' && tool == 'query'"
    ```
* **Data Storage/State:**
    Policies are loaded from disk and compiled at startup. Hot-reload is supported via the config watcher.

## 5. Alternatives Considered
* **Hardcoded Validation**: Rejected due to lack of flexibility and difficulty in updating rules.
* **Upstream-side Validation**: Rejected because it doesn't provide a unified interface across different protocols (HTTP, gRPC, CLI).

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The PFE is the core of our Zero Trust architecture. Policies should follow "Default Deny".
* **Observability**: Every decision (Allow/Deny) is recorded with metadata including the rule name and the failing parameter.

## 7. Evolutionary Changelog
* **2026-02-23**: Initial Document Creation.
