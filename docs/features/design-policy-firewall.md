# Design Doc: Policy Firewall Engine
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
With the rise of autonomous agents like Claude Code and OpenClaw, AI agents are increasingly performing actions on local machines and production environments. A simple "allow/deny" model is no longer sufficient. MCP Any needs a robust "Policy Firewall" that can inspect tool calls in real-time and enforce complex, context-aware security rules.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept every MCP `tools/call` request.
    * Evaluate requests against Rego (Open Policy Agent) or CEL (Common Expression Language) policies.
    * Support "Least Privilege" access (e.g., restrict filesystem access to specific directories).
    * Provide detailed audit logs for policy decisions.
* **Non-Goals:**
    * Implementing a new policy language (we will use OPA/Rego or CEL).
    * Providing a GUI for policy authoring (initially config-driven).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Operations Engineer / AI Developer
* **Primary Goal:** Prevent an autonomous agent from reading sensitive files (e.g., `/etc/passwd`) while allowing it to read project files.
* **The Happy Path (Tasks):**
    1. User defines a Rego policy in `policies/fs_access.rego`.
    2. User configures MCP Any to use this policy for the `filesystem` service.
    3. Agent attempts to call `read_file` with path `/etc/passwd`.
    4. Policy Firewall intercepts the call, evaluates it, and returns an "Access Denied" error to the agent.
    5. Agent receives the error and continues with a different approach.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> MCP Any Server -> [Policy Firewall Middleware] -> Upstream Adapter -> Upstream Service`
* **APIs / Interfaces:**
    * New middleware component `pkg/middleware/policy_firewall`.
    * Policy Engine interface: `Evaluate(ctx, request) (Decision, error)`.
* **Data Storage/State:**
    * Policies are loaded from the filesystem or a remote URL.
    * Policy evaluation is stateless per request.

## 5. Alternatives Considered
* **Hardcoded Rules:** Too inflexible for enterprise use cases.
* **Upstream-side Security:** Many upstreams don't support granular security; MCP Any provides a centralized point of enforcement.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The firewall is the core of our Zero Trust strategy.
* **Observability:** Every policy evaluation (Allow/Deny) is logged to the Audit Log with the reason.

## 7. Evolutionary Changelog
* **2026-02-23:** Initial Document Creation.
