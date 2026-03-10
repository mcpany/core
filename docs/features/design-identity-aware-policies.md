# Design Doc: Identity-Aware Tool Policy Engine
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
As AI agent swarms (like OpenClaw) become more complex, the "Agent-to-Agent" (A2A) interaction surface increases. Today, MCP Any applies policies globally or per-session. However, this allows a low-privilege subagent to potentially "trick" a higher-privilege agent into executing a dangerous tool call (Sociality Illusion). MCP Any needs to scope permissions to the specific identity of the agent making the request.

## 2. Goals & Non-Goals
* **Goals:**
    * Implement a verifiable Agent Identity token (JWT-based).
    * Allow policy definitions that reference `agent.id` and `agent.role`.
    * Prevent a subagent from using tools not explicitly granted to its identity.
* **Non-Goals:**
    * Implementing a full-blown IAM system (e.g., Auth0).
    * Managing human user identities (this is for Agent identities).

## 3. Critical User Journey (CUJ)
* **User Persona:** Swarm Orchestrator
* **Primary Goal:** Ensure the "Researcher" agent can read files, but the "Writer" agent can only write to a specific directory, even if they share the same MCP Any session.
* **The Happy Path (Tasks):**
    1. Orchestrator registers "Researcher" and "Writer" agents with MCP Any.
    2. MCP Any issues ephemeral Identity Tokens for each.
    3. "Researcher" calls `read_file` with its token; MCP Any validates the token and allows the call.
    4. "Writer" calls `read_file` with its token; MCP Any rejects the call based on the Identity-Aware Policy.

## 4. Design & Architecture
* **System Flow:**
    Agent -> [Request + Identity Token] -> MCP Any Gateway -> Policy Engine [Check agent.role against tool.required_role] -> MCP Server.
* **APIs / Interfaces:**
    * `POST /v1/identities/register`: Create a new agent identity.
    * `mcp_identity` header: Required for all tool execution calls.
* **Data Storage/State:**
    * Identities stored in the internal SQLite "Blackboard."

## 5. Alternatives Considered
* **Session-only isolation:** Rejected because multi-agent swarms often share a single session for context persistence but need different privilege levels.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** Tokens must be short-lived and cryptographically signed.
* **Observability:** All tool calls must log the Agent ID that initiated them.

## 7. Evolutionary Changelog
* **2026-03-10:** Initial Document Creation.
