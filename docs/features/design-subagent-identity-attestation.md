# Design Doc: Subagent Identity Attestation
**Status:** Draft
**Created:** 2026-03-10

## 1. Context and Scope
With the rise of hierarchical agent frameworks like OpenClaw, complex tasks are increasingly delegated to specialized subagents. However, current MCP implementations lack a standardized way to verify the identity and authorized intent of these subagents. This creates a security gap where a compromised subagent (or one spawned via a malicious local configuration) could call sensitive tools without the parent agent's knowledge or consent. MCP Any needs to provide a cryptographic "Chain of Custody" to ensure that every tool call in a hierarchy is fully authorized.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a cryptographic signature scheme for tool calls that traces back to the root agent.
    *   Provide a mechanism for parent agents to "lease" specific capabilities to subagents with time-bound and intent-bound constraints.
    *   Enable tools to verify the entire "Agent Chain" before execution.
*   **Non-Goals:**
    *   Replacing existing LLM-based authentication (e.g., OpenAI API keys).
    *   Managing the internal state of third-party agent frameworks.

## 3. Critical User Journey (CUJ)
*   **User Persona:** Enterprise Security Architect / Lead AI Engineer
*   **Primary Goal:** Ensure that subagents spawned by a "Research Agent" can only read files and cannot perform write operations or network calls, even if the subagent's local configuration is tampered with.
*   **The Happy Path (Tasks):**
    1.  Root Agent initializes a session with MCP Any and receives a "Root Identity Token."
    2.  Root Agent spawns a "Subagent" and issues an "Attested Delegation Token" containing specific allowed tools and an "Intent Scope" (e.g., "Read Research Papers").
    3.  Subagent calls a tool (e.g., `fs:read`) through MCP Any, passing the Delegation Token.
    4.  MCP Any validates the signature chain, checks the Intent Scope against the Policy Engine, and executes the tool.
    5.  Subagent attempts to call `fs:write`. MCP Any rejects the call because it is outside the Attested Delegation Scope.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    sequenceDiagram
        participant Root as Root Agent
        participant Any as MCP Any Gateway
        participant Sub as Subagent
        participant Tool as Upstream Tool

        Root->>Any: Authenticate & Get Root Token
        Root->>Sub: Spawn with Delegation Token (Signed by Root)
        Sub->>Any: call_tool(fs:read, Token)
        Any->>Any: Verify Signature Chain (Root -> Sub)
        Any->>Any: Validate Intent Scope
        Any->>Tool: Execute Read
        Tool-->>Any: Data
        Any-->>Sub: Result
    ```
*   **APIs / Interfaces:**
    *   `auth/delegate`: New endpoint for agents to create signed delegation tokens for subagents.
    *   `tools/call`: Updated to accept an optional `x-mcp-attestation` header containing the identity chain.
*   **Data Storage/State:**
    *   Tokens are short-lived and stateless (JWT-based) to ensure performance in high-scale OpenClaw deployments.
    *   Public keys for root agents are stored in the Service Registry.

## 5. Alternatives Considered
*   **Centralized Session Store**: Rejected due to latency and scalability concerns in distributed agent swarms. Cryptographic chains allow for decentralized verification.
*   **Simple Capability Tokens**: Rejected because they don't provide provenance. A subagent could leak a token to a third party; a signed chain ensures the call came from the authorized hierarchy.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** Mandatory signature verification for all hierarchical calls. Integration with the Path-Normalized Config Validator to prevent "Shadow Token" injections.
*   **Observability:** Every entry in the audit log will include the full `Agent-Chain-ID` for forensic analysis.

## 7. Evolutionary Changelog
*   **2026-03-10:** Initial Document Creation.
