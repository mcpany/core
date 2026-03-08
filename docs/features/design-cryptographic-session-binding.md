# Design Doc: Cryptographic Session Binding
**Status:** Draft
**Created:** 2026-03-06

## 1. Context and Scope
With the rise of "AI Predator Swarms" (Hivenets), a simple API key or static tool signature is no longer enough to prove that a tool call is legitimate. Attackers can hijack an existing, authorized agent and "poison" its context to perform unauthorized actions. Cryptographic Session Binding ensures that every tool call is part of a verified, human-initiated session, preventing "orphaned" or "hijacked" agentic flows from escalating privileges.

## 2. Goals & Non-Goals
* **Goals:**
    * Bind every tool call to a specific, unique `session_id` signed by the initiator (Human or High-Trust Proxy).
    * Require a "Chain of Trust" (Provenance) from the initial user request down to the leaf subagent tool call.
    * Implement "Time-Bound Tokens" that expire if not used within a narrow window.
    * Prevent "Out-of-Band" tool calls from malicious scripts that aren't part of the active agentic flow.
* **Non-Goals:**
    * Eliminating the need for tool-level permissions.
    * Encrypting the tool payloads (Focus is on provenance and binding).

## 3. Critical User Journey (CUJ)
* **User Persona:** Security Engineer
* **Primary Goal:** Prevent a malicious website from sending a JSON-RPC request to the local MCP Any server to execute a tool (like `shell_execute`).
* **The Happy Path (Tasks):**
    1. Human user initiates a task via a verified UI/CLI.
    2. MCP Any generates a `SessionToken` (signed by its private key) and binds it to the user's `Identity`.
    3. The Agent receives the `SessionToken` and must include a "Subagent-Signed" variant in every tool call.
    4. MCP Any validates that the `SessionToken` is active, has not expired, and matches the originating session's `Identity` and `RootIntent`.
    5. A malicious request (e.g., from a web browser without the `SessionToken`) is rejected immediately as "Unbound Request."

## 4. Design & Architecture
* **System Flow:**
    ```mermaid
    sequenceDiagram
        User->>MCP Any: Start Session (Auth: JWT/MFA)
        MCP Any-->>User: SessionToken (Signed: MCP_Any_Key)
        User->>Agent: Run Task (Token: SessionToken)
        Agent->>MCP Any: Tool Call (Tool: list_files, Token: SessionToken + SubagentSign)
        MCP Any->>Auth Engine: VerifyBinding(SessionToken, Call)
        Auth Engine-->>MCP Any: Result (VALID)
        MCP Any->>Tool: Execute
    ```
* **APIs / Interfaces:**
    * Every JSON-RPC request now includes a mandatory `x-session-binding` header or parameter.
* **Data Storage/State:**
    * Short-lived `SessionMetadata` in-memory (Redis or similar) for high-performance validation.

## 5. Alternatives Considered
* **IP-based filtering:** *Rejected* because it doesn't prevent "Same-Host" attacks (e.g., from a browser to `localhost`).
* **Persistent API Keys:** *Rejected* because they can be leaked and used indefinitely.

## 6. Cross-Cutting Concerns
* **Security (Zero Trust):** The "Root of Trust" is the initial human-initiated auth (e.g., OAuth/OIDC or local MFA).
* **Observability:** Audit logs will record the "Binding Chain" for every tool call to enable post-incident forensics.

## 7. Evolutionary Changelog
* **2026-03-06:** Initial Document Creation.
