# Design Doc: Policy Firewall Engine
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
As agents become more autonomous, the risk of "prompt injection" or "rogue behavior" increasing. MCP Any currently lacks a granular, declarative way to restrict tool calls based on context, user role, or dangerous patterns. The Policy Firewall Engine provides a middleware layer to evaluate tool calls against Rego (Open Policy Agent) or CEL (Common Expression Language) rules.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept all `tools/call` requests.
    *   Apply declarative policies (Allow/Deny/Request Approval).
    *   Provide context-aware policy evaluation (e.g., "Allow `fs:read` only in `/tmp`").
*   **Non-Goals:**
    *   Implementing the LLM itself.
    *   Providing a full-blown IAM system (it integrates with existing ones).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect
*   **Primary Goal:** Prevent an agent from deleting production database records without human approval.
*   **The Happy Path (Tasks):**
    1.  Architect defines a policy in `policy.rego` that blocks `DELETE` operations on services tagged `production`.
    2.  Agent attempts to call `delete_user`.
    3.  MCP Any Policy Firewall intercepts the call.
    4.  Rego engine evaluates the call against the policy and returns `DENY`.
    5.  MCP Any returns a standardized error to the agent explaining the policy violation.

## 4. Design & Architecture
*   **System Flow:**
    `Agent -> [JSON-RPC] -> MCP Any Core -> Policy Middleware -> Rego Engine -> [ALLOW] -> Upstream Adapter`
*   **APIs / Interfaces:**
    *   `POST /v1/policies`: Upload new policy files.
    *   `GET /v1/policies/status`: Check evaluation stats.
*   **Data Storage/State:**
    Policies are stored as part of the server configuration or in a dedicated `policies/` directory.

## 5. Alternatives Considered
*   **Hardcoded Rules:** Rejected because it's not flexible enough for enterprise needs.
*   **LLM-based Validation:** Rejected because it's non-deterministic and expensive.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Policies are evaluated in a local OPA sidecar or embedded library.
*   **Observability:** All policy evaluations are logged to the Audit Log with `Decision`, `RuleID`, and `Reason`.

## 7. Evolutionary Changelog
*   **2026-02-23:** Initial Document Creation.
