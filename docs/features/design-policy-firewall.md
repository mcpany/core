# Design Doc: Policy Firewall Engine

**Status:** Draft
**Created:** 2026-02-07

## 1. Context and Scope
As AI agents become more autonomous, they increasingly require access to sensitive tools and data. However, granting broad access introduces significant security risks, including the "Lethal Trifecta" (bridging private and untrusted data). MCP Any needs a centralized, programmable mechanism to intercept and validate every tool call before it reaches an upstream service.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept all `tools/call` requests in the middleware pipeline.
    *   Support declarative policy definitions using Rego (Open Policy Agent) or CEL.
    *   Enable granular access control based on tool name, arguments, user identity, and session context.
    *   Provide a "Dry Run" mode for policy testing.
*   **Non-Goals:**
    *   Implement an identity provider (it should integrate with existing ones).
    *   Enforce policies on the upstream service itself (this is a gateway-level firewall).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security Administrator
*   **Primary Goal:** Prevent an agent from calling a "write" tool on a production database if the session context contains untrusted data from a public URL.
*   **The Happy Path (Tasks):**
    1.  Admin defines a Rego policy that blocks `db:write` if `session.has_untrusted_data` is true.
    2.  An agent attempts to call `db:insert_user`.
    3.  The Policy Firewall intercepts the call.
    4.  The Engine evaluates the request against the Rego policy.
    5.  The Engine returns a "Forbidden" error to the agent, preventing the call.

## 4. Design & Architecture
*   **System Flow:**
    `Client Request` -> `Auth Middleware` -> `Context Assembly` -> **`Policy Engine`** -> `Upstream Adapter`
*   **APIs / Interfaces:**
    *   `CheckAccess(Request) (Response, Error)`: Core interface for the engine.
*   **Data Storage/State:** Policies are stored in the configuration store and hot-reloaded.

## 5. Alternatives Considered
*   **Hardcoded Rules:** Rejected because they lack the flexibility needed for diverse enterprise environments.
*   **Upstream-only Security:** Rejected because many upstreams lack granular MCP-aware permission systems.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The firewall is the primary enforcer of Zero Trust. Policies should default to "Deny All."
*   **Observability:** All policy evaluations (Permit/Deny) must be logged in the Audit Log with reasons.

## 7. Evolutionary Changelog
*   **2026-02-07:** Initial Document Creation.
