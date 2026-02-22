# Design Doc: Recursive Scope Enforcer

**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
In complex agent swarms, a parent agent might delegate a task to a subagent. Currently, MCP Any lacks a mechanism to ensure that the subagent's capabilities are a strict subset of the parent's capabilities. This "Privilege Escalation" risk is a significant barrier to deploying autonomous swarms in production.

The Recursive Scope Enforcer will allow MCP Any to track the "Call Chain" and enforce decreasing privilege levels.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Introduce `X-MCP-Context-Scope` headers for all inter-agent calls.
    *   Validate that any sub-call scope is "lesser than or equal to" the parent scope.
    *   Provide a standard library of scopes (e.g., `fs:read`, `net:connect`).
*   **Non-Goals:**
    *   Dynamic scope negotiation (initially static based on config).
    *   Third-party IAM integration (OAuth2 etc. are separate concerns).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Ensure a subagent spawned to "read logs" cannot "delete databases" even if it somehow gets access to the tool.
*   **The Happy Path (Tasks):**
    1.  Parent Agent is granted `logs:read:*`.
    2.  Parent Agent calls Subagent with `logs:read:/var/log/app.log`.
    3.  MCP Any verifies `logs:read:/var/log/app.log` is a subset of `logs:read:*`.
    4.  If Subagent tries to call `db:delete`, MCP Any blocks it because it's outside the inherited scope.

## 4. Design & Architecture
*   **System Flow:**
    `[Agent A (Scope: *)] -> [MCP Any] -> [Agent B (Scope: X)] -> [MCP Any] -> [Tool]`
    MCP Any checks Scope X against Agent A's scope, and then checks the Tool call against Scope X.
*   **APIs / Interfaces:**
    *   Middleware: `ScopeValidationMiddleware`
    *   Policy Language: CEL (Common Expression Language) for scope matching.
*   **Data Storage/State:**
    *   Scoped tokens stored in the session context.

## 5. Alternatives Considered
*   **Rego (OPA):** Powerful but might be overkill/too slow for high-frequency tool calls. CEL is faster and easier to embed in Go.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Core feature for Zero Trust agent execution.
*   **Observability:** Scope violations logged with high severity.

## 7. Evolutionary Changelog
*   **2026-02-22:** Initial Document Creation.
