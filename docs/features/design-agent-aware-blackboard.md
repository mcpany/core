# Design Doc: Agent-Aware Blackboard Isolation
**Status:** Draft
**Created:** 2026-03-09

## 1. Context and Scope
The `Shared KV Store` (Blackboard) is a core tool in MCP Any that allows multiple agents and subagents to share state. However, as agent swarms become more complex (e.g., OpenClaw's multi-agent refinement), there is an increasing risk of "Cross-Agent State Injection." A malicious or misconfigured subagent could overwrite critical state used by a parent agent or other specialized subagents, leading to hallucinations or security bypasses. MCP Any must implement "Agent-Aware" isolation to ensure each agent can only access and modify state within its authorized "Intent Scope."

## 2. Goals & Non-Goals
* **Goals:**
    * Implement mandatory "Agent Identity" headers for all Blackboard tool calls.
    * Provide row-level security (RLS) within the SQLite-based KV store.
    * Define "Shared Intent Scopes" where multiple agents can explicitly collaborate on the same data.
    * Prevent unauthorized data exfiltration between unrelated agent sessions.
* **Non-Goals:**
    * Encrypting the data at rest (handled by other security layers).
    * Restricting *total* storage usage (handled by quota management).

## 3. Critical User Journey (CUJ)
* **User Persona:** Multi-Agent Swarm Orchestrator (e.g., OpenClaw).
* **Primary Goal:** Ensure the "Security Auditor" subagent cannot modify the "Code Generator" subagent's temporary state.
* **The Happy Path (Tasks):**
    1. Parent Agent creates a session with a unique `AgentID` and `IntentScope`.
    2. Subagent A is spawned with `AgentID: Sub-A` and inherits the same `IntentScope`.
    3. Subagent A writes a temporary variable to the Blackboard.
    4. Subagent B (different `IntentScope`) attempts to read Subagent A's variable.
    5. MCP Any rejects the request with a "Permission Denied: Intent Mismatch" error.

## 4. Design & Architecture
* **System Flow:**
    `Agent (with Auth Headers)` -> `MCP Any (Blackboard Tool)` -> `SQLite with RLS Middleware`
    1. **Identity Injection**: The `Recursive Context Protocol` injects `AgentID` and `IntentScope` into all tool calls.
    2. **Access Control**: The Blackboard tool uses these headers to append `WHERE agent_id = ? OR intent_scope = ?` to all SQL queries.
    3. **Audit**: Every access is logged with the associated `AgentID`.
* **APIs / Interfaces:**
    * `Blackboard.get(key, scope_override?)`
    * `Blackboard.set(key, value, scope_override?)`
* **Data Storage/State:**
    * `blackboard.db`: SQLite table with columns `(key, value, agent_id, intent_scope, created_at)`.

## 5. Alternatives Considered
* **Separate Databases per Agent**: Rejected due to high overhead and difficulty in sharing state when *explicitly* requested.
* **In-Memory only**: Rejected because long-running agent swarms need persistence across restarts.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: Follows the principle of least privilege. By default, an agent only sees its own data.
* **Observability**: The `Recursive Context Dashboard` will visualize which agents are accessing which parts of the Blackboard.

## 7. Evolutionary Changelog
* **2026-03-09:** Initial Document Creation.
### Update: 2026-03-10 - Hardening Swarm Isolation with Intent-Bound Cryptography
**Context**: Multi-agent refinement loops in OpenClaw are vulnerable to cross-agent state injection.
**Architecture Adjustment**:
* **Cryptographic Intent-Scopes**: Transitioning from string-based scopes to cryptographically signed tokens.
* **Read-Only Shared State**: By default, shared state from a parent is read-only for subagents unless explicit "Write" capabilities are granted via the policy engine.
* **Automatic Cleanup**: Intent-bound data is automatically purged when the parent session expires, preventing long-term state leakage.
**Security Impact**: Prevents "Cross-Agent State Poisoning," ensuring that compromised or misbehaving subagents cannot influence the execution of the parent or other specialized agents.

### Update: 2026-11-02 - Integration with Memory Integrity Verification (MIV)
**Context:** Disclosure of "Memory Injection" (Lakera AI) and "Sleeper Agent" vulnerabilities confirms that persistent state isolation is insufficient.
**Architecture Adjustment:**
* Extending the Blackboard to support **MIV-compliant** integrity scanning.
* Implementing a "State Verification Sidecar" that periodically re-attests the semantic alignment of keys against the Mission Root.
* Introducing "Shadow-Snapshotting" to allow the MIV service to scan memory without locking active agent reasoning loops.
**Security Impact:** Detects and neutralizes "Sleeper" instructions injected via third-party data sources before they can be activated by the agent.
