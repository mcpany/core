# Design Doc: Policy Firewall
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
With the rise of autonomous agents like OpenClaw, AI models are increasingly given access to sensitive host-level capabilities (shell execution, filesystem access). The current "allow-all" or "static config" models are insufficient for enterprise or secure personal use. MCP Any needs a dynamic, rule-based interceptor that validates tool calls against a security policy before they reach the upstream adapter.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept every MCP `tools/call` request.
    * Evaluate request parameters against Rego (Open Policy Agent) or CEL (Common Expression Language) rules.
    * Support "Zero Trust" by default (deny all unless explicitly allowed).
    * Provide detailed audit logs for blocked requests.
* **Non-Goals:**
    * Implementing the sandbox environment itself (this is handled by the Upstream Adapters).
    * Handling authentication (handled by Auth Middleware).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Local Agent User
* **Primary Goal:** Prevent an agent from executing `rm -rf /` while allowing `ls` in a specific directory.
* **The Happy Path (Tasks):**
    1. User defines a Rego policy file (`policy.rego`) restricting `shell_command` tool to a whitelist of commands.
    2. User starts MCP Any with `--policy-path=policy.rego`.
    3. Agent sends a request to execute `ls -la /tmp`.
    4. Policy Firewall evaluates the request, matches it against the whitelist, and allows it.
    5. Agent sends a request to execute `rm -rf /`.
    6. Policy Firewall evaluates the request, finds no matching allow rule, and returns a `Permission Denied` error to the agent.

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    graph TD
        Agent[AI Agent] -->|tools/call| Core[MCP Any Core]
        Core -->|Intercept| PF[Policy Firewall]
        PF -->|Query| Engine[OPA/CEL Engine]
        Engine -->|Policy File| Rules[Rego/CEL Rules]
        Engine -->|Decision| PF
        PF -->|Allow| Adapter[Upstream Adapter]
        PF -->|Deny| Error[Error Response]
        Adapter --> Upstream[Backend API/CLI]
    ```
* **APIs / Interfaces:**
    * `PolicyEngine` interface with `Evaluate(request ToolCall) (Decision, error)` method.
* **Data Storage/State:**
    * Policies are stored as `.rego` or `.yaml` files.
    * Audit logs are stored in the SQLite `mcpany.db`.

## 5. Alternatives Considered
* **Hardcoded Whitelists**: Rejected as too inflexible for complex agent workflows.
* **LLM-based Validation**: Rejected due to latency, cost, and potential for "jailbreaking" the validator.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The firewall is the core of our Zero Trust implementation. It must fail-closed if the policy engine crashes.
* **Observability:** Every decision (Allow/Deny) must be logged with the associated rule ID for debugging.

## 7. Evolutionary Changelog
* **2026-02-23:** Initial Document Creation.
