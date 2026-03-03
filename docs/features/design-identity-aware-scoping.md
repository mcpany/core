# Design Doc: Identity-Aware Tool Scoping

**Status:** Draft
**Created:** 2026-03-03

## 1. Context and Scope
As agents move from simple single-agent executors to complex multi-agent swarms (e.g., OpenClaw, CrewAI), the "identity" of the calling subagent becomes a critical security boundary. Currently, tools are authorized based on the gateway's global credentials. MCP Any must transition to an "Identity-Aware" model where tool access is scoped to the specific role or identity of the calling agent (e.g., a "Critic" agent shouldn't have "Write" access to a database).

## 2. Goals & Non-Goals
*   **Goals:**
    *   Enable agents to "Self-Identify" via a new `mcp_identity` header or metadata field.
    *   Enforce granular access control policies (using Rego/CEL) based on the caller's identity.
    *   Support "Delegated Authority" where a parent agent can grant a subset of its permissions to a subagent.
    *   Integrate with the `Policy Firewall` for real-time enforcement.
*   **Non-Goals:**
    *   Implementing a full-blown IAM system (e.g., OAuth2 server).
    *   Verifying the internal logic of the agent (we trust the identity provided via secure channels).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Lead Systems Architect for an Agent Swarm.
*   **Primary Goal:** Ensure the "Code Reviewer" agent cannot accidentally (or maliciously) execute "Delete" commands on the production DB.
*   **The Happy Path (Tasks):**
    1.  Architect defines an "Identity Policy" in MCP Any: `identity == 'reviewer' -> tool_action != 'delete'`.
    2.  The Swarm Orchestrator calls MCP Any, passing `mcp_identity: "reviewer"`.
    3.  If the Reviewer agent attempts to call a `delete_record` tool, MCP Any blocks the call and returns a `403 Forbidden (Identity Policy Violation)`.
    4.  The action is logged in the Audit Trail with the "reviewer" identity attached.

## 4. Design & Architecture
*   **System Flow:**
    - **Header Extraction**: Middleware extracts identity metadata from incoming JSON-RPC or HTTP requests.
    - **Policy Lookup**: Queries the `Policy Engine` with the `(Identity, Tool, Action)` tuple.
    - **Context Enrichment**: If allowed, the tool call is enriched with the identity before being sent to the upstream MCP server (allowing for upstream identity-aware logging).
*   **APIs / Interfaces:**
    - Standardized `mcp_identity` field in the `meta` object of MCP requests.
    - Policy definition schema extensions for identity-based rules.
*   **Data Storage/State:** Policies are stored in the main `config.yaml` or as separate `.rego` files.

## 5. Alternatives Considered
*   **Unique API Keys per Subagent**: Give every subagent its own API key. *Rejected* as it becomes a management nightmare for dynamic swarms.
*   **Process Isolation**: Run every subagent in its own MCP Any instance. *Rejected* as it wastes resources and breaks shared state.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is the core of "Zero Trust" for multi-agent systems.
*   **Observability:** The "Identity" must be a first-class dimension in all logs and metrics.

## 7. Evolutionary Changelog
*   **2026-03-03:** Initial Document Creation.
