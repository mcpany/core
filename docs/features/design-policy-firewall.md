# Design Doc: Rego-based Policy Firewall
**Status:** Draft
**Created:** 2025-05-22

## 1. Context and Scope
Autonomous agents are susceptible to prompt injection and malicious tool outputs. A static "Allow/Deny" list is insufficient for complex workflows. We need a dynamic policy engine that can evaluate the state, the tool being called, and the previous tool call sequence.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Integrate Open Policy Agent (OPA) / Rego for tool call validation.
    *   Support stateful policies (e.g., "Allow `rm` only if `ls` was called on the same path first").
    *   Provide clear "Policy Denied" error messages with remediation hints.
*   **Non-Goals:**
    *   Compiling Rego policies into Go code (use the OPA Go library).
    *   Handling application-level business logic (focus on infrastructure safety).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Admin
*   **Primary Goal:** Prevent an agent from deleting files outside of a specific directory, even if the LLM is tricked.
*   **The Happy Path (Tasks):**
    1.  Admin uploads a Rego policy to MCP Any.
    2.  Agent attempts a `filesystem:write` call to `/etc/passwd`.
    3.  Policy Firewall intercepts the call and evaluates it against the Rego policy.
    4.  Policy returns `deny` with reason "Path restricted".
    5.  MCP Any blocks the call and logs a security alert.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph LR
        Call[Tool Call] --> Middleware[Policy Middleware]
        Middleware --> OPA[OPA Engine]
        OPA -- Policy --> Rego[Rego Rules]
        OPA -- State --> History[Call History]
        OPA --> Decision{Decision}
        Decision -- Allow --> Upstream[Upstream Adapter]
        Decision -- Deny --> Error[Error Response]
    ```
*   **APIs / Interfaces:**
    *   `PolicyEngineInterface` with `Evaluate(request, context)` method.
    *   Internal storage for "Call History" to support stateful rules.

## 5. Alternatives Considered
*   **Hardcoded Go Policies**: Too rigid, requires recompilation.
*   **CEL (Common Expression Language)**: Good for simple rules, but less powerful than Rego for complex stateful logic.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The policy firewall itself must be protected from tampering.
*   **Observability:** Every policy decision must be logged in the Audit Log with the full evaluation context.

## 7. Evolutionary Changelog
*   **2025-05-22:** Initial Document Creation.
