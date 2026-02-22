# Design Doc: Policy Firewall Engine
**Status:** Draft
**Created:** 2026-02-22

## 1. Context and Scope
Autonomous agents operating in local environments pose a significant security risk. If an agent is compromised or hallucinations lead to destructive tool calls, the impact can be catastrophic. MCP Any needs a robust, programmable firewall that intercepts tool calls and validates them against predefined security policies before execution.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a policy-based interception layer for all tool calls.
    * Support CEL (Common Expression Language) or Rego for policy definitions.
    * Provide "Safe Defaults" for common tools (e.g., restrict `fs:write` to `/tmp`).
* **Non-Goals:**
    * Building a full Sandbox/Containerization solution (we rely on the host's isolation).
    * Blocking every possible prompt injection (focus is on tool-call validation).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Developer
* **Primary Goal:** Prevent an agent from deleting files outside of its designated workspace.
* **The Happy Path (Tasks):**
    1. Developer defines a policy: `allow(tool == "fs:delete") if path.startsWith("/home/jules/workspace")`.
    2. Agent attempts `fs:delete` on `/etc/passwd`.
    3. Policy Firewall Engine intercepts the call and evaluates the policy.
    4. The call is blocked, and an "Access Denied" error is returned to the agent.

## 4. Design & Architecture
* **System Flow:** Tool Call -> Policy Engine Middleware -> Upstream Adapter.
* **APIs / Interfaces:** `PolicyEngineInterface` with `Evaluate(request)` method.
* **Data Storage/State:** Policies loaded from `config.yaml` or a dedicated `policies/` directory.

## 5. Alternatives Considered
* **Hardcoded restrictions:** Rejected as it lacks flexibility for different user needs.
* **Model-based filtering:** Rejected because LLMs are susceptible to jailbreaking and cannot be trusted as the primary security layer.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The firewall is the core of the Zero Trust model for agents.
* **Observability:** Every blocked call must be logged in the Audit Log with the specific policy rule that triggered it.

## 7. Evolutionary Changelog
* **2026-02-22:** Initial Document Creation.
