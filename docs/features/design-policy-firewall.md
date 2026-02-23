# Design Doc: Policy Firewall Engine
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
As AI agents become more autonomous, the risk of "prompt injection" leading to unauthorized tool execution increases. MCP Any needs a robust way to intercept and validate tool calls before they reach the upstream services. The Policy Firewall Engine provides a declarative way to define security rules that govern tool execution.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept all `tools/call` requests.
    *   Evaluate requests against Rego (OPA) or CEL (Common Expression Language) policies.
    *   Support granular access control based on tool name, arguments, and agent metadata.
    *   Provide clear audit logs for blocked requests.
*   **Non-Goals:**
    *   Modifying the upstream service response (handled by other middleware).
    *   Implementing a full-blown IAM system (it integrates with existing ones).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security Engineer
*   **Primary Goal:** Prevent an agent from calling the `delete_database` tool unless the request includes a specific `approval_token`.
*   **The Happy Path (Tasks):**
    1.  Engineer defines a CEL policy in `config.yaml`: `tool == "delete_database" && !request.metadata.has("approval_token")`.
    2.  Agent attempts to call `delete_database`.
    3.  Policy Firewall intercepts the call.
    4.  The expression evaluates to `true` (blocked).
    5.  MCP Any returns a `MethodForbidden` error to the agent and logs the attempt.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph LR
        Client[AI Agent] -->|Tool Call| Middleware[Policy Firewall]
        Middleware -->|Check| Engine[Policy Engine (Rego/CEL)]
        Engine -->|Allowed| Upstream[Upstream Adapter]
        Engine -->|Blocked| Client
    ```
*   **APIs / Interfaces:**
    *   New configuration block `policy_firewall` under global or service-level config.
    *   `CheckAccess(ctx, req) (bool, error)` internal interface.
*   **Data Storage/State:**
    *   Policies are loaded from configuration files.
    *   Stateless evaluation; can integrate with external data sources for dynamic lookups.

## 5. Alternatives Considered
*   **Hardcoded Rules:** Rejected due to lack of flexibility for enterprise users.
*   **Post-execution Validation:** Rejected because it's too late; the damage (e.g., data deletion) would already be done.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Policies default to "Deny All" if misconfigured.
*   **Observability:** Every policy evaluation (Allow/Deny) is logged with the specific rule ID that matched.

## 7. Evolutionary Changelog
*   **2026-02-23:** Initial Document Creation.
