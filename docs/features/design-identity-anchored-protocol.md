# Design Doc: Identity-Anchored Header Protocol
**Status:** Draft
**Created:** 2026-02-27

## 1. Context and Scope
In complex multi-agent swarms, an initial request from a user may pass through several layers of agents (Orchestrator -> Researcher -> Analyst -> Coder). As the request moves deeper, the original user identity and intent context are often lost or replaced by the immediate parent agent's identity. This leads to "Identity Spreading," where a subagent might erroneously gain access to tools it shouldn't, or where audit logs fail to attribute actions to the end user.

## 2. Goals & Non-Goals
* **Goals:**
    * Cryptographically anchor the end-user identity to every tool call in a swarm.
    * Prevent privilege escalation by subagents.
    * Enable end-to-end auditability from tool call back to the originating user.
    * Provide a mechanism for "Intent Scoping" where a parent can restrict the subagent's temporary identity.
* **Non-Goals:**
    * Replacing existing OIDC/OAuth providers.
    * Managing long-term user passwords or credentials.

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Compliance Officer
* **Primary Goal:** Verify that a database deletion performed by a "Cleanup Subagent" was actually authorized by the original user "Alice."
* **The Happy Path (Tasks):**
    1. Alice initiates a task via the Orchestrator.
    2. MCP Any issues a signed "Identity Anchor" token to the Orchestrator.
    3. The Orchestrator passes this token to the Cleanup Subagent.
    4. The Cleanup Subagent calls the `delete_records` tool.
    5. MCP Any verifies the Identity Anchor, ensuring the token is still valid and within the "Alice" session.
    6. The tool call is executed, and the audit log records: `User: Alice | Agent: CleanupSubagent | Tool: delete_records`.

## 4. Design & Architecture
* **System Flow:**
    `User -> MCP Any (Issuer) -> Parent Agent -> Subagent -> MCP Any (Validator) -> Tool`
* **APIs / Interfaces:**
    * `X-MCP-Identity-Anchor`: A JWT or PASETO token containing `orig_user_id`, `session_id`, `intent_hash`, and `expiry`.
    * `X-MCP-Agent-Path`: A recursive header tracking the chain of agents (e.g., `orchestrator/researcher/analyst`).
* **Data Storage/State:**
    * Public keys for anchor verification are stored in the MCP Any internal keystore.

## 5. Alternatives Considered
* **Simple Header Passing**: Rejected because it's easily spoofed by a compromised subagent.
* **Centralized Session DB**: Rejected because it creates a performance bottleneck and doesn't work well in federated/distributed agent environments.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust)**: The anchor is "Intent-Bound." If a subagent tries to use the anchor for a task outside the original intent (detected via hash or LLM verification), the call is blocked.
* **Observability**: Every anchor verification failure is logged as a high-priority security event.

## 7. Evolutionary Changelog
* **2026-02-27:** Initial Document Creation.
