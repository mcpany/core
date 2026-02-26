# Design Doc: Policy Firewall

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
As MCP Any becomes the central gateway for AI agents, it must enforce strict security boundaries. Agents often have broad access to tools that can perform destructive or sensitive operations. The Policy Firewall provides a Rego (Open Policy Agent) or CEL-based engine that intercepts every tool call to ensure it aligns with security policies, user intent, and "Zero Trust" principles.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Intercept and validate every `tools/call` request against a declarative policy.
    *   Support granular, context-aware rules (e.g., "Allow `fs:read` only in `/tmp`").
    *   Integrate with the `Policy Engine` to verify high-level intent.
    *   Provide clear audit logs and "denied" responses to agents.
*   **Non-Goals:**
    *   Implementing the underlying tools' security (e.g., OS-level permissions).
    *   Replacing authentication (handled by the Auth middleware).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Administrator.
*   **Primary Goal:** Prevent an agent from deleting production database records while allowing it to read logs.
*   **The Happy Path (Tasks):**
    1.  Administrator defines a Rego policy: `deny if input.tool == "db:delete" and input.env == "prod"`.
    2.  An agent attempts to call `db:delete` on a production instance.
    3.  Policy Firewall intercepts the call, evaluates the policy, and returns a `403 Forbidden` error.
    4.  The attempt is logged for compliance auditing.

## 4. Design & Architecture
*   **System Flow:**
    - **Hooking**: Every tool call passes through the `PolicyMiddleware`.
    - **Context Injection**: The middleware injects session metadata (user, role, intent) into the policy input.
    - **Evaluation**: The OPA/CEL engine evaluates the input against loaded `.rego` or `.cel` files.
    - **Action**: Based on the result (Allow/Deny/Audit), the call is either forwarded to the upstream MCP server or blocked.
*   **APIs / Interfaces:**
    - **Configuration**: `policies: [{ path: "./security.rego", type: "rego" }]`.
    - **Tool Interface**: No changes to standard MCP; returns `error` field in response on denial.
*   **Data Storage/State:** Policies are stored as part of the server configuration; audit logs are streamed to the structured log sink.

## 5. Alternatives Considered
*   **Hardcoded Rules in Go**: *Rejected* for lack of flexibility and difficulty in updating without restarts.
*   **Tool-Side Enforcement**: *Rejected* because it requires modifying every individual MCP server, violating the "Universal Adapter" principle.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The firewall is the core of the Zero Trust architecture. It assumes all tool calls are suspect until proven compliant.
*   **Observability:** Every policy decision is recorded in the `Audit Log` with the specific rule that triggered the action.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
