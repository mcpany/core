# Design Doc: Policy Firewall Engine
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
With the rise of autonomous agent swarms, the risk of "Prompt Injection" leading to malicious tool execution has increased. MCP Any needs a robust, programmable firewall that can intercept and validate tool calls before they reach upstream services.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a Rego (Open Policy Agent) or CEL based engine for tool call validation.
    *   Support dynamic policy updates without server restart.
    *   Enable "Zero Trust" by default for all incoming tool requests.
*   **Non-Goals:**
    *   Providing a full OPA server (we will embed the engine).
    *   Validating the LLM's "intent" (focus is on the safety of the resulting tool call).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Prevent agents from calling `exec` on non-whitelisted commands or accessing sensitive files.
*   **The Happy Path (Tasks):**
    1. Architect defines a policy in `policy.rego`: `deny if tool == "exec" and not cmd in whitelist`.
    2. Agent attempts to call `exec` with `rm -rf /`.
    3. MCP Any Policy Firewall intercepts the call.
    4. Policy engine evaluates and returns a `deny`.
    5. Agent receives a standard MCP error: `403 Forbidden - Policy Violation`.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    graph LR
        Agent -->|ToolCall| PF[Policy Firewall]
        PF -->|Query| Engine[Rego Engine]
        Engine -->|Decision| PF
        PF -->|Allow| Upstream[Upstream Service]
        PF -->|Deny| Error[MCP Error Response]
    ```
*   **APIs / Interfaces:**
    *   Middleware Hook: `OnBeforeToolCall(request) -> Response`.
    *   Policy API: `PUT /api/v1/policies` for hot-reloading.
*   **Data Storage/State:**
    *   Policies stored in the configuration database (SQLite).

## 5. Alternatives Considered
*   **Hardcoded Whitelists**: Rejected as too inflexible for complex agent behaviors.
*   **Upstream-side Validation**: Rejected because many upstreams are third-party APIs that don't support custom logic.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The firewall itself must be secured. Only administrators can update policies.
*   **Observability:** Every policy decision (Allow/Deny) must be logged for audit purposes.

## 7. Evolutionary Changelog
*   **2026-02-23:** Initial Document Creation.
