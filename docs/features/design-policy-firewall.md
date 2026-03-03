# Design Doc: Policy Firewall (Discovery & Runtime)

**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
As agents become more autonomous and tools more numerous, the risk of unauthorized tool execution or malicious tool injection increases. The Policy Firewall is the core security layer of MCP Any, responsible for enforcing Zero Trust principles. Recent findings on "Meta-Injection" (prompt injection in tool metadata) necessitate that this firewall must now operate not just at runtime (during tool calls), but also at discovery-time (during tool listing).

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enforce granular, capability-based access control (RBAC/ABAC) for all tool calls.
    *   Scan tool schemas and metadata for prompt injection patterns before they reach the agent.
    *   Provide a "Pre-flight" verification step for every MCP request.
    *   Support declarative policy definitions using Rego (Open Policy Agent) or CEL.
*   **Non-Goals:**
    *   Modifying the underlying LLM's weights or behavior.
    *   Implementing a general-purpose WAF for the host machine.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect.
*   **Primary Goal:** Ensure that a "Customer Support Agent" can only access `read_kb` and `update_ticket` tools, and that these tools do not contain hidden instructions in their descriptions.
*   **The Happy Path (Tasks):**
    1.  Architect defines a Rego policy restricting the `support_agent` role to a specific toolset.
    2.  The agent requests `tools/list`.
    3.  The Policy Firewall intercepts the list, scans metadata for anomalies, and filters the list based on the agent's role.
    4.  The agent calls `update_ticket`.
    5.  The Policy Firewall validates the input arguments against the policy before allowing the call to proceed to the upstream MCP server.

## 4. Design & Architecture
*   **System Flow:**
    - **Interception Layer**: A middleware that wraps the `mcpserver` handler.
    - **Metadata Scanner**: A specialized regex/LLM-based scanner that runs during `tools/list` and `tools/get`.
    - **Policy Evaluator**: Uses the `opa` Go library to evaluate Rego policies against the current context (user, agent, tool, arguments).
*   **APIs / Interfaces:**
    - Internal `PolicyEngine` interface with `EvaluateListing()` and `EvaluateCall()` methods.
    - `POST /api/policies` for updating Rego rules.
*   **Data Storage/State:** Policies are stored in the local filesystem or a protected SQLite table.

## 5. Alternatives Considered
*   **Simple ACLs**: Hardcoded lists of allowed tools. *Rejected* as it doesn't scale for complex enterprise requirements.
*   **Tool-Side Security**: Relying on upstream MCP servers to enforce security. *Rejected* because MCP Any must provide a "Unified" security layer for all (including third-party) upstreams.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is the primary concern. All tool access is "Denied by Default."
*   **Performance**: Policy evaluation must be ultra-fast (<5ms) to avoid degrading agent responsiveness.

## 7. Evolutionary Changelog
*   **2026-03-03:** Initial Document Creation. Expanded scope to include Metadata Scanning.
