# Design Doc: Policy Firewall Engine (CEL-based)
**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
As the MCP ecosystem expands, static allow/deny lists for tool calls have become insufficient. Agents increasingly operate in shared environments where permissions must be context-sensitive (e.g., "Allow `read_file` only if the path is in `/tmp` and the agent intent is `testing`"). This document proposes a dynamic, expression-based Policy Firewall utilizing the **Common Expression Language (CEL)** to provide granular, safe-by-default governance.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a middleware that intercepts all `tools/call` and `resources/read` requests.
    * Evaluate each request against a set of CEL expressions.
    * Support access to request metadata (tool name, arguments, client ID) and session context (project, risk level).
    * Provide "Safe-by-Default" templates for common use cases (e.g., Read-Only, Sandbox-Only).
* **Non-Goals:**
    * Implementing a new expression language (use existing CEL libraries).
    * Handling LLM-side prompt injection (this is a tool-execution guardrail).

## 3. Critical User Journey (CUJ)
* **User Persona:** Enterprise AI Governance Lead
* **Primary Goal:** Restrict a "Research Agent" from accessing production credentials, even if it has access to a "Secret Manager" tool.
* **The Happy Path (Tasks):**
    1. Admin defines a policy: `request.tool == "get_secret" && !request.args.key.startsWith("prod_")`.
    2. Agent attempts to call `get_secret(key="prod_db_password")`.
    3. Policy Firewall intercepts the call, evaluates the CEL expression to `false`.
    4. Call is rejected with a 403 Forbidden error and a clear policy violation message.

## 4. Design & Architecture
* **System Flow:**
    `Client Request` -> `Auth Middleware` -> **`Policy Firewall (CEL Evaluation)`** -> `Routing` -> `Upstream Adapter`
* **APIs / Interfaces:**
    ```yaml
    # Example Policy Configuration
    policies:
      - id: restrict-prod-secrets
        description: "Prevent access to production secrets in non-prod sessions"
        expression: "request.tool == 'get_secret' ? !request.args.key.startsWith('prod_') : true"
        action: DENY
    ```
* **Data Storage/State:** Policies are loaded from the central `config.yaml` or a dedicated `policies/` directory. The engine is stateless per request.

## 5. Alternatives Considered
* **Rego (OPA):** Powerful but has a steeper learning curve and higher runtime overhead than CEL for simple tool-call expressions.
* **Hardcoded Middlewares:** Inflexible and requires a server recompile for every new policy change.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The firewall is the final gatekeeper before execution. It must fail-closed if an expression cannot be evaluated.
* **Observability:** Every policy evaluation (Allow/Deny) is logged to the audit trail with the specific expression ID that triggered it.

## 7. Evolutionary Changelog
* **2026-03-05:** Initial Document Creation.
