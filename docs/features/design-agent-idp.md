# Design Doc: Agent Identity Provider (IdP) for Non-Human Identities (NHI)

**Status:** Draft
**Created:** 2026-03-04

## 1. Context and Scope
As AI agent swarms (like OpenClaw) become more complex, the lack of a standardized identity layer has become a critical security bottleneck. Agents currently operate with "borrowed" human credentials or no identity at all, making it impossible to audit, restrict, or verify inter-agent communications. This design introduces a dedicated Agent Identity Provider (IdP) within MCP Any to manage and verify Non-Human Identities (NHI).

## 2. Goals & Non-Goals
*   **Goals:**
    *   Issue cryptographically verifiable identities (DID-based or JWT-signed) to every agent and subagent.
    *   Enable "Identity-Bound Capability Tokens" that restrict tool access based on the specific agent's identity.
    *   Provide an "Identity Audit Log" that tracks which specific agent (not just which user) invoked a tool.
    *   Support "Delegated Identity" where a parent agent can spawn subagents with a subset of its own permissions.
*   **Non-Goals:**
    *   Replacing human OIDC/SAML providers (MCP Any IdP is for *agents*, not humans).
    *   Managing full PKI infrastructure (should rely on simple, local Ed25519 keypairs for now).

## 3. Critical User Journey (CUJ)
*   **User Persona:** OpenClaw Swarm Orchestrator.
*   **Primary Goal:** Ensure the "File Researcher" subagent can only read files, while the "Code Executor" can only run commands, even though they share the same parent session.
*   **The Happy Path (Tasks):**
    1.  Orchestrator requests a "Subagent Identity" from MCP Any for the "File Researcher."
    2.  MCP Any issues a signed JWS token bound to the `fs:read` capability.
    3.  The "File Researcher" includes this token in its MCP tool calls.
    4.  MCP Any's Policy Firewall verifies the token and identity before allowing the `read_file` call.
    5.  If the "File Researcher" attempts to call `shell_execute`, MCP Any rejects it as the identity lacks that capability.

## 4. Design & Architecture
*   **System Flow:**
    - **Identity Registry**: A local SQLite-backed store for active agent identities and their associated public keys.
    - **Token Issuance Service**: A local endpoint that signs identity tokens using the MCP Any master key.
    - **Verification Middleware**: Intercepts all incoming MCP requests, extracts the identity token, and validates it against the registry and policy engine.
*   **APIs / Interfaces:**
    - `/api/v1/identity/issue`: POST to create a new agent identity.
    - `/api/v1/identity/verify`: GET/POST to validate a token.
*   **Data Storage/State:** `identities` table in `mcpany.db` storing `agent_id`, `public_key`, `parent_id`, and `metadata`.

## 5. Alternatives Considered
*   **Using SPIFFE/Spire**: Too heavyweight for local-first agent development.
*   **Simple API Keys**: Hard to manage at scale for dynamic swarms and doesn't support delegation/inheritance easily.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is the core of NHI Zero Trust. It prevents "lateral movement" between agents in a swarm.
*   **Observability:** Every tool execution log will now be enriched with an `agent_id` field.

## 7. Evolutionary Changelog
*   **2026-03-04:** Initial Document Creation.
