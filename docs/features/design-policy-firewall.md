# Design Doc: Policy Firewall Engine

**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
As agents become more autonomous, the risk of them executing dangerous or unauthorized tool calls increases. MCP Any needs a way to inspect, validate, and potentially block tool calls before they reach the upstream services. The Policy Firewall Engine provides a programmable layer for enforcing these security and compliance rules.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept all `tools/call` requests.
    *   Evaluate requests against Rego (OPA) or CEL (Common Expression Language) policies.
    *   Support granular access control based on tool name, arguments, and session context.
    *   Allow for "dry-run" and "alert-only" modes.
*   **Non-Goals:**
    *   This is not a network firewall (IP/Port level).
    *   It does not handle authentication (handled by Auth Middleware).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Admin
*   **Primary Goal:** Prevent agents from calling `delete_database` on production environments during non-working hours.
*   **The Happy Path (Tasks):**
    1.  Admin defines a CEL policy: `request.tool == "delete_database" && request.args.env == "prod" && now.hour() >= 18`.
    2.  Agent attempts to call `delete_database` with `env: "prod"` at 7 PM.
    3.  Policy Firewall intercepts the call.
    4.  Policy evaluates to `true` (deny).
    5.  MCP Any returns an error to the agent: "Tool call blocked by Policy Firewall: Production deletions not allowed after hours."

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    sequenceDiagram
        Client->>+Core: tools/call
        Core->>+Firewall: Evaluate(Request)
        Firewall->>+Engine: Match(Rules)
        Engine-->>-Firewall: Allow/Deny
        Firewall-->>-Core: Decision
        alt Allow
            Core->>+Upstream: Execute
            Upstream-->>-Core: Result
            Core-->>-Client: Result
        else Deny
            Core-->>Client: Error (Blocked)
        end
    ```
*   **APIs / Interfaces:**
    *   `FirewallInterface`: `Evaluate(ctx, request) (Decision, error)`
*   **Data Storage/State:**
    *   Policies are stored as YAML/JSON configuration files or loaded from a remote OPA server.

## 5. Alternatives Considered
*   **Hardcoded Rules:** Too inflexible for enterprise needs.
*   **Javascript Hooks:** High security risk and performance overhead. Rego/CEL are safer and more performant for policy evaluation.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The firewall itself must be isolated and have its own resource limits.
*   **Observability:** Every firewall decision must be logged in the Audit Log with the specific rule that triggered the action.

## 7. Evolutionary Changelog
*   **2026-02-23:** Initial Document Creation.
