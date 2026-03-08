# Design Doc: Cryptographic Agent Identity (A2A-ID)

**Status:** Draft
**Created:** 2026-03-05

## 1. Context and Scope
The rapid expansion of agent swarms and multi-agent coordination (A2A) has introduced a major security gap: the lack of verifiable identity. In a mesh of agents, how does a "Coding Agent" know that the "Research Agent" requesting a file write is actually authorized by the user? The recent "ClawJacked" and "Calendar Injection" exploits show that unauthenticated "intent" is the primary attack vector. MCP Any must provide a way for agents to cryptographically sign their messages and handoffs.

## 2. Goals & Non-Goals
*   **Goals:**
    *   Provide every agent registered in MCP Any with a unique cryptographic identity (Ed25519 keypair).
    *   Enable agents to sign A2A messages and "Intent Tokens" for tool calls.
    *   Support non-repudiable audit logs of agent-to-agent delegations.
    *   Implement a "Web of Trust" where agents can verify each other's identities via the MCP Any gateway.
*   **Non-Goals:**
    *   Managing human user identities (this is handled by the existing Auth layer).
    *   Defining the signing algorithm for third-party frameworks (we provide the standard; they must adopt it).

## 3. Critical User Journey (CUJ)
*   **User Persona:** Security-Conscious Agent Developer.
*   **Primary Goal:** Ensure that only my "Verified Manager Agent" can trigger the "Deployment Tool."
*   **The Happy Path (Tasks):**
    1.  User starts MCP Any; it generates a root `AgentCA` and individual keys for registered agents.
    2.  The Manager Agent receives its private key and identity certificate.
    3.  When the Manager Agent calls the Deployment Agent (A2A), it includes a `Sec-Agent-Sig` header.
    4.  The Deployment Agent (or MCP Any gateway) verifies the signature against the Manager's public key.
    5.  The tool call proceeds only if the identity and intent signature are valid.

## 4. Design & Architecture
*   **System Flow:**
    - **Identity Provisioning**: The `IdentityManager` service manages the lifecycle of agent keys (stored in `MCPANY_DB_PATH`).
    - **Signature Verification Middleware**: A new middleware in the pipeline that intercepts A2A and MCP calls to verify the `Sec-Agent-Sig`.
    - **Trust Registry**: A public-key distribution endpoint where agents can fetch the certificates of their peers.
*   **APIs / Interfaces:**
    - `GET /api/v1/identity/public-keys`: Returns a list of trusted agent public keys.
    - `POST /api/v1/identity/sign`: Internal-only endpoint for agents to request a signature for a payload.
*   **Data Storage/State:** Encrypted storage of private keys in the SQLite database, protected by the master `MCPANY_API_KEY`.

## 5. Alternatives Considered
*   **Simple API Keys per Agent**: High rotation overhead and no non-repudiation. *Rejected*.
*   **OIDC for Agents**: Too heavyweight for local/low-latency agent swarms. *Rejected* in favor of lightweight Ed25519 signatures.

## 6. Cross-Cutting Concerns
*   **Security (Zero Trust):** This is the "Identity" pillar of Zero Trust for agents. It prevents "Shadow Agents" from executing unauthorized tasks.
*   **Observability:** The UI "Agent Chain Tracer" will now display a "Verified" badge next to every signed handoff.

## 7. Evolutionary Changelog
*   **2026-03-05:** Initial Document Creation.
