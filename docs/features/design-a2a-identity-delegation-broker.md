# Design Doc: A2A Identity Delegation Broker

**Status:** Draft
**Created:** 2026-04-24

## 1. Context and Scope
As AI agent swarms grow in complexity, specialized sub-agents often require access to tools or data originally authorized for a parent agent. Currently, this requires either manual re-authentication or the insecure sharing of high-privilege session tokens. The "Identity Shadowing" vulnerability (CVE-2026-31045) highlights the risks of improper token management in these handoffs.

MCP Any needs a secure mechanism to facilitate "Leased Identity," where a parent can delegate a subset of its authority to a sub-agent for a specific mission without compromising its primary identity.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Implement a lifecycle manager for Identity Delegation Tokens (IDT).
    *   Enable cryptographic binding of IDTs to specific sub-agents and mission intents.
    *   Provide real-time revocation of delegated authority.
    *   Support UAB-compliant identity metadata.
*   **Non-Goals:**
    *   Replacing framework-specific internal authentication (e.g., OpenClaw session tokens).
    *   Managing user-to-parent agent authentication (handled by primary gateway auth).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Heterogeneous Swarm Orchestrator
*   **Primary Goal:** Safely delegate "Read-Only" access to a private GitHub repo from a Senior Architect agent to a Documentation Specialist sub-agent.
*   **The Happy Path (Tasks):**
    1.  The Senior Architect agent requests an IDT from MCP Any, specifying the sub-agent ID, the GitHub tool scope, and a 1-hour TTL.
    2.  MCP Any validates the request against the Senior Architect's current session and issues a signed IDT.
    3.  The Senior Architect hands the IDT to the Documentation Specialist.
    4.  The Documentation Specialist calls the GitHub tool via MCP Any, providing the IDT.
    5.  MCP Any verifies the IDT signature, scope, and TTL before proxying the call to the GitHub MCP server.

## 4. Design & Architecture
*   **System Flow:**
    ```mermaid
    sequenceDiagram
        ParentAgent->>IdentityBroker: Request IDT (SubID, Scope, TTL)
        IdentityBroker-->>ParentAgent: Signed IDT
        ParentAgent->>SubAgent: Delegate Task (IDT)
        SubAgent->>MCP_Any_Gateway: Call Tool (IDT)
        MCP_Any_Gateway->>IdentityBroker: Validate IDT
        IdentityBroker-->>MCP_Any_Gateway: Success (Scoped Permissions)
        MCP_Any_Gateway->>MCPServer: Execute Tool
    ```
*   **APIs / Interfaces:**
    *   `POST /v1/identity/delegate`: Issue a new IDT.
    *   `GET /v1/identity/verify`: Internal endpoint for the gateway to validate tokens.
    *   `POST /v1/identity/revoke`: Invalidate an IDT before expiry.
*   **Data Storage/State:**
    *   Tokens are stored in the Shared KV Store (Blackboard) with a "Delegation" namespace, indexed by Hash and Sub-Agent ID.

## 5. Alternatives Considered
*   **Full Token Proxying**: Parent gives its own token to the sub-agent. Rejected due to lack of scoping and high risk of privilege escalation.
*   **Ephemeral Tool Passwords**: Generating one-time passwords for specific tool calls. Rejected due to high latency and lack of support for stateful multi-call sub-missions.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** IDTs must be mission-scoped. An IDT issued for "GitHub Read" cannot be used for "Slack Write," even by the same sub-agent.
*   **Observability:** All delegation events (issuance, usage, revocation) are logged in the Global Audit Trail, linked to the Root Mission Intent.

## 7. Evolutionary Changelog
*   **2026-04-24:** Initial Document Creation.
