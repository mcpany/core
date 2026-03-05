# Design Doc: Policy Firewall (Rego/CEL)

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
As AI agents gain more autonomy, the risk of "Goal Drift" and unauthorized tool execution increases. Simple static permissions are insufficient for dynamic agent swarms. MCP Any needs a robust, declarative Policy Firewall that can intercept tool calls and evaluate them against sophisticated rules (e.g., "only allow file writes in `/tmp` if the agent is in `coding-mode`").

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a middleware that intercepts all `tools/call` requests.
    *   Support Rego (Open Policy Agent) or CEL (Common Expression Language) for policy definitions.
    *   Provide "Intent-Aware" validation by checking the `_mcp_context` headers.
    *   Enable "Human-in-the-Loop" (HITL) triggers based on policy outcomes.
*   **Non-Goals:**
    *   Defining the specific business policies (MCP Any provides the *mechanism*, users provide the *policy*).
    *   Replacing host-level security (e.g., AppArmor/SELinux).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security Administrator.
*   **Primary Goal:** Prevent an agent from making any network calls to internal IP ranges unless it's a "Verified Admin Agent."
*   **The Happy Path (Tasks):**
    1.  Admin uploads a Rego policy file to the MCP Any configuration.
    2.  An agent attempts to call a tool that targets `http://192.168.1.50/reboot`.
    3.  The Policy Firewall intercepts the call and extracts the target URL.
    4.  The Policy Engine evaluates the Rego rule against the tool arguments and agent context.
    5.  Evaluation returns `deny`. The tool call is aborted with a `403 Forbidden` error.

## 4. Design & Architecture
*   **System Flow:**
    - **Intercept**: The `PolicyFirewallMiddleware` wraps the tool execution pipeline.
    - **Evaluation**: The middleware sends the tool input and agent metadata to the internal `OPA/CEL Runner`.
    - **Enforcement**: If denied, execution stops. If allowed, it proceeds to the next middleware.
*   **APIs / Interfaces:**
    - `POST /v1/policies`: Upload/Update policies.
    - `GET /v1/policies/evaluate`: Test a policy against a mock payload.
*   **Data Storage/State:** Policies are stored as part of the server configuration, potentially persisted in the `Shared KV Store`.

## 5. Alternatives Considered
*   **Hardcoded Rules**: Faster but lacks flexibility for complex enterprise needs.
*   **LLM-based Validation**: Uses another LLM to "check" the call. *Rejected* due to latency and the risk of the "checker" also hallucinating.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Policy Firewall is the "Policy Enforcement Point" (PEP) in our Zero Trust architecture.
*   **Observability:** Every policy evaluation (Allow/Deny) is logged to the Audit Log with the full evaluation trace.

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
