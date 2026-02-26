# Design Doc: Universal Policy-as-Code Engine (Policy Firewall)

**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
With the rise of autonomous agents like OpenClaw and background execution in Claude Code, traditional tool allow-lists are insufficient. Agents need to execute complex tasks that may involve sensitive file operations or network calls. MCP Any must provide a centralized, programmable policy layer that can evaluate the safety of a tool call based on the agent's identity, the current session context, and the specific parameters of the call.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide a unified policy enforcement point (PEP) for all MCP tool calls.
    *   Support industry-standard policy languages: Rego (Open Policy Agent) and CEL (Common Expression Language).
    *   Enable "Intent-Aware" governance by inspecting session context and recursive lineage.
    *   Seamlessly integrate with the HITL (Human-in-the-Loop) middleware for high-risk escalations.
*   **Non-Goals:**
    *   Defining the specific security policies for every possible tool (MCP Any provides the engine, users provide the policy).
    *   Implementing identity management (MCP Any relies on upstream agent framework tokens).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect.
*   **Primary Goal:** Prevent an autonomous subagent from deleting any file outside of its assigned project directory, even if the parent agent has broader permissions.
*   **The Happy Path (Tasks):**
    1.  Architect defines a Rego policy that restricts `fs_delete` calls to a specific `base_path` passed in the `SessionContext`.
    2.  An autonomous subagent attempts to call `fs_delete(path="/etc/passwd")`.
    3.  The Policy Firewall intercepts the JSON-RPC call.
    4.  The engine evaluates the Rego policy using the tool parameters and the current `Recursive Context` headers.
    5.  The policy returns `deny`.
    6.  MCP Any rejects the tool call with a `403 Forbidden` error and logs the violation to the Audit Trail.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception**: Every `tools/call` request passes through the `PolicyFirewallMiddleware`.
    - **Evaluation**: The middleware extracts the tool name, arguments, and context headers, passing them to the Rego/CEL evaluator.
    - **Decision**: If the policy allows, the call proceeds. If it denies, the call is blocked. If it requires "approval," it is routed to the HITL Middleware.
*   **APIs / Interfaces:**
    - **Policy Management API**: CRUD endpoints for managing Rego/CEL policy files.
    - **Validation API**: Test endpoint for dry-running policies against mock tool calls.
*   **Data Storage/State:** Policies are stored as `.rego` or `.yaml` files in the `config/policies` directory. Evaluation results are cached for performance.

## 5. Alternatives Considered
*   **Hardcoded Rules in Middleware**: Fast but inflexible. *Rejected* because it doesn't scale to complex enterprise requirements.
*   **Agent-Side Enforcement**: Relies on the agent to be "good." *Rejected* because it violates Zero Trust principles.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** The Policy Engine itself must be protected. Only authorized administrators can modify policies.
*   **Observability:** Every policy evaluation (Allow/Deny/Escalate) is recorded in the Audit Log with the associated Trace ID.

## 7. Evolutionary Changelog
*   **2026-02-27:** Initial Document Creation.
