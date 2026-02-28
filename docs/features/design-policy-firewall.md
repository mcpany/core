# Design Doc: Policy Firewall Engine (CEL/Rego)

**Status:** Draft
**Created:** 2026-02-28

## 1. Context and Scope
With the rise of autonomous agent swarms and the discovery of vulnerabilities like CVE-2026-2008 (Fermat-MCP), a static permission model is no longer sufficient. Agents need a way to verify that a tool call is not only authorized by the user but also safe in the current execution context. The Policy Firewall Engine provides a machine-enforceable layer using Common Expression Language (CEL) or Rego (Open Policy Agent) to validate tool inputs, outputs, and execution context.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a middleware that intercept tool calls and evaluates them against declarative policies.
    *   Support both CEL (performance-oriented) and Rego (expressiveness-oriented) policy languages.
    *   Enable "Intent-Aware" filtering (e.g., blocking a `DELETE` call if the high-level task is `READ_ONLY`).
    *   Provide a standardized way to define "Security Contracts" for MCP tools.
*   **Non-Goals:**
    *   Replacing the primary authentication layer (API Keys/MFA).
    *   Defining the policies themselves (MCP Any provides the engine and enforcement).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security Engineer.
*   **Primary Goal:** Prevent an agent from performing accidental mass-deletion in a production database.
*   **The Happy Path (Tasks):**
    1.  Engineer defines a Rego policy that restricts `sql_query` tool calls containing `DROP` or `TRUNCATE` unless the session has `SUPERUSER` scope.
    2.  An agent attempts to call `sql_query(query="DROP TABLE users")`.
    3.  The Policy Firewall Engine evaluates the call, matches the forbidden pattern, and rejects the request with a `403 Forbidden` error.
    4.  The action is logged in the Audit Log with the specific policy ID that triggered the rejection.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: The engine sits in the middleware pipeline between the `AuthMiddleware` and the `UpstreamAdapter`.
    - **Evaluation**: The `PolicyEvaluator` fetches relevant policies for the current tool and profile.
    - **Decision**: Policies return `allow`, `deny`, or `suspend` (triggering a HITL flow).
*   **APIs / Interfaces:**
    - `PolicyEngineInterface`: `Evaluate(context, toolCall) (Decision, error)`.
    - New config block: `policies: [{ name: "read-only-guard", type: "cel", script: "..." }]`.
*   **Data Storage/State:** Policies are stored in the `Configuration Store` and can be hot-reloaded.

## 5. Alternatives Considered
*   **Hardcoded Sanitization**: Writing regex-based filters for every tool. *Rejected* as it is unscalable and prone to bypasses.
*   **External OPA Server**: Requiring an external OPA deployment. *Rejected* to maintain the "Single Binary" principle of MCP Any, though we may support it as an optional upstream later.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The firewall is the core enforcer of the Zero Trust architecture, ensuring every call is validated at runtime.
*   **Performance**: CEL evaluation must be sub-millisecond to avoid impacting agent reasoning latency.

## 7. Evolutionary Changelog
*   **2026-02-28:** Initial Document Creation.
