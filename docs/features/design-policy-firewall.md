# Design Doc: Policy Firewall Engine
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
As AI agents become more autonomous, the risk of "prompt injection" leading to unauthorized tool execution increases. MCP Any currently lacks a granular, declarative way to restrict tool calls based on context, user role, or resource sensitivity. The Policy Firewall Engine aims to provide a "Zero Trust" layer between the agent and the upstream tools.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a middleware layer that intercepts all `tools/call` requests.
    * Support declarative policy definitions using Rego (Open Policy Agent) or CEL.
    * Allow policies to access call context (e.g., arguments, user metadata, agent history).
    * Provide a default-deny posture for high-sensitivity tools.
* **Non-Goals:**
    * Implementing identity management (AuthN). This system assumes identity is already resolved.
    * Modifying upstream tools to enforce security.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Administrator
* **Primary Goal:** Prevent an agent from performing `fs:write` on `/etc/` while allowing it on `/tmp/`.
* **The Happy Path (Tasks):**
    1. Admin defines a Rego policy file.
    2. Admin registers the policy with MCP Any via configuration.
    3. Agent attempts to call `fs:write` with path `/etc/passwd`.
    4. Policy Firewall intercepts the call, evaluates the Rego script, and returns a "Permission Denied" error to the agent.
    5. Agent attempts to call `fs:write` with path `/tmp/output.txt`.
    6. Policy Firewall approves, and the call proceeds to the upstream.

## 4. Design & Architecture
* **System Flow:**
    `Agent -> [JSON-RPC Handler] -> [Policy Firewall Middleware] -> [Adapter] -> [Upstream Tool]`
* **APIs / Interfaces:**
    * `PolicyEngine` interface with `Evaluate(ctx, toolCall) (Result, error)`.
* **Data Storage/State:**
    Policies are loaded from the configuration directory and cached in memory.

## 5. Alternatives Considered
* **Hardcoded Rules:** Rejected because it lacks flexibility for complex enterprise requirements.
* **Upstream Enforcement:** Rejected because many existing APIs don't have the concept of "agent-aware" permissions.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The firewall is the central enforcement point. Fail-closed on evaluation errors.
* **Observability:** All policy decisions (Allow/Deny) are logged with the trace ID and evaluation reason.

## 7. Evolutionary Changelog
* **2026-02-23:** Initial Document Creation.
