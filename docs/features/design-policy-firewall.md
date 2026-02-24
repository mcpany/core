# Design Doc: Policy Firewall

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the rise of autonomous agents and specialized subagents (e.g., OpenClaw swarms), the risk of "Agentic Overreach" has become critical. Agents often have broad access to tools that can perform destructive actions or leak sensitive data. The Policy Firewall is a Zero-Trust execution engine that intercepts every tool call to ensure it complies with defined security policies, particularly focusing on subprocess isolation and input sanitization.

## 2. Goals & Non-Goals
* **Goals:**
    * Intercept and validate all MCP tool calls against a central policy engine.
    * Enforce strict sandboxing for any tool call that triggers a subprocess.
    * Sanitize inputs and outputs to prevent shell injection and data leakage.
    * Provide a declarative way (Rego/CEL) to define tool-level permissions.
* **Non-Goals:**
    * Hardcoding security rules for every possible tool (must be policy-driven).
    * Replacing the underlying OS-level security (provides an additional application-layer boundary).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security-Conscious Agent Operator.
* **Primary Goal:** Prevent an autonomous agent from executing `rm -rf /` even if it has access to a "Shell" tool.
* **The Happy Path (Tasks):**
    1. Operator defines a policy that restricts the "Shell" tool to a specific workspace directory.
    2. Agent attempts to call `shell_execute(command="rm -rf /")`.
    3. Policy Firewall intercepts the call and evaluates it against the Rego policy.
    4. The call is rejected with a "Policy Violation" error.
    5. The attempt is logged in the Sandbox Security Dashboard for review.

## 4. Design & Architecture
* **System Flow:**
    - **Interception**: Every `tools/call` request passes through the `PolicyMiddleware`.
    - **Evaluation**: The middleware sends the call context (tool name, arguments, agent identity) to the Policy Engine (OPV/CEL).
    - **Enforcement**: If approved, the call proceeds. If it involves a subprocess, it is wrapped in a containerized or restricted environment.
* **APIs / Interfaces:**
    - Internal `EvaluatePolicy(context)` interface.
    - Admin API for updating policies: `POST /v1/policies`.
* **Data Storage/State:** Policies are stored in a version-controlled repository or the internal SQLite store.

## 5. Alternatives Considered
- **Standard Linux Permissions (chmod/chown)**: *Rejected* as too coarse-grained for granular tool/agent-level scoping.
- **Model-Side Guardrails**: *Rejected* as easily bypassed via prompt injection; security must be enforced at the infrastructure layer.

## 6. Cross-Cutting Concerns
- **Security (Zero Trust)**: Follows the principle of least privilege. Deny-by-default for all tool calls unless an explicit policy allows them.
- **Observability**: Every policy decision (Allow/Deny) is tracked and visible in the UI Timeline.

## 7. Evolutionary Changelog
- **2026-02-27**: Initial Document Creation. Focused on subprocess sandboxing and input sanitization in response to recent ecosystem vulnerabilities.
