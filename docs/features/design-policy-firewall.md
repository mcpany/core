# Design Doc: Policy Firewall Engine
**Status:** Draft
**Created:** 2026-02-23

## 1. Context and Scope
As AI agents become more autonomous, the risk of unauthorized or dangerous tool calls increases. MCP Any needs a robust way to intercept, evaluate, and potentially block tool calls based on granular policies. The Policy Firewall Engine (PFE) acts as a security gatekeeper between the MCP client and the upstream adapters.

## 2. Goals & Non-Goals
* **Goals:**
    * Provide a unified policy enforcement point for all tool calls.
    * Support dynamic policy updates without server restarts.
    * Enable granular control (e.g., "Allow user 'Alice' to call 'read_file' but only in '/tmp'").
    * Integrate with external policy engines (Rego/OPA).
* **Non-Goals:**
    * Replacing upstream authentication (e.g., API keys).
    * Providing a full IAM system for agents.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Engineer
* **Primary Goal:** Restrict a subagent's ability to delete files to only a specific "scratch" directory.
* **The Happy Path (Tasks):**
    1. The engineer defines a Rego policy that checks the `path` argument of `delete_file`.
    2. The policy is loaded into MCP Any's configuration.
    3. An agent attempts to call `delete_file` with `/etc/passwd`.
    4. The PFE intercepts the call, evaluates the policy, and rejects the request with a 403 Forbidden error.
    5. The agent then attempts `delete_file` with `/tmp/scratch/data.txt`.
    6. The PFE evaluates and allows the call.

## 4. Design & Architecture
* **System Flow:**
    `Client Request -> MCP Any Core -> Policy Firewall Engine -> Upstream Adapter`
* **APIs / Interfaces:**
    * `PolicyEngineInterface`: `Evaluate(ctx, request) (decision, error)`
* **Data Storage/State:**
    * Policies are stored as part of the server configuration (YAML/JSON) or in a dedicated policy directory.

## 5. Alternatives Considered
* **Hardcoded Rules:** Rejected because it lacks flexibility and requires re-compilation.
* **Upstream-only Security:** Rejected because it requires every upstream service to implement its own security logic, leading to "Policy Drift".

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The PFE is the foundation of Zero Trust in MCP Any. It ensures no tool call is executed without explicit policy approval.
* **Observability:** Every policy decision is logged in the Audit Log, including the reason for denial.

## 7. Evolutionary Changelog
* **2026-02-23:** Initial Document Creation.
